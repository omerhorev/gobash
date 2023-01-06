package gobash

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"

	"github.com/omerhorev/gobash/ast"
	"github.com/omerhorev/gobash/command"
	"github.com/omerhorev/gobash/utils"

	"github.com/pkg/errors"
)

// internally used to represent io redirections
type ioRedirection struct {
	To   string
	Fd   int
	Mode ast.IORedirectionMode
}

// Will be used instead of os.OpenFile when opening files by the shell
type OpenFileFunc func(path string, flag int, perm os.FileMode) (io.ReadWriteCloser, error)

// Will be used when changing a folder using cd
type ChangeDirFunc func(path string) (newPath string, err error)

const (
	defaultIFS = "\n\t "
	retErr     = 127 // the return code when error happens
)

// Settings for Executor
type ExecutorSettings struct {
	// Remove the cd command
	NoCd bool

	// Remove the exec command
	NoExec bool

	// Will be used instead of os.OpenFile when opening files by the shell
	OpenFunc OpenFileFunc

	// The method used in the cd builtin execution-unit to change the working directory
	// if null, implementation based on os.Stat will be used
	CdFunc ChangeDirFunc

	// Disable opening new files by the shell
	// If set, the following commands will result an error:
	//  - `echo 1 > /tmp/x`
	//  - `echo 1 < /tmp/x`
	//  - `cat 3<>/tmp/x `
	//
	// Note: this only affect file openings, no the redirection subsystem.
	// The command `echo 1>&2` would not raise an error.
	DisableFileOpen bool

	// Exit the execution when an IO redirection error happens
	// (see 2.8.1 Consequences of Shell Errors)
	StopOnIORedirectionError bool

	// Exit the execution when an unknown command error happens
	// (see 2.8.1 Consequences of Shell Errors)
	StopOnUnknownCommand bool
}

// The Executor receives an AST and executes it.
// It executes the script if an almost-compliant way to IEEE1003.1.
// The key differences are:
// - commands: The Executor supports special commands that connect to a Golang method. Use RegisterCommand to add such commands.
type Executor struct {
	Settings     ExecutorSettings  // Settings for the executor
	ExecEnv      *ExecEnv          // The current execution environment (env-vars, open files, etc)
	Commands     []command.Command // The current registered command
	astNodeStack []ast.Node        // the current ast node stack
}

// Creates a new executor with settings. The newly created Executor has no
// stdin/stdout/stderr streams. You must manually set them using the Std[in/out/err]
// fields.
func NewExecutor(settings ExecutorSettings) *Executor {
	executor := &Executor{
		Settings:     settings,
		Commands:     []command.Command{},
		ExecEnv:      newExecEnv(),
		astNodeStack: []ast.Node{},
	}

	return executor
}

// Run the program specified.
//
// The program will be executed on the same Goroutine and will block until
// it finishes execution.
func (e *Executor) Run(program *ast.Program) error {
	for _, node := range program.Commands {
		if _, err := e.executeNode(node, e.ExecEnv); err != nil {
			return err
		}
	}

	return nil
}

// Register a one or more new commands
//
// For example, add all the default commands:
//
//	e.AddCommands(command.Default...)
func (e *Executor) AddCommands(commands ...command.Command) {
	e.Commands = append(e.Commands, commands...)
}

// Sets the stdin of the executor. If the reader is also an io.Closer, it uses
// the reader's close method. Otherwise, a no-op Closer is used.
func (e *Executor) SetStdin(r io.Reader) {
	var rc io.Reader = nil
	if c, ok := r.(io.ReadCloser); ok {
		rc = c
	} else {
		rc = io.NopCloser(r)
	}

	e.ExecEnv.Files[0] = &utils.ErrorReadWriterErrW{Reader: rc}
}

// Sets the stdout of the executor. If the writer is also an io.Closer, it uses
// the writer's close method. Otherwise, a no-op Closer is used.
func (e *Executor) SetStdout(w io.Writer) {
	var wc io.Writer = nil
	if c, ok := w.(io.WriteCloser); ok {
		wc = c
	} else {
		wc = utils.NewNopWriteCloser(w)
	}

	e.ExecEnv.Files[1] = utils.ErrorReadWriterErrR{Writer: wc}
}

// Sets the stderr of the executor. If the writer is also an io.Closer, it uses
// the writer's close method. Otherwise, a no-op Closer is used.
func (e *Executor) SetStderr(w io.Writer) {
	var wc io.Writer = nil
	if c, ok := w.(io.WriteCloser); ok {
		wc = c
	} else {
		wc = utils.NewNopWriteCloser(w)
	}

	e.ExecEnv.Files[2] = utils.ErrorReadWriterErrR{Writer: wc}
}

// Change the shell's working directory.
// This method will use the CdFunc in the settings if one exists.
func (e *Executor) Cd(path string) error {
	cdFunc := e.cdFunc()

	newPath, err := cdFunc(path)
	if err != nil {
		return err
	}

	e.ExecEnv.WorkingDirectory = newPath

	return nil
}

func (e *Executor) getCommand(name string) (command.Command, error) {
	commands := append(e.Commands, &cdBuiltinCommand{Executor: e})

	for _, command := range commands {
		if command.Match(name) {
			return command, nil
		}
	}

	return nil, newUnknownCommandError(name)
}

func (e *Executor) executeNode(node ast.Node, env *ExecEnv) (ret int, err error) {
	e.astNodeStack = append(e.astNodeStack, node)

	switch n := node.(type) {
	case *ast.Background:
		ret, err = e.executeBackground(n, env)
	case *ast.Binary:
		ret, err = e.executeBinary(n, env)
	case *ast.Pipe:
		ret, err = e.executePipe(n, env)
	case *ast.SimpleCommand:
		ret, err = e.executeSimpleCommand(n, env)
	case *ast.String:
		ret, err = e.executeString(n, env)
	case *ast.Expr:
		ret, err = e.executeExpr(n, env)
	case *ast.Backtick:
		ret, err = e.executeBacktick(n, env)
	default:
		ret, err = retErr, fmt.Errorf("unsupported execution %T", n)
	}

	e.astNodeStack = e.astNodeStack[:len(e.astNodeStack)-1]

	if newErr := e.HandleError(err); newErr != nil {
		ret, err = retErr, newErr
	} else {
		err = nil
	}

	return
}

func (e *Executor) isRunInBackground() bool {
	for i := range e.astNodeStack {
		v := e.astNodeStack[len(e.astNodeStack)-1-i]
		if _, ok := v.(ast.Background); ok {
			return true
		}
	}

	return false
}

func (e *Executor) executeExpr(node *ast.Expr, env *ExecEnv) (int, error) {
	b := bytes.Buffer{}

	for _, n := range node.Nodes {
		b.Reset()

		if _, err := e.executeNodeOverrideStdInOut(n, env, env.Stdin(), &b); err != nil {
			return retErr, err
		}

		if _, ok := n.(ast.DoubleQuote); !ok {
			sb := bufio.NewWriter(env.Stdout())

			s := e.newFieldSplitScanner(&b)
			more := s.Scan()
			for more {
				sb.WriteString(s.Text())

				more = s.Scan()

				if more {
					sb.WriteRune(' ')
				}
			}

			sb.Flush()
		} else {
			io.Copy(env.Stdout(), &b)
		}
	}

	return 0, nil
}

func (e *Executor) executeString(node *ast.String, env *ExecEnv) (int, error) {
	env.Stdout().Write([]byte(node.Value))

	return 0, nil
}

func (e *Executor) executeBacktick(node *ast.Backtick, env *ExecEnv) (int, error) {
	return e.executeNode(node.Node, env)
}

func (e *Executor) executeBackground(node *ast.Background, env *ExecEnv) (int, error) {
	e.executeNode(node.Child, env)

	return 0, nil
}

func (e *Executor) executeBinary(node *ast.Binary, env *ExecEnv) (int, error) {
	ret, err := e.executeNode(node.Left, env)
	if err != nil {
		return retErr, err
	}

	if ret == 0 && node.IsAnd() || ret != 0 && node.IsOr() {
		ret, err := e.executeNode(node.Right, env)
		if err != nil {
			return retErr, err
		}

		return ret, nil
	}

	return ret, nil
}

func (e *Executor) executePipe(node *ast.Pipe, env *ExecEnv) (int, error) {
	if len(node.Commands) == 0 {
		return retErr, errors.New("pipe with no commands")
	}

	if len(node.Commands) == 1 {
		return e.executeNode(node.Commands[0], env)
	}

	// setup stdin, stdout and stderr
	_r := io.NopCloser(e.ExecEnv.Stdin())
	wg := sync.WaitGroup{}

	for i := 0; i < len(node.Commands)-1; i++ {
		n := node.Commands[i]

		r, w := io.Pipe()

		wg.Add(1)
		go func(reader io.ReadCloser, writer io.WriteCloser) {
			e.executeNodeOverrideStdInOut(n, env, reader, writer)
			writer.Close()
			reader.Close()

			wg.Done()
		}(_r, w)

		_r = r
	}

	n := node.Commands[len(node.Commands)-1]
	ret, err := e.executeNodeOverrideStdInOut(n, env, _r, e.ExecEnv.Stdout())
	_r.Close()

	wg.Wait()

	return ret, err
}

func (e *Executor) executeSimpleCommand(node *ast.SimpleCommand, env *ExecEnv) (int, error) {
	name, args, assignments, redirects, err := e.expandSimpleCommand(node)
	if err != nil {
		return retErr, err
	}

	newEnv := env.New()

	for _, v := range redirects {
		if file, err := e.getIORedirectFile(v, newEnv); err != nil {
			return retErr, err
		} else {
			defer file.Close()

			newEnv.Files[v.Fd] = file
		}
	}

	cmdEnv := e.createCommandEnv(newEnv)
	for k, v := range assignments {
		cmdEnv.Env[k] = v
	}
	cmdEnv.Args = append([]string{name}, args...)

	if e.isRunInBackground() {
		// TODO: run in background
		return retErr, errors.New("unimplemented")
	}

	cmd, err := e.getCommand(name)
	if err != nil {
		return retErr, err
	}

	return cmd.Execute(cmdEnv.Args, cmdEnv), nil
}

func (e *Executor) createCommandEnv(env *ExecEnv) *command.Env {
	filesWithoutClose := map[int]io.ReadWriter{}
	for fd, f := range env.Files {
		filesWithoutClose[fd] = f
	}

	envVars := map[string]string{}
	for k, v := range env.Params {
		envVars[k] = v
	}

	return &command.Env{
		Files:    filesWithoutClose,
		Env:      envVars,
		OpenFunc: e.openFileFunc(),
	}
}

func (e *Executor) executeNodeOverrideStdInOut(node ast.Node, env *ExecEnv, in io.Reader, out io.Writer) (int, error) {
	envCopy := env.New()
	envCopy.Files[0] = &utils.ErrorReadWriterErrW{Reader: in}
	envCopy.Files[1] = &utils.ErrorReadWriterErrR{Writer: out}

	return e.executeNode(node, envCopy)
}

func (e *Executor) getIORedirectFile(redirection *ioRedirection, env *ExecEnv) (io.ReadWriteCloser, error) {
	if redirection.Mode == ast.IORedirectionModeInputFd || redirection.Mode == ast.IORedirectionModeOutputFd {
		fd, err := strconv.Atoi(redirection.To)
		if err != nil {
			return nil, newIORedirectionError(errors.Errorf("bad fd number %s", redirection.To))
		}

		file, ok := env.Files[fd]
		if !ok {
			return nil, newIORedirectionError(errors.Errorf("%d: bad file descriptor", fd))
		}

		if redirection.Mode == ast.IORedirectionModeOutputFd {
			return utils.ErrorReadWriterErrR{Writer: file}, nil
		} else {
			return &utils.ErrorReadWriterErrW{Reader: file}, nil
		}

	} else {
		path := redirection.To
		flags := 0

		switch redirection.Mode {
		case ast.IORedirectionModeInput:
			flags = os.O_RDONLY
		case ast.IORedirectionModeOutput:
			flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		case ast.IORedirectionModeInputOutput:
			flags = os.O_RDWR | os.O_CREATE | os.O_TRUNC
		case ast.IORedirectionModeOutputAppend:
			flags = os.O_WRONLY | os.O_CREATE | os.O_APPEND
		}

		f, err := e.openFile(path, flags, 0666)
		if err != nil {
			return nil, newIORedirectionError(errors.Wrap(err, path))
		}

		return f, nil
	}
}

func (e *Executor) openFile(path string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
	if e.Settings.DisableFileOpen {
		return nil, errors.Errorf("open disabled")
	}

	return e.openFileFunc()(path, flag, perm)
}

func (e *Executor) openFileFunc() OpenFileFunc {
	if e.Settings.OpenFunc != nil {
		return e.Settings.OpenFunc
	}

	return func(path string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
		return os.OpenFile(path, flag, perm)
	}
}

func (e *Executor) cdFunc() ChangeDirFunc {
	if e.Settings.CdFunc != nil {
		return e.Settings.CdFunc
	} else {
		return defaultCdFunc
	}
}

func (e *Executor) HandleError(err error) error {
	if err := e.error(err); err != nil {
		return err
	}

	if IsIORedirectionError(err) {
		if e.Settings.StopOnIORedirectionError {
			return err
		} else {
			return nil
		}
	}

	if IsUnknownCommandError(err) {
		if e.Settings.StopOnUnknownCommand {
			return err
		} else {
			return nil
		}
	}

	return err
}

func (e *Executor) error(err error) (retErr error) {
	if err != nil {
		if str := err.Error(); str != "" {
			_, retErr = e.ExecEnv.Stderr().Write([]byte(str + "\n"))
		}
	}

	return
}

func (e *Executor) expandSimpleCommand(node *ast.SimpleCommand) (command string, args []string, assignments map[string]string, redirects []*ioRedirection, err error) {
	command = ""
	args = []string{}
	assignments = map[string]string{}
	redirects = []*ioRedirection{}
	var val string

	command, err = e.expandExpr(node.Word)
	if err != nil {
		return
	}

	for k, v := range node.Assignments {
		val, err = e.expandExpr(v)
		if err != nil {
			return
		}

		assignments[k] = val
	}

	for _, v := range node.Args {
		val, err = e.expandExpr(v)
		if err != nil {
			return
		}

		args = append(args, val)
	}

	for _, v := range node.Redirects {
		val, err = e.expandExpr(v.Value)
		if err != nil {
			return
		}

		redirects = append(redirects, &ioRedirection{
			Fd:   v.Fd,
			Mode: v.Mode,
			To:   val,
		})
	}

	return
}

func (e *Executor) expandExpr(node ast.Node) (string, error) {
	b := bytes.Buffer{}
	_, err := e.executeNodeOverrideStdInOut(node, e.ExecEnv, utils.Null, &b)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

func (e *Executor) newFieldSplitScanner(reader io.Reader) *bufio.Scanner {
	return utils.NewRunesScanner(reader, e.getIFS())
}

func (e *Executor) getIFS() []rune {
	s := e.ExecEnv.GetParamDefault("IFS", defaultIFS)
	runes := []rune{}
	for _, r := range s {
		runes = append(runes, r)
	}

	return runes
}

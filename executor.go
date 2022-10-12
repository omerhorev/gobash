package gobash

import (
	"bytes"
	"io"
	"os"
	"sort"

	"github.com/omerhorev/gobash/ast"
	"github.com/omerhorev/gobash/utils"
	"github.com/pkg/errors"
)

// Will be used instead of os.OpenFile when opening files by the shell
type ExecutorOpenFileFunc func(path string, flag int, perm os.FileMode) (io.ReadWriteCloser, error)

const (
	retErr = 127 // the return code when error happens
)

// Settings for Executor
type ExecutorSettings struct {
	// Will be used instead of os.OpenFile when opening files by the shell
	OpenFunc ExecutorOpenFileFunc

	// Disable opening new files by the shell
	// If set, the following commands will result an error:
	//  - `echo 1 > /tmp/x`
	//  - `echo 1 < /tmp/x`
	//  - `cat 3<>/tmp/x `
	//
	// Note: this only affect file openings, no the redirection subsystem.
	// The command `echo 1>&2` would not raise an error.
	DisableFileOpen bool
}

// The Executor receives an AST and executes it.
// It executes the script if an almost-compliant way to IEEE1003.1.
// The key differences are:
// - commands: The Executor supports special commands that connect to a Golang method. Use RegisterCommand to add such commands.
type Executor struct {
	Settings ExecutorSettings // Settings for the executor
	ExecEnv  *ExecEnv         // The current execution environmnet (env-vars, open files, etc)
	Commands []Command        // The current registered command
	Stdin    io.Reader        // the stdin stream
	Stdout   io.Writer
	Stderr   io.Writer
}

// Creates a new executor with settings. The newly created Executor has no
// stdin/stdout/stderr streams. You must manually set them using the Std[in/out/err]
// fields.
func NewExecutor(settings ExecutorSettings) *Executor {
	return &Executor{
		Settings: settings,
		Commands: []Command{},
		ExecEnv:  newExecEnv(),
		Stdout:   io.Discard,
		Stderr:   io.Discard,
		Stdin:    bytes.NewReader([]byte{}),
	}
}

// Run the program specified.
//
// The program will be executed on the same Goroutine and will block until
// it finishes execution.
func (e *Executor) Run(program *ast.Program) error {
	for _, cmd := range program.Commands {
		if _, err := e.executeNode(cmd); err != nil {
			return err
		}
	}

	return nil
}

// Register a new command
func (e *Executor) RegisterCommand(command Command) {
	e.Commands = append(e.Commands, command)
	sort.Slice(e.Commands, func(i, j int) bool {
		return e.Commands[i].Priority() > e.Commands[j].Priority()
	})
}

func (e *Executor) getCommand(word string) (Command, error) {
	for _, command := range e.Commands {
		if command.Match(word) {
			return command, nil
		}
	}

	return nil, errors.Errorf("%s: command not found", word)
}

func (e *Executor) executeNode(node ast.Node) (int, error) {
	switch n := node.(type) {
	case *ast.Program:
		return e.executeProgram(n)
	case *ast.Background:
		return e.executeBackground(n)
	case *ast.Binary:
		return e.executeBinary(n)
	case *ast.Pipe:
		return e.executePipe(n)
	}

	return retErr, errors.New("unsupported simple execution")
}

func (e *Executor) executeCommandEnv(node ast.Node, env *CommandExecEnv) (int, error) {
	switch n := node.(type) {
	case *ast.SimpleCommand:
		return e.executeSimpleCommand(n, env)
	}

	return retErr, errors.New("unsupported simple execution")
}

func (e *Executor) executeProgram(node *ast.Program) (int, error) {
	for _, node := range node.Commands {
		if _, err := e.executeNode(node); err != nil {
			return retErr, err
		}
	}

	return 0, nil
}

func (e *Executor) executeBackground(node *ast.Background) (int, error) {
	// TODO: support run in background
	e.executeNode(node.Child)

	return 0, nil
}

func (e *Executor) executeBinary(node *ast.Binary) (int, error) {
	ret, err := e.executeNode(node.Left)
	if err != nil {
		return retErr, err
	}

	if ret == 0 && node.IsAnd() || ret != 0 && node.IsOr() {
		ret, err := e.executeNode(node.Right)
		if err != nil {
			return retErr, err
		}

		return ret, nil
	}

	return ret, nil
}

func (e *Executor) executePipe(node *ast.Pipe) (int, error) {
	if len(node.Commands) == 0 {
		return retErr, errors.New("pipe with no commands")
	}

	if len(node.Commands) == 1 {
		return e.executeCommand(node.Commands[0], e.Stdin, e.Stdout, e.Stderr)
	}

	// setup stdin, stdout and stderr
	_r := e.Stdin

	for i := 0; i < len(node.Commands)-1; i++ {
		n := node.Commands[i]

		r, w := io.Pipe()

		go func(reader io.Reader) {
			e.executeCommand(n, reader, w, e.Stderr)
			w.Close()
		}(_r)

		_r = r
	}

	n := node.Commands[len(node.Commands)-1]
	return e.executeCommand(n, _r, e.Stdout, e.Stderr)
}

func (e *Executor) executeSimpleCommand(node *ast.SimpleCommand, env *CommandExecEnv) (int, error) {
	cmd, err := e.getCommand(node.Word)
	if err != nil {
		return 1, err
	}

	for k, v := range node.Assignments {
		env.Env[k] = v
	}

	for k, v := range node.Redirects {
		if file, err := e.getIORedirectFile(v, env); err != nil {
			return retErr, err
		} else {
			env.Files[k] = file
		}
	}

	cmd.Execute(append([]string{node.Word}, node.Args...), env)

	return 0, nil
}

func (e *Executor) executeCommand(node ast.Node, in io.Reader, out io.Writer, err io.Writer) (int, error) {
	env := e.ExecEnv.CommandExecEnv() // shallow copy
	env.Files[0] = &utils.ErrorReadWriterErrW{Reader: in}
	env.Files[1] = &utils.ErrorReadWriterErrR{Writer: out}
	env.Files[2] = &utils.ErrorReadWriterErrR{Writer: err}

	return e.executeCommandEnv(node, env)
}

func (e *Executor) getIORedirectFile(redirection *ast.SimpleCommandIORedirection, env *CommandExecEnv) (io.ReadWriteCloser, error) {
	if redirection.Mode == ast.SimpleCommandIORedirectionModeInputFd {
		fd := redirection.Value.(int) // value must be int at this point
		f, ok := env.Files[fd]
		if !ok {
			return nil, errors.Errorf("%d: bad file descriptor", fd)
		}

		return &utils.ErrorReadWriterErrW{Reader: f}, nil
	}

	if redirection.Mode == ast.SimpleCommandIORedirectionModeInput {
		path := redirection.Value.(string) // value must be string at this point
		f, err := e.openFile(path, os.O_RDONLY, 0)
		if err != nil {
			return nil, errors.Wrap(err, path)
		}

		return f, nil
	}

	if redirection.Mode == ast.SimpleCommandIORedirectionModeOutputFd {
		fd := redirection.Value.(int) // value must be int at this point
		f, ok := env.Files[fd]
		if !ok {
			return nil, errors.Errorf("%d: bad file descriptor", fd)
		}

		return utils.ErrorReadWriterErrR{Writer: f}, nil
	}

	if redirection.Mode == ast.SimpleCommandIORedirectionModeOutput {
		path := redirection.Value.(string) // value must be string at this point
		f, err := e.openFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)

		if err != nil {
			return nil, errors.Wrap(err, path)
		}

		return f, nil
	}

	if redirection.Mode == ast.SimpleCommandIORedirectionModeInputOutput {
		path := redirection.Value.(string) // value must be string at this point
		f, err := e.openFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)

		if err != nil {
			return nil, errors.Wrap(err, path)
		}

		return f, nil
	}

	if redirection.Mode == ast.SimpleCommandIORedirectionModeOutputAppend {
		path := redirection.Value.(string) // value must be string at this point
		f, err := e.openFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

		if err != nil {
			return nil, errors.Wrap(err, path)
		}

		return f, nil
	}

	return nil, errors.Errorf("unknown redirection")
}

func (e *Executor) openFile(path string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
	if e.Settings.DisableFileOpen {
		return nil, errors.Errorf("open disabled")
	}

	if e.Settings.OpenFunc != nil {
		return e.Settings.OpenFunc(path, flag, perm)
	}

	return os.OpenFile(path, flag, perm)
}

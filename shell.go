package gobash

import (
	"errors"
	"io"

	"github.com/omerhorev/gobash/command"
)

type ShellSettings struct {
	ExecutorSettings

	// Starts an interactive-mode shell
	Interactive bool
}

var InteractiveDefaultSettings ShellSettings = ShellSettings{
	ExecutorSettings: ExecutorSettings{
		NoCd:                     false,
		NoExec:                   false,
		OpenFunc:                 nil, // use default os.OpenFile
		CdFunc:                   nil, // use default fs based implementation
		DisableFileOpen:          false,
		StopOnBuiltinError:       false,
		StopOnIORedirectionError: false,
		StopOnUnknownCommand:     false,
	},

	Interactive: true,
}

// Shell is a golang implemented bash-like shell
type Shell struct {
	Settings ShellSettings
	executor *Executor
}

func NewShell(settings ShellSettings) *Shell {
	return &Shell{
		executor: NewExecutor(settings.ExecutorSettings),
		Settings: settings,
	}
}

// Sets the stdin of the shell. If the reader is also an io.Closer, it uses
// the reader's close method. Otherwise, a no-op Closer is used.
func (s *Shell) SetStdin(r io.Reader) {
	s.executor.SetStdin(r)
}

// Sets the stdout of the shell. If the writer is also an io.Closer, it uses
// the writer's close method. Otherwise, a no-op Closer is used.
func (s *Shell) SetStdout(w io.Writer) {
	s.executor.SetStdout(w)
}

// Sets the stderr of the shell. If the writer is also an io.Closer, it uses
// the writer's close method. Otherwise, a no-op Closer is used.
func (s *Shell) SetStderr(w io.Writer) {
	s.executor.SetStderr(w)
}

// Register a one or more new commands
//
// For example, add all the default commands:
//
//	s.AddCommands(command.Default...)
func (s *Shell) AddCommands(cmd ...command.Command) {
	s.executor.AddCommands(cmd...)
}

func (s *Shell) Run(expression string) error {
	tokenizer := NewTokenizerShort(expression)

	tokens, err := tokenizer.ReadAll()
	if err != nil {
		return err
	}

	parser := NewParser(tokens, ParserSettings{})

	if err := parser.Parse(); err != nil {
		return err
	}

	return s.executor.Run(parser.Program())
}

// Runs the shell in interactive mode.
//
// Each line is read from the reader and evaluated by the shell. This mode mimics
// the behavior of an interactive terminal session. Prompt is printed between
// line reads.
func (s *Shell) RunInteractive() error {
	if !s.Settings.Interactive {
		return errors.New("unsupported in non-interactive mode")
	}

	lr := NewLineReader(s.executor.ExecEnv.Stdin())
	for {
		// TODO: Prompt

		line, err := lr.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		if err := s.Run(line); s.handleError(err) != nil {
			return err
		}
	}

	return nil
}

func (s *Shell) RunReader(reader io.Reader) error {
	if !s.Settings.Interactive {
		return s.RunScript(reader)
	} else {

	}

	return nil
}

// Evaluate the entire content of the script
func (s *Shell) RunScript(reader io.Reader) error {
	tokenizer := NewTokenizerLong(reader)

	tokens, err := tokenizer.ReadAll()
	if err != nil {
		return err
	}

	parser := NewParser(tokens, ParserSettings{})

	if err := parser.Parse(); err != nil {
		return err
	}

	return s.executor.Run(parser.Program())
}

func (s *Shell) handleError(err error) error {
	if !s.Settings.Interactive {
		return err
	}

	if errors.Is(err, SyntaxError{}) ||
		errors.Is(err, BuiltinError{}) ||
		errors.Is(err, IoRedirectionError{}) {
		return nil
	} else {
		return err
	}
}

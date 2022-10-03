package gobash

import (
	"sort"

	"github.com/omerhorev/gobash/ast"
	"github.com/pkg/errors"
)

var (
	ErrBadAST = errors.New("bad AST node")
)

// settings for Executor
type ExecutorSettings struct{}

// The executor receives an AST and executes it
type Executor struct {
	Settings ExecutorSettings
	Commands []Command
	program  ast.Program
}

// Creates a new executor
func NewExecutor(settings ExecutorSettings, program ast.Program) *Executor {
	return &Executor{
		Settings: settings,
		Commands: []Command{},
		program:  program,
	}
}

// Run the program
//
// The program will be executed on the same thread and will be blocking until
// it finishes execution.
func (e *Executor) Run() error {
	for _, cmd := range e.program.Commands {
		if err := e.executeCommand(cmd); err != nil {
			return err
		}
	}

	return nil
}

// Register a new command
func (e *Executor) Register(command Command) {
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

func (e *Executor) executeCommand(node ast.Command) error {
	if c, ok := node.(*ast.SimpleCommand); ok {
		return e.executeSimpleCommand(c)
	}

	return ErrBadAST
}

func (e *Executor) executeSimpleCommand(node *ast.SimpleCommand) error {
	cmd, err := e.getCommand(node.Word)
	if err != nil {
		return err
	}

	cmd.Execute(nil, nil, nil)
	// TODO: pipe

	return nil
}

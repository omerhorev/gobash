package gobash

import (
	"errors"

	"github.com/omerhorev/gobash/ast"
)

func builtinCd(e *Executor, node *ast.SimpleCommand) error {
	args := append([]string{node.Word}, node.Args...) // like command args

	path := "/"
	if len(args) == 1 {
		if val := e.ExecEnv.GetParam("HOME"); val != "" {
			path = val
		}
	} else if len(args) == 2 {
		path = args[1]
	} else {
		return errors.New("too many arguments")
	}

	newPath, err := e.cdFunc()(path)
	if err != nil {
		return err
	}

	e.ExecEnv.WorkingDirectory = newPath

	return nil
}

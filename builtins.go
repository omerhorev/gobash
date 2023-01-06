package gobash

import (
	"errors"

	"github.com/omerhorev/gobash/command"
)

type cdBuiltinCommand struct {
	*Executor
}

func (c *cdBuiltinCommand) Match(word string) bool { return word == "cd" && !c.Executor.Settings.NoCd }
func (c *cdBuiltinCommand) Execute(args []string, env *command.Env) int {
	path := "/"
	if len(args) == 1 {
		if val := c.Executor.ExecEnv.GetParam("HOME"); val != "" {
			path = val
		} else {
			env.Error(errors.New("HOME not set"))
			return 1
		}
	} else if len(args) == 2 {
		path = args[1]
	} else {
		env.Error(errors.New("too many arguments"))
		return 1
	}

	if err := c.Executor.Cd(path); err != nil {
		env.Error(err)
		return 1
	}

	return 0
}

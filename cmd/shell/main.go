package main

import (
	"os"

	"github.com/omerhorev/gobash"
	"github.com/omerhorev/gobash/command"
)

func main() {
	s := gobash.NewShell(gobash.InteractiveDefaultSettings)
	s.SetStdin(os.Stdin)
	s.SetStdout(os.Stdout)
	s.SetStderr(os.Stderr)

	s.AddCommands(command.Default...)

	if err := s.RunInteractive(); err != nil {
		panic(err)
	}
}

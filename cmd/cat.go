package cmd

import (
	"bufio"

	"github.com/omerhorev/gobash/utils"
)

// like /bin/cat without flags
//
//	cat [FILE]...
var Cat = &SimpleMatchCommand{
	Name: "cat",
	F: func(args []string, e *Env) int {
		doFile := func(path string) error {
			f, err := fileOrStdin(path, e)
			if err != nil {
				return err
			}
			defer f.Close()

			scanner := bufio.NewScanner(f)

			for scanner.Scan() {
				if _, err := e.Print(scanner.Text()); err != nil {
					return err
				}
			}

			return scanner.Err()
		}

		if len(args) == 1 {
			doFile("-")
		}

		for _, arg := range args[1:] {
			if err := doFile(arg); err != nil {
				e.Error(err)
			}
		}

		return 0
	},
}

// like /bin/cat without flags
//
//	cat [FILE]...
var Tac = &SimpleMatchCommand{
	Name: "tac",
	F: func(args []string, e *Env) int {
		doFile := func(path string) error {
			f, err := fileOrStdin(path, e)
			if err != nil {
				return err
			}
			defer f.Close()

			scanner := bufio.NewScanner(f)

			for scanner.Scan() {
				if _, err := e.Print(utils.ReverseString(scanner.Text())); err != nil {
					return err
				}
			}

			return scanner.Err()
		}

		if len(args) == 1 {
			doFile("-")
		}

		for _, arg := range args[1:] {
			if err := doFile(arg); err != nil {
				e.Error(err)
			}
		}

		return 0
	},
}

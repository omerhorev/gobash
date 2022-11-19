package command

import (
	"bufio"
	"io"

	"github.com/omerhorev/gobash/utils"
)

// like /bin/rev
//
//	rev [FILE]...
var Rev = &SimpleMatchCommand{
	Name: "rev",
	F: func(args []string, e *Env) int {
		doFile := func(r io.Reader) error {
			scanner := bufio.NewScanner(r)

			for scanner.Scan() {
				if _, err := e.Print(utils.ReverseString(scanner.Text())); err != nil {
					return err
				}
			}

			return scanner.Err()
		}

		if len(args) == 1 {
			doFile(e.Stdin())
		}

		for _, arg := range args[1:] {
			f, err := e.Open(arg)
			if err != nil {
				e.Error(err)
				continue
			}

			defer f.Close()

			if err := doFile(f); err != nil {
				e.Error(err)
				continue
			}
		}

		return 0
	},
}

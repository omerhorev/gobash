package command

// like /bin/echo
var Echo = &SimpleMatchCommand{
	Name: "echo",
	F: func(args []string, e *Env) int {
		if len(args) > 1 {
			for _, arg := range args[1 : len(args)-1] {
				e.Print(arg, " ")
			}

			e.Print(args[len(args)-1])
		}

		e.Println()

		return 0
	},
}

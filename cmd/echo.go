package cmd

// like /bin/echo
var Echo = &SimpleMatchCommand{
	Name: "echo",
	F: func(args []string, e *Env) int {
		for _, arg := range args[1 : len(args)-1] {
			e.Print(arg, " ")
		}

		e.Print(args[len(args)-1])

		return 0
	},
}

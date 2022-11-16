package cmd

// like /bin/env without flags and command execution
var EnvCmd = &SimpleMatchCommand{
	Name: "env",
	F: func(args []string, e *Env) int {
		if len(args) != 1 {
			e.Println("too many arguments")
			return 1
		}

		for k, v := range e.Env {
			e.Printf("%s=%s\n", k, v)
		}

		return 0
	},
}

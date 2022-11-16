package cmd

var (
	// like /bin/true
	True = &SimpleMatchCommand{
		Name: "true",
		F: func(s []string, e *Env) int {
			return 0
		},
	}

	// like /bin/false
	False = &SimpleMatchCommand{
		Name: "false",
		F: func(s []string, e *Env) int {
			return 1
		},
	}
)

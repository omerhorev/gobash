package gobash

// Exexution environment the contains all the parameters of the current
// execution, like:
//  - Working directory
//  - Shell parameters
//  - Aliases
//  - Shell functions
type ExecEnv struct {
	// Working directory as set by cd
	WorkingDirectory string // Current working directory

	// Shell parameters that are set by variable assignment (set command) or from the
	// System Interfaces volume of POSIX.1-2017 environment inherited by the shell
	// when it begins (export special built-in)
	Params map[string]string
}

func newExecEnv() *ExecEnv {
	return &ExecEnv{
		WorkingDirectory: "/",
		Params:           map[string]string{},
	}
}

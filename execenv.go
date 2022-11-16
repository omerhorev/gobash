package gobash

import (
	"io"

	"github.com/omerhorev/gobash/utils"
)

// Execution environment contains all the parameters of the current
// execution:
//   - Working directory
//   - Shell parameters
//   - Aliases
//   - Shell functions
//   - Open Files (std in/out/err)
type ExecEnv struct {
	// Working directory as set by cd
	WorkingDirectory string

	// Shell parameters that set by variable assignment (`set` command) or from the
	// System Interfaces volume of POSIX.1-2017 environment inherited by the shell
	// when it begins (`export` special built-in).
	Params map[string]string

	// functions

	// Open files that can be used by the process (like stdin[0], stdout[1] and
	// stderr[2]). Just like a file with fd, it can be read from and written to.
	Files map[int]io.ReadWriteCloser
}

func newExecEnv() *ExecEnv {
	return &ExecEnv{
		WorkingDirectory: "/",
		Params:           map[string]string{},
		Files:            map[int]io.ReadWriteCloser{},
	}
}

// Returns the File with fd #0 in a read-only mode.
// if there is no such file, a null writer is returned (io.EOF to all reads)
func (e *ExecEnv) Stdin() io.Reader {
	if r, ok := e.Files[0]; ok {
		return r
	}

	return utils.Null
}

// Returns the File with fd #1 in a write-only mode.
// if there is no such file, a null writer is returned (io.EOF to all writes)
func (e *ExecEnv) Stdout() io.Writer {
	if r, ok := e.Files[1]; ok {
		return r
	}

	return utils.Null
}

// Returns the File with fd #2 in a write-only mode.
// if there is no such file, a null writer is returned (io.EOF to all writes)
func (e *ExecEnv) Stderr() io.Writer {
	if r, ok := e.Files[2]; ok {
		return r
	}

	return utils.Null
}

// Creates a new environment from this environment
func (e *ExecEnv) New() *ExecEnv {
	commandExecEnv := &ExecEnv{
		WorkingDirectory: e.WorkingDirectory,
		Params:           make(map[string]string),
		Files:            make(map[int]io.ReadWriteCloser),
	}

	for k, v := range e.Params {
		commandExecEnv.Params[k] = v
	}

	for k, v := range e.Files {
		commandExecEnv.Files[k] = v
	}

	return commandExecEnv
}

// GetParam retrieves the value of the parameter variable named by the key.
// It returns the value, which will be empty if the variable is not present.
func (e *ExecEnv) GetParam(key string) string {
	if value, exists := e.Params[key]; exists {
		return value
	} else {
		return ""
	}
}

// SetParam sets the value of the parameter variable named as key
func (e *ExecEnv) SetParam(key string, value string) {
	e.Params[key] = value
}

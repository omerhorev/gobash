package gobash

import (
	"io"

	"github.com/omerhorev/gobash/utils"
)

// Execution environment contains all the parameters of the current
// execution:
//  - Working directory
//  - Shell parameters
//  - Aliases
//  - Shell functions
//  - Open Files (std in/out/err)
type ExecEnv struct {
	// Working directory as set by cd
	WorkingDirectory string

	// Shell parameters that set by variable assignment (`set` command) or from the
	// System Interfaces volume of POSIX.1-2017 environment inherited by the shell
	// when it begins (`export` special built-in).
	Params map[string]string

	// Envrionment variables.
	Env map[string]string

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

// Creates a new CommandExecEnv from this ExecEnv
func (e ExecEnv) CommandExecEnv() *CommandExecEnv {
	commandExecEnv := &CommandExecEnv{
		WorkingDirectory: e.WorkingDirectory,
		Env:              make(map[string]string),
		Files:            make(map[int]io.ReadWriteCloser),
	}

	for k, v := range e.Params {
		commandExecEnv.Env[k] = v
	}

	for k, v := range e.Env {
		commandExecEnv.Env[k] = v
	}

	for k, v := range e.Files {
		commandExecEnv.Files[k] = v
	}

	return commandExecEnv
}

// Execution environment that is passed into the command.
// Changes does not affect the main ExecEnv.
type CommandExecEnv struct {
	// Working directory as set by cd
	WorkingDirectory string

	// Environment variables that are passed to the command. Environment variables are
	// inherited from the main ExecEnv, and include it's environment variables and params
	Env map[string]string

	// Open files that can be used by the process (like stdin[0], stdout[1] and
	// stderr[2]). Just like a file with fd, it can be read from and written to.
	Files map[int]io.ReadWriteCloser
}

// Returns the File with fd #0 in a read-only mode.
// if there is no such file, a null writer is returned (io.EOF to all reads)
func (e *CommandExecEnv) Stdin() io.Reader {
	if r, ok := e.Files[0]; ok {
		return r
	}

	return utils.Null
}

// Returns the File with fd #1 in a write-only mode.
// if there is no such file, a null writer is returned (io.EOF to all writes)
func (e *CommandExecEnv) Stdout() io.Writer {
	if r, ok := e.Files[1]; ok {
		return r
	}

	return utils.Null
}

// Returns the File with fd #2 in a write-only mode.
// if there is no such file, a null writer is returned (io.EOF to all writes)
func (e *CommandExecEnv) Stderr() io.Writer {
	if r, ok := e.Files[2]; ok {
		return r
	}

	return utils.Null
}

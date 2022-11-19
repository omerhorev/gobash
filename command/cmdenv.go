package command

import (
	"fmt"
	"io"
	"os"

	"github.com/omerhorev/gobash/utils"
)

var (
	fdStdin  = 0
	fdStdout = 1
	fdStderr = 2
)

// The Env struct used to emulate the process environment that a
// command is executed in. It allows the command implementation to access
// fd, open files, print to stdout and more
type Env struct {
	// The Open function used by the shell. Use this function to open files
	// to keep compatibility (like os.OpenFile)
	OpenFunc func(path string, flag int, perm os.FileMode) (io.ReadWriteCloser, error)

	// Open file descriptors (0 is stdin, 1 stdout, 2 stderr)
	Files map[int]io.ReadWriter

	// Environment variables passed to the (like os.Env)
	Env map[string]string

	// The arguments that will be passed to the command (like os.Args)
	Args []string
}

// Returns the file with the fd provided or io.ErrClosed
func (e *Env) GetFile(fd int) (io.ReadWriter, error) {
	if file, exists := e.Files[fd]; exists {
		return file, nil
	} else {
		return nil, os.ErrClosed
	}
}

// Returns the stdin file
func (e *Env) Stdin() io.Reader {
	if file, err := e.GetFile(fdStdin); err != nil {
		return utils.Null
	} else {
		return file
	}
}

// Returns the stdout file
func (e *Env) Stdout() io.Writer {
	if file, err := e.GetFile(fdStdout); err != nil {
		return utils.Null
	} else {
		return file
	}
}

// Returns the stderr file
func (e *Env) Stderr() io.Writer {
	if file, err := e.GetFile(fdStderr); err != nil {
		return utils.Null
	} else {
		return file
	}
}

// like fmt.Print (to stdout)
func (e *Env) Print(v ...any) (int, error) {
	return fmt.Fprint(e.Stdout(), v...)
}

// like fmt.Printf (to stdout)
func (e *Env) Printf(format string, v ...any) (int, error) {
	return fmt.Fprintf(e.Stdout(), format, v...)
}

// like fmt.Println (to stdout)
func (e *Env) Println(v ...any) (int, error) {
	return fmt.Fprintln(e.Stdout(), v...)
}

// like os.Open (with OpenFileFunc)
func (e *Env) Open(path string) (io.ReadWriteCloser, error) {
	return e.OpenFile(path, os.O_RDONLY, 0)
}

// like os.OpenFile (with OpenFileFunc)
func (e *Env) OpenFile(path string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
	return e.OpenFunc(path, flag, perm)
}

// Returns the program name (e.Args[0])
func (e *Env) Name() string {
	return e.Args[0]
}

// Prints out errors to stderr in the format:
//
//	[command name]: [error string]
//
// For example:
//
//	rev: cannot open /tmp/fil1: No such file or directory
func (e *Env) Error(err error) {
	fmt.Fprintf(e.Stderr(), "%s: %s", e.Name(), err.Error())
}

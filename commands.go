package gobash

import "io"

var (
	DefaultCmdPriority = 100
)

// Command is a the basic execution element of the shell. The user can implement
// custom command behaviour using this interface.
//
// For the default behaviour use Cmd struct
type Command interface {
	// In case several commands match, execute the one with the highest priority.
	Priority() int

	// Given a command word, return whether this is the command.
	// for example:
	//  return word == "ls"
	// will implement the ls command.
	//
	//  return true
	// will implement a
	Match(word string) bool

	// Execute the command. If error is returned, the shell will stop execution
	// and return the error.
	Execute(stdin io.Reader, stdout, stderr io.Writer) (returnCode int, err error)
}

type Cmd struct {
	Name string
	f    func(stdin io.Reader, stdout, stderr io.Writer) (int, error)
}

func (c *Cmd) Priority() int          { return DefaultCmdPriority }
func (c *Cmd) Match(word string) bool { return word == c.Name }
func (c *Cmd) Execute(stdin io.Reader, stdout, stderr io.Writer) (int, error) {
	return c.f(stdin, stdout, stderr)
}

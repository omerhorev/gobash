package command

var (
	DefaultExecUnitPriority = 100
	BuiltinExecUnitPriority = 10
)

// Command is the basic execution unit of the shell. It allows the user to define
// functions that can be called from the shell but implemented in Golang.
//
// When the shell executes a command, it searches for a matching Command using the
// Match function and then executes it using the
type Command interface {
	// Given a command word, return whether this is the command.
	// for example:
	//  return word == "ls"
	// will implement the ls command.
	//
	//  return true
	// will implement a
	Match(word string) bool

	// Execute the command. If an error is returned, the shell will stop execution
	// and return the error.
	Execute([]string, *Env) int
}

// SimpleMatchCommand is a command with only one name (simplified match function)
type SimpleMatchCommand struct {
	Name string                   // The command name. Cannot contain slash (/).
	F    func([]string, *Env) int // The function that will be executed.
}

func (c *SimpleMatchCommand) Match(word string) bool              { return word == c.Name }
func (c *SimpleMatchCommand) Execute(args []string, env *Env) int { return c.F(args, env) }

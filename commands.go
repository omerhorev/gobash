package gobash

var (
	DefaultCmdPriority = 100
)

type CommandSettings interface {
}

// Command is one of the basic execution units of the shell.
//
// The user can implement a custom command that will be executed after builtin commands
// and before the path execution (if supported by the shell)
//
// Commands can be used to implement "bash functions" that are written in Golang, instead
// of a scripting language.
//
// The Command interface offers various customizable behaviors. If you need a "simple"
// command with default properties, use the Cmd struct.
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
	Match(word string) bool // TODO: Document name constraints

	// Execute the command. If an error is returned, the shell will stop execution
	// and return the error.
	Execute([]string, *CommandExecEnv) (returnCode int, err error)
}

// Cmd is syntax-sugar for "regular" commands.
// The Cmd command has a single name that is used to invoke it and a default priority.
// Just like a "normal" bash command, The Cmd command has no way to stop the execution of
// the program, thus, can't return an error (but can return a return code)
type Cmd struct {
	Name string // TODO: Document name constraints
	Run  func([]string, *CommandExecEnv) int
}

func (c *Cmd) Priority() int          { return DefaultCmdPriority }
func (c *Cmd) Match(word string) bool { return word == c.Name }
func (c *Cmd) Execute(args []string, execEnv *CommandExecEnv) (int, error) {
	return c.Run(args, execEnv), nil
}

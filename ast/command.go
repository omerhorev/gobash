package ast

// The Command node represnts a simple command in the form of:
//  - `X=1 ls > 1`
//  - `cat 2>&1`
//  - `<file`
type SimpleCommand struct {
	// Arguments, including command name (arg[0])
	Args []string

	// Stores pre-command assignments
	Assignments map[string]string

	// Stores the command word if provided
	// TODO: Support name
	Word string

	// Stores redirections
	Redirects map[int]*SimpleCommandIORedirection // redirects may
}

func NewSimpleCommand() *SimpleCommand {
	return &SimpleCommand{
		Assignments: make(map[string]string),
		Word:        "",
		Redirects:   make(map[int]*SimpleCommandIORedirection),
		Args:        []string{},
	}
}

func (s *SimpleCommand) AddRedirect(fd int, io *SimpleCommandIORedirection) {
	s.Redirects[fd] = io
}

func (s *SimpleCommand) AddAssignment(key, value string) {
	s.Assignments[key] = value
}

func (s *SimpleCommand) AddArgument(value string) {
	s.Args = append(s.Args, value)
}

// IO redirection mode (controlled by the operator found in the Token)
type SimpleCommandIORedirectionMode string

var (
	SimpleCommandIORedirectionModeOutput       = SimpleCommandIORedirectionMode(">")
	SimpleCommandIORedirectionModeOutputFd     = SimpleCommandIORedirectionMode(">&")
	SimpleCommandIORedirectionModeOutputAppend = SimpleCommandIORedirectionMode(">>")
	SimpleCommandIORedirectionModeInput        = SimpleCommandIORedirectionMode("<")
	SimpleCommandIORedirectionModeInputFd      = SimpleCommandIORedirectionMode("<&")
	SimpleCommandIORedirectionModeInputOutput  = SimpleCommandIORedirectionMode("<>")
	SimpleCommandIORedirectionModeOutputForce  = SimpleCommandIORedirectionMode(">|")
)

type SimpleCommandIORedirection struct {
	Mode  SimpleCommandIORedirectionMode
	Value any // If Fd redirect this value is int, otherwise string
}

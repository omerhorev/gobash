package ast

// The Command node represents a simple command in the form of:
//   - `X=1 ls > 1`
//   - `cat 2>&1`
//   - `<file`
type SimpleCommand struct {
	// Arguments, including command name (arg[0])
	Args []*Expr

	// Stores pre-command assignments
	Assignments map[string]*Expr

	// Stores the command word if provided
	// TODO: Support name
	Word *Expr

	// Stores redirects
	Redirects []*IORedirection // redirects may
}

func NewSimpleCommand() *SimpleCommand {
	return &SimpleCommand{
		Assignments: make(map[string]*Expr),
		Word:        nil,
		Redirects:   []*IORedirection{},
		Args:        []*Expr{},
	}
}

func (s *SimpleCommand) AddRedirect(fd int, io *IORedirection) {
	s.Redirects[fd] = io
}

func (s *SimpleCommand) AddAssignment(key string, value *Expr) {
	s.Assignments[key] = value
}

func (s *SimpleCommand) AddArgument(value *Expr) {
	s.Args = append(s.Args, value)
}

// IO redirection mode (controlled by the operator found in the Token)
type IORedirectionMode string

var (
	IORedirectionModeOutput       = IORedirectionMode(">")
	IORedirectionModeOutputFd     = IORedirectionMode(">&")
	IORedirectionModeOutputAppend = IORedirectionMode(">>")
	IORedirectionModeInput        = IORedirectionMode("<")
	IORedirectionModeInputFd      = IORedirectionMode("<&")
	IORedirectionModeInputOutput  = IORedirectionMode("<>")
	IORedirectionModeOutputForce  = IORedirectionMode(">|")
)

type IORedirection struct {
	Fd    int
	Mode  IORedirectionMode
	Value *Expr
}

// Returns whether the redirection mode is a duplication of another fd
func (m IORedirectionMode) IsDup() bool {
	return m == IORedirectionModeOutputFd ||
		m == IORedirectionModeInputFd
}

// // Returns whether the file needs to be closed: the mode is dup and the value is "-"
// func (io IORedirection) IsClose() bool {
// 	return io.Mode.IsDup() && io.To == "-"
// }

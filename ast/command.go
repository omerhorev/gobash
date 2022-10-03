package ast

type Command interface{}

// can be any
type SimpleCommand struct {
	Assignments map[string]string
	Word        string
	Redirects   []*IORedirect
}

func NewSimpleCommand() *SimpleCommand {
	return &SimpleCommand{
		Assignments: make(map[string]string),
		Word:        "",
		Redirects:   make([]*IORedirect, 0),
	}
}

func (s *SimpleCommand) AddRedirect(io *IORedirect) {
	s.Redirects = append(s.Redirects, io)
}

func (s *SimpleCommand) AddAssignment(key, value string) {
	s.Assignments[key] = value
}

type Assignment struct {
}

type IORedirect struct {
	IO  int
	Dst string
}

type IORedirectFile struct {
	Path string
}

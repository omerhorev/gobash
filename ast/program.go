package ast

// The Program node is the base node. It contains a list of commands to execute.
type Program struct {
	Commands []Node
}

func NewProgram() *Program {
	return &Program{
		Commands: make([]Node, 0),
	}
}

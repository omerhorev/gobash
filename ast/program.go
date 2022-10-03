package ast

type Program struct {
	Commands []Node
}

func NewProgram() *Program {
	return &Program{
		Commands: make([]Node, 0),
	}
}

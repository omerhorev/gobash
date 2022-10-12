package ast

// The Pipe node states that the child nods in Commands should be piped to one another.
// This node is the result of the "|" operator
type Pipe struct {
	Commands []Node
}

func NewPipe() *Pipe {
	return &Pipe{
		Commands: []Node{},
	}
}

func (p *Pipe) AddCommand(cmd Node) {
	p.Commands = append(p.Commands, cmd)
}

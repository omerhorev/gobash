package ast

type Pipe struct {
	Commands   []Command
	LogicalNot bool
}

func NewPipe() *Pipe {
	return &Pipe{
		Commands:   []Command{},
		LogicalNot: false,
	}
}

func (p *Pipe) AddCommand(cmd Command) {
	p.Commands = append(p.Commands, cmd)
}

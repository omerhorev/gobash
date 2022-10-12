package ast

// The Background node states that the child AST should be executed in the background
// as prt of the job control subsystem. It is generated when facing an '&' token at the
// end of a Command (pipe, binary, etc..)
type Background struct {
	Child Node
}

func NewBackground(child Node) *Background {
	return &Background{
		Child: child,
	}
}

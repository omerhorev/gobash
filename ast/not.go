package ast

// Not node states that logical not (!) should be applied to the child AST.
type Not struct {
	Child Node
}

func NewNot(child Node) *Not {
	return &Not{
		Child: child,
	}
}

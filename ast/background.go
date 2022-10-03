package ast

// background states that the following AST
// is executed on the background
type Background struct {
	Node Node
}

func NewBackground(node Node) *Background {
	return &Background{
		Node: node,
	}
}

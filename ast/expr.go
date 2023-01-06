package ast

// The Program node is the base node. It contains a list of commands to execute.
type Expr struct {
	Nodes []Node
}

func NewExpr(nodes ...Node) *Expr {
	return &Expr{
		Nodes: nodes,
	}
}

func NewExprStr(s string) *Expr {
	return NewExpr(NewString(s))
}

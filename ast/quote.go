package ast

type DoubleQuote struct{ Nodes []Node }

func NewDoubleQuote() *DoubleQuote { return &DoubleQuote{Nodes: []Node{}} }

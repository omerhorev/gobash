package ast

type Backtick struct{ Node Node }

func NewBacktick() *Backtick { return &Backtick{Node: nil} }

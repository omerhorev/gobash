package ast

// The Program node is the base node. It contains a list of commands to execute.
type String struct {
	Value string
}

func NewString() *String {
	return &String{
		Value: "",
	}
}

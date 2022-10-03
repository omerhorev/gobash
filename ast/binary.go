package ast

type BinaryType int

var (
	BinaryTypeOr  = BinaryType(0)
	BinaryTypeAnd = BinaryType(1)
)

type Binary struct {
	Left  Node // can be pipe, or another binary
	Right Node // can be pipe, or another binary
	Type  BinaryType
}

func NewBinary(t BinaryType) *Binary {
	return &Binary{
		Left:  nil,
		Right: nil,
		Type:  t,
	}
}

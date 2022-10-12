package ast

type BinaryType int

var (
	BinaryTypeOr  = BinaryType(0)
	BinaryTypeAnd = BinaryType(1)
)

// The Binary node states that the following two nodes should be executed in a Binary Or (||)
// or Binary And (&&) mode.
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

func (b Binary) IsAnd() bool { return b.Type == BinaryTypeAnd }
func (b Binary) IsOr() bool  { return b.Type == BinaryTypeOr }

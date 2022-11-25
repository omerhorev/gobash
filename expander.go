package gobash

import (
	"github.com/omerhorev/gobash/ast"
	"github.com/omerhorev/gobash/rdp"
)

type ExpanderToken rune

func (t ExpanderToken) Accept(token ExpanderToken) bool {
	return t == token
}

func (t ExpanderToken) String() string {
	return string(t)
}

// A helper structure used to expand words
type Expander struct {
	rdp rdp.RDP[ExpanderToken, ExpanderToken]
}

// Creates a new expander object
func NewExpander() *Expander {
	return &Expander{}
}

func Parse() *ast.Node {
	return nil
}

// returns the next instance of an unescaped, unquoted rune in a string

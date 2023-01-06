package gobash

import (
	"unicode/utf8"

	"github.com/omerhorev/gobash/ast"
	"github.com/omerhorev/gobash/rdp"
)

type ExpanderToken rune

func (t ExpanderToken) Accept(token ExpanderToken) bool {
	return t == token
}

func (t ExpanderToken) String() string {
	if t == utf8.MaxRune {
		return "EOF"
	}

	return string(t)
}

var (
	expanderTokenBacktick  = ExpanderToken('`')
	expanderTokenBackslash = ExpanderToken('\\')
	expanderTokenEOF       = ExpanderToken(utf8.MaxRune)
)

// A helper structure used to parse the syntax of word expansion
type Expander struct {
	rdp  rdp.RDP[ExpanderToken, ExpanderToken]
	Expr *ast.Expr
}

// Creates a new expander object
func NewExpander(expression string) *Expander {
	return &Expander{
		rdp: rdp.RDP[ExpanderToken, ExpanderToken]{
			Tokens: append([]ExpanderToken(expression), expanderTokenEOF),
		},
		Expr: nil,
	}
}

func (e *Expander) Parse() error {
	expr := ast.NewExpr()
	for {
		if e.rdp.Error() != nil {
			return e.rdp.Error()
		}

		if e.rdp.Accept(expanderTokenEOF) {
			break
		}

		if node, ok := e.string(); ok {
			expr.Nodes = append(expr.Nodes, node)
			continue
		}

		if node, ok := e.backtick(); ok {
			expr.Nodes = append(expr.Nodes, node)
			continue
		}
	}

	e.Expr = expr

	return e.rdp.Error()
}

func (e *Expander) backtick() (*ast.Backtick, bool) {
	if !e.rdp.Accept(expanderTokenBacktick) {
		return nil, false
	}

	s := ""

	for {
		if r, ok := e.char(); ok {
			s += string(r)
		} else {
			break
		}
	}

	if !e.rdp.Expect(expanderTokenBacktick) {
		return nil, false
	}

	e.rdp.Consume()

	e2 := NewExpander(s)
	if err := e2.Parse(); err != nil {
		e.rdp.SetError(err)
		return nil, false
	}

	node := ast.NewBacktick()
	node.Node = e2.Expr

	return node, true
}

func (e *Expander) string() (*ast.String, bool) {
	s := ast.NewString("")

	for {
		if r, ok := e.char(); ok {
			s.Value += string(r)
		} else {
			break
		}
	}

	if s.Value == "" {
		return nil, false
	}

	return s, true
}

func (e *Expander) char() (rune, bool) {
	if r, ok := e.backslash(); ok {
		return r, true
	} else if e.checkNotSpecial() {
		currentRune := rune(e.rdp.Current())
		e.rdp.Consume()
		return currentRune, true
	}

	return utf8.RuneError, false
}

func (e *Expander) backslash() (rune, bool) {
	if e.rdp.Accept(expanderTokenBackslash) {
		if e.rdp.Check(expanderTokenEOF) {
			return utf8.RuneError, false
		}

		currentRune := rune(e.rdp.Current())
		e.rdp.Consume()
		return currentRune, true
	}

	return utf8.RuneError, false
}

func (e *Expander) checkNotSpecial() bool {
	return !e.rdp.Check(expanderTokenBacktick, expanderTokenEOF)
}

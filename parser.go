package gobash

import (
	"strconv"
	"strings"

	"github.com/omerhorev/gobash/ast"
	"github.com/omerhorev/gobash/rdp"
)

// Represents settings for the Parser
type ParserSettings struct{}

var parserDefaultSettings = ParserSettings{}

// The Parser transforms tokens into Abstract Syntax Tree (AST).
type Parser struct {
	Settings ParserSettings

	rdp  rdp.RDP[*Token, TokenIdentifier]
	node ast.Node
}

// Creates a new parser with settings.
func NewParser(tokens []*Token, settings ParserSettings) *Parser {
	return &Parser{
		rdp: rdp.RDP[*Token, TokenIdentifier]{
			Tokens: tokens,
		},

		Settings: settings,
		node:     nil,
	}
}

// Creates a new parser object with default settings.
func NewParserDefault(tokens []*Token) *Parser {
	return NewParser(tokens, parserDefaultSettings)
}

// Start parsing the tokens. This method is not thread safe and can
// be executed only once for the Parser object.
func (p *Parser) Parse() error {
	return p.parse()
}

// Return errors from the parsing process.
func (p *Parser) Error() error {
	if rdp.IsSyntaxError(p.rdp.Error()) {
		return newSyntaxError(p.rdp.Error().(rdp.SyntaxError).Unwrap())
	}
	return p.rdp.Error()
}

// Returns the generated AST as a program node to be used by the executor
// This method can be used only after calling the Parse method
func (p *Parser) Program() *ast.Program {
	return p.AST().(*ast.Program)
}

// Return the generated AST after the parse. This method can be used only after calling
// the Parse method
func (p *Parser) AST() ast.Node {
	return p.node
}

// func (p *Parser) error(format string, args ...any) {
// 	p.rdp.Error() = newSyntaxError(errors.Errorf(format, args...))
// }

// func (p *Parser) current() *Token {
// 	return p.tokens[p.rdp.CurrentIndex]
// }

// func (p *Parser) prev() *Token {
// 	return p.tokens[p.rdp.CurrentIndex-1]
// }

// func (p *Parser) consume() {
// 	p.rdp.CurrentIndex++
// }

// func (p *Parser) backup() int {
// 	return p.rdp.CurrentIndex
// }

// func (p *Parser) restore(index int) {
// 	p.rdp.CurrentIndex = index
// }

// func (p *Parser) check(terminal ...TokenIdentifier) bool {
// 	for _, t := range terminal {
// 		if p.rdp.Current().Identifier == t {
// 			return true
// 		}
// 	}

// 	return false
// }

// func (p *Parser) accept(terminal ...TokenIdentifier) bool {
// 	if p.rdp.Error() != nil {
// 		return false
// 	}

// 	if p.check(terminal...) {
// 		p.rdp.Consume()
// 		return true
// 	}

// 	return false
// }

// func (p *Parser) expect(identifier TokenIdentifier) bool {
// 	if !p.rdp.Accept(identifier) {
// 		found := p.rdp.Current().Identifier
// 		if p.rdp.Current().Identifier == tokenIdentifierWord ||
// 			p.rdp.Current().Identifier == tokenIdentifierAssignmentWord {
// 			found = TokenIdentifier(p.rdp.Current().Value)
// 		}

// 		p.rdp.SetError(fmt.Errorf("expected %s but found %s", identifier, found))
// 		return false
// 	}

// 	return true
// }

func (p *Parser) parse() error {
	b := p.rdp.Backup()

	program, _ := p.program()

	if !p.rdp.Expect(tokenIdentifierEOF) {
		p.rdp.Restore(b)

		return p.rdp.Error()
	}

	p.node = program

	return nil
}

func (p *Parser) program() (*ast.Program, bool) {
	p.linebreak()

	b := p.rdp.Backup()
	program := ast.NewProgram()

	if nodes, ok := p.completeCommands(); !ok {
		p.rdp.Restore(b)
	} else {
		program.Commands = nodes
	}

	return program, true
}

func (p *Parser) completeCommands() (nodes []ast.Node, ok bool) {
	b := p.rdp.Backup()

	nodes, ok2 := p.completeCommand()
	if !ok2 {
		p.rdp.Restore(b)
		return nil, false
	}

	p.linebreak()

	nodes2, ok2 := p.completeCommands()
	if ok2 {
		nodes = append(nodes, nodes2...)
	}

	return nodes, true
}

// this is a merge between complete_command and list
func (p *Parser) completeCommand() ([]ast.Node, bool) {
	b := p.rdp.Backup()

	node, ok := p.andOr()
	if !ok {
		p.rdp.Restore(b)
		return nil, false
	}

	if p.rdp.Accept(tokenIdentifierSemicolon) {
		// do nothing, just a semicolon
	} else {
		if p.rdp.Accept(tokenIdentifierAnd) {
			node = ast.NewBackground(node)
		}
	}

	nodes := []ast.Node{node}

	if nodes2, ok := p.completeCommand(); ok {
		nodes = append(nodes, nodes2...)
	}

	return nodes, true
}

func (p *Parser) andOr() (node ast.Node, ok bool) {
	b := p.rdp.Backup()

	pipe, found := p.pipeline()
	if !found {
		p.rdp.Restore(b)
		return nil, false
	}

	binaryType, found := p.andOrBinaryType()
	if !found {
		return pipe, true
	}

	p.linebreak()

	next, ok2 := p.andOr()
	if !ok2 {
		return nil, false
	}

	binary := ast.NewBinary(binaryType)
	binary.Left = pipe
	binary.Right = next

	return binary, true
}

// derived from the and_or grammar rule
func (p *Parser) andOrBinaryType() (BinaryType ast.BinaryType, ok bool) {
	if p.rdp.Accept(tokenIdentifierDAnd) {
		return ast.BinaryTypeAnd, true
	}

	if p.rdp.Accept(tokenIdentifierDPipe) {
		return ast.BinaryTypeOr, true
	}

	return 0, false
}

// this represents both the pipeline and pipeline_sequence syntax because
// it can be simplified if not using a recursion
func (p *Parser) pipeline() (ast.Node, bool) {
	b := p.rdp.Backup()

	pipe := ast.NewPipe()

	cmd, ok2 := p.command()
	if !ok2 {
		p.rdp.Restore(b)
		return nil, false
	}

	pipe.AddCommand(cmd)

	// don't use recursion because we don't want more nodes in the AST. instead
	// just loop through the pipes
	for {
		b2 := p.rdp.Backup()
		if !p.rdp.Accept(tokenIdentifierPipe) {
			p.rdp.Restore(b2)
			break
		}

		cmd, ok2 := p.command()
		if !ok2 {
			p.rdp.Restore(b2)
			break
		}

		pipe.AddCommand(cmd)
	}

	node := ast.Node(pipe)
	if len(pipe.Commands) == 1 {
		node = pipe.Commands[0]
	}

	return node, true
}

func (p *Parser) command() (ast.Node, bool) {
	// the first word of a command
	b := p.rdp.Backup()
	upgraded := p.rdp.Current().tryUpgradeToReservedWord()

	not := p.rdp.Accept(tokenIdentifierBang)

	cmdNode, ok := p.simpleCommand()
	if !ok {
		p.rdp.Restore(b)
		// restore the upgraded token, if it was upgraded
		if upgraded {
			p.rdp.Current().Identifier = tokenIdentifierWord
		}

		return nil, false
	}

	node := ast.Node(cmdNode)
	if not {
		node = ast.NewNot(cmdNode)
	}

	return node, true
}

func (p *Parser) simpleCommand() (node *ast.SimpleCommand, ok bool) {
	c := ast.NewSimpleCommand()

	if p.cmdPrefix(c) {
		if p.cmdWord(c) {
			p.cmdSuffix(c)
		}
	} else if p.cmdName(c) {
		p.cmdSuffix(c)
	} else {
		return nil, false
	}

	return c, true
}

func (p *Parser) cmdPrefix(cmd *ast.SimpleCommand) (ok bool) {
	b := p.rdp.Backup()

	if !p.ioRedirect(cmd) {
		if k, word, ok := p.assignmentWord(); !ok {
			p.rdp.Restore(b)
			return false
		} else {
			cmd.AddAssignment(k, word)
		}
	}

	p.cmdPrefix(cmd)

	return true
}

func (p *Parser) cmdWord(cmd *ast.SimpleCommand) bool {
	if !p.rdp.Check(tokenIdentifierWord) {
		return false
	}

	e := NewExpander(p.rdp.Current().Value)
	if err := e.Parse(); err != nil {
		return false
	}

	cmd.Word = e.Expr
	p.rdp.Consume()

	return true
}

func (p *Parser) cmdSuffix(cmd *ast.SimpleCommand) (ok bool) {
	b := p.rdp.Backup()

	if !p.ioRedirectOrArg(cmd) {
		p.rdp.Restore(b)
		return false
	}

	p.cmdSuffix(cmd)

	return true
}

func (p *Parser) ioRedirectOrArg(cmd *ast.SimpleCommand) bool {
	b := p.rdp.Backup()

	if p.ioRedirect(cmd) {
		return true
	}

	if p.rdp.Accept(tokenIdentifierWord) {
		e := NewExpander(p.rdp.Prev().Value)
		if err := e.Parse(); err != nil {
			return false
		}

		cmd.AddArgument(e.Expr)
		return true
	}

	p.rdp.Restore(b)
	return false
}

func (p *Parser) cmdName(cmd *ast.SimpleCommand) bool {
	e := NewExpander(p.rdp.Current().Value)
	if err := e.Parse(); err != nil {
		return false
	}

	cmd.Word = e.Expr

	return p.rdp.Accept(tokenIdentifierWord)
}

func (p *Parser) ioRedirect(cmd *ast.SimpleCommand) (ok bool) {
	b := p.rdp.Backup()

	a := &ast.IORedirection{}
	var fdSet *int = nil

	if p.rdp.Accept(tokenIdentifierIONumber) {
		fd, _ := strconv.Atoi(p.rdp.Prev().Value)
		fdSet = &fd
	}

	fdAsserted := 0
	if p.rdp.Accept(tokenIdentifierLess, tokenIdentifierLessAnd) { // <
		fdAsserted = 0
	} else if p.rdp.Accept(tokenIdentifierGreat, tokenIdentifierDGreat, tokenIdentifierGreatAnd, tokenIdentifierClobber) {
		fdAsserted = 1
	} else if p.rdp.Accept(tokenIdentifierLessGreat) {
		// nothing
	} else {
		p.rdp.Restore(b)
		return false
	}

	a.Mode = ast.IORedirectionMode(p.rdp.Prev().Value)

	a.Value = ast.NewExprStr(p.rdp.Current().Value)

	to := fdAsserted
	if fdSet != nil {
		to = *fdSet
	}

	a.Fd = to
	cmd.Redirects = append(cmd.Redirects, a)

	p.rdp.Consume()

	return true
}

func (p *Parser) linebreak() bool {
	p.newlineList()

	return true
}

func (p *Parser) newlineList() bool {
	// read at least one newline
	if !p.rdp.Accept(tokenIdentifierNewline) {
		return false
	}

	for {
		if !p.rdp.Accept(tokenIdentifierNewline) {
			break
		}
	}

	return true
}

// in the grammar this is a terminal. it will be represented
// here as a non-terminal to parse context depended information
func (p *Parser) assignmentWord() (string, *ast.Expr, bool) {
	if p.rdp.Current().tryUpgradeToAssignmentWord() { // rule 7b
		v := p.rdp.Current().Value
		i := strings.IndexRune(v, '=')
		key := v[:i]
		value := v[i+1:]
		p.rdp.Consume()

		return key, ast.NewExprStr(value), true
	}

	return "", nil, false
}

// func (p *Parser) expandAll() (ast.Node, bool) {
// 	b := p.rdp.Backup()

// 	t := NewTokenizerShort(p.rdp.Current().Value)

// 	tokens, err := t.ReadAll()
// 	if err != nil {
// 		p.rdp.Restore(b)
// 		return nil, false
// 	}

// 	p2 := NewParser(tokens, p.Settings)
// 	p2.Parse()

// 	// b := p.rdp.Backup()

// 	// TODO: expansion

// 	n := ast.NewString()

// 	nextIsBackslash := false

// 	v := p.rdp.Current().Value
// 	for _, r := range v {
// 		if nextIsBackslash {
// 			if r == 'n' {
// 				r = '\n'
// 			}

// 			if r == 'r' {
// 				r = '\r'
// 			}

// 			if r == ' ' {
// 				r = ' '
// 			}

// 			nextIsBackslash = false
// 		}

// 		if isBackslash(r) {
// 			nextIsBackslash = true
// 		} else {
// 			n.Value = n.Value + string(r)
// 		}
// 	}

// 	// p.rdp.Restore(b)

// 	return n, true
// }

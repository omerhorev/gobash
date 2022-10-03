package gobash

import (
	"strconv"
	"strings"

	"github.com/omerhorev/gobash/ast"
	"github.com/pkg/errors"
)

// Represents settings for Parser
type ParserSettings struct{}

var parserDefaultSettings = ParserSettings{}

// Transforms tokens into abstract syntax tree (AST)
type Parser struct {
	tokens       []*Token
	currentIndex int
	Settings     ParserSettings
	err          error
	node         ast.Node
}

// Creates a new parser with settings
func NewParser(tokens []*Token, settings ParserSettings) *Parser {
	return &Parser{
		tokens:       tokens,
		currentIndex: 0,
		Settings:     settings,
		err:          nil,
		node:         nil,
	}
}

// Creates a new parser object with default settings
func NewParserDefault(tokens []*Token) *Parser {
	return NewParser(tokens, parserDefaultSettings)
}

// Start parsing the tokens. This method is not thread safe and should
// be executed only once for the Parser object.
//
// Return any syntax error that rise durring the parsing.
func (p *Parser) Parse() error {
	return p.parse()
}

// Return any error that rised during the parsing
func (p *Parser) Error() error {
	return p.err
}

// Return the generated AST after the parse. This method should be used after calling
// the Parse() method
func (p *Parser) AST() ast.Node {
	return p.node
}

func (p *Parser) error(format string, args ...any) {
	p.err = NewSyntaxError(errors.Errorf(format, args...))
}

func (p *Parser) current() *Token {
	return p.tokens[p.currentIndex]
}

func (p *Parser) consume() {
	p.currentIndex++
}

func (p *Parser) backup() int {
	return p.currentIndex
}

func (p *Parser) restore(index int) {
	p.currentIndex = index
}

func (p *Parser) check(terminal ...TokenIdentifier) bool {
	for _, t := range terminal {
		if p.current().Identifier == t {
			return true
		}
	}

	return false
}

func (p *Parser) accept(terminal ...TokenIdentifier) bool {
	if p.err != nil {
		return false
	}

	if p.check(terminal...) {
		p.consume()
		return true
	}

	return false
}

func (p *Parser) expect(identifier TokenIdentifier) bool {
	if !p.accept(identifier) {
		found := p.current().Identifier
		if p.current().Identifier == tokenIdentifierWord ||
			p.current().Identifier == tokenIdentifierAssignmentWord {
			found = TokenIdentifier(p.current().Value)
		}

		p.error("expected %s but found %s", identifier, found)
		return false
	}

	return true
}

func (p *Parser) parse() error {
	b := p.backup()
	program, ok := p.program()

	if !ok {
		if p.err != nil {
			return p.err
		} else {
			p.error("syntax error: unexpected %s", p.current().Identifier)
			return p.err
		}
	}

	if !p.expect(tokenIdentifierEOF) {
		p.restore(b)

		return p.err
	}

	p.node = program

	return nil
}

func (p *Parser) program() (program *ast.Program, ok bool) {
	p.linebreak()

	b := p.backup()
	program = ast.NewProgram()

	nodes, ok2 := p.completeCommands()
	if !ok2 {
		p.restore(b)

		return nil, false
	}

	program.Commands = nodes

	return program, true
}

func (p *Parser) completeCommands() (nodes []ast.Node, ok bool) {
	b := p.backup()

	nodes, ok2 := p.completeCommand()
	if !ok2 {
		p.restore(b)
		return nil, false
	}

	p.linebreak()

	nodes2, ok2 := p.completeCommands()
	if ok2 {
		nodes = append(nodes, nodes2...)
	}

	return nodes, true
}

// this is a join between complete_command and list
func (p *Parser) completeCommand() ([]ast.Node, bool) {
	b := p.backup()

	node, ok := p.andOr()
	if !ok {
		p.restore(b)
		return nil, false
	}

	if p.accept(tokenIdentifierSemicolon) {
		// do nothing, just a semicolon
	} else {
		if p.accept(tokenIdentifierAnd) {
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
	b := p.backup()

	pipe, found := p.pipeline()
	if !found {
		p.restore(b)
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

// derived from the and_or grammer rule
func (p *Parser) andOrBinaryType() (BinaryType ast.BinaryType, ok bool) {
	if p.accept(tokenIdentifierDAnd) {
		return ast.BinaryTypeAnd, true
	}

	if p.accept(tokenIdentifierDPipe) {
		return ast.BinaryTypeOr, true
	}

	return 0, false
}

// this represents both the pipeline and pipeline_sequence syntax because
// it can be simplefied if not using a recursion
func (p *Parser) pipeline() (pipe *ast.Pipe, ok bool) {
	b := p.backup()

	pipe = ast.NewPipe()

	if p.accept(tokenIdentifierBang) {
		pipe.LogicalNot = true
	}

	cmd, ok2 := p.command()
	if !ok2 {
		p.restore(b)
		return nil, false
	}

	pipe.AddCommand(cmd)

	// don't use recursion because we don't want more nodes in the AST. instead
	// just loop through the pipes
	for {
		b2 := p.backup()
		if !p.accept(tokenIdentifierPipe) {
			p.restore(b2)
			break
		}

		cmd, ok2 := p.command()
		if !ok2 {
			p.restore(b2)
			break
		}

		pipe.AddCommand(cmd)
	}

	return pipe, true
}

func (p *Parser) command() (ast.Command, bool) {
	return p.simpleCommand()

	// TODO: Implement more command types
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
	b := p.backup()

	if r, ok := p.ioRedirect(); !ok {
		if k, v, ok := p.assignmentWord(); !ok {
			p.restore(b)
			return false
		} else {
			cmd.AddAssignment(k, v)
		}
	} else {
		cmd.AddRedirect(r)
	}

	p.cmdPrefix(cmd)

	return true
}

func (p *Parser) cmdWord(cmd *ast.SimpleCommand) bool {
	if p.current().tryUpgradeToAssignmentWord() {
		return false
	}

	if p.check(tokenIdentifierWord) {
		return false
	}

	cmd.Word = p.current().Value
	p.consume()

	return true
}

func (p *Parser) cmdSuffix(cmd *ast.SimpleCommand) (ok bool) {
	b := p.backup()

	redirect, ok2 := p.ioRedirect()
	if !ok2 {
		p.restore(b)
		return false
	}
	cmd.AddRedirect(redirect)

	for {
		b2 := p.backup()
		redirect, ok2 := p.ioRedirect()
		if !ok2 {
			p.restore(b2)
			break
		}
		cmd.AddRedirect(redirect)
	}

	return true
}

func (p *Parser) cmdName(cmd *ast.SimpleCommand) bool {
	// TODO:

	cmd.Word = p.current().Value

	return p.accept(tokenIdentifierWord)
}

func (p *Parser) ioRedirect() (node *ast.IORedirect, ok bool) {
	node = &ast.IORedirect{IO: 1}

	if p.accept(tokenIdentifierIONumber) {
		n, _ := strconv.Atoi(p.current().Value)
		node.IO = n
	}

	file, ok := p.ioFile()
	if !ok {
		return nil, false
	}

	node.Dst = file

	return node, true
}

func (p *Parser) ioFile() (path string, ok bool) {
	b := p.backup()

	if !p.accept(tokenIdentifierLess) &&
		!p.accept(tokenIdentifierGreat) &&
		!p.accept(tokenIdentifierDGreat) &&
		!p.accept(tokenIdentifierLessAnd) &&
		!p.accept(tokenIdentifierLessGreat) &&
		!p.accept(tokenIdentifierClobber) {
		p.restore(b)
		return "", false
	}

	p.accept(tokenIdentifierWord)

	return p.current().Value, true
}

func (p *Parser) linebreak() bool {
	p.newlineList()

	return true
}

func (s *Parser) newlineList() bool {
	// read at least one newline
	if !s.accept(tokenIdentifierNewline) {
		return false
	}

	for {
		if !s.accept(tokenIdentifierNewline) {
			break
		}
	}

	return true
}

// in the grammer this is a terminal. it will be represented
// here as a non-terminal to parse context depended information
func (p *Parser) assignmentWord() (string, string, bool) {
	if p.current().tryUpgradeToAssignmentWord() { // rule 7b
		v := p.current().Value
		i := strings.IndexRune(v, '=')
		key := v[:i]
		value := v[i:]
		p.consume()

		return key, value, true
	}

	return "", "", false
}

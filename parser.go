package gobash

import (
	"strconv"
	"strings"

	"github.com/omerhorev/gobash/ast"
	"github.com/pkg/errors"
)

// Represents settings for the Parser
type ParserSettings struct{}

var parserDefaultSettings = ParserSettings{}

// The Parser transforms tokens into Abstract Syntax Tree (AST).
type Parser struct {
	Settings ParserSettings

	tokens       []*Token
	currentIndex int
	err          error
	node         ast.Node
}

// Creates a new parser with settings.
func NewParser(tokens []*Token, settings ParserSettings) *Parser {
	return &Parser{
		tokens:       tokens,
		currentIndex: 0,
		Settings:     settings,
		err:          nil,
		node:         nil,
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
	return p.err
}

// Return the generated AST after the parse. This method can be used only after calling
// the Parse method
func (p *Parser) AST() ast.Node {
	return p.node
}

func (p *Parser) error(format string, args ...any) {
	p.err = newSyntaxError(errors.Errorf(format, args...))
}

func (p *Parser) current() *Token {
	return p.tokens[p.currentIndex]
}

func (p *Parser) prev() *Token {
	return p.tokens[p.currentIndex-1]
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

	program, _ := p.program()

	if !p.expect(tokenIdentifierEOF) {
		p.restore(b)

		return p.err
	}

	p.node = program

	return nil
}

func (p *Parser) program() (*ast.Program, bool) {
	p.linebreak()

	b := p.backup()
	program := ast.NewProgram()

	if nodes, ok := p.completeCommands(); !ok {
		p.restore(b)
	} else {
		program.Commands = nodes
	}

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

// this is a merge between complete_command and list
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

// derived from the and_or grammar rule
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
// it can be simplified if not using a recursion
func (p *Parser) pipeline() (ast.Node, bool) {
	b := p.backup()

	pipe := ast.NewPipe()

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

	node := ast.Node(pipe)
	if len(pipe.Commands) == 1 {
		node = pipe.Commands[0]
	}

	return node, true
}

func (p *Parser) command() (ast.Node, bool) {
	// the first word of a command
	b := p.backup()
	upgraded := p.current().tryUpgradeToReservedWord()

	not := p.accept(tokenIdentifierBang)

	cmdNode, ok := p.simpleCommand()
	if !ok {
		p.restore(b)
		// restore the upgraded token, if it was upgraded
		if upgraded {
			p.current().Identifier = tokenIdentifierWord
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
	b := p.backup()

	if !p.ioRedirect(cmd) {
		if k, v, ok := p.assignmentWord(); !ok {
			p.restore(b)
			return false
		} else {
			cmd.AddAssignment(k, v)
		}
	}

	p.cmdPrefix(cmd)

	return true
}

func (p *Parser) cmdWord(cmd *ast.SimpleCommand) bool {
	if !p.check(tokenIdentifierWord) {
		return false
	}

	cmd.Word = p.current().Value
	p.consume()

	return true
}

func (p *Parser) cmdSuffix(cmd *ast.SimpleCommand) (ok bool) {
	b := p.backup()

	if !p.ioRedirectOrArg(cmd) {
		p.restore(b)
		return false
	}

	p.cmdSuffix(cmd)

	return true
}

func (p *Parser) ioRedirectOrArg(cmd *ast.SimpleCommand) bool {
	b := p.backup()

	if p.ioRedirect(cmd) {
		return true
	}

	if p.accept(tokenIdentifierWord) {
		cmd.AddArgument(p.prev().Value)
		return true
	}

	p.restore(b)
	return false
}

func (p *Parser) cmdName(cmd *ast.SimpleCommand) bool {
	// TODO:

	cmd.Word = p.current().Value

	return p.accept(tokenIdentifierWord)
}

func (p *Parser) ioRedirect(cmd *ast.SimpleCommand) (ok bool) {
	b := p.backup()

	a := &ast.IORedirection{}
	var fdSet *int = nil

	if p.accept(tokenIdentifierIONumber) {
		fd, _ := strconv.Atoi(p.prev().Value)
		fdSet = &fd
	}

	fdAsserted := 0
	if p.accept(tokenIdentifierLess, tokenIdentifierLessAnd) { // <
		fdAsserted = 0
	} else if p.accept(tokenIdentifierGreat, tokenIdentifierDGreat, tokenIdentifierGreatAnd, tokenIdentifierClobber) {
		fdAsserted = 1
	} else if p.accept(tokenIdentifierLessGreat) {
		// nothing
	} else {
		p.restore(b)
		return false
	}

	a.Mode = ast.IORedirectionMode(p.prev().Value)

	if a.Mode == ast.IORedirectionModeInputFd ||
		a.Mode == ast.IORedirectionModeOutputFd {
		if fd, err := strconv.Atoi(p.current().Value); err != nil {
			p.error("bad fd number")
			p.restore(b)
			return false
		} else {
			a.Value = fd
		}
	} else {
		a.Value = p.current().Value
	}

	to := fdAsserted
	if fdSet != nil {
		to = *fdSet
	}

	a.Fd = to
	cmd.Redirects = append(cmd.Redirects, a)

	p.consume()

	return true
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

// in the grammar this is a terminal. it will be represented
// here as a non-terminal to parse context depended information
func (p *Parser) assignmentWord() (string, string, bool) {
	if p.current().tryUpgradeToAssignmentWord() { // rule 7b
		v := p.current().Value
		i := strings.IndexRune(v, '=')
		key := v[:i]
		value := v[i+1:]
		p.consume()

		return key, value, true
	}

	return "", "", false
}

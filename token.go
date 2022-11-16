package gobash

import "unicode/utf8"

var (
	operatorsStrings = []string{
		"&&", "||", ";;", "<<", ">>", "<&", ">&", "<>", "<<-", ">|",
		"&", "|", ";", "<", ">",
	}

	reservedWordsStrings = []string{
		"!", "{", "}", "case", "do", "done", "elif", "else",
		"esac", "fi", "for", "if", "in", "then", "until", "while",
	}
)

type TokenIdentifier string

var (
	tokenIdentifierEOF            = TokenIdentifier("<eof>")
	tokenIdentifierWord           = TokenIdentifier("<word>")
	tokenIdentifierAssignmentWord = TokenIdentifier("<assignment_word>")
	tokenIdentifierIONumber       = TokenIdentifier("<io>")
	tokenIdentifierNewline        = TokenIdentifier("\n")
	tokenIdentifierAnd            = TokenIdentifier("&")
	tokenIdentifierSemicolon      = TokenIdentifier(";")
	tokenIdentifierDAnd           = TokenIdentifier("&&")
	tokenIdentifierDPipe          = TokenIdentifier("||")
	tokenIdentifierBang           = TokenIdentifier("!")
	tokenIdentifierPipe           = TokenIdentifier("|")
	tokenIdentifierLess           = TokenIdentifier("<")
	tokenIdentifierLessAnd        = TokenIdentifier("<&")
	tokenIdentifierGreat          = TokenIdentifier(">")
	tokenIdentifierGreatAnd       = TokenIdentifier(">&")
	tokenIdentifierDGreat         = TokenIdentifier(">>")
	tokenIdentifierLessGreat      = TokenIdentifier("<>")
	tokenIdentifierClobber        = TokenIdentifier(">|")
)

// The Token Represents a part of the string produced by the Tokenizer.
// The Token is used in the grammatical processing of shell expressions and is later
// transformed into AST by the Parser.
type Token struct {
	Value      string          // The actual value of the token
	Identifier TokenIdentifier // The type of token
}

// Returns whether a token is of specific type.
func (t Token) Is(identifier TokenIdentifier) bool {
	return t.Identifier == identifier
}

// Returns wether the token is EOF token.
func (t Token) IsEOF() bool {
	return t.Is(tokenIdentifierEOF)
}

// Returns whether the token is all numbers.
func (t Token) IsAllNumbers() bool {
	return isStringNumber(t.Value)
}

// Assignment-Word tokens are context-dependent. Try upgrading this token to
// assignment-word if possible. Returns whether the upgrade was successful or not.
func (t *Token) tryUpgradeToAssignmentWord() bool {
	if t.Identifier != tokenIdentifierWord {
		return false
	}

	// an assignment must be at least 2 runes (X=)
	if len(t.Value) < 1 {
		return false
	}

	r, _ := utf8.DecodeRuneInString(t.Value)
	if !isNameRune(r) || isDigit(r) {
		return false
	}

	for _, r := range t.Value {
		if !isNameRune(r) {
			if isEqualSign(r) {
				t.Identifier = tokenIdentifierAssignmentWord
				return true
			}
		}
	}

	return false
}

// Reserved-Word tokens are context-dependent. Try upgrading this token to
// Reserved-word if possible. Returns whether the upgrade was successful or not.
func (t *Token) tryUpgradeToReservedWord() bool {
	if t.Identifier != tokenIdentifierWord {
		return false
	}

	for _, r := range reservedWordsStrings {
		if r == t.Value {
			t.Identifier = TokenIdentifier(r)
		}
	}
	if t.Value == "!" {
		t.Identifier = tokenIdentifierBang

		return true
	}

	return false
}

func newTokenFromString(value string, delimitedBy rune) *Token {
	// apply 2.10.1 Shell Grammar Lexical Conventions

	identifier := tokenIdentifierWord

	for _, op := range operatorsStrings {
		if value == op {
			identifier = TokenIdentifier(op)
			break
		}
	}

	// rule #2 - all number string delimited by < or >
	if isStringNumber(value) && (delimitedBy == '<' || delimitedBy == '>') {
		identifier = tokenIdentifierIONumber
	}

	// implied from standard, not stated
	if r, _ := utf8.DecodeRuneInString(value); isNewLine(r) && len(value) == 1 {
		identifier = tokenIdentifierNewline
	}

	// implied from standard, not stated
	if value == "" {
		identifier = tokenIdentifierEOF
	}

	return &Token{
		Value:      value,
		Identifier: identifier,
	}
}

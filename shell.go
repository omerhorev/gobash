package gobash

import (
	"context"
	"io"
	"strings"
	"unicode"

	"github.com/pkg/errors"
)

type Shell struct {
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader
}

func NewShell(stdin io.Reader, stdout io.Writer, stderr io.Writer) *Shell {
	return &Shell{Stdin: stdin, Stdout: stdout, Stderr: stderr}
}

// Just like RunContext be receives a string instead of a reader
func (s *Shell) RunStringContext(ctx context.Context, str string) error {
	return s.RunContext(ctx, strings.NewReader(str))
}

// Run the shell with the current environment and settings, while evaluating
// the miltu-line script in r.
//
// The context can be used to terminate the executaion of the entire shell,
func (s *Shell) RunContext(ctx context.Context, r io.Reader) error {
	lr := NewLineReader(r)
	lineNumber := 0

	for {
		lineNumber++
		line, err := lr.ReadLine()
		if errors.Is(err, io.EOF) {
			if line == "" {
				break
			}
		} else if err != nil {
			return errors.Wrap(err, "readline")
		}

		if err := s.processLine(line); err != nil {
			if IsSyntaxError(err) {
				se := err.(SyntaxError)
				se.SetLine(lineNumber)
				return se
			} else {
				return err
			}
		}
	}

	return nil
}

func (s *Shell) tokenizeLine(line string) ([]string, error) {
	// Tokenize the expression
	tokens := []string{}
	isInOperator := false
	isBackslashed := false
	isApostrophed := false
	isQuotationMarked := false

	preserveMeaning := func() bool {
		return isBackslashed || isApostrophed || isQuotationMarked
	}

	token := ""

	delim := func() {
		if token != "" {
			tokens = append(tokens, token)
		}

		token = ""
		isInOperator = false
		isBackslashed = false
	}

	for _, r := range line {
		if !preserveMeaning() && isInOperator {
			if canBeUsedInOperator(token + string(r)) {
				// rule #2 - the next charected can be used in an operator
				// add()
				token += string(r)
				continue
			} else {
				// rule #3 - the next charected cannot be used in an operator
				delim()
			}
		}

		// rule #4 - backslash
		// backslash is not applied when already backslashing or inside a double quote
		if !(isBackslashed || isApostrophed) && isEscape(r) {
			isBackslashed = true
			token += string(r)
			continue
		}

		// rule #4
		if !preserveMeaning() && !isQuotationMarked && isQuotationMark(r) {
			isQuotationMarked = true
			token += string(r)
			continue
		}

		if !isBackslashed && isQuotationMarked && isQuotationMark(r) {
			isQuotationMarked = false
			token += string(r)
			continue
		}

		if !preserveMeaning() && !isApostrophed && isApostrophe(r) {
			isApostrophed = true
			token += string(r)
			continue
		}

		if isApostrophed && isApostrophe(r) {
			isApostrophed = false
			token += string(r)
			continue
		}

		// rule #6, new operator
		if !preserveMeaning() && canBeUsedInOperator(string(r)) {
			delim()
			// add()
			token += string(r)
			isInOperator = true
			continue
		}

		// rule 7#, spaces
		if !preserveMeaning() && unicode.IsSpace(r) {
			delim()
			// discard()
			continue
		}

		// rule #8, comment
		if !preserveMeaning() && isNumberSign(r) {
			break
		}

		// add to word
		isBackslashed = false
		token += string(r)
	}

	if isApostrophed || isQuotationMarked {
		return nil, NewSyntaxErrorWithOffset(ErrUnterminatedQuotedString, len(line))
	}

	// rule #1
	delim()

	return tokens, nil
}

func (s *Shell) processLine(line string) error {
	_, err := s.tokenizeLine(line)
	if err != nil {
		return err
	}

	return nil
}

func canBeUsedInOperator(token string) bool {
	operators := []string{"<", "<<", "<&", ">&", "<>", "<<-", ">|", ">", ">>", "&&", "||", ";", "&"}

	for _, o := range operators {
		if strings.HasPrefix(o, token) {
			return true
		}
	}

	return false
}

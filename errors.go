package gobash

import (
	"errors"
	"fmt"
)

var ErrUnterminatedQuotedString = errors.New("unterminated quoted string")

// An syntax error. The error contain the cause (unterminated quote, etc) and the
// location of the error.
//
// The location may contain the offset of the error inside the line
// and the line number
type SyntaxError struct {
	Err    error
	offset *int // optional, may add a char offset
	line   *int // optional, may add a line
}

// Creates a new syntax error from an existing error
func NewSyntaxError(err error) SyntaxError {
	return SyntaxError{
		Err:    err,
		offset: nil,
		line:   nil,
	}
}

// Creates a new syntax error from an existing error
func NewSyntaxErrorWithOffset(err error, offset int) SyntaxError {
	return SyntaxError{
		Err:    err,
		offset: &offset,
		line:   nil,
	}
}

// Returns whether the error provided is a syntax error
func IsSyntaxError(err error) bool {
	return errors.Is(err, SyntaxError{})
}

// Returns whether the error has offset data available
func (err SyntaxError) HasOffset() bool { return err.offset != nil }

// Returns the offset of the error inside the line. panic if no offset data
// is available. Check first with HasOffset()
func (err SyntaxError) Offset() int { return *err.offset }

// Sets the offset the error happened inside the line
func (err *SyntaxError) SetOffset(off int) { err.offset = &off }

// Returns whether the error has line data available
func (err SyntaxError) HasLine() bool { return err.line != nil }

// Returns the line of the error inside the script. panic if no line data
// is available. Check first with HasLine()
func (err SyntaxError) Line() int { return *err.line }

// Sets the line the error happened
func (err *SyntaxError) SetLine(line int) { err.line = &line }

func (err SyntaxError) Error() (description string) {
	if err.HasOffset() && err.HasLine() {
		return fmt.Sprintf("syntax error at %d:%d", err.Line(), err.Offset())
	} else if !err.HasOffset() && err.HasLine() {
		return fmt.Sprintf("syntax error at line %d", err.Line())
	} else if err.HasOffset() && !err.HasLine() {
		return fmt.Sprintf("syntax error at offset %d", err.Offset())
	} else {
		return "syntax error"
	}
}

func (err SyntaxError) Unwrap() error {
	return err.Err
}

func (err SyntaxError) Is(err2 error) bool {
	_, ok := err2.(SyntaxError)
	return ok
}

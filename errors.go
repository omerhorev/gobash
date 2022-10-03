package gobash

import (
	"errors"
	"fmt"
)

var ErrUnterminatedQuotedString = errors.New("unterminated quoted string")

type SyntaxError struct {
	Err error
}

// Creates a new syntax error from an existing error
func NewSyntaxError(err error) SyntaxError {
	return SyntaxError{
		Err: err,
	}
}

// Returns whether the error provided is a syntax error
func IsSyntaxError(err error) bool {
	return errors.Is(err, SyntaxError{})
}

func (err SyntaxError) Error() (description string) {
	return fmt.Sprintf("syntax error: %s", err.Err.Error())
}

func (err SyntaxError) Unwrap() error {
	return err.Err
}

func (err SyntaxError) Is(err2 error) bool {
	_, ok := err2.(SyntaxError)
	return ok
}

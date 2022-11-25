package rdp

import (
	"errors"
	"fmt"
)

type SyntaxError struct{ Err error }

func newSyntaxError(err error) SyntaxError {
	return SyntaxError{
		Err: err,
	}
}

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

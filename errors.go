package gobash

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

type IoRedirectionError struct{ Err error }

func IsIORedirectionError(err error) bool {
	return errors.Is(err, IoRedirectionError{})
}

func newIORedirectionError(err error) IoRedirectionError {
	return IoRedirectionError{
		Err: err,
	}
}

func (err IoRedirectionError) Error() (description string) {
	return fmt.Sprintf("io error: %s", err.Err.Error())
}

func (err IoRedirectionError) Unwrap() error {
	return err.Err
}

func (err IoRedirectionError) Is(err2 error) bool {
	_, ok := err2.(IoRedirectionError)
	return ok
}

type UnknownCommandError struct{ Command string }

func IsUnknownCommandError(err error) bool {
	return errors.Is(err, UnknownCommandError{Command: ""})
}

func newUnknownCommandError(command string) UnknownCommandError {
	return UnknownCommandError{
		Command: command,
	}
}

func (err UnknownCommandError) Error() string {
	return fmt.Sprintf("%s: command not found", err.Command)
}

func (err UnknownCommandError) Is(err2 error) bool {
	_, ok := err2.(UnknownCommandError)
	return ok
}

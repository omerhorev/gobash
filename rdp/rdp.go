package rdp

import (
	"errors"
	"fmt"
	"io"
)

var (
	ErrOutOfBounds = errors.New("restore out of bounds")
)

type Terminal[T any] interface {
	Accept(T) bool
}

type Token interface {
	// String() string
}

// Recursive descent parser with backtracking support
type RDP[T Token, TR Terminal[T]] struct {
	Tokens []T
	Index  int
	err    error
}

func (r *RDP[T, TR]) Consume() error {
	if r.err != nil {
		return r.err
	}

	if r.Index == len(r.Tokens) {
		r.SetError(newSyntaxError(io.ErrUnexpectedEOF))
		return r.err
	}

	r.Index++

	return nil
}

func (r *RDP[T, TR]) Current() T {
	return r.Tokens[r.Index]
}

func (r *RDP[T, TR]) Prev() T {
	if r.Index == 0 {
		r.SetError(newSyntaxError(ErrOutOfBounds))
		return r.Tokens[0]
	} else {
		return r.Tokens[r.Index-1]
	}
}

func (r *RDP[T, TR]) Restore(b int) {
	if b > len(r.Tokens) {
		r.SetError(newSyntaxError(ErrOutOfBounds))
	} else {
		r.Index = b
	}
}

func (r *RDP[T, TR]) Backup() int {
	return r.Index
}

func (r *RDP[T, TR]) Check(terminal ...TR) bool {
	for _, t := range terminal {
		if t.Accept(r.Current()) {
			return true
		}
	}

	return false
}

func (r *RDP[T, TR]) Accept(expected ...TR) bool {
	if r.err != nil {
		return false
	}

	if r.Check(expected...) {
		if r.Consume() != nil {
			return false
		} else {
			return true
		}
	}

	return false
}

func (r *RDP[T, TR]) Expect(terminal TR) bool {
	if !terminal.Accept(r.Current()) {
		r.SetError(newSyntaxError(fmt.Errorf("expected %v but found %v", terminal, r.Current())))
		return false
	}

	return true
}

func (r *RDP[T, TR]) SetError(err error) {
	r.err = err
}

func (r *RDP[T, TR]) Error() error {
	return r.err
}

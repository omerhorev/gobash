package utils

import "io"

// Null is a reader/writer that returns EOF to all reads and writes
var Null = null{}

type null struct{}

func (null) Read([]byte) (int, error)  { return 0, io.EOF }
func (null) Write([]byte) (int, error) { return 0, io.EOF }

type NopWriteCloser struct {
	w io.Writer
}

func NewNopWriteCloser(w io.Writer) *NopWriteCloser {
	return &NopWriteCloser{
		w: w,
	}
}

func (w NopWriteCloser) Write(b []byte) (int, error) { return w.w.Write(b) }
func (w NopWriteCloser) Close() error                { return nil }

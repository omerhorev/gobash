package utils

import (
	"errors"
	"io"
)

var (
	ErrUnsupportedIO = errors.New("unsupported io operation")
)

// A ReadWriter that returns ErrUnsupportedIO for read operations
type ErrorReadWriterErrR struct{ io.Writer }

func (e ErrorReadWriterErrR) Write(p []byte) (n int, err error) {
	return e.Writer.Write(p)
}

func (e ErrorReadWriterErrR) Read(p []byte) (n int, err error) {
	return 0, ErrUnsupportedIO
}

func (e ErrorReadWriterErrR) Close() error {
	if closer, ok := e.Writer.(io.Closer); ok {
		return closer.Close()
	}

	return nil
}

// A ReadWriter that returns ErrUnsupportedIO for write operations
type ErrorReadWriterErrW struct{ io.Reader }

func (e ErrorReadWriterErrW) Write(p []byte) (n int, err error) {
	return 0, ErrUnsupportedIO
}

func (e ErrorReadWriterErrW) Read(p []byte) (n int, err error) {
	return e.Reader.Read(p)
}

func (e ErrorReadWriterErrW) Close() error {
	if closer, ok := e.Reader.(io.Closer); ok {
		return closer.Close()
	}

	return nil
}

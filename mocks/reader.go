package mocks

import "io"

// Hook type that will be called on the entrance to a read method.
// If the method returns an error, the execution will be stopped and the
// error and n will be returned
type ReadEntryHook func(r *MockReader, data []byte) (n int, err error)

// Hook type that will be called on the exit from a read method.
// The original return values and argument are passed to this method
// This method return values will be passed instead of the original read
type ReadExitHook func(r *MockReader, data []byte, n int, err error) (newN int, newErr error)

// MockReader is an testing helper class that allows to mock various
// reader's read errors.
//
// It allows to set hooks after and before reads and alter the input/output
type MockReader struct {
	Reader io.Reader // The original io.Reader

	// The amount of times the original io.Reader was called
	// This value is increased after the original io.Reader Read method is called
	// The exit hook is caled before this value is increased
	ReadCalls int

	// The amount of data read from the original io.Reader so far
	// This value is increased after the original io.Reader Read method is called
	// The exit hook is caled after this value is increased
	DataRead int

	ReadEntryHook ReadEntryHook // The read entry hook
	ReadExitHook  ReadExitHook  // The read exit hook
}

// Create a new mock reader from an existing reader
func NewReader(r io.Reader) *MockReader {
	return &MockReader{
		ReadCalls:     0,
		DataRead:      0,
		ReadEntryHook: nil,
		ReadExitHook:  nil,
		Reader:        r,
	}
}

// Execute the various hooks and the original io.Reader Read method
func (r *MockReader) Read(d []byte) (n int, err error) {
	defer func() { r.ReadCalls++ }()

	if mockN, mockErr := r.readEntryHook(d); mockErr != nil {
		return mockN, mockErr
	}

	n, err = r.Reader.Read(d)
	r.DataRead += n

	return r.readExitHook(d, n, err)
}

func (r *MockReader) readEntryHook(data []byte) (int, error) {
	if r.ReadEntryHook == nil {
		return 0, nil
	}

	return r.ReadEntryHook(r, data)
}

func (r *MockReader) readExitHook(data []byte, n int, err error) (int, error) {
	if r.ReadExitHook == nil {
		return n, err
	}

	return r.ReadExitHook(r, data, n, err)
}

// Return io.ErrUnexpectedEOF after the before the n call to the
// original io.Reader Read method
func UnexpectedEOFBeforeCall(n int) ReadEntryHook {
	return func(r *MockReader, data []byte) (int, error) {
		if r.ReadCalls == n {
			return 0, io.ErrUnexpectedEOF
		}

		return 0, nil
	}
}

// Return io.ErrUnexpectedEOF on every read call
func AlwaysUnexpectedEOF() ReadEntryHook {
	return func(r *MockReader, data []byte) (int, error) {
		return 0, io.ErrUnexpectedEOF
	}
}

package mocks

import (
	"bytes"
	"os"
)

// Mock the behavior of a file
type MockFile struct {
	AllowRead  bool // Allow read operations from the mock file
	AllowWrite bool // Allow write operations to the mock file

	b      *bytes.Buffer
	opened bool
}

// Creates a new mock file. The flag and perm mocks the behavior of os.OpenFile
func NewMockFile(flag int, perm os.FileMode, buffer *bytes.Buffer) *MockFile {
	rwMode := flag & 0x2

	f := MockFile{
		AllowRead:  rwMode == os.O_RDONLY || rwMode == os.O_RDWR,
		AllowWrite: rwMode == os.O_WRONLY || rwMode == os.O_RDWR,
		opened:     true,
		b:          buffer,
	}

	if flag&os.O_APPEND == 0 && f.AllowWrite {
		f.b.Reset()
	}

	return &f
}

func (f *MockFile) Read(d []byte) (int, error) {
	if !f.opened {
		return 0, os.ErrClosed
	}

	return f.b.Read(d)
}

func (f *MockFile) Write(d []byte) (int, error) {
	if !f.opened {
		return 0, os.ErrClosed
	}

	return f.b.Write(d)
}

func (f *MockFile) Close() error {
	if !f.opened {
		return os.ErrClosed
	}

	f.opened = false

	return nil
}

// Testing operation. Returns the amount of data in the buffer
func (f *MockFile) Len() int {
	return f.b.Len()
}

// Testing operation. Returns the data in the buffer
func (f *MockFile) Bytes() []byte {
	return f.b.Bytes()
}

// Testing operation. Returns the data in the buffer decoded as string
func (f *MockFile) String() string {
	return f.b.String()
}

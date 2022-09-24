package gobash

import (
	"io"
	"strings"
	"testing"

	"github.com/omerhorev/gobash/mocks"
	"github.com/stretchr/testify/assert"
)

// Test LineReader functionality including non-english runes
func TestReadLine(t *testing.T) {
	line, err := "", error(nil)
	l := NewLineReader(strings.NewReader(`line1
line2
line3.1\
line3.2
line4
שורה5
שורה6.1\
שורה6.2`))
	line, err = l.ReadLine()
	assert.NoError(t, err)
	assert.Equal(t, "line1", line)

	line, err = l.ReadLine()
	assert.NoError(t, err)
	assert.Equal(t, "line2", line)

	line, err = l.ReadLine()
	assert.NoError(t, err)
	assert.Equal(t, "line3.1\\\nline3.2", line)

	line, err = l.ReadLine()
	assert.NoError(t, err)
	assert.Equal(t, "line4", line)

	line, err = l.ReadLine()
	assert.NoError(t, err)
	assert.Equal(t, "שורה5", line)

	line, err = l.ReadLine()
	assert.ErrorIs(t, err, io.EOF)
	assert.Equal(t, "שורה6.1\\\nשורה6.2", line)
}

// Test LineReader functionality including non-english runes
func TestReadLineEOF(t *testing.T) {
	l := NewLineReader(strings.NewReader(`la\`))

	_, err := l.ReadLine()
	assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
}

func TestReadLineError(t *testing.T) {
	r := mocks.NewReader(strings.NewReader("abcdabcdabcdabcdabcdabcdabcdabcd"))
	r.ReadEntryHook = mocks.AlwaysUnexpectedEOF()
	_, err := NewLineReaderSize(r, 1).ReadLine()
	assert.ErrorIs(t, err, io.ErrUnexpectedEOF)

	r = mocks.NewReader(strings.NewReader("abcdabcdabcdabcdabcdabcdabcdabcd"))
	r.ReadEntryHook = mocks.UnexpectedEOFBeforeCall(2)
	_, err = NewLineReaderSize(r, 1).ReadLine()
	assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
}

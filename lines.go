package gobash

import (
	"bufio"
	"io"
	"strings"

	"github.com/pkg/errors"
)

var (
	// The default buffer read size that will be used in NewLineReader
	DefaultBufferReadSize = 4096
)

// LineReader is used to read lines as defined in IEEE1003.1
// The struct can be initialized via NewLineReader
//
// briefly, LineReader reads line by line, while supporting line end escaping using
// backslash while fully supporting UTF-8.
type LineReader struct {
	reader *bufio.Reader
}

// Initialize a new LineReader using an existing reader.
//
// The LineReader uses bufio.Reader to read runes, so all the read operations are
// buffered, it may read more data than needed, and pass it over to the next line.
func NewLineReader(r io.Reader) *LineReader {
	return NewLineReaderSize(r, DefaultBufferReadSize)
}

// Initialize a new LineReader using an existing reader and a buffer size.
//
// The LineReader uses bufio.Reader to read runes, so all the read operations are
// buffered, it may read more data than needed, and pass it over to the next line.
func NewLineReaderSize(r io.Reader, n int) *LineReader {
	return &LineReader{
		reader: bufio.NewReaderSize(r, n),
	}
}

// ReadLine will read from the buffer an unescaped newline
func (lr *LineReader) ReadLine() (string, error) {
	isEscaped := false
	builder := strings.Builder{}

	for {
		r, _, err := lr.reader.ReadRune()
		if isEscaped && errors.Is(err, io.EOF) {
			// the data can't end inside an escape
			return "", errors.Wrap(io.ErrUnexpectedEOF, "ReadRune")
		} else if errors.Is(err, io.EOF) {
			// Valid EOF, end of line
			return builder.String(), io.EOF
		} else if err != nil {
			// Other error
			return "", errors.Wrap(err, "ReadRune")
		}

		if !isEscaped && isNewLine(r) {
			break
		}

		isEscaped = !isEscaped && isEscape(r)

		if _, err := builder.WriteRune(r); err != nil {
			return "", errors.Wrap(err, "WriteRune")
		}
	}

	return builder.String(), nil
}

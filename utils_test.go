package gobash

import (
	"bytes"
)

func createTestShell(stdin string) (shell *Shell, stdout *bytes.Buffer, stderr *bytes.Buffer, err error) {
	inReader := bytes.NewReader([]byte(stdin))
	outWriter := new(bytes.Buffer)
	errWriter := new(bytes.Buffer)

	s := NewShell(inReader, outWriter, errWriter)

	return s, outWriter, errWriter, err
}

package gobash

import (
	"bufio"
	"context"
	"io"
	"strings"
)

type Shell struct {
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader

	// Current shell execution environment
	execEnv *ExecEnv

	reader *bufio.Reader // the script/line/interactive reader
	token  *Token        // the current token in the recursice discect parsing algorithm
}

func NewShell(stdin io.Reader, stdout io.Writer, stderr io.Writer) *Shell {
	return &Shell{
		Stdin:   stdin,
		Stdout:  stdout,
		Stderr:  stderr,
		execEnv: newExecEnv(),
		reader:  nil,
	}
}

// Run the shell with the current environment and settings, while evaluating
// the miltu-line script in r.
//
// The context can be used to terminate the executaion of the entire shell,
func (s *Shell) RunContext(ctx context.Context, r io.Reader) error {

	s.reader = bufio.NewReader(r)

	return nil
}

func canBeUsedInOperator(token string) bool {
	for _, o := range operatorsStrings {
		if strings.HasPrefix(o, token) {
			return true
		}
	}

	return false
}

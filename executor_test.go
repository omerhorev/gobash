package gobash

import (
	"bytes"
	"io"
	"testing"

	"github.com/omerhorev/gobash/ast"
	"github.com/stretchr/testify/require"
)

func TestExecutorPipe(t *testing.T) {
	executor := createTestExecutor()
	b := bytes.Buffer{}
	executor.Stdout = &b

	executor.Run(&ast.Program{
		Commands: []ast.Node{
			&ast.Pipe{
				Commands: []ast.Node{
					&ast.SimpleCommand{
						Word: "echo",
						Args: []string{"a", "b", "c"},
					},
					&ast.SimpleCommand{
						Word: "rev",
					},
				},
			},
		},
	})

	require.Equal(t, "c b a", b.String())
}

func createTestExecutor() *Executor {
	e := NewExecutor(ExecutorSettings{})

	e.RegisterCommand(&Cmd{
		Name: "echo",
		Run: func(args []string, env *CommandExecEnv) int {
			args = args[1:]
			for i, a := range args {
				env.Stdout().Write([]byte(a))
				if i < len(args)-1 {
					env.Stdout().Write([]byte(" "))
				}
			}

			return 0
		},
	})

	e.RegisterCommand(&Cmd{
		Name: "rev",
		Run: func(args []string, env *CommandExecEnv) int {
			if len(args) != 1 {
				return 1
			}

			buf, err := io.ReadAll(env.Stdin())
			if err != nil {
				return 1
			}

			for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
				buf[i], buf[j] = buf[j], buf[i]
			}

			env.Stdout().Write(buf)

			return 0
		},
	})

	return e
}

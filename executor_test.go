package gobash

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/omerhorev/gobash/ast"
	"github.com/omerhorev/gobash/command"
	"github.com/omerhorev/gobash/mocks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestExecutorPipe(t *testing.T) {
	executor := createTestExecutor()
	b := bytes.Buffer{}
	executor.SetStdout(&b)

	require.NoError(t, executor.Run(&ast.Program{
		Commands: []ast.Node{
			&ast.Pipe{
				Commands: []ast.Node{
					&ast.SimpleCommand{Word: "echo", Args: []string{"a", "b", "c"}},
					&ast.SimpleCommand{Word: "rev"},
				},
			},
			&ast.Pipe{
				Commands: []ast.Node{
					&ast.SimpleCommand{Word: "echo", Args: []string{"d"}},
				},
			},
		},
	}))

	require.Equal(t, "c b a", b.String())

	require.Error(t, executor.Run(&ast.Program{Commands: []ast.Node{&ast.Pipe{}}}))

	require.NoError(t, executor.Run(&ast.Program{
		Commands: []ast.Node{
			&ast.Pipe{
				Commands: []ast.Node{
					&ast.SimpleCommand{Word: "echo"},
					&ast.SimpleCommand{Word: "echo"},
				},
			},
		},
	}))
}

func TestExecutorCommands(t *testing.T) {
	executor := createTestExecutor()
	bufferStdout := bytes.Buffer{}
	bufferStderr := bytes.Buffer{}
	executor.SetStdout(&bufferStdout)
	executor.SetStderr(&bufferStderr)

	prog := &ast.Program{
		Commands: []ast.Node{
			&ast.SimpleCommand{Word: "unknown"},
		},
	}

	executor.Settings.StopOnUnknownCommand = true
	require.Error(t, executor.Run(prog))
	require.Equal(t, "unknown: command not found\n", bufferStderr.String())

	bufferStderr.Reset()

	executor.Settings.StopOnUnknownCommand = false
	require.NoError(t, executor.Run(prog))
	require.Equal(t, "unknown: command not found\n", bufferStderr.String())
}

func TestExecutorIORedirection(t *testing.T) {
	executor := createTestExecutor()
	bufferStdout := bytes.Buffer{}
	bufferStderr := bytes.Buffer{}
	executor.SetStdout(&bufferStdout)
	executor.SetStderr(&bufferStderr)

	files := map[string]*bytes.Buffer{
		"input/1": bytes.NewBufferString("123"),
	}

	fileOpener := func(path string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
		if buffer, exists := files[path]; exists {
			return mocks.NewMockFile(flag, perm, buffer), nil
		} else if flag&os.O_CREATE == 0 {
			return nil, os.ErrNotExist
		} else {
			files[path] = bytes.NewBufferString("")
			return mocks.NewMockFile(flag, perm, files[path]), nil
		}
	}

	executor.Settings.OpenFunc = fileOpener

	executor.Run(&ast.Program{
		Commands: []ast.Node{
			&ast.SimpleCommand{
				Word: "echo",
				Args: []string{"1"},
				Redirects: []*ast.IORedirection{
					{Fd: 1, Mode: ast.IORedirectionModeOutput, Value: "output/1"},
				},
			},
			&ast.SimpleCommand{
				Word: "echo",
				Args: []string{"2"},
				Redirects: []*ast.IORedirection{
					{Fd: 1, Mode: ast.IORedirectionModeOutputAppend, Value: "output/1"},
				},
			},
			&ast.SimpleCommand{
				Word: "echo",
				Args: []string{"fd_io_redirect"},
				Redirects: []*ast.IORedirection{
					{Fd: 10, Mode: ast.IORedirectionModeOutputAppend, Value: "output/2"},
					{Fd: 1, Mode: ast.IORedirectionModeOutputFd, Value: 10},
				},
			},
			&ast.SimpleCommand{
				Word: "rev",
				Redirects: []*ast.IORedirection{
					{Fd: 10, Mode: ast.IORedirectionModeInput, Value: "input/1"},
					{Fd: 0, Mode: ast.IORedirectionModeInputFd, Value: 10},
					{Fd: 1, Mode: ast.IORedirectionModeOutput, Value: "output/3"},
				},
			},
		},
	})

	require.Contains(t, files, "output/1")
	require.Contains(t, files, "output/2")
	require.Contains(t, files, "output/3")
	require.Equal(t, "1\n2\n", files["output/1"].String())
	require.Equal(t, "fd_io_redirect\n", files["output/2"].String())
	require.Equal(t, "321", files["output/3"].String())

	progErrorIORedirectionFile := &ast.Program{
		Commands: []ast.Node{
			&ast.SimpleCommand{Word: "true", Redirects: []*ast.IORedirection{
				{Fd: 1, Mode: ast.IORedirectionModeInput, Value: "missing_file"},
			}},
		},
	}

	progErrorIORedirectionFd := &ast.Program{
		Commands: []ast.Node{
			&ast.SimpleCommand{Word: "true", Redirects: []*ast.IORedirection{
				{Fd: 1, Mode: ast.IORedirectionModeInputFd, Value: 22},
			}},
		},
	}

	executor.Settings.StopOnIORedirectionError = true
	require.ErrorIs(t, executor.Run(progErrorIORedirectionFile), newIORedirectionError(errors.New("some error")))
	require.Equal(t, "io error: missing_file: file does not exist\n", bufferStderr.String())
	bufferStderr.Reset()

	require.ErrorIs(t, executor.Run(progErrorIORedirectionFd), newIORedirectionError(errors.New("some error")))
	require.Equal(t, "io error: 22: bad file descriptor\n", bufferStderr.String())
	bufferStderr.Reset()

	executor.Settings.StopOnIORedirectionError = false
	require.NoError(t, executor.Run(progErrorIORedirectionFile))
	require.Equal(t, "io error: missing_file: file does not exist\n", bufferStderr.String())
	bufferStderr.Reset()
}

func TestExecutorBinary(t *testing.T) {
	executor := createTestExecutor()
	b := bytes.Buffer{}
	executor.SetStdout(&b)

	executor.Run(&ast.Program{
		Commands: []ast.Node{
			&ast.Binary{
				Left:  makeSimpleCommand("true"),
				Right: makeSimpleCommand("echo", "1"),
				Type:  ast.BinaryTypeAnd,
			},
		},
	})

	require.Equal(t, "1\n", b.String())
	b.Reset()

	executor.Run(&ast.Program{
		Commands: []ast.Node{
			&ast.Binary{
				Left:  makeSimpleCommand("false"),
				Right: makeSimpleCommand("echo", "1"),
				Type:  ast.BinaryTypeAnd,
			},
		},
	})

	require.Equal(t, "", b.String())
	b.Reset()

	executor.Run(&ast.Program{
		Commands: []ast.Node{
			&ast.Binary{
				Left:  makeSimpleCommand("true"),
				Right: makeSimpleCommand("echo", "1"),
				Type:  ast.BinaryTypeOr,
			},
		},
	})

	require.Equal(t, "", b.String())
	b.Reset()

	executor.Run(&ast.Program{
		Commands: []ast.Node{
			&ast.Binary{
				Left:  makeSimpleCommand("false"),
				Right: makeSimpleCommand("echo", "1"),
				Type:  ast.BinaryTypeOr,
			},
		},
	})

	require.Equal(t, "1\n", b.String())
	b.Reset()
}

func TestExecutorBuiltinCd(t *testing.T) {
	executor := createTestExecutor()
	bufferStdout := bytes.Buffer{}
	bufferStderr := bytes.Buffer{}
	executor.SetStdout(&bufferStdout)
	executor.SetStderr(&bufferStderr)
	executor.ExecEnv.SetParam("HOME", "/home/user/")

	executor.Settings.CdFunc = func(path string) (newPath string, err error) {
		if path == "/tmp" {
			return "/tmp", nil
		} else if path == "/" {
			return "/", nil
		} else if path == "/home/user/" {
			return "/home/user2/", nil
		} else {
			return "/", errors.Errorf("unknown path %s", path)
		}
	}

	prog1 := &ast.Program{Commands: []ast.Node{
		&ast.SimpleCommand{Word: "cd", Args: []string{"/tmp"}},
		&ast.SimpleCommand{Word: "cd", Args: []string{"/"}},
		&ast.SimpleCommand{Word: "cd", Args: []string{}},
	}}

	prog2 := &ast.Program{Commands: []ast.Node{
		&ast.SimpleCommand{Word: "cd", Args: []string{"/bad"}},
	}}

	prog3 := &ast.Program{Commands: []ast.Node{
		&ast.SimpleCommand{Word: "cd", Args: []string{"/", "/"}},
	}}

	executor.Settings.StopOnBuiltinError = true
	require.NoError(t, executor.Run(prog1))
	require.ErrorIs(t, executor.Run(prog2), newBuiltinError(errors.New("")))
	require.ErrorIs(t, executor.Run(prog3), newBuiltinError(errors.New("")))

	executor.Settings.StopOnBuiltinError = false
	require.NoError(t, executor.Run(prog2))
}

func TestExecutorEnv(t *testing.T) {
	executor := createTestExecutor()
	bufferStdout := bytes.Buffer{}
	bufferStderr := bytes.Buffer{}
	executor.SetStdout(&bufferStdout)
	executor.SetStderr(&bufferStderr)
	executor.ExecEnv.SetParam("X", "Y")
	executor.ExecEnv.SetParam("C", "D")

	prog1 := &ast.Program{Commands: []ast.Node{
		&ast.SimpleCommand{Word: "env", Assignments: map[string]string{"A": "B", "C": "E"}},
	}}

	require.NoError(t, executor.Run(prog1))
	b := bufio.NewScanner(&bufferStdout)
	lines := []string{}
	for b.Scan() {
		lines = append(lines, b.Text())
	}

	require.Contains(t, lines, "X=Y")
	require.Contains(t, lines, "A=B")
	require.Contains(t, lines, "C=E")
	require.Len(t, lines, 3)
}

func createTestExecutor() *Executor {
	e := NewExecutor(ExecutorSettings{})
	e.AddCommands(command.Default...)

	return e
}

func makeSimpleCommand(name string, args ...string) *ast.SimpleCommand {
	return &ast.SimpleCommand{
		Word: name,
		Args: args,
	}
}

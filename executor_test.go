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
					&ast.SimpleCommand{
						Word: ast.NewExpr(ast.NewString("echo")),
						Args: []*ast.Expr{
							ast.NewExpr(ast.NewString("a")),
							ast.NewExpr(ast.NewString("b")),
							ast.NewExpr(ast.NewString("c")),
						},
					},
					&ast.SimpleCommand{Word: ast.NewExpr(ast.NewString("rev"))},
				},
			},
			&ast.Pipe{
				Commands: []ast.Node{
					&ast.SimpleCommand{
						Word: ast.NewExpr(ast.NewString("echo")),
						Args: []*ast.Expr{
							ast.NewExpr(ast.NewString("d")),
						},
					},
				},
			},
		},
	}))

	require.Equal(t, "c b a\nd\n", b.String())

	require.Error(t, executor.Run(&ast.Program{Commands: []ast.Node{&ast.Pipe{}}}))

	require.NoError(t, executor.Run(&ast.Program{
		Commands: []ast.Node{
			&ast.Pipe{
				Commands: []ast.Node{
					&ast.SimpleCommand{Word: ast.NewExpr(ast.NewString("echo"))},
					&ast.SimpleCommand{Word: ast.NewExpr(ast.NewString("echo"))},
					&ast.SimpleCommand{Word: ast.NewExpr(ast.NewString("echo"))},
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
			&ast.SimpleCommand{Word: ast.NewExpr(ast.NewString("unknown"))},
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
				Word: ast.NewExpr(ast.NewString("echo")),
				Args: []*ast.Expr{ast.NewExpr(ast.NewString("1"))},
				Redirects: []*ast.IORedirection{
					{Fd: 1, Mode: ast.IORedirectionModeOutput, Value: ast.NewExprStr("output/1")},
				},
			},
			&ast.SimpleCommand{
				Word: ast.NewExpr(ast.NewString("echo")),
				Args: []*ast.Expr{ast.NewExpr(ast.NewString("2"))},
				Redirects: []*ast.IORedirection{
					{Fd: 1, Mode: ast.IORedirectionModeOutputAppend, Value: ast.NewExprStr("output/1")},
				},
			},
			&ast.SimpleCommand{
				Word: ast.NewExpr(ast.NewString("echo")),
				Args: []*ast.Expr{ast.NewExpr(ast.NewString("fd_io_redirect"))},
				Redirects: []*ast.IORedirection{
					{Fd: 10, Mode: ast.IORedirectionModeOutputAppend, Value: ast.NewExprStr("output/2")},
					{Fd: 1, Mode: ast.IORedirectionModeOutputFd, Value: ast.NewExprStr("10")},
				},
			},
			&ast.SimpleCommand{
				Word: ast.NewExpr(ast.NewString("rev")),
				Redirects: []*ast.IORedirection{
					{Fd: 10, Mode: ast.IORedirectionModeInput, Value: ast.NewExprStr("input/1")},
					{Fd: 0, Mode: ast.IORedirectionModeInputFd, Value: ast.NewExprStr("10")},
					{Fd: 1, Mode: ast.IORedirectionModeOutput, Value: ast.NewExprStr("output/3")},
				},
			},
		},
	})

	require.Contains(t, files, "output/1")
	require.Contains(t, files, "output/2")
	require.Contains(t, files, "output/3")
	require.Equal(t, "1\n2\n", files["output/1"].String())
	require.Equal(t, "fd_io_redirect\n", files["output/2"].String())
	require.Equal(t, "321\n", files["output/3"].String())

	progErrorIORedirectionFile := &ast.Program{
		Commands: []ast.Node{
			&ast.SimpleCommand{Word: ast.NewExprStr("true"), Redirects: []*ast.IORedirection{
				{Fd: 1, Mode: ast.IORedirectionModeInput, Value: ast.NewExprStr("missing_file")},
			}},
		},
	}

	progErrorIORedirectionFd := &ast.Program{
		Commands: []ast.Node{
			&ast.SimpleCommand{Word: ast.NewExprStr("true"), Redirects: []*ast.IORedirection{
				{Fd: 1, Mode: ast.IORedirectionModeInputFd, Value: ast.NewExprStr("22")},
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
				Left:  &ast.SimpleCommand{Word: ast.NewExprStr("true")},
				Right: &ast.SimpleCommand{Word: ast.NewExprStr("echo"), Args: []*ast.Expr{ast.NewExprStr("1")}},
				Type:  ast.BinaryTypeAnd,
			},
		},
	})

	require.Equal(t, "1\n", b.String())
	b.Reset()

	executor.Run(&ast.Program{
		Commands: []ast.Node{
			&ast.Binary{
				Left:  &ast.SimpleCommand{Word: ast.NewExprStr("false")},
				Right: &ast.SimpleCommand{Word: ast.NewExprStr("echo"), Args: []*ast.Expr{ast.NewExprStr("1")}},
				Type:  ast.BinaryTypeAnd,
			},
		},
	})

	require.Equal(t, "", b.String())
	b.Reset()

	executor.Run(&ast.Program{
		Commands: []ast.Node{
			&ast.Binary{
				Left:  &ast.SimpleCommand{Word: ast.NewExprStr("true")},
				Right: &ast.SimpleCommand{Word: ast.NewExprStr("echo"), Args: []*ast.Expr{ast.NewExprStr("1")}},
				Type:  ast.BinaryTypeOr,
			},
		},
	})

	require.Equal(t, "", b.String())
	b.Reset()

	executor.Run(&ast.Program{
		Commands: []ast.Node{
			&ast.Binary{
				Left:  &ast.SimpleCommand{Word: ast.NewExprStr("false")},
				Right: &ast.SimpleCommand{Word: ast.NewExprStr("echo"), Args: []*ast.Expr{ast.NewExprStr("1")}},
				Type:  ast.BinaryTypeOr,
			},
		},
	})

	require.Equal(t, "1\n", b.String())
	b.Reset()
}

func TestExecutorFieldSplitting(t *testing.T) {
	executor := createTestExecutor()
	bufferStdout := bytes.Buffer{}
	bufferStderr := bytes.Buffer{}
	executor.SetStdout(&bufferStdout)
	executor.SetStderr(&bufferStderr)

	prog1 := &ast.Program{Commands: []ast.Node{
		&ast.SimpleCommand{
			Word: ast.NewExprStr("echo"),
			Args: []*ast.Expr{
				ast.NewExpr(testBacktickEcho("1")),
			},
		},
	}}

	require.NoError(t, executor.Run(prog1))
}

func TestExecutorExpander(t *testing.T) {
	executor := createTestExecutor()
	bufferStdout := bytes.Buffer{}
	bufferStderr := bytes.Buffer{}
	executor.SetStdout(&bufferStdout)
	executor.SetStderr(&bufferStderr)

	//
	// X=`echo 1` e`echo ch`o h`echo ell`o
	//
	prog1 := &ast.Program{Commands: []ast.Node{
		&ast.SimpleCommand{
			Assignments: map[string]*ast.Expr{
				"X": ast.NewExpr(testBacktickEcho("1")),
			},
			Word: ast.NewExpr(ast.NewString("e"), testBacktickEcho("ch"), ast.NewString("o")),
			Args: []*ast.Expr{
				ast.NewExpr(ast.NewString("h"), testBacktickEcho("ell"), ast.NewString("o")),
			},
		},
	}}

	require.NoError(t, executor.Run(prog1))
	require.Equal(t, "hello\n", bufferStdout.String())
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
		&ast.SimpleCommand{Word: ast.NewExprStr("cd"), Args: []*ast.Expr{ast.NewExprStr("/tmp")}},
		&ast.SimpleCommand{Word: ast.NewExprStr("cd"), Args: []*ast.Expr{ast.NewExprStr("/")}},
		&ast.SimpleCommand{Word: ast.NewExprStr("cd"), Args: []*ast.Expr{ast.NewExprStr("")}},
	}}

	prog2 := &ast.Program{Commands: []ast.Node{
		&ast.SimpleCommand{Word: ast.NewExprStr("cd"), Args: []*ast.Expr{ast.NewExprStr("/bad")}},
	}}

	prog3 := &ast.Program{Commands: []ast.Node{
		&ast.SimpleCommand{Word: ast.NewExprStr("cd"), Args: []*ast.Expr{ast.NewExprStr("/"), ast.NewExprStr("/")}},
	}}

	require.NoError(t, executor.Run(prog1))
	require.Equal(t, "cd: unknown path ", bufferStderr.String())
	bufferStderr.Reset()

	require.NoError(t, executor.Run(prog2))
	require.Equal(t, "cd: unknown path /bad", bufferStderr.String())
	bufferStderr.Reset()

	require.NoError(t, executor.Run(prog3))
	require.Equal(t, "cd: too many arguments", bufferStderr.String())
	bufferStderr.Reset()
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
		&ast.SimpleCommand{Word: ast.NewExprStr("env"), Assignments: map[string]*ast.Expr{"A": ast.NewExprStr("B"), "C": ast.NewExprStr("E")}},
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

func testBacktickEcho(args ...string) ast.Node {
	args2 := []*ast.Expr{}
	for _, arg := range args {
		args2 = append(args2, ast.NewExprStr(arg))
	}

	return &ast.Backtick{
		Node: &ast.SimpleCommand{
			Word: ast.NewExprStr("echo"),
			Args: args2,
		},
	}
}

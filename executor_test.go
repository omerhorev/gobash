package gobash

// import (
// 	"io"
// 	"testing"

// 	"github.com/omerhorev/gobash/ast"
// 	"github.com/stretchr/testify/assert"
// )

// func TestExecutor(t *testing.T) {
// 	e := NewExecutor(
// 		ExecutorSettings{},
// 		ast.Program{
// 			Commands: []ast.{
// 				&ast.SimpleCommand{
// 					Word: "ls",
// 					PipeTo: &ast.SimpleCommand{
// 						Word: "grep",
// 						IORedirect: &ast.IORedirectFile{
// 							Path: "/tmp/file",
// 						},
// 					},
// 				},
// 				&ast.SimpleCommand{
// 					Word:   "cat",
// 					PipeTo: nil,
// 				},
// 			},
// 		},
// 	)

// 	e.Register(&Cmd{
// 		Name: "ls",
// 		f: func(stdin io.Reader, stdout, stderr io.Writer) (int, error) {
// 			return 0, nil
// 		},
// 	})

// 	e.Register(&Cmd{
// 		Name: "cat",
// 		f: func(stdin io.Reader, stdout, stderr io.Writer) (int, error) {
// 			return 0, nil
// 		},
// 	})

// 	assert.NoError(t, e.Run())
// }

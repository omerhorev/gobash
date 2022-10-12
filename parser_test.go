package gobash

import (
	"testing"

	"github.com/omerhorev/gobash/ast"
	"github.com/stretchr/testify/require"
)

func TestParserErrors(t *testing.T) {
	parserTest(t, "ls")
	parserTest(t, "ls > 1")
	parserTest(t, "ls >> 1")
	parserTest(t, "ls 1> 1")
	parserTest(t, "ls >| 1")
	parserTest(t, "ls < 1")
	parserTest(t, "ls <> 1")
	parserTest(t, "ls &")
	parserTest(t, "ls ; ls")
	parserTest(t, "ls ; ls; ls")
	parserTest(t, "ls ; ls; ls && ls")
	parserTest(t, "ls ; ls; ls && ls; ls")
	parserTest(t, "ls ; ls; ls && ls; ls > 1")
	parserTest(t, "ls ; ls; ls && ls > 1; ls > 1")
	parserTest(t, "ls > 1; ls; ls && ls > 1; ls > 1")
	parserTest(t, "x > y && y > x")
	parserTest(t, "x 1")
	parserTest(t, "x 1> y")
	parserTest(t, "x 1 > y")
	parserTest(t, "a\nb\nc")
	parserTest(t, "\na\nb\n")
	parserTest(t, "\n\na\nb;\n\n")

	parserTestError(t, "ls 1<&a")
	parserTestError(t, "ls |")
	parserTestError(t, "ls &&&")
	parserTestError(t, "ls &&")
	parserTestError(t, "\na\nb &&")
}

func TestParserAST1(t *testing.T) {
	p := parseDefaultText(t, `
'c'>1 | A=1 d 4<&2 && c || y
ls&x;ls&
`)
	requireNode(t, p.AST(), &ast.Program{
		Commands: []ast.Node{
			&ast.Binary{
				Left: &ast.Pipe{
					Commands: []ast.Node{
						&ast.SimpleCommand{
							Word:        "'c'",
							Assignments: map[string]string{},
							Redirects: map[int]*ast.SimpleCommandIORedirection{
								1: {
									Mode:  ast.SimpleCommandIORedirectionModeOutput,
									Value: "1",
								},
							},
						},
						&ast.SimpleCommand{
							Word: "d",
							Assignments: map[string]string{
								"A": "1",
							},
							Redirects: map[int]*ast.SimpleCommandIORedirection{
								4: {
									Mode:  ast.SimpleCommandIORedirectionModeInputFd,
									Value: 2,
								},
							},
						},
					},
				},
				Right: &ast.Binary{
					Left: &ast.SimpleCommand{
						Word:        "c",
						Assignments: map[string]string{},
						Redirects:   map[int]*ast.SimpleCommandIORedirection{},
					},
					Right: &ast.SimpleCommand{
						Word:        "y",
						Assignments: map[string]string{},
						Redirects:   map[int]*ast.SimpleCommandIORedirection{},
					},
					Type: ast.BinaryTypeOr,
				},
				Type: ast.BinaryTypeAnd,
			},
			&ast.Background{
				Child: &ast.SimpleCommand{
					Word:        "ls",
					Assignments: map[string]string{},
					Redirects:   map[int]*ast.SimpleCommandIORedirection{},
				},
			},
			&ast.SimpleCommand{
				Word:        "x",
				Assignments: map[string]string{},
				Redirects:   map[int]*ast.SimpleCommandIORedirection{},
			},
			&ast.Background{
				Child: &ast.SimpleCommand{
					Word:        "ls",
					Assignments: map[string]string{},
					Redirects:   map[int]*ast.SimpleCommandIORedirection{},
				},
			},
		},
	})
}

func TestParserNot(t *testing.T) {
	p := parseDefaultText(t, "x;! y")
	requireNode(t, p.AST(), &ast.Program{
		Commands: []ast.Node{
			&ast.SimpleCommand{
				Word:        "x",
				Assignments: map[string]string{},
				Redirects:   map[int]*ast.SimpleCommandIORedirection{},
			},
			&ast.Not{
				Child: &ast.SimpleCommand{
					Word:        "y",
					Assignments: map[string]string{},
					Redirects:   map[int]*ast.SimpleCommandIORedirection{},
				},
			},
		},
	})
}

func TestParserASTArguments(t *testing.T) {
	p := parseDefaultText(t, `A=1 x a b c | y a <1 b c`)
	requireNode(t, p.AST(), &ast.Program{
		Commands: []ast.Node{
			&ast.Pipe{
				Commands: []ast.Node{
					&ast.SimpleCommand{
						Word:        "x",
						Assignments: map[string]string{"A": "1"},
						Args:        []string{"a", "b", "c"},
					},
					&ast.SimpleCommand{
						Word: "y",
						Args: []string{"a", "b", "c"},
						Redirects: map[int]*ast.SimpleCommandIORedirection{
							0: {
								Mode:  ast.SimpleCommandIORedirectionModeInput,
								Value: "1",
							},
						},
					},
				},
			},
		},
	},
	)
}

func TestParserASTRedirects(t *testing.T) {
	p := parseDefaultText(t, `x <1 >1 2<1 3<>1 4>>1 ; <&1 >&1 2>|1 10<2`)
	requireNode(t, p.AST(), &ast.Program{
		Commands: []ast.Node{
			&ast.SimpleCommand{
				Word:        "x",
				Assignments: map[string]string{},
				Redirects: map[int]*ast.SimpleCommandIORedirection{
					1: {
						Mode:  ast.SimpleCommandIORedirectionModeOutput,
						Value: "1",
					},
					0: {
						Mode:  ast.SimpleCommandIORedirectionModeInput,
						Value: "1",
					},
					2: {
						Mode:  ast.SimpleCommandIORedirectionModeInput,
						Value: "1",
					},
					3: {
						Mode:  ast.SimpleCommandIORedirectionModeInputOutput,
						Value: "1",
					},
					4: {
						Mode:  ast.SimpleCommandIORedirectionModeOutputAppend,
						Value: "1",
					},
				},
			},
			&ast.SimpleCommand{
				Word:        "",
				Assignments: map[string]string{},
				Redirects: map[int]*ast.SimpleCommandIORedirection{
					0: {
						Mode:  ast.SimpleCommandIORedirectionModeInputFd,
						Value: 1,
					},
					1: {
						Mode:  ast.SimpleCommandIORedirectionModeOutputFd,
						Value: 1,
					},
					2: {
						Mode:  ast.SimpleCommandIORedirectionModeOutputForce,
						Value: "1",
					},
					10: {
						Mode:  ast.SimpleCommandIORedirectionModeInput,
						Value: "2",
					},
				},
			},
		},
	})
}

func TestParserEmptyScript(t *testing.T) {
	p := parseDefaultText(t, "")
	requireNode(t, p.AST(), &ast.Program{Commands: []ast.Node{}})

	p = parseDefaultText(t, "\n")
	requireNode(t, p.AST(), &ast.Program{Commands: []ast.Node{}})

	p = parseDefaultText(t, "\n\n\n\n")
	requireNode(t, p.AST(), &ast.Program{Commands: []ast.Node{}})
}

// creates a token list from strings and add EOF
func parserTest(t *testing.T, text string) {
	p := parseDefaultText(t, text)
	require.NoError(t, p.Error())
}

func parserTestError(t *testing.T, text string) {
	p := parseDefaultText(t, text)
	require.Error(t, p.Error())
}

func parseDefaultText(t *testing.T, text string) *Parser {
	tokenizer := NewTokenizerShort(text)
	tokens, err := tokenizer.ReadAll()
	require.NoError(t, err)

	parser := NewParserDefault(tokens)
	parser.Parse()

	return parser
}

func requireNode(t *testing.T, actual ast.Node, expected ast.Node) {
	switch a := expected.(type) {
	case *ast.Binary:
		requireBinary(t, actual, a)
	case *ast.Pipe:
		requirePipe(t, actual, a)
	case *ast.SimpleCommand:
		requireSimpleCmd(t, actual, a)
	case *ast.Program:
		requireProgram(t, actual, a)
	case *ast.Background:
		requireBackground(t, actual, a)
	case *ast.Not:
		requireNot(t, actual, a)
	default:
		require.Nil(t, actual)
		require.Nil(t, expected)
	}
}
func requireNot(t *testing.T, node ast.Node, not *ast.Not) {
	require.IsType(t, node, &ast.Not{})
	n := node.(*ast.Not)

	requireNode(t, n.Child, not.Child)
}

func requireBackground(t *testing.T, node ast.Node, background *ast.Background) {
	require.IsType(t, node, &ast.Background{})
	b := node.(*ast.Background)

	requireNode(t, b.Child, background.Child)
}

func requireProgram(t *testing.T, node ast.Node, prog *ast.Program) {
	require.IsType(t, node, &ast.Program{})
	p := node.(*ast.Program)

	require.Len(t, p.Commands, len(prog.Commands))

	for i := range p.Commands {
		requireNode(t, prog.Commands[i], p.Commands[i])
	}
}

func requireBinary(t *testing.T, node ast.Node, binary *ast.Binary) {
	require.IsType(t, node, &ast.Binary{})
	b := node.(*ast.Binary)

	requireNode(t, b.Left, binary.Left)
	requireNode(t, b.Right, binary.Right)
	require.Equal(t, binary.Type, b.Type)
}

func requirePipe(t *testing.T, node ast.Node, expectedPipe *ast.Pipe) {
	require.IsType(t, node, &ast.Pipe{})
	pipe := node.(*ast.Pipe)

	require.Len(t, pipe.Commands, len(expectedPipe.Commands))

	for i := range pipe.Commands {
		requireNode(t, pipe.Commands[i], expectedPipe.Commands[i])
	}
}

func requireSimpleCmd(t *testing.T, node ast.Node, expectedCmd *ast.SimpleCommand) {
	require.IsType(t, node, &ast.SimpleCommand{})
	cmd := node.(*ast.SimpleCommand)
	require.Equal(t, expectedCmd.Word, cmd.Word)
	require.Len(t, cmd.Assignments, len(cmd.Assignments))
	require.Len(t, cmd.Redirects, len(cmd.Redirects))

	for k, v := range expectedCmd.Assignments {
		require.Contains(t, cmd.Assignments, k)
		require.Equal(t, v, cmd.Assignments[k])
	}

	for k, v := range expectedCmd.Redirects {
		require.Contains(t, cmd.Redirects, k)
		require.Equal(t, cmd.Redirects[k].Mode, v.Mode)
		require.Equal(t, cmd.Redirects[k].Value, v.Value)
	}
}

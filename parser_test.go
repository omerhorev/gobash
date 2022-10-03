package gobash

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

	parserTestError(t, "ls |")
	parserTestError(t, "ls &&&")
	parserTestError(t, "ls &&")
	parserTestError(t, "\na\nb &&")
}

// creates a token list from strings and add EOF
func parserTest(t *testing.T, text string) {
	tokenizer := NewTokenizerShort(text)
	tokens, err := tokenizer.ReadAll()
	assert.NoError(t, err)

	parser := NewParserDefault(tokens)
	assert.NoError(t, parser.Parse())
}

func parserTestAST(t *testing.T, text string) {
	tokenizer := NewTokenizerShort(text)
	tokens, err := tokenizer.ReadAll()
	assert.NoError(t, err)

	parser := NewParserDefault(tokens)
	assert.NoError(t, parser.Parse())
}

func parserTestError(t *testing.T, text string) {
	tokenizer := NewTokenizerShort(text)
	tokens, err := tokenizer.ReadAll()
	assert.NoError(t, err)

	parser := NewParserDefault(tokens)
	assert.Error(t, parser.Parse())
}

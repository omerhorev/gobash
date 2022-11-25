package gobash

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTokenOperators(t *testing.T) {
	testTokens(t, ">file", ">", "file")               // operator with multi-space
	testTokens(t, " >file", ">", "file")              // operator with multi-space
	testTokens(t, "ls>file", "ls", ">", "file")       // single char operator
	testTokens(t, "ls>>file", "ls", ">>", "file")     // multi char operator
	testTokens(t, "ls > file", "ls", ">", "file")     // operator with space
	testTokens(t, "ls   >   file", "ls", ">", "file") // operator with multi-space
}

func TestTokenWords(t *testing.T) {
	testTokens(t, "x yy zzz", "x", "yy", "zzz")       // words
	testTokens(t, "x yy    zzz   ", "x", "yy", "zzz") // words with multi-space
}
func TestTokenEscaping(t *testing.T) {
	testTokens(t, `x\ y`, `x\ y`)              // escape
	testTokens(t, `x\ y z\ t`, `x\ y`, `z\ t`) // multi-escape
	testTokens(t, `x\ yz\ t`, `x\ yz\ t`)      // multi-escape, same word
	testTokens(t, `x\\y`, `x\\y`)              // escape backslash
	testTokens(t, `x\>y`, `x\>y`)              // escape operator
	testTokens(t, `x>\>y`, `x`, `>`, `\>y`)    // escape operator, middle
}

func TestTokenQuotes(t *testing.T) {
	testTokens(t, `"x"`, `"x"`)                                     // quote
	testTokens(t, `"x y"`, `"x y"`)                                 // quote with space
	testTokens(t, `"x"'y''z' "a"'b'`, `"x"'y''z'`, `"a"'b'`)        // quote joined
	testTokens(t, `"x y""t" abc`, `"x y""t"`, "abc")                // quote with space, joined with another quote
	testTokens(t, `"abc \"t" a`, `"abc \"t"`, "a")                  // escaped quote inside a quote
	testTokens(t, `a\"y z`, `a\"y`, "z")                            // escaped quote
	testTokens(t, `abc ">" z`, `abc`, `">"`, `z`)                   // quoted operator
	testTokens(t, `0"1\"2\\\"3\\\"2\"1"0`, `0"1\"2\\\"3\\\"2\"1"0`) // quote inside quote
	testTokens(t, `ls "file with ' "`, `ls`, `"file with ' "`)      // quote inside quote
	testTokens(t, `ls 'file with " '`, `ls`, `'file with " '`)      // quote inside quote
	testTokens(t, `'\' 'x'`, `'\'`, `'x'`)                          // backslash inside quote
	testTokens(t, `"\"" 'x'`, `"\""`, `'x'`)                        // backslash inside quote

	// unterminated quote
	// _, err := s.tokenize(`"abc`)
	// assert.True(t, IsSyntaxError(err))
	// assert.ErrorIs(t, err, ErrUnterminatedQuotedString)

	// // unterminated quote
	// _, err = s.tokenize(`'abc`)
	// assert.True(t, IsSyntaxError(err))
	// assert.ErrorIs(t, err, ErrUnterminatedQuotedString)
}

func TestTokenExpressions(t *testing.T) {
	testTokens(t, "`x`", "`x`")
	testTokens(t, "`x `", "`x `")
	testTokens(t, "`x y`", "`x y`")
	testTokens(t, "`x 'y'`", "`x 'y'`")
	testTokens(t, "`x 'y\"y\"'`", "`x 'y\"y\"'`")

	// nested
	testTokens(t, "`\\`x\\``", "`\\`x\\``")
	testTokens(t, "`abc \\`a \\\\`b\\\\` c\\``", "`abc \\`a \\\\`b\\\\` c\\``")
}

func TestTokenMultiLine(t *testing.T) {
	testTokens(t, `ls y
cat x`, "ls", "y", "\n", "cat", "x")

	testTokens(t, `ls y
`, "ls", "y", "\n")

	testTokens(t, `ls y\
x`, "ls", "yx")
}

func testTokens(t *testing.T, line string, tokensStr ...string) {
	tokenizer := NewTokenizerShort(line)

	actualTokens, err := tokenizer.ReadAll()
	require.NoError(t, err)
	require.Equal(t, len(tokensStr)+1, len(actualTokens))

	for i := range tokensStr {
		require.Equal(t, tokensStr[i], actualTokens[i].Value)
	}

	require.True(t, actualTokens[len(tokensStr)].IsEOF())
}

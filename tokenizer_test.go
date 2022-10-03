package gobash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenOperators(t *testing.T) {
	s, _, _, _ := createTestShell("")

	testTokens(t, s, ">file", ">", "file")               // operator with multi-space
	testTokens(t, s, " >file", ">", "file")              // operator with multi-space
	testTokens(t, s, "ls>file", "ls", ">", "file")       // single char operator
	testTokens(t, s, "ls>>file", "ls", ">>", "file")     // multi char operator
	testTokens(t, s, "ls > file", "ls", ">", "file")     // operator with space
	testTokens(t, s, "ls   >   file", "ls", ">", "file") // operator with multi-space
}

func TestTokenWords(t *testing.T) {
	s, _, _, _ := createTestShell("")

	testTokens(t, s, "x yy zzz", "x", "yy", "zzz")       // words
	testTokens(t, s, "x yy    zzz   ", "x", "yy", "zzz") // words with multi-space
}
func TestTokenEscaping(t *testing.T) {
	s, _, _, _ := createTestShell("")

	testTokens(t, s, `x\ y`, `x\ y`)              // escape
	testTokens(t, s, `x\ y z\ t`, `x\ y`, `z\ t`) // multi-escape
	testTokens(t, s, `x\ yz\ t`, `x\ yz\ t`)      // multi-escape, same word
	testTokens(t, s, `x\\y`, `x\\y`)              // escape backslash
	testTokens(t, s, `x\>y`, `x\>y`)              // escape operator
	testTokens(t, s, `x>\>y`, `x`, `>`, `\>y`)    // escape operator, middle
}

func TestTokenQuotes(t *testing.T) {
	s, _, _, _ := createTestShell("")

	testTokens(t, s, `"x"`, `"x"`)                                     // quote
	testTokens(t, s, `"x y"`, `"x y"`)                                 // quote with space
	testTokens(t, s, `"x"'y''z' "a"'b'`, `"x"'y''z'`, `"a"'b'`)        // quote joined
	testTokens(t, s, `"x y""t" abc`, `"x y""t"`, "abc")                // quote with space, joined with another quote
	testTokens(t, s, `"abc \"t" a`, `"abc \"t"`, "a")                  // escaped quote inside a quote
	testTokens(t, s, `a\"y z`, `a\"y`, "z")                            // escaped quote
	testTokens(t, s, `abc ">" z`, `abc`, `">"`, `z`)                   // quoted operator
	testTokens(t, s, `0"1\"2\\\"3\\\"2\"1"0`, `0"1\"2\\\"3\\\"2\"1"0`) // quote inside quote
	testTokens(t, s, `ls "file with ' "`, `ls`, `"file with ' "`)      // quote inside quote
	testTokens(t, s, `ls 'file with " '`, `ls`, `'file with " '`)      // quote inside quote
	testTokens(t, s, `'\' 'x'`, `'\'`, `'x'`)                          // backslash inside quote
	testTokens(t, s, `"\"" 'x'`, `"\""`, `'x'`)                        // backslash inside quote

	// unterminated quote
	// _, err := s.tokenize(`"abc`)
	// assert.True(t, IsSyntaxError(err))
	// assert.ErrorIs(t, err, ErrUnterminatedQuotedString)

	// // unterminated quote
	// _, err = s.tokenize(`'abc`)
	// assert.True(t, IsSyntaxError(err))
	// assert.ErrorIs(t, err, ErrUnterminatedQuotedString)
}

func TestTokenMultiLine(t *testing.T) {
	s, _, _, _ := createTestShell("")
	testTokens(t, s, `ls y
cat x`, "ls", "y", "\n", "cat", "x")

	testTokens(t, s, `ls y
`, "ls", "y", "\n")

	testTokens(t, s, `ls y\
x`, "ls", "yx")
}

func testTokens(t *testing.T, s *Shell, line string, tokensStr ...string) {
	tokenizer := NewTokenizerShort(line)

	actualTokens, err := tokenizer.ReadAll()
	assert.NoError(t, err)
	assert.Equal(t, len(tokensStr)+1, len(actualTokens))

	for i := range tokensStr {
		assert.Equal(t, tokensStr[i], actualTokens[i].Value)
	}

	assert.True(t, actualTokens[len(tokensStr)].IsEOF())
}

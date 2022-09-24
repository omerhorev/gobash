package gobash

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// import (
// 	"bufio"
// 	"strings"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// func TestStrings(t *testing.T) {
// 	testSplit(t, "mv /path/1 /path/2", "mv", "/path/1", "/path/2")
// 	testSplit(t, "mv     /path/1 2", "mv", "/path/1", "2")
// 	testSplit(t, "\tmv     /path/1   /path/2    ", "mv", "/path/1", "/path/2")
// }

// func TestEmpty(t *testing.T) {
// 	testSplit(t, "")
// }

// func TestAllSpaces(t *testing.T) {
// 	testSplit(t, " ")
// 	testSplit(t, "\t")
// 	testSplit(t, "    \t ")
// }

// func TestStringsWithQuotes(t *testing.T) {
// 	testSplit(t, `mv "x" "y"`, `mv`, `x`, `y`)
// 	testSplit(t, `mv 'x' 'y'`, `mv`, `x`, `y`)
// 	testSplit(t, `'x"'`, `x"`)
// 	testSplit(t, `x'y'`, "xy")
// 	testSplit(t, `"x"'y'`, "xy")
// 	testSplit(t, `x'"y'`, `x"y`)
// 	testSplit(t, `'בדיקה' "בדיקה2" xxx`, `בדיקה`, "בדיקה2", "xxx")
// 	testSplit(t, `find "$DIR" -type f -atime +5 -exec rm {} \;`, "find", "$DIR", "-type", "f", "-atime", "+5", "-exec", "rm", "{}", ";")
// }

// func TestUntilRune(t *testing.T) {
// 	// with only short runes
// 	testScanUntilRune(t, "abc", rune('a'), "", false)
// 	testScanUntilRune(t, "abc", rune('b'), "a", false)
// 	testScanUntilRune(t, "abc", rune('c'), "ab", false)
// 	testScanUntilRune(t, "abc", rune('d'), "", true)

// 	// with longer runes
// 	testScanUntilRune(t, "aאbבcג", rune('a'), "", false)
// 	testScanUntilRune(t, "aאbבcג", rune('א'), "a", false)
// 	testScanUntilRune(t, "aאbבcג", rune('b'), "aא", false)
// 	testScanUntilRune(t, "aאbבcג", rune('ב'), "aאb", false)
// 	testScanUntilRune(t, "aאbבcג", rune('c'), "aאbב", false)
// 	testScanUntilRune(t, "aאbבcג", rune('ג'), "aאbבc", false)
// 	testScanUntilRune(t, "aאbבcג", rune('d'), "", true)
// 	testScanUntilRune(t, "aאbבcג", rune('ד'), "", true)
// }

// func testSplit(t *testing.T, sentence string, words ...string) {
// 	b := bufio.NewScanner(strings.NewReader(sentence))
// 	b.Split(SplitTokens)
// 	for i, w := range words {
// 		assert.True(t, b.Scan(), "b.Scan() returned false; ", i, w)
// 		assert.Equal(t, w, b.Text())
// 	}

// 	assert.False(t, b.Scan(), "b.Scan() returned true after end")
// }

// func testScanUntilRune(t *testing.T, sentence string, delimeter rune, expected string, notFound bool) {
// 	i := scanUntilRune([]byte(sentence), delimeter)
// 	if notFound {
// 		assert.Equal(t, -1, i, "sentence", "expected not found but scanned until", i)
// 	} else {
// 		assert.Equal(t, i, len(expected), sentence)
// 		assert.Equal(t, []byte(expected), []byte(sentence)[:i], sentence)
// 	}
// }

func TestToken(t *testing.T) {
	s, _, _, _ := createTestShell("")

	// operators
	testTokens(t, s, "ls", "ls")
	testTokens(t, s, "ls>file", "ls", ">", "file")       // single char operator
	testTokens(t, s, "ls>>file", "ls", ">>", "file")     // multi char operator
	testTokens(t, s, "ls > file", "ls", ">", "file")     // operator with space
	testTokens(t, s, "ls   >   file", "ls", ">", "file") // operator with multi-space

	// words
	testTokens(t, s, "x yy zzz", "x", "yy", "zzz")       // words
	testTokens(t, s, "x yy    zzz   ", "x", "yy", "zzz") // words with multi-space

	// escaping
	testTokens(t, s, `x\ y`, `x\ y`)              // escape
	testTokens(t, s, `x\ y z\ t`, `x\ y`, `z\ t`) // multi-escape
	testTokens(t, s, `x\ yz\ t`, `x\ yz\ t`)      // multi-escape, same word
	testTokens(t, s, `x\\y`, `x\\y`)              // escape backslash
	testTokens(t, s, `x\>y`, `x\>y`)              // escape operator
	testTokens(t, s, `x>\>y`, `x`, `>`, `\>y`)    // escape operator, middle

	// quotes
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
	_, err := s.tokenizeLine(`"abc`)
	assert.True(t, IsSyntaxError(err))
	assert.ErrorIs(t, err, ErrUnterminatedQuotedString)

	// unterminated quote
	_, err = s.tokenizeLine(`'abc`)
	assert.True(t, IsSyntaxError(err))
	assert.ErrorIs(t, err, ErrUnterminatedQuotedString)
}

func TestRun(t *testing.T) {
	s, _, _, _ := createTestShell("")
	s.RunStringContext(context.Background(), "ls")
}

func testTokens(t *testing.T, s *Shell, line string, tokens ...string) {
	foundTokens, err := s.tokenizeLine(line)
	assert.NoError(t, err)
	assert.Equal(t, append([]string{}, tokens...), foundTokens)
}

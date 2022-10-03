package gobash

import (
	"context"
	"strings"
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

func TestRdp(t *testing.T) {
	s, _, _, _ := createTestShell("")
	// noSyntaxError(t, s, "")
	// noSyntaxError(t, s, "\n")
	// noSyntaxError(t, s, "a")
	// noSyntaxError(t, s, "a; b")
	// noSyntaxError(t, s, "a\nb")
	noSyntaxError(t, s, "a &&")
	// noSyntaxError(t, s, "a | b")
}

func noSyntaxError(t *testing.T, s *Shell, text string) {
	assert.NoError(t, s.RunContext(context.Background(), strings.NewReader(text)))
}

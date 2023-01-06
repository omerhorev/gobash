package utils

import (
	"bufio"
	"io"
	"unicode/utf8"

	"golang.org/x/exp/slices"
)

// Creates a new bufio.Scanner that is configured to scan words separated by
// one of the runes provided
func NewRunesScanner(reader io.Reader, runes []rune) *bufio.Scanner {
	s := bufio.NewScanner(reader)
	// s.Split(bufio.ScanWords)
	s.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// Skip leading spaces.
		start := 0
		for width := 0; start < len(data); start += width {
			var r rune
			r, width = utf8.DecodeRune(data[start:])

			if !slices.Contains(runes, r) {
				break
			}
		}
		// Scan until space, marking end of word.
		for width, i := 0, start; i < len(data); i += width {
			var r rune
			r, width = utf8.DecodeRune(data[i:])
			if slices.Contains(runes, r) {
				return i + width, data[start:i], nil
			}
		}
		// If we're at EOF, we have a final, non-empty, non-terminated word. Return it.
		if atEOF && len(data) > start {
			return len(data), data[start:], nil
		}
		// Request more data.
		return start, nil, nil
	})

	return s
}

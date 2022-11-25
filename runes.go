package gobash

import (
	"unicode"
)

// Returns whether the character is a NL newline.
func isNewLine(r rune) bool {
	return r == '\n'
}

// Returns whether the character is an IEEE1003.1 escape code (backslash).
func isEscape(r rune) bool {
	return r == '\\'
}

// Returns whether the character is an apostrophe U+0027.
func isApostrophe(r rune) bool {
	return r == '\''
}

// Returns whether the character is a backtick or a dollar sign.
func isExpressionStart(r rune) bool {
	return r == '$' || r == '`'
}

// Returns whether the character is a dollar sign.
func isDollarSign(r rune) bool {
	return r == '$'
}

// Returns whether the character is a quotation mark U+0022.
func isQuotationMark(r rune) bool {
	return r == '"'
}

// Returns whether the character is a number sign U+0023.
func isNumberSign(r rune) bool {
	return r == '#'
}

// Returns whether a string is an unsigned number.
func isStringNumber(str string) bool {
	for _, c := range str {
		if !unicode.IsNumber(c) {
			return false
		}
	}

	return true
}

// Returns whether the rune is tab or space (not newline).
func isBlank(r rune) bool {
	return unicode.IsSpace(r) && !isNewLine(r)
}

// Returns whether the rune can be used in a name (digits, alphabet and underscore).
func isNameRune(r rune) bool {
	return isAlphabetLetter(r) || isDigit(r) || isUnderscore(r)
}

// Returns whether the rune is in the english alphabet (lower and upper).
func isAlphabetLetter(r rune) bool {
	return !((r < 'a' || r > 'z') && (r < 'A' || r > 'Z'))
}

// Returns whether the rune is a digit.
func isDigit(r rune) bool {
	return unicode.IsDigit(r)
}

// Returns whether the rune is an underscore U+005F.
func isUnderscore(r rune) bool {
	return r == '_'
}

// Returns whether the rune is an equal sign U+003D.
func isEqualSign(r rune) bool {
	return r == '='
}

func nextUnescaped(str string, expected rune) int {
	isBackslashed := false
	isApostrophed := false
	isQuotationMarked := false

	preserveMeaning := func() bool {
		return isBackslashed || isApostrophed || isQuotationMarked
	}

	for i, r := range str {
		if !preserveMeaning() && expected == r {
			return i
		}

		if !(isBackslashed || isApostrophed) && isEscape(r) {
			isBackslashed = true
			continue
		}

		if !preserveMeaning() && !isQuotationMarked && isQuotationMark(r) {
			isQuotationMarked = true
			continue
		}

		if !isBackslashed && isQuotationMarked && isQuotationMark(r) {
			isQuotationMarked = false
			continue
		}

		if !preserveMeaning() && !isApostrophed && isApostrophe(r) {
			isApostrophed = true
			continue
		}

		if isApostrophed && isApostrophe(r) {
			isApostrophed = false
			continue
		}

		// add to word
		isBackslashed = false
	}

	return -1
}

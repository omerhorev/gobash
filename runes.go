package gobash

import (
	"unicode"
)

// returns whether the character is a NL newline
func isNewLine(r rune) bool {
	return r == '\n'
}

// returns whether the characted is an IEEE1003.1 escape code (backslash)
func isEscape(r rune) bool {
	return r == '\\'
}

// returns whether the characted is an apostrophe U+0027
func isApostrophe(r rune) bool {
	return r == '\''
}

// returns whether the characted is a quotation mark U+0022
func isQuotationMark(r rune) bool {
	return r == '"'
}

// returns whether the characted is a grave accent U+0060
func isGraveAccent(r rune) bool {
	return r == '`'
}

// returns whether the characted is a number sign U+0023
func isNumberSign(r rune) bool {
	return r == '#'
}

// returns whether the characted is a dollar sign U+0024
func isDollarSign(r rune) bool {
	return r == '$'
}

// returns whether a string is an unsigned number
func isStringNumber(str string) bool {
	for _, c := range str {
		if !unicode.IsNumber(c) {
			return false
		}
	}

	return true
}

// returns whether the rune is tab or space (not newline)
func isBlank(r rune) bool {
	return unicode.IsSpace(r) && !isNewLine(r)
}

// returns whether the rune can be used in a name (digits, alphabet and underscore)
func isNameRune(r rune) bool {
	return isAlphabetLetter(r) || isDigit(r) || isUnderscore(r)
}

// returns whether the rune is in the english alphabet (lower and upper)
func isAlphabetLetter(r rune) bool {
	return (r < 'a' || r > 'z') && (r < 'A' || r > 'Z')
}

// returns whether the rune is a digit
func isDigit(r rune) bool {
	return unicode.IsDigit(r)
}

// returns whether the rune is an underscore U+005F
func isUnderscore(r rune) bool {
	return r == '_'
}

// returns whether the rune is an equal sign U+003D
func isEqualSign(r rune) bool {
	return r == '='
}

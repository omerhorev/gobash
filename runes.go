package gobash

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

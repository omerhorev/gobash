package gobash

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	defaultShortTokensCount = 20
	defaultLongTokensCount  = 200
)

// used to alter the behavior of the tokenizer
type TokenizerSettings struct {
	// used for tokens memory allocation
	// can speed up the tokenizing process in expense of memory usage
	ExpectedTokensCount int
}

// tokenizer is a structure that receives a stream and produces tokens
// it is used to parse shell scripts and expressions into tokens
// that can be used in grammar
type Tokenizer struct {
	reader   *bufio.Reader
	settings TokenizerSettings
	err      error
}

// Create a tokenizer that is optimized for short expressions (usually received
// from terminal)
func NewTokenizerShort(text string) *Tokenizer {
	return &Tokenizer{
		settings: TokenizerSettings{
			ExpectedTokensCount: defaultShortTokensCount,
		},

		reader: bufio.NewReader(strings.NewReader(text)),
		err:    nil,
	}
}

// Create a tokenizer that is optimized for long expressions (usually received
// from a script)
func NewTokenizerLong(reader io.Reader) *Tokenizer {
	return &Tokenizer{
		settings: TokenizerSettings{
			ExpectedTokensCount: defaultLongTokensCount,
		},

		reader: bufio.NewReader(reader),
		err:    nil,
	}
}

// Create a tokenizer that is optimized for long expressions (usually received
// from a script)
func NewTokenizer(reader io.Reader, settings TokenizerSettings) *Tokenizer {
	return &Tokenizer{
		settings: settings,
		reader:   bufio.NewReader(reader),
		err:      nil,
	}
}

// reads the next token from the stream and return it.
//
// when the stream reaches the end, it will produce another EOF token without an error.
// the next call will return an EOF error.
func (t *Tokenizer) ReadToken() (*Token, error) {
	token, err := t.readToken()
	if err != nil {
		t.err = err
		return nil, err
	}

	if token.IsEOF() {
		// the next call will produce an io.EOF error
		t.err = io.EOF
	}

	return token, nil
}

// reads all tokens from the reader until EOF is reached.
//
// returns an error only if one rises from the reader. if EOF stops reading tokens
// and without returning an error
func (t *Tokenizer) ReadAll() ([]*Token, error) {
	if t.err != nil {
		return nil, t.err
	}

	tokens := make([]*Token, 0, t.settings.ExpectedTokensCount)

	for {
		token, err := t.ReadToken()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}

		tokens = append(tokens, token)
	}

	return tokens, nil
}

func (t *Tokenizer) readToken() (*Token, error) {
	if t.err != nil {
		return nil, t.err
	}

	isInOperator := false
	isBackslashed := false
	isApostrophed := false
	isQuotationMarked := false

	preserveMeaning := func() bool {
		return isBackslashed || isApostrophed || isQuotationMarked
	}

	tokenStr := ""

	var r = utf8.RuneError
	var err = error(nil)

	// peek the first rune. If its a space, we can skip it.
	// only start tokenizing in the first non-space rune
	for {
		r, _, err = t.reader.ReadRune()
		if err == io.EOF {
			return newTokenFromString("", utf8.RuneError), nil
		} else if err != nil {
			return nil, err
		}

		if isNewLine(r) {
			return newTokenFromString("\n", r), nil
		}

		if !isBlank(r) {
			if err := t.reader.UnreadRune(); err != nil {
				return nil, err
			}

			break
		}
	}

	isInOperator = canBeUsedInOperator(string(r))

	for {
		r, _, err := t.reader.ReadRune()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if !preserveMeaning() && isInOperator {
			if canBeUsedInOperator(tokenStr + string(r)) {
				// rule #2 - the next rune can be used in an operator
				// add()
				tokenStr += string(r)
				continue
			} else {
				// rule #3 - the next rune cannot be used in an operator
				// keep the next rune
				if err := t.reader.UnreadRune(); err != nil {
					return nil, err
				}

				return newTokenFromString(tokenStr, r), nil
			}
		}

		// rule #4 - backslash
		// backslash is not applied when already backslashing or inside a double quote
		if !(isBackslashed || isApostrophed) && isEscape(r) {
			isBackslashed = true
			tokenStr += string(r)
			continue
		}

		// rule #4
		if !preserveMeaning() && !isQuotationMarked && isQuotationMark(r) {
			isQuotationMarked = true
			tokenStr += string(r)
			continue
		}

		if !isBackslashed && isQuotationMarked && isQuotationMark(r) {
			isQuotationMarked = false
			tokenStr += string(r)
			continue
		}

		if !preserveMeaning() && !isApostrophed && isApostrophe(r) {
			isApostrophed = true
			tokenStr += string(r)
			continue
		}

		if !preserveMeaning() && isExpressionStart(r) {
			t.reader.UnreadRune()

			expr, err := t.readExpression()
			if err != nil {
				return nil, err
			}

			tokenStr += expr
		}

		if isApostrophed && isApostrophe(r) {
			isApostrophed = false
			tokenStr += string(r)
			continue
		}

		// rule #6, new operator
		if !preserveMeaning() && canBeUsedInOperator(string(r)) {
			// The rune will linger on the next call to nextToken()
			if err := t.reader.UnreadRune(); err != nil {
				return nil, err
			}

			return newTokenFromString(tokenStr, r), nil
			// add()
			// token += string(r)
			// isInOperator = true
			// continue
		}

		// line joining
		if !(isQuotationMarked || isApostrophed) && isBackslashed && isNewLine(r) {
			// we know for sure the token ends with '\' (isBackslashed)
			// remove the backslash from the token
			tokenStr = tokenStr[:len(tokenStr)-1]
			continue
		}

		// rule 7#, spaces
		if !preserveMeaning() && unicode.IsSpace(r) {
			// skip multiple spaces
			if tokenStr != "" {
				if err := t.reader.UnreadRune(); err != nil {
					return nil, err
				}

				return newTokenFromString(tokenStr, r), nil
			} else {
				// discard the space
				continue
			}
		}

		// rule #8, comment
		if !preserveMeaning() && isNumberSign(r) {
			break
		}

		if !preserveMeaning() && isNewLine(r) {
			// delim(r)
			// the newline will linger on to the next call to nextToken
			if err := t.reader.UnreadRune(); err != nil {
				return nil, err
			}

			return newTokenFromString(tokenStr, r), nil
			// token += string(r)
		}

		// add to word
		isBackslashed = false
		tokenStr += string(r)
	}

	if isApostrophed || isQuotationMarked {
		return nil, newSyntaxError(errors.New("unterminated quoted string"))
	}

	return newTokenFromString(tokenStr, utf8.RuneError), nil
}

func (t *Tokenizer) readExpression() (string, error) {
	r, _, err := t.reader.ReadRune()
	if err != nil {
		return "", err
	}

	if r == '`' {
		result, err := t.readUntilUnescaped("`")
		if err != nil {
			return "", err
		}

		return "`" + result, nil
	}

	return "", errors.New("unexpected expression")
}

func (t *Tokenizer) readUntilUnescaped(suffix string) (string, error) {
	if t.err != nil {
		return "", t.err
	}

	isBackslashed := false
	isApostrophed := false
	isQuotationMarked := false

	preserveMeaning := func() bool {
		return isBackslashed || isApostrophed || isQuotationMarked
	}

	str := ""

	for {
		r, _, err := t.reader.ReadRune()
		if err == io.EOF {
			return "", io.ErrUnexpectedEOF
		} else if err != nil {
			return "", err
		}

		if !preserveMeaning() && strings.HasSuffix(str+string(r), suffix) {
			return str, nil
		}

		if !(isBackslashed || isApostrophed) && isEscape(r) {
			isBackslashed = true
			str += string(r)
			continue
		}

		if !preserveMeaning() && !isQuotationMarked && isQuotationMark(r) {
			isQuotationMarked = true
			str += string(r)
			continue
		}

		if !isBackslashed && isQuotationMarked && isQuotationMark(r) {
			isQuotationMarked = false
			str += string(r)
			continue
		}

		if !preserveMeaning() && !isApostrophed && isApostrophe(r) {
			isApostrophed = true
			str += string(r)
			continue
		}

		if !preserveMeaning() && isExpressionStart(r) {
			t.reader.UnreadRune()

			expr, err := t.readExpression()
			if err != nil {
				return "", err
			}

			str += expr
		}

		if isApostrophed && isApostrophe(r) {
			isApostrophed = false
			str += string(r)
			continue
		}

		// add to word
		isBackslashed = false
		str += string(r)
	}
}

func canBeUsedInOperator(token string) bool {
	for _, o := range operatorsStrings {
		if strings.HasPrefix(o, token) {
			return true
		}
	}

	return false
}

package gobash

// A helper structure used to expand words
type Expander struct {
	str string
	i   int
}

func accept() bool {
	return true
}

// Creates a new expander object
func NewExpander() *Expander {
	return &Expander{}
}

// returns the next instance of an unescaped, unquoted rune in a string

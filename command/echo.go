package command

import (
	"flag"
	"strconv"
)

// like /bin/echo
var Echo = &SimpleMatchCommand{
	Name: "echo",
	F: func(args []string, e *Env) int {
		if len(args) == 0 {
			return 1
		}

		flags := flag.NewFlagSet("echo", flag.ContinueOnError)
		noNewLine := flags.Bool("n", false, "do not output the trailing newline")
		escape := flags.Bool("e", false, "enable interpretation of backslash escapes")

		flags.Usage = func() {}

		if flags.Parse(args[1:]) != nil {
			return 1
		}

		doEcho(e, flags.Args(), *noNewLine, *escape)
		return 0
	},
}

func doEcho(e *Env, s []string, noNewLine bool, escape bool) {
	if len(s) > 0 {
		for _, arg := range s[:len(s)-1] {
			if escape {
				arg = doEscape(arg)
			}
			e.Print(arg, " ")
		}

		if escape {
			s[len(s)-1] = doEscape(s[len(s)-1])
		}
		e.Print(s[len(s)-1])
	}

	if !noNewLine {
		e.Println()
	}
}

func doEscape(s string) string {
	i := 0
	for i+1 < len(s) {
		if s[i] == '\\' {
			switch s[i+1] {
			case '\\':
				s = s[:i] + `\` + s[i+2:]
			case 'a':
				s = s[:i] + "\x07" + s[i+2:]
			case 'b':
				s = s[:i] + "\x07" + s[i+2:]
			case 'c':
				s = s[:i]
			case 'n':
				s = s[:i] + "\n" + s[i+2:]
			case 'r':
				s = s[:i] + "\r" + s[i+2:]
			case 't':
				s = s[:i] + "\t" + s[i+2:]
			case 'v':
				s = s[:i] + "\v" + s[i+2:]
			case '0':
				j := 0
				for j = 0; j < 3; j++ {
					if i+j+1 >= len(s) {
						break
					}
					r := s[i+1+j]
					if r < '0' || r > '7' {
						break
					}
				}
				n, _ := strconv.ParseUint(s[i+1:i+1+j], 8, 32)
				r := rune(n)
				s = s[:i] + string(r) + s[i+1+j:]
			case 'x':
				j := 0
				for j = 0; j < 2; j++ {
					if i+j+2 >= len(s) {
						break
					}
					r := s[i+2+j]
					if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
						break
					}
				}
				n, _ := strconv.ParseUint(s[i+2:i+2+j], 16, 32)
				r := rune(n)
				s = s[:i] + string(r) + s[i+2+j:]
			}
		}
		i++
	}

	// 	s = strings.Replace(s, "\\\\", "\\", -1)
	// 	s = strings.Replace(s, "\\a", "\a", -1)
	// 	s = strings.Replace(s, "\\b", "\b", -1)
	// 	s = strings.Replace(s, "\\f", "\f", -1)
	// 	s = strings.Replace(s, "\\n", "\n", -1)
	// 	s = strings.Replace(s, "\\r", "\r", -1)
	// 	s = strings.Replace(s, "\\t", "\t", -1)
	// 	s = strings.Replace(s, "\\v", "\v", -1)
	// 	s = strings.Replace(s, "\\v", "\v", -1)

	// 	if len(s) > 1 {
	// 	for i := range s[:len(s)-2] {
	// 		if s[i] == '\\' && s[i+1]
	// 	}
	// }

	// 	if endIndex := strings.Index(s, "\\c"); endIndex > 0 {
	// 		s = s[:endIndex]
	// 	}

	return s
}

package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
)

// Tok is one token: Kind 'w' word, 'o' operator, 'n' IO number.
type Tok struct {
	Kind   byte
	Text   string
	Quoted bool
}

// Operators, longest first. A real tokenizer must find these without spaces around them.
var ops = []string{"<<-", "&&", "||", ">>", "<<", ">&", "<&", "|", "&", ";", "<", ">"}

// checkIsOp reports whether s is a backslash followed by a character that the
// backslash can escape: an operator character, a quote, $, #, `, a space or
// another backslash.
func checkIsOp(s string) bool {
	if len(s) < 2 || s[0] != '\\' {
		return false
	}
	return strings.ContainsRune("|&;<>'\"$#` \\", rune(s[1]))
}

// tokenize splits one command line. The error string is empty on success, or
// "unterminated quote" / "trailing backslash".
func tokenize(line string) ([]Tok, string) {
	// TODO: replace this space split with a character-by-character walk that handles
	// quotes, backslashes, operators, IO numbers and comments.
	var toks []Tok
	var doubleQuoteFlag bool
	var singleQuoteFlag bool
	var word string
	var precedingOp bool
	var skip int
	var noPower bool
	var opFlag bool

	for i, char := range line {
		if skip > 0 {
			skip -= 1
			continue
		}

		if char == '\'' && singleQuoteFlag {
			if noPower {
				word += string(char)
				noPower = false
				continue
			}
			singleQuoteFlag = false
		} else if char == '\'' && !singleQuoteFlag && !doubleQuoteFlag {
			if noPower {
				word += string(char)
				noPower = false
				continue
			}
			singleQuoteFlag = true
		} else if char == '"' && doubleQuoteFlag {
			if noPower {
				word += string(char)
				noPower = false
				continue
			}
			doubleQuoteFlag = false
		} else if char == '"' && !singleQuoteFlag && !doubleQuoteFlag {
			if noPower {
				word += string(char)
				noPower = false
				continue
			}
			doubleQuoteFlag = true
		} else if (char == ' ' || char == '\t') && !singleQuoteFlag && !doubleQuoteFlag {

			if !precedingOp {
				toks = append(toks, Tok{'w', word, false})
				word = ""
			} else {
				precedingOp = false
			}
		} else if !singleQuoteFlag && !doubleQuoteFlag && slices.Contains(ops, string(char)) {
			if noPower {
				word += string(char)
				noPower = false
				continue
			}

			if i+3 <= len(line) && slices.Contains(ops, line[i:i+3]) {
				skip = 2
			} else if i+2 <= len(line) && slices.Contains(ops, line[i:i+2]) {
				skip = 1
			}

			if line[i : i+skip+1][0] == '>' || line[i : i+skip+1][0] == '<' {
				if word != "" {
					_, err := strconv.Atoi(word)
					if err == nil {
						if !strings.ContainsAny(line[i-len(word):i], "\"'") {

							toks = append(toks, Tok{'i', word, false})
							word = ""
						}
					}

				}
			}

			if word != "" {
				toks = append(toks, Tok{'w', word, false})
				word = ""
			}

			toks = append(toks, Tok{'o', line[i : i+skip+1], false})

			opFlag = true
			precedingOp = true

		} else if word == "" && !singleQuoteFlag && !doubleQuoteFlag && (char == '#') {
			if noPower {
				word += string(char)
				noPower = false
				continue
			}
			return toks, ""
		} else if char == '\\' {

			if i+2 <= len(line) && checkIsOp(line[i:i+2]) {
				noPower = true
				continue
			} else {
				word += string(char)
			}

		} else if char == '$' {
			if noPower {
				word += string(char)
				noPower = false
				continue
			} else {
				word += string(char)
			}
		} else {
			if !doubleQuoteFlag && !singleQuoteFlag && char == ' ' {
				continue
			}
			word += string(char)
		}

		if opFlag {
			opFlag = false
		} else {
			precedingOp = false
		}

		// for _, char := range w {
		// 	if char == '\'' && singleQuoteFlag {
		// 		singleQuoteFlag = false
		// 	} else if char == '\'' && !singleQuoteFlag && !doubleQuoteFlag {
		// 		singleQuoteFlag = true
		// 	} else if char == '"' && doubleQuoteFlag {
		// 		doubleQuoteFlag = false
		// 	} else if char == '"' && !singleQuoteFlag && !doubleQuoteFlag {
		// 		doubleQuoteFlag = true
		// 	} else {
		// 		word += string(char)
		// 	}
		// }
		// if !singleQuoteFlag && !doubleQuoteFlag {
		// 	toks = append(toks, Tok{'w', word, false})
		// }

	}
	if doubleQuoteFlag || singleQuoteFlag {
		return nil, "unterminated quote"
	}

	if len(word) > 0 && word[len(word)-1] == '\\' {
		return nil, "trailing backslash"
	}
	if line == "" {
		return toks, ""
	}
	toks = append(toks, Tok{'w', word, false})
	return toks, ""
}

// fmtTok shows one token the way the lessons do: [word] (operator) {io-number}.
func fmtTok(t Tok) string {
	if t.Kind == 'w' {
		return "[" + t.Text + "]"
	}
	if t.Kind == 'o' {
		return "(" + t.Text + ")"
	}
	return "{" + t.Text + "}"
}

func main() {
	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		toks, err := tokenize(line)
		if err != "" {
			fmt.Fprintln(w, "ERR "+err)
			continue
		}
		if len(toks) == 0 {
			continue
		}
		parts := make([]string, len(toks))
		for i, t := range toks {
			parts[i] = fmtTok(t)
		}
		fmt.Fprintln(w, strings.Join(parts, " "))
	}
}

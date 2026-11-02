package main

import (
	"bufio"
	"fmt"
	"os"
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

// tokenize splits one command line. The error string is empty on success, or
// "unterminated quote" / "trailing backslash".
func tokenize(line string) ([]Tok, string) {
	// TODO: replace this space split with a character-by-character walk that handles
	// quotes, backslashes, operators, IO numbers and comments.
	var toks []Tok
	for _, w := range strings.Fields(line) {
		toks = append(toks, Tok{'w', w, false})
	}
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

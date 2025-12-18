package eval

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

func stripCommas(s string) string {
	if !strings.Contains(s, ",") {
		return s
	}
	return strings.ReplaceAll(s, ",", "")
}

func Lex(input string) ([]Token, error) {
	l := &lexer{s: normalize(input)}
	var toks []Token
	for {
		tok, err := l.next()
		if err != nil {
			return nil, err
		}
		toks = append(toks, tok)
		if tok.Kind == tokEOF {
			return toks, nil
		}
	}
}

func normalize(s string) string {
	s = strings.TrimSpace(s)
	repl := strings.NewReplacer(
		"×", "*",
		"−", "-",
		"–", "-",
		"—", "-",
	)
	s = repl.Replace(s)
	return normalizeMulX(s)
}

func normalizeMulX(s string) string {
	// Convert 'x' / 'X' to '*' only when it looks like a binary operator.
	// This avoids breaking identifiers like "max".
	var b strings.Builder
	b.Grow(len(s))

	n := len(s)
	for i := 0; i < n; {
		r, size := utf8.DecodeRuneInString(s[i:])
		if (r == 'x' || r == 'X') && isMulContext(s, i) {
			b.WriteByte('*')
			i += size
			continue
		}
		b.WriteRune(r)
		i += size
	}

	return b.String()
}

func isMulContext(s string, idx int) bool {
	left := prevNonSpaceRune(s, idx)
	right := nextNonSpaceRune(s, idx+1)
	if left == 0 || right == 0 {
		return false
	}

	leftOk := unicode.IsDigit(left) || left == ')' || left == '%' || left == '$' || left == '.'
	rightOk := unicode.IsDigit(right) || right == '(' || right == '$' || right == '.' || right == '\\' || unicode.IsLetter(right)
	return leftOk && rightOk
}

func prevNonSpaceRune(s string, idx int) rune {
	for i := idx - 1; i >= 0; {
		r, size := utf8.DecodeLastRuneInString(s[:i+1])
		if r == utf8.RuneError && size == 1 {
			return 0
		}
		if !unicode.IsSpace(r) {
			return r
		}
		i -= size
	}
	return 0
}

func nextNonSpaceRune(s string, idx int) rune {
	for i := idx; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			return 0
		}
		if !unicode.IsSpace(r) {
			return r
		}
		i += size
	}
	return 0
}

func (l *lexer) advance(n int) {
	l.i += n
}

func (l *lexer) skipSpaces() {
	for l.i < len(l.s) {
		r, size := utf8.DecodeRuneInString(l.s[l.i:])
		if r == 0 || !unicode.IsSpace(r) {
			return
		}
		l.i += size
	}
}

func (l *lexer) next() (Token, error) {
	l.skipSpaces()
	if l.i >= len(l.s) {
		return Token{Kind: tokEOF}, nil
	}

	r, size := utf8.DecodeRuneInString(l.s[l.i:])
	switch r {
	case '+':
		l.advance(size)
		return Token{Kind: tokPlus, Text: "+"}, nil
	case '-':
		l.advance(size)
		return Token{Kind: tokMinus, Text: "-"}, nil
	case '*':
		l.advance(size)
		return Token{Kind: tokMul, Text: "*"}, nil
	case '/':
		l.advance(size)
		return Token{Kind: tokDiv, Text: "/"}, nil
	case '^':
		l.advance(size)
		return Token{Kind: tokPow, Text: "^"}, nil
	case '>':
		l.advance(size)
		// Check for >=
		if l.i < len(l.s) && l.s[l.i] == '=' {
			l.advance(1)
			return Token{Kind: tokGTE, Text: ">="}, nil
		}
		return Token{Kind: tokGT, Text: ">"}, nil
	case '<':
		l.advance(size)
		// Check for <=
		if l.i < len(l.s) && l.s[l.i] == '=' {
			l.advance(1)
			return Token{Kind: tokLTE, Text: "<="}, nil
		}
		return Token{Kind: tokLT, Text: "<"}, nil
	case '=':
		l.advance(size)
		// Check for ==
		if l.i < len(l.s) && l.s[l.i] == '=' {
			l.advance(1)
			return Token{Kind: tokEQ, Text: "=="}, nil
		}
		// Single = is not a valid operator in expressions, skip it
		return l.next()
	case '!':
		l.advance(size)
		// Check for !=
		if l.i < len(l.s) && l.s[l.i] == '=' {
			l.advance(1)
			return Token{Kind: tokNE, Text: "!="}, nil
		}
		return Token{}, fmt.Errorf("unexpected '!'")
	case '(':
		l.advance(size)
		return Token{Kind: tokLParen, Text: "("}, nil
	case ')':
		l.advance(size)
		return Token{Kind: tokRParen, Text: ")"}, nil
	case '\\':
		// Line reference: \\1, \\2, ...
		l.advance(size)
		start := l.i
		for l.i < len(l.s) {
			r2, s2 := utf8.DecodeRuneInString(l.s[l.i:])
			if !unicode.IsDigit(r2) {
				break
			}
			l.i += s2
		}
		if start == l.i {
			return Token{}, fmt.Errorf("unexpected '\\\\'")
		}
		refStr := l.s[start:l.i]
		ref, err := strconv.Atoi(refStr)
		if err != nil {
			return Token{}, err
		}
		return Token{Kind: tokRef, Text: "\\\\" + refStr, Ref: ref}, nil
	case '$':
		// Currency prefix: $95.88 -> 95.88
		l.advance(size)
		start := l.i
		if start >= len(l.s) {
			return Token{}, fmt.Errorf("unexpected '$'")
		}
		r0, _ := utf8.DecodeRuneInString(l.s[l.i:])
		if !(unicode.IsDigit(r0) || r0 == '.') {
			return Token{}, fmt.Errorf("unexpected '$'")
		}
		dotSeen := r0 == '.'
		_, s0 := utf8.DecodeRuneInString(l.s[l.i:])
		l.i += s0
		for l.i < len(l.s) {
			r2, s2 := utf8.DecodeRuneInString(l.s[l.i:])
			if unicode.IsDigit(r2) {
				l.i += s2
				continue
			}
			if r2 == ',' {
				l.i += s2
				continue
			}
			if r2 == '.' && !dotSeen {
				dotSeen = true
				l.i += s2
				continue
			}
			break
		}
		n, err := strconv.ParseFloat(stripCommas(l.s[start:l.i]), 64)
		if err != nil {
			return Token{}, err
		}
		return Token{Kind: tokNumber, Text: "$" + l.s[start:l.i], Num: n}, nil
	}

	if unicode.IsDigit(r) || r == '.' {
		start := l.i
		dotSeen := r == '.'
		l.advance(size)
		for l.i < len(l.s) {
			r2, s2 := utf8.DecodeRuneInString(l.s[l.i:])
			if unicode.IsDigit(r2) {
				l.i += s2
				continue
			}
			if r2 == ',' {
				l.i += s2
				continue
			}
			if r2 == '.' && !dotSeen {
				dotSeen = true
				l.i += s2
				continue
			}
			break
		}
		n, err := strconv.ParseFloat(stripCommas(l.s[start:l.i]), 64)
		if err != nil {
			return Token{}, err
		}
		if l.i < len(l.s) {
			r2, s2 := utf8.DecodeRuneInString(l.s[l.i:])
			if r2 == '%' {
				l.i += s2
				n = n / 100.0
				return Token{Kind: tokNumber, Text: l.s[start:l.i], Num: n, Pct: true}, nil
			}
		}
		return Token{Kind: tokNumber, Text: l.s[start:l.i], Num: n}, nil
	}

	if unicode.IsLetter(r) {
		start := l.i
		l.advance(size)
		for l.i < len(l.s) {
			r2, s2 := utf8.DecodeRuneInString(l.s[l.i:])
			if unicode.IsLetter(r2) || unicode.IsDigit(r2) || r2 == '_' {
				l.i += s2
				continue
			}
			break
		}
		return Token{Kind: tokIdent, Text: strings.ToLower(l.s[start:l.i])}, nil
	}

	return Token{}, fmt.Errorf("unexpected character: %q", r)
}

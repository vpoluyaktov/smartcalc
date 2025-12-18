package eval

type TokenKind int

const (
	tokEOF TokenKind = iota
	tokNumber
	tokIdent
	tokRef
	tokPlus
	tokMinus
	tokMul
	tokDiv
	tokPow
	tokLParen
	tokRParen
	tokGT  // >
	tokLT  // <
	tokGTE // >=
	tokLTE // <=
	tokEQ  // ==
	tokNE  // !=
)

type Token struct {
	Kind TokenKind
	Text string
	Num  float64
	Ref  int
	Pct  bool
}

type lexer struct {
	s string
	i int
}

type parser struct {
	toks []Token
	pos  int
	refs func(n int) (float64, error)
}

type val struct {
	v   float64
	pct bool // true only if the entire expression is a percent literal, like 20%
}

// Pratt parser precedence
const (
	precCmp = 5 // comparison operators (lowest precedence)
	precAdd = 10
	precMul = 20
	precPow = 30
)

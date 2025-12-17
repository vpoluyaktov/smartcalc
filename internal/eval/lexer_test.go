package eval

import (
	"testing"
)

func TestStripCommas(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1000", "1000"},
		{"1,000", "1000"},
		{"1,000,000", "1000000"},
		{"1,234.56", "1234.56"},
		{"no commas here", "no commas here"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := stripCommas(tt.input)
			if result != tt.expected {
				t.Errorf("stripCommas(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"2 × 3", "2 * 3"},
		{"5 − 2", "5 - 2"},
		{"10 – 5", "10 - 5"},
		{"20 — 10", "20 - 10"},
		{"  spaces  ", "spaces"},
		{"2 x 3", "2 * 3"},
		{"max(5)", "max(5)"},   // 'x' in identifier should not be replaced
		{"2 x \\1", "2 * \\1"}, // 'x' before reference should be replaced
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalize(tt.input)
			if result != tt.expected {
				t.Errorf("normalize(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLexNumbers(t *testing.T) {
	tests := []struct {
		input       string
		expectedNum float64
		expectedPct bool
	}{
		{"42", 42, false},
		{"3.14", 3.14, false},
		{"1,000", 1000, false},
		{"1,234.56", 1234.56, false},
		{"20%", 0.20, true},
		{"50%", 0.50, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			toks, err := Lex(tt.input)
			if err != nil {
				t.Fatalf("Lex(%q) error: %v", tt.input, err)
			}
			if len(toks) < 1 {
				t.Fatalf("Lex(%q) returned no tokens", tt.input)
			}
			if toks[0].Kind != tokNumber {
				t.Errorf("Lex(%q) first token kind = %v, want tokNumber", tt.input, toks[0].Kind)
			}
			if toks[0].Num != tt.expectedNum {
				t.Errorf("Lex(%q) Num = %v, want %v", tt.input, toks[0].Num, tt.expectedNum)
			}
			if toks[0].Pct != tt.expectedPct {
				t.Errorf("Lex(%q) Pct = %v, want %v", tt.input, toks[0].Pct, tt.expectedPct)
			}
		})
	}
}

func TestLexCurrency(t *testing.T) {
	tests := []struct {
		input       string
		expectedNum float64
	}{
		{"$100", 100},
		{"$99.99", 99.99},
		{"$1,000", 1000},
		{"$1,234.56", 1234.56},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			toks, err := Lex(tt.input)
			if err != nil {
				t.Fatalf("Lex(%q) error: %v", tt.input, err)
			}
			if len(toks) < 1 {
				t.Fatalf("Lex(%q) returned no tokens", tt.input)
			}
			if toks[0].Kind != tokNumber {
				t.Errorf("Lex(%q) first token kind = %v, want tokNumber", tt.input, toks[0].Kind)
			}
			if toks[0].Num != tt.expectedNum {
				t.Errorf("Lex(%q) Num = %v, want %v", tt.input, toks[0].Num, tt.expectedNum)
			}
		})
	}
}

func TestLexOperators(t *testing.T) {
	tests := []struct {
		input        string
		expectedKind TokenKind
	}{
		{"+", tokPlus},
		{"-", tokMinus},
		{"*", tokMul},
		{"/", tokDiv},
		{"^", tokPow},
		{"(", tokLParen},
		{")", tokRParen},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			toks, err := Lex(tt.input)
			if err != nil {
				t.Fatalf("Lex(%q) error: %v", tt.input, err)
			}
			if len(toks) < 1 {
				t.Fatalf("Lex(%q) returned no tokens", tt.input)
			}
			if toks[0].Kind != tt.expectedKind {
				t.Errorf("Lex(%q) first token kind = %v, want %v", tt.input, toks[0].Kind, tt.expectedKind)
			}
		})
	}
}

func TestLexReferences(t *testing.T) {
	tests := []struct {
		input       string
		expectedRef int
	}{
		{"\\1", 1},
		{"\\2", 2},
		{"\\10", 10},
		{"\\99", 99},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			toks, err := Lex(tt.input)
			if err != nil {
				t.Fatalf("Lex(%q) error: %v", tt.input, err)
			}
			if len(toks) < 1 {
				t.Fatalf("Lex(%q) returned no tokens", tt.input)
			}
			if toks[0].Kind != tokRef {
				t.Errorf("Lex(%q) first token kind = %v, want tokRef", tt.input, toks[0].Kind)
			}
			if toks[0].Ref != tt.expectedRef {
				t.Errorf("Lex(%q) Ref = %v, want %v", tt.input, toks[0].Ref, tt.expectedRef)
			}
		})
	}
}

func TestLexIdentifiers(t *testing.T) {
	tests := []struct {
		input        string
		expectedText string
	}{
		{"sin", "sin"},
		{"cos", "cos"},
		{"tan", "tan"},
		{"sqrt", "sqrt"},
		{"max", "max"},
		{"min", "min"},
		{"abs", "abs"},
		{"log", "log"},
		{"ln", "ln"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			toks, err := Lex(tt.input)
			if err != nil {
				t.Fatalf("Lex(%q) error: %v", tt.input, err)
			}
			if len(toks) < 1 {
				t.Fatalf("Lex(%q) returned no tokens", tt.input)
			}
			if toks[0].Kind != tokIdent {
				t.Errorf("Lex(%q) first token kind = %v, want tokIdent", tt.input, toks[0].Kind)
			}
			if toks[0].Text != tt.expectedText {
				t.Errorf("Lex(%q) Text = %q, want %q", tt.input, toks[0].Text, tt.expectedText)
			}
		})
	}
}

func TestLexComplexExpressions(t *testing.T) {
	tests := []struct {
		input         string
		expectedCount int // number of tokens including EOF
	}{
		{"2 + 3", 4},           // NUM + NUM EOF
		{"$100 - 20%", 4},      // NUM - NUM EOF
		{"sin(45)", 5},         // IDENT ( NUM ) EOF
		{"\\1 * 2", 4},         // REF * NUM EOF
		{"(1 + 2) * 3", 8},     // ( NUM + NUM ) * NUM EOF
		{"$7.99 * 4 * \\1", 6}, // NUM * NUM * REF EOF
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			toks, err := Lex(tt.input)
			if err != nil {
				t.Fatalf("Lex(%q) error: %v", tt.input, err)
			}
			if len(toks) != tt.expectedCount {
				t.Errorf("Lex(%q) token count = %d, want %d", tt.input, len(toks), tt.expectedCount)
				for i, tok := range toks {
					t.Logf("  token %d: Kind=%v Text=%q", i, tok.Kind, tok.Text)
				}
			}
		})
	}
}

func TestLexErrors(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"\\"},   // backslash without number
		{"$"},    // dollar without number
		{"$abc"}, // dollar with non-number
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := Lex(tt.input)
			if err == nil {
				t.Errorf("Lex(%q) expected error, got nil", tt.input)
			}
		})
	}
}

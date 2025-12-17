package eval

import (
	"math"
	"testing"
)

func TestEvalExprBasicArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"2 + 3", 5},
		{"10 - 4", 6},
		{"3 * 4", 12},
		{"20 / 5", 4},
		{"2 ^ 3", 8},
		{"2 + 3 * 4", 14},   // precedence: 2 + (3*4)
		{"(2 + 3) * 4", 20}, // parentheses
		{"10 - 2 - 3", 5},   // left associative
		{"2 ^ 3 ^ 2", 512},  // right associative: 2^(3^2) = 2^9
		{"-5", -5},          // unary minus
		{"+5", 5},           // unary plus
		{"--5", 5},          // double negative
		{"10 / 2 / 2", 2.5}, // left associative division
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := EvalExpr(tt.input, nil)
			if err != nil {
				t.Fatalf("EvalExpr(%q) error: %v", tt.input, err)
			}
			if math.Abs(result-tt.expected) > 0.0001 {
				t.Errorf("EvalExpr(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEvalExprCurrency(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"$100", 100},
		{"$99.99", 99.99},
		{"$1,000", 1000},
		{"$100 + $50", 150},
		{"$100 * 2", 200},
		{"$1,234.56 + $765.44", 2000},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := EvalExpr(tt.input, nil)
			if err != nil {
				t.Fatalf("EvalExpr(%q) error: %v", tt.input, err)
			}
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("EvalExpr(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEvalExprPercentage(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"100 + 20%", 120}, // 100 * 1.20
		{"100 - 20%", 80},  // 100 * 0.80
		{"$100 - 10%", 90}, // currency with percent
		{"200 + 50%", 300}, // 200 * 1.50
		{"50%", 0.50},      // standalone percent
		{"100 * 20%", 20},  // multiplication (not percent-of-left)
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := EvalExpr(tt.input, nil)
			if err != nil {
				t.Fatalf("EvalExpr(%q) error: %v", tt.input, err)
			}
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("EvalExpr(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEvalExprReferences(t *testing.T) {
	// Mock resolver that returns line values
	values := map[int]float64{
		1: 100,
		2: 50,
		3: 25,
	}
	resolver := func(n int) (float64, error) {
		if v, ok := values[n]; ok {
			return v, nil
		}
		return 0, nil
	}

	tests := []struct {
		input    string
		expected float64
	}{
		{"\\1", 100},
		{"\\2", 50},
		{"\\1 + \\2", 150},
		{"\\1 * 2", 200},
		{"\\1 + \\2 + \\3", 175},
		{"2 * \\1", 200},
		{"$7.99 * 4 * \\1", 3196}, // This was failing - let's verify
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := EvalExpr(tt.input, resolver)
			if err != nil {
				t.Fatalf("EvalExpr(%q) error: %v", tt.input, err)
			}
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("EvalExpr(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEvalExprMultiplicationWithX(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"2 x 3", 6},
		{"$7.99 x 4", 31.96},
		{"10 x 10", 100},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := EvalExpr(tt.input, nil)
			if err != nil {
				t.Fatalf("EvalExpr(%q) error: %v", tt.input, err)
			}
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("EvalExpr(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEvalExprErrors(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"2 +"},        // incomplete expression
		{"* 3"},        // missing left operand
		{"sin"},        // function without parens
		{"unknown(5)"}, // unknown function
		{"(2 + 3"},     // unclosed paren
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := EvalExpr(tt.input, nil)
			if err == nil {
				t.Errorf("EvalExpr(%q) expected error, got nil", tt.input)
			}
		})
	}
}

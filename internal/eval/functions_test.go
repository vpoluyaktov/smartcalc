package eval

import (
	"math"
	"testing"
)

func TestCallFn(t *testing.T) {
	tests := []struct {
		name     string
		fn       string
		arg      float64
		expected float64
		wantErr  bool
	}{
		// Trig functions use radians
		{"sin 0", "sin", 0, 0, false},
		{"sin pi/2", "sin", math.Pi / 2, 1, false},
		{"cos 0", "cos", 0, 1, false},
		{"cos pi/2", "cos", math.Pi / 2, 0, false},
		{"tan 0", "tan", 0, 0, false},
		{"tan pi/4", "tan", math.Pi / 4, 1, false},
		{"asin 0", "asin", 0, 0, false},
		{"acos 1", "acos", 1, 0, false},
		{"atan 0", "atan", 0, 0, false},
		{"sqrt 4", "sqrt", 4, 2, false},
		{"sqrt 9", "sqrt", 9, 3, false},
		{"sqrt 2", "sqrt", 2, math.Sqrt(2), false},
		{"abs positive", "abs", 5, 5, false},
		{"abs negative", "abs", -5, 5, false},
		{"abs zero", "abs", 0, 0, false},
		{"log 10", "log", 10, 1, false},
		{"log 100", "log", 100, 2, false},
		{"ln e", "ln", math.E, 1, false},
		{"unknown function", "unknown", 5, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := callFn(tt.fn, tt.arg)
			if tt.wantErr {
				if err == nil {
					t.Errorf("callFn(%q, %v) expected error, got nil", tt.fn, tt.arg)
				}
				return
			}
			if err != nil {
				t.Fatalf("callFn(%q, %v) error: %v", tt.fn, tt.arg, err)
			}
			if math.Abs(result-tt.expected) > 0.0001 {
				t.Errorf("callFn(%q, %v) = %v, want %v", tt.fn, tt.arg, result, tt.expected)
			}
		})
	}
}

func TestEvalExprFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"sin(0)", 0},
		{"cos(0)", 1},
		{"sqrt(16)", 4},
		{"abs(-10)", 10},
		{"log(100)", 2},
		{"ln(2.718281828)", 1}, // ln(e) ≈ 1
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

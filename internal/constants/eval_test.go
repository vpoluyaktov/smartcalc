package constants

import (
	"strings"
	"testing"
)

func TestMathematicalConstants(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"pi", "3.14159"},
		{"e", "2.71828"},
		{"phi", "1.618"},
		{"golden ratio", "1.618"},
		{"sqrt2", "1.414"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalConstants(tt.expr)
			if err != nil {
				t.Errorf("EvalConstants(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("EvalConstants(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

func TestPhysicalConstants(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"speed of light", "2.99792458e+08"},
		{"c", "m/s"},
		{"gravity", "9.80665"},
		{"g", "m/s"},
		{"avogadro", "6.022"},
		{"planck", "6.626"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalConstants(tt.expr)
			if err != nil {
				t.Errorf("EvalConstants(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("EvalConstants(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

func TestValueOfPattern(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"value of pi", "3.14159"},
		{"value of speed of light", "2.99792458e+08"},
		{"value of gravity", "9.80665"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalConstants(tt.expr)
			if err != nil {
				t.Errorf("EvalConstants(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("EvalConstants(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

func TestIsConstantExpression(t *testing.T) {
	tests := []struct {
		expr     string
		expected bool
	}{
		{"pi", true},
		{"speed of light", true},
		{"value of gravity", true},
		{"golden ratio", true},
		{"100 + 50", false},
		{"5 miles in km", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := IsConstantExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("IsConstantExpression(%q) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}

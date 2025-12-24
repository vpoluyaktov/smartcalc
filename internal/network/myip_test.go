package network

import (
	"testing"
)

func TestIsMyIPExpression(t *testing.T) {
	tests := []struct {
		expr     string
		expected bool
	}{
		{"what is my ip", true},
		{"What is my IP", true},
		{"what's my ip", true},
		{"whats my ip", true},
		{"my ip", true},
		{"my ip address", true},
		{"show my ip", true},
		{"get my ip", true},

		// Invalid expressions
		{"what is my name", false},
		{"geoip 8.8.8.8", false},
		{"ip lookup 1.1.1.1", false},
		{"hello world", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := IsMyIPExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("IsMyIPExpression(%q) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}

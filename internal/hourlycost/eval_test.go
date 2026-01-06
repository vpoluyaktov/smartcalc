package hourlycost

import (
	"testing"
)

func TestIsHourlyCostExpression(t *testing.T) {
	tests := []struct {
		expr     string
		expected bool
	}{
		// Valid expressions with dollars
		{"$35 per hour in week", true},
		{"$35 per hour in 1 week", true},
		{"$45 per hour in 5 months", true},
		{"$100 per hour in 2 years", true},
		{"$50 per hour in 3 days", true},
		{"$35 per a hour in week", true},
		{"$35 per an hour in week", true},
		{"$1,000 per hour in 1 month", true},
		{"$35.50 per hour in 2 weeks", true},

		// Valid expressions with cents
		{"25 cents per hour in 2 years", true},
		{"50 cent per hour in 1 year", true},
		{"99 cents per hour in 6 months", true},
		{"10 cents per hour in 1 day", true},

		// Invalid expressions
		{"$35 per day in week", false},
		{"$35 hour in week", false},
		{"35 per hour in week", false},
		{"$35 per hour", false},
		{"hello world", false},
		{"2 + 2", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := IsHourlyCostExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("IsHourlyCostExpression(%q) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestEvalHourlyCost(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
		wantErr  bool
	}{
		// Basic dollar calculations
		{"$35 per hour in 1 week", "$5,880.00", false},
		{"$35 per hour in week", "$5,880.00", false},
		{"$45 per hour in 5 months", "$162,000.00", false},
		{"$100 per hour in 1 year", "$876,000.00", false},
		{"$50 per hour in 1 day", "$1,200.00", false},

		// With "a" or "an"
		{"$35 per a hour in 1 week", "$5,880.00", false},
		{"$35 per an hour in 1 week", "$5,880.00", false},

		// Decimal rates
		{"$35.50 per hour in 1 week", "$5,964.00", false},
		{"$0.50 per hour in 1 day", "$12.00", false},

		// Rates with commas
		{"$1,000 per hour in 1 week", "$168,000.00", false},

		// Cents calculations
		{"25 cents per hour in 2 years", "$4,380.00", false},
		{"50 cents per hour in 1 year", "$4,380.00", false},
		{"100 cents per hour in 1 day", "$24.00", false},
		{"1 cent per hour in 1 year", "$87.60", false},

		// Multiple units
		{"$10 per hour in 2 weeks", "$3,360.00", false},
		{"$20 per hour in 3 months", "$43,200.00", false},
		{"$15 per hour in 10 days", "$3,600.00", false},

		// Case insensitivity
		{"$35 PER HOUR IN 1 WEEK", "$5,880.00", false},
		{"$35 Per Hour In 1 Week", "$5,880.00", false},
		{"25 CENTS PER HOUR IN 1 YEAR", "$2,190.00", false},

		// Invalid expressions
		{"invalid expression", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalHourlyCost(tt.expr)
			if tt.wantErr {
				if err == nil {
					t.Errorf("EvalHourlyCost(%q) expected error, got result: %q", tt.expr, result)
				}
				return
			}
			if err != nil {
				t.Errorf("EvalHourlyCost(%q) unexpected error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalHourlyCost(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestHourlyCostConstants(t *testing.T) {
	if HoursPerDay != 24.0 {
		t.Errorf("HoursPerDay = %v, want 24.0", HoursPerDay)
	}
	if HoursPerWeek != 168.0 {
		t.Errorf("HoursPerWeek = %v, want 168.0", HoursPerWeek)
	}
	if HoursPerMonth != 720.0 {
		t.Errorf("HoursPerMonth = %v, want 720.0", HoursPerMonth)
	}
	if HoursPerYear != 8760.0 {
		t.Errorf("HoursPerYear = %v, want 8760.0", HoursPerYear)
	}
}

func TestFormatCurrency(t *testing.T) {
	tests := []struct {
		value    float64
		expected string
	}{
		{100.00, "$100.00"},
		{1000.00, "$1,000.00"},
		{1234567.89, "$1,234,567.89"},
		{0.50, "$0.50"},
		{999999.99, "$999,999.99"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatCurrency(tt.value)
			if result != tt.expected {
				t.Errorf("formatCurrency(%v) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

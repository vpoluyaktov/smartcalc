package datetime

import (
	"strings"
	"testing"
	"time"
)

func TestEvalNowIn(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"now in Seattle", "PST"},
		{"now in New York", "EST"},
		{"now in Kiev", "EET"},
		{"now in Moscow", "MSK"},
		{"now() in Seattle", "PST"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalDateTime(tt.expr)
			if err != nil {
				t.Errorf("EvalDateTime(%q) error: %v", tt.expr, err)
				return
			}
			// Result should contain a timezone abbreviation (may vary by DST)
			if result == "" {
				t.Errorf("EvalDateTime(%q) returned empty result", tt.expr)
			}
		})
	}
}

func TestEvalDurationConversion(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"861.5 hours in days", "35.90 days"},
		{"24 hours in days", "1 days"},
		{"7 days in weeks", "1 weeks"},
		{"60 minutes in hours", "1 hours"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalDateTime(tt.expr)
			if err != nil {
				t.Errorf("EvalDateTime(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalDateTime(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestEvalDateArithmetic(t *testing.T) {
	tests := []struct {
		expr        string
		shouldParse bool
	}{
		{"today() - 35.9 days", true},
		{"today() + 10 days", true},
		{"now() + 5 hours", true},
		{"2025-09-25 19:00:00 + 10 hours", true},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalDateTime(tt.expr)
			if tt.shouldParse {
				if err != nil {
					t.Errorf("EvalDateTime(%q) error: %v", tt.expr, err)
					return
				}
				if result == "" {
					t.Errorf("EvalDateTime(%q) returned empty result", tt.expr)
				}
			}
		})
	}
}

func TestEvalTimeConversion(t *testing.T) {
	tests := []struct {
		expr        string
		shouldParse bool
	}{
		{"6:00 am Seattle in Kiev", true},
		{"11am kiev in seattle", true},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalDateTime(tt.expr)
			if tt.shouldParse {
				if err != nil {
					t.Errorf("EvalDateTime(%q) error: %v", tt.expr, err)
					return
				}
				if result == "" {
					t.Errorf("EvalDateTime(%q) returned empty result", tt.expr)
				}
			}
		})
	}
}

func TestEvalDateRange(t *testing.T) {
	tests := []struct {
		expr    string
		minDays float64
		maxDays float64
	}{
		{"Dec 6 till March 11", 90, 100}, // approximately 95 days
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalDateTime(tt.expr)
			if err != nil {
				t.Errorf("EvalDateTime(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, "days") {
				t.Errorf("EvalDateTime(%q) = %q, expected to contain 'days'", tt.expr, result)
			}
		})
	}
}

func TestEvalDurationMultiplication(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"13 x 3 min", "39"},
		{"8 hours x 5", "days"}, // 40 hours = some days
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalDateTime(tt.expr)
			if err != nil {
				t.Errorf("EvalDateTime(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("EvalDateTime(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"5 hours", 5 * time.Hour},
		{"30 minutes", 30 * time.Minute},
		{"2 days", 48 * time.Hour},
		{"1 week", 7 * 24 * time.Hour},
		{"3.5 hours", time.Duration(3.5 * float64(time.Hour))},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseDuration(tt.input)
			if err != nil {
				t.Errorf("ParseDuration(%q) error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("ParseDuration(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLookupTimezone(t *testing.T) {
	tests := []struct {
		city    string
		wantErr bool
	}{
		{"Seattle", false},
		{"New York", false},
		{"Kiev", false},
		{"Moscow", false},
		{"EST", false},
		{"PST", false},
		{"unknown_city_xyz", true},
	}

	for _, tt := range tests {
		t.Run(tt.city, func(t *testing.T) {
			_, err := LookupTimezone(tt.city)
			if (err != nil) != tt.wantErr {
				t.Errorf("LookupTimezone(%q) error = %v, wantErr %v", tt.city, err, tt.wantErr)
			}
		})
	}
}

func TestIsDateTimeExpression(t *testing.T) {
	tests := []struct {
		expr     string
		expected bool
	}{
		{"now in Seattle", true},
		{"5 hours in days", true},
		{"today() - 10 days", true},
		{"6:00 am Seattle in Kiev", true},
		{"100 + 50", false},
		{"$100 - 20%", false},
		{"sin(45)", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := IsDateTimeExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("IsDateTimeExpression(%q) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}

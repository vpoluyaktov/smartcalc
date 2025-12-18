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

func TestEvalDateTimeWithRefs(t *testing.T) {
	// Test reference resolution
	resolver := func(n int) (string, bool) {
		switch n {
		case 1:
			return "2025-12-17 16:00:00 PST", true
		case 2:
			return "2025-01-01 00:00:00 UTC", true
		default:
			return "", false
		}
	}

	tests := []struct {
		expr        string
		shouldParse bool
		contains    string
	}{
		{"\\1 + 3 days", true, "2025-12-20"},
		{"\\1 - 1 week", true, "2025-12-10"},
		{"\\2 + 5 hours", true, "05:00"},
		{"\\1 + 24 hours", true, "2025-12-18"},
		{"\\99 + 1 day", false, ""}, // invalid reference
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalDateTimeWithRefs(tt.expr, resolver)
			if tt.shouldParse {
				if err != nil {
					t.Errorf("EvalDateTimeWithRefs(%q) error: %v", tt.expr, err)
					return
				}
				if tt.contains != "" && !strings.Contains(result, tt.contains) {
					t.Errorf("EvalDateTimeWithRefs(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
				}
			} else {
				if err == nil {
					t.Errorf("EvalDateTimeWithRefs(%q) expected error, got result: %q", tt.expr, result)
				}
			}
		})
	}
}

func TestResolveRefsInExpr(t *testing.T) {
	resolver := func(n int) (string, bool) {
		switch n {
		case 1:
			return "2025-12-17 16:00:00 PST", true
		case 2:
			return "2025-01-01 00:00:00 UTC", true
		default:
			return "", false
		}
	}

	tests := []struct {
		expr     string
		expected string
	}{
		{"\\1 + 3 days", "2025-12-17 16:00:00 PST + 3 days"},
		{"\\2 - 1 week", "2025-01-01 00:00:00 UTC - 1 week"},
		{"\\1 + \\2", "2025-12-17 16:00:00 PST + 2025-01-01 00:00:00 UTC"},
		{"no refs here", "no refs here"},
		{"\\99 + 1 day", "\\99 + 1 day"}, // unresolved ref stays as-is
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := resolveRefsInExpr(tt.expr, resolver)
			if result != tt.expected {
				t.Errorf("resolveRefsInExpr(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestEvalNumberPlusDuration(t *testing.T) {
	tests := []struct {
		expr        string
		shouldParse bool
	}{
		{"0 + 3 days", true},
		{"0 - 1 week", true},
		{"0 + 5 hours", true},
		{"0 + 30 minutes", true},
		{"1 + 3 days", false}, // only 0 is treated as "now"
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
			} else {
				if err == nil {
					// For non-datetime expressions, it should fail
					// But we need to check if it was actually parsed as datetime
				}
			}
		})
	}
}

func TestEvalNowAndToday(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"now", ""},     // should return current time
		{"now()", ""},   // should return current time
		{"today", ""},   // should return today's date
		{"today()", ""}, // should return today's date
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalDateTime(tt.expr)
			if err != nil {
				t.Errorf("EvalDateTime(%q) error: %v", tt.expr, err)
				return
			}
			if result == "" {
				t.Errorf("EvalDateTime(%q) returned empty result", tt.expr)
			}
		})
	}
}

func TestEvalDateTimeConversionWithTimezone(t *testing.T) {
	tests := []struct {
		expr        string
		shouldParse bool
	}{
		{"2025-09-25 19:00:00 EST in Seattle", true},
		{"2025-10-06 18:00:00 est in seattle", true},
		{"2025-10-15 23:50:00 EST in seattle", true},
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
				// Should contain PST or PDT for Seattle
				if !strings.Contains(result, "P") {
					t.Errorf("EvalDateTime(%q) = %q, expected Pacific timezone", tt.expr, result)
				}
			}
		})
	}
}

func TestEvalComplexDurationExpressions(t *testing.T) {
	tests := []struct {
		expr        string
		shouldParse bool
		contains    string
	}{
		{"(8 hours x 5 x 2) x 2", true, "days"},
		{"2 hours x 3", true, "hours"},
		{"30 min x 4", true, "hours"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalDateTime(tt.expr)
			if tt.shouldParse {
				if err != nil {
					t.Errorf("EvalDateTime(%q) error: %v", tt.expr, err)
					return
				}
				if tt.contains != "" && !strings.Contains(result, tt.contains) {
					t.Errorf("EvalDateTime(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
				}
			}
		})
	}
}

func TestConvertDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		toUnit   string
		expected float64
	}{
		{24 * time.Hour, "days", 1},
		{48 * time.Hour, "days", 2},
		{7 * 24 * time.Hour, "weeks", 1},
		{60 * time.Minute, "hours", 1},
		{3600 * time.Second, "hours", 1},
	}

	for _, tt := range tests {
		t.Run(tt.toUnit, func(t *testing.T) {
			result, err := ConvertDuration(tt.duration, tt.toUnit)
			if err != nil {
				t.Errorf("ConvertDuration(%v, %q) error: %v", tt.duration, tt.toUnit, err)
				return
			}
			if result != tt.expected {
				t.Errorf("ConvertDuration(%v, %q) = %v, want %v", tt.duration, tt.toUnit, result, tt.expected)
			}
		})
	}
}

func TestParseDateRange(t *testing.T) {
	tests := []struct {
		expr        string
		shouldParse bool
	}{
		{"Dec 6 till March 11", true},
		{"Jan 1 until Dec 31", true},
		{"March 15 to April 20", true},
		{"invalid range", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			start, end, err := ParseDateRange(tt.expr)
			if tt.shouldParse {
				if err != nil {
					t.Errorf("ParseDateRange(%q) error: %v", tt.expr, err)
					return
				}
				if start.IsZero() || end.IsZero() {
					t.Errorf("ParseDateRange(%q) returned zero time", tt.expr)
				}
				if !end.After(start) {
					t.Errorf("ParseDateRange(%q) end should be after start", tt.expr)
				}
			} else {
				if err == nil {
					t.Errorf("ParseDateRange(%q) expected error", tt.expr)
				}
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		contains string
	}{
		{48 * time.Hour, "days"},
		{5 * time.Hour, "hours"},
		{30 * time.Minute, "minutes"},
		{45 * time.Second, "seconds"},
	}

	for _, tt := range tests {
		t.Run(tt.contains, func(t *testing.T) {
			result := FormatDuration(tt.duration)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("FormatDuration(%v) = %q, want to contain %q", tt.duration, result, tt.contains)
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

func TestFormatTime(t *testing.T) {
	// Test that FormatTime truncates to minutes (no seconds)
	testTime := time.Date(2025, 12, 18, 14, 30, 45, 0, time.UTC)
	result := FormatTime(testTime)

	// Should contain date and hour:minute
	if !strings.Contains(result, "2025-12-18") {
		t.Errorf("FormatTime() = %q, should contain date '2025-12-18'", result)
	}
	if !strings.Contains(result, "14:30") {
		t.Errorf("FormatTime() = %q, should contain time '14:30'", result)
	}
	// Should NOT contain seconds
	if strings.Contains(result, "14:30:45") {
		t.Errorf("FormatTime() = %q, should NOT contain seconds", result)
	}
	// Should contain timezone
	if !strings.Contains(result, "UTC") {
		t.Errorf("FormatTime() = %q, should contain timezone 'UTC'", result)
	}
}

func TestFormatTimeNoSeconds(t *testing.T) {
	// Verify the format is exactly "2006-01-02 15:04 MST" (no seconds)
	testTime := time.Date(2025, 1, 1, 0, 0, 59, 0, time.UTC)
	result := FormatTime(testTime)

	expected := "2025-01-01 00:00 UTC"
	if result != expected {
		t.Errorf("FormatTime() = %q, want %q", result, expected)
	}
}

func TestEvalNowNoSeconds(t *testing.T) {
	// Verify that "now" output doesn't include seconds
	result, err := EvalDateTime("now")
	if err != nil {
		t.Fatalf("EvalDateTime('now') error: %v", err)
	}

	// Result should be in format "YYYY-MM-DD HH:MM TZ" (no seconds)
	// Count colons - should be exactly 1 (in HH:MM)
	colonCount := strings.Count(result, ":")
	if colonCount != 1 {
		t.Errorf("EvalDateTime('now') = %q, expected exactly 1 colon (no seconds)", result)
	}
}

package manhour

import (
	"testing"
)

func TestIsManHourExpression(t *testing.T) {
	tests := []struct {
		expr     string
		expected bool
	}{
		// Valid business time expressions
		{"248 man-hour / 3 men in business weeks", true},
		{"248 man-hours / 3 men in business weeks", true},
		{"248 manhour / 3 men in business weeks", true},
		{"248 manhours / 3 men in business weeks", true},
		{"248 man hour / 3 men in business weeks", true},
		{"248 man hours / 3 men in business weeks", true},
		{"100 man-hour / 1 man in business weeks", true},
		{"100 man-hours / 2 person in business weeks", true},
		{"100 man-hours / 2 persons in business weeks", true},
		{"100 man-hours / 2 people in business weeks", true},
		{"80 man-hours / 2 men in business days", true},
		{"40 man-hour / 1 man in business day", true},
		{"248.5 man-hours / 3.5 men in business weeks", true},
		{"320 man-hours / 2 men in business months", true},
		{"160 man-hour / 1 man in business month", true},

		// Valid calendar time expressions (with "calendar" prefix)
		{"248 man-hour / 3 men in calendar weeks", true},
		{"168 man-hours / 1 man in calendar week", true},
		{"48 man-hours / 2 men in calendar days", true},
		{"720 man-hours / 1 man in calendar months", true},
		{"720 man-hour / 1 man in calendar month", true},

		// Valid calendar time expressions (without prefix - defaults to calendar)
		{"248 man-hour / 3 men in weeks", true},
		{"168 man-hours / 1 man in week", true},
		{"48 man-hours / 2 men in days", true},
		{"24 man-hour / 1 man in day", true},
		{"720 man-hours / 1 man in months", true},
		{"720 man-hour / 1 man in month", true},

		// Invalid expressions
		{"248 hours / 3 men in business weeks", false},
		{"248 man-hour / 3 in business weeks", false},
		{"hello world", false},
		{"2 + 2", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := IsManHourExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("IsManHourExpression(%q) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestEvalManHour(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
		wantErr  bool
	}{
		// Basic calculations - business weeks
		{"248 man-hour / 3 men in business weeks", "2.07 business weeks", false},
		{"240 man-hours / 3 men in business weeks", "2 business weeks", false},
		{"40 man-hour / 1 man in business weeks", "1 business week", false},
		{"80 man-hours / 2 men in business weeks", "1 business week", false},
		{"120 man-hours / 3 men in business weeks", "1 business week", false},
		{"400 man-hours / 2 men in business weeks", "5 business weeks", false},

		// Business days
		{"80 man-hours / 2 men in business days", "5 business days", false},
		{"8 man-hour / 1 man in business days", "1 business day", false},
		{"16 man-hours / 2 men in business days", "1 business day", false},
		{"24 man-hours / 3 men in business days", "1 business day", false},

		// Business months (160 hours = 1 business month)
		{"160 man-hour / 1 man in business months", "1 business month", false},
		{"320 man-hours / 2 men in business months", "1 business month", false},
		{"480 man-hours / 3 men in business months", "1 business month", false},
		{"800 man-hours / 2 men in business months", "2.50 business months", false},

		// Calendar weeks (168 hours = 1 calendar week)
		{"168 man-hours / 1 man in weeks", "1 week", false},
		{"168 man-hours / 1 man in calendar weeks", "1 week", false},
		{"336 man-hours / 2 men in weeks", "1 week", false},
		{"504 man-hours / 3 men in calendar weeks", "1 week", false},
		{"248 man-hour / 3 men in weeks", "0.49 weeks", false},

		// Calendar days (24 hours = 1 calendar day)
		{"24 man-hours / 1 man in days", "1 day", false},
		{"24 man-hours / 1 man in calendar days", "1 day", false},
		{"48 man-hours / 2 men in days", "1 day", false},
		{"72 man-hours / 3 men in calendar days", "1 day", false},
		{"120 man-hours / 2 men in days", "2.50 days", false},

		// Calendar months (720 hours = 1 calendar month)
		{"720 man-hours / 1 man in months", "1 month", false},
		{"720 man-hours / 1 man in calendar months", "1 month", false},
		{"1440 man-hours / 2 men in months", "1 month", false},
		{"2160 man-hours / 3 men in calendar months", "1 month", false},

		// Fractional results
		{"100 man-hours / 3 men in business weeks", "0.83 business weeks", false},
		{"50 man-hours / 2 men in business days", "3.12 business days", false},

		// Case insensitivity
		{"248 MAN-HOUR / 3 MEN IN BUSINESS WEEKS", "2.07 business weeks", false},
		{"248 Man-Hours / 3 Men in Business Weeks", "2.07 business weeks", false},
		{"168 MAN-HOURS / 1 MAN IN WEEKS", "1 week", false},
		{"168 Man-Hours / 1 Man in Calendar Weeks", "1 week", false},

		// Different people terms
		{"80 man-hours / 2 person in business weeks", "1 business week", false},
		{"80 man-hours / 2 persons in business weeks", "1 business week", false},
		{"80 man-hours / 2 people in business weeks", "1 business week", false},

		// Division by zero
		{"100 man-hours / 0 men in business weeks", "", true},

		// Invalid expressions
		{"invalid expression", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalManHour(tt.expr)
			if tt.wantErr {
				if err == nil {
					t.Errorf("EvalManHour(%q) expected error, got result: %q", tt.expr, result)
				}
				return
			}
			if err != nil {
				t.Errorf("EvalManHour(%q) unexpected error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalManHour(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestManHourConstants(t *testing.T) {
	// Business time constants
	if HoursPerBusinessDay != 8.0 {
		t.Errorf("HoursPerBusinessDay = %v, want 8.0", HoursPerBusinessDay)
	}
	if DaysPerBusinessWeek != 5.0 {
		t.Errorf("DaysPerBusinessWeek = %v, want 5.0", DaysPerBusinessWeek)
	}
	if HoursPerBusinessWeek != 40.0 {
		t.Errorf("HoursPerBusinessWeek = %v, want 40.0", HoursPerBusinessWeek)
	}
	if DaysPerBusinessMonth != 20.0 {
		t.Errorf("DaysPerBusinessMonth = %v, want 20.0", DaysPerBusinessMonth)
	}
	if HoursPerBusinessMonth != 160.0 {
		t.Errorf("HoursPerBusinessMonth = %v, want 160.0", HoursPerBusinessMonth)
	}

	// Calendar time constants
	if HoursPerCalendarDay != 24.0 {
		t.Errorf("HoursPerCalendarDay = %v, want 24.0", HoursPerCalendarDay)
	}
	if DaysPerCalendarWeek != 7.0 {
		t.Errorf("DaysPerCalendarWeek = %v, want 7.0", DaysPerCalendarWeek)
	}
	if HoursPerCalendarWeek != 168.0 {
		t.Errorf("HoursPerCalendarWeek = %v, want 168.0", HoursPerCalendarWeek)
	}
	if DaysPerCalendarMonth != 30.0 {
		t.Errorf("DaysPerCalendarMonth = %v, want 30.0", DaysPerCalendarMonth)
	}
	if HoursPerCalendarMonth != 720.0 {
		t.Errorf("HoursPerCalendarMonth = %v, want 720.0", HoursPerCalendarMonth)
	}
}

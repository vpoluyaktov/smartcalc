package stats

import (
	"testing"
)

func TestAverage(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"avg(10, 20, 30, 40)", "25"},
		{"average(1, 2, 3, 4, 5)", "3"},
		{"mean(100, 200)", "150"},
		{"avg(1.5, 2.5, 3.5)", "2.5"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalStats(tt.expr)
			if err != nil {
				t.Errorf("EvalStats(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalStats(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestMedian(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"median(1, 2, 3, 4, 5)", "3"},
		{"median(1, 2, 3, 4)", "2.5"},
		{"median(1, 2, 3, 4, 100)", "3"},
		{"median(10, 20, 30)", "20"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalStats(tt.expr)
			if err != nil {
				t.Errorf("EvalStats(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalStats(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestSum(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"sum(10, 20, 30)", "60"},
		{"sum(1, 2, 3, 4, 5)", "15"},
		{"sum(100)", "100"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalStats(tt.expr)
			if err != nil {
				t.Errorf("EvalStats(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalStats(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestMinMax(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"min(10, 5, 20, 3)", "3"},
		{"max(10, 5, 20, 3)", "20"},
		{"min(-5, 0, 5)", "-5"},
		{"max(-5, 0, 5)", "5"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalStats(tt.expr)
			if err != nil {
				t.Errorf("EvalStats(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalStats(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestStdDev(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"stddev(2, 4, 4, 4, 5, 5, 7, 9)", "2"},
		{"stdev(10, 10, 10)", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalStats(tt.expr)
			if err != nil {
				t.Errorf("EvalStats(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalStats(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestVariance(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"variance(2, 4, 4, 4, 5, 5, 7, 9)", "4"},
		{"var(10, 10, 10)", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalStats(tt.expr)
			if err != nil {
				t.Errorf("EvalStats(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalStats(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestCount(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"count(1, 2, 3, 4, 5)", "5"},
		{"count(10)", "1"},
		{"count(1, 2, 3)", "3"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalStats(tt.expr)
			if err != nil {
				t.Errorf("EvalStats(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalStats(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestRange(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"range(1, 5, 10, 3)", "9"},
		{"range(100, 200, 150)", "100"},
		{"range(-10, 10)", "20"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalStats(tt.expr)
			if err != nil {
				t.Errorf("EvalStats(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalStats(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestIsStatsExpression(t *testing.T) {
	tests := []struct {
		expr     string
		expected bool
	}{
		{"avg(10, 20, 30)", true},
		{"median(1, 2, 3)", true},
		{"sum(1, 2, 3)", true},
		{"stddev(1, 2, 3)", true},
		{"100 + 50", false},
		{"5 miles in km", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := IsStatsExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("IsStatsExpression(%q) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}

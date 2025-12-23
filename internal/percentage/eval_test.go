package percentage

import (
	"strings"
	"testing"
)

func TestWhatIsPercentOf(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"what is 15% of 200", "30"},
		{"what is 50% of 100", "50"},
		{"20% of 150", "30"},
		{"10% of 1000", "100"},
		{"what is 25 % of 80", "20"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalPercentage(tt.expr)
			if err != nil {
				t.Errorf("EvalPercentage(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalPercentage(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestWhatPercentIs(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"50 is what % of 200", "25.00%"},
		{"25 is what percent of 100", "25.00%"},
		{"75 is what percentage of 300", "25.00%"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalPercentage(tt.expr)
			if err != nil {
				t.Errorf("EvalPercentage(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalPercentage(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestIncreaseByPercent(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"increase 100 by 20%", "120"},
		{"increase 50 by 10%", "55"},
		{"200 increased by 25%", "250"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalPercentage(tt.expr)
			if err != nil {
				t.Errorf("EvalPercentage(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalPercentage(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestDecreaseByPercent(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"decrease 500 by 15%", "425"},
		{"decrease 100 by 20%", "80"},
		{"200 decreased by 50%", "100"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalPercentage(tt.expr)
			if err != nil {
				t.Errorf("EvalPercentage(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalPercentage(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestPercentChange(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"percent change from 50 to 75", "+50.00%"},
		{"percent change from 100 to 50", "-50.00%"},
		{"percentage change from 80 to 100", "+25.00%"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalPercentage(tt.expr)
			if err != nil {
				t.Errorf("EvalPercentage(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalPercentage(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestTipCalculation(t *testing.T) {
	tests := []struct {
		expr     string
		contains []string
	}{
		{"tip 20% on $85.50", []string{"Tip: $17.10", "Total: $102.60"}},
		{"15% tip on 100", []string{"Tip: $15.00", "Total: $115.00"}},
		{"tip 18% on 50", []string{"Tip: $9.00", "Total: $59.00"}},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalPercentage(tt.expr)
			if err != nil {
				t.Errorf("EvalPercentage(%q) error: %v", tt.expr, err)
				return
			}
			for _, c := range tt.contains {
				if !strings.Contains(result, c) {
					t.Errorf("EvalPercentage(%q) = %q, want to contain %q", tt.expr, result, c)
				}
			}
		})
	}
}

func TestSplitBill(t *testing.T) {
	tests := []struct {
		expr     string
		contains []string
	}{
		{"$150 split 4 ways", []string{"Per person: $37.50"}},
		{"$100 split 5 ways with 20% tip", []string{"Total: $120.00", "Per person: $24.00"}},
		{"split $200 4 ways with 15% tip", []string{"Total: $230.00", "Per person: $57.50"}},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalPercentage(tt.expr)
			if err != nil {
				t.Errorf("EvalPercentage(%q) error: %v", tt.expr, err)
				return
			}
			for _, c := range tt.contains {
				if !strings.Contains(result, c) {
					t.Errorf("EvalPercentage(%q) = %q, want to contain %q", tt.expr, result, c)
				}
			}
		})
	}
}

func TestIsPercentageExpression(t *testing.T) {
	tests := []struct {
		expr     string
		expected bool
	}{
		{"what is 15% of 200", true},
		{"50 is what % of 200", true},
		{"increase 100 by 20%", true},
		{"decrease 500 by 15%", true},
		{"tip 20% on $85", true},
		{"100 + 50", false},
		{"5 miles in km", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := IsPercentageExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("IsPercentageExpression(%q) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}

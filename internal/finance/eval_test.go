package finance

import (
	"strings"
	"testing"
)

func TestLoanPayment(t *testing.T) {
	tests := []struct {
		expr     string
		contains []string
	}{
		{"loan $250000 at 6.5% for 30 years", []string{"Monthly: $1580", "Total: $", "Interest: $"}},
		{"loan 100000 at 5% for 15 years", []string{"Monthly: $790", "Total: $"}},
		{"loan $50000 at 4% for 5 years", []string{"Monthly: $920", "Total: $"}},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalFinance(tt.expr)
			if err != nil {
				t.Errorf("EvalFinance(%q) error: %v", tt.expr, err)
				return
			}
			for _, c := range tt.contains {
				if !strings.Contains(result, c) {
					t.Errorf("EvalFinance(%q) = %q, want to contain %q", tt.expr, result, c)
				}
			}
		})
	}
}

func TestCompoundInterest(t *testing.T) {
	tests := []struct {
		expr     string
		contains []string
	}{
		{"$10000 at 5% for 10 years compounded monthly", []string{"Final: $16470", "Interest earned: $6470"}},
		{"compound interest $5000 at 7% for 5 years", []string{"Final: $7012", "Interest earned: $2012"}},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalFinance(tt.expr)
			if err != nil {
				t.Errorf("EvalFinance(%q) error: %v", tt.expr, err)
				return
			}
			for _, c := range tt.contains {
				if !strings.Contains(result, c) {
					t.Errorf("EvalFinance(%q) = %q, want to contain %q", tt.expr, result, c)
				}
			}
		})
	}
}

func TestSimpleInterest(t *testing.T) {
	tests := []struct {
		expr     string
		contains []string
	}{
		{"simple interest $5000 at 3% for 2 years", []string{"Interest: $300", "Total: $5300"}},
		{"simple interest $10000 at 5% for 5 years", []string{"Interest: $2500", "Total: $12500"}},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalFinance(tt.expr)
			if err != nil {
				t.Errorf("EvalFinance(%q) error: %v", tt.expr, err)
				return
			}
			for _, c := range tt.contains {
				if !strings.Contains(result, c) {
					t.Errorf("EvalFinance(%q) = %q, want to contain %q", tt.expr, result, c)
				}
			}
		})
	}
}

func TestMortgagePayment(t *testing.T) {
	tests := []struct {
		expr     string
		contains []string
	}{
		{"mortgage $350000 at 7% for 30 years", []string{"Monthly: $2328", "Total: $", "Interest: $"}},
		{"mortgage $200000 at 4.5% for 15 years", []string{"Monthly: $1529", "Total: $"}},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalFinance(tt.expr)
			if err != nil {
				t.Errorf("EvalFinance(%q) error: %v", tt.expr, err)
				return
			}
			for _, c := range tt.contains {
				if !strings.Contains(result, c) {
					t.Errorf("EvalFinance(%q) = %q, want to contain %q", tt.expr, result, c)
				}
			}
		})
	}
}

func TestInvestmentGrowth(t *testing.T) {
	tests := []struct {
		expr     string
		contains []string
	}{
		{"invest $1000 at 7% for 20 years", []string{"Final: $3869", "Growth: $2869"}},
		{"invest $5000 at 10% for 10 years", []string{"Final: $12968", "Growth: $7968"}},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalFinance(tt.expr)
			if err != nil {
				t.Errorf("EvalFinance(%q) error: %v", tt.expr, err)
				return
			}
			for _, c := range tt.contains {
				if !strings.Contains(result, c) {
					t.Errorf("EvalFinance(%q) = %q, want to contain %q", tt.expr, result, c)
				}
			}
		})
	}
}

func TestIsFinanceExpression(t *testing.T) {
	tests := []struct {
		expr     string
		expected bool
	}{
		{"loan $250000 at 6.5% for 30 years", true},
		{"mortgage $350000 at 7% for 30 years", true},
		{"compound interest $10000 at 5% for 10 years", true},
		{"simple interest $5000 at 3% for 2 years", true},
		{"invest $1000 at 7% for 20 years", true},
		{"100 + 50", false},
		{"5 miles in km", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := IsFinanceExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("IsFinanceExpression(%q) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}

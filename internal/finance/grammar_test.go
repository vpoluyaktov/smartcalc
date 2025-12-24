package finance

import (
	"strings"
	"testing"
)

func TestGrammarParseLoan(t *testing.T) {
	tests := []struct {
		expr      string
		principal float64
		rate      float64
		years     float64
	}{
		{"loan $250000 at 6.5% for 30 years", 250000, 6.5, 30},
		{"loan 100000 at 5% for 15 years", 100000, 5, 15},
		{"loan $50,000 at 4% for 5 years", 50000, 4, 5},
		{"LOAN $100000 AT 7% FOR 10 YEARS", 100000, 7, 10}, // case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			parsed, err := Parse(tt.expr)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tt.expr, err)
			}
			if parsed.Loan == nil {
				t.Fatalf("Parse(%q) did not produce LoanExpr", tt.expr)
			}
			if got := parsed.Loan.Principal.Float64(); got != tt.principal {
				t.Errorf("Principal = %v, want %v", got, tt.principal)
			}
			if got := parsed.Loan.Rate.Float64(); got != tt.rate {
				t.Errorf("Rate = %v, want %v", got, tt.rate)
			}
			if got := parsed.Loan.Term.Years(); got != tt.years {
				t.Errorf("Years = %v, want %v", got, tt.years)
			}
		})
	}
}

func TestGrammarParseMortgage(t *testing.T) {
	tests := []struct {
		expr         string
		principal    float64
		rate         float64
		years        float64
		extraPayment float64
		paySchedule  bool
	}{
		{"mortgage $350000 at 7% for 30 years", 350000, 7, 30, 0, false},
		{"mortgage $200000 at 4.5% for 15 years", 200000, 4.5, 15, 0, false},
		{"mortgage $350000 at 7% for 30 years extra payment $500", 350000, 7, 30, 500, false},
		{"mortgage $200000 at 5% for 30 years extra $200", 200000, 5, 30, 200, false},
		{"mortgage $100000 at 5% for 1 year pay schedule", 100000, 5, 1, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			parsed, err := Parse(tt.expr)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tt.expr, err)
			}
			if parsed.Mortgage == nil {
				t.Fatalf("Parse(%q) did not produce MortgageExpr", tt.expr)
			}
			if got := parsed.Mortgage.Principal.Float64(); got != tt.principal {
				t.Errorf("Principal = %v, want %v", got, tt.principal)
			}
			if got := parsed.Mortgage.Rate.Float64(); got != tt.rate {
				t.Errorf("Rate = %v, want %v", got, tt.rate)
			}
			if got := parsed.Mortgage.Term.Years(); got != tt.years {
				t.Errorf("Years = %v, want %v", got, tt.years)
			}
			if tt.extraPayment > 0 {
				if parsed.Mortgage.ExtraPayment == nil {
					t.Errorf("ExtraPayment is nil, want %v", tt.extraPayment)
				} else if got := parsed.Mortgage.ExtraPayment.Float64(); got != tt.extraPayment {
					t.Errorf("ExtraPayment = %v, want %v", got, tt.extraPayment)
				}
			}
			if got := parsed.Mortgage.PaySchedule; got != tt.paySchedule {
				t.Errorf("PaySchedule = %v, want %v", got, tt.paySchedule)
			}
		})
	}
}

func TestGrammarParseCompoundInterest(t *testing.T) {
	tests := []struct {
		expr      string
		principal float64
		rate      float64
		years     float64
		frequency string
	}{
		{"compound interest $10000 at 5% for 10 years", 10000, 5, 10, ""},
		{"compound interest $5000 at 7% for 5 years", 5000, 7, 5, ""},
		{"compound interest $10000 at 5% for 10 years compounded monthly", 10000, 5, 10, "monthly"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			parsed, err := Parse(tt.expr)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tt.expr, err)
			}
			if parsed.CompoundInterest == nil {
				t.Fatalf("Parse(%q) did not produce CompoundInterestExpr", tt.expr)
			}
			if got := parsed.CompoundInterest.Principal.Float64(); got != tt.principal {
				t.Errorf("Principal = %v, want %v", got, tt.principal)
			}
			if got := parsed.CompoundInterest.Rate.Float64(); got != tt.rate {
				t.Errorf("Rate = %v, want %v", got, tt.rate)
			}
			if got := parsed.CompoundInterest.Term.Years(); got != tt.years {
				t.Errorf("Years = %v, want %v", got, tt.years)
			}
			if got := parsed.CompoundInterest.Frequency; got != tt.frequency {
				t.Errorf("Frequency = %q, want %q", got, tt.frequency)
			}
		})
	}
}

func TestGrammarParseSimpleInterest(t *testing.T) {
	tests := []struct {
		expr      string
		principal float64
		rate      float64
		years     float64
	}{
		{"simple interest $5000 at 3% for 2 years", 5000, 3, 2},
		{"simple interest $10000 at 5% for 5 years", 10000, 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			parsed, err := Parse(tt.expr)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tt.expr, err)
			}
			if parsed.SimpleInterest == nil {
				t.Fatalf("Parse(%q) did not produce SimpleInterestExpr", tt.expr)
			}
			if got := parsed.SimpleInterest.Principal.Float64(); got != tt.principal {
				t.Errorf("Principal = %v, want %v", got, tt.principal)
			}
			if got := parsed.SimpleInterest.Rate.Float64(); got != tt.rate {
				t.Errorf("Rate = %v, want %v", got, tt.rate)
			}
			if got := parsed.SimpleInterest.Term.Years(); got != tt.years {
				t.Errorf("Years = %v, want %v", got, tt.years)
			}
		})
	}
}

func TestGrammarParseInvestment(t *testing.T) {
	tests := []struct {
		expr      string
		principal float64
		rate      float64
		years     float64
	}{
		{"invest $1000 at 7% for 20 years", 1000, 7, 20},
		{"invest $5000 at 10% for 10 years", 5000, 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			parsed, err := Parse(tt.expr)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tt.expr, err)
			}
			if parsed.Investment == nil {
				t.Fatalf("Parse(%q) did not produce InvestmentExpr", tt.expr)
			}
			if got := parsed.Investment.Principal.Float64(); got != tt.principal {
				t.Errorf("Principal = %v, want %v", got, tt.principal)
			}
			if got := parsed.Investment.Rate.Float64(); got != tt.rate {
				t.Errorf("Rate = %v, want %v", got, tt.rate)
			}
			if got := parsed.Investment.Term.Years(); got != tt.years {
				t.Errorf("Years = %v, want %v", got, tt.years)
			}
		})
	}
}

func TestGrammarEvaluate(t *testing.T) {
	tests := []struct {
		expr     string
		contains []string
	}{
		{"loan $250000 at 6.5% for 30 years", []string{"Monthly:", "$1,580", "Total:", "Interest:"}},
		{"mortgage $350000 at 7% for 30 years", []string{"Monthly:", "$2,328", "Payoff:"}},
		{"simple interest $5000 at 3% for 2 years", []string{"Interest:", "$300", "Total:", "$5,300"}},
		{"invest $1000 at 7% for 20 years", []string{"Final:", "$3,869", "Growth:"}},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, ok := EvalFinanceGrammar(tt.expr)
			if !ok {
				t.Fatalf("EvalFinanceGrammar(%q) returned false", tt.expr)
			}
			for _, c := range tt.contains {
				if !strings.Contains(result, c) {
					t.Errorf("Result missing %q\nGot: %s", c, result)
				}
			}
		})
	}
}

func TestGrammarParseInvalid(t *testing.T) {
	tests := []string{
		"100 + 50",
		"5 miles in km",
		"hello world",
		"loan without amount",
	}

	for _, expr := range tests {
		t.Run(expr, func(t *testing.T) {
			_, err := Parse(expr)
			if err == nil {
				t.Errorf("Parse(%q) should have failed", expr)
			}
		})
	}
}

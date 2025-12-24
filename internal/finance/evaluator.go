package finance

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"smartcalc/internal/utils"
)

// Evaluate evaluates a parsed financial expression and returns the result string.
func (e *FinanceExpr) Evaluate() (string, error) {
	switch {
	case e.Loan != nil:
		return e.Loan.Evaluate()
	case e.Mortgage != nil:
		return e.Mortgage.Evaluate()
	case e.CompoundInterest != nil:
		return e.CompoundInterest.Evaluate()
	case e.SimpleInterest != nil:
		return e.SimpleInterest.Evaluate()
	case e.Investment != nil:
		return e.Investment.Evaluate()
	default:
		return "", fmt.Errorf("unknown financial expression type")
	}
}

// Evaluate calculates loan payment details.
func (l *LoanExpr) Evaluate() (string, error) {
	principal := l.Principal.Float64()
	annualRate := l.Rate.Float64() / 100
	years := l.Term.Years()

	if principal == 0 || years == 0 {
		return "", fmt.Errorf("invalid loan parameters")
	}

	monthlyRate := annualRate / 12
	numPayments := int(years * 12)

	var monthlyPayment float64
	if monthlyRate == 0 {
		monthlyPayment = principal / float64(numPayments)
	} else {
		monthlyPayment = principal * (monthlyRate * math.Pow(1+monthlyRate, float64(numPayments))) /
			(math.Pow(1+monthlyRate, float64(numPayments)) - 1)
	}

	totalPayment := monthlyPayment * float64(numPayments)
	totalInterest := totalPayment - principal

	return fmt.Sprintf("\n> Monthly: %s\n> Total: %s\n> Interest: %s",
		utils.FormatCurrency(monthlyPayment),
		utils.FormatCurrency(totalPayment),
		utils.FormatCurrency(totalInterest)), nil
}

// Evaluate calculates mortgage payment details.
func (m *MortgageExpr) Evaluate() (string, error) {
	principal := m.Principal.Float64()
	annualRate := m.Rate.Float64() / 100
	years := m.Term.Years()

	if principal == 0 || years == 0 {
		return "", fmt.Errorf("invalid mortgage parameters")
	}

	// Handle pay schedule variant
	if m.PaySchedule {
		return m.evaluatePaySchedule(principal, annualRate, years)
	}

	// Handle extra payment variant
	if m.ExtraPayment != nil {
		return m.evaluateWithExtraPayment(principal, annualRate, years)
	}

	// Standard mortgage calculation
	monthlyRate := annualRate / 12
	numPayments := int(years * 12)

	var monthlyPayment float64
	if monthlyRate == 0 {
		monthlyPayment = principal / float64(numPayments)
	} else {
		monthlyPayment = principal * (monthlyRate * math.Pow(1+monthlyRate, float64(numPayments))) /
			(math.Pow(1+monthlyRate, float64(numPayments)) - 1)
	}

	totalPayment := monthlyPayment * float64(numPayments)
	totalInterest := totalPayment - principal

	startDate := time.Now()
	payoffDate := startDate.AddDate(0, numPayments, 0)

	return fmt.Sprintf("\n> Monthly: %s\n> Total: %s\n> Interest: %s\n> Payoff: %s",
		utils.FormatCurrency(monthlyPayment),
		utils.FormatCurrency(totalPayment),
		utils.FormatCurrency(totalInterest),
		payoffDate.Format("Jan 2006")), nil
}

func (m *MortgageExpr) evaluatePaySchedule(principal, annualRate, years float64) (string, error) {
	monthlyRate := annualRate / 12
	numPayments := int(years * 12)

	var monthlyPayment float64
	if monthlyRate == 0 {
		monthlyPayment = principal / float64(numPayments)
	} else {
		monthlyPayment = principal * (monthlyRate * math.Pow(1+monthlyRate, float64(numPayments))) /
			(math.Pow(1+monthlyRate, float64(numPayments)) - 1)
	}

	var sb strings.Builder
	sb.WriteString("\n> Payment Schedule:\n")
	sb.WriteString("> ──────────────────────────────────────────────────────────────\n")
	sb.WriteString("> Month      | Payment    | Principal  | Interest   | Balance\n")
	sb.WriteString("> ──────────────────────────────────────────────────────────────\n")

	balance := principal
	startDate := time.Now()
	totalInterest := 0.0

	for i := 1; i <= numPayments; i++ {
		interestPayment := balance * monthlyRate
		principalPayment := monthlyPayment - interestPayment
		balance -= principalPayment
		totalInterest += interestPayment

		if balance < 0 {
			balance = 0
		}

		paymentDate := startDate.AddDate(0, i, 0)
		sb.WriteString(fmt.Sprintf("> %s | %10s | %10s | %10s | %10s\n",
			paymentDate.Format("Jan 2006"),
			utils.FormatCurrency(monthlyPayment),
			utils.FormatCurrency(principalPayment),
			utils.FormatCurrency(interestPayment),
			utils.FormatCurrency(balance)))
	}

	sb.WriteString("> ──────────────────────────────────────────────────────────────\n")
	sb.WriteString(fmt.Sprintf("> Total Interest: %s", utils.FormatCurrency(totalInterest)))

	return sb.String(), nil
}

func (m *MortgageExpr) evaluateWithExtraPayment(principal, annualRate, years float64) (string, error) {
	extraPayment := m.ExtraPayment.Float64()
	monthlyRate := annualRate / 12
	numPayments := int(years * 12)

	var monthlyPayment float64
	if monthlyRate == 0 {
		monthlyPayment = principal / float64(numPayments)
	} else {
		monthlyPayment = principal * (monthlyRate * math.Pow(1+monthlyRate, float64(numPayments))) /
			(math.Pow(1+monthlyRate, float64(numPayments)) - 1)
	}

	// Calculate standard mortgage totals
	standardTotal := monthlyPayment * float64(numPayments)
	standardInterest := standardTotal - principal
	startDate := time.Now()
	standardPayoffDate := startDate.AddDate(0, numPayments, 0)

	// Calculate with extra payment
	balance := principal
	totalInterestWithExtra := 0.0
	monthsWithExtra := 0

	for balance > 0 {
		monthsWithExtra++
		interestPayment := balance * monthlyRate
		totalInterestWithExtra += interestPayment

		totalPaymentThisMonth := monthlyPayment + extraPayment
		principalPayment := totalPaymentThisMonth - interestPayment

		balance -= principalPayment
		if balance < 0 {
			balance = 0
		}

		if monthsWithExtra > numPayments*2 {
			break
		}
	}

	extraPayoffDate := startDate.AddDate(0, monthsWithExtra, 0)
	interestSavings := standardInterest - totalInterestWithExtra
	timeSaved := numPayments - monthsWithExtra

	yearsSaved := timeSaved / 12
	monthsSaved := timeSaved % 12
	var timeSavedStr string
	if yearsSaved > 0 && monthsSaved > 0 {
		timeSavedStr = fmt.Sprintf("%d years, %d months", yearsSaved, monthsSaved)
	} else if yearsSaved > 0 {
		timeSavedStr = fmt.Sprintf("%d years", yearsSaved)
	} else {
		timeSavedStr = fmt.Sprintf("%d months", monthsSaved)
	}

	return fmt.Sprintf("\n> Monthly: %s (+ %s extra)\n> Standard Interest: %s\n> With Extra Payment: %s\n> Interest Savings: %s\n> Standard Payoff: %s\n> New Payoff: %s\n> Time Saved: %s",
		utils.FormatCurrency(monthlyPayment),
		utils.FormatCurrency(extraPayment),
		utils.FormatCurrency(standardInterest),
		utils.FormatCurrency(totalInterestWithExtra),
		utils.FormatCurrency(interestSavings),
		standardPayoffDate.Format("Jan 2006"),
		extraPayoffDate.Format("Jan 2006"),
		timeSavedStr), nil
}

// Evaluate calculates compound interest.
func (c *CompoundInterestExpr) Evaluate() (string, error) {
	principal := c.Principal.Float64()
	annualRate := c.Rate.Float64() / 100
	years := c.Term.Years()

	if principal == 0 || years == 0 {
		return "", fmt.Errorf("invalid compound interest parameters")
	}

	frequency := c.Frequency
	if frequency == "" {
		frequency = "annually"
	}

	n := getCompoundingFrequency(frequency)
	amount := principal * math.Pow(1+annualRate/float64(n), float64(n)*years)
	interest := amount - principal

	return fmt.Sprintf("\n> Final: %s\n> Interest earned: %s",
		utils.FormatCurrency(amount),
		utils.FormatCurrency(interest)), nil
}

// Evaluate calculates simple interest.
func (s *SimpleInterestExpr) Evaluate() (string, error) {
	principal := s.Principal.Float64()
	rate := s.Rate.Float64() / 100
	years := s.Term.Years()

	if principal == 0 || years == 0 {
		return "", fmt.Errorf("invalid simple interest parameters")
	}

	interest := principal * rate * years
	total := principal + interest

	return fmt.Sprintf("\n> Interest: %s\n> Total: %s",
		utils.FormatCurrency(interest),
		utils.FormatCurrency(total)), nil
}

// Evaluate calculates investment growth.
func (inv *InvestmentExpr) Evaluate() (string, error) {
	principal := inv.Principal.Float64()
	annualRate := inv.Rate.Float64() / 100
	years := inv.Term.Years()

	if principal == 0 || years == 0 {
		return "", fmt.Errorf("invalid investment parameters")
	}

	// Assume annual compounding for simple invest command
	amount := principal * math.Pow(1+annualRate, years)
	growth := amount - principal
	growthPercent := (growth / principal) * 100

	return fmt.Sprintf("\n> Final: %s\n> Growth: %s (+%.1f%%)",
		utils.FormatCurrency(amount),
		utils.FormatCurrency(growth),
		growthPercent), nil
}

// Float64 converts an Amount to a float64 value.
func (a *Amount) Float64() float64 {
	if a == nil {
		return 0
	}
	// Remove commas from the number string
	cleaned := strings.ReplaceAll(a.Value, ",", "")
	val, _ := strconv.ParseFloat(cleaned, 64)
	return val
}

// Float64 converts a Rate to a float64 value (the percentage number, not decimal).
func (r *Rate) Float64() float64 {
	if r == nil {
		return 0
	}
	val, _ := strconv.ParseFloat(r.Value, 64)
	return val
}

// Years converts a Term to years as a float64.
func (t *Term) Years() float64 {
	if t == nil {
		return 0
	}
	val, _ := strconv.ParseFloat(t.Value, 64)
	unit := strings.ToLower(t.Unit)
	if strings.HasPrefix(unit, "month") {
		return val / 12
	}
	return val
}

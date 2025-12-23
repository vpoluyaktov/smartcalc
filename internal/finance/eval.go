package finance

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"smartcalc/internal/utils"
)

// Handler defines the interface for financial calculation handlers.
type Handler interface {
	Handle(expr, exprLower string) (string, bool)
}

// HandlerFunc is an adapter to allow ordinary functions to be used as Handlers.
type HandlerFunc func(expr, exprLower string) (string, bool)

// Handle calls the underlying function.
func (f HandlerFunc) Handle(expr, exprLower string) (string, bool) {
	return f(expr, exprLower)
}

// handlerChain is the ordered list of handlers for financial calculations.
var handlerChain = []Handler{
	HandlerFunc(handleLoanPayment),
	HandlerFunc(handleCompoundInterest),
	HandlerFunc(handleSimpleInterest),
	HandlerFunc(handleMortgagePayment),
	HandlerFunc(handleInvestmentGrowth),
}

// EvalFinance evaluates a financial expression and returns the result.
func EvalFinance(expr string) (string, error) {
	expr = strings.TrimSpace(expr)
	exprLower := strings.ToLower(expr)

	for _, h := range handlerChain {
		if result, ok := h.Handle(expr, exprLower); ok {
			return result, nil
		}
	}

	return "", fmt.Errorf("unable to evaluate financial expression: %s", expr)
}

// IsFinanceExpression checks if an expression looks like a financial calculation.
func IsFinanceExpression(expr string) bool {
	exprLower := strings.ToLower(expr)

	patterns := []string{
		`loan\s+\$?[\d,]+`,
		`mortgage\s+\$?[\d,]+`,
		`compound\s+interest`,
		`simple\s+interest`,
		`invest\s+\$?[\d,]+`,
		`\$[\d,]+\s+at\s+[\d.]+%`,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, exprLower); matched {
			return true
		}
	}

	return false
}

func handleLoanPayment(expr, exprLower string) (string, bool) {
	// Pattern: "loan $250000 at 6.5% for 30 years" or "loan 250000 at 6.5% for 30 years"
	re := regexp.MustCompile(`loan\s+\$?([\d,]+)\s+at\s+([\d.]+)%\s+for\s+(\d+)\s+years?`)
	matches := re.FindStringSubmatch(exprLower)
	if matches == nil {
		return "", false
	}

	principal := parseAmount(matches[1])
	annualRate := parseFloat(matches[2]) / 100
	years := parseInt(matches[3])

	if principal == 0 || years == 0 {
		return "", false
	}

	monthlyRate := annualRate / 12
	numPayments := years * 12

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
		utils.FormatCurrency(monthlyPayment), utils.FormatCurrency(totalPayment), utils.FormatCurrency(totalInterest)), true
}

func handleCompoundInterest(expr, exprLower string) (string, bool) {
	// Pattern: "$10000 at 5% for 10 years compounded monthly" or "compound interest $10000 at 5% for 10 years"
	re := regexp.MustCompile(`(?:compound\s+interest\s+)?\$?([\d,]+)\s+at\s+([\d.]+)%\s+for\s+(\d+)\s+years?\s*(?:compounded\s+)?(\w+)?`)
	matches := re.FindStringSubmatch(exprLower)
	if matches == nil {
		return "", false
	}

	// Must contain "compound" keyword
	if !strings.Contains(exprLower, "compound") {
		return "", false
	}

	principal := parseAmount(matches[1])
	annualRate := parseFloat(matches[2]) / 100
	years := parseInt(matches[3])
	compoundFreq := "annually"
	if len(matches) > 4 && matches[4] != "" {
		compoundFreq = matches[4]
	}

	if principal == 0 || years == 0 {
		return "", false
	}

	n := getCompoundingFrequency(compoundFreq)
	amount := principal * math.Pow(1+annualRate/float64(n), float64(n*years))
	interest := amount - principal

	return fmt.Sprintf("\n> Final: %s\n> Interest earned: %s", utils.FormatCurrency(amount), utils.FormatCurrency(interest)), true
}

func handleSimpleInterest(expr, exprLower string) (string, bool) {
	// Pattern: "simple interest $5000 at 3% for 2 years"
	re := regexp.MustCompile(`simple\s+interest\s+\$?([\d,]+)\s+at\s+([\d.]+)%\s+for\s+(\d+)\s+years?`)
	matches := re.FindStringSubmatch(exprLower)
	if matches == nil {
		return "", false
	}

	principal := parseAmount(matches[1])
	rate := parseFloat(matches[2]) / 100
	years := parseInt(matches[3])

	if principal == 0 || years == 0 {
		return "", false
	}

	interest := principal * rate * float64(years)
	total := principal + interest

	return fmt.Sprintf("\n> Interest: %s\n> Total: %s", utils.FormatCurrency(interest), utils.FormatCurrency(total)), true
}

func handleMortgagePayment(expr, exprLower string) (string, bool) {
	// Pattern: "mortgage $350000 at 7% for 30 years"
	re := regexp.MustCompile(`mortgage\s+\$?([\d,]+)\s+at\s+([\d.]+)%\s+for\s+(\d+)\s+years?`)
	matches := re.FindStringSubmatch(exprLower)
	if matches == nil {
		return "", false
	}

	principal := parseAmount(matches[1])
	annualRate := parseFloat(matches[2]) / 100
	years := parseInt(matches[3])

	if principal == 0 || years == 0 {
		return "", false
	}

	monthlyRate := annualRate / 12
	numPayments := years * 12

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
		utils.FormatCurrency(monthlyPayment), utils.FormatCurrency(totalPayment), utils.FormatCurrency(totalInterest)), true
}

func handleInvestmentGrowth(expr, exprLower string) (string, bool) {
	// Pattern: "invest $1000 at 7% for 20 years"
	re := regexp.MustCompile(`invest\s+\$?([\d,]+)\s+at\s+([\d.]+)%\s+for\s+(\d+)\s+years?`)
	matches := re.FindStringSubmatch(exprLower)
	if matches == nil {
		return "", false
	}

	principal := parseAmount(matches[1])
	annualRate := parseFloat(matches[2]) / 100
	years := parseInt(matches[3])

	if principal == 0 || years == 0 {
		return "", false
	}

	// Assume annual compounding for simple invest command
	amount := principal * math.Pow(1+annualRate, float64(years))
	growth := amount - principal
	growthPercent := (growth / principal) * 100

	return fmt.Sprintf("\n> Final: %s\n> Growth: %s (+%.1f%%)", utils.FormatCurrency(amount), utils.FormatCurrency(growth), growthPercent), true
}

func parseAmount(s string) float64 {
	s = strings.ReplaceAll(s, ",", "")
	val, _ := strconv.ParseFloat(s, 64)
	return val
}

func parseFloat(s string) float64 {
	val, _ := strconv.ParseFloat(s, 64)
	return val
}

func parseInt(s string) int {
	val, _ := strconv.Atoi(s)
	return val
}

func getCompoundingFrequency(freq string) int {
	switch strings.ToLower(freq) {
	case "daily":
		return 365
	case "weekly":
		return 52
	case "monthly":
		return 12
	case "quarterly":
		return 4
	case "semiannually", "semi-annually":
		return 2
	case "annually", "yearly":
		return 1
	default:
		return 1
	}
}

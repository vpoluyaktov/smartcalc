package finance

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"smartcalc/internal/utils"
)

// EvalFinanceGrammar attempts to evaluate a financial expression using the grammar-based parser.
// Returns the result string and true if successful, or empty string and false if parsing fails.
func EvalFinanceGrammar(expr string) (string, bool) {
	parsed, err := Parse(expr)
	if err != nil {
		return "", false
	}
	result, err := parsed.Evaluate()
	if err != nil {
		return "", false
	}
	return result, true
}

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
// It first tries the grammar-based parser, then falls back to the handler chain.
func EvalFinance(expr string) (string, error) {
	expr = strings.TrimSpace(expr)

	// Try grammar-based parsing first
	if result, ok := EvalFinanceGrammar(expr); ok {
		return result, nil
	}

	// Fall back to handler chain for expressions not covered by grammar
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
	// Check for extra payment variant first
	// Pattern: "mortgage $350000 at 7% for 30 years extra payment $500" or "extra $500"
	extraRe := regexp.MustCompile(`mortgage\s+\$?([\d,]+)\s+at\s+([\d.]+)%\s+for\s+(\d+)\s+years?\s+extra\s+(?:payment\s+)?\$?([\d,]+)`)
	extraMatches := extraRe.FindStringSubmatch(exprLower)
	if extraMatches != nil {
		return handleMortgageWithExtraPayment(extraMatches)
	}

	// Check for pay schedule variant
	// Pattern: "mortgage $350000 at 7% for 30 years pay schedule"
	scheduleRe := regexp.MustCompile(`mortgage\s+\$?([\d,]+)\s+at\s+([\d.]+)%\s+for\s+(\d+)\s+years?\s+pay\s+schedule`)
	scheduleMatches := scheduleRe.FindStringSubmatch(exprLower)
	if scheduleMatches != nil {
		return handleMortgagePaySchedule(scheduleMatches)
	}

	// Standard mortgage pattern: "mortgage $350000 at 7% for 30 years"
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

	// Calculate payoff date (assuming payments start next month)
	startDate := time.Now()
	payoffDate := startDate.AddDate(0, numPayments, 0)

	return fmt.Sprintf("\n> Monthly: %s\n> Total: %s\n> Interest: %s\n> Payoff: %s",
		utils.FormatCurrency(monthlyPayment), utils.FormatCurrency(totalPayment),
		utils.FormatCurrency(totalInterest), payoffDate.Format("Jan 2006")), true
}

func handleMortgagePaySchedule(matches []string) (string, bool) {
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

	// Build amortization schedule
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

		// Ensure balance doesn't go negative due to rounding
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

	return sb.String(), true
}

func handleMortgageWithExtraPayment(matches []string) (string, bool) {
	principal := parseAmount(matches[1])
	annualRate := parseFloat(matches[2]) / 100
	years := parseInt(matches[3])
	extraPayment := parseAmount(matches[4])

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

		// Apply regular payment + extra payment
		totalPaymentThisMonth := monthlyPayment + extraPayment
		principalPayment := totalPaymentThisMonth - interestPayment

		balance -= principalPayment
		if balance < 0 {
			balance = 0
		}

		// Safety check to prevent infinite loop
		if monthsWithExtra > numPayments*2 {
			break
		}
	}

	extraPayoffDate := startDate.AddDate(0, monthsWithExtra, 0)
	interestSavings := standardInterest - totalInterestWithExtra
	timeSaved := numPayments - monthsWithExtra

	// Format time saved
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
		utils.FormatCurrency(monthlyPayment), utils.FormatCurrency(extraPayment),
		utils.FormatCurrency(standardInterest), utils.FormatCurrency(totalInterestWithExtra),
		utils.FormatCurrency(interestSavings),
		standardPayoffDate.Format("Jan 2006"), extraPayoffDate.Format("Jan 2006"),
		timeSavedStr), true
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

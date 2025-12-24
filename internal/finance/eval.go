package finance

import (
	"fmt"
	"regexp"
	"strings"
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

// EvalFinance evaluates a financial expression and returns the result.
func EvalFinance(expr string) (string, error) {
	expr = strings.TrimSpace(expr)

	if result, ok := EvalFinanceGrammar(expr); ok {
		return result, nil
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

// getCompoundingFrequency returns the number of compounding periods per year.
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

package percentage

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"smartcalc/internal/utils"
)

// Handler defines the interface for percentage calculation handlers.
type Handler interface {
	Handle(expr, exprLower string) (string, bool)
}

// HandlerFunc is an adapter to allow ordinary functions to be used as Handlers.
type HandlerFunc func(expr, exprLower string) (string, bool)

// Handle calls the underlying function.
func (f HandlerFunc) Handle(expr, exprLower string) (string, bool) {
	return f(expr, exprLower)
}

// handlerChain is the ordered list of handlers for percentage calculations.
var handlerChain = []Handler{
	HandlerFunc(handleWhatIsPercentOf),
	HandlerFunc(handleWhatPercentIs),
	HandlerFunc(handleDecreaseByPercent), // must be before increase to avoid false matches
	HandlerFunc(handleIncreaseByPercent),
	HandlerFunc(handlePercentChange),
	HandlerFunc(handleTipCalculation),
	HandlerFunc(handleSplitBill),
}

// EvalPercentage evaluates a percentage expression and returns the result.
func EvalPercentage(expr string) (string, error) {
	expr = strings.TrimSpace(expr)
	exprLower := strings.ToLower(expr)

	for _, h := range handlerChain {
		if result, ok := h.Handle(expr, exprLower); ok {
			return result, nil
		}
	}

	return "", fmt.Errorf("unable to evaluate percentage expression: %s", expr)
}

// IsPercentageExpression checks if an expression looks like a percentage calculation.
func IsPercentageExpression(expr string) bool {
	exprLower := strings.ToLower(expr)

	patterns := []string{
		`what\s+is\s+[\d.]+%?\s+of`,
		`[\d.]+\s+is\s+what\s+(?:%|percent|percentage)`,
		`increase\s+[\d.]+\s+by`,
		`decrease\s+[\d.]+\s+by`,
		`percent\s+change`,
		`tip\s+[\d.]+%?\s+on`,
		`split\s+\$?[\d.]+`,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, exprLower); matched {
			return true
		}
	}

	return false
}

func handleWhatIsPercentOf(expr, exprLower string) (string, bool) {
	// Pattern: "what is 15% of 200" or "15% of 200"
	re := regexp.MustCompile(`(?:what\s+is\s+)?([\d.]+)\s*%?\s+of\s+([\d.]+)`)
	matches := re.FindStringSubmatch(exprLower)
	if matches == nil {
		return "", false
	}

	percent, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return "", false
	}

	value, err := strconv.ParseFloat(matches[2], 64)
	if err != nil {
		return "", false
	}

	result := value * percent / 100
	return formatResult(result), true
}

func handleWhatPercentIs(expr, exprLower string) (string, bool) {
	// Pattern: "50 is what % of 200" or "50 is what percent of 200"
	re := regexp.MustCompile(`([\d.]+)\s+is\s+what\s+(?:%|percent|percentage)\s+of\s+([\d.]+)`)
	matches := re.FindStringSubmatch(exprLower)
	if matches == nil {
		return "", false
	}

	part, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return "", false
	}

	whole, err := strconv.ParseFloat(matches[2], 64)
	if err != nil {
		return "", false
	}

	if whole == 0 {
		return "undefined (division by zero)", true
	}

	result := (part / whole) * 100
	return fmt.Sprintf("%.2f%%", result), true
}

func handleIncreaseByPercent(expr, exprLower string) (string, bool) {
	// Pattern: "increase 100 by 20%" or "100 increased by 20%"
	// Must contain "increase" keyword
	if !strings.Contains(exprLower, "increase") {
		return "", false
	}

	re := regexp.MustCompile(`(?:increase\s+)([\d.]+)\s+by\s+([\d.]+)\s*%`)
	matches := re.FindStringSubmatch(exprLower)
	if matches == nil {
		// Try alternate pattern: "100 increased by 20%"
		re = regexp.MustCompile(`([\d.]+)\s+increased\s+by\s+([\d.]+)\s*%`)
		matches = re.FindStringSubmatch(exprLower)
		if matches == nil {
			return "", false
		}
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return "", false
	}

	percent, err := strconv.ParseFloat(matches[2], 64)
	if err != nil {
		return "", false
	}

	result := value * (1 + percent/100)
	return formatResult(result), true
}

func handleDecreaseByPercent(expr, exprLower string) (string, bool) {
	// Pattern: "decrease 500 by 15%" or "500 decreased by 15%"
	// Must contain "decrease" keyword to avoid matching increase expressions
	if !strings.Contains(exprLower, "decrease") {
		return "", false
	}

	re := regexp.MustCompile(`(?:decrease\s+)([\d.]+)\s+by\s+([\d.]+)\s*%`)
	matches := re.FindStringSubmatch(exprLower)
	if matches == nil {
		// Try alternate pattern: "500 decreased by 15%"
		re = regexp.MustCompile(`([\d.]+)\s+decreased\s+by\s+([\d.]+)\s*%`)
		matches = re.FindStringSubmatch(exprLower)
		if matches == nil {
			return "", false
		}
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return "", false
	}

	percent, err := strconv.ParseFloat(matches[2], 64)
	if err != nil {
		return "", false
	}

	result := value * (1 - percent/100)
	return formatResult(result), true
}

func handlePercentChange(expr, exprLower string) (string, bool) {
	// Pattern: "percent change from 50 to 75" or "percentage change 100 to 150"
	re := regexp.MustCompile(`(?:percent(?:age)?\s+change\s+)?(?:from\s+)?([\d.]+)\s+to\s+([\d.]+)`)
	matches := re.FindStringSubmatch(exprLower)
	if matches == nil {
		return "", false
	}

	// Only match if it contains "percent change" or "percentage change"
	if !strings.Contains(exprLower, "percent") {
		return "", false
	}

	oldVal, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return "", false
	}

	newVal, err := strconv.ParseFloat(matches[2], 64)
	if err != nil {
		return "", false
	}

	if oldVal == 0 {
		return "undefined (division by zero)", true
	}

	change := ((newVal - oldVal) / oldVal) * 100
	sign := ""
	if change > 0 {
		sign = "+"
	}
	return fmt.Sprintf("%s%.2f%%", sign, change), true
}

func handleTipCalculation(expr, exprLower string) (string, bool) {
	// Pattern: "tip 20% on $85.50" or "20% tip on 85.50"
	re := regexp.MustCompile(`(?:tip\s+)?([\d.]+)\s*%\s*(?:tip\s+)?on\s+\$?([\d.]+)`)
	matches := re.FindStringSubmatch(exprLower)
	if matches == nil {
		return "", false
	}

	tipPercent, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return "", false
	}

	amount, err := strconv.ParseFloat(matches[2], 64)
	if err != nil {
		return "", false
	}

	tip := amount * tipPercent / 100
	total := amount + tip
	return fmt.Sprintf("Tip: $%.2f, Total: $%.2f", tip, total), true
}

func handleSplitBill(expr, exprLower string) (string, bool) {
	// Pattern: "$150 split 4 ways" or "split $150 4 ways" or "$150 split 4 ways with 18% tip"
	re := regexp.MustCompile(`(?:split\s+)?\$?([\d.]+)\s+split\s+(\d+)\s+ways?(?:\s+with\s+([\d.]+)\s*%\s*tip)?`)
	matches := re.FindStringSubmatch(exprLower)
	if matches == nil {
		// Try alternate pattern
		re = regexp.MustCompile(`split\s+\$?([\d.]+)\s+(\d+)\s+ways?(?:\s+with\s+([\d.]+)\s*%\s*tip)?`)
		matches = re.FindStringSubmatch(exprLower)
		if matches == nil {
			return "", false
		}
	}

	amount, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return "", false
	}

	ways, err := strconv.Atoi(matches[2])
	if err != nil || ways == 0 {
		return "", false
	}

	tipPercent := 0.0
	if len(matches) > 3 && matches[3] != "" {
		tipPercent, _ = strconv.ParseFloat(matches[3], 64)
	}

	tip := amount * tipPercent / 100
	total := amount + tip
	perPerson := total / float64(ways)

	if tipPercent > 0 {
		return fmt.Sprintf("Total: $%.2f (incl. $%.2f tip), Per person: $%.2f", total, tip, perPerson), true
	}
	return fmt.Sprintf("Per person: $%.2f", perPerson), true
}

func formatResult(value float64) string {
	return utils.FormatResult(false, value)
}

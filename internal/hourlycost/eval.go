package hourlycost

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	HoursPerDay   = 24.0
	HoursPerWeek  = 168.0  // 24 * 7
	HoursPerMonth = 720.0  // 24 * 30
	HoursPerYear  = 8760.0 // 24 * 365
)

// IsHourlyCostExpression checks if an expression is an hourly cost calculation.
// Pattern: "$X per hour in Y days/weeks/months/years" or "X cents per hour in Y ..."
func IsHourlyCostExpression(expr string) bool {
	exprLower := strings.ToLower(strings.TrimSpace(expr))

	// Pattern: $X per [a|an] hour in [Y] unit OR X cents/cent per [a|an] hour in [Y] unit
	// The number Y is optional (e.g., "in week" means "in 1 week")
	pattern := `^(?:\$[\d,.]+|[\d,.]+\s*(?:cents?|¢))\s+per\s+(?:a\s+|an\s+)?hours?\s+in\s+(?:[\d,.]+\s+)?(?:days?|weeks?|months?|years?)$`
	matched, _ := regexp.MatchString(pattern, exprLower)
	return matched
}

// EvalHourlyCost evaluates an hourly cost expression.
// Example: "$35 per hour in week" -> "$5,880.00"
// Example: "$45 per hour in 5 months" -> "$162,000.00"
// Example: "25 cents per hour in 2 years" -> "$4,380.00"
func EvalHourlyCost(expr string) (string, error) {
	exprLower := strings.ToLower(strings.TrimSpace(expr))

	// Pattern to extract: (rate) per hour in (duration) (unit)
	// Supports: $X, X cents, X cent, X¢
	// Duration is optional (defaults to 1)
	re := regexp.MustCompile(`^(\$[\d,.]+|[\d,.]+\s*(?:cents?|¢))\s+per\s+(?:a\s+|an\s+)?hours?\s+in\s+([\d,.]+\s+)?(days?|weeks?|months?|years?)$`)
	matches := re.FindStringSubmatch(exprLower)
	if matches == nil {
		return "", fmt.Errorf("unable to parse hourly cost expression: %s", expr)
	}

	rateStr := matches[1]
	durationStr := strings.TrimSpace(matches[2])
	unit := matches[3]

	// Default duration to 1 if not specified
	if durationStr == "" {
		durationStr = "1"
	}

	// Parse the hourly rate
	var hourlyRate float64
	var err error

	if strings.HasPrefix(rateStr, "$") {
		// Dollar amount
		cleanRate := strings.ReplaceAll(rateStr[1:], ",", "")
		hourlyRate, err = strconv.ParseFloat(cleanRate, 64)
		if err != nil {
			return "", fmt.Errorf("invalid rate value: %s", rateStr)
		}
	} else {
		// Cents amount
		cleanRate := strings.TrimSpace(rateStr)
		cleanRate = strings.TrimSuffix(cleanRate, "cents")
		cleanRate = strings.TrimSuffix(cleanRate, "cent")
		cleanRate = strings.TrimSuffix(cleanRate, "¢")
		cleanRate = strings.TrimSpace(cleanRate)
		cleanRate = strings.ReplaceAll(cleanRate, ",", "")
		centsValue, err := strconv.ParseFloat(cleanRate, 64)
		if err != nil {
			return "", fmt.Errorf("invalid cents value: %s", rateStr)
		}
		hourlyRate = centsValue / 100.0 // Convert cents to dollars
	}

	// Parse duration
	cleanDuration := strings.ReplaceAll(durationStr, ",", "")
	duration, err := strconv.ParseFloat(cleanDuration, 64)
	if err != nil {
		return "", fmt.Errorf("invalid duration value: %s", durationStr)
	}

	// Calculate total hours based on unit
	var totalHours float64
	switch {
	case strings.HasPrefix(unit, "day"):
		totalHours = duration * HoursPerDay
	case strings.HasPrefix(unit, "week"):
		totalHours = duration * HoursPerWeek
	case strings.HasPrefix(unit, "month"):
		totalHours = duration * HoursPerMonth
	case strings.HasPrefix(unit, "year"):
		totalHours = duration * HoursPerYear
	default:
		return "", fmt.Errorf("unknown time unit: %s", unit)
	}

	// Calculate total cost
	totalCost := hourlyRate * totalHours

	// Format result as currency
	return formatCurrency(totalCost), nil
}

// formatCurrency formats a float as a currency string with thousands separators
func formatCurrency(value float64) string {
	// Format with 2 decimal places
	str := fmt.Sprintf("%.2f", value)

	// Split into integer and decimal parts
	parts := strings.Split(str, ".")
	intPart := parts[0]
	decPart := parts[1]

	// Add thousands separators to integer part
	var result strings.Builder
	for i, digit := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(digit)
	}

	return "$" + result.String() + "." + decPart
}

package datetime

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// EvalDateTime evaluates a date/time expression and returns the result
func EvalDateTime(expr string) (string, error) {
	expr = strings.TrimSpace(expr)
	exprLower := strings.ToLower(expr)

	// Check for "now in <city>" pattern
	if strings.HasPrefix(exprLower, "now in ") || strings.HasPrefix(exprLower, "now() in ") {
		return evalNowIn(expr)
	}

	// Check for "now" or "now()"
	if exprLower == "now" || exprLower == "now()" {
		return FormatTime(time.Now()), nil
	}

	// Check for "today" or "today()"
	if exprLower == "today" || exprLower == "today()" {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).Format("2006-01-02"), nil
	}

	// Check for time conversion: "6:00 am Seattle in Kiev"
	if result, ok := tryTimeConversion(expr); ok {
		return result, nil
	}

	// Check for duration conversion: "861.5 hours in days"
	if result, ok := tryDurationConversion(expr); ok {
		return result, nil
	}

	// Check for date arithmetic: "today() - 35.9 days" or "2025-09-25 19:00:00 + 10 hours"
	if result, ok := tryDateArithmetic(expr); ok {
		return result, nil
	}

	// Check for datetime with timezone conversion: "2025-09-25 19:00:00 EST in Seattle"
	if result, ok := tryDateTimeConversion(expr); ok {
		return result, nil
	}

	// Check for date range: "Dec 6 till March 11"
	if result, ok := tryDateRange(expr); ok {
		return result, nil
	}

	// Check for duration multiplication: "(8 hours x 5 x 2) x 2" or "13 x 3 min"
	if result, ok := tryDurationMultiplication(expr); ok {
		return result, nil
	}

	return "", fmt.Errorf("unable to evaluate date/time expression: %s", expr)
}

// IsDateTimeExpression checks if an expression looks like a date/time expression
func IsDateTimeExpression(expr string) bool {
	exprLower := strings.ToLower(expr)

	// Keywords that indicate date/time
	keywords := []string{
		"now", "today", "yesterday", "tomorrow",
		"hours", "hour", "hrs", "hr",
		"minutes", "minute", "mins", "min",
		"seconds", "second", "secs", "sec",
		"days", "day",
		"weeks", "week",
		"months", "month",
		"years", "year", "yrs", "yr",
		" in ", " till ", " until ", " to ",
		"am", "pm",
	}

	for _, kw := range keywords {
		if strings.Contains(exprLower, kw) {
			return true
		}
	}

	// Check for date patterns
	datePatterns := []string{
		`\d{4}-\d{2}-\d{2}`,     // 2025-09-25
		`\d{1,2}/\d{1,2}/\d{4}`, // 09/25/2025
		`\d{1,2}:\d{2}`,         // 6:00
	}

	for _, pattern := range datePatterns {
		if matched, _ := regexp.MatchString(pattern, expr); matched {
			return true
		}
	}

	return false
}

func evalNowIn(expr string) (string, error) {
	// Extract city name
	re := regexp.MustCompile(`(?i)now(?:\(\))?\s+in\s+(.+)`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", fmt.Errorf("invalid 'now in' expression")
	}

	city := strings.TrimSpace(matches[1])
	loc, err := LookupTimezone(city)
	if err != nil {
		return "", fmt.Errorf("unknown timezone/city: %s", city)
	}

	return FormatTime(time.Now().In(loc)), nil
}

func tryTimeConversion(expr string) (string, bool) {
	// Pattern: "6:00 am Seattle in Kiev" or "11am kiev in seattle"
	// More flexible pattern to handle various time formats
	re := regexp.MustCompile(`(?i)^(\d{1,2}(?::\d{2})?(?::\d{2})?\s*(?:am|pm)?)\s+(.+?)\s+in\s+(.+)$`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	timeStr := strings.TrimSpace(matches[1])
	fromCity := strings.TrimSpace(matches[2])
	toCity := strings.TrimSpace(matches[3])

	fromLoc, err := LookupTimezone(fromCity)
	if err != nil {
		return "", false
	}

	toLoc, err := LookupTimezone(toCity)
	if err != nil {
		return "", false
	}

	// Parse the time
	t, err := ParseDateTime(timeStr, fromLoc)
	if err != nil {
		return "", false
	}

	// Convert to target timezone
	result := t.In(toLoc)
	return FormatTime(result), true
}

func tryDurationConversion(expr string) (string, bool) {
	// Pattern: "861.5 hours in days"
	re := regexp.MustCompile(`(?i)^([\d.]+)\s*(seconds?|secs?|minutes?|mins?|hours?|hrs?|days?|weeks?|months?|years?|yrs?)\s+in\s+(seconds?|secs?|minutes?|mins?|hours?|hrs?|days?|weeks?|months?|years?|yrs?)$`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return "", false
	}

	fromUnit := matches[2]
	toUnit := matches[3]

	// Parse as duration
	d, err := ParseDuration(fmt.Sprintf("%f %s", value, fromUnit))
	if err != nil {
		return "", false
	}

	// Convert to target unit
	result, err := ConvertDuration(d, toUnit)
	if err != nil {
		return "", false
	}

	// Format nicely
	if result == float64(int(result)) {
		return fmt.Sprintf("%.0f %s", result, toUnit), true
	}
	return fmt.Sprintf("%.2f %s", result, toUnit), true
}

func tryDateArithmetic(expr string) (string, bool) {
	// Pattern: "today() - 35.9 days" or "2025-09-25 19:00:00 + 10 hours"
	re := regexp.MustCompile(`(?i)^(.+?)\s*([+−-])\s*([\d.]+)\s*(seconds?|secs?|minutes?|mins?|hours?|hrs?|days?|weeks?|months?|years?|yrs?)$`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	dateExpr := strings.TrimSpace(matches[1])
	op := matches[2]
	value, _ := strconv.ParseFloat(matches[3], 64)
	unit := matches[4]

	// Parse the date
	var baseTime time.Time
	dateLower := strings.ToLower(dateExpr)

	if dateLower == "today" || dateLower == "today()" {
		now := time.Now()
		baseTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	} else if dateLower == "now" || dateLower == "now()" {
		baseTime = time.Now()
	} else {
		var err error
		baseTime, err = ParseDateTime(dateExpr, time.Local)
		if err != nil {
			return "", false
		}
	}

	// Parse duration
	d, err := ParseDuration(fmt.Sprintf("%f %s", value, unit))
	if err != nil {
		return "", false
	}

	// Apply operation
	if op == "-" || op == "−" {
		baseTime = baseTime.Add(-d)
	} else {
		baseTime = baseTime.Add(d)
	}

	return FormatTime(baseTime), true
}

func tryDateTimeConversion(expr string) (string, bool) {
	// Pattern: "2025-09-25 19:00:00 EST in Seattle"
	re := regexp.MustCompile(`(?i)^(.+?)\s+([A-Z]{2,4})\s+in\s+(\w+(?:\s+\w+)?)$`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	dateTimeStr := strings.TrimSpace(matches[1])
	fromTz := strings.TrimSpace(matches[2])
	toCity := strings.TrimSpace(matches[3])

	fromLoc, err := LookupTimezone(fromTz)
	if err != nil {
		return "", false
	}

	toLoc, err := LookupTimezone(toCity)
	if err != nil {
		return "", false
	}

	t, err := ParseDateTime(dateTimeStr, fromLoc)
	if err != nil {
		return "", false
	}

	return FormatTime(t.In(toLoc)), true
}

func tryDateRange(expr string) (string, bool) {
	start, end, err := ParseDateRange(expr)
	if err != nil {
		return "", false
	}

	days := DaysBetween(start, end)
	if days == float64(int(days)) {
		return fmt.Sprintf("%.0f days", days), true
	}
	return fmt.Sprintf("%.1f days", days), true
}

func tryDurationMultiplication(expr string) (string, bool) {
	// Handle expressions like "(8 hours x 5 x 2) x 2" or "13 x 3 min"
	exprLower := strings.ToLower(expr)

	// Check if it contains duration units
	hasUnit := false
	for _, unit := range []string{"hour", "hr", "min", "sec", "day", "week"} {
		if strings.Contains(exprLower, unit) {
			hasUnit = true
			break
		}
	}
	if !hasUnit {
		return "", false
	}

	// Simple pattern: "number x number unit" like "13 x 3 min"
	re := regexp.MustCompile(`(?i)^([\d.]+)\s*[x×*]\s*([\d.]+)\s*(seconds?|secs?|minutes?|mins?|hours?|hrs?|days?|weeks?)$`)
	if matches := re.FindStringSubmatch(expr); matches != nil {
		v1, _ := strconv.ParseFloat(matches[1], 64)
		v2, _ := strconv.ParseFloat(matches[2], 64)
		unit := matches[3]

		result := v1 * v2
		d, err := ParseDuration(fmt.Sprintf("%f %s", result, unit))
		if err != nil {
			return "", false
		}
		return FormatDuration(d), true
	}

	// Pattern: "number unit x number" like "8 hours x 5"
	re = regexp.MustCompile(`(?i)^([\d.]+)\s*(seconds?|secs?|minutes?|mins?|hours?|hrs?|days?|weeks?)\s*[x×*]\s*([\d.]+)$`)
	if matches := re.FindStringSubmatch(expr); matches != nil {
		v1, _ := strconv.ParseFloat(matches[1], 64)
		unit := matches[2]
		v2, _ := strconv.ParseFloat(matches[3], 64)

		d, err := ParseDuration(fmt.Sprintf("%f %s", v1, unit))
		if err != nil {
			return "", false
		}
		result := time.Duration(float64(d) * v2)
		return FormatDuration(result), true
	}

	// More complex expressions with parentheses - simplified handling
	// "(8 hours x 5 x 2) x 2"
	re = regexp.MustCompile(`(?i)^\(([\d.]+)\s*(hours?|hrs?|minutes?|mins?|days?)\s*[x×*]\s*([\d.]+)\s*[x×*]\s*([\d.]+)\)\s*[x×*]\s*([\d.]+)$`)
	if matches := re.FindStringSubmatch(expr); matches != nil {
		baseVal, _ := strconv.ParseFloat(matches[1], 64)
		unit := matches[2]
		m1, _ := strconv.ParseFloat(matches[3], 64)
		m2, _ := strconv.ParseFloat(matches[4], 64)
		m3, _ := strconv.ParseFloat(matches[5], 64)

		d, err := ParseDuration(fmt.Sprintf("%f %s", baseVal, unit))
		if err != nil {
			return "", false
		}
		result := time.Duration(float64(d) * m1 * m2 * m3)
		return FormatDuration(result), true
	}

	return "", false
}

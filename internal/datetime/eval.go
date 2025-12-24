package datetime

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// RefResolver is a function that resolves line references like \1 to their string values
type RefResolver func(n int) (string, bool)

// Handler defines the interface for datetime expression handlers.
// Each handler attempts to process an expression and returns the result
// along with a boolean indicating whether it handled the expression.
type Handler interface {
	Handle(expr, exprLower string) (string, bool)
}

// HandlerFunc is an adapter to allow ordinary functions to be used as Handlers.
type HandlerFunc func(expr, exprLower string) (string, bool)

// Handle calls the underlying function.
func (f HandlerFunc) Handle(expr, exprLower string) (string, bool) {
	return f(expr, exprLower)
}

// handlerChain is the ordered list of handlers for datetime expressions.
// Handlers are tried in order; the first one that returns ok=true wins.
var handlerChain = []Handler{
	HandlerFunc(handleNowIn),
	HandlerFunc(handleNow),
	HandlerFunc(handleToday),
	HandlerFunc(handleNumberPlusDuration),
	HandlerFunc(handleTimeConversion),
	HandlerFunc(handleDurationConversion),
	HandlerFunc(handleDateArithmetic),
	HandlerFunc(handleDateTimeConversion),
	HandlerFunc(handleDateRange),
	HandlerFunc(handleDateDifference),
	HandlerFunc(handleDurationMultiplication),
}

// EvalDateTimeWithRefs evaluates a date/time expression with line reference support
func EvalDateTimeWithRefs(expr string, resolver RefResolver) (string, error) {
	// First, replace any line references with their values
	if resolver != nil {
		expr = resolveRefsInExpr(expr, resolver)
	}
	return EvalDateTime(expr)
}

// resolveRefsInExpr replaces \n references with their resolved values
func resolveRefsInExpr(expr string, resolver RefResolver) string {
	result := expr
	for i := 0; i < len(result); i++ {
		if result[i] != '\\' {
			continue
		}
		j := i + 1
		if j >= len(result) || result[j] < '0' || result[j] > '9' {
			continue
		}
		n := 0
		for j < len(result) && result[j] >= '0' && result[j] <= '9' {
			n = n*10 + int(result[j]-'0')
			j++
		}
		if val, ok := resolver(n); ok {
			result = result[:i] + val + result[j:]
			i += len(val) - 1
		}
	}
	return result
}

// EvalDateTime evaluates a date/time expression and returns the result.
// It uses the Chain of Responsibility pattern to delegate to handlers.
func EvalDateTime(expr string) (string, error) {
	expr = strings.TrimSpace(expr)
	exprLower := strings.ToLower(expr)

	for _, h := range handlerChain {
		if result, ok := h.Handle(expr, exprLower); ok {
			return result, nil
		}
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

func handleNowIn(expr, exprLower string) (string, bool) {
	// Check for "now in <city>" pattern
	if !strings.HasPrefix(exprLower, "now in ") && !strings.HasPrefix(exprLower, "now() in ") {
		return "", false
	}

	// Extract city name
	re := regexp.MustCompile(`(?i)now(?:\(\))?\s+in\s+(.+)`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	city := strings.TrimSpace(matches[1])
	loc, err := LookupTimezone(city)
	if err != nil {
		return "", false
	}

	return FormatTime(time.Now().In(loc)), true
}

func handleNow(expr, exprLower string) (string, bool) {
	if exprLower == "now" || exprLower == "now()" {
		return FormatTime(time.Now()), true
	}
	return "", false
}

func handleToday(expr, exprLower string) (string, bool) {
	if exprLower == "today" || exprLower == "today()" {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).Format("2006-01-02"), true
	}
	return "", false
}

func handleTimeConversion(expr, exprLower string) (string, bool) {
	// Pattern: "6:00 am Seattle in Kiev" or "11am kiev in seattle" or "2:00 am UTC to PST"
	// More flexible pattern to handle various time formats
	// Supports both "in" and "to" as separators
	re := regexp.MustCompile(`(?i)^(\d{1,2}(?::\d{2})?(?::\d{2})?\s*(?:am|pm)?)\s+(.+?)\s+(?:in|to)\s+(.+)$`)
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

func handleDurationConversion(expr, exprLower string) (string, bool) {
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

func handleDateArithmetic(expr, exprLower string) (string, bool) {
	// Pattern: "today() - 35.9 days" or "2025-09-25 19:00:00 + 10 hours" or "2025-12-17 16:00:00 PST + 3 days"
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
		// Try to parse time with timezone like "12 am PST" or "3:00 pm EST"
		if t, ok := parseTimeWithTimezone(dateExpr); ok {
			baseTime = t
		} else {
			var err error
			baseTime, err = ParseDateTime(dateExpr, time.Local)
			if err != nil {
				return "", false
			}
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

// parseTimeWithTimezone parses time expressions like "12 am PST", "3:00 pm EST", "14:00 UTC"
func parseTimeWithTimezone(expr string) (time.Time, bool) {
	// Pattern: time followed by timezone
	re := regexp.MustCompile(`(?i)^(\d{1,2}(?::\d{2})?(?::\d{2})?\s*(?:am|pm)?)\s+([A-Za-z]+(?:\s+[A-Za-z]+)?)$`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return time.Time{}, false
	}

	timeStr := strings.TrimSpace(matches[1])
	tzStr := strings.TrimSpace(matches[2])

	// Look up timezone
	loc, err := LookupTimezone(tzStr)
	if err != nil {
		return time.Time{}, false
	}

	// Parse the time part
	timeFormats := []string{
		"3:04pm", "3:04 pm", "3pm", "3 pm",
		"15:04", "15:04:05",
		"3:04:05pm", "3:04:05 pm",
	}

	for _, format := range timeFormats {
		if t, err := time.ParseInLocation(format, strings.ToLower(timeStr), loc); err == nil {
			now := time.Now().In(loc)
			return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), 0, loc), true
		}
	}

	return time.Time{}, false
}

func handleDateTimeConversion(expr, exprLower string) (string, bool) {
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

func handleDateRange(expr, exprLower string) (string, bool) {
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

func handleDateDifference(expr, exprLower string) (string, bool) {
	// Pattern: "date1 - date2" or "date1 − date2" (with minus or en-dash)
	// Examples: "19/01/22 - now", "2020-01-15 - today", "now - 2020-01-15"

	// Split by minus sign (handle both regular minus and en-dash)
	var date1Str, date2Str string
	if idx := strings.Index(expr, " − "); idx > 0 {
		date1Str = strings.TrimSpace(expr[:idx])
		date2Str = strings.TrimSpace(expr[idx+len(" − "):])
	} else if idx := strings.Index(expr, " - "); idx > 0 {
		date1Str = strings.TrimSpace(expr[:idx])
		date2Str = strings.TrimSpace(expr[idx+3:])
	} else {
		return "", false
	}

	// Parse date1
	var date1 time.Time
	date1Lower := strings.ToLower(date1Str)
	if date1Lower == "now" || date1Lower == "now()" {
		date1 = time.Now()
	} else if date1Lower == "today" || date1Lower == "today()" {
		now := time.Now()
		date1 = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	} else {
		var err error
		date1, err = ParseDateTime(date1Str, time.Local)
		if err != nil {
			return "", false
		}
	}

	// Parse date2
	var date2 time.Time
	date2Lower := strings.ToLower(date2Str)
	if date2Lower == "now" || date2Lower == "now()" {
		date2 = time.Now()
	} else if date2Lower == "today" || date2Lower == "today()" {
		now := time.Now()
		date2 = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	} else {
		var err error
		date2, err = ParseDateTime(date2Str, time.Local)
		if err != nil {
			return "", false
		}
	}

	// Calculate and format the difference
	return FormatDetailedDuration(date1, date2), true
}

func handleNumberPlusDuration(expr, exprLower string) (string, bool) {
	// Pattern: "0 + 3 days" or "0 - 5 hours" - treat 0 as "now"
	re := regexp.MustCompile(`(?i)^(\d+)\s*([+−-])\s*([\d.]+)\s*(seconds?|secs?|minutes?|mins?|hours?|hrs?|days?|weeks?|months?|years?|yrs?)$`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	baseNum, _ := strconv.ParseFloat(matches[1], 64)
	op := matches[2]
	value, _ := strconv.ParseFloat(matches[3], 64)
	unit := matches[4]

	// If base is 0, treat as "now"
	var baseTime time.Time
	if baseNum == 0 {
		baseTime = time.Now()
	} else {
		// Otherwise, this isn't a datetime expression
		return "", false
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

func handleDurationMultiplication(expr, exprLower string) (string, bool) {
	// Handle expressions like "(8 hours x 5 x 2) x 2" or "13 x 3 min"

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

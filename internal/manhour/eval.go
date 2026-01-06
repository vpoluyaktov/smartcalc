package manhour

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	// Business time constants (8-hour workday, 5-day workweek)
	HoursPerBusinessDay   = 8.0
	DaysPerBusinessWeek   = 5.0
	HoursPerBusinessWeek  = HoursPerBusinessDay * DaysPerBusinessWeek  // 40 hours
	DaysPerBusinessMonth  = 20.0                                       // ~4 weeks
	HoursPerBusinessMonth = HoursPerBusinessDay * DaysPerBusinessMonth // 160 hours

	// Calendar time constants (24-hour day, 7-day week)
	HoursPerCalendarDay   = 24.0
	DaysPerCalendarWeek   = 7.0
	HoursPerCalendarWeek  = HoursPerCalendarDay * DaysPerCalendarWeek  // 168 hours
	DaysPerCalendarMonth  = 30.0                                       // approximate
	HoursPerCalendarMonth = HoursPerCalendarDay * DaysPerCalendarMonth // 720 hours
)

// IsManHourExpression checks if an expression is a man-hour calculation.
// Pattern: "X man-hour(s) / Y men/man in [business|calendar] weeks/days/months"
func IsManHourExpression(expr string) bool {
	exprLower := strings.ToLower(strings.TrimSpace(expr))

	// Pattern: number man-hour(s) / number men/man in [business|calendar] weeks/days/months
	// "business" prefix is optional for business time, "calendar" is explicit for calendar time
	// Without prefix, defaults to calendar time (e.g., "in weeks" = "in calendar weeks")
	pattern := `^\d+(?:\.\d+)?\s*man[- ]?hours?\s*/\s*\d+(?:\.\d+)?\s*(?:men|man|person|persons|people)\s+in\s+(?:business\s+|calendar\s+)?(?:weeks?|days?|months?)$`
	matched, _ := regexp.MatchString(pattern, exprLower)
	return matched
}

// EvalManHour evaluates a man-hour expression.
// Example: "248 man-hour / 3 men in business weeks" -> "2.07 business weeks"
// Example: "248 man-hour / 3 men in weeks" -> "0.49 weeks" (calendar weeks)
func EvalManHour(expr string) (string, error) {
	exprLower := strings.ToLower(strings.TrimSpace(expr))

	// Pattern: (manhours) man-hour(s) / (people) men/man in [business|calendar] (unit)
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*man[- ]?hours?\s*/\s*(\d+(?:\.\d+)?)\s*(?:men|man|person|persons|people)\s+in\s+(business\s+|calendar\s+)?(weeks?|days?|months?)$`)
	matches := re.FindStringSubmatch(exprLower)
	if matches == nil {
		return "", fmt.Errorf("unable to parse man-hour expression: %s", expr)
	}

	manHours, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return "", fmt.Errorf("invalid man-hours value: %s", matches[1])
	}

	people, err := strconv.ParseFloat(matches[2], 64)
	if err != nil {
		return "", fmt.Errorf("invalid people count: %s", matches[2])
	}

	if people == 0 {
		return "", fmt.Errorf("cannot divide by zero people")
	}

	timeType := strings.TrimSpace(matches[3]) // "business", "calendar", or ""
	targetUnit := matches[4]

	// Calculate hours per person
	hoursPerPerson := manHours / people

	var result float64
	var unitLabel string

	// Determine if business or calendar time
	isBusiness := strings.HasPrefix(timeType, "business")
	isCalendar := strings.HasPrefix(timeType, "calendar") || timeType == ""

	if isBusiness {
		switch {
		case strings.HasPrefix(targetUnit, "month"):
			result = hoursPerPerson / HoursPerBusinessMonth
			unitLabel = pluralize(result, "business month", "business months")
		case strings.HasPrefix(targetUnit, "week"):
			result = hoursPerPerson / HoursPerBusinessWeek
			unitLabel = pluralize(result, "business week", "business weeks")
		case strings.HasPrefix(targetUnit, "day"):
			result = hoursPerPerson / HoursPerBusinessDay
			unitLabel = pluralize(result, "business day", "business days")
		default:
			return "", fmt.Errorf("unknown target unit: %s", targetUnit)
		}
	} else if isCalendar {
		switch {
		case strings.HasPrefix(targetUnit, "month"):
			result = hoursPerPerson / HoursPerCalendarMonth
			unitLabel = pluralize(result, "month", "months")
		case strings.HasPrefix(targetUnit, "week"):
			result = hoursPerPerson / HoursPerCalendarWeek
			unitLabel = pluralize(result, "week", "weeks")
		case strings.HasPrefix(targetUnit, "day"):
			result = hoursPerPerson / HoursPerCalendarDay
			unitLabel = pluralize(result, "day", "days")
		default:
			return "", fmt.Errorf("unknown target unit: %s", targetUnit)
		}
	}

	// Format result
	if result == float64(int(result)) {
		return fmt.Sprintf("%.0f %s", result, unitLabel), nil
	}
	return fmt.Sprintf("%.2f %s", result, unitLabel), nil
}

// pluralize returns singular or plural form based on value
func pluralize(value float64, singular, plural string) string {
	if value == 1 {
		return singular
	}
	return plural
}

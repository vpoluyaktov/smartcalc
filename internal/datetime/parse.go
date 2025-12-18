package datetime

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Common date/time formats to try when parsing
var dateFormats = []string{
	"2006-01-02 15:04:05 MST",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006-01-02",
	"01/02/2006 15:04:05 MST",
	"01/02/2006 15:04:05",
	"01/02/2006 15:04",
	"01/02/2006",
	"02/01/2006", // dd/mm/yyyy
	"02/01/06",   // dd/mm/yy
	"Jan 2, 2006 15:04:05 MST",
	"Jan 2, 2006 15:04:05",
	"Jan 2, 2006 15:04",
	"Jan 2, 2006",
	"Jan 2 2006",
	"January 2, 2006",
	"January 2 2006",
	"2 Jan 2006",
	"2 January 2006",
}

// Time-only formats
var timeFormats = []string{
	"3:04pm",
	"3:04 pm",
	"3pm",
	"3 pm",
	"15:04",
	"15:04:05",
	"3:04:05pm",
	"3:04:05 pm",
}

// Month name to number mapping
var monthNames = map[string]time.Month{
	"jan": time.January, "january": time.January,
	"feb": time.February, "february": time.February,
	"mar": time.March, "march": time.March,
	"apr": time.April, "april": time.April,
	"may": time.May,
	"jun": time.June, "june": time.June,
	"jul": time.July, "july": time.July,
	"aug": time.August, "august": time.August,
	"sep": time.September, "september": time.September,
	"oct": time.October, "october": time.October,
	"nov": time.November, "november": time.November,
	"dec": time.December, "december": time.December,
}

// ParseDateTime attempts to parse a date/time string
func ParseDateTime(s string, defaultLoc *time.Location) (time.Time, error) {
	s = strings.TrimSpace(s)
	if defaultLoc == nil {
		defaultLoc = time.Local
	}

	// Try each format
	for _, format := range dateFormats {
		if t, err := time.ParseInLocation(format, s, defaultLoc); err == nil {
			return t, nil
		}
	}

	// Try time-only formats (use today's date)
	for _, format := range timeFormats {
		if t, err := time.ParseInLocation(format, strings.ToLower(s), defaultLoc); err == nil {
			now := time.Now().In(defaultLoc)
			return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), 0, defaultLoc), nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date/time: %s", s)
}

// ParseDuration parses duration expressions like "5 hours", "3.5 days", "30 minutes"
func ParseDuration(s string) (time.Duration, error) {
	s = strings.ToLower(strings.TrimSpace(s))

	// Pattern: number followed by unit
	re := regexp.MustCompile(`^([\d.]+)\s*(seconds?|secs?|s|minutes?|mins?|m|hours?|hrs?|h|days?|d|weeks?|w|months?|years?|yrs?|y)$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return 0, fmt.Errorf("unable to parse duration: %s", s)
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, err
	}

	unit := matches[2]
	switch {
	case strings.HasPrefix(unit, "sec") || unit == "s":
		return time.Duration(value * float64(time.Second)), nil
	case strings.HasPrefix(unit, "min") || unit == "m":
		return time.Duration(value * float64(time.Minute)), nil
	case strings.HasPrefix(unit, "hour") || strings.HasPrefix(unit, "hr") || unit == "h":
		return time.Duration(value * float64(time.Hour)), nil
	case strings.HasPrefix(unit, "day") || unit == "d":
		return time.Duration(value * 24 * float64(time.Hour)), nil
	case strings.HasPrefix(unit, "week") || unit == "w":
		return time.Duration(value * 7 * 24 * float64(time.Hour)), nil
	case strings.HasPrefix(unit, "month"):
		// Approximate: 30.44 days per month
		return time.Duration(value * 30.44 * 24 * float64(time.Hour)), nil
	case strings.HasPrefix(unit, "year") || strings.HasPrefix(unit, "yr") || unit == "y":
		// Approximate: 365.25 days per year
		return time.Duration(value * 365.25 * 24 * float64(time.Hour)), nil
	}

	return 0, fmt.Errorf("unknown duration unit: %s", unit)
}

// ConvertDuration converts a duration to a specific unit and returns the value
func ConvertDuration(d time.Duration, toUnit string) (float64, error) {
	toUnit = strings.ToLower(strings.TrimSpace(toUnit))

	switch {
	case strings.HasPrefix(toUnit, "sec") || toUnit == "s":
		return d.Seconds(), nil
	case strings.HasPrefix(toUnit, "min") || toUnit == "m":
		return d.Minutes(), nil
	case strings.HasPrefix(toUnit, "hour") || strings.HasPrefix(toUnit, "hr") || toUnit == "h":
		return d.Hours(), nil
	case strings.HasPrefix(toUnit, "day") || toUnit == "d":
		return d.Hours() / 24, nil
	case strings.HasPrefix(toUnit, "week") || toUnit == "w":
		return d.Hours() / (24 * 7), nil
	case strings.HasPrefix(toUnit, "month"):
		return d.Hours() / (24 * 30.44), nil
	case strings.HasPrefix(toUnit, "year") || strings.HasPrefix(toUnit, "yr") || toUnit == "y":
		return d.Hours() / (24 * 365.25), nil
	}

	return 0, fmt.Errorf("unknown duration unit: %s", toUnit)
}

// ParseDateRange parses expressions like "Dec 6 till March 11"
func ParseDateRange(s string) (time.Time, time.Time, error) {
	s = strings.ToLower(strings.TrimSpace(s))

	// Find separator
	var start, end string
	for _, sep := range []string{" till ", " until ", " to ", " - ", " through "} {
		if idx := strings.Index(s, sep); idx > 0 {
			start = s[:idx]
			end = s[idx+len(sep):]
			break
		}
	}

	if start == "" || end == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("unable to parse date range: %s", s)
	}

	startTime, err := parsePartialDate(start)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	endTime, err := parsePartialDate(end)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	// If end is before start, assume it's next year
	if endTime.Before(startTime) {
		endTime = endTime.AddDate(1, 0, 0)
	}

	return startTime, endTime, nil
}

// parsePartialDate parses dates like "Dec 6" or "March 11"
func parsePartialDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)

	// Try "Month Day" format
	re := regexp.MustCompile(`^([a-zA-Z]+)\s+(\d+)$`)
	if matches := re.FindStringSubmatch(s); matches != nil {
		monthStr := strings.ToLower(matches[1])
		day, _ := strconv.Atoi(matches[2])

		if month, ok := monthNames[monthStr]; ok {
			now := time.Now()
			return time.Date(now.Year(), month, day, 0, 0, 0, 0, time.Local), nil
		}
	}

	// Try "Day Month" format
	re = regexp.MustCompile(`^(\d+)\s+([a-zA-Z]+)$`)
	if matches := re.FindStringSubmatch(s); matches != nil {
		day, _ := strconv.Atoi(matches[1])
		monthStr := strings.ToLower(matches[2])

		if month, ok := monthNames[monthStr]; ok {
			now := time.Now()
			return time.Date(now.Year(), month, day, 0, 0, 0, 0, time.Local), nil
		}
	}

	// Try full date parsing
	return ParseDateTime(s, time.Local)
}

// FormatTime formats a time for display (truncated to minutes to avoid constant updates)
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04 MST")
}

// FormatDuration formats a duration for display
func FormatDuration(d time.Duration) string {
	if d < 0 {
		return "-" + FormatDuration(-d)
	}

	days := d.Hours() / 24
	if days >= 1 {
		hours := d.Hours() - float64(int(days))*24
		if hours > 0 {
			return fmt.Sprintf("%.0f days %.1f hours", days, hours)
		}
		if days == float64(int(days)) {
			return fmt.Sprintf("%.0f days", days)
		}
		return fmt.Sprintf("%.2f days", days)
	}

	if d.Hours() >= 1 {
		return fmt.Sprintf("%.2f hours", d.Hours())
	}

	if d.Minutes() >= 1 {
		return fmt.Sprintf("%.2f minutes", d.Minutes())
	}

	return fmt.Sprintf("%.2f seconds", d.Seconds())
}

// DaysBetween calculates the number of days between two dates
func DaysBetween(start, end time.Time) float64 {
	return end.Sub(start).Hours() / 24
}

// FormatDetailedDuration formats a duration between two dates showing years, months, weeks, days, hours, minutes
func FormatDetailedDuration(from, to time.Time) string {
	if to.Before(from) {
		from, to = to, from
	}

	// Calculate years
	years := 0
	for {
		next := from.AddDate(1, 0, 0)
		if next.After(to) {
			break
		}
		years++
		from = next
	}

	// Calculate months
	months := 0
	for {
		next := from.AddDate(0, 1, 0)
		if next.After(to) {
			break
		}
		months++
		from = next
	}

	// Calculate remaining duration
	remaining := to.Sub(from)

	weeks := int(remaining.Hours() / (24 * 7))
	remaining -= time.Duration(weeks) * 7 * 24 * time.Hour

	days := int(remaining.Hours() / 24)
	remaining -= time.Duration(days) * 24 * time.Hour

	hours := int(remaining.Hours())
	remaining -= time.Duration(hours) * time.Hour

	minutes := int(remaining.Minutes())

	// Build result string
	var parts []string
	if years > 0 {
		if years == 1 {
			parts = append(parts, "1 year")
		} else {
			parts = append(parts, fmt.Sprintf("%d years", years))
		}
	}
	if months > 0 {
		if months == 1 {
			parts = append(parts, "1 month")
		} else {
			parts = append(parts, fmt.Sprintf("%d months", months))
		}
	}
	if weeks > 0 {
		if weeks == 1 {
			parts = append(parts, "1 week")
		} else {
			parts = append(parts, fmt.Sprintf("%d weeks", weeks))
		}
	}
	if days > 0 {
		if days == 1 {
			parts = append(parts, "1 day")
		} else {
			parts = append(parts, fmt.Sprintf("%d days", days))
		}
	}
	if hours > 0 {
		if hours == 1 {
			parts = append(parts, "1 hour")
		} else {
			parts = append(parts, fmt.Sprintf("%d hours", hours))
		}
	}
	if minutes > 0 {
		if minutes == 1 {
			parts = append(parts, "1 min")
		} else {
			parts = append(parts, fmt.Sprintf("%d min", minutes))
		}
	}

	if len(parts) == 0 {
		return "0 min"
	}

	return strings.Join(parts, " ")
}

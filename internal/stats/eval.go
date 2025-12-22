package stats

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"smartcalc/internal/utils"
)

// Handler defines the interface for statistics handlers.
type Handler interface {
	Handle(expr, exprLower string) (string, bool)
}

// HandlerFunc is an adapter to allow ordinary functions to be used as Handlers.
type HandlerFunc func(expr, exprLower string) (string, bool)

// Handle calls the underlying function.
func (f HandlerFunc) Handle(expr, exprLower string) (string, bool) {
	return f(expr, exprLower)
}

// handlerChain is the ordered list of handlers for statistics.
var handlerChain = []Handler{
	HandlerFunc(handleAverage),
	HandlerFunc(handleMedian),
	HandlerFunc(handleSum),
	HandlerFunc(handleMin),
	HandlerFunc(handleMax),
	HandlerFunc(handleStdDev),
	HandlerFunc(handleVariance),
	HandlerFunc(handleCount),
	HandlerFunc(handleRange),
}

// EvalStats evaluates a statistics expression and returns the result.
func EvalStats(expr string) (string, error) {
	expr = strings.TrimSpace(expr)
	exprLower := strings.ToLower(expr)

	for _, h := range handlerChain {
		if result, ok := h.Handle(expr, exprLower); ok {
			return result, nil
		}
	}

	return "", fmt.Errorf("unable to evaluate statistics expression: %s", expr)
}

// IsStatsExpression checks if an expression looks like a statistics calculation.
func IsStatsExpression(expr string) bool {
	exprLower := strings.ToLower(expr)

	statsFunctions := []string{
		"avg(", "average(", "mean(",
		"median(",
		"sum(",
		"min(", "max(",
		"stddev(", "stdev(",
		"variance(", "var(",
		"count(",
		"range(",
	}

	for _, fn := range statsFunctions {
		if strings.Contains(exprLower, fn) {
			return true
		}
	}

	return false
}

func parseNumbers(expr string) ([]float64, bool) {
	// Extract content between parentheses
	re := regexp.MustCompile(`\((.*)\)`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return nil, false
	}

	content := matches[1]
	// Split by comma or space
	parts := regexp.MustCompile(`[,\s]+`).Split(content, -1)

	var numbers []float64
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		val, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return nil, false
		}
		numbers = append(numbers, val)
	}

	if len(numbers) == 0 {
		return nil, false
	}

	return numbers, true
}

func handleAverage(expr, exprLower string) (string, bool) {
	// Pattern: avg(1, 2, 3) or average(1, 2, 3) or mean(1, 2, 3)
	if !strings.HasPrefix(exprLower, "avg(") &&
		!strings.HasPrefix(exprLower, "average(") &&
		!strings.HasPrefix(exprLower, "mean(") {
		return "", false
	}

	numbers, ok := parseNumbers(expr)
	if !ok {
		return "", false
	}

	sum := 0.0
	for _, n := range numbers {
		sum += n
	}
	avg := sum / float64(len(numbers))

	return formatResult(avg), true
}

func handleMedian(expr, exprLower string) (string, bool) {
	if !strings.HasPrefix(exprLower, "median(") {
		return "", false
	}

	numbers, ok := parseNumbers(expr)
	if !ok {
		return "", false
	}

	sort.Float64s(numbers)
	n := len(numbers)

	var median float64
	if n%2 == 0 {
		median = (numbers[n/2-1] + numbers[n/2]) / 2
	} else {
		median = numbers[n/2]
	}

	return formatResult(median), true
}

func handleSum(expr, exprLower string) (string, bool) {
	if !strings.HasPrefix(exprLower, "sum(") {
		return "", false
	}

	numbers, ok := parseNumbers(expr)
	if !ok {
		return "", false
	}

	sum := 0.0
	for _, n := range numbers {
		sum += n
	}

	return formatResult(sum), true
}

func handleMin(expr, exprLower string) (string, bool) {
	if !strings.HasPrefix(exprLower, "min(") {
		return "", false
	}

	numbers, ok := parseNumbers(expr)
	if !ok {
		return "", false
	}

	min := numbers[0]
	for _, n := range numbers[1:] {
		if n < min {
			min = n
		}
	}

	return formatResult(min), true
}

func handleMax(expr, exprLower string) (string, bool) {
	if !strings.HasPrefix(exprLower, "max(") {
		return "", false
	}

	numbers, ok := parseNumbers(expr)
	if !ok {
		return "", false
	}

	max := numbers[0]
	for _, n := range numbers[1:] {
		if n > max {
			max = n
		}
	}

	return formatResult(max), true
}

func handleStdDev(expr, exprLower string) (string, bool) {
	if !strings.HasPrefix(exprLower, "stddev(") && !strings.HasPrefix(exprLower, "stdev(") {
		return "", false
	}

	numbers, ok := parseNumbers(expr)
	if !ok || len(numbers) < 2 {
		return "", false
	}

	// Calculate mean
	sum := 0.0
	for _, n := range numbers {
		sum += n
	}
	mean := sum / float64(len(numbers))

	// Calculate variance
	variance := 0.0
	for _, n := range numbers {
		variance += (n - mean) * (n - mean)
	}
	variance /= float64(len(numbers))

	stddev := math.Sqrt(variance)
	return formatResult(stddev), true
}

func handleVariance(expr, exprLower string) (string, bool) {
	if !strings.HasPrefix(exprLower, "variance(") && !strings.HasPrefix(exprLower, "var(") {
		return "", false
	}

	numbers, ok := parseNumbers(expr)
	if !ok || len(numbers) < 2 {
		return "", false
	}

	// Calculate mean
	sum := 0.0
	for _, n := range numbers {
		sum += n
	}
	mean := sum / float64(len(numbers))

	// Calculate variance
	variance := 0.0
	for _, n := range numbers {
		variance += (n - mean) * (n - mean)
	}
	variance /= float64(len(numbers))

	return formatResult(variance), true
}

func handleCount(expr, exprLower string) (string, bool) {
	if !strings.HasPrefix(exprLower, "count(") {
		return "", false
	}

	numbers, ok := parseNumbers(expr)
	if !ok {
		return "", false
	}

	return fmt.Sprintf("%d", len(numbers)), true
}

func handleRange(expr, exprLower string) (string, bool) {
	if !strings.HasPrefix(exprLower, "range(") {
		return "", false
	}

	numbers, ok := parseNumbers(expr)
	if !ok {
		return "", false
	}

	min := numbers[0]
	max := numbers[0]
	for _, n := range numbers[1:] {
		if n < min {
			min = n
		}
		if n > max {
			max = n
		}
	}

	return formatResult(max - min), true
}

func formatResult(value float64) string {
	return utils.FormatResult(false, value)
}

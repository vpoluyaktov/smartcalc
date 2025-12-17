package calc

import (
	"fmt"
	"strings"

	"supercalc/internal/eval"
	"supercalc/internal/utils"
)

// LineResult holds the result of evaluating a single line.
type LineResult struct {
	Output     string
	Value      float64
	HasResult  bool
	IsCurrency bool
}

// EvalLines evaluates all lines and returns the processed output lines.
func EvalLines(lines []string) []LineResult {
	results := make([]LineResult, len(lines))
	values := make([]float64, len(lines))
	haveRes := make([]bool, len(lines))
	currencyByLine := make([]bool, len(lines))

	for i, line := range lines {
		results[i].Output = line
		trim := strings.TrimSpace(line)
		if trim == "" {
			continue
		}
		eq := strings.IndexRune(line, '=')
		if eq < 0 {
			continue
		}
		expr := strings.TrimSpace(line[:eq])
		if expr == "" {
			continue
		}

		isCurrency := strings.Contains(expr, "$") || eval.ExprReferencesCurrency(expr, currencyByLine)

		val, err := eval.EvalExpr(expr, func(n int) (float64, error) {
			idx := n - 1
			if idx < 0 || idx >= len(values) {
				return 0, fmt.Errorf("bad reference \\\\%d", n)
			}
			if !haveRes[idx] {
				return 0, fmt.Errorf("unresolved reference \\\\%d", n)
			}
			return values[idx], nil
		})
		if err != nil {
			results[i].Output = strings.TrimRight(line[:eq+1], " ") + " ERR"
			continue
		}

		values[i] = val
		haveRes[i] = true
		currencyByLine[i] = isCurrency

		results[i].Output = strings.TrimRight(line[:eq+1], " ") + " " + utils.FormatResult(isCurrency, val)
		results[i].Value = val
		results[i].HasResult = true
		results[i].IsCurrency = isCurrency
	}

	return results
}

// BuildLineNumbers generates line number text for n lines.
func BuildLineNumbers(n int) string {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		b.WriteString(fmt.Sprintf("%d\n", i))
	}
	return strings.TrimRight(b.String(), "\n")
}

// GetLineValues returns a map of line number (1-based) to formatted result string.
// This is used for replacing references with actual values when copying.
func GetLineValues(lines []string) map[int]string {
	results := EvalLines(lines)
	values := make(map[int]string)
	for i, r := range results {
		if r.HasResult {
			values[i+1] = utils.FormatResult(r.IsCurrency, r.Value)
		}
	}
	return values
}

// ReplaceRefsWithValues takes text and replaces all \n references with actual values.
func ReplaceRefsWithValues(text string) string {
	lines := strings.Split(text, "\n")
	values := GetLineValues(lines)
	return eval.ReplaceReferencesWithValues(text, values)
}

package calc

import (
	"fmt"
	"regexp"
	"strings"

	"smartcalc/internal/datetime"
	"smartcalc/internal/eval"
	"smartcalc/internal/network"
	"smartcalc/internal/utils"
)

// formatExpression adds proper spacing around operators in an expression.
// Transforms "2+3" to "2 + 3", "$100-20%" to "$100 - 20%", etc.
func formatExpression(expr string) string {
	// Operators that should have spaces around them
	// Handle multi-char operators first, then single-char
	result := expr

	// Add space around basic operators: +, -, *, /, x, ×, ÷, ^
	// But be careful not to affect:
	// - Negative numbers at start or after operator
	// - Currency symbols like $
	// - Decimal points
	// - CIDR notation like /24
	// - Time notation like 6:00

	// First, normalize multiple spaces to single space
	spaceRe := regexp.MustCompile(`\s+`)
	result = spaceRe.ReplaceAllString(result, " ")

	// Add spaces around operators (but not inside numbers or special notations)
	// Match operator not preceded/followed by space
	operators := []struct {
		pattern string
		replace string
	}{
		// Multiplication variants
		{`(\S)\s*×\s*(\S)`, `$1 × $2`},
		{`(\S)\s*÷\s*(\S)`, `$1 ÷ $2`},
		{`(\d)\s*x\s*(\d)`, `$1 x $2`},  // x between digits (multiplication)
		{`(\d)\s*\*\s*(\d)`, `$1 * $2`}, // * between digits
		{`(\d)\s*\^\s*(\d)`, `$1 ^ $2`}, // ^ between digits
		// Addition - digit/paren/percent followed by +
		{`([\d\)%])\s*\+\s*(\S)`, `$1 + $2`},
		// Subtraction - digit/paren/percent followed by -
		{`([\d\)%])\s*-\s*(\S)`, `$1 - $2`},
		// Division - but not CIDR notation (/24)
		{`(\d)\s*/\s*(\d{3,})`, `$1 / $2`}, // Only if divisor is 3+ digits (not CIDR)
	}

	for _, op := range operators {
		re := regexp.MustCompile(op.pattern)
		result = re.ReplaceAllString(result, op.replace)
	}

	return result
}

// LineResult holds the result of evaluating a single line.
type LineResult struct {
	Output      string
	Value       float64
	HasResult   bool
	IsCurrency  bool
	IsDateTime  bool
	DateTimeStr string // raw datetime result for reference
}

// cleanOutputLines removes stale output lines ("> " prefixed) that follow expression lines.
// This ensures old multi-line output is cleared before new evaluation.
func cleanOutputLines(lines []string) []string {
	var result []string
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		// Skip output lines - they will be regenerated
		if strings.HasPrefix(trim, "> ") {
			continue
		}
		result = append(result, line)
	}
	return result
}

// EvalLines evaluates all lines and returns the processed output lines.
func EvalLines(lines []string) []LineResult {
	// First pass: remove stale output lines ("> " lines that follow an expression)
	cleanedLines := cleanOutputLines(lines)

	results := make([]LineResult, len(cleanedLines))
	values := make([]float64, len(cleanedLines))
	haveRes := make([]bool, len(cleanedLines))
	currencyByLine := make([]bool, len(cleanedLines))

	for i, line := range cleanedLines {
		results[i].Output = line
		trim := strings.TrimSpace(line)
		if trim == "" {
			continue
		}
		// Skip output lines (prefixed with "> ")
		if strings.HasPrefix(trim, "> ") {
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

		// Try network/IP evaluation first
		if network.IsNetworkExpression(expr) {
			netResult, err := network.EvalNetwork(expr)
			if err == nil {
				results[i].Output = formatExpression(expr) + " = " + netResult
				results[i].HasResult = true
				continue
			}
			// Fall through if network eval fails
		}

		// Try date/time evaluation (with reference support)
		if datetime.IsDateTimeExpression(expr) || strings.Contains(expr, "\\") {
			// Create resolver for line references
			resolver := func(n int) (string, bool) {
				idx := n - 1
				if idx < 0 || idx >= len(results) {
					return "", false
				}
				if results[idx].IsDateTime && results[idx].DateTimeStr != "" {
					return results[idx].DateTimeStr, true
				}
				return "", false
			}

			dtResult, err := datetime.EvalDateTimeWithRefs(expr, resolver)
			if err == nil {
				results[i].Output = formatExpression(expr) + " = " + dtResult
				results[i].HasResult = true
				results[i].IsDateTime = true
				results[i].DateTimeStr = dtResult
				continue
			}
			// Fall through to numeric evaluation if datetime fails
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
			results[i].Output = formatExpression(expr) + " = ERR"
			continue
		}

		values[i] = val
		haveRes[i] = true
		currencyByLine[i] = isCurrency

		results[i].Output = formatExpression(expr) + " = " + utils.FormatResult(isCurrency, val)
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

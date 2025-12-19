package calc

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"smartcalc/internal/constants"
	"smartcalc/internal/datetime"
	"smartcalc/internal/eval"
	"smartcalc/internal/finance"
	"smartcalc/internal/network"
	"smartcalc/internal/percentage"
	"smartcalc/internal/programmer"
	"smartcalc/internal/stats"
	"smartcalc/internal/units"
	"smartcalc/internal/utils"
)

// tryBaseConversion handles expressions like "24 in dec", "25 in hex", "25 in oct", "25 in bin"
// Also handles hex input like "0xFF in dec" or "0b1010 in dec"
func tryBaseConversion(expr string) (string, bool) {
	exprLower := strings.ToLower(strings.TrimSpace(expr))

	// Pattern: "number in base"
	re := regexp.MustCompile(`(?i)^(0x[0-9a-fA-F]+|0b[01]+|0o[0-7]+|\d+)\s+in\s+(dec|decimal|hex|hexadecimal|oct|octal|bin|binary)$`)
	matches := re.FindStringSubmatch(exprLower)
	if matches == nil {
		return "", false
	}

	numStr := matches[1]
	targetBase := matches[2]

	// Parse the input number
	var value int64
	var err error

	if strings.HasPrefix(numStr, "0x") {
		// Hexadecimal input
		value, err = strconv.ParseInt(numStr[2:], 16, 64)
	} else if strings.HasPrefix(numStr, "0b") {
		// Binary input
		value, err = strconv.ParseInt(numStr[2:], 2, 64)
	} else if strings.HasPrefix(numStr, "0o") {
		// Octal input
		value, err = strconv.ParseInt(numStr[2:], 8, 64)
	} else {
		// Decimal input
		value, err = strconv.ParseInt(numStr, 10, 64)
	}

	if err != nil {
		return "", false
	}

	// Convert to target base
	var result string
	switch {
	case strings.HasPrefix(targetBase, "dec"):
		result = strconv.FormatInt(value, 10)
	case strings.HasPrefix(targetBase, "hex"):
		result = "0x" + strings.ToUpper(strconv.FormatInt(value, 16))
	case strings.HasPrefix(targetBase, "oct"):
		result = "0o" + strconv.FormatInt(value, 8)
	case strings.HasPrefix(targetBase, "bin"):
		result = "0b" + strconv.FormatInt(value, 2)
	default:
		return "", false
	}

	return result, true
}

// isBaseConversionExpr checks if expression is a base conversion
func isBaseConversionExpr(expr string) bool {
	exprLower := strings.ToLower(expr)
	return strings.Contains(exprLower, " in dec") ||
		strings.Contains(exprLower, " in hex") ||
		strings.Contains(exprLower, " in oct") ||
		strings.Contains(exprLower, " in bin") ||
		strings.Contains(exprLower, " in decimal") ||
		strings.Contains(exprLower, " in hexadecimal") ||
		strings.Contains(exprLower, " in octal") ||
		strings.Contains(exprLower, " in binary")
}

// findResultEquals finds the position of the trailing '=' that marks the result,
// skipping '=' characters that are part of comparison operators (>=, <=, ==, !=).
// Returns -1 if no result '=' is found.
func findResultEquals(s string) int {
	// Find the last '=' that is not part of a comparison operator
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '=' {
			// Check if this '=' is part of >=, <=, ==, or !=
			if i > 0 {
				prev := s[i-1]
				if prev == '>' || prev == '<' || prev == '=' || prev == '!' {
					continue // Skip this '=', it's part of a comparison operator
				}
			}
			return i
		}
	}
	return -1
}

// extractInlineComment extracts an inline comment from a line.
// Returns the comment string (including the # prefix) if found, empty string otherwise.
// The comment must appear after the result '=' to be preserved.
func extractInlineComment(line string, eqPos int) string {
	// Look for # after the = sign
	afterEq := line[eqPos+1:]
	hashIdx := strings.Index(afterEq, "#")
	if hashIdx >= 0 {
		return " " + strings.TrimSpace(afterEq[hashIdx:])
	}
	return ""
}

// isComparisonExpr checks if an expression contains comparison operators
func isComparisonExpr(expr string) bool {
	// Check for comparison operators: >, <, >=, <=, ==, !=
	if strings.Contains(expr, ">=") || strings.Contains(expr, "<=") ||
		strings.Contains(expr, "==") || strings.Contains(expr, "!=") {
		return true
	}
	// Check for single > or < (but not part of >= or <=)
	for i, r := range expr {
		if r == '>' || r == '<' {
			// Make sure it's not followed by =
			if i+1 < len(expr) && expr[i+1] == '=' {
				continue
			}
			return true
		}
	}
	return false
}

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
		// Skip comment lines (starting with #)
		if strings.HasPrefix(trim, "#") {
			continue
		}

		// Handle inline comments - strip everything after #
		workingLine := line
		inlineComment := ""
		if hashIdx := strings.Index(line, "#"); hashIdx >= 0 {
			workingLine = line[:hashIdx]
		}

		eq := findResultEquals(workingLine)
		if eq < 0 {
			continue
		}
		expr := strings.TrimSpace(workingLine[:eq])
		if expr == "" {
			continue
		}

		// Extract inline comment from original line (after the = sign)
		inlineComment = extractInlineComment(line, eq)

		// Try base conversion first (24 in hex, 0xFF in dec, etc.)
		if isBaseConversionExpr(expr) {
			if baseResult, ok := tryBaseConversion(expr); ok {
				results[i].Output = expr + " = " + baseResult + inlineComment
				results[i].HasResult = true
				continue
			}
		}

		// Try physical constants
		if constants.IsConstantExpression(expr) {
			constResult, err := constants.EvalConstants(expr)
			if err == nil {
				results[i].Output = formatExpression(expr) + " = " + constResult + inlineComment
				results[i].HasResult = true
				continue
			}
		}

		// Try unit conversions
		if units.IsUnitExpression(expr) {
			unitResult, err := units.EvalUnits(expr)
			if err == nil {
				results[i].Output = formatExpression(expr) + " = " + unitResult + inlineComment
				results[i].HasResult = true
				continue
			}
		}

		// Try percentage calculations
		if percentage.IsPercentageExpression(expr) {
			pctResult, err := percentage.EvalPercentage(expr)
			if err == nil {
				results[i].Output = formatExpression(expr) + " = " + pctResult + inlineComment
				results[i].HasResult = true
				continue
			}
		}

		// Try financial calculations
		if finance.IsFinanceExpression(expr) {
			finResult, err := finance.EvalFinance(expr)
			if err == nil {
				results[i].Output = formatExpression(expr) + " = " + finResult + inlineComment
				results[i].HasResult = true
				continue
			}
		}

		// Try statistics functions
		if stats.IsStatsExpression(expr) {
			statsResult, err := stats.EvalStats(expr)
			if err == nil {
				results[i].Output = formatExpression(expr) + " = " + statsResult + inlineComment
				results[i].HasResult = true
				continue
			}
		}

		// Try programmer utilities
		if programmer.IsProgrammerExpression(expr) {
			progResult, err := programmer.EvalProgrammer(expr)
			if err == nil {
				results[i].Output = formatExpression(expr) + " = " + progResult + inlineComment
				results[i].HasResult = true
				continue
			}
		}

		// Try network/IP evaluation
		if network.IsNetworkExpression(expr) {
			netResult, err := network.EvalNetwork(expr)
			if err == nil {
				results[i].Output = formatExpression(expr) + " = " + netResult + inlineComment
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
				results[i].Output = formatExpression(expr) + " = " + dtResult + inlineComment
				results[i].HasResult = true
				results[i].IsDateTime = true
				results[i].DateTimeStr = dtResult
				continue
			}
			// Fall through to numeric evaluation if datetime fails
		}

		isCurrency := strings.Contains(expr, "$") || eval.ExprReferencesCurrency(expr, currencyByLine)
		isComparison := isComparisonExpr(expr)

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
			results[i].Output = formatExpression(expr) + " = ERR" + inlineComment
			continue
		}

		values[i] = val
		haveRes[i] = true
		currencyByLine[i] = isCurrency

		var resultStr string
		if isComparison {
			resultStr = utils.FormatBoolResult(val)
		} else {
			resultStr = utils.FormatResult(isCurrency, val)
		}
		results[i].Output = formatExpression(expr) + " = " + resultStr + inlineComment
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

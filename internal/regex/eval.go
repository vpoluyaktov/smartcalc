package regex

import (
	"fmt"
	"regexp"
	"strings"
)

// MatchResult represents a single match with its position and groups
type MatchResult struct {
	Start  int      // Start position in the test string
	End    int      // End position in the test string
	Match  string   // The full match
	Groups []string // Captured groups (index 0 is the full match)
}

// RegexResult represents the full result of a regex test
type RegexResult struct {
	Matches     bool          // Whether the regex matches
	MatchCount  int           // Number of matches found
	Results     []MatchResult // All match results
	Error       string        // Error message if regex is invalid
	Highlighted string        // Test string with matches highlighted using markers
}

// IsRegexExpression checks if an expression looks like a regex test
func IsRegexExpression(expr string) bool {
	exprLower := strings.ToLower(strings.TrimSpace(expr))

	// Pattern: regex /pattern/ test "string"
	// Pattern: regex /pattern/ "string"
	// Pattern: /pattern/ test "string"
	// Pattern: /pattern/ match "string"
	patterns := []string{
		`^regex\s+/.+/[gimsuvy]*\s+(?:test|match|against|on)\s+`,
		`^regex\s+/.+/[gimsuvy]*\s+"`,
		`^regex\s+/.+/[gimsuvy]*\s+'`,
		`^regex\s+/.+/[gimsuvy]*\s+` + "`",
		`^/.+/[gimsuvy]*\s+(?:test|match|against|on)\s+`,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, exprLower); matched {
			return true
		}
	}

	return false
}

// EvalRegex evaluates a regex expression and returns the result
func EvalRegex(expr string) (string, error) {
	result := TestRegex(expr)

	if result.Error != "" {
		return "", fmt.Errorf("%s", result.Error)
	}

	return FormatResult(result), nil
}

// TestRegex parses and tests a regex expression
func TestRegex(expr string) RegexResult {
	expr = strings.TrimSpace(expr)

	// Parse the expression to extract pattern and test string
	pattern, testStr, err := parseRegexExpression(expr)
	if err != nil {
		return RegexResult{Error: err.Error()}
	}

	// Compile the regex
	re, err := regexp.Compile(pattern)
	if err != nil {
		return RegexResult{Error: fmt.Sprintf("invalid regex: %s", err.Error())}
	}

	// Find all matches
	allMatches := re.FindAllStringSubmatchIndex(testStr, -1)

	if len(allMatches) == 0 {
		return RegexResult{
			Matches:     false,
			MatchCount:  0,
			Results:     nil,
			Highlighted: testStr,
		}
	}

	// Build match results
	results := make([]MatchResult, 0, len(allMatches))
	for _, match := range allMatches {
		if len(match) >= 2 {
			mr := MatchResult{
				Start:  match[0],
				End:    match[1],
				Match:  testStr[match[0]:match[1]],
				Groups: make([]string, 0),
			}

			// Extract captured groups
			for i := 0; i < len(match); i += 2 {
				if match[i] >= 0 && match[i+1] >= 0 {
					mr.Groups = append(mr.Groups, testStr[match[i]:match[i+1]])
				} else {
					mr.Groups = append(mr.Groups, "") // Empty group
				}
			}

			results = append(results, mr)
		}
	}

	// Build highlighted string with markers
	highlighted := buildHighlightedString(testStr, results)

	return RegexResult{
		Matches:     true,
		MatchCount:  len(results),
		Results:     results,
		Highlighted: highlighted,
	}
}

// parseRegexExpression extracts the pattern and test string from the expression
func parseRegexExpression(expr string) (pattern string, testStr string, err error) {
	// Remove "regex " prefix if present
	exprLower := strings.ToLower(expr)
	if strings.HasPrefix(exprLower, "regex ") {
		expr = strings.TrimSpace(expr[6:])
	}

	// Find the regex pattern between / /
	if !strings.HasPrefix(expr, "/") {
		return "", "", fmt.Errorf("regex pattern must start with /")
	}

	// Find the closing / (accounting for escaped slashes)
	patternEnd := -1
	for i := 1; i < len(expr); i++ {
		if expr[i] == '/' && (i == 1 || expr[i-1] != '\\') {
			patternEnd = i
			break
		}
	}

	if patternEnd == -1 {
		return "", "", fmt.Errorf("regex pattern must end with /")
	}

	pattern = expr[1:patternEnd]
	remaining := strings.TrimSpace(expr[patternEnd+1:])

	// Skip optional flags (we don't use them in Go, but allow them for compatibility)
	flagsEnd := 0
	for flagsEnd < len(remaining) {
		c := remaining[flagsEnd]
		if c == 'g' || c == 'i' || c == 'm' || c == 's' || c == 'u' || c == 'v' || c == 'y' {
			flagsEnd++
		} else {
			break
		}
	}
	remaining = strings.TrimSpace(remaining[flagsEnd:])

	// Skip optional "test", "match", "against", "on" keywords
	remainingLower := strings.ToLower(remaining)
	for _, keyword := range []string{"test ", "match ", "against ", "on "} {
		if strings.HasPrefix(remainingLower, keyword) {
			remaining = strings.TrimSpace(remaining[len(keyword):])
			remainingLower = strings.ToLower(remaining)
			break
		}
	}

	// Extract the test string (can be quoted with ", ', or `)
	testStr, err = extractQuotedString(remaining)
	if err != nil {
		// If not quoted, use the rest as the test string
		testStr = remaining
	}

	if testStr == "" {
		return "", "", fmt.Errorf("test string is required")
	}

	return pattern, testStr, nil
}

// extractQuotedString extracts a string from quotes
func extractQuotedString(s string) (string, error) {
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return "", fmt.Errorf("string too short")
	}

	quote := s[0]
	if quote != '"' && quote != '\'' && quote != '`' {
		return "", fmt.Errorf("not a quoted string")
	}

	// Find the closing quote
	for i := 1; i < len(s); i++ {
		if s[i] == quote && (i == 1 || s[i-1] != '\\') {
			return s[1:i], nil
		}
	}

	return "", fmt.Errorf("unclosed quote")
}

// buildHighlightedString creates a string with match markers
// Uses «» markers around matches for frontend highlighting
func buildHighlightedString(testStr string, results []MatchResult) string {
	if len(results) == 0 {
		return testStr
	}

	var sb strings.Builder
	lastEnd := 0

	for _, r := range results {
		// Add text before this match
		if r.Start > lastEnd {
			sb.WriteString(testStr[lastEnd:r.Start])
		}
		// Add the match with markers
		sb.WriteString("«")
		sb.WriteString(r.Match)
		sb.WriteString("»")
		lastEnd = r.End
	}

	// Add remaining text after last match
	if lastEnd < len(testStr) {
		sb.WriteString(testStr[lastEnd:])
	}

	return sb.String()
}

// FormatResult formats the regex result for display
func FormatResult(result RegexResult) string {
	if result.Error != "" {
		return "ERR: " + result.Error
	}

	if !result.Matches {
		return "no match"
	}

	var sb strings.Builder

	// First line: match status with highlighted string
	if result.MatchCount == 1 {
		sb.WriteString(fmt.Sprintf("match: %s", result.Highlighted))
	} else {
		sb.WriteString(fmt.Sprintf("%d matches: %s", result.MatchCount, result.Highlighted))
	}

	// Add captured groups if any (beyond the full match)
	hasGroups := false
	for _, r := range result.Results {
		if len(r.Groups) > 1 {
			hasGroups = true
			break
		}
	}

	if hasGroups {
		for i, r := range result.Results {
			if len(r.Groups) > 1 {
				if result.MatchCount > 1 {
					sb.WriteString(fmt.Sprintf("\n> Match %d groups:", i+1))
				} else {
					sb.WriteString("\n> Groups:")
				}
				for j := 1; j < len(r.Groups); j++ {
					if r.Groups[j] != "" {
						sb.WriteString(fmt.Sprintf("\n>   [%d]: \"%s\"", j, r.Groups[j]))
					} else {
						sb.WriteString(fmt.Sprintf("\n>   [%d]: (empty)", j))
					}
				}
			}
		}
	}

	return sb.String()
}

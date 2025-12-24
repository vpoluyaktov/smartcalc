package regex

import (
	"strings"
	"testing"
)

func TestIsRegexExpression(t *testing.T) {
	tests := []struct {
		expr     string
		expected bool
	}{
		// Valid regex expressions
		{`regex /hello/ test "hello world"`, true},
		{`regex /\d+/ match "abc123def"`, true},
		{`regex /foo/ "foobar"`, true},
		{`regex /bar/g test "bar bar bar"`, true},
		{`/hello/ test "hello world"`, true},
		{`/\d+/ match "123"`, true},
		{`regex /test/i "TEST"`, true},

		// Invalid expressions
		{"hello world", false},
		{"2 + 2", false},
		{"now", false},
		{"regex", false},
		{"regex hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := IsRegexExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("IsRegexExpression(%q) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestTestRegex_BasicMatch(t *testing.T) {
	result := TestRegex(`regex /hello/ test "hello world"`)

	if !result.Matches {
		t.Error("Expected match, got no match")
	}
	if result.MatchCount != 1 {
		t.Errorf("Expected 1 match, got %d", result.MatchCount)
	}
	if len(result.Results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result.Results))
	}
	if result.Results[0].Match != "hello" {
		t.Errorf("Expected match 'hello', got '%s'", result.Results[0].Match)
	}
}

func TestTestRegex_NoMatch(t *testing.T) {
	result := TestRegex(`regex /xyz/ test "hello world"`)

	if result.Matches {
		t.Error("Expected no match, got match")
	}
	if result.MatchCount != 0 {
		t.Errorf("Expected 0 matches, got %d", result.MatchCount)
	}
}

func TestTestRegex_MultipleMatches(t *testing.T) {
	result := TestRegex(`regex /\d+/ test "abc123def456ghi789"`)

	if !result.Matches {
		t.Error("Expected match, got no match")
	}
	if result.MatchCount != 3 {
		t.Errorf("Expected 3 matches, got %d", result.MatchCount)
	}
	if len(result.Results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(result.Results))
	}

	expectedMatches := []string{"123", "456", "789"}
	for i, expected := range expectedMatches {
		if result.Results[i].Match != expected {
			t.Errorf("Match %d: expected '%s', got '%s'", i, expected, result.Results[i].Match)
		}
	}
}

func TestTestRegex_CaptureGroups(t *testing.T) {
	result := TestRegex(`regex /(\w+)@(\w+)\.(\w+)/ test "email: test@example.com"`)

	if !result.Matches {
		t.Error("Expected match, got no match")
	}
	if len(result.Results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result.Results))
	}

	groups := result.Results[0].Groups
	if len(groups) != 4 {
		t.Fatalf("Expected 4 groups (full match + 3 captures), got %d", len(groups))
	}

	expectedGroups := []string{"test@example.com", "test", "example", "com"}
	for i, expected := range expectedGroups {
		if groups[i] != expected {
			t.Errorf("Group %d: expected '%s', got '%s'", i, expected, groups[i])
		}
	}
}

func TestTestRegex_Highlighting(t *testing.T) {
	result := TestRegex(`regex /\d+/ test "abc123def"`)

	if !result.Matches {
		t.Error("Expected match, got no match")
	}

	expected := "abc«123»def"
	if result.Highlighted != expected {
		t.Errorf("Expected highlighted '%s', got '%s'", expected, result.Highlighted)
	}
}

func TestTestRegex_MultipleHighlighting(t *testing.T) {
	result := TestRegex(`regex /\d+/ test "a1b2c3"`)

	if !result.Matches {
		t.Error("Expected match, got no match")
	}

	expected := "a«1»b«2»c«3»"
	if result.Highlighted != expected {
		t.Errorf("Expected highlighted '%s', got '%s'", expected, result.Highlighted)
	}
}

func TestTestRegex_InvalidRegex(t *testing.T) {
	result := TestRegex(`regex /[invalid/ test "hello"`)

	if result.Error == "" {
		t.Error("Expected error for invalid regex, got none")
	}
	if !strings.Contains(result.Error, "invalid regex") {
		t.Errorf("Expected 'invalid regex' in error, got '%s'", result.Error)
	}
}

func TestTestRegex_DifferentQuotes(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{`regex /hello/ test "hello world"`, "hello"},
		{`regex /hello/ test 'hello world'`, "hello"},
		{"regex /hello/ test `hello world`", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := TestRegex(tt.expr)
			if !result.Matches {
				t.Errorf("Expected match for %q", tt.expr)
				return
			}
			if result.Results[0].Match != tt.expected {
				t.Errorf("Expected match '%s', got '%s'", tt.expected, result.Results[0].Match)
			}
		})
	}
}

func TestTestRegex_WithoutRegexPrefix(t *testing.T) {
	result := TestRegex(`/\d+/ test "abc123"`)

	if !result.Matches {
		t.Error("Expected match, got no match")
	}
	if result.Results[0].Match != "123" {
		t.Errorf("Expected match '123', got '%s'", result.Results[0].Match)
	}
}

func TestTestRegex_MatchKeyword(t *testing.T) {
	result := TestRegex(`regex /foo/ match "foobar"`)

	if !result.Matches {
		t.Error("Expected match, got no match")
	}
	if result.Results[0].Match != "foo" {
		t.Errorf("Expected match 'foo', got '%s'", result.Results[0].Match)
	}
}

func TestTestRegex_AgainstKeyword(t *testing.T) {
	result := TestRegex(`regex /bar/ against "foobar"`)

	if !result.Matches {
		t.Error("Expected match, got no match")
	}
	if result.Results[0].Match != "bar" {
		t.Errorf("Expected match 'bar', got '%s'", result.Results[0].Match)
	}
}

func TestEvalRegex(t *testing.T) {
	tests := []struct {
		expr        string
		shouldMatch bool
		contains    string
	}{
		{`regex /hello/ test "hello world"`, true, "match"},
		{`regex /xyz/ test "hello world"`, true, "no match"},
		{`regex /(\w+)@(\w+)/ test "user@domain"`, true, "Groups"},
		{`regex /\d+/ test "a1b2c3"`, true, "3 matches"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalRegex(tt.expr)
			if tt.shouldMatch && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected result to contain '%s', got '%s'", tt.contains, result)
			}
		})
	}
}

func TestFormatResult(t *testing.T) {
	// Test no match
	result := RegexResult{Matches: false}
	formatted := FormatResult(result)
	if formatted != "no match" {
		t.Errorf("Expected 'no match', got '%s'", formatted)
	}

	// Test error
	result = RegexResult{Error: "test error"}
	formatted = FormatResult(result)
	if formatted != "ERR: test error" {
		t.Errorf("Expected 'ERR: test error', got '%s'", formatted)
	}

	// Test single match
	result = RegexResult{
		Matches:     true,
		MatchCount:  1,
		Highlighted: "«hello» world",
		Results: []MatchResult{
			{Start: 0, End: 5, Match: "hello", Groups: []string{"hello"}},
		},
	}
	formatted = FormatResult(result)
	if !strings.HasPrefix(formatted, "match [0-5]:") {
		t.Errorf("Expected to start with 'match [0-5]:', got '%s'", formatted)
	}

	// Test multiple matches
	result = RegexResult{
		Matches:     true,
		MatchCount:  2,
		Highlighted: "«a»b«c»",
		Results: []MatchResult{
			{Match: "a", Groups: []string{"a"}},
			{Match: "c", Groups: []string{"c"}},
		},
	}
	formatted = FormatResult(result)
	if !strings.HasPrefix(formatted, "2 matches:") {
		t.Errorf("Expected to start with '2 matches:', got '%s'", formatted)
	}
}

func TestTestRegex_SpecialCharacters(t *testing.T) {
	// Test regex with special characters
	result := TestRegex(`regex /\$\d+\.\d{2}/ test "Price: $19.99"`)

	if !result.Matches {
		t.Error("Expected match, got no match")
	}
	if result.Results[0].Match != "$19.99" {
		t.Errorf("Expected match '$19.99', got '%s'", result.Results[0].Match)
	}
}

func TestTestRegex_WordBoundary(t *testing.T) {
	result := TestRegex(`regex /\bword\b/ test "a word here"`)

	if !result.Matches {
		t.Error("Expected match, got no match")
	}
	if result.MatchCount != 1 {
		t.Errorf("Expected 1 match, got %d", result.MatchCount)
	}
}

func TestTestRegex_CaseInsensitive(t *testing.T) {
	// Go's regexp doesn't use flags like /i, but we can use (?i) in the pattern
	result := TestRegex(`regex /(?i)hello/ test "HELLO World"`)

	if !result.Matches {
		t.Error("Expected match with case-insensitive flag, got no match")
	}
	if result.Results[0].Match != "HELLO" {
		t.Errorf("Expected match 'HELLO', got '%s'", result.Results[0].Match)
	}
}

func TestTestRegex_NestedGroups(t *testing.T) {
	result := TestRegex(`regex /((a)(b))/ test "ab"`)

	if !result.Matches {
		t.Error("Expected match, got no match")
	}

	groups := result.Results[0].Groups
	// Groups: [0]=full match "ab", [1]="ab", [2]="a", [3]="b"
	if len(groups) < 4 {
		t.Fatalf("Expected at least 4 groups, got %d", len(groups))
	}

	expectedGroups := []string{"ab", "ab", "a", "b"}
	for i, expected := range expectedGroups {
		if groups[i] != expected {
			t.Errorf("Group %d: expected '%s', got '%s'", i, expected, groups[i])
		}
	}
}

func TestTestRegex_NamedCaptureGroups(t *testing.T) {
	// Go uses (?P<name>...) syntax for named groups
	result := TestRegex(`regex /(?P<user>\w+)@(?P<domain>\w+)\.(?P<tld>\w+)/ test "test@example.com"`)

	if !result.Matches {
		t.Error("Expected match, got no match")
	}

	r := result.Results[0]
	// Groups: [0]=full match, [1]=user, [2]=domain, [3]=tld
	if len(r.Groups) != 4 {
		t.Fatalf("Expected 4 groups, got %d", len(r.Groups))
	}

	expectedGroups := []string{"test@example.com", "test", "example", "com"}
	for i, expected := range expectedGroups {
		if r.Groups[i] != expected {
			t.Errorf("Group %d: expected '%s', got '%s'", i, expected, r.Groups[i])
		}
	}

	// Check group names
	expectedNames := []string{"", "user", "domain", "tld"}
	for i, expected := range expectedNames {
		if i < len(r.GroupNames) && r.GroupNames[i] != expected {
			t.Errorf("GroupName %d: expected '%s', got '%s'", i, expected, r.GroupNames[i])
		}
	}
}

func TestEvalRegex_NamedGroupsOutput(t *testing.T) {
	result, err := EvalRegex(`regex /(?P<name>\w+)=(?P<value>\d+)/ test "count=42"`)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should contain named group labels
	if !strings.Contains(result, "name") {
		t.Errorf("Expected result to contain group name 'name', got: %s", result)
	}
	if !strings.Contains(result, "value") {
		t.Errorf("Expected result to contain group name 'value', got: %s", result)
	}
}

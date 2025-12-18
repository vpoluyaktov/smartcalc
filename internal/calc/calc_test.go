package calc

import (
	"math"
	"testing"
)

func TestEvalLinesBasic(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		expected []LineResult
	}{
		{
			name:  "simple addition",
			lines: []string{"2 + 3 ="},
			expected: []LineResult{
				{Output: "2 + 3 = 5", Value: 5, HasResult: true, IsCurrency: false},
			},
		},
		{
			name:  "empty line",
			lines: []string{""},
			expected: []LineResult{
				{Output: "", Value: 0, HasResult: false, IsCurrency: false},
			},
		},
		{
			name:  "no equals sign",
			lines: []string{"2 + 3"},
			expected: []LineResult{
				{Output: "2 + 3", Value: 0, HasResult: false, IsCurrency: false},
			},
		},
		{
			name:  "currency expression",
			lines: []string{"$100 + $50 ="},
			expected: []LineResult{
				{Output: "$100 + $50 = $150.00", Value: 150, HasResult: true, IsCurrency: true},
			},
		},
		{
			name:  "percentage calculation",
			lines: []string{"100 - 20% ="},
			expected: []LineResult{
				{Output: "100 - 20% = 80", Value: 80, HasResult: true, IsCurrency: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := EvalLines(tt.lines)
			if len(results) != len(tt.expected) {
				t.Fatalf("EvalLines() returned %d results, want %d", len(results), len(tt.expected))
			}
			for i, exp := range tt.expected {
				if results[i].Output != exp.Output {
					t.Errorf("result[%d].Output = %q, want %q", i, results[i].Output, exp.Output)
				}
				if results[i].HasResult != exp.HasResult {
					t.Errorf("result[%d].HasResult = %v, want %v", i, results[i].HasResult, exp.HasResult)
				}
				if results[i].IsCurrency != exp.IsCurrency {
					t.Errorf("result[%d].IsCurrency = %v, want %v", i, results[i].IsCurrency, exp.IsCurrency)
				}
				if exp.HasResult && math.Abs(results[i].Value-exp.Value) > 0.01 {
					t.Errorf("result[%d].Value = %v, want %v", i, results[i].Value, exp.Value)
				}
			}
		})
	}
}

func TestEvalLinesWithReferences(t *testing.T) {
	lines := []string{
		"100 =",
		"50 =",
		"\\1 + \\2 =",
	}

	results := EvalLines(lines)

	if len(results) != 3 {
		t.Fatalf("EvalLines() returned %d results, want 3", len(results))
	}

	// Line 1: 100
	if results[0].Value != 100 {
		t.Errorf("result[0].Value = %v, want 100", results[0].Value)
	}

	// Line 2: 50
	if results[1].Value != 50 {
		t.Errorf("result[1].Value = %v, want 50", results[1].Value)
	}

	// Line 3: \1 + \2 = 150
	if results[2].Value != 150 {
		t.Errorf("result[2].Value = %v, want 150", results[2].Value)
	}
}

func TestEvalLinesReferenceError(t *testing.T) {
	lines := []string{
		"\\1 + 5 =", // references itself - should error
	}

	results := EvalLines(lines)

	if !containsERR(results[0].Output) {
		t.Errorf("expected ERR in output for self-reference, got %q", results[0].Output)
	}
}

func TestEvalLinesCurrencyPropagation(t *testing.T) {
	lines := []string{
		"$100 =",
		"\\1 + 50 =", // should inherit currency from line 1
	}

	results := EvalLines(lines)

	if !results[0].IsCurrency {
		t.Error("result[0] should be currency")
	}
	if !results[1].IsCurrency {
		t.Error("result[1] should inherit currency from reference")
	}
}

func TestEvalLinesMultipleExpressions(t *testing.T) {
	lines := []string{
		"10 + 20 * 3 =",
		"$100 - 20% =",
		"$1,500.00 + $250.50 =",
		"100 =",
		"\\1 * 2 =",
		"sin(45) + cos(30) =",
		"$1,000 x 12 - 15% + $500 =",
	}

	results := EvalLines(lines)

	// Verify each line has a result (no ERR)
	for i, r := range results {
		if containsERR(r.Output) {
			t.Errorf("line %d unexpectedly has ERR: %q", i+1, r.Output)
		}
		if !r.HasResult {
			t.Errorf("line %d should have result", i+1)
		}
	}

	// Check specific values
	if math.Abs(results[0].Value-70) > 0.01 {
		t.Errorf("line 1 value = %v, want 70", results[0].Value)
	}
	if math.Abs(results[1].Value-80) > 0.01 {
		t.Errorf("line 2 value = %v, want 80", results[1].Value)
	}
	if math.Abs(results[2].Value-1750.50) > 0.01 {
		t.Errorf("line 3 value = %v, want 1750.50", results[2].Value)
	}
}

func containsERR(s string) bool {
	return len(s) >= 3 && s[len(s)-3:] == "ERR"
}

func TestBuildLineNumbers(t *testing.T) {
	tests := []struct {
		n        int
		expected string
	}{
		{1, "1"},
		{3, "1\n2\n3"},
		{5, "1\n2\n3\n4\n5"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := BuildLineNumbers(tt.n)
			if result != tt.expected {
				t.Errorf("BuildLineNumbers(%d) = %q, want %q", tt.n, result, tt.expected)
			}
		})
	}
}

func TestCleanOutputLines(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "no output lines",
			input:    []string{"10.100.0.0/16 / 4 subnets =", "some other line"},
			expected: []string{"10.100.0.0/16 / 4 subnets =", "some other line"},
		},
		{
			name: "with output lines",
			input: []string{
				"10.100.0.0/16 / 4 subnets =",
				"> 1: 10.100.0.0/18 (16382 hosts)",
				"> 2: 10.100.64.0/18 (16382 hosts)",
				"> 3: 10.100.128.0/18 (16382 hosts)",
				"> 4: 10.100.192.0/18 (16382 hosts)",
			},
			expected: []string{"10.100.0.0/16 / 4 subnets ="},
		},
		{
			name: "mixed content",
			input: []string{
				"10.100.0.0/16 / 4 subnets =",
				"> 1: 10.100.0.0/18 (16382 hosts)",
				"> 2: 10.100.64.0/18 (16382 hosts)",
				"",
				"100 + 50 =",
			},
			expected: []string{"10.100.0.0/16 / 4 subnets =", "", "100 + 50 ="},
		},
		{
			name:     "empty input",
			input:    []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanOutputLines(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("cleanOutputLines() returned %d lines, want %d\ngot: %v\nwant: %v",
					len(result), len(tt.expected), result, tt.expected)
			}
			for i, exp := range tt.expected {
				if result[i] != exp {
					t.Errorf("result[%d] = %q, want %q", i, result[i], exp)
				}
			}
		})
	}
}

func TestEvalLinesSkipsOutputLines(t *testing.T) {
	// Simulate re-evaluation with stale output lines
	lines := []string{
		"10.100.0.0/16 / 2 subnets =",
		"> 1: 10.100.0.0/17 (32766 hosts)",
		"> 2: 10.100.128.0/17 (32766 hosts)",
	}

	results := EvalLines(lines)

	// Should only have 1 result (the expression line) after cleaning
	if len(results) != 1 {
		t.Fatalf("EvalLines() returned %d results, want 1", len(results))
	}

	// The output should contain the subnet results
	if !results[0].HasResult {
		t.Error("result[0] should have result")
	}

	// Output should start with the expression and contain newlines for subnets
	if results[0].Output[:30] != "10.100.0.0/16 / 2 subnets = \n>" {
		t.Errorf("unexpected output format: %q", results[0].Output[:50])
	}
}

func TestEvalLinesSubnetOutputFormat(t *testing.T) {
	lines := []string{"10.100.0.0/24 / 2 subnets ="}

	results := EvalLines(lines)

	if len(results) != 1 {
		t.Fatalf("EvalLines() returned %d results, want 1", len(results))
	}

	output := results[0].Output

	// Should contain newline before first subnet
	expectedStart := "10.100.0.0/24 / 2 subnets = \n> 1:"
	if len(output) < len(expectedStart) || output[:len(expectedStart)] != expectedStart {
		t.Errorf("output should start with expression followed by newline and first subnet\ngot: %q", output)
	}

	// Should contain both subnets
	if !containsSubstring(output, "> 1:") {
		t.Error("output should contain '> 1:'")
	}
	if !containsSubstring(output, "> 2:") {
		t.Error("output should contain '> 2:'")
	}

	// Should contain host counts
	if !containsSubstring(output, "hosts)") {
		t.Error("output should contain host counts")
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

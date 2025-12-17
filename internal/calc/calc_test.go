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

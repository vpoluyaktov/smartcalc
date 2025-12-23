package eval

import (
	"testing"
)

func TestFindInsertionPoint(t *testing.T) {
	tests := []struct {
		name     string
		oldLines []string
		newLines []string
		expected int
	}{
		{
			name:     "insert at beginning",
			oldLines: []string{"line1", "line2"},
			newLines: []string{"new", "line1", "line2"},
			expected: 1,
		},
		{
			name:     "insert in middle",
			oldLines: []string{"line1", "line2", "line3"},
			newLines: []string{"line1", "new", "line2", "line3"},
			expected: 2,
		},
		{
			name:     "insert at end",
			oldLines: []string{"line1", "line2"},
			newLines: []string{"line1", "line2", "new"},
			expected: 3,
		},
		{
			name:     "insert empty line after first",
			oldLines: []string{"line1", "line2"},
			newLines: []string{"line1", "", "line2"},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindInsertionPoint(tt.oldLines, tt.newLines)
			if result != tt.expected {
				t.Errorf("FindInsertionPoint() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestFindDeletionPoint(t *testing.T) {
	tests := []struct {
		name     string
		oldLines []string
		newLines []string
		expected int
	}{
		{
			name:     "delete from beginning",
			oldLines: []string{"line1", "line2", "line3"},
			newLines: []string{"line2", "line3"},
			expected: 1,
		},
		{
			name:     "delete from middle",
			oldLines: []string{"line1", "line2", "line3"},
			newLines: []string{"line1", "line3"},
			expected: 2,
		},
		{
			name:     "delete from end",
			oldLines: []string{"line1", "line2", "line3"},
			newLines: []string{"line1", "line2"},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindDeletionPoint(tt.oldLines, tt.newLines)
			if result != tt.expected {
				t.Errorf("FindDeletionPoint() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestAdjustReferencesForInsert(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		insertAt int
		delta    int
		expected string
	}{
		{
			name:     "shift ref after insert point",
			text:     "\\2 + 5",
			insertAt: 2,
			delta:    1,
			expected: "\\3 + 5",
		},
		{
			name:     "no shift for ref before insert point",
			text:     "\\1 + 5",
			insertAt: 3,
			delta:    1,
			expected: "\\1 + 5",
		},
		{
			name:     "multiple refs",
			text:     "\\1 + \\2 + \\3",
			insertAt: 2,
			delta:    1,
			expected: "\\1 + \\3 + \\4",
		},
		{
			name:     "insert multiple lines",
			text:     "\\5",
			insertAt: 2,
			delta:    3,
			expected: "\\8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AdjustReferencesForInsert(tt.text, tt.insertAt, tt.delta)
			if result != tt.expected {
				t.Errorf("AdjustReferencesForInsert(%q, %d, %d) = %q, want %q",
					tt.text, tt.insertAt, tt.delta, result, tt.expected)
			}
		})
	}
}

func TestAdjustReferencesForDelete(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		deleteAt int
		delta    int
		expected string
	}{
		{
			name:     "shift ref after delete point",
			text:     "\\3 + 5",
			deleteAt: 2,
			delta:    1,
			expected: "\\2 + 5",
		},
		{
			name:     "no shift for ref before delete point",
			text:     "\\1 + 5",
			deleteAt: 3,
			delta:    1,
			expected: "\\1 + 5",
		},
		{
			name:     "ref in deleted range stays",
			text:     "\\2 + 5",
			deleteAt: 2,
			delta:    1,
			expected: "\\2 + 5",
		},
		{
			name:     "multiple refs with delete",
			text:     "\\1 + \\3 + \\4",
			deleteAt: 2,
			delta:    1,
			expected: "\\1 + \\2 + \\3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AdjustReferencesForDelete(tt.text, tt.deleteAt, tt.delta)
			if result != tt.expected {
				t.Errorf("AdjustReferencesForDelete(%q, %d, %d) = %q, want %q",
					tt.text, tt.deleteAt, tt.delta, result, tt.expected)
			}
		})
	}
}

func TestAdjustReferences(t *testing.T) {
	tests := []struct {
		name     string
		oldText  string
		newText  string
		expected string
	}{
		{
			name:     "no change",
			oldText:  "line1\nline2",
			newText:  "line1\nline2",
			expected: "line1\nline2",
		},
		{
			name:     "insert line shifts refs",
			oldText:  "100 =\n\\1 + 5 =",
			newText:  "100 =\n\n\\1 + 5 =",
			expected: "100 =\n\n\\1 + 5 =", // ref to line 1 stays as \1
		},
		{
			name:     "insert after line 1 shifts ref to line 2",
			oldText:  "100 =\n50 =\n\\2 + 5 =",
			newText:  "100 =\n\n50 =\n\\2 + 5 =",
			expected: "100 =\n\n50 =\n\\3 + 5 =",
		},
		{
			name:     "delete line shifts refs down",
			oldText:  "100 =\n50 =\n\\2 + 5 =",
			newText:  "100 =\n\\2 + 5 =",
			expected: "100 =\n\\2 + 5 =", // ref to deleted line stays (will error)
		},
		{
			name:     "delete line shifts refs after deleted range",
			oldText:  "100 =\n50 =\n200 =\n\\3 + 5 =",
			newText:  "100 =\n200 =\n\\3 + 5 =",
			expected: "100 =\n200 =\n\\2 + 5 =", // ref to line 3 becomes line 2
		},
		{
			name:     "insert at beginning shifts all refs",
			oldText:  "100 =\n\\1 * 2 =",
			newText:  "\n100 =\n\\1 * 2 =",
			expected: "\n100 =\n\\2 * 2 =",
		},
		{
			name:     "delete at beginning shifts all refs",
			oldText:  "\n100 =\n\\2 * 2 =",
			newText:  "100 =\n\\2 * 2 =",
			expected: "100 =\n\\1 * 2 =",
		},
		{
			name:     "insert multiple lines",
			oldText:  "100 =\n\\1 * 2 =",
			newText:  "100 =\n\n\n\\1 * 2 =",
			expected: "100 =\n\n\n\\1 * 2 =", // ref to line 1 stays
		},
		{
			name:     "insert between ref and target",
			oldText:  "100 =\n200 =\n\\1 + \\2 =",
			newText:  "100 =\n\n200 =\n\\1 + \\2 =",
			expected: "100 =\n\n200 =\n\\1 + \\3 =",
		},
		{
			name:     "delete multiple lines",
			oldText:  "100 =\n\n\n\\1 * 2 =",
			newText:  "100 =\n\\1 * 2 =",
			expected: "100 =\n\\1 * 2 =", // ref to line 1 stays
		},
		{
			name:     "complex chain insert",
			oldText:  "100 =\n\\1 * 2 =\n\\2 + 10 =",
			newText:  "100 =\n\n\\1 * 2 =\n\\2 + 10 =",
			expected: "100 =\n\n\\1 * 2 =\n\\3 + 10 =",
		},
		{
			name:     "complex chain delete",
			oldText:  "100 =\n\n\\1 * 2 =\n\\3 + 10 =",
			newText:  "100 =\n\\1 * 2 =\n\\3 + 10 =",
			expected: "100 =\n\\1 * 2 =\n\\2 + 10 =",
		},
		{
			name:     "with results - insert",
			oldText:  "100 = 100\n\\1 * 2 = 200",
			newText:  "100 = 100\n\n\\1 * 2 = 200",
			expected: "100 = 100\n\n\\1 * 2 = 200",
		},
		{
			name:     "with results - delete",
			oldText:  "100 = 100\n\n\\1 * 2 = 200",
			newText:  "100 = 100\n\\1 * 2 = 200",
			expected: "100 = 100\n\\1 * 2 = 200",
		},
		{
			name:     "ref on same line as change",
			oldText:  "100 =\n200 =",
			newText:  "100 =\n\\1 + 50 =\n200 =",
			expected: "100 =\n\\1 + 50 =\n200 =",
		},
		{
			name:     "insert empty line before refs - user scenario",
			oldText:  "# line1\n# line2\n# line3\n\n\n2 + 2 = 4\n\n\\6 x 4 = 16\n\n\\8 / 2 = 8\n\n",
			newText:  "# line1\n# line2\n# line3\n\n\n\n2 + 2 = 4\n\n\\6 x 4 = 16\n\n\\8 / 2 = 8\n\n",
			expected: "# line1\n# line2\n# line3\n\n\n\n2 + 2 = 4\n\n\\7 x 4 = 16\n\n\\9 / 2 = 8\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AdjustReferences(tt.oldText, tt.newText)
			if result != tt.expected {
				t.Errorf("AdjustReferences() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestReplaceReferencesWithValues(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		values   map[int]string
		expected string
	}{
		{
			name:     "single reference",
			text:     "\\1 + 5 =",
			values:   map[int]string{1: "100"},
			expected: "100 + 5 =",
		},
		{
			name:     "multiple references",
			text:     "\\1 + \\2 =",
			values:   map[int]string{1: "100", 2: "50"},
			expected: "100 + 50 =",
		},
		{
			name:     "currency values",
			text:     "\\1 * 2 =",
			values:   map[int]string{1: "$1,500.00"},
			expected: "$1,500.00 * 2 =",
		},
		{
			name:     "missing reference",
			text:     "\\3 + 5 =",
			values:   map[int]string{1: "100"},
			expected: "\\3 + 5 =", // keeps original
		},
		{
			name:     "no references",
			text:     "100 + 50 =",
			values:   map[int]string{1: "200"},
			expected: "100 + 50 =",
		},
		{
			name:     "multiline with references",
			text:     "100 =\n\\1 * 2 =",
			values:   map[int]string{1: "100"},
			expected: "100 =\n100 * 2 =",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReplaceReferencesWithValues(tt.text, tt.values)
			if result != tt.expected {
				t.Errorf("ReplaceReferencesWithValues() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExprReferencesCurrency(t *testing.T) {
	tests := []struct {
		name           string
		expr           string
		currencyByLine []bool
		expected       bool
	}{
		{
			name:           "no references",
			expr:           "2 + 3",
			currencyByLine: []bool{true, false},
			expected:       false,
		},
		{
			name:           "ref to currency line",
			expr:           "\\1 + 5",
			currencyByLine: []bool{true, false},
			expected:       true,
		},
		{
			name:           "ref to non-currency line",
			expr:           "\\2 + 5",
			currencyByLine: []bool{true, false},
			expected:       false,
		},
		{
			name:           "multiple refs one currency",
			expr:           "\\1 + \\2",
			currencyByLine: []bool{false, true},
			expected:       true,
		},
		{
			name:           "ref out of bounds",
			expr:           "\\5 + 5",
			currencyByLine: []bool{true, false},
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExprReferencesCurrency(tt.expr, tt.currencyByLine)
			if result != tt.expected {
				t.Errorf("ExprReferencesCurrency(%q) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}

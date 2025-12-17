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

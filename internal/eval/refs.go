package eval

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// FindInsertionPoint compares old and new lines to find where insertion happened.
// Returns the 1-based line number where new lines were inserted.
func FindInsertionPoint(oldLines, newLines []string) int {
	// Find first line that differs
	minLen := len(oldLines)
	if len(newLines) < minLen {
		minLen = len(newLines)
	}

	for i := 0; i < minLen; i++ {
		if oldLines[i] != newLines[i] {
			return i + 1 // 1-based
		}
	}

	// If all compared lines match, insertion is at the end of the shorter
	return minLen + 1
}

// FindDeletionPoint compares old and new lines to find where deletion happened.
// Returns the 1-based line number where lines were deleted.
func FindDeletionPoint(oldLines, newLines []string) int {
	// Find first line that differs
	minLen := len(newLines)
	if len(oldLines) < minLen {
		minLen = len(oldLines)
	}

	for i := 0; i < minLen; i++ {
		if oldLines[i] != newLines[i] {
			return i + 1 // 1-based
		}
	}

	return minLen + 1
}

// AdjustReferencesForInsert updates \n references when lines are inserted.
// insertAt is 1-based line number where insertion happened.
// delta is the number of lines inserted (positive).
func AdjustReferencesForInsert(text string, insertAt, delta int) string {
	re := regexp.MustCompile(`\\(\d+)`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		numStr := match[1:] // strip leading \
		n, _ := strconv.Atoi(numStr)
		// References to lines >= insertAt should shift up
		if n >= insertAt {
			return fmt.Sprintf("\\%d", n+delta)
		}
		return match
	})
}

// AdjustReferencesForDelete updates \n references when lines are deleted.
// deleteAt is 1-based line number where deletion started.
// delta is the number of lines deleted (positive).
func AdjustReferencesForDelete(text string, deleteAt, delta int) string {
	re := regexp.MustCompile(`\\(\d+)`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		numStr := match[1:] // strip leading \
		n, _ := strconv.Atoi(numStr)
		deletedEnd := deleteAt + delta
		// References in deleted range stay as-is (will error)
		if n >= deleteAt && n < deletedEnd {
			return match
		}
		// References after deleted range shift down
		if n >= deletedEnd {
			return fmt.Sprintf("\\%d", n-delta)
		}
		return match
	})
}

// AdjustReferences is a convenience function that detects insert/delete and adjusts.
func AdjustReferences(oldText, newText string) string {
	oldLines := strings.Split(oldText, "\n")
	newLines := strings.Split(newText, "\n")

	delta := len(newLines) - len(oldLines)
	if delta == 0 {
		return newText // no line count change
	}

	if delta > 0 {
		// Lines inserted
		insertAt := FindInsertionPoint(oldLines, newLines)
		return AdjustReferencesForInsert(newText, insertAt, delta)
	}

	// Lines deleted
	deleteAt := FindDeletionPoint(oldLines, newLines)
	return AdjustReferencesForDelete(newText, deleteAt, -delta)
}

// ReplaceReferencesWithValues replaces \n references with actual numeric values.
// values is a map from line number (1-based) to the formatted result string.
func ReplaceReferencesWithValues(text string, values map[int]string) string {
	re := regexp.MustCompile(`\\(\d+)`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		numStr := match[1:] // strip leading \
		n, _ := strconv.Atoi(numStr)
		if val, ok := values[n]; ok {
			return val
		}
		return match // keep original if no value found
	})
}

// ExprReferencesCurrency detects references like \1, \2 in the expression
// and returns true if any referenced line was currency.
func ExprReferencesCurrency(expr string, currencyByLine []bool) bool {
	for i := 0; i < len(expr); i++ {
		if expr[i] != '\\' {
			continue
		}
		j := i + 1
		if j >= len(expr) || expr[j] < '0' || expr[j] > '9' {
			continue
		}
		n := 0
		for j < len(expr) && expr[j] >= '0' && expr[j] <= '9' {
			n = n*10 + int(expr[j]-'0')
			j++
		}
		idx := n - 1
		if idx >= 0 && idx < len(currencyByLine) && currencyByLine[idx] {
			return true
		}
		i = j - 1
	}
	return false
}

package programmer

import (
	"strings"
	"testing"
)

func TestBitwiseAnd(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"0xFF AND 0x0F", "15 (0xF)"},
		{"255 and 15", "15 (0xF)"},
		{"0xF0 and 0x0F", "0 (0x0)"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalProgrammer(tt.expr)
			if err != nil {
				t.Errorf("EvalProgrammer(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("EvalProgrammer(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

func TestBitwiseOr(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"0xF0 OR 0x0F", "255 (0xFF)"},
		{"8 or 4", "12 (0xC)"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalProgrammer(tt.expr)
			if err != nil {
				t.Errorf("EvalProgrammer(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("EvalProgrammer(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

func TestBitwiseXor(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"0xFF XOR 0x0F", "240 (0xF0)"},
		{"10 xor 10", "0 (0x0)"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalProgrammer(tt.expr)
			if err != nil {
				t.Errorf("EvalProgrammer(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("EvalProgrammer(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

func TestLeftShift(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"1 << 8", "256 (0x100)"},
		{"1 << 4", "16 (0x10)"},
		{"0xFF << 4", "4080 (0xFF0)"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalProgrammer(tt.expr)
			if err != nil {
				t.Errorf("EvalProgrammer(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("EvalProgrammer(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

func TestRightShift(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"256 >> 4", "16 (0x10)"},
		{"0xFF >> 4", "15 (0xF)"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalProgrammer(tt.expr)
			if err != nil {
				t.Errorf("EvalProgrammer(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("EvalProgrammer(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

func TestAsciiChar(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"char 65", "'A'"},
		{"char 0x41", "'A'"},
		{"char 97", "'a'"},
		{"ascii A", "65 (0x41)"},
		{"ascii a", "97 (0x61)"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalProgrammer(tt.expr)
			if err != nil {
				t.Errorf("EvalProgrammer(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalProgrammer(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestUUID(t *testing.T) {
	result, err := EvalProgrammer("uuid")
	if err != nil {
		t.Errorf("EvalProgrammer(uuid) error: %v", err)
		return
	}
	// UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	if len(result) != 36 {
		t.Errorf("EvalProgrammer(uuid) = %q, want 36 char UUID", result)
	}
	if result[8] != '-' || result[13] != '-' || result[18] != '-' || result[23] != '-' {
		t.Errorf("EvalProgrammer(uuid) = %q, invalid UUID format", result)
	}
}

func TestMD5(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"md5 hello", "5d41402abc4b2a76b9719d911017c592"},
		{"md5 test", "098f6bcd4621d373cade4e832627b4f6"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalProgrammer(tt.expr)
			if err != nil {
				t.Errorf("EvalProgrammer(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalProgrammer(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestSHA256(t *testing.T) {
	result, err := EvalProgrammer("sha256 hello")
	if err != nil {
		t.Errorf("EvalProgrammer(sha256 hello) error: %v", err)
		return
	}
	expected := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if result != expected {
		t.Errorf("EvalProgrammer(sha256 hello) = %q, want %q", result, expected)
	}
}

func TestAsciiTable(t *testing.T) {
	result, err := EvalProgrammer("ascii table")
	if err != nil {
		t.Errorf("EvalProgrammer(ascii table) error: %v", err)
		return
	}
	// Check that it contains expected content
	if !strings.Contains(result, "Control Characters") {
		t.Errorf("ascii table should contain 'Control Characters'")
	}
	if !strings.Contains(result, "Printable Characters") {
		t.Errorf("ascii table should contain 'Printable Characters'")
	}
	if !strings.Contains(result, "NUL") {
		t.Errorf("ascii table should contain 'NUL'")
	}
	if !strings.Contains(result, "> ") {
		t.Errorf("ascii table output should be prefixed with '> '")
	}
}

func TestIsProgrammerExpression(t *testing.T) {
	tests := []struct {
		expr     string
		expected bool
	}{
		{"0xFF AND 0x0F", true},
		{"1 << 8", true},
		{"ascii A", true},
		{"ascii table", true},
		{"uuid", true},
		{"md5 hello", true},
		{"100 + 50", false},
		{"5 miles in km", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := IsProgrammerExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("IsProgrammerExpression(%q) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}

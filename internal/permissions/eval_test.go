package permissions

import (
	"strings"
	"testing"
)

func TestIsPermissionsExpression(t *testing.T) {
	tests := []struct {
		expr     string
		expected bool
	}{
		// chmod octal
		{"chmod 755", true},
		{"chmod 0755", true},
		{"chmod 4755", true},
		{"chmod 777", true},
		{"chmod 644", true},
		{"CHMOD 755", true},

		// chmod symbolic
		{"chmod rwxr-xr-x", true},
		{"chmod rw-r--r--", true},
		{"chmod rwxrwxrwx", true},

		// chmod symbolic spaced
		{"chmod rwx r-x r-x", true},
		{"chmod rw- r-- r--", true},

		// umask
		{"umask 022", true},
		{"umask 0022", true},
		{"umask 077", true},
		{"UMASK 022", true},

		// conversion expressions
		{"755 to symbolic", true},
		{"755 symbolic", true},
		{"rwxr-xr-x to octal", true},
		{"rwxr-xr-x octal", true},

		// permission keyword
		{"permission 755", true},
		{"permissions 755", true},

		// Not permissions expressions
		{"chmod", false},
		{"755", false},
		{"hello world", false},
		{"2 + 2", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := IsPermissionsExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("IsPermissionsExpression(%q) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestChmodOctalToSymbolic(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"chmod 755", "rwxr-xr-x"},
		{"chmod 644", "rw-r--r--"},
		{"chmod 777", "rwxrwxrwx"},
		{"chmod 000", "---------"},
		{"chmod 700", "rwx------"},
		{"chmod 600", "rw-------"},
		{"chmod 400", "r--------"},
		{"chmod 100", "--x------"},
		{"chmod 070", "---rwx---"},
		{"chmod 007", "------rwx"},
		{"chmod 111", "--x--x--x"},
		{"chmod 222", "-w--w--w-"},
		{"chmod 333", "-wx-wx-wx"},
		{"chmod 444", "r--r--r--"},
		{"chmod 555", "r-xr-xr-x"},
		{"chmod 666", "rw-rw-rw-"},

		// With leading zero
		{"chmod 0755", "rwxr-xr-x"},
		{"chmod 0644", "rw-r--r--"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalPermissions(tt.expr)
			if err != nil {
				t.Errorf("EvalPermissions(%q) returned error: %v", tt.expr, err)
				return
			}
			if !strings.HasPrefix(result, tt.expected) {
				t.Errorf("EvalPermissions(%q) = %q, want prefix %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestChmodSymbolicToOctal(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"chmod rwxr-xr-x", "755"},
		{"chmod rw-r--r--", "644"},
		{"chmod rwxrwxrwx", "777"},
		{"chmod ---------", "000"},
		{"chmod rwx------", "700"},
		{"chmod rw-------", "600"},
		{"chmod r--------", "400"},
		{"chmod --x------", "100"},
		{"chmod ---rwx---", "070"},
		{"chmod ------rwx", "007"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalPermissions(tt.expr)
			if err != nil {
				t.Errorf("EvalPermissions(%q) returned error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalPermissions(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestChmodSymbolicSpaced(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"chmod rwx r-x r-x", "755"},
		{"chmod rw- r-- r--", "644"},
		{"chmod rwx rwx rwx", "777"},
		{"chmod --- --- ---", "000"},
		{"chmod rwx --- ---", "700"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalPermissions(tt.expr)
			if err != nil {
				t.Errorf("EvalPermissions(%q) returned error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalPermissions(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestSpecialBits(t *testing.T) {
	tests := []struct {
		expr            string
		expectedSymbol  string
		expectedSpecial string
	}{
		// Setuid (4xxx)
		{"chmod 4755", "rwsr-xr-x", "setuid"},
		{"chmod 4644", "rwSr--r--", "setuid"},

		// Setgid (2xxx)
		{"chmod 2755", "rwxr-sr-x", "setgid"},
		{"chmod 2644", "rw-r-Sr--", "setgid"},

		// Sticky bit (1xxx)
		{"chmod 1755", "rwxr-xr-t", "sticky"},
		{"chmod 1644", "rw-r--r-T", "sticky"},

		// Combined special bits
		{"chmod 6755", "rwsr-sr-x", "setuid, setgid"},
		{"chmod 7755", "rwsr-sr-t", "setuid, setgid, sticky"},
		{"chmod 3755", "rwxr-sr-t", "setgid, sticky"},
		{"chmod 5755", "rwsr-xr-t", "setuid, sticky"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalPermissions(tt.expr)
			if err != nil {
				t.Errorf("EvalPermissions(%q) returned error: %v", tt.expr, err)
				return
			}
			if !strings.HasPrefix(result, tt.expectedSymbol) {
				t.Errorf("EvalPermissions(%q) = %q, want prefix %q", tt.expr, result, tt.expectedSymbol)
			}
			if !strings.Contains(result, tt.expectedSpecial) {
				t.Errorf("EvalPermissions(%q) = %q, want to contain %q", tt.expr, result, tt.expectedSpecial)
			}
		})
	}
}

func TestSpecialBitsSymbolicToOctal(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"chmod rwsr-xr-x", "4755"},
		{"chmod rwSr--r--", "4644"},
		{"chmod rwxr-sr-x", "2755"},
		{"chmod rw-r-Sr--", "2644"},
		{"chmod rwxr-xr-t", "1755"},
		{"chmod rw-r--r-T", "1644"},
		{"chmod rwsr-sr-x", "6755"},
		{"chmod rwsr-sr-t", "7755"},
	}

	// Note: The symbolic notation uses:
	// - 's' = setuid/setgid WITH execute
	// - 'S' = setuid/setgid WITHOUT execute
	// - 't' = sticky WITH execute
	// - 'T' = sticky WITHOUT execute

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalPermissions(tt.expr)
			if err != nil {
				t.Errorf("EvalPermissions(%q) returned error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalPermissions(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestUmask(t *testing.T) {
	tests := []struct {
		expr         string
		fileMode     string
		dirMode      string
		fileSymbolic string
		dirSymbolic  string
	}{
		{"umask 022", "644", "755", "rw-r--r--", "rwxr-xr-x"},
		{"umask 077", "600", "700", "rw-------", "rwx------"},
		{"umask 000", "666", "777", "rw-rw-rw-", "rwxrwxrwx"},
		{"umask 027", "640", "750", "rw-r-----", "rwxr-x---"},
		{"umask 002", "664", "775", "rw-rw-r--", "rwxrwxr-x"},
		{"umask 007", "660", "770", "rw-rw----", "rwxrwx---"},
		{"umask 0022", "644", "755", "rw-r--r--", "rwxr-xr-x"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalPermissions(tt.expr)
			if err != nil {
				t.Errorf("EvalPermissions(%q) returned error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.fileMode) {
				t.Errorf("EvalPermissions(%q) = %q, want to contain file mode %q", tt.expr, result, tt.fileMode)
			}
			if !strings.Contains(result, tt.dirMode) {
				t.Errorf("EvalPermissions(%q) = %q, want to contain dir mode %q", tt.expr, result, tt.dirMode)
			}
			if !strings.Contains(result, tt.fileSymbolic) {
				t.Errorf("EvalPermissions(%q) = %q, want to contain file symbolic %q", tt.expr, result, tt.fileSymbolic)
			}
			if !strings.Contains(result, tt.dirSymbolic) {
				t.Errorf("EvalPermissions(%q) = %q, want to contain dir symbolic %q", tt.expr, result, tt.dirSymbolic)
			}
		})
	}
}

func TestOctalToSymbolicConversion(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"755 to symbolic", "rwxr-xr-x"},
		{"755 symbolic", "rwxr-xr-x"},
		{"644 to symbolic", "rw-r--r--"},
		{"777 symbolic", "rwxrwxrwx"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalPermissions(tt.expr)
			if err != nil {
				t.Errorf("EvalPermissions(%q) returned error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalPermissions(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestSymbolicToOctalConversion(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"rwxr-xr-x to octal", "755"},
		{"rwxr-xr-x octal", "755"},
		{"rw-r--r-- to octal", "644"},
		{"rwxrwxrwx octal", "777"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalPermissions(tt.expr)
			if err != nil {
				t.Errorf("EvalPermissions(%q) returned error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalPermissions(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestPermissionKeyword(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"permission 755", "rwxr-xr-x"},
		{"permissions 755", "rwxr-xr-x"},
		{"permission 644", "rw-r--r--"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalPermissions(tt.expr)
			if err != nil {
				t.Errorf("EvalPermissions(%q) returned error: %v", tt.expr, err)
				return
			}
			if !strings.HasPrefix(result, tt.expected) {
				t.Errorf("EvalPermissions(%q) = %q, want prefix %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestInvalidExpressions(t *testing.T) {
	tests := []string{
		"chmod 888",      // Invalid octal
		"chmod rwxrwxrw", // Too short
		"chmod abc",      // Invalid
		"umask 888",      // Invalid octal
	}

	for _, expr := range tests {
		t.Run(expr, func(t *testing.T) {
			_, err := EvalPermissions(expr)
			if err == nil {
				t.Errorf("EvalPermissions(%q) should return error", expr)
			}
		})
	}
}

func TestOctalToSymbolicFunction(t *testing.T) {
	tests := []struct {
		mode     int
		expected string
	}{
		{0755, "rwxr-xr-x"},
		{0644, "rw-r--r--"},
		{0777, "rwxrwxrwx"},
		{0000, "---------"},
		{04755, "rwsr-xr-x"},
		{02755, "rwxr-sr-x"},
		{01755, "rwxr-xr-t"},
		{07777, "rwsrwsrwt"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := octalToSymbolic(tt.mode)
			if result != tt.expected {
				t.Errorf("octalToSymbolic(%04o) = %q, want %q", tt.mode, result, tt.expected)
			}
		})
	}
}

func TestSymbolicToOctalFunction(t *testing.T) {
	tests := []struct {
		symbolic string
		expected int
	}{
		{"rwxr-xr-x", 0755},
		{"rw-r--r--", 0644},
		{"rwxrwxrwx", 0777},
		{"---------", 0000},
		{"rwsr-xr-x", 04755},
		{"rwxr-sr-x", 02755},
		{"rwxr-xr-t", 01755},
		{"rwsrwsrwt", 07777},
	}

	for _, tt := range tests {
		t.Run(tt.symbolic, func(t *testing.T) {
			result, err := symbolicToOctal(tt.symbolic)
			if err != nil {
				t.Errorf("symbolicToOctal(%q) returned error: %v", tt.symbolic, err)
				return
			}
			if result != tt.expected {
				t.Errorf("symbolicToOctal(%q) = %04o, want %04o", tt.symbolic, result, tt.expected)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	// Test that converting octal -> symbolic -> octal gives the same result
	modes := []int{0755, 0644, 0777, 0000, 0700, 0600, 04755, 02755, 01755, 07777, 06755, 05755, 03755}

	for _, mode := range modes {
		t.Run(string(rune(mode)), func(t *testing.T) {
			symbolic := octalToSymbolic(mode)
			result, err := symbolicToOctal(symbolic)
			if err != nil {
				t.Errorf("Round trip failed for %04o: %v", mode, err)
				return
			}
			if result != mode {
				t.Errorf("Round trip failed: %04o -> %q -> %04o", mode, symbolic, result)
			}
		})
	}
}

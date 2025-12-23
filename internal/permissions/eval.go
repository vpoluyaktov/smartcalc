package permissions

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Permission bit constants
const (
	// Basic permission bits
	OwnerRead    = 0400
	OwnerWrite   = 0200
	OwnerExecute = 0100
	GroupRead    = 0040
	GroupWrite   = 0020
	GroupExecute = 0010
	OtherRead    = 0004
	OtherWrite   = 0002
	OtherExecute = 0001

	// Special bits
	Setuid = 04000
	Setgid = 02000
	Sticky = 01000

	// Default creation masks
	DefaultFileMode = 0666
	DefaultDirMode  = 0777
)

// IsPermissionsExpression checks if an expression is a Unix permissions expression
func IsPermissionsExpression(expr string) bool {
	exprLower := strings.ToLower(strings.TrimSpace(expr))

	patterns := []string{
		`^chmod\s+[0-7]{3,4}$`,     // chmod 755, chmod 0755, chmod 4755
		`^chmod\s+[rwxst-]{9,10}$`, // chmod rwxr-xr-x
		`^chmod\s+[ugoa]*[rwxst-]+\s+[ugoa]*[rwxst-]+\s+[ugoa]*[rwxst-]+$`, // chmod rwx r-x r-x
		`^umask\s+[0-7]{3,4}$`, // umask 022, umask 0022
		`^[0-7]{3,4}\s+(?:to\s+)?(?:symbolic|sym|permissions?)$`, // 755 to symbolic
		`^[rwxst-]{9,10}\s+(?:to\s+)?(?:octal|numeric|number)$`,  // rwxr-xr-x to octal
		`^permissions?\s+[0-7]{3,4}$`,                            // permission 755
		`^permissions?\s+[rwxst-]{9,10}$`,                        // permission rwxr-xr-x
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, exprLower); matched {
			return true
		}
	}

	return false
}

// EvalPermissions evaluates a Unix permissions expression
func EvalPermissions(expr string) (string, error) {
	expr = strings.TrimSpace(expr)
	exprLower := strings.ToLower(expr)

	// chmod with octal: chmod 755, chmod 0755, chmod 4755
	if matched, _ := regexp.MatchString(`^chmod\s+[0-7]{3,4}$`, exprLower); matched {
		return handleChmodOctal(expr)
	}

	// chmod with symbolic (single string): chmod rwxr-xr-x
	if matched, _ := regexp.MatchString(`^chmod\s+[rwxstST-]{9,10}$`, exprLower); matched {
		return handleChmodSymbolic(expr)
	}

	// chmod with symbolic (spaced): chmod rwx r-x r-x
	if matched, _ := regexp.MatchString(`(?i)^chmod\s+[rwxstST-]+\s+[rwxstST-]+\s+[rwxstST-]+$`, expr); matched {
		return handleChmodSymbolicSpaced(expr)
	}

	// umask: umask 022
	if matched, _ := regexp.MatchString(`^umask\s+[0-7]{3,4}$`, exprLower); matched {
		return handleUmask(expr)
	}

	// octal to symbolic: 755 to symbolic
	if matched, _ := regexp.MatchString(`^[0-7]{3,4}\s+(?:to\s+)?(?:symbolic|sym|permissions?)$`, exprLower); matched {
		return handleOctalToSymbolic(expr)
	}

	// symbolic to octal: rwxr-xr-x to octal
	if matched, _ := regexp.MatchString(`^[rwxst-]{9,10}\s+(?:to\s+)?(?:octal|numeric|number)$`, exprLower); matched {
		return handleSymbolicToOctal(expr)
	}

	// permission octal: permission 755
	if matched, _ := regexp.MatchString(`^permissions?\s+[0-7]{3,4}$`, exprLower); matched {
		return handleChmodOctal(expr)
	}

	// permission symbolic: permission rwxr-xr-x
	if matched, _ := regexp.MatchString(`^permissions?\s+[rwxst-]{9,10}$`, exprLower); matched {
		return handleChmodSymbolic(expr)
	}

	return "", fmt.Errorf("unable to evaluate permissions expression: %s", expr)
}

// handleChmodOctal converts octal permissions to symbolic
// e.g., chmod 755 -> rwxr-xr-x
func handleChmodOctal(expr string) (string, error) {
	re := regexp.MustCompile(`(?i)(?:chmod|permissions?)\s+([0-7]{3,4})`)
	matches := re.FindStringSubmatch(expr)
	if len(matches) < 2 {
		return "", fmt.Errorf("invalid chmod expression: %s", expr)
	}

	octalStr := matches[1]
	mode, err := strconv.ParseInt(octalStr, 8, 64)
	if err != nil {
		return "", fmt.Errorf("invalid octal number: %s", octalStr)
	}

	symbolic := octalToSymbolic(int(mode))
	specialBits := getSpecialBitsDescription(int(mode))

	result := symbolic
	if specialBits != "" {
		result += " (" + specialBits + ")"
	}

	return result, nil
}

// handleChmodSymbolic converts symbolic permissions to octal
// e.g., chmod rwxr-xr-x -> 755
func handleChmodSymbolic(expr string) (string, error) {
	re := regexp.MustCompile(`(?i)(?:chmod|permissions?)\s+([rwxstST-]{9,10})`)
	matches := re.FindStringSubmatch(expr)
	if len(matches) < 2 {
		return "", fmt.Errorf("invalid chmod expression: %s", expr)
	}

	symbolic := matches[1]
	mode, err := symbolicToOctal(symbolic)
	if err != nil {
		return "", err
	}

	// Format with leading zero for special bits
	if mode > 0777 {
		return fmt.Sprintf("%04o", mode), nil
	}
	return fmt.Sprintf("%03o", mode), nil
}

// handleChmodSymbolicSpaced handles spaced symbolic notation
// e.g., chmod rwx r-x r-x -> 755
func handleChmodSymbolicSpaced(expr string) (string, error) {
	re := regexp.MustCompile(`(?i)chmod\s+([rwxstST-]+)\s+([rwxstST-]+)\s+([rwxstST-]+)`)
	matches := re.FindStringSubmatch(expr)
	if len(matches) < 4 {
		return "", fmt.Errorf("invalid chmod expression: %s", expr)
	}

	// Combine the three parts
	symbolic := matches[1] + matches[2] + matches[3]
	mode, err := symbolicToOctal(symbolic)
	if err != nil {
		return "", err
	}

	// Format with leading zero for special bits
	if mode > 0777 {
		return fmt.Sprintf("%04o", mode), nil
	}
	return fmt.Sprintf("%03o", mode), nil
}

// handleUmask calculates file and directory permissions from umask
func handleUmask(expr string) (string, error) {
	re := regexp.MustCompile(`(?i)umask\s+([0-7]{3,4})`)
	matches := re.FindStringSubmatch(expr)
	if len(matches) < 2 {
		return "", fmt.Errorf("invalid umask expression: %s", expr)
	}

	umaskStr := matches[1]
	umask, err := strconv.ParseInt(umaskStr, 8, 64)
	if err != nil {
		return "", fmt.Errorf("invalid octal number: %s", umaskStr)
	}

	fileMode := DefaultFileMode &^ int(umask)
	dirMode := DefaultDirMode &^ int(umask)

	fileSymbolic := octalToSymbolic(fileMode)
	dirSymbolic := octalToSymbolic(dirMode)

	return fmt.Sprintf("files: %03o (%s), directories: %03o (%s)",
		fileMode, fileSymbolic, dirMode, dirSymbolic), nil
}

// handleOctalToSymbolic converts octal to symbolic
func handleOctalToSymbolic(expr string) (string, error) {
	re := regexp.MustCompile(`([0-7]{3,4})`)
	matches := re.FindStringSubmatch(expr)
	if len(matches) < 2 {
		return "", fmt.Errorf("invalid expression: %s", expr)
	}

	octalStr := matches[1]
	mode, err := strconv.ParseInt(octalStr, 8, 64)
	if err != nil {
		return "", fmt.Errorf("invalid octal number: %s", octalStr)
	}

	return octalToSymbolic(int(mode)), nil
}

// handleSymbolicToOctal converts symbolic to octal
func handleSymbolicToOctal(expr string) (string, error) {
	re := regexp.MustCompile(`([rwxstST-]{9,10})`)
	matches := re.FindStringSubmatch(expr)
	if len(matches) < 2 {
		return "", fmt.Errorf("invalid expression: %s", expr)
	}

	symbolic := matches[1]
	mode, err := symbolicToOctal(symbolic)
	if err != nil {
		return "", err
	}

	if mode > 0777 {
		return fmt.Sprintf("%04o", mode), nil
	}
	return fmt.Sprintf("%03o", mode), nil
}

// octalToSymbolic converts an octal mode to symbolic notation
func octalToSymbolic(mode int) string {
	var result strings.Builder

	// Owner permissions
	if mode&OwnerRead != 0 {
		result.WriteByte('r')
	} else {
		result.WriteByte('-')
	}
	if mode&OwnerWrite != 0 {
		result.WriteByte('w')
	} else {
		result.WriteByte('-')
	}
	// Owner execute with setuid
	if mode&Setuid != 0 {
		if mode&OwnerExecute != 0 {
			result.WriteByte('s')
		} else {
			result.WriteByte('S')
		}
	} else {
		if mode&OwnerExecute != 0 {
			result.WriteByte('x')
		} else {
			result.WriteByte('-')
		}
	}

	// Group permissions
	if mode&GroupRead != 0 {
		result.WriteByte('r')
	} else {
		result.WriteByte('-')
	}
	if mode&GroupWrite != 0 {
		result.WriteByte('w')
	} else {
		result.WriteByte('-')
	}
	// Group execute with setgid
	if mode&Setgid != 0 {
		if mode&GroupExecute != 0 {
			result.WriteByte('s')
		} else {
			result.WriteByte('S')
		}
	} else {
		if mode&GroupExecute != 0 {
			result.WriteByte('x')
		} else {
			result.WriteByte('-')
		}
	}

	// Other permissions
	if mode&OtherRead != 0 {
		result.WriteByte('r')
	} else {
		result.WriteByte('-')
	}
	if mode&OtherWrite != 0 {
		result.WriteByte('w')
	} else {
		result.WriteByte('-')
	}
	// Other execute with sticky bit
	if mode&Sticky != 0 {
		if mode&OtherExecute != 0 {
			result.WriteByte('t')
		} else {
			result.WriteByte('T')
		}
	} else {
		if mode&OtherExecute != 0 {
			result.WriteByte('x')
		} else {
			result.WriteByte('-')
		}
	}

	return result.String()
}

// symbolicToOctal converts symbolic notation to octal mode
func symbolicToOctal(symbolic string) (int, error) {
	// Don't convert to lowercase - S/s and T/t have different meanings

	if len(symbolic) != 9 && len(symbolic) != 10 {
		return 0, fmt.Errorf("invalid symbolic notation: %s (expected 9 or 10 characters)", symbolic)
	}

	// Handle 10-character notation with file type prefix
	if len(symbolic) == 10 {
		symbolic = symbolic[1:] // Skip the file type character
	}

	mode := 0

	// Owner permissions (positions 0-2)
	if symbolic[0] == 'r' {
		mode |= OwnerRead
	} else if symbolic[0] != '-' {
		return 0, fmt.Errorf("invalid character at position 1: %c", symbolic[0])
	}

	if symbolic[1] == 'w' {
		mode |= OwnerWrite
	} else if symbolic[1] != '-' {
		return 0, fmt.Errorf("invalid character at position 2: %c", symbolic[1])
	}

	switch symbolic[2] {
	case 'x':
		mode |= OwnerExecute
	case 's':
		mode |= OwnerExecute | Setuid
	case 'S':
		mode |= Setuid
	case '-':
		// no execute
	default:
		return 0, fmt.Errorf("invalid character at position 3: %c", symbolic[2])
	}

	// Group permissions (positions 3-5)
	if symbolic[3] == 'r' {
		mode |= GroupRead
	} else if symbolic[3] != '-' {
		return 0, fmt.Errorf("invalid character at position 4: %c", symbolic[3])
	}

	if symbolic[4] == 'w' {
		mode |= GroupWrite
	} else if symbolic[4] != '-' {
		return 0, fmt.Errorf("invalid character at position 5: %c", symbolic[4])
	}

	switch symbolic[5] {
	case 'x':
		mode |= GroupExecute
	case 's':
		mode |= GroupExecute | Setgid
	case 'S':
		mode |= Setgid
	case '-':
		// no execute
	default:
		return 0, fmt.Errorf("invalid character at position 6: %c", symbolic[5])
	}

	// Other permissions (positions 6-8)
	if symbolic[6] == 'r' {
		mode |= OtherRead
	} else if symbolic[6] != '-' {
		return 0, fmt.Errorf("invalid character at position 7: %c", symbolic[6])
	}

	if symbolic[7] == 'w' {
		mode |= OtherWrite
	} else if symbolic[7] != '-' {
		return 0, fmt.Errorf("invalid character at position 8: %c", symbolic[7])
	}

	switch symbolic[8] {
	case 'x':
		mode |= OtherExecute
	case 't':
		mode |= OtherExecute | Sticky
	case 'T':
		mode |= Sticky
	case '-':
		// no execute
	default:
		return 0, fmt.Errorf("invalid character at position 9: %c", symbolic[8])
	}

	return mode, nil
}

// getSpecialBitsDescription returns a description of special bits if any are set
func getSpecialBitsDescription(mode int) string {
	var parts []string

	if mode&Setuid != 0 {
		parts = append(parts, "setuid")
	}
	if mode&Setgid != 0 {
		parts = append(parts, "setgid")
	}
	if mode&Sticky != 0 {
		parts = append(parts, "sticky")
	}

	return strings.Join(parts, ", ")
}

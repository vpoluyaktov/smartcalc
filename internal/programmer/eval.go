package programmer

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// Handler defines the interface for programmer utility handlers.
type Handler interface {
	Handle(expr, exprLower string) (string, bool)
}

// HandlerFunc is an adapter to allow ordinary functions to be used as Handlers.
type HandlerFunc func(expr, exprLower string) (string, bool)

// Handle calls the underlying function.
func (f HandlerFunc) Handle(expr, exprLower string) (string, bool) {
	return f(expr, exprLower)
}

// handlerChain is the ordered list of handlers for programmer utilities.
var handlerChain = []Handler{
	HandlerFunc(handleAsciiTable),
	HandlerFunc(handleBitwiseAnd),
	HandlerFunc(handleBitwiseOr),
	HandlerFunc(handleBitwiseXor),
	HandlerFunc(handleBitwiseNot),
	HandlerFunc(handleLeftShift),
	HandlerFunc(handleRightShift),
	HandlerFunc(handleAsciiToChar),
	HandlerFunc(handleCharToAscii),
	HandlerFunc(handleUUID),
	HandlerFunc(handleMD5),
	HandlerFunc(handleSHA1),
	HandlerFunc(handleSHA256),
	HandlerFunc(handleBase64Encode),
	HandlerFunc(handleBase64Decode),
	HandlerFunc(handleRandomNumber),
}

// EvalProgrammer evaluates a programmer utility expression and returns the result.
func EvalProgrammer(expr string) (string, error) {
	expr = strings.TrimSpace(expr)
	exprLower := strings.ToLower(expr)

	for _, h := range handlerChain {
		if result, ok := h.Handle(expr, exprLower); ok {
			return result, nil
		}
	}

	return "", fmt.Errorf("unable to evaluate programmer expression: %s", expr)
}

// IsProgrammerExpression checks if an expression looks like a programmer utility.
func IsProgrammerExpression(expr string) bool {
	exprLower := strings.ToLower(expr)

	patterns := []string{
		`0x[0-9a-f]+\s+and\s+0x[0-9a-f]+`,
		`0x[0-9a-f]+\s+or\s+0x[0-9a-f]+`,
		`0x[0-9a-f]+\s+xor\s+0x[0-9a-f]+`,
		`not\s+0x[0-9a-f]+`,
		`(?:0x[0-9a-f]+|\d+)\s*<<\s*\d+`,
		`(?:0x[0-9a-f]+|\d+)\s*>>\s*\d+`,
		`^ascii\s+`,
		`^ascii\s*table$`,
		`^char\s+`,
		`^uuid$`,
		`^md5\s+`,
		`^sha1\s+`,
		`^sha256\s+`,
		`^random\s+`,
		`^base64\s+encode\s+`,
		`^base64\s+decode\s+`,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, exprLower); matched {
			return true
		}
	}

	return false
}

func handleAsciiTable(expr, exprLower string) (string, bool) {
	// Pattern: "ascii table"
	if exprLower != "ascii table" {
		return "", false
	}

	var sb strings.Builder

	// Control characters (0-31)
	sb.WriteString("\n> Control Characters (0-31):")
	sb.WriteString("\n> Dec Hex  Char | Dec Hex  Char | Dec Hex  Char | Dec Hex  Char")
	sb.WriteString("\n> --- ---- ---- | --- ---- ---- | --- ---- ---- | --- ---- ----")
	controlNames := []string{
		"NUL", "SOH", "STX", "ETX", "EOT", "ENQ", "ACK", "BEL",
		"BS", "TAB", "LF", "VT", "FF", "CR", "SO", "SI",
		"DLE", "DC1", "DC2", "DC3", "DC4", "NAK", "SYN", "ETB",
		"CAN", "EM", "SUB", "ESC", "FS", "GS", "RS", "US",
	}
	for i := 0; i < 32; i += 4 {
		sb.WriteString(fmt.Sprintf("\n> %3d 0x%02X %-4s", i, i, controlNames[i]))
		if i+1 < 32 {
			sb.WriteString(fmt.Sprintf(" | %3d 0x%02X %-4s", i+1, i+1, controlNames[i+1]))
		}
		if i+2 < 32 {
			sb.WriteString(fmt.Sprintf(" | %3d 0x%02X %-4s", i+2, i+2, controlNames[i+2]))
		}
		if i+3 < 32 {
			sb.WriteString(fmt.Sprintf(" | %3d 0x%02X %-4s", i+3, i+3, controlNames[i+3]))
		}
	}

	// Printable characters (32-127)
	sb.WriteString("\n> ")
	sb.WriteString("\n> Printable Characters (32-127):")
	sb.WriteString("\n> Dec Hex Char | Dec Hex Char | Dec Hex Char | Dec Hex Char")
	sb.WriteString("\n> --- --- ---- | --- --- ---- | --- --- ---- | --- --- ----")
	for i := 32; i < 128; i += 4 {
		sb.WriteString(fmt.Sprintf("\n> %3d %02X  %-4c", i, i, rune(i)))
		if i+1 < 128 {
			sb.WriteString(fmt.Sprintf(" | %3d %02X  %-4c", i+1, i+1, rune(i+1)))
		}
		if i+2 < 128 {
			sb.WriteString(fmt.Sprintf(" | %3d %02X  %-4c", i+2, i+2, rune(i+2)))
		}
		if i+3 < 128 {
			if i+3 == 127 {
				sb.WriteString(fmt.Sprintf(" | %3d %02X  DEL", i+3, i+3))
			} else {
				sb.WriteString(fmt.Sprintf(" | %3d %02X  %-4c", i+3, i+3, rune(i+3)))
			}
		}
	}

	return sb.String(), true
}

func parseHexOrDec(s string) (int64, bool) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(strings.ToLower(s), "0x") {
		val, err := strconv.ParseInt(s[2:], 16, 64)
		return val, err == nil
	}
	if strings.HasPrefix(strings.ToLower(s), "0b") {
		val, err := strconv.ParseInt(s[2:], 2, 64)
		return val, err == nil
	}
	val, err := strconv.ParseInt(s, 10, 64)
	return val, err == nil
}

func handleBitwiseAnd(expr, exprLower string) (string, bool) {
	// Pattern: "0xFF AND 0x0F" or "255 and 15"
	re := regexp.MustCompile(`(?i)^(0x[0-9a-f]+|\d+)\s+and\s+(0x[0-9a-f]+|\d+)$`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	a, ok1 := parseHexOrDec(matches[1])
	b, ok2 := parseHexOrDec(matches[2])
	if !ok1 || !ok2 {
		return "", false
	}

	result := a & b
	return fmt.Sprintf("%d (0x%X)", result, result), true
}

func handleBitwiseOr(expr, exprLower string) (string, bool) {
	// Pattern: "0xFF OR 0x0F"
	re := regexp.MustCompile(`(?i)^(0x[0-9a-f]+|\d+)\s+or\s+(0x[0-9a-f]+|\d+)$`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	a, ok1 := parseHexOrDec(matches[1])
	b, ok2 := parseHexOrDec(matches[2])
	if !ok1 || !ok2 {
		return "", false
	}

	result := a | b
	return fmt.Sprintf("%d (0x%X)", result, result), true
}

func handleBitwiseXor(expr, exprLower string) (string, bool) {
	// Pattern: "0xFF XOR 0x0F"
	re := regexp.MustCompile(`(?i)^(0x[0-9a-f]+|\d+)\s+xor\s+(0x[0-9a-f]+|\d+)$`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	a, ok1 := parseHexOrDec(matches[1])
	b, ok2 := parseHexOrDec(matches[2])
	if !ok1 || !ok2 {
		return "", false
	}

	result := a ^ b
	return fmt.Sprintf("%d (0x%X)", result, result), true
}

func handleBitwiseNot(expr, exprLower string) (string, bool) {
	// Pattern: "NOT 0xFF" or "~0xFF"
	re := regexp.MustCompile(`(?i)^(?:not|~)\s*(0x[0-9a-f]+|\d+)$`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	a, ok := parseHexOrDec(matches[1])
	if !ok {
		return "", false
	}

	// Use 32-bit NOT for reasonable output
	result := ^uint32(a)
	return fmt.Sprintf("%d (0x%X)", result, result), true
}

func handleLeftShift(expr, exprLower string) (string, bool) {
	// Pattern: "1 << 8"
	re := regexp.MustCompile(`^(0x[0-9a-fA-F]+|\d+)\s*<<\s*(\d+)$`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	a, ok := parseHexOrDec(matches[1])
	if !ok {
		return "", false
	}
	shift, _ := strconv.Atoi(matches[2])

	result := a << shift
	return fmt.Sprintf("%d (0x%X)", result, result), true
}

func handleRightShift(expr, exprLower string) (string, bool) {
	// Pattern: "256 >> 4"
	re := regexp.MustCompile(`^(0x[0-9a-fA-F]+|\d+)\s*>>\s*(\d+)$`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	a, ok := parseHexOrDec(matches[1])
	if !ok {
		return "", false
	}
	shift, _ := strconv.Atoi(matches[2])

	result := a >> shift
	return fmt.Sprintf("%d (0x%X)", result, result), true
}

func handleAsciiToChar(expr, exprLower string) (string, bool) {
	// Pattern: "char 65" or "char 0x41"
	re := regexp.MustCompile(`(?i)^char\s+(0x[0-9a-f]+|\d+)$`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	code, ok := parseHexOrDec(matches[1])
	if !ok || code < 0 || code > 127 {
		return "", false
	}

	if code >= 32 && code <= 126 {
		return fmt.Sprintf("'%c'", rune(code)), true
	}
	// Non-printable characters
	names := map[int64]string{
		0: "NUL", 9: "TAB", 10: "LF", 13: "CR", 27: "ESC", 32: "SPACE", 127: "DEL",
	}
	if name, ok := names[code]; ok {
		return name, true
	}
	return fmt.Sprintf("0x%02X", code), true
}

func handleCharToAscii(expr, exprLower string) (string, bool) {
	// Pattern: "ascii A" or "ascii 'A'"
	re := regexp.MustCompile(`(?i)^ascii\s+['"]?([^'"]+)['"]?$`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	char := matches[1]
	if len(char) != 1 {
		return "", false
	}

	code := int(char[0])
	return fmt.Sprintf("%d (0x%02X)", code, code), true
}

func handleUUID(expr, exprLower string) (string, bool) {
	if exprLower != "uuid" && exprLower != "uuid()" {
		return "", false
	}

	return uuid.New().String(), true
}

func handleMD5(expr, exprLower string) (string, bool) {
	// Pattern: "md5 hello" or "md5 'hello world'"
	re := regexp.MustCompile(`(?i)^md5\s+['"]?(.+?)['"]?$`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	input := matches[1]
	hash := md5.Sum([]byte(input))
	return hex.EncodeToString(hash[:]), true
}

func handleSHA1(expr, exprLower string) (string, bool) {
	// Pattern: "sha1 hello"
	re := regexp.MustCompile(`(?i)^sha1\s+['"]?(.+?)['"]?$`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	input := matches[1]
	hash := sha1.Sum([]byte(input))
	return hex.EncodeToString(hash[:]), true
}

func handleSHA256(expr, exprLower string) (string, bool) {
	// Pattern: "sha256 hello"
	re := regexp.MustCompile(`(?i)^sha256\s+['"]?(.+?)['"]?$`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	input := matches[1]
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:]), true
}

func handleBase64Encode(expr, exprLower string) (string, bool) {
	// Pattern: "base64 encode hello" or "base64 encode 'hello world'"
	re := regexp.MustCompile(`(?i)^base64\s+encode\s+['"]?(.+?)['"]?$`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	input := matches[1]
	encoded := base64.StdEncoding.EncodeToString([]byte(input))
	return encoded, true
}

func handleBase64Decode(expr, exprLower string) (string, bool) {
	// Pattern: "base64 decode SGVsbG8gV29ybGQ="
	re := regexp.MustCompile(`(?i)^base64\s+decode\s+['"]?(.+?)['"]?$`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	input := matches[1]
	decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return "ERR: invalid base64", true
	}
	return string(decoded), true
}

func handleRandomNumber(expr, exprLower string) (string, bool) {
	// Pattern: "random 1 to 100" or "random 1-100"
	re := regexp.MustCompile(`(?i)^random\s+(\d+)\s*(?:to|-)\s*(\d+)$`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	min, _ := strconv.Atoi(matches[1])
	max, _ := strconv.Atoi(matches[2])

	if min > max {
		min, max = max, min
	}

	// Use uuid to generate randomness (simple approach)
	u := uuid.New()
	bytes := u[:]
	var num uint64
	for _, b := range bytes[:8] {
		num = num<<8 | uint64(b)
	}

	result := min + int(num%uint64(max-min+1))
	return fmt.Sprintf("%d", result), true
}

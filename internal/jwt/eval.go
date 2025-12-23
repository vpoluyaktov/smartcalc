package jwt

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// IsJWTExpression checks if an expression is a JWT decode expression
func IsJWTExpression(expr string) bool {
	exprLower := strings.ToLower(strings.TrimSpace(expr))

	patterns := []string{
		`^jwt\s+decode\s+`, // jwt decode <token>
		`^decode\s+jwt\s+`, // decode jwt <token>
		`^jwt\s+[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]*$`, // jwt <token>
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, exprLower); matched {
			return true
		}
	}

	// Also check if it looks like a raw JWT token (three base64 parts separated by dots)
	if isJWTToken(expr) {
		return true
	}

	return false
}

// isJWTToken checks if a string looks like a JWT token
func isJWTToken(s string) bool {
	s = strings.TrimSpace(s)
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return false
	}

	// Check if first two parts are valid base64url
	for i := 0; i < 2; i++ {
		if len(parts[i]) == 0 {
			return false
		}
		// Try to decode to verify it's valid base64url
		_, err := base64URLDecode(parts[i])
		if err != nil {
			return false
		}
	}

	// Third part (signature) can be empty for unsigned tokens
	return true
}

// EvalJWT evaluates a JWT expression and returns the decoded result
func EvalJWT(expr string) (string, error) {
	expr = strings.TrimSpace(expr)
	exprLower := strings.ToLower(expr)

	var token string

	// Extract token from different formats
	if strings.HasPrefix(exprLower, "jwt decode ") {
		token = strings.TrimSpace(expr[11:])
	} else if strings.HasPrefix(exprLower, "decode jwt ") {
		token = strings.TrimSpace(expr[11:])
	} else if strings.HasPrefix(exprLower, "jwt ") {
		token = strings.TrimSpace(expr[4:])
	} else {
		// Assume the whole expression is a JWT token
		token = expr
	}

	// Remove quotes if present
	token = strings.Trim(token, `"'`)

	return decodeJWT(token)
}

// decodeJWT decodes a JWT token and returns formatted output
func decodeJWT(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid JWT: expected 3 parts, got %d", len(parts))
	}

	// Decode header
	headerJSON, err := base64URLDecode(parts[0])
	if err != nil {
		return "", fmt.Errorf("invalid JWT header: %v", err)
	}

	var header map[string]interface{}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return "", fmt.Errorf("invalid JWT header JSON: %v", err)
	}

	// Decode payload
	payloadJSON, err := base64URLDecode(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid JWT payload: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return "", fmt.Errorf("invalid JWT payload JSON: %v", err)
	}

	// Format output
	return formatJWTOutput(header, payload, parts[2])
}

// base64URLDecode decodes a base64url encoded string
func base64URLDecode(s string) ([]byte, error) {
	// Add padding if necessary
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}

	// Replace URL-safe characters
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")

	return base64.StdEncoding.DecodeString(s)
}

// formatJWTOutput formats the decoded JWT for display
func formatJWTOutput(header, payload map[string]interface{}, signature string) (string, error) {
	var result strings.Builder

	// Format header
	result.WriteString("Header:\n")
	headerFormatted, err := formatJSON(header)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(headerFormatted, "\n") {
		result.WriteString(">   " + line + "\n")
	}

	// Format payload
	result.WriteString("> Payload:\n")
	payloadFormatted, err := formatJSONWithTimestamps(payload)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(payloadFormatted, "\n") {
		result.WriteString(">   " + line + "\n")
	}

	// Signature info
	result.WriteString("> Signature: ")
	if signature == "" {
		result.WriteString("(none - unsigned token)")
	} else {
		// Show truncated signature
		if len(signature) > 20 {
			result.WriteString(signature[:20] + "...")
		} else {
			result.WriteString(signature)
		}
	}

	// Add expiration status if present
	if exp, ok := payload["exp"]; ok {
		expTime := parseTimestamp(exp)
		if expTime != nil {
			result.WriteString("\n> Status: ")
			if time.Now().After(*expTime) {
				result.WriteString("⚠ EXPIRED")
			} else {
				remaining := time.Until(*expTime)
				result.WriteString(fmt.Sprintf("✓ Valid (expires in %s)", formatDuration(remaining)))
			}
		}
	}

	return result.String(), nil
}

// formatJSON formats a map as indented JSON
func formatJSON(data map[string]interface{}) (string, error) {
	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(formatted), nil
}

// formatJSONWithTimestamps formats JSON and adds human-readable timestamps
func formatJSONWithTimestamps(data map[string]interface{}) (string, error) {
	// Create a copy with formatted timestamps
	dataCopy := make(map[string]interface{})
	for k, v := range data {
		dataCopy[k] = v
	}

	// Add human-readable timestamps for common JWT time fields
	timeFields := []string{"exp", "iat", "nbf", "auth_time"}
	for _, field := range timeFields {
		if val, ok := dataCopy[field]; ok {
			if ts := parseTimestamp(val); ts != nil {
				dataCopy[field+"_readable"] = ts.Format("2006-01-02 15:04:05 MST")
			}
		}
	}

	formatted, err := json.MarshalIndent(dataCopy, "", "  ")
	if err != nil {
		return "", err
	}
	return string(formatted), nil
}

// parseTimestamp parses a Unix timestamp from various formats
func parseTimestamp(val interface{}) *time.Time {
	var ts int64

	switch v := val.(type) {
	case float64:
		ts = int64(v)
	case int64:
		ts = v
	case int:
		ts = int64(v)
	case json.Number:
		if i, err := v.Int64(); err == nil {
			ts = i
		} else {
			return nil
		}
	default:
		return nil
	}

	t := time.Unix(ts, 0)
	return &t
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < 0 {
		return "expired"
	}

	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
}

package jwt

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// Helper function to create a test JWT token
func createTestJWT(header, payload map[string]interface{}, signature string) string {
	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	return headerB64 + "." + payloadB64 + "." + signature
}

func TestIsJWTExpression(t *testing.T) {
	// Create a valid test token
	header := map[string]interface{}{"alg": "HS256", "typ": "JWT"}
	payload := map[string]interface{}{"sub": "1234567890", "name": "Test User"}
	validToken := createTestJWT(header, payload, "testsignature")

	tests := []struct {
		expr     string
		expected bool
	}{
		// jwt decode prefix
		{"jwt decode " + validToken, true},
		{"JWT decode " + validToken, true},
		{"JWT DECODE " + validToken, true},

		// decode jwt prefix
		{"decode jwt " + validToken, true},
		{"DECODE JWT " + validToken, true},

		// jwt prefix
		{"jwt " + validToken, true},
		{"JWT " + validToken, true},

		// Raw token
		{validToken, true},

		// Not JWT expressions
		{"jwt", false},
		{"decode", false},
		{"hello world", false},
		{"2 + 2", false},
		{"invalid.token", false},
		{"a.b", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr[:min(30, len(tt.expr))], func(t *testing.T) {
			result := IsJWTExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("IsJWTExpression(%q) = %v, want %v", tt.expr[:min(50, len(tt.expr))], result, tt.expected)
			}
		})
	}
}

func TestEvalJWT_BasicDecode(t *testing.T) {
	header := map[string]interface{}{
		"alg": "HS256",
		"typ": "JWT",
	}
	payload := map[string]interface{}{
		"sub":  "1234567890",
		"name": "John Doe",
		"iat":  float64(1516239022),
	}
	token := createTestJWT(header, payload, "SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c")

	result, err := EvalJWT("jwt decode " + token)
	if err != nil {
		t.Fatalf("EvalJWT returned error: %v", err)
	}

	// Check that result contains expected parts
	if !strings.Contains(result, "Header:") {
		t.Error("Result should contain 'Header:'")
	}
	if !strings.Contains(result, "Payload:") {
		t.Error("Result should contain 'Payload:'")
	}
	if !strings.Contains(result, "Signature:") {
		t.Error("Result should contain 'Signature:'")
	}
	if !strings.Contains(result, "HS256") {
		t.Error("Result should contain algorithm 'HS256'")
	}
	if !strings.Contains(result, "John Doe") {
		t.Error("Result should contain name 'John Doe'")
	}
	if !strings.Contains(result, "1234567890") {
		t.Error("Result should contain sub '1234567890'")
	}
}

func TestEvalJWT_WithExpiration(t *testing.T) {
	header := map[string]interface{}{
		"alg": "HS256",
		"typ": "JWT",
	}

	// Test with future expiration
	futureExp := time.Now().Add(24 * time.Hour).Unix()
	payload := map[string]interface{}{
		"sub": "1234567890",
		"exp": float64(futureExp),
	}
	token := createTestJWT(header, payload, "signature")

	result, err := EvalJWT("jwt " + token)
	if err != nil {
		t.Fatalf("EvalJWT returned error: %v", err)
	}

	if !strings.Contains(result, "Valid") {
		t.Error("Result should indicate token is valid")
	}
}

func TestEvalJWT_ExpiredToken(t *testing.T) {
	header := map[string]interface{}{
		"alg": "HS256",
		"typ": "JWT",
	}

	// Test with past expiration
	pastExp := time.Now().Add(-24 * time.Hour).Unix()
	payload := map[string]interface{}{
		"sub": "1234567890",
		"exp": float64(pastExp),
	}
	token := createTestJWT(header, payload, "signature")

	result, err := EvalJWT("jwt " + token)
	if err != nil {
		t.Fatalf("EvalJWT returned error: %v", err)
	}

	if !strings.Contains(result, "EXPIRED") {
		t.Error("Result should indicate token is expired")
	}
}

func TestEvalJWT_DifferentFormats(t *testing.T) {
	header := map[string]interface{}{"alg": "RS256", "typ": "JWT"}
	payload := map[string]interface{}{"sub": "user123", "role": "admin"}
	token := createTestJWT(header, payload, "sig")

	tests := []struct {
		name string
		expr string
	}{
		{"jwt decode prefix", "jwt decode " + token},
		{"decode jwt prefix", "decode jwt " + token},
		{"jwt prefix", "jwt " + token},
		{"raw token", token},
		{"with quotes", `jwt decode "` + token + `"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EvalJWT(tt.expr)
			if err != nil {
				t.Errorf("EvalJWT(%q) returned error: %v", tt.name, err)
				return
			}
			if !strings.Contains(result, "RS256") {
				t.Errorf("Result should contain algorithm 'RS256'")
			}
			if !strings.Contains(result, "user123") {
				t.Errorf("Result should contain sub 'user123'")
			}
			if !strings.Contains(result, "admin") {
				t.Errorf("Result should contain role 'admin'")
			}
		})
	}
}

func TestEvalJWT_InvalidToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{"too few parts", "header.payload"},
		{"too many parts", "a.b.c.d"},
		{"invalid base64 header", "!!!.payload.sig"},
		{"invalid json header", "aGVsbG8.payload.sig"}, // "hello" in base64
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := EvalJWT("jwt decode " + tt.token)
			if err == nil {
				t.Errorf("EvalJWT should return error for invalid token: %s", tt.name)
			}
		})
	}
}

func TestEvalJWT_UnsignedToken(t *testing.T) {
	header := map[string]interface{}{"alg": "none", "typ": "JWT"}
	payload := map[string]interface{}{"sub": "1234567890"}
	token := createTestJWT(header, payload, "")

	result, err := EvalJWT("jwt " + token)
	if err != nil {
		t.Fatalf("EvalJWT returned error: %v", err)
	}

	if !strings.Contains(result, "none") || !strings.Contains(result, "unsigned") {
		t.Error("Result should indicate unsigned token")
	}
}

func TestEvalJWT_CommonClaims(t *testing.T) {
	header := map[string]interface{}{"alg": "HS256", "typ": "JWT"}
	now := time.Now()
	payload := map[string]interface{}{
		"iss":       "https://example.com",
		"sub":       "user@example.com",
		"aud":       "my-app",
		"exp":       float64(now.Add(1 * time.Hour).Unix()),
		"nbf":       float64(now.Unix()),
		"iat":       float64(now.Unix()),
		"jti":       "unique-token-id",
		"name":      "Test User",
		"email":     "test@example.com",
		"role":      "admin",
		"auth_time": float64(now.Unix()),
	}
	token := createTestJWT(header, payload, "signature")

	result, err := EvalJWT("jwt decode " + token)
	if err != nil {
		t.Fatalf("EvalJWT returned error: %v", err)
	}

	// Check common claims are present
	expectedStrings := []string{
		"https://example.com",
		"user@example.com",
		"my-app",
		"unique-token-id",
		"Test User",
		"test@example.com",
		"admin",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("Result should contain %q", expected)
		}
	}

	// Check that readable timestamps are added
	if !strings.Contains(result, "_readable") {
		t.Error("Result should contain readable timestamp fields")
	}
}

func TestBase64URLDecode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"SGVsbG8", "Hello"},
		{"SGVsbG8gV29ybGQ", "Hello World"},
		{"eyJhbGciOiJIUzI1NiJ9", `{"alg":"HS256"}`},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result, err := base64URLDecode(tt.input)
			if err != nil {
				t.Errorf("base64URLDecode(%q) returned error: %v", tt.input, err)
				return
			}
			if string(result) != tt.expected {
				t.Errorf("base64URLDecode(%q) = %q, want %q", tt.input, string(result), tt.expected)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Second, "30s"},
		{5 * time.Minute, "5m"},
		{2 * time.Hour, "2h 0m"},
		{25 * time.Hour, "1d 1h 0m"},
		{48*time.Hour + 30*time.Minute, "2d 0h 30m"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}

func TestIsJWTToken(t *testing.T) {
	header := map[string]interface{}{"alg": "HS256"}
	payload := map[string]interface{}{"sub": "123"}
	validToken := createTestJWT(header, payload, "sig")

	tests := []struct {
		token    string
		expected bool
	}{
		{validToken, true},
		{"a.b.c", false},          // Invalid base64
		{"header.payload", false}, // Only 2 parts
		{"a.b.c.d", false},        // 4 parts
		{"", false},               // Empty
		{"..", false},             // Empty parts
	}

	for _, tt := range tests {
		t.Run(tt.token[:min(20, len(tt.token))], func(t *testing.T) {
			result := isJWTToken(tt.token)
			if result != tt.expected {
				t.Errorf("isJWTToken(%q) = %v, want %v", tt.token[:min(30, len(tt.token))], result, tt.expected)
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

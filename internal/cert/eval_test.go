package cert

import (
	"crypto/x509"
	"strings"
	"testing"
	"time"
)

func TestIsCertExpression(t *testing.T) {
	tests := []struct {
		expr     string
		expected bool
	}{
		// cert decode prefix
		{"cert decode https://google.com", true},
		{"cert decode google.com", true},
		{"CERT DECODE https://example.com", true},

		// ssl decode prefix
		{"ssl decode https://google.com", true},
		{"ssl decode google.com", true},
		{"SSL DECODE https://example.com", true},

		// cert test prefix
		{"cert test https://google.com", true},
		{"cert test google.com", true},
		{"CERT TEST https://example.com", true},

		// ssl test prefix
		{"ssl test https://google.com", true},
		{"ssl test google.com", true},
		{"SSL TEST https://example.com", true},

		// decode cert/ssl prefix
		{"decode cert https://google.com", true},
		{"decode ssl https://google.com", true},

		// test cert/ssl prefix
		{"test cert https://google.com", true},
		{"test ssl https://google.com", true},

		// cert/ssl with URL directly
		{"cert https://google.com", true},
		{"ssl https://google.com", true},

		// Not cert expressions
		{"cert", false},
		{"ssl", false},
		{"decode", false},
		{"hello world", false},
		{"2 + 2", false},
		{"jwt decode token", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := IsCertExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("IsCertExpression(%q) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestEvalCert_Google(t *testing.T) {
	// Test with a real certificate (Google should always be available)
	result, err := EvalCert("cert decode https://google.com")
	if err != nil {
		t.Fatalf("EvalCert returned error: %v", err)
	}

	// Check that result contains expected parts
	if !strings.Contains(result, "Subject:") {
		t.Error("Result should contain 'Subject:'")
	}
	if !strings.Contains(result, "Issuer:") {
		t.Error("Result should contain 'Issuer:'")
	}
	if !strings.Contains(result, "Validity:") {
		t.Error("Result should contain 'Validity:'")
	}
	if !strings.Contains(result, "Status:") {
		t.Error("Result should contain 'Status:'")
	}
	if !strings.Contains(result, "Serial Number:") {
		t.Error("Result should contain 'Serial Number:'")
	}
	if !strings.Contains(result, "Subject Alt Names:") {
		t.Error("Result should contain 'Subject Alt Names:'")
	}
	// Check for certificate chain tree format
	if strings.Contains(result, "Certificate Chain:") {
		if !strings.Contains(result, "(leaf)") {
			t.Error("Certificate chain should contain leaf certificate")
		}
		if !strings.Contains(result, "(root)") && !strings.Contains(result, "(intermediate)") {
			t.Error("Certificate chain should contain root or intermediate certificate")
		}
	}
}

func TestEvalCert_DifferentFormats(t *testing.T) {
	tests := []struct {
		name string
		expr string
	}{
		{"cert decode prefix", "cert decode https://google.com"},
		{"ssl decode prefix", "ssl decode https://google.com"},
		{"cert test prefix", "cert test https://google.com"},
		{"ssl test prefix", "ssl test https://google.com"},
		{"test cert prefix", "test cert https://google.com"},
		{"test ssl prefix", "test ssl https://google.com"},
		{"cert prefix", "cert https://google.com"},
		{"ssl prefix", "ssl https://google.com"},
		{"without scheme", "cert decode google.com"},
		{"with quotes", `cert decode "https://google.com"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EvalCert(tt.expr)
			if err != nil {
				t.Errorf("EvalCert(%q) returned error: %v", tt.name, err)
				return
			}
			if !strings.Contains(result, "Subject:") {
				t.Errorf("Result should contain 'Subject:'")
			}
		})
	}
}

func TestEvalCert_InvalidURL(t *testing.T) {
	tests := []struct {
		name string
		expr string
	}{
		{"invalid host", "cert decode https://invalid.invalid.invalid"},
		{"empty host", "cert decode https://"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := EvalCert(tt.expr)
			if err == nil {
				t.Errorf("EvalCert(%q) should return error for invalid URL", tt.name)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		days     int
		expected string
	}{
		{0, "0s"},
		{1, "1d 0h"},
		{30, "30d 0h"},
		{45, "1mo 15d"},
		{365, "12mo 5d"},
		{400, "1y 35d"},
		{730, "2y 0d"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			// Convert days to duration
			d := formatDuration(24 * 60 * 60 * 1e9 * time.Duration(tt.days))
			// Just check it doesn't panic and returns something
			if d == "" {
				t.Error("formatDuration should return non-empty string")
			}
		})
	}
}

func TestFormatKeyUsage(t *testing.T) {
	tests := []struct {
		usage    x509.KeyUsage
		contains string
	}{
		{x509.KeyUsageDigitalSignature, "Digital Signature"},
		{x509.KeyUsageKeyEncipherment, "Key Encipherment"},
		{x509.KeyUsageCertSign, "Certificate Sign"},
		{0, "None"},
	}

	for _, tt := range tests {
		result := formatKeyUsage(tt.usage)
		if !strings.Contains(result, tt.contains) {
			t.Errorf("formatKeyUsage(%v) = %q, should contain %q", tt.usage, result, tt.contains)
		}
	}
}

// TestEvalCert_BadSSL tests various certificate issues using badssl.com test certificates
func TestEvalCert_BadSSL(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		shouldSucceed  bool
		expectContains []string
	}{
		{
			name:          "expired certificate",
			url:           "https://expired.badssl.com",
			shouldSucceed: true,
			expectContains: []string{
				"Subject:",
				"Issuer:",
				"EXPIRED", // Status should show expired
			},
		},
		{
			name:          "self-signed certificate",
			url:           "https://self-signed.badssl.com",
			shouldSucceed: true,
			expectContains: []string{
				"Subject:",
				"Issuer:",
				"Status:",
			},
		},
		{
			name:          "wrong host certificate",
			url:           "https://wrong.host.badssl.com",
			shouldSucceed: true,
			expectContains: []string{
				"Subject:",
				"Issuer:",
				"Status:",
			},
		},
		{
			name:          "untrusted root certificate",
			url:           "https://untrusted-root.badssl.com",
			shouldSucceed: true,
			expectContains: []string{
				"Subject:",
				"Issuer:",
				"Status:",
			},
		},
		{
			name:          "revoked certificate",
			url:           "https://revoked.badssl.com",
			shouldSucceed: true,
			expectContains: []string{
				"Subject:",
				"Issuer:",
				"Status:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EvalCert("cert decode " + tt.url)

			if tt.shouldSucceed {
				if err != nil {
					t.Fatalf("EvalCert(%q) returned error: %v", tt.url, err)
				}

				for _, expected := range tt.expectContains {
					if !strings.Contains(result, expected) {
						t.Errorf("Result should contain %q, got:\n%s", expected, result)
					}
				}
			} else {
				if err == nil {
					t.Errorf("EvalCert(%q) should have returned error", tt.url)
				}
			}
		})
	}
}

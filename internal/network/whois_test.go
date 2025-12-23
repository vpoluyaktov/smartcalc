package network

import (
	"strings"
	"testing"
)

func TestIsWhoisExpression(t *testing.T) {
	tests := []struct {
		expr     string
		expected bool
	}{
		{"whois google.com", true},
		{"whois example.org", true},
		{"WHOIS GOOGLE.COM", true},
		{"Whois github.com", true},
		{"whois", false},
		{"google.com", false},
		{"dig google.com", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := IsWhoisExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("IsWhoisExpression(%q) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestEvalWhois_Google(t *testing.T) {
	result, err := EvalWhois("whois google.com")
	if err != nil {
		t.Fatalf("EvalWhois returned error: %v", err)
	}

	// Check that result contains expected parts
	if !strings.Contains(result, "WHOIS: google.com") {
		t.Error("Result should contain 'WHOIS: google.com'")
	}
	// Google.com should have registrar info
	if !strings.Contains(result, "Registrar") {
		t.Error("Result should contain 'Registrar'")
	}
}

func TestEvalWhois_DifferentFormats(t *testing.T) {
	tests := []struct {
		name string
		expr string
	}{
		{"simple domain", "whois example.com"},
		{"with quotes", "whois \"example.com\""},
		{"with https", "whois https://example.com"},
		{"with path", "whois example.com/path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EvalWhois(tt.expr)
			if err != nil {
				t.Errorf("EvalWhois(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, "WHOIS:") {
				t.Errorf("EvalWhois(%q) should contain 'WHOIS:'", tt.expr)
			}
		})
	}
}

func TestEvalWhois_EmptyDomain(t *testing.T) {
	_, err := EvalWhois("whois ")
	if err == nil {
		t.Error("EvalWhois with empty domain should return error")
	}
}

func TestGetWhoisServer(t *testing.T) {
	tests := []struct {
		domain   string
		expected string
	}{
		{"google.com", "whois.verisign-grs.com"},
		{"example.net", "whois.verisign-grs.com"},
		{"example.org", "whois.pir.org"},
		{"example.io", "whois.nic.io"},
		{"example.unknown", "whois.iana.org"},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			result := getWhoisServer(tt.domain)
			if result != tt.expected {
				t.Errorf("getWhoisServer(%q) = %q, want %q", tt.domain, result, tt.expected)
			}
		})
	}
}

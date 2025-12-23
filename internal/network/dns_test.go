package network

import (
	"strings"
	"testing"
)

func TestIsDNSExpression(t *testing.T) {
	tests := []struct {
		expr     string
		expected bool
	}{
		{"dig google.com", true},
		{"nslookup google.com", true},
		{"dns google.com", true},
		{"lookup google.com", true},
		{"resolve google.com", true},
		{"DIG GOOGLE.COM", true},
		{"NSLOOKUP google.com", true},
		{"dig", false},
		{"google.com", false},
		{"10.0.0.1/24", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := IsDNSExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("IsDNSExpression(%q) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestEvalDNS_Google(t *testing.T) {
	result, err := EvalDNS("dig google.com")
	if err != nil {
		t.Fatalf("EvalDNS returned error: %v", err)
	}

	// Check that result contains expected parts
	if !strings.Contains(result, "DNS Lookup: google.com") {
		t.Error("Result should contain 'DNS Lookup: google.com'")
	}
	if !strings.Contains(result, "A Records:") {
		t.Error("Result should contain 'A Records:'")
	}
}

func TestEvalDNS_DifferentFormats(t *testing.T) {
	tests := []struct {
		name string
		expr string
	}{
		{"dig prefix", "dig google.com"},
		{"nslookup prefix", "nslookup google.com"},
		{"dns prefix", "dns google.com"},
		{"lookup prefix", "lookup google.com"},
		{"resolve prefix", "resolve google.com"},
		{"with quotes", "dig \"google.com\""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EvalDNS(tt.expr)
			if err != nil {
				t.Errorf("EvalDNS(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, "DNS Lookup:") {
				t.Errorf("EvalDNS(%q) should contain 'DNS Lookup:'", tt.expr)
			}
		})
	}
}

func TestEvalDNS_InvalidDomain(t *testing.T) {
	tests := []struct {
		name string
		expr string
	}{
		{"empty domain", "dig "},
		{"invalid domain", "dig thisisnotavaliddomainname12345.invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := EvalDNS(tt.expr)
			if err == nil {
				t.Errorf("EvalDNS(%q) should return error", tt.expr)
			}
		})
	}
}

func TestEvalDNS_MXRecords(t *testing.T) {
	result, err := EvalDNS("dig gmail.com")
	if err != nil {
		t.Fatalf("EvalDNS returned error: %v", err)
	}

	// Gmail should have MX records
	if !strings.Contains(result, "MX Records:") {
		t.Error("Result for gmail.com should contain 'MX Records:'")
	}
}

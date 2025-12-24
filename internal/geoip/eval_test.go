package geoip

import (
	"testing"
)

func TestIsGeoIPExpression(t *testing.T) {
	tests := []struct {
		expr     string
		expected bool
	}{
		{"geoip 8.8.8.8", true},
		{"geoip 1.1.1.1", true},
		{"geoip 192.168.1.1", true},
		{"ip location 8.8.8.8", true},
		{"locate ip 8.8.8.8", true},
		{"where is 8.8.8.8", true},
		{"geoip 2001:4860:4860::8888", true},

		// Invalid expressions
		{"hello world", false},
		{"8.8.8.8", false},
		{"geoip", false},
		{"geoip hello", false},
		{"100 + 50", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := IsGeoIPExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("IsGeoIPExpression(%q) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestExtractIP(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"geoip 8.8.8.8", "8.8.8.8"},
		{"geoip 1.1.1.1", "1.1.1.1"},
		{"ip location 192.168.1.1", "192.168.1.1"},
		{"where is 10.0.0.1", "10.0.0.1"},
		{"no ip here", ""},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := extractIP(tt.expr)
			if result != tt.expected {
				t.Errorf("extractIP(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"127.0.0.1", true},
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"208.67.222.222", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			result := isPrivateIP(tt.ip)
			if result != tt.expected {
				t.Errorf("isPrivateIP(%q) = %v, want %v", tt.ip, result, tt.expected)
			}
		})
	}
}

func TestEvalGeoIP_PrivateIP(t *testing.T) {
	_, err := EvalGeoIP("geoip 192.168.1.1")
	if err == nil {
		t.Error("Expected error for private IP, got nil")
	}
}

func TestEvalGeoIP_InvalidIP(t *testing.T) {
	_, err := EvalGeoIP("geoip 999.999.999.999")
	if err == nil {
		t.Error("Expected error for invalid IP, got nil")
	}
}

func TestFormatGeoIPResult(t *testing.T) {
	result := &GeoIPResponse{
		City:       "Mountain View",
		RegionName: "California",
		Country:    "United States",
		ISP:        "Google LLC",
		Lat:        37.4056,
		Lon:        -122.0775,
		Timezone:   "America/Los_Angeles",
	}

	formatted := formatGeoIPResult(result)

	// Check that all expected fields are present
	if !contains(formatted, "Mountain View") {
		t.Error("Expected city in output")
	}
	if !contains(formatted, "California") {
		t.Error("Expected region in output")
	}
	if !contains(formatted, "United States") {
		t.Error("Expected country in output")
	}
	if !contains(formatted, "Google LLC") {
		t.Error("Expected ISP in output")
	}
	if !contains(formatted, "37.4056") {
		t.Error("Expected latitude in output")
	}
	if !contains(formatted, "America/Los_Angeles") {
		t.Error("Expected timezone in output")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

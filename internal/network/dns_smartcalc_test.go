package network

import (
	"strings"
	"testing"
)

func TestEvalDNS_SmartCalcOrg(t *testing.T) {
	result, err := EvalDNS("dig smart-calc.org")
	if err != nil {
		t.Fatalf("EvalDNS returned error: %v", err)
	}

	// Check that result contains expected parts
	if !strings.Contains(result, "DNS Lookup: smart-calc.org") {
		t.Error("Result should contain 'DNS Lookup: smart-calc.org'")
	}

	// Verify it returns GitHub Pages IPs (not Cisco Umbrella IPs)
	expectedIPs := []string{"185.199.108.153", "185.199.109.153", "185.199.110.153", "185.199.111.153"}
	foundCount := 0
	for _, ip := range expectedIPs {
		if strings.Contains(result, ip) {
			foundCount++
		}
	}

	if foundCount == 0 {
		t.Errorf("Result should contain at least one GitHub Pages IP.\nGot: %s", result)
	}

	// Make sure it does NOT contain Cisco Umbrella IPs (which would indicate system DNS interception)
	ciscoIPs := []string{"146.112.38."}
	for _, ip := range ciscoIPs {
		if strings.Contains(result, ip) {
			t.Errorf("Result should NOT contain Cisco Umbrella IP %s (DNS interception detected).\nGot: %s", ip, result)
		}
	}

	t.Logf("DNS lookup result:\n%s", result)
}

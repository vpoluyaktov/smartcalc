package network

import (
	"strings"
	"testing"
)

func TestEvalSplitToSubnets(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"split 10.100.0.0/16 to 6 subnets", "10.100.0.0/19"},
		{"split 10.100.0.0/24 to 4 subnets", "10.100.0.0/26"},
		{"divide 192.168.0.0/24 into 2 subnets", "192.168.0.0/25"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalNetwork(tt.expr)
			if err != nil {
				t.Errorf("EvalNetwork(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("EvalNetwork(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

func TestEvalSplitByHosts(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"split 10.200.0.0/16 to subnets with 1024 hosts", "/21"},
		{"split 192.168.0.0/24 to subnets with 60 hosts", "/26"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalNetwork(tt.expr)
			if err != nil {
				t.Errorf("EvalNetwork(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("EvalNetwork(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

func TestEvalHostCount(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"how many hosts in 10.100.0.0/28", "14 hosts"},
		{"hosts in 10.100.0.0/24", "254 hosts"},
		{"hosts in /24", "254 hosts"},
		{"hosts in /30", "2 hosts"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalNetwork(tt.expr)
			if err != nil {
				t.Errorf("EvalNetwork(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalNetwork(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestEvalSubnetInfo(t *testing.T) {
	tests := []struct {
		expr     string
		contains []string
	}{
		{"subnet info 10.100.0.0/24", []string{"Network:", "Mask:", "Hosts:", "Range:", "Broadcast:"}},
		{"info for 192.168.1.0/28", []string{"192.168.1.0", "255.255.255.240", "14"}},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalNetwork(tt.expr)
			if err != nil {
				t.Errorf("EvalNetwork(%q) error: %v", tt.expr, err)
				return
			}
			for _, c := range tt.contains {
				if !strings.Contains(result, c) {
					t.Errorf("EvalNetwork(%q) = %q, want to contain %q", tt.expr, result, c)
				}
			}
		})
	}
}

func TestEvalMaskForPrefix(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"mask for /24", "255.255.255.0"},
		{"netmask /16", "255.255.0.0"},
		{"subnet mask for /28", "255.255.255.240"},
		{"mask for 8", "255.0.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalNetwork(tt.expr)
			if err != nil {
				t.Errorf("EvalNetwork(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalNetwork(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestEvalWildcardMask(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"wildcard for /24", "0.0.0.255"},
		{"wildcard mask /16", "0.0.255.255"},
		{"wildcard /28", "0.0.0.15"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalNetwork(tt.expr)
			if err != nil {
				t.Errorf("EvalNetwork(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalNetwork(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestEvalPrefixFromMask(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"prefix for 255.255.255.0", "/24"},
		{"cidr for 255.255.0.0", "/16"},
		{"prefix for 255.255.255.240", "/28"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalNetwork(tt.expr)
			if err != nil {
				t.Errorf("EvalNetwork(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalNetwork(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestEvalIPInRange(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"is 10.100.0.50 in 10.100.0.0/24", "yes"},
		{"is 10.100.1.50 in 10.100.0.0/24", "no"},
		{"is 192.168.1.100 in 192.168.1.0/28", "no"},
		{"is 192.168.1.10 in 192.168.1.0/28", "yes"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalNetwork(tt.expr)
			if err != nil {
				t.Errorf("EvalNetwork(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalNetwork(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestEvalNextSubnet(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"next subnet after 10.100.0.0/24", "10.100.1.0/24"},
		{"next subnet after 192.168.0.0/28", "192.168.0.16/28"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalNetwork(tt.expr)
			if err != nil {
				t.Errorf("EvalNetwork(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalNetwork(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestEvalBroadcast(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"broadcast for 10.100.0.0/24", "10.100.0.255"},
		{"broadcast of 192.168.1.0/28", "192.168.1.15"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalNetwork(tt.expr)
			if err != nil {
				t.Errorf("EvalNetwork(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalNetwork(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestEvalNetworkAddress(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"network for 10.100.0.50/24", "10.100.0.0/24"},
		{"network address 192.168.1.100/28", "192.168.1.96/28"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalNetwork(tt.expr)
			if err != nil {
				t.Errorf("EvalNetwork(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalNetwork(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestEvalJustCIDR(t *testing.T) {
	tests := []struct {
		expr     string
		contains []string
	}{
		{"10.100.0.0/24", []string{"10.100.0.0/24", "254", "255.255.255.0"}},
		{"192.168.1.0/28", []string{"192.168.1.0/28", "14", "255.255.255.240"}},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalNetwork(tt.expr)
			if err != nil {
				t.Errorf("EvalNetwork(%q) error: %v", tt.expr, err)
				return
			}
			for _, c := range tt.contains {
				if !strings.Contains(result, c) {
					t.Errorf("EvalNetwork(%q) = %q, want to contain %q", tt.expr, result, c)
				}
			}
		})
	}
}

func TestIsNetworkExpression(t *testing.T) {
	tests := []struct {
		expr     string
		expected bool
	}{
		{"split 10.100.0.0/16 to 6 subnets", true},
		{"hosts in 10.100.0.0/24", true},
		{"mask for /24", true},
		{"10.100.0.0/24", true},
		{"100 + 50", false},
		{"now in Seattle", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := IsNetworkExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("IsNetworkExpression(%q) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestParseCIDR(t *testing.T) {
	tests := []struct {
		cidr      string
		hostCount int64
		mask      string
	}{
		{"10.0.0.0/8", 16777214, "255.0.0.0"},
		{"192.168.0.0/24", 254, "255.255.255.0"},
		{"192.168.0.0/28", 14, "255.255.255.240"},
		{"192.168.0.0/30", 2, "255.255.255.252"},
		{"192.168.0.0/32", 1, "255.255.255.255"},
	}

	for _, tt := range tests {
		t.Run(tt.cidr, func(t *testing.T) {
			info, err := ParseCIDR(tt.cidr)
			if err != nil {
				t.Errorf("ParseCIDR(%q) error: %v", tt.cidr, err)
				return
			}
			if info.HostCount != tt.hostCount {
				t.Errorf("ParseCIDR(%q).HostCount = %d, want %d", tt.cidr, info.HostCount, tt.hostCount)
			}
			if info.Mask != tt.mask {
				t.Errorf("ParseCIDR(%q).Mask = %q, want %q", tt.cidr, info.Mask, tt.mask)
			}
		})
	}
}

func TestSplitToSubnets(t *testing.T) {
	tests := []struct {
		cidr  string
		count int
		want  int
	}{
		{"10.0.0.0/24", 4, 4},
		{"10.0.0.0/24", 2, 2},
		{"10.0.0.0/16", 8, 8},
	}

	for _, tt := range tests {
		t.Run(tt.cidr, func(t *testing.T) {
			subnets, err := SplitToSubnets(tt.cidr, tt.count)
			if err != nil {
				t.Errorf("SplitToSubnets(%q, %d) error: %v", tt.cidr, tt.count, err)
				return
			}
			if len(subnets) != tt.want {
				t.Errorf("SplitToSubnets(%q, %d) returned %d subnets, want %d", tt.cidr, tt.count, len(subnets), tt.want)
			}
		})
	}
}

func TestSplitByHostCount(t *testing.T) {
	tests := []struct {
		cidr           string
		hostsPerSubnet int
		wantPrefix     int
	}{
		{"10.0.0.0/16", 1024, 21},  // 2046 hosts per /21
		{"192.168.0.0/24", 60, 26}, // 62 hosts per /26
	}

	for _, tt := range tests {
		t.Run(tt.cidr, func(t *testing.T) {
			subnets, err := SplitByHostCount(tt.cidr, tt.hostsPerSubnet)
			if err != nil {
				t.Errorf("SplitByHostCount(%q, %d) error: %v", tt.cidr, tt.hostsPerSubnet, err)
				return
			}
			if len(subnets) == 0 {
				t.Errorf("SplitByHostCount(%q, %d) returned no subnets", tt.cidr, tt.hostsPerSubnet)
				return
			}
			if subnets[0].CIDR != tt.wantPrefix {
				t.Errorf("SplitByHostCount(%q, %d) prefix = /%d, want /%d", tt.cidr, tt.hostsPerSubnet, subnets[0].CIDR, tt.wantPrefix)
			}
		})
	}
}

func TestHostsInPrefix(t *testing.T) {
	tests := []struct {
		prefix int
		want   int64
	}{
		{24, 254},
		{28, 14},
		{30, 2},
		{31, 2},
		{32, 1},
		{16, 65534},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.prefix)), func(t *testing.T) {
			got := HostsInPrefix(tt.prefix)
			if got != tt.want {
				t.Errorf("HostsInPrefix(%d) = %d, want %d", tt.prefix, got, tt.want)
			}
		})
	}
}

func TestCalculateMask(t *testing.T) {
	tests := []struct {
		prefix int
		want   string
	}{
		{8, "255.0.0.0"},
		{16, "255.255.0.0"},
		{24, "255.255.255.0"},
		{28, "255.255.255.240"},
		{32, "255.255.255.255"},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.prefix)), func(t *testing.T) {
			got, err := CalculateMask(tt.prefix)
			if err != nil {
				t.Errorf("CalculateMask(%d) error: %v", tt.prefix, err)
				return
			}
			if got != tt.want {
				t.Errorf("CalculateMask(%d) = %q, want %q", tt.prefix, got, tt.want)
			}
		})
	}
}

func TestWildcardMask(t *testing.T) {
	tests := []struct {
		prefix int
		want   string
	}{
		{24, "0.0.0.255"},
		{16, "0.0.255.255"},
		{28, "0.0.0.15"},
		{32, "0.0.0.0"},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.prefix)), func(t *testing.T) {
			got, err := WildcardMask(tt.prefix)
			if err != nil {
				t.Errorf("WildcardMask(%d) error: %v", tt.prefix, err)
				return
			}
			if got != tt.want {
				t.Errorf("WildcardMask(%d) = %q, want %q", tt.prefix, got, tt.want)
			}
		})
	}
}

func TestIPInRange(t *testing.T) {
	tests := []struct {
		ip   string
		cidr string
		want bool
	}{
		{"10.100.0.50", "10.100.0.0/24", true},
		{"10.100.1.50", "10.100.0.0/24", false},
		{"192.168.1.10", "192.168.1.0/28", true},
		{"192.168.1.20", "192.168.1.0/28", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			got, err := IPInRange(tt.ip, tt.cidr)
			if err != nil {
				t.Errorf("IPInRange(%q, %q) error: %v", tt.ip, tt.cidr, err)
				return
			}
			if got != tt.want {
				t.Errorf("IPInRange(%q, %q) = %v, want %v", tt.ip, tt.cidr, got, tt.want)
			}
		})
	}
}

func TestNextSubnet(t *testing.T) {
	tests := []struct {
		cidr string
		want string
	}{
		{"10.100.0.0/24", "10.100.1.0/24"},
		{"192.168.0.0/28", "192.168.0.16/28"},
		{"10.0.0.0/8", "11.0.0.0/8"},
	}

	for _, tt := range tests {
		t.Run(tt.cidr, func(t *testing.T) {
			got, err := NextSubnet(tt.cidr)
			if err != nil {
				t.Errorf("NextSubnet(%q) error: %v", tt.cidr, err)
				return
			}
			if got != tt.want {
				t.Errorf("NextSubnet(%q) = %q, want %q", tt.cidr, got, tt.want)
			}
		})
	}
}

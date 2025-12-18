package network

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// EvalNetwork evaluates a network/IP expression and returns the result
func EvalNetwork(expr string) (string, error) {
	expr = strings.TrimSpace(expr)
	exprLower := strings.ToLower(expr)

	// "split 10.100.0.0/16 to 6 subnets"
	if result, ok := trySplitToSubnets(exprLower); ok {
		return result, nil
	}

	// "split 10.200.0.0/16 to subnets with 1024 hosts" or "split ... with 1024 hosts each"
	if result, ok := trySplitByHosts(exprLower); ok {
		return result, nil
	}

	// "how many hosts in 10.100.0.0/28" or "hosts in 10.100.0.0/28"
	if result, ok := tryHostCount(exprLower); ok {
		return result, nil
	}

	// "subnet info 10.100.0.0/24" or "info 10.100.0.0/24"
	if result, ok := trySubnetInfo(exprLower); ok {
		return result, nil
	}

	// "wildcard for /24" or "wildcard mask /24" - check before regular mask
	if result, ok := tryWildcardMask(exprLower); ok {
		return result, nil
	}

	// "mask for /24" or "netmask /24" or "subnet mask for /24"
	if result, ok := tryMaskForPrefix(exprLower); ok {
		return result, nil
	}

	// "prefix for 255.255.255.0" or "cidr for 255.255.255.0"
	if result, ok := tryPrefixFromMask(exprLower); ok {
		return result, nil
	}

	// "is 10.100.0.50 in 10.100.0.0/24"
	if result, ok := tryIPInRange(exprLower); ok {
		return result, nil
	}

	// "next subnet after 10.100.0.0/24"
	if result, ok := tryNextSubnet(exprLower); ok {
		return result, nil
	}

	// "broadcast for 10.100.0.0/24" or "broadcast of 10.100.0.0/24"
	if result, ok := tryBroadcast(exprLower); ok {
		return result, nil
	}

	// "network for 10.100.0.50/24" or "network address 10.100.0.50/24"
	if result, ok := tryNetworkAddress(exprLower); ok {
		return result, nil
	}

	// Just a CIDR - return info
	if result, ok := tryJustCIDR(expr); ok {
		return result, nil
	}

	return "", fmt.Errorf("unable to evaluate network expression: %s", expr)
}

// IsNetworkExpression checks if an expression looks like a network/IP expression
func IsNetworkExpression(expr string) bool {
	exprLower := strings.ToLower(expr)

	// Check for IP address patterns with CIDR first - most reliable indicator
	ipCIDRPattern := `\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,2}`
	if matched, _ := regexp.MatchString(ipCIDRPattern, expr); matched {
		return true
	}

	// Keywords that indicate network expressions (must have IP-like context)
	networkKeywords := []string{
		"subnet", "subnets", "cidr", "netmask",
		"wildcard", "broadcast", "split",
	}

	for _, kw := range networkKeywords {
		if strings.Contains(exprLower, kw) {
			return true
		}
	}

	// "hosts in" only if followed by IP or prefix
	if strings.Contains(exprLower, "hosts") {
		// Check if it has IP pattern or /prefix
		if matched, _ := regexp.MatchString(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`, expr); matched {
			return true
		}
		if matched, _ := regexp.MatchString(`/\d{1,2}`, expr); matched {
			return true
		}
	}

	// "mask for" with prefix
	if strings.Contains(exprLower, "mask") && strings.Contains(exprLower, "/") {
		return true
	}

	// "prefix for" with IP-like mask
	if strings.Contains(exprLower, "prefix for") || strings.Contains(exprLower, "cidr for") {
		if matched, _ := regexp.MatchString(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`, expr); matched {
			return true
		}
	}

	return false
}

func trySplitToSubnets(expr string) (string, bool) {
	// Pattern: "split 10.100.0.0/16 to 6 subnets" or "divide 10.100.0.0/16 into 6 subnets"
	re := regexp.MustCompile(`(?:split|divide)\s+(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,2})\s+(?:to|into)\s+(\d+)\s+subnets?`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	cidr := matches[1]
	count, _ := strconv.Atoi(matches[2])

	subnets, err := SplitToSubnets(cidr, count)
	if err != nil {
		return fmt.Sprintf("Error: %s", err), true
	}

	return FormatSubnetList(subnets), true
}

func trySplitByHosts(expr string) (string, bool) {
	// Pattern: "split 10.200.0.0/16 to subnets with 1024 hosts" or "... with 1024 hosts each"
	re := regexp.MustCompile(`(?:split|divide)\s+(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,2})\s+(?:to|into)\s+subnets?\s+(?:with|of)\s+(\d+)\s+hosts?`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	cidr := matches[1]
	hosts, _ := strconv.Atoi(matches[2])

	subnets, err := SplitByHostCount(cidr, hosts)
	if err != nil {
		return fmt.Sprintf("Error: %s", err), true
	}

	return FormatSubnetList(subnets), true
}

func tryHostCount(expr string) (string, bool) {
	// Pattern: "how many hosts in 10.100.0.0/28" or "hosts in 10.100.0.0/28" or "host count 10.100.0.0/28"
	re := regexp.MustCompile(`(?:how\s+many\s+)?hosts?\s+(?:in|for|count)?\s*(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,2})`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		// Try just prefix: "hosts in /24"
		re = regexp.MustCompile(`(?:how\s+many\s+)?hosts?\s+(?:in|for)?\s*/(\d{1,2})`)
		matches = re.FindStringSubmatch(expr)
		if matches == nil {
			return "", false
		}
		prefix, _ := strconv.Atoi(matches[1])
		hosts := HostsInPrefix(prefix)
		return fmt.Sprintf("%d hosts", hosts), true
	}

	cidr := matches[1]
	info, err := ParseCIDR(cidr)
	if err != nil {
		return fmt.Sprintf("Error: %s", err), true
	}

	return fmt.Sprintf("%d hosts", info.HostCount), true
}

func trySubnetInfo(expr string) (string, bool) {
	// Pattern: "subnet info 10.100.0.0/24" or "info for 10.100.0.0/24"
	re := regexp.MustCompile(`(?:subnet\s+)?info\s+(?:for\s+)?(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,2})`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	cidr := matches[1]
	info, err := ParseCIDR(cidr)
	if err != nil {
		return fmt.Sprintf("Error: %s", err), true
	}

	return fmt.Sprintf("Network: %s/%d\nMask: %s\nHosts: %d\nRange: %s - %s\nBroadcast: %s",
		info.NetworkAddr, info.CIDR, info.Mask, info.HostCount, info.FirstHost, info.LastHost, info.Broadcast), true
}

func tryMaskForPrefix(expr string) (string, bool) {
	// Pattern: "mask for /24" or "netmask /24" or "subnet mask for /24"
	re := regexp.MustCompile(`(?:subnet\s+)?(?:net)?mask\s+(?:for\s+)?/?(\d{1,2})`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	prefix, _ := strconv.Atoi(matches[1])
	mask, err := CalculateMask(prefix)
	if err != nil {
		return fmt.Sprintf("Error: %s", err), true
	}

	return mask, true
}

func tryWildcardMask(expr string) (string, bool) {
	// Pattern: "wildcard for /24" or "wildcard mask /24"
	re := regexp.MustCompile(`wildcard\s+(?:mask\s+)?(?:for\s+)?/?(\d{1,2})`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	prefix, _ := strconv.Atoi(matches[1])
	wildcard, err := WildcardMask(prefix)
	if err != nil {
		return fmt.Sprintf("Error: %s", err), true
	}

	return wildcard, true
}

func tryPrefixFromMask(expr string) (string, bool) {
	// Pattern: "prefix for 255.255.255.0" or "cidr for 255.255.255.0"
	re := regexp.MustCompile(`(?:prefix|cidr)\s+(?:for\s+)?(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	mask := matches[1]
	prefix, err := PrefixFromMask(mask)
	if err != nil {
		return fmt.Sprintf("Error: %s", err), true
	}

	return fmt.Sprintf("/%d", prefix), true
}

func tryIPInRange(expr string) (string, bool) {
	// Pattern: "is 10.100.0.50 in 10.100.0.0/24"
	re := regexp.MustCompile(`is\s+(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\s+in\s+(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,2})`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	ip := matches[1]
	cidr := matches[2]

	inRange, err := IPInRange(ip, cidr)
	if err != nil {
		return fmt.Sprintf("Error: %s", err), true
	}

	if inRange {
		return "yes", true
	}
	return "no", true
}

func tryNextSubnet(expr string) (string, bool) {
	// Pattern: "next subnet after 10.100.0.0/24"
	re := regexp.MustCompile(`next\s+subnet\s+(?:after\s+)?(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,2})`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	cidr := matches[1]
	next, err := NextSubnet(cidr)
	if err != nil {
		return fmt.Sprintf("Error: %s", err), true
	}

	return next, true
}

func tryBroadcast(expr string) (string, bool) {
	// Pattern: "broadcast for 10.100.0.0/24" or "broadcast of 10.100.0.0/24"
	re := regexp.MustCompile(`broadcast\s+(?:for|of|address)?\s*(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,2})`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	cidr := matches[1]
	info, err := ParseCIDR(cidr)
	if err != nil {
		return fmt.Sprintf("Error: %s", err), true
	}

	return info.Broadcast, true
}

func tryNetworkAddress(expr string) (string, bool) {
	// Pattern: "network for 10.100.0.50/24" or "network address 10.100.0.50/24"
	re := regexp.MustCompile(`network\s+(?:for|of|address)?\s*(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,2})`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	cidr := matches[1]
	info, err := ParseCIDR(cidr)
	if err != nil {
		return fmt.Sprintf("Error: %s", err), true
	}

	return fmt.Sprintf("%s/%d", info.NetworkAddr, info.CIDR), true
}

func tryJustCIDR(expr string) (string, bool) {
	// Just a CIDR notation - return basic info
	re := regexp.MustCompile(`^(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,2})$`)
	matches := re.FindStringSubmatch(strings.TrimSpace(expr))
	if matches == nil {
		return "", false
	}

	cidr := matches[1]
	info, err := ParseCIDR(cidr)
	if err != nil {
		return fmt.Sprintf("Error: %s", err), true
	}

	return FormatSubnetInfo(info), true
}

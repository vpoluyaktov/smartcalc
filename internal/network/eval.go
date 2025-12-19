package network

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Handler defines the interface for network expression handlers.
// Each handler attempts to process an expression and returns the result
// along with a boolean indicating whether it handled the expression.
type Handler interface {
	Handle(expr, exprLower string) (string, bool)
}

// HandlerFunc is an adapter to allow ordinary functions to be used as Handlers.
type HandlerFunc func(expr, exprLower string) (string, bool)

// Handle calls the underlying function.
func (f HandlerFunc) Handle(expr, exprLower string) (string, bool) {
	return f(expr, exprLower)
}

// handlerChain is the ordered list of handlers for network expressions.
// Handlers are tried in order; the first one that returns ok=true wins.
var handlerChain = []Handler{
	HandlerFunc(handleDivideToSubnets),
	HandlerFunc(handleDivideByHosts),
	HandlerFunc(handleHostCount),
	HandlerFunc(handleSubnetInfo),
	HandlerFunc(handleWildcardMask), // must be before handleMaskForPrefix
	HandlerFunc(handleMaskForPrefix),
	HandlerFunc(handlePrefixFromMask),
	HandlerFunc(handleIPInRange),
	HandlerFunc(handleNextSubnet),
	HandlerFunc(handleBroadcast),
	HandlerFunc(handleNetworkAddress),
	HandlerFunc(handleJustCIDR),
}

// EvalNetwork evaluates a network/IP expression and returns the result.
// It uses the Chain of Responsibility pattern to delegate to handlers.
func EvalNetwork(expr string) (string, error) {
	expr = strings.TrimSpace(expr)
	exprLower := strings.ToLower(expr)

	for _, h := range handlerChain {
		if result, ok := h.Handle(expr, exprLower); ok {
			return result, nil
		}
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
		"subnet", "subnets", "network", "networks", "cidr", "netmask",
		"wildcard", "broadcast",
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

func handleDivideToSubnets(expr, exprLower string) (string, bool) {
	// Pattern: "10.100.0.0/16 / 4 subnets" or "10.100.0.0/16 / 4 networks"
	re := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,2})\s*/\s*(\d+)\s+(?:subnets?|networks?)`)
	matches := re.FindStringSubmatch(exprLower)
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

func handleDivideByHosts(expr, exprLower string) (string, bool) {
	// Pattern: "10.100.0.0/16 / 1024 hosts"
	re := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,2})\s*/\s*(\d+)\s+hosts?`)
	matches := re.FindStringSubmatch(exprLower)
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

func handleHostCount(expr, exprLower string) (string, bool) {
	// Pattern: "how many hosts in 10.100.0.0/28" or "hosts in 10.100.0.0/28" or "host count 10.100.0.0/28"
	re := regexp.MustCompile(`(?:how\s+many\s+)?hosts?\s+(?:in|for|count)?\s*(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,2})`)
	matches := re.FindStringSubmatch(exprLower)
	if matches == nil {
		// Try just prefix: "hosts in /24"
		re = regexp.MustCompile(`(?:how\s+many\s+)?hosts?\s+(?:in|for)?\s*/(\d{1,2})`)
		matches = re.FindStringSubmatch(exprLower)
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

func handleSubnetInfo(expr, exprLower string) (string, bool) {
	// Pattern: "subnet info 10.100.0.0/24" or "info for 10.100.0.0/24"
	re := regexp.MustCompile(`(?:subnet\s+)?info\s+(?:for\s+)?(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,2})`)
	matches := re.FindStringSubmatch(exprLower)
	if matches == nil {
		return "", false
	}

	cidr := matches[1]
	info, err := ParseCIDR(cidr)
	if err != nil {
		return fmt.Sprintf("Error: %s", err), true
	}

	return fmt.Sprintf("\n> Network: %s/%d\n> Mask: %s\n> Hosts: %d\n> Range: %s - %s\n> Broadcast: %s",
		info.NetworkAddr, info.CIDR, info.Mask, info.HostCount, info.FirstHost, info.LastHost, info.Broadcast), true
}

func handleMaskForPrefix(expr, exprLower string) (string, bool) {
	// Pattern: "mask for /24" or "netmask /24" or "subnet mask for /24"
	re := regexp.MustCompile(`(?:subnet\s+)?(?:net)?mask\s+(?:for\s+)?/?(\d{1,2})`)
	matches := re.FindStringSubmatch(exprLower)
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

func handleWildcardMask(expr, exprLower string) (string, bool) {
	// Pattern: "wildcard for /24" or "wildcard mask /24"
	re := regexp.MustCompile(`wildcard\s+(?:mask\s+)?(?:for\s+)?/?(\d{1,2})`)
	matches := re.FindStringSubmatch(exprLower)
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

func handlePrefixFromMask(expr, exprLower string) (string, bool) {
	// Pattern: "prefix for 255.255.255.0" or "cidr for 255.255.255.0"
	re := regexp.MustCompile(`(?:prefix|cidr)\s+(?:for\s+)?(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)
	matches := re.FindStringSubmatch(exprLower)
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

func handleIPInRange(expr, exprLower string) (string, bool) {
	// Pattern: "is 10.100.0.50 in 10.100.0.0/24"
	re := regexp.MustCompile(`is\s+(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\s+in\s+(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,2})`)
	matches := re.FindStringSubmatch(exprLower)
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

func handleNextSubnet(expr, exprLower string) (string, bool) {
	// Pattern: "next subnet after 10.100.0.0/24"
	re := regexp.MustCompile(`next\s+subnet\s+(?:after\s+)?(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,2})`)
	matches := re.FindStringSubmatch(exprLower)
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

func handleBroadcast(expr, exprLower string) (string, bool) {
	// Pattern: "broadcast for 10.100.0.0/24" or "broadcast of 10.100.0.0/24"
	re := regexp.MustCompile(`broadcast\s+(?:for|of|address)?\s*(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,2})`)
	matches := re.FindStringSubmatch(exprLower)
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

func handleNetworkAddress(expr, exprLower string) (string, bool) {
	// Pattern: "network for 10.100.0.50/24" or "network address 10.100.0.50/24"
	re := regexp.MustCompile(`network\s+(?:for|of|address)?\s*(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,2})`)
	matches := re.FindStringSubmatch(exprLower)
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

func handleJustCIDR(expr, exprLower string) (string, bool) {
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

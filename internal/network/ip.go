package network

import (
	"fmt"
	"math"
	"net"
	"strings"
)

// SubnetInfo holds information about a subnet
type SubnetInfo struct {
	Network     net.IPNet
	NetworkAddr string
	Broadcast   string
	FirstHost   string
	LastHost    string
	HostCount   int64
	Mask        string
	CIDR        int
}

// ParseCIDR parses a CIDR notation string and returns subnet info
func ParseCIDR(cidr string) (*SubnetInfo, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR: %s", cidr)
	}
	return getSubnetInfo(ipnet), nil
}

// getSubnetInfo calculates all subnet information
func getSubnetInfo(ipnet *net.IPNet) *SubnetInfo {
	ones, bits := ipnet.Mask.Size()

	// Calculate host count (excluding network and broadcast for IPv4)
	hostBits := bits - ones
	var hostCount int64
	if hostBits > 1 {
		hostCount = (1 << hostBits) - 2 // subtract network and broadcast
	} else if hostBits == 1 {
		hostCount = 2 // /31 point-to-point
	} else {
		hostCount = 1 // /32 single host
	}

	networkAddr := ipnet.IP.String()
	broadcast := getBroadcast(ipnet)
	firstHost, lastHost := getHostRange(ipnet)

	return &SubnetInfo{
		Network:     *ipnet,
		NetworkAddr: networkAddr,
		Broadcast:   broadcast,
		FirstHost:   firstHost,
		LastHost:    lastHost,
		HostCount:   hostCount,
		Mask:        net.IP(ipnet.Mask).String(),
		CIDR:        ones,
	}
}

// getBroadcast calculates the broadcast address
func getBroadcast(ipnet *net.IPNet) string {
	ip := ipnet.IP.To4()
	if ip == nil {
		ip = ipnet.IP.To16()
	}

	broadcast := make(net.IP, len(ip))
	for i := range ip {
		broadcast[i] = ip[i] | ^ipnet.Mask[i]
	}
	return broadcast.String()
}

// getHostRange returns first and last usable host addresses
func getHostRange(ipnet *net.IPNet) (string, string) {
	ones, bits := ipnet.Mask.Size()
	hostBits := bits - ones

	ip := ipnet.IP.To4()
	if ip == nil {
		ip = ipnet.IP.To16()
	}

	if hostBits <= 1 {
		// /31 or /32 - return network address
		return ip.String(), ip.String()
	}

	// First host = network + 1
	firstHost := make(net.IP, len(ip))
	copy(firstHost, ip)
	firstHost[len(firstHost)-1]++

	// Last host = broadcast - 1
	broadcast := make(net.IP, len(ip))
	for i := range ip {
		broadcast[i] = ip[i] | ^ipnet.Mask[i]
	}
	broadcast[len(broadcast)-1]--

	return firstHost.String(), broadcast.String()
}

// SplitToSubnets splits a network into n equal subnets
func SplitToSubnets(cidr string, count int) ([]SubnetInfo, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR: %s", cidr)
	}

	ones, bits := ipnet.Mask.Size()

	// Calculate required additional bits
	additionalBits := int(math.Ceil(math.Log2(float64(count))))
	newPrefix := ones + additionalBits

	if newPrefix > bits {
		return nil, fmt.Errorf("cannot split /%d into %d subnets", ones, count)
	}

	// Calculate actual number of subnets
	actualCount := 1 << additionalBits

	subnets := make([]SubnetInfo, 0, count)
	ip := ipnet.IP.To4()
	if ip == nil {
		ip = ipnet.IP.To16()
	}

	// Size of each subnet in IP addresses
	subnetSize := 1 << (bits - newPrefix)

	for i := 0; i < actualCount && i < count; i++ {
		// Calculate subnet address
		offset := i * subnetSize
		subnetIP := addToIP(ip, offset)

		newMask := net.CIDRMask(newPrefix, bits)
		newNet := &net.IPNet{IP: subnetIP, Mask: newMask}

		subnets = append(subnets, *getSubnetInfo(newNet))
	}

	return subnets, nil
}

// SplitByHostCount splits a network into subnets with at least n hosts each
func SplitByHostCount(cidr string, hostsPerSubnet int) ([]SubnetInfo, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR: %s", cidr)
	}

	ones, bits := ipnet.Mask.Size()

	// Calculate required host bits (add 2 for network and broadcast)
	requiredHostBits := int(math.Ceil(math.Log2(float64(hostsPerSubnet + 2))))
	if requiredHostBits < 2 {
		requiredHostBits = 2
	}

	newPrefix := bits - requiredHostBits
	if newPrefix < ones {
		return nil, fmt.Errorf("cannot fit %d hosts in /%d network", hostsPerSubnet, ones)
	}

	// Calculate how many subnets we can create
	subnetCount := 1 << (newPrefix - ones)

	subnets := make([]SubnetInfo, 0, subnetCount)
	ip := ipnet.IP.To4()
	if ip == nil {
		ip = ipnet.IP.To16()
	}

	subnetSize := 1 << requiredHostBits

	for i := 0; i < subnetCount; i++ {
		offset := i * subnetSize
		subnetIP := addToIP(ip, offset)

		newMask := net.CIDRMask(newPrefix, bits)
		newNet := &net.IPNet{IP: subnetIP, Mask: newMask}

		subnets = append(subnets, *getSubnetInfo(newNet))
	}

	return subnets, nil
}

// addToIP adds an offset to an IP address
func addToIP(ip net.IP, offset int) net.IP {
	result := make(net.IP, len(ip))
	copy(result, ip)

	for i := len(result) - 1; i >= 0 && offset > 0; i-- {
		sum := int(result[i]) + (offset & 0xFF)
		result[i] = byte(sum & 0xFF)
		offset = (offset >> 8) + (sum >> 8)
	}

	return result
}

// CalculateMask returns subnet mask for a given prefix length
func CalculateMask(prefix int) (string, error) {
	if prefix < 0 || prefix > 32 {
		return "", fmt.Errorf("invalid prefix length: %d", prefix)
	}
	mask := net.CIDRMask(prefix, 32)
	return net.IP(mask).String(), nil
}

// PrefixFromMask converts a subnet mask to prefix length
func PrefixFromMask(mask string) (int, error) {
	ip := net.ParseIP(mask)
	if ip == nil {
		return 0, fmt.Errorf("invalid mask: %s", mask)
	}

	ip4 := ip.To4()
	if ip4 == nil {
		return 0, fmt.Errorf("not an IPv4 mask: %s", mask)
	}

	ones, _ := net.IPMask(ip4).Size()
	return ones, nil
}

// HostsInPrefix calculates number of usable hosts for a prefix
func HostsInPrefix(prefix int) int64 {
	if prefix > 30 {
		if prefix == 31 {
			return 2
		}
		return 1
	}
	hostBits := 32 - prefix
	return (1 << hostBits) - 2
}

// FormatSubnetInfo formats subnet info for display
func FormatSubnetInfo(info *SubnetInfo) string {
	return fmt.Sprintf("%s/%d (hosts: %d, range: %s - %s, mask: %s)",
		info.NetworkAddr, info.CIDR, info.HostCount, info.FirstHost, info.LastHost, info.Mask)
}

// FormatSubnetList formats a list of subnets for display
func FormatSubnetList(subnets []SubnetInfo) string {
	if len(subnets) == 0 {
		return "no subnets"
	}

	var sb strings.Builder
	for i, s := range subnets {
		// Each line starts with newline (first line too, so it appears on its own line)
		sb.WriteString("\n")
		// Prefix with "> " so output lines are not re-parsed
		sb.WriteString(fmt.Sprintf("> %d: %s/%d (%d hosts)", i+1, s.NetworkAddr, s.CIDR, s.HostCount))
	}
	return sb.String()
}

// IPInRange checks if an IP is within a CIDR range
func IPInRange(ip string, cidr string) (bool, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false, fmt.Errorf("invalid IP: %s", ip)
	}

	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false, fmt.Errorf("invalid CIDR: %s", cidr)
	}

	return ipnet.Contains(parsedIP), nil
}

// NextSubnet returns the next subnet of the same size
func NextSubnet(cidr string) (string, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", fmt.Errorf("invalid CIDR: %s", cidr)
	}

	ones, bits := ipnet.Mask.Size()
	subnetSize := 1 << (bits - ones)

	nextIP := addToIP(ipnet.IP, subnetSize)
	return fmt.Sprintf("%s/%d", nextIP.String(), ones), nil
}

// WildcardMask returns the wildcard mask for a prefix
func WildcardMask(prefix int) (string, error) {
	if prefix < 0 || prefix > 32 {
		return "", fmt.Errorf("invalid prefix: %d", prefix)
	}

	mask := net.CIDRMask(prefix, 32)
	wildcard := make(net.IP, 4)
	for i := range mask {
		wildcard[i] = ^mask[i]
	}
	return wildcard.String(), nil
}

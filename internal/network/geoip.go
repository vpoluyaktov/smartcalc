package network

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// GeoIPResponse represents the response from ip-api.com
type GeoIPResponse struct {
	Status      string  `json:"status"`
	Message     string  `json:"message"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
	Query       string  `json:"query"`
}

// IsGeoIPExpression checks if an expression is a geoip lookup
func IsGeoIPExpression(expr string) bool {
	expr = strings.TrimSpace(strings.ToLower(expr))
	patterns := []string{
		`^geoip\s+\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`,
		`^geoip\s+[a-f0-9:]+$`, // IPv6
		`^ip\s+location\s+\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`,
		`^locate\s+ip\s+\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`,
		`^where\s+is\s+\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, expr); matched {
			return true
		}
	}

	return false
}

// EvalGeoIP evaluates a geoip expression and returns location info
func EvalGeoIP(expr string) (string, error) {
	ip := extractIP(expr)
	if ip == "" {
		return "", fmt.Errorf("no valid IP address found")
	}

	// Validate IP address
	if net.ParseIP(ip) == nil {
		return "", fmt.Errorf("invalid IP address: %s", ip)
	}

	// Check for private/reserved IPs
	if isPrivateIP(ip) {
		return "", fmt.Errorf("cannot geolocate private IP address: %s", ip)
	}

	// Query ip-api.com
	result, err := lookupIP(ip)
	if err != nil {
		return "", err
	}

	return formatGeoIPResult(result), nil
}

// extractIP extracts the IP address from the expression
func extractIP(expr string) string {
	// IPv4 pattern
	ipv4Pattern := regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
	if match := ipv4Pattern.FindString(expr); match != "" {
		return match
	}

	// IPv6 pattern (simplified)
	ipv6Pattern := regexp.MustCompile(`[a-fA-F0-9:]+:[a-fA-F0-9:]+`)
	if match := ipv6Pattern.FindString(expr); match != "" {
		return match
	}

	return ""
}

// isPrivateIP checks if an IP is private/reserved
func isPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// Check for private ranges
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"::1/128",
		"fc00::/7",
		"fe80::/10",
	}

	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}

	return false
}

// lookupIP queries ip-api.com for geolocation data
func lookupIP(ip string) (*GeoIPResponse, error) {
	url := fmt.Sprintf("http://ip-api.com/json/%s", ip)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to query geoip service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geoip service returned status %d", resp.StatusCode)
	}

	var result GeoIPResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse geoip response: %v", err)
	}

	if result.Status == "fail" {
		return nil, fmt.Errorf("geoip lookup failed: %s", result.Message)
	}

	return &result, nil
}

// formatGeoIPResult formats the geoip response for display
func formatGeoIPResult(r *GeoIPResponse) string {
	var sb strings.Builder

	// Location line
	location := r.City
	if r.RegionName != "" && r.RegionName != r.City {
		location += ", " + r.RegionName
	}
	if r.Country != "" {
		location += ", " + r.Country
	}
	sb.WriteString(fmt.Sprintf("\n> Location: %s", location))

	// ISP/Org
	if r.ISP != "" {
		sb.WriteString(fmt.Sprintf("\n> ISP: %s", r.ISP))
	} else if r.Org != "" {
		sb.WriteString(fmt.Sprintf("\n> Org: %s", r.Org))
	}

	// Coordinates
	sb.WriteString(fmt.Sprintf("\n> Coords: %.4f, %.4f", r.Lat, r.Lon))

	// Timezone
	if r.Timezone != "" {
		sb.WriteString(fmt.Sprintf("\n> Timezone: %s", r.Timezone))
	}

	return sb.String()
}

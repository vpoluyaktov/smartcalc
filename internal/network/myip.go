package network

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// MyIPResponse represents the response from ip-api.com for current IP
type MyIPResponse struct {
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

// IsMyIPExpression checks if an expression is asking for the user's IP
func IsMyIPExpression(expr string) bool {
	expr = strings.TrimSpace(strings.ToLower(expr))
	patterns := []string{
		`^what\s+is\s+my\s+ip$`,
		`^what'?s\s+my\s+ip$`,
		`^my\s+ip$`,
		`^my\s+ip\s+address$`,
		`^show\s+my\s+ip$`,
		`^get\s+my\s+ip$`,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, expr); matched {
			return true
		}
	}

	return false
}

// EvalMyIP returns the user's public IP address with location info
func EvalMyIP() (string, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// ip-api.com returns info about the requesting IP when no IP is specified
	resp, err := client.Get("http://ip-api.com/json/")
	if err != nil {
		return "", fmt.Errorf("failed to get IP info: %v", err)
	}
	defer resp.Body.Close()

	var result MyIPResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if result.Status != "success" {
		return "", fmt.Errorf("API error: %s", result.Message)
	}

	// Format the response
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n> IP: %s", result.Query))
	sb.WriteString(fmt.Sprintf("\n> Location: %s, %s, %s", result.City, result.RegionName, result.Country))
	if result.Zip != "" {
		sb.WriteString(fmt.Sprintf(" %s", result.Zip))
	}
	sb.WriteString(fmt.Sprintf("\n> ISP: %s", result.ISP))
	if result.Org != "" && result.Org != result.ISP {
		sb.WriteString(fmt.Sprintf("\n> Org: %s", result.Org))
	}
	sb.WriteString(fmt.Sprintf("\n> Coordinates: %.4f, %.4f", result.Lat, result.Lon))
	sb.WriteString(fmt.Sprintf("\n> Timezone: %s", result.Timezone))

	return sb.String(), nil
}

package network

import (
	"bufio"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
)

// IsWhoisExpression checks if an expression is a whois expression
func IsWhoisExpression(expr string) bool {
	exprLower := strings.ToLower(strings.TrimSpace(expr))
	return strings.HasPrefix(exprLower, "whois ")
}

// EvalWhois evaluates a whois expression and returns the result
func EvalWhois(expr string) (string, error) {
	expr = strings.TrimSpace(expr)
	exprLower := strings.ToLower(expr)

	if !strings.HasPrefix(exprLower, "whois ") {
		return "", fmt.Errorf("invalid whois expression")
	}

	domain := strings.TrimSpace(expr[6:])

	// Remove quotes if present
	domain = strings.Trim(domain, "\"'")

	// Remove protocol if present
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")

	// Remove path if present
	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}

	if domain == "" {
		return "", fmt.Errorf("no domain specified")
	}

	return queryWhois(domain)
}

// queryWhois queries the whois server for domain information
func queryWhois(domain string) (string, error) {
	// Determine the appropriate whois server based on TLD
	whoisServer := getWhoisServer(domain)

	// Connect to whois server
	conn, err := net.DialTimeout("tcp", whoisServer+":43", 10*time.Second)
	if err != nil {
		return "", fmt.Errorf("failed to connect to whois server: %v", err)
	}
	defer conn.Close()

	// Set read/write deadline
	conn.SetDeadline(time.Now().Add(10 * time.Second))

	// Send query
	_, err = conn.Write([]byte(domain + "\r\n"))
	if err != nil {
		return "", fmt.Errorf("failed to send whois query: %v", err)
	}

	// Read response
	var response strings.Builder
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		response.WriteString(scanner.Text() + "\n")
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read whois response: %v", err)
	}

	rawResponse := response.String()
	if rawResponse == "" {
		return "", fmt.Errorf("empty response from whois server")
	}

	// Parse and format the response
	return formatWhoisResponse(domain, rawResponse), nil
}

// getWhoisServer returns the appropriate whois server for a domain
func getWhoisServer(domain string) string {
	// Extract TLD
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return "whois.iana.org"
	}

	tld := strings.ToLower(parts[len(parts)-1])

	// Common TLD whois servers
	servers := map[string]string{
		"com":  "whois.verisign-grs.com",
		"net":  "whois.verisign-grs.com",
		"org":  "whois.pir.org",
		"info": "whois.afilias.net",
		"io":   "whois.nic.io",
		"co":   "whois.nic.co",
		"me":   "whois.nic.me",
		"us":   "whois.nic.us",
		"uk":   "whois.nic.uk",
		"de":   "whois.denic.de",
		"fr":   "whois.nic.fr",
		"nl":   "whois.domain-registry.nl",
		"eu":   "whois.eu",
		"ru":   "whois.tcinet.ru",
		"cn":   "whois.cnnic.cn",
		"jp":   "whois.jprs.jp",
		"au":   "whois.auda.org.au",
		"ca":   "whois.cira.ca",
		"br":   "whois.registro.br",
		"in":   "whois.registry.in",
		"edu":  "whois.educause.edu",
		"gov":  "whois.dotgov.gov",
		"mil":  "whois.nic.mil",
		"biz":  "whois.biz",
		"name": "whois.nic.name",
		"tv":   "whois.nic.tv",
		"cc":   "ccwhois.verisign-grs.com",
		"app":  "whois.nic.google",
		"dev":  "whois.nic.google",
	}

	if server, ok := servers[tld]; ok {
		return server
	}

	return "whois.iana.org"
}

// formatWhoisResponse extracts and formats key information from whois response
func formatWhoisResponse(domain, rawResponse string) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("> WHOIS: %s\n", domain))

	// Key fields to extract
	fields := []struct {
		label    string
		patterns []string
	}{
		{"Registrar", []string{`(?i)Registrar:\s*(.+)`, `(?i)Registrar Name:\s*(.+)`}},
		{"Created", []string{`(?i)Creation Date:\s*(.+)`, `(?i)Created:\s*(.+)`, `(?i)Registration Date:\s*(.+)`}},
		{"Updated", []string{`(?i)Updated Date:\s*(.+)`, `(?i)Last Updated:\s*(.+)`}},
		{"Expires", []string{`(?i)Expir(?:y|ation) Date:\s*(.+)`, `(?i)Registry Expiry Date:\s*(.+)`}},
		{"Status", []string{`(?i)Domain Status:\s*(.+)`}},
		{"Name Servers", []string{`(?i)Name Server:\s*(.+)`}},
	}

	foundFields := make(map[string][]string)

	for _, field := range fields {
		for _, pattern := range field.patterns {
			re := regexp.MustCompile(pattern)
			matches := re.FindAllStringSubmatch(rawResponse, -1)
			for _, match := range matches {
				if len(match) > 1 {
					value := strings.TrimSpace(match[1])
					// Clean up the value
					if idx := strings.Index(value, " "); field.label == "Status" && idx != -1 {
						value = value[:idx] // Remove extra status info URLs
					}
					foundFields[field.label] = append(foundFields[field.label], value)
				}
			}
			if len(foundFields[field.label]) > 0 {
				break // Found matches with this pattern, no need to try others
			}
		}
	}

	// Output found fields
	for _, field := range fields {
		if values, ok := foundFields[field.label]; ok && len(values) > 0 {
			if field.label == "Name Servers" || field.label == "Status" {
				// Show multiple values
				result.WriteString(fmt.Sprintf("> %s:\n", field.label))
				seen := make(map[string]bool)
				for _, v := range values {
					vLower := strings.ToLower(v)
					if !seen[vLower] {
						seen[vLower] = true
						result.WriteString(fmt.Sprintf(">   %s\n", v))
					}
				}
			} else {
				// Show first value only
				result.WriteString(fmt.Sprintf("> %s: %s\n", field.label, values[0]))
			}
		}
	}

	output := result.String()
	if output == fmt.Sprintf("> WHOIS: %s\n", domain) {
		// No fields found, return raw response (truncated)
		lines := strings.Split(rawResponse, "\n")
		result.WriteString("> Raw Response:\n")
		count := 0
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "%") && !strings.HasPrefix(line, "#") {
				result.WriteString(fmt.Sprintf(">   %s\n", line))
				count++
				if count >= 20 {
					result.WriteString(">   ... (truncated)\n")
					break
				}
			}
		}
		return strings.TrimSuffix(result.String(), "\n")
	}

	return strings.TrimSuffix(output, "\n")
}

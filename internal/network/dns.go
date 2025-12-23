package network

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

// IsDNSExpression checks if an expression is a DNS lookup expression
func IsDNSExpression(expr string) bool {
	exprLower := strings.ToLower(strings.TrimSpace(expr))

	patterns := []string{
		`^dig\s+`,      // dig <domain>
		`^nslookup\s+`, // nslookup <domain>
		`^dns\s+`,      // dns <domain>
		`^lookup\s+`,   // lookup <domain>
		`^resolve\s+`,  // resolve <domain>
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, exprLower); matched {
			return true
		}
	}

	return false
}

// EvalDNS evaluates a DNS lookup expression and returns the result
func EvalDNS(expr string) (string, error) {
	expr = strings.TrimSpace(expr)
	exprLower := strings.ToLower(expr)

	var domain string

	// Extract domain from different formats
	switch {
	case strings.HasPrefix(exprLower, "dig "):
		domain = strings.TrimSpace(expr[4:])
	case strings.HasPrefix(exprLower, "nslookup "):
		domain = strings.TrimSpace(expr[9:])
	case strings.HasPrefix(exprLower, "dns "):
		domain = strings.TrimSpace(expr[4:])
	case strings.HasPrefix(exprLower, "lookup "):
		domain = strings.TrimSpace(expr[7:])
	case strings.HasPrefix(exprLower, "resolve "):
		domain = strings.TrimSpace(expr[8:])
	default:
		return "", fmt.Errorf("invalid DNS expression")
	}

	// Remove quotes if present
	domain = strings.Trim(domain, "\"'")

	// Remove trailing dot if present
	domain = strings.TrimSuffix(domain, ".")

	if domain == "" {
		return "", fmt.Errorf("no domain specified")
	}

	return lookupDomain(domain)
}

// lookupDomain performs DNS lookups for a domain
func lookupDomain(domain string) (string, error) {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("> DNS Lookup: %s\n", domain))

	// Follow CNAME chain and collect all records
	cnameChain := followCNAMEChain(domain)

	if len(cnameChain) > 1 {
		result.WriteString("> Resolution Chain:\n")
		// Find max name length for alignment
		maxLen := 0
		for _, entry := range cnameChain {
			if len(entry.name) > maxLen {
				maxLen = len(entry.name)
			}
		}
		for i, entry := range cnameChain {
			if i < len(cnameChain)-1 {
				// CNAME entry
				result.WriteString(fmt.Sprintf(">   %-*s  CNAME  %s\n", maxLen, entry.name, entry.target))
			} else {
				// Final A/AAAA records
				if len(entry.ips) > 0 {
					for _, ip := range entry.ips {
						result.WriteString(fmt.Sprintf(">   %-*s  A      %s\n", maxLen, entry.name, ip))
					}
				}
			}
		}
	} else {
		// Direct A records (no CNAME chain)
		ips, err := net.LookupIP(domain)
		if err == nil {
			var ipv4s, ipv6s []string
			for _, ip := range ips {
				if ip.To4() != nil {
					ipv4s = append(ipv4s, ip.String())
				} else {
					ipv6s = append(ipv6s, ip.String())
				}
			}
			if len(ipv4s) > 0 {
				result.WriteString("> A Records:\n")
				for _, ip := range ipv4s {
					result.WriteString(fmt.Sprintf(">   %s\n", ip))
				}
			}
			if len(ipv6s) > 0 {
				result.WriteString("> AAAA Records:\n")
				for _, ip := range ipv6s {
					result.WriteString(fmt.Sprintf(">   %s\n", ip))
				}
			}
		}
	}

	// MX records
	mxs, err := net.LookupMX(domain)
	if err == nil && len(mxs) > 0 {
		result.WriteString("> MX Records:\n")
		for _, mx := range mxs {
			result.WriteString(fmt.Sprintf(">   %s (priority: %d)\n", strings.TrimSuffix(mx.Host, "."), mx.Pref))
		}
	}

	// NS records
	nss, err := net.LookupNS(domain)
	if err == nil && len(nss) > 0 {
		result.WriteString("> NS Records:\n")
		for _, ns := range nss {
			result.WriteString(fmt.Sprintf(">   %s\n", strings.TrimSuffix(ns.Host, ".")))
		}
	}

	// TXT records
	txts, err := net.LookupTXT(domain)
	if err == nil && len(txts) > 0 {
		result.WriteString("> TXT Records:\n")
		for _, txt := range txts {
			// Truncate long TXT records
			if len(txt) > 80 {
				txt = txt[:77] + "..."
			}
			result.WriteString(fmt.Sprintf(">   \"%s\"\n", txt))
		}
	}

	output := result.String()
	if output == fmt.Sprintf("> DNS Lookup: %s\n", domain) {
		return "", fmt.Errorf("no DNS records found for %s", domain)
	}

	return strings.TrimSuffix(output, "\n"), nil
}

// cnameEntry represents a step in the CNAME resolution chain
type cnameEntry struct {
	name   string
	target string
	ips    []string
}

// followCNAMEChain follows CNAME records until it reaches A/AAAA records
func followCNAMEChain(domain string) []cnameEntry {
	var chain []cnameEntry
	seen := make(map[string]bool)
	current := domain
	maxDepth := 10 // Prevent infinite loops

	for i := 0; i < maxDepth; i++ {
		if seen[current] {
			break // Circular reference
		}
		seen[current] = true

		// Look up CNAME for current domain
		cname, err := net.LookupCNAME(current)
		if err != nil {
			break
		}

		cname = strings.TrimSuffix(cname, ".")

		if cname == current || cname == "" {
			// No CNAME, this is the final domain - get A records
			ips, err := net.LookupIP(current)
			if err == nil && len(ips) > 0 {
				var ipStrs []string
				for _, ip := range ips {
					if ip.To4() != nil {
						ipStrs = append(ipStrs, ip.String())
					}
				}
				chain = append(chain, cnameEntry{name: current, ips: ipStrs})
			}
			break
		}

		// Add CNAME to chain
		chain = append(chain, cnameEntry{name: current, target: cname})
		current = cname
	}

	// If we followed CNAMEs, get the final A records
	if len(chain) > 0 && len(chain[len(chain)-1].ips) == 0 {
		ips, err := net.LookupIP(current)
		if err == nil && len(ips) > 0 {
			var ipStrs []string
			for _, ip := range ips {
				if ip.To4() != nil {
					ipStrs = append(ipStrs, ip.String())
				}
			}
			chain = append(chain, cnameEntry{name: current, ips: ipStrs})
		}
	}

	return chain
}

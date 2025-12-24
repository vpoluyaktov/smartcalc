package network

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// DNS-over-HTTPS servers (bypasses network-level DNS interception)
var dohServers = []string{
	"https://cloudflare-dns.com/dns-query",
	"https://dns.google/dns-query",
	"https://dns.quad9.net/dns-query",
}

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

// queryDNS sends a DNS query using DNS-over-HTTPS to bypass network interception
func queryDNS(domain string, qtype uint16) (*dns.Msg, error) {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), qtype)
	m.RecursionDesired = true

	// Pack the DNS message
	dnsData, err := m.Pack()
	if err != nil {
		return nil, fmt.Errorf("failed to pack DNS message: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	var lastErr error
	for _, server := range dohServers {
		req, err := http.NewRequest("POST", server, bytes.NewReader(dnsData))
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("Content-Type", "application/dns-message")
		req.Header.Set("Accept", "application/dns-message")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("DoH server returned status %d", resp.StatusCode)
			continue
		}

		r := new(dns.Msg)
		if err := r.Unpack(body); err != nil {
			lastErr = err
			continue
		}

		if r.Rcode != dns.RcodeSuccess {
			lastErr = fmt.Errorf("DNS query failed with rcode: %d", r.Rcode)
			continue
		}

		return r, nil
	}
	return nil, lastErr
}

// lookupDomain performs DNS lookups for a domain using public DNS servers
func lookupDomain(domain string) (string, error) {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("> DNS Lookup: %s\n", domain))

	// Follow CNAME chain and collect all records using public DNS
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
		// Direct A records (no CNAME chain) - use public DNS
		ipv4s, ipv6s := lookupIPsPublicDNS(domain)
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

	// MX records using public DNS
	mxRecords := lookupMXPublicDNS(domain)
	if len(mxRecords) > 0 {
		result.WriteString("> MX Records:\n")
		for _, mx := range mxRecords {
			result.WriteString(fmt.Sprintf(">   %s (priority: %d)\n", mx.host, mx.pref))
		}
	}

	// NS records using public DNS
	nsRecords := lookupNSPublicDNS(domain)
	if len(nsRecords) > 0 {
		result.WriteString("> NS Records:\n")
		for _, ns := range nsRecords {
			result.WriteString(fmt.Sprintf(">   %s\n", ns))
		}
	}

	// TXT records using public DNS
	txtRecords := lookupTXTPublicDNS(domain)
	if len(txtRecords) > 0 {
		result.WriteString("> TXT Records:\n")
		for _, txt := range txtRecords {
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

// lookupIPsPublicDNS queries A and AAAA records using public DNS
func lookupIPsPublicDNS(domain string) (ipv4s, ipv6s []string) {
	// Query A records
	r, err := queryDNS(domain, dns.TypeA)
	if err == nil {
		for _, ans := range r.Answer {
			if a, ok := ans.(*dns.A); ok {
				ipv4s = append(ipv4s, a.A.String())
			}
		}
	}

	// Query AAAA records
	r, err = queryDNS(domain, dns.TypeAAAA)
	if err == nil {
		for _, ans := range r.Answer {
			if aaaa, ok := ans.(*dns.AAAA); ok {
				ipv6s = append(ipv6s, aaaa.AAAA.String())
			}
		}
	}

	return ipv4s, ipv6s
}

type mxRecord struct {
	host string
	pref uint16
}

// lookupMXPublicDNS queries MX records using public DNS
func lookupMXPublicDNS(domain string) []mxRecord {
	var records []mxRecord
	r, err := queryDNS(domain, dns.TypeMX)
	if err == nil {
		for _, ans := range r.Answer {
			if mx, ok := ans.(*dns.MX); ok {
				records = append(records, mxRecord{
					host: strings.TrimSuffix(mx.Mx, "."),
					pref: mx.Preference,
				})
			}
		}
	}
	return records
}

// lookupNSPublicDNS queries NS records using public DNS
func lookupNSPublicDNS(domain string) []string {
	var records []string
	r, err := queryDNS(domain, dns.TypeNS)
	if err == nil {
		for _, ans := range r.Answer {
			if ns, ok := ans.(*dns.NS); ok {
				records = append(records, strings.TrimSuffix(ns.Ns, "."))
			}
		}
	}
	return records
}

// lookupTXTPublicDNS queries TXT records using public DNS
func lookupTXTPublicDNS(domain string) []string {
	var records []string
	r, err := queryDNS(domain, dns.TypeTXT)
	if err == nil {
		for _, ans := range r.Answer {
			if txt, ok := ans.(*dns.TXT); ok {
				records = append(records, strings.Join(txt.Txt, ""))
			}
		}
	}
	return records
}

// cnameEntry represents a step in the CNAME resolution chain
type cnameEntry struct {
	name   string
	target string
	ips    []string
}

// lookupCNAMEPublicDNS queries CNAME records using public DNS
func lookupCNAMEPublicDNS(domain string) (string, error) {
	r, err := queryDNS(domain, dns.TypeCNAME)
	if err != nil {
		return "", err
	}
	for _, ans := range r.Answer {
		if cname, ok := ans.(*dns.CNAME); ok {
			return strings.TrimSuffix(cname.Target, "."), nil
		}
	}
	return "", fmt.Errorf("no CNAME record found")
}

// followCNAMEChain follows CNAME records until it reaches A/AAAA records using public DNS
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

		// Look up CNAME for current domain using public DNS
		cname, err := lookupCNAMEPublicDNS(current)
		if err != nil || cname == "" || cname == current {
			// No CNAME, this is the final domain - get A records
			ipv4s, _ := lookupIPsPublicDNS(current)
			if len(ipv4s) > 0 {
				chain = append(chain, cnameEntry{name: current, ips: ipv4s})
			}
			break
		}

		// Add CNAME to chain
		chain = append(chain, cnameEntry{name: current, target: cname})
		current = cname
	}

	// If we followed CNAMEs, get the final A records
	if len(chain) > 0 && len(chain[len(chain)-1].ips) == 0 {
		ipv4s, _ := lookupIPsPublicDNS(current)
		if len(ipv4s) > 0 {
			chain = append(chain, cnameEntry{name: current, ips: ipv4s})
		}
	}

	return chain
}

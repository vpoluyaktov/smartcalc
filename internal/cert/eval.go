package cert

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// IsCertExpression checks if an expression is a certificate decode expression
func IsCertExpression(expr string) bool {
	exprLower := strings.ToLower(strings.TrimSpace(expr))

	patterns := []string{
		`^cert\s+decode\s+`, // cert decode <url>
		`^ssl\s+decode\s+`,  // ssl decode <url>
		`^cert\s+test\s+`,   // cert test <url>
		`^ssl\s+test\s+`,    // ssl test <url>
		`^cert\s+https?://`, // cert <url>
		`^ssl\s+https?://`,  // ssl <url>
		`^decode\s+cert\s+`, // decode cert <url>
		`^decode\s+ssl\s+`,  // decode ssl <url>
		`^test\s+cert\s+`,   // test cert <url>
		`^test\s+ssl\s+`,    // test ssl <url>
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, exprLower); matched {
			return true
		}
	}

	return false
}

// EvalCert evaluates a certificate expression and returns the decoded result
func EvalCert(expr string) (string, error) {
	expr = strings.TrimSpace(expr)
	exprLower := strings.ToLower(expr)

	var urlStr string

	// Extract URL from different formats
	if strings.HasPrefix(exprLower, "cert decode ") {
		urlStr = strings.TrimSpace(expr[12:])
	} else if strings.HasPrefix(exprLower, "ssl decode ") {
		urlStr = strings.TrimSpace(expr[11:])
	} else if strings.HasPrefix(exprLower, "cert test ") {
		urlStr = strings.TrimSpace(expr[10:])
	} else if strings.HasPrefix(exprLower, "ssl test ") {
		urlStr = strings.TrimSpace(expr[9:])
	} else if strings.HasPrefix(exprLower, "decode cert ") {
		urlStr = strings.TrimSpace(expr[12:])
	} else if strings.HasPrefix(exprLower, "decode ssl ") {
		urlStr = strings.TrimSpace(expr[11:])
	} else if strings.HasPrefix(exprLower, "test cert ") {
		urlStr = strings.TrimSpace(expr[10:])
	} else if strings.HasPrefix(exprLower, "test ssl ") {
		urlStr = strings.TrimSpace(expr[9:])
	} else if strings.HasPrefix(exprLower, "cert ") {
		urlStr = strings.TrimSpace(expr[5:])
	} else if strings.HasPrefix(exprLower, "ssl ") {
		urlStr = strings.TrimSpace(expr[4:])
	} else {
		return "", fmt.Errorf("invalid certificate expression")
	}

	// Remove quotes if present
	urlStr = strings.Trim(urlStr, `"'`)

	// Add https:// if no scheme provided
	if !strings.HasPrefix(strings.ToLower(urlStr), "http://") && !strings.HasPrefix(strings.ToLower(urlStr), "https://") {
		urlStr = "https://" + urlStr
	}

	return decodeCertificate(urlStr)
}

// decodeCertificate fetches and decodes the SSL certificate from the given URL
func decodeCertificate(urlStr string) (string, error) {
	// Parse the URL to get the host
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %v", err)
	}

	host := parsedURL.Host
	if host == "" {
		return "", fmt.Errorf("invalid URL: no host specified")
	}

	// Add default port if not specified
	if !strings.Contains(host, ":") {
		if parsedURL.Scheme == "http" {
			host += ":80"
		} else {
			host += ":443"
		}
	}

	// Connect with TLS, skipping verification to handle expired/untrusted certs
	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 10 * time.Second},
		"tcp",
		host,
		&tls.Config{
			InsecureSkipVerify: true, // Allow expired/untrusted certificates
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to connect: %v", err)
	}
	defer conn.Close()

	// Get the certificate chain
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return "", fmt.Errorf("no certificates found")
	}

	// Format the certificate information
	return formatCertificates(certs, host)
}

// formatCertificates formats the certificate chain for display
func formatCertificates(certs []*x509.Certificate, host string) (string, error) {
	var result strings.Builder

	// Show the leaf certificate (first in chain) in detail
	cert := certs[0]

	// Subject
	result.WriteString("Subject:\n")
	if cert.Subject.CommonName != "" {
		result.WriteString(fmt.Sprintf(">   Common Name: %s\n", cert.Subject.CommonName))
	}
	if len(cert.Subject.Organization) > 0 {
		result.WriteString(fmt.Sprintf(">   Organization: %s\n", strings.Join(cert.Subject.Organization, ", ")))
	}
	if len(cert.Subject.OrganizationalUnit) > 0 {
		result.WriteString(fmt.Sprintf(">   Org Unit: %s\n", strings.Join(cert.Subject.OrganizationalUnit, ", ")))
	}
	if len(cert.Subject.Country) > 0 {
		result.WriteString(fmt.Sprintf(">   Country: %s\n", strings.Join(cert.Subject.Country, ", ")))
	}
	if len(cert.Subject.Province) > 0 {
		result.WriteString(fmt.Sprintf(">   State/Province: %s\n", strings.Join(cert.Subject.Province, ", ")))
	}
	if len(cert.Subject.Locality) > 0 {
		result.WriteString(fmt.Sprintf(">   Locality: %s\n", strings.Join(cert.Subject.Locality, ", ")))
	}

	// Issuer
	result.WriteString("> Issuer:\n")
	if cert.Issuer.CommonName != "" {
		result.WriteString(fmt.Sprintf(">   Common Name: %s\n", cert.Issuer.CommonName))
	}
	if len(cert.Issuer.Organization) > 0 {
		result.WriteString(fmt.Sprintf(">   Organization: %s\n", strings.Join(cert.Issuer.Organization, ", ")))
	}
	if len(cert.Issuer.Country) > 0 {
		result.WriteString(fmt.Sprintf(">   Country: %s\n", strings.Join(cert.Issuer.Country, ", ")))
	}

	// Validity
	result.WriteString("> Validity:\n")
	result.WriteString(fmt.Sprintf(">   Not Before: %s\n", cert.NotBefore.Format("2006-01-02 15:04:05 MST")))
	result.WriteString(fmt.Sprintf(">   Not After: %s\n", cert.NotAfter.Format("2006-01-02 15:04:05 MST")))

	// Status
	now := time.Now()
	result.WriteString("> Status: ")
	if now.Before(cert.NotBefore) {
		result.WriteString("⚠ NOT YET VALID\n")
	} else if now.After(cert.NotAfter) {
		result.WriteString("⚠ EXPIRED\n")
	} else {
		remaining := time.Until(cert.NotAfter)
		result.WriteString(fmt.Sprintf("✓ Valid (expires in %s)\n", formatDuration(remaining)))
	}

	// Serial Number
	result.WriteString(fmt.Sprintf("> Serial Number: %s\n", formatSerialNumber(cert.SerialNumber.Bytes())))

	// Signature Algorithm
	result.WriteString(fmt.Sprintf("> Signature Algorithm: %s\n", cert.SignatureAlgorithm.String()))

	// Public Key Info
	result.WriteString("> Public Key:\n")
	result.WriteString(fmt.Sprintf(">   Algorithm: %s\n", cert.PublicKeyAlgorithm.String()))
	if keySize := getKeySize(cert); keySize > 0 {
		result.WriteString(fmt.Sprintf(">   Key Size: %d bits\n", keySize))
	}

	// Subject Alternative Names (SANs)
	if len(cert.DNSNames) > 0 || len(cert.IPAddresses) > 0 {
		result.WriteString("> Subject Alt Names:\n")
		for _, dns := range cert.DNSNames {
			result.WriteString(fmt.Sprintf(">   DNS: %s\n", dns))
		}
		for _, ip := range cert.IPAddresses {
			result.WriteString(fmt.Sprintf(">   IP: %s\n", ip.String()))
		}
	}

	// Key Usage
	if cert.KeyUsage != 0 {
		result.WriteString("> Key Usage: ")
		result.WriteString(formatKeyUsage(cert.KeyUsage))
		result.WriteString("\n")
	}

	// Extended Key Usage
	if len(cert.ExtKeyUsage) > 0 {
		result.WriteString("> Extended Key Usage: ")
		result.WriteString(formatExtKeyUsage(cert.ExtKeyUsage))
		result.WriteString("\n")
	}

	// Certificate chain info as ASCII tree
	if len(certs) > 1 {
		result.WriteString(fmt.Sprintf("> Certificate Chain: %d certificates\n", len(certs)))
		// Display chain in reverse order (root at top, leaf at bottom)
		for i := len(certs) - 1; i >= 0; i-- {
			c := certs[i]
			depth := len(certs) - 1 - i

			// Determine certificate type label
			var label string
			if i == 0 {
				label = "(leaf)"
			} else if c.IsCA {
				if i == len(certs)-1 {
					label = "(root)"
				} else {
					label = "(intermediate)"
				}
			}

			name := c.Subject.CommonName
			if name == "" && len(c.Subject.Organization) > 0 {
				name = c.Subject.Organization[0]
			}

			// Build tree line
			if depth == 0 {
				// Root certificate
				result.WriteString(fmt.Sprintf("> 🔐 %s %s\n", name, label))
			} else {
				// All child certificates at same indentation level
				if i == 0 {
					// Last certificate (leaf)
					result.WriteString(fmt.Sprintf(">    └── %s %s\n", name, label))
				} else {
					// Intermediate certificate
					result.WriteString(fmt.Sprintf(">    ├── %s %s\n", name, label))
				}
			}
		}
	}

	return result.String(), nil
}

// formatSerialNumber formats a serial number as hex
func formatSerialNumber(bytes []byte) string {
	if len(bytes) == 0 {
		return "N/A"
	}
	var parts []string
	for _, b := range bytes {
		parts = append(parts, fmt.Sprintf("%02X", b))
	}
	return strings.Join(parts, ":")
}

// getKeySize returns the key size in bits
func getKeySize(cert *x509.Certificate) int {
	switch pub := cert.PublicKey.(type) {
	case interface{ Size() int }:
		return pub.Size() * 8
	default:
		return 0
	}
}

// formatKeyUsage formats the key usage flags
func formatKeyUsage(ku x509.KeyUsage) string {
	var usages []string
	if ku&x509.KeyUsageDigitalSignature != 0 {
		usages = append(usages, "Digital Signature")
	}
	if ku&x509.KeyUsageContentCommitment != 0 {
		usages = append(usages, "Content Commitment")
	}
	if ku&x509.KeyUsageKeyEncipherment != 0 {
		usages = append(usages, "Key Encipherment")
	}
	if ku&x509.KeyUsageDataEncipherment != 0 {
		usages = append(usages, "Data Encipherment")
	}
	if ku&x509.KeyUsageKeyAgreement != 0 {
		usages = append(usages, "Key Agreement")
	}
	if ku&x509.KeyUsageCertSign != 0 {
		usages = append(usages, "Certificate Sign")
	}
	if ku&x509.KeyUsageCRLSign != 0 {
		usages = append(usages, "CRL Sign")
	}
	if ku&x509.KeyUsageEncipherOnly != 0 {
		usages = append(usages, "Encipher Only")
	}
	if ku&x509.KeyUsageDecipherOnly != 0 {
		usages = append(usages, "Decipher Only")
	}
	if len(usages) == 0 {
		return "None"
	}
	return strings.Join(usages, ", ")
}

// formatExtKeyUsage formats the extended key usage
func formatExtKeyUsage(eku []x509.ExtKeyUsage) string {
	var usages []string
	for _, u := range eku {
		switch u {
		case x509.ExtKeyUsageAny:
			usages = append(usages, "Any")
		case x509.ExtKeyUsageServerAuth:
			usages = append(usages, "Server Authentication")
		case x509.ExtKeyUsageClientAuth:
			usages = append(usages, "Client Authentication")
		case x509.ExtKeyUsageCodeSigning:
			usages = append(usages, "Code Signing")
		case x509.ExtKeyUsageEmailProtection:
			usages = append(usages, "Email Protection")
		case x509.ExtKeyUsageTimeStamping:
			usages = append(usages, "Time Stamping")
		case x509.ExtKeyUsageOCSPSigning:
			usages = append(usages, "OCSP Signing")
		default:
			usages = append(usages, fmt.Sprintf("Unknown(%d)", u))
		}
	}
	if len(usages) == 0 {
		return "None"
	}
	return strings.Join(usages, ", ")
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < 0 {
		return "expired"
	}

	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 365 {
		years := days / 365
		remainingDays := days % 365
		return fmt.Sprintf("%dy %dd", years, remainingDays)
	}
	if days > 30 {
		months := days / 30
		remainingDays := days % 30
		return fmt.Sprintf("%dmo %dd", months, remainingDays)
	}
	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
}

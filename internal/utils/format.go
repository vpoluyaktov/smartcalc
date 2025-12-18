package utils

import (
	"fmt"
	"math"
	"strings"
)

func addThousandsSeparators(s string) string {
	if s == "" {
		return s
	}
	sign := ""
	if strings.HasPrefix(s, "-") {
		sign = "-"
		s = strings.TrimPrefix(s, "-")
	}
	n := len(s)
	if n <= 3 {
		return sign + s
	}
	rem := n % 3
	if rem == 0 {
		rem = 3
	}
	var b strings.Builder
	b.Grow(n + (n / 3))
	b.WriteString(sign)
	b.WriteString(s[:rem])
	for i := rem; i < n; i += 3 {
		b.WriteByte(',')
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

func formatNumberWithThousands(v float64) string {
	s := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.10f", v), "0"), ".")
	intPart := s
	fracPart := ""
	if dot := strings.IndexByte(s, '.'); dot >= 0 {
		intPart = s[:dot]
		fracPart = s[dot:]
	}
	return addThousandsSeparators(intPart) + fracPart
}

func formatCurrency(v float64) string {
	abs := math.Abs(v)
	whole := int64(abs)
	frac := int64(math.Round((abs - float64(whole)) * 100))
	if frac == 100 {
		whole++
		frac = 0
	}
	out := fmt.Sprintf("%s.%02d", addThousandsSeparators(fmt.Sprintf("%d", whole)), frac)
	if v < 0 {
		out = "-" + out
	}
	return "$" + out
}

func FormatResult(isCurrency bool, v float64) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return "NaN"
	}
	if isCurrency {
		return formatCurrency(v)
	}
	return formatNumberWithThousands(v)
}

// FormatBoolResult formats a comparison result as true/false
func FormatBoolResult(v float64) string {
	if v == 1 {
		return "true"
	}
	return "false"
}

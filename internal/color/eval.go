package color

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// IsColorExpression checks if an expression is a color conversion
func IsColorExpression(expr string) bool {
	expr = strings.TrimSpace(strings.ToLower(expr))

	patterns := []string{
		// Hex to RGB/HSL
		`^#[0-9a-f]{6}\s+(?:to|in)\s+(?:rgb|hsl)$`,
		`^#[0-9a-f]{3}\s+(?:to|in)\s+(?:rgb|hsl)$`,
		// RGB to Hex/HSL
		`^rgb\s*\(\s*\d+\s*,\s*\d+\s*,\s*\d+\s*\)\s+(?:to|in)\s+(?:hex|hsl)$`,
		// HSL to RGB/Hex
		`^hsl\s*\(\s*\d+\s*,\s*\d+%?\s*,\s*\d+%?\s*\)\s+(?:to|in)\s+(?:rgb|hex)$`,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, expr); matched {
			return true
		}
	}

	return false
}

// EvalColor evaluates a color conversion expression
func EvalColor(expr string) (string, error) {
	expr = strings.TrimSpace(expr)
	exprLower := strings.ToLower(expr)

	// Parse the expression to get source color and target format
	parts := regexp.MustCompile(`\s+(?:to|in)\s+`).Split(exprLower, 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid color expression")
	}

	sourceColor := strings.TrimSpace(parts[0])
	targetFormat := strings.TrimSpace(parts[1])

	// Determine source format and convert
	if strings.HasPrefix(sourceColor, "#") {
		return convertFromHex(sourceColor, targetFormat)
	} else if strings.HasPrefix(sourceColor, "rgb") {
		return convertFromRGB(sourceColor, targetFormat)
	} else if strings.HasPrefix(sourceColor, "hsl") {
		return convertFromHSL(sourceColor, targetFormat)
	}

	return "", fmt.Errorf("unknown color format: %s", sourceColor)
}

// convertFromHex converts a hex color to the target format
func convertFromHex(hex string, target string) (string, error) {
	r, g, b, err := parseHex(hex)
	if err != nil {
		return "", err
	}

	switch target {
	case "rgb":
		return fmt.Sprintf("rgb(%d, %d, %d)", r, g, b), nil
	case "hsl":
		h, s, l := rgbToHSL(r, g, b)
		return fmt.Sprintf("hsl(%d, %d%%, %d%%)", h, s, l), nil
	default:
		return "", fmt.Errorf("unknown target format: %s", target)
	}
}

// convertFromRGB converts an RGB color to the target format
func convertFromRGB(rgb string, target string) (string, error) {
	r, g, b, err := parseRGB(rgb)
	if err != nil {
		return "", err
	}

	switch target {
	case "hex":
		return fmt.Sprintf("#%02X%02X%02X", r, g, b), nil
	case "hsl":
		h, s, l := rgbToHSL(r, g, b)
		return fmt.Sprintf("hsl(%d, %d%%, %d%%)", h, s, l), nil
	default:
		return "", fmt.Errorf("unknown target format: %s", target)
	}
}

// convertFromHSL converts an HSL color to the target format
func convertFromHSL(hsl string, target string) (string, error) {
	h, s, l, err := parseHSL(hsl)
	if err != nil {
		return "", err
	}

	r, g, b := hslToRGB(h, s, l)

	switch target {
	case "rgb":
		return fmt.Sprintf("rgb(%d, %d, %d)", r, g, b), nil
	case "hex":
		return fmt.Sprintf("#%02X%02X%02X", r, g, b), nil
	default:
		return "", fmt.Errorf("unknown target format: %s", target)
	}
}

// parseHex parses a hex color string and returns RGB values
func parseHex(hex string) (int, int, int, error) {
	hex = strings.TrimPrefix(hex, "#")

	// Handle shorthand hex (e.g., #FFF -> #FFFFFF)
	if len(hex) == 3 {
		hex = string(hex[0]) + string(hex[0]) +
			string(hex[1]) + string(hex[1]) +
			string(hex[2]) + string(hex[2])
	}

	if len(hex) != 6 {
		return 0, 0, 0, fmt.Errorf("invalid hex color: #%s", hex)
	}

	r, err := strconv.ParseInt(hex[0:2], 16, 64)
	if err != nil {
		return 0, 0, 0, err
	}
	g, err := strconv.ParseInt(hex[2:4], 16, 64)
	if err != nil {
		return 0, 0, 0, err
	}
	b, err := strconv.ParseInt(hex[4:6], 16, 64)
	if err != nil {
		return 0, 0, 0, err
	}

	return int(r), int(g), int(b), nil
}

// parseRGB parses an RGB color string and returns RGB values
func parseRGB(rgb string) (int, int, int, error) {
	re := regexp.MustCompile(`rgb\s*\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)\s*\)`)
	matches := re.FindStringSubmatch(strings.ToLower(rgb))
	if matches == nil {
		return 0, 0, 0, fmt.Errorf("invalid RGB color: %s", rgb)
	}

	r, _ := strconv.Atoi(matches[1])
	g, _ := strconv.Atoi(matches[2])
	b, _ := strconv.Atoi(matches[3])

	// Clamp values to 0-255
	r = clamp(r, 0, 255)
	g = clamp(g, 0, 255)
	b = clamp(b, 0, 255)

	return r, g, b, nil
}

// parseHSL parses an HSL color string and returns H, S, L values
func parseHSL(hsl string) (int, int, int, error) {
	re := regexp.MustCompile(`hsl\s*\(\s*(\d+)\s*,\s*(\d+)%?\s*,\s*(\d+)%?\s*\)`)
	matches := re.FindStringSubmatch(strings.ToLower(hsl))
	if matches == nil {
		return 0, 0, 0, fmt.Errorf("invalid HSL color: %s", hsl)
	}

	h, _ := strconv.Atoi(matches[1])
	s, _ := strconv.Atoi(matches[2])
	l, _ := strconv.Atoi(matches[3])

	// Normalize values
	h = h % 360
	s = clamp(s, 0, 100)
	l = clamp(l, 0, 100)

	return h, s, l, nil
}

// rgbToHSL converts RGB values to HSL
func rgbToHSL(r, g, b int) (int, int, int) {
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))
	delta := max - min

	// Lightness
	l := (max + min) / 2.0

	if delta == 0 {
		return 0, 0, int(math.Round(l * 100))
	}

	// Saturation
	var s float64
	if l < 0.5 {
		s = delta / (max + min)
	} else {
		s = delta / (2.0 - max - min)
	}

	// Hue
	var h float64
	switch max {
	case rf:
		h = (gf - bf) / delta
		if gf < bf {
			h += 6
		}
	case gf:
		h = (bf-rf)/delta + 2
	case bf:
		h = (rf-gf)/delta + 4
	}
	h *= 60

	return int(math.Round(h)), int(math.Round(s * 100)), int(math.Round(l * 100))
}

// hslToRGB converts HSL values to RGB
func hslToRGB(h, s, l int) (int, int, int) {
	hf := float64(h) / 360.0
	sf := float64(s) / 100.0
	lf := float64(l) / 100.0

	if sf == 0 {
		v := int(math.Round(lf * 255))
		return v, v, v
	}

	var q float64
	if lf < 0.5 {
		q = lf * (1 + sf)
	} else {
		q = lf + sf - lf*sf
	}
	p := 2*lf - q

	r := hueToRGB(p, q, hf+1.0/3.0)
	g := hueToRGB(p, q, hf)
	b := hueToRGB(p, q, hf-1.0/3.0)

	return int(math.Round(r * 255)), int(math.Round(g * 255)), int(math.Round(b * 255))
}

// hueToRGB is a helper function for HSL to RGB conversion
func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

// clamp restricts a value to a range
func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

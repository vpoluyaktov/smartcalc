package color

import (
	"strings"
	"testing"
)

func TestIsColorExpression(t *testing.T) {
	tests := []struct {
		expr     string
		expected bool
	}{
		// Hex to RGB/HSL
		{"#FF5733 to rgb", true},
		{"#FF5733 in rgb", true},
		{"#FF5733 to hsl", true},
		{"#FFF to rgb", true},
		{"#abc to hsl", true},

		// RGB to Hex/HSL
		{"rgb(255, 87, 51) to hex", true},
		{"rgb(255,87,51) in hex", true},
		{"rgb(255, 87, 51) to hsl", true},

		// HSL to RGB/Hex
		{"hsl(14, 100%, 60%) to rgb", true},
		{"hsl(14, 100, 60) to hex", true},
		{"hsl(14,100%,60%) in rgb", true},

		// Invalid expressions
		{"hello world", false},
		{"100 + 50", false},
		{"#FF5733", false},
		{"rgb(255, 87, 51)", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := IsColorExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("IsColorExpression(%q) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestParseHex(t *testing.T) {
	tests := []struct {
		hex      string
		r, g, b  int
		hasError bool
	}{
		{"#FF5733", 255, 87, 51, false},
		{"#ffffff", 255, 255, 255, false},
		{"#000000", 0, 0, 0, false},
		{"#FFF", 255, 255, 255, false},
		{"#000", 0, 0, 0, false},
		{"#abc", 170, 187, 204, false},
	}

	for _, tt := range tests {
		t.Run(tt.hex, func(t *testing.T) {
			r, g, b, err := parseHex(tt.hex)
			if tt.hasError && err == nil {
				t.Errorf("parseHex(%q) expected error", tt.hex)
				return
			}
			if !tt.hasError && err != nil {
				t.Errorf("parseHex(%q) unexpected error: %v", tt.hex, err)
				return
			}
			if r != tt.r || g != tt.g || b != tt.b {
				t.Errorf("parseHex(%q) = (%d, %d, %d), want (%d, %d, %d)", tt.hex, r, g, b, tt.r, tt.g, tt.b)
			}
		})
	}
}

func TestParseRGB(t *testing.T) {
	tests := []struct {
		rgb      string
		r, g, b  int
		hasError bool
	}{
		{"rgb(255, 87, 51)", 255, 87, 51, false},
		{"rgb(0, 0, 0)", 0, 0, 0, false},
		{"rgb(255,255,255)", 255, 255, 255, false},
		{"rgb( 100 , 150 , 200 )", 100, 150, 200, false},
	}

	for _, tt := range tests {
		t.Run(tt.rgb, func(t *testing.T) {
			r, g, b, err := parseRGB(tt.rgb)
			if tt.hasError && err == nil {
				t.Errorf("parseRGB(%q) expected error", tt.rgb)
				return
			}
			if !tt.hasError && err != nil {
				t.Errorf("parseRGB(%q) unexpected error: %v", tt.rgb, err)
				return
			}
			if r != tt.r || g != tt.g || b != tt.b {
				t.Errorf("parseRGB(%q) = (%d, %d, %d), want (%d, %d, %d)", tt.rgb, r, g, b, tt.r, tt.g, tt.b)
			}
		})
	}
}

func TestParseHSL(t *testing.T) {
	tests := []struct {
		hsl      string
		h, s, l  int
		hasError bool
	}{
		{"hsl(14, 100%, 60%)", 14, 100, 60, false},
		{"hsl(0, 0%, 0%)", 0, 0, 0, false},
		{"hsl(360, 100, 50)", 0, 100, 50, false}, // 360 wraps to 0
		{"hsl( 180 , 50% , 75% )", 180, 50, 75, false},
	}

	for _, tt := range tests {
		t.Run(tt.hsl, func(t *testing.T) {
			h, s, l, err := parseHSL(tt.hsl)
			if tt.hasError && err == nil {
				t.Errorf("parseHSL(%q) expected error", tt.hsl)
				return
			}
			if !tt.hasError && err != nil {
				t.Errorf("parseHSL(%q) unexpected error: %v", tt.hsl, err)
				return
			}
			if h != tt.h || s != tt.s || l != tt.l {
				t.Errorf("parseHSL(%q) = (%d, %d, %d), want (%d, %d, %d)", tt.hsl, h, s, l, tt.h, tt.s, tt.l)
			}
		})
	}
}

func TestRGBToHSL(t *testing.T) {
	tests := []struct {
		r, g, b int
		h, s, l int
	}{
		{255, 0, 0, 0, 100, 50},    // Red
		{0, 255, 0, 120, 100, 50},  // Green
		{0, 0, 255, 240, 100, 50},  // Blue
		{255, 255, 255, 0, 0, 100}, // White
		{0, 0, 0, 0, 0, 0},         // Black
		{128, 128, 128, 0, 0, 50},  // Gray
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			h, s, l := rgbToHSL(tt.r, tt.g, tt.b)
			// Allow small rounding differences
			if abs(h-tt.h) > 1 || abs(s-tt.s) > 1 || abs(l-tt.l) > 1 {
				t.Errorf("rgbToHSL(%d, %d, %d) = (%d, %d, %d), want (%d, %d, %d)",
					tt.r, tt.g, tt.b, h, s, l, tt.h, tt.s, tt.l)
			}
		})
	}
}

func TestHSLToRGB(t *testing.T) {
	tests := []struct {
		h, s, l int
		r, g, b int
	}{
		{0, 100, 50, 255, 0, 0},    // Red
		{120, 100, 50, 0, 255, 0},  // Green
		{240, 100, 50, 0, 0, 255},  // Blue
		{0, 0, 100, 255, 255, 255}, // White
		{0, 0, 0, 0, 0, 0},         // Black
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			r, g, b := hslToRGB(tt.h, tt.s, tt.l)
			// Allow small rounding differences
			if abs(r-tt.r) > 1 || abs(g-tt.g) > 1 || abs(b-tt.b) > 1 {
				t.Errorf("hslToRGB(%d, %d, %d) = (%d, %d, %d), want (%d, %d, %d)",
					tt.h, tt.s, tt.l, r, g, b, tt.r, tt.g, tt.b)
			}
		})
	}
}

func TestEvalColor(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		// Hex to RGB
		{"#FF5733 to rgb", "rgb(255, 87, 51)"},
		{"#FFF to rgb", "rgb(255, 255, 255)"},
		{"#000 to rgb", "rgb(0, 0, 0)"},

		// Hex to HSL
		{"#FF0000 to hsl", "hsl(0, 100%, 50%)"},

		// RGB to Hex
		{"rgb(255, 87, 51) to hex", "#FF5733"},
		{"rgb(0, 0, 0) to hex", "#000000"},

		// RGB to HSL
		{"rgb(255, 0, 0) to hsl", "hsl(0, 100%, 50%)"},

		// HSL to RGB
		{"hsl(0, 100%, 50%) to rgb", "rgb(255, 0, 0)"},
		{"hsl(120, 100%, 50%) to rgb", "rgb(0, 255, 0)"},

		// HSL to Hex
		{"hsl(0, 100%, 50%) to hex", "#FF0000"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalColor(tt.expr)
			if err != nil {
				t.Errorf("EvalColor(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(strings.ToUpper(result), strings.ToUpper(tt.contains)) {
				t.Errorf("EvalColor(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

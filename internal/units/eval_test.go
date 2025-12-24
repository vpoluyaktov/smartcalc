package units

import (
	"strings"
	"testing"
)

func TestEvalLengthConversion(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"5 miles in km", "8.04"},
		{"100 cm to inches", "39.37"},
		{"1 km to meters", "1000"},
		{"10 feet to meters", "3.04"},
		{"1 meter to feet", "3.28"},
		{"12 inches to cm", "30.48"},
		{"1 yard to meters", "0.91"},
		{"1000 mm to meters", "1 m"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalUnits(tt.expr)
			if err != nil {
				t.Errorf("EvalUnits(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("EvalUnits(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

func TestEvalWeightConversion(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"10 kg in lbs", "22.04"},
		{"5 lbs to kg", "2.26"},
		{"100 grams to oz", "3.52"},
		{"16 oz to grams", "453"},
		{"1 kg to grams", "1000"},
		{"2000 lbs to tons", "1.00"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalUnits(tt.expr)
			if err != nil {
				t.Errorf("EvalUnits(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("EvalUnits(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

func TestEvalTemperatureConversion(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"100 f to c", "37.78°C"},
		{"0 c to f", "32°F"},
		{"100 c to f", "212°F"},
		{"0 kelvin to c", "-273.15°C"},
		{"25 celsius to fahrenheit", "77°F"},
		{"-40 f to c", "-40°C"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalUnits(tt.expr)
			if err != nil {
				t.Errorf("EvalUnits(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EvalUnits(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestEvalVolumeConversion(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"5 gallons in liters", "18.92"},
		{"1 liter to ml", "1000"},
		{"2 cups to ml", "473"},
		{"1 quart to liters", "0.94"},
		{"1 pint to cups", "2"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalUnits(tt.expr)
			if err != nil {
				t.Errorf("EvalUnits(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("EvalUnits(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

func TestEvalDataConversion(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		// SI units (base 1000)
		{"500 mb in gb", "0.5"},
		{"1 tb to gb", "1000"},
		{"1 gb to mb", "1000"},
		{"1000 kb to mb", "1 MB"},
		{"1 pb to tb", "1000"},
		// IEC units (base 1024)
		{"1 gib to mib", "1024"},
		{"1024 mib to gib", "1 GIB"},
		{"1 tib to gib", "1024"},
		// Cross conversions (SI to IEC)
		{"1000 mb to mib", "953.67"},
		{"1024 mib to mb", "1073.74"},
		// Bytes conversions
		{"1234567 bytes to mb", "1.23"},
		{"1234567 bytes to mib", "1.17"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalUnits(tt.expr)
			if err != nil {
				t.Errorf("EvalUnits(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("EvalUnits(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

func TestEvalSpeedConversion(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"60 mph to kph", "96.56"},
		{"100 kph to mph", "62.13"},
		{"10 m/s to kph", "36"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalUnits(tt.expr)
			if err != nil {
				t.Errorf("EvalUnits(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("EvalUnits(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

func TestEvalAreaConversion(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"1 acre to sqft", "43560"},
		{"1 hectare to acres", "2.47"},
		{"100 sqft to sqm", "9.29"},
		{"1 sqkm to hectares", "100"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalUnits(tt.expr)
			if err != nil {
				t.Errorf("EvalUnits(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("EvalUnits(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

func TestIsUnitExpression(t *testing.T) {
	tests := []struct {
		expr     string
		expected bool
	}{
		{"5 miles in km", true},
		{"100 f to c", true},
		{"10 kg in lbs", true},
		{"500 mb in gb", true},
		{"100 + 50", false},
		{"now in Seattle", false},
		{"sin(45)", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := IsUnitExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("IsUnitExpression(%q) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}

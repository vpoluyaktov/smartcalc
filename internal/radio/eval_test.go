package radio

import (
	"strings"
	"testing"
)

func TestIsRadioExpression(t *testing.T) {
	tests := []struct {
		expr     string
		expected bool
	}{
		{"14.2 MHz to meters", true},
		{"146 MHz in m", true},
		{"2 m to MHz", true},
		{"70 cm to MHz", true},
		{"dipole for 14.2 MHz", true},
		{"quarter wave 146 MHz", true},
		{"1/4 wave 7.1 MHz", true},
		{"yagi for 144 MHz", true},
		{"swr 1.5", true},
		{"30 dbm to watts", true},
		{"ham band 14.2 MHz", true},
		{"20m band", true},
		{"simple math 2+2", false},
		{"hello world", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := IsRadioExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("IsRadioExpression(%q) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestFrequencyToWavelength(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"14.2 MHz to meters", "21.1"},  // ~21.1 m
		{"146 MHz to m", "2.0"},         // ~2.05 m
		{"7.1 MHz in meters", "42.2"},   // ~42.2 m
		{"1 GHz to meters", "29.98"},    // ~29.98 cm
		{"3.5 MHz wavelength", "85.6"},  // ~85.6 m
		{"1000 kHz to meters", "299.7"}, // ~299.8 m
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalRadio(tt.expr)
			if err != nil {
				t.Errorf("EvalRadio(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("EvalRadio(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

func TestWavelengthToFrequency(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"2 m to MHz", "149.8"},      // ~149.9 MHz
		{"70 cm to MHz", "428.2"},    // ~428.3 MHz
		{"20 meters to MHz", "14.9"}, // ~15 MHz
		{"10 m in MHz", "29.9"},      // ~30 MHz
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalRadio(tt.expr)
			if err != nil {
				t.Errorf("EvalRadio(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("EvalRadio(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

func TestDipoleAntenna(t *testing.T) {
	tests := []struct {
		expr     string
		contains []string
	}{
		{"dipole for 14.2 MHz", []string{"Half-wave dipole", "14.200 MHz", "m", "ft"}},
		{"dipole 7.1 MHz", []string{"Half-wave dipole", "7.100 MHz"}},
		{"half-wave dipole 146 MHz", []string{"Half-wave dipole", "146.000 MHz"}},
		{"dipole for 14.2 MHz vf=0.95", []string{"Half-wave dipole", "14.200 MHz"}},
		{"dipole for 2 m", []string{"Half-wave dipole", "149.896 MHz"}},
		{"dipole for 70 cm", []string{"Half-wave dipole", "428.275 MHz"}},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalRadio(tt.expr)
			if err != nil {
				t.Errorf("EvalRadio(%q) error: %v", tt.expr, err)
				return
			}
			for _, c := range tt.contains {
				if !strings.Contains(result, c) {
					t.Errorf("EvalRadio(%q) = %q, want to contain %q", tt.expr, result, c)
				}
			}
		})
	}
}

func TestQuarterWaveVertical(t *testing.T) {
	tests := []struct {
		expr     string
		contains []string
	}{
		{"quarter wave for 14.2 MHz", []string{"Quarter-wave vertical", "14.200 MHz"}},
		{"1/4 wave 146 MHz", []string{"Quarter-wave vertical", "146.000 MHz"}},
		{"quarter-wave 7.1 MHz", []string{"Quarter-wave vertical", "7.100 MHz"}},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalRadio(tt.expr)
			if err != nil {
				t.Errorf("EvalRadio(%q) error: %v", tt.expr, err)
				return
			}
			for _, c := range tt.contains {
				if !strings.Contains(result, c) {
					t.Errorf("EvalRadio(%q) = %q, want to contain %q", tt.expr, result, c)
				}
			}
		})
	}
}

func TestYagiElements(t *testing.T) {
	tests := []struct {
		expr     string
		contains []string
	}{
		{"yagi for 144 MHz", []string{"Yagi antenna", "144.000 MHz", "Reflector", "Driven", "Director"}},
		{"yagi 14.2 MHz", []string{"Yagi antenna", "14.200 MHz"}},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalRadio(tt.expr)
			if err != nil {
				t.Errorf("EvalRadio(%q) error: %v", tt.expr, err)
				return
			}
			for _, c := range tt.contains {
				if !strings.Contains(result, c) {
					t.Errorf("EvalRadio(%q) = %q, want to contain %q", tt.expr, result, c)
				}
			}
		})
	}
}

func TestSWR(t *testing.T) {
	tests := []struct {
		expr     string
		contains []string
	}{
		{"swr 50 75", []string{"SWR", "1.50:1", "Reflection coefficient", "Return loss"}},
		{"swr 1.5", []string{"SWR 1.50:1", "Reflection coefficient", "Return loss", "Power reflected"}},
		{"vswr 2.0", []string{"SWR 2.00:1"}},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalRadio(tt.expr)
			if err != nil {
				t.Errorf("EvalRadio(%q) error: %v", tt.expr, err)
				return
			}
			for _, c := range tt.contains {
				if !strings.Contains(result, c) {
					t.Errorf("EvalRadio(%q) = %q, want to contain %q", tt.expr, result, c)
				}
			}
		})
	}
}

func TestDecibelConversion(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"3 db to times", "1.995"},         // 10^(3/10) ≈ 2
		{"10 db to times", "10.0"},         // 10^(10/10) = 10
		{"6 db to times voltage", "1.995"}, // 10^(6/20) ≈ 2
		{"2 times to db", "3.0"},           // 10*log10(2) ≈ 3
		{"10 times to db", "10.0"},         // 10*log10(10) = 10
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalRadio(tt.expr)
			if err != nil {
				t.Errorf("EvalRadio(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("EvalRadio(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

func TestPowerConversion(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"30 dbm to watts", "1.000 W"}, // 10^(30/10) mW = 1000 mW = 1 W
		{"0 dbm to watts", "1.000 mW"}, // 10^(0/10) mW = 1 mW
		{"1 watt to dbm", "30.0 dBm"},  // 10*log10(1000) = 30
		{"100 mw to dbm", "20.0 dBm"},  // 10*log10(100) = 20
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalRadio(tt.expr)
			if err != nil {
				t.Errorf("EvalRadio(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("EvalRadio(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

func TestBandInfo(t *testing.T) {
	tests := []struct {
		expr     string
		contains []string
	}{
		{"ham band 14.2 MHz", []string{"20 meters", "14.000", "14.350"}},
		{"ham band 146 MHz", []string{"2 meters", "144.000", "148.000"}},
		{"amateur band 7.1 MHz", []string{"40 meters", "7.000", "7.300"}},
		{"ham band 432 MHz", []string{"70 centimeters"}},
		{"20m band", []string{"20 meters", "14.000", "14.350"}},
		{"ham band 20m", []string{"20 meters", "14.000", "14.350"}},
		{"ham band 2m", []string{"2 meters", "144.000", "148.000"}},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalRadio(tt.expr)
			if err != nil {
				t.Errorf("EvalRadio(%q) error: %v", tt.expr, err)
				return
			}
			for _, c := range tt.contains {
				if !strings.Contains(result, c) {
					t.Errorf("EvalRadio(%q) = %q, want to contain %q", tt.expr, result, c)
				}
			}
		})
	}
}

func TestVelocityFactor(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"10m vf=0.66", "6.60 m in cable"},
		{"2m cable vf 0.82", "1.64 m in cable"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalRadio(tt.expr)
			if err != nil {
				t.Errorf("EvalRadio(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("EvalRadio(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

func TestOhmsLaw(t *testing.T) {
	tests := []struct {
		expr     string
		contains []string
	}{
		// Voltage and Current given
		{"12v 2a", []string{"Voltage: 12.000 V", "Current: 2.000 A", "Resistance: 6.000 Ω", "Power: 24.000 W"}},
		{"12 volts 2 amps", []string{"Voltage: 12.000 V", "Current: 2.000 A"}},
		// Voltage and Resistance given
		{"24v 100ohm", []string{"Voltage: 24.000 V", "Resistance: 100.000 Ω", "Current: 0.240 A", "Power: 5.760 W"}},
		{"12v 50 ohm", []string{"Voltage: 12.000 V", "Resistance: 50.000 Ω"}},
		// Current and Resistance given
		{"2a 10ohm", []string{"Current: 2.000 A", "Resistance: 10.000 Ω", "Voltage: 20.000 V", "Power: 40.000 W"}},
		// Power and Resistance given
		{"100w 50ohm", []string{"Power: 100.000 W", "Resistance: 50.000 Ω", "Voltage: 70.711 V", "Current: 1.414 A"}},
		// Power and Voltage given
		{"100w 50v", []string{"Power: 100.000 W", "Voltage: 50.000 V", "Current: 2.000 A", "Resistance: 25.000 Ω"}},
		// Power and Current given
		{"100w 5a", []string{"Power: 100.000 W", "Current: 5.000 A", "Voltage: 20.000 V", "Resistance: 4.000 Ω"}},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalRadio(tt.expr)
			if err != nil {
				t.Errorf("EvalRadio(%q) error: %v", tt.expr, err)
				return
			}
			for _, c := range tt.contains {
				if !strings.Contains(result, c) {
					t.Errorf("EvalRadio(%q) = %q, want to contain %q", tt.expr, result, c)
				}
			}
		})
	}
}

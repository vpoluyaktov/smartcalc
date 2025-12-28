package hamradio

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Speed of light in meters per second
const speedOfLight = 299792458.0

// Handler defines the interface for ham radio expression handlers.
type Handler interface {
	Handle(expr, exprLower string) (string, bool)
}

// HandlerFunc is an adapter to allow ordinary functions to be used as Handlers.
type HandlerFunc func(expr, exprLower string) (string, bool)

// Handle calls the underlying function.
func (f HandlerFunc) Handle(expr, exprLower string) (string, bool) {
	return f(expr, exprLower)
}

// handlerChain is the ordered list of handlers for ham radio expressions.
var handlerChain = []Handler{
	HandlerFunc(handleFrequencyToWavelength),
	HandlerFunc(handleWavelengthToFrequency),
	HandlerFunc(handleDipoleAntenna),
	HandlerFunc(handleQuarterWaveVertical),
	HandlerFunc(handleYagiElements),
	HandlerFunc(handleFreeToCable),
	HandlerFunc(handleSWR),
	HandlerFunc(handleDecibelConversion),
	HandlerFunc(handlePowerConversion),
	HandlerFunc(handleBandInfo),
	HandlerFunc(handleOhmsLaw),
}

// EvalHamRadio evaluates a ham radio expression and returns the result.
func EvalHamRadio(expr string) (string, error) {
	expr = strings.TrimSpace(expr)
	exprLower := strings.ToLower(expr)

	for _, h := range handlerChain {
		if result, ok := h.Handle(expr, exprLower); ok {
			return result, nil
		}
	}

	return "", fmt.Errorf("unable to evaluate ham radio expression: %s", expr)
}

// IsHamRadioExpression checks if an expression looks like a ham radio expression.
func IsHamRadioExpression(expr string) bool {
	exprLower := strings.ToLower(expr)

	// Keywords that indicate ham radio expressions
	keywords := []string{
		"mhz to m", "mhz to meters", "mhz in m", "mhz in meters",
		"khz to m", "khz to meters", "khz in m", "khz in meters",
		"ghz to m", "ghz to meters", "ghz in m", "ghz in meters",
		"m to mhz", "meters to mhz", "m in mhz", "meters in mhz",
		"m to khz", "meters to khz", "m in khz", "meters in khz",
		"dipole", "quarter wave", "quarter-wave", "1/4 wave", "λ/4",
		"yagi", "velocity factor", "vf=", "vf ", "cable vf",
		"swr", "vswr",
		"dbm to watt", "watt to dbm", "dbm to w", "w to dbm",
		"db to times", "times to db",
		"ham band", "amateur band", "m band", "cm band",
		"ohm", "volts", "amps", "watts",
	}

	for _, kw := range keywords {
		if strings.Contains(exprLower, kw) {
			return true
		}
	}

	// Pattern for frequency/wavelength with ham radio context
	patterns := []string{
		`\d+\.?\d*\s*(?:mhz|khz|ghz)\s+(?:to|in)\s+(?:m|meters?|wavelength)`,
		`\d+\.?\d*\s*(?:m|meters?)\s+(?:to|in)\s+(?:mhz|khz|ghz)`,
		`dipole\s+(?:for\s+)?\d+`,
		`(?:quarter[- ]?wave|1/4\s*wave|λ/4)\s+(?:for\s+)?\d+`,
		`swr\s+\d+`,
		`\d+\.?\d*\s*dbm`,
		`\d+\.?\d*\s*(?:v|volts?)\s+\d+\.?\d*\s*(?:a|amps?|ohms?|w|watts?)`,
		`\d+\.?\d*\s*(?:a|amps?)\s+\d+\.?\d*\s*(?:v|volts?|ohms?|w|watts?)`,
		`\d+\.?\d*\s*(?:ohms?)\s+\d+\.?\d*\s*(?:v|volts?|a|amps?|w|watts?)`,
		`\d+\.?\d*\s*(?:w|watts?)\s+\d+\.?\d*\s*(?:v|volts?|a|amps?|ohms?)`,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, exprLower); matched {
			return true
		}
	}

	return false
}

// handleFrequencyToWavelength converts frequency to wavelength
// Examples: "14.2 MHz to meters", "146 MHz in m", "7.1 mhz wavelength"
func handleFrequencyToWavelength(expr, exprLower string) (string, bool) {
	patterns := []string{
		`(?i)^([\d.]+)\s*(mhz|khz|ghz)\s+(?:to|in|->)\s+(?:m|meters?|wavelength)$`,
		`(?i)^([\d.]+)\s*(mhz|khz|ghz)\s+wavelength$`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(expr)
		if matches != nil {
			value, _ := strconv.ParseFloat(matches[1], 64)
			unit := strings.ToLower(matches[2])

			// Convert to Hz
			var freqHz float64
			switch unit {
			case "khz":
				freqHz = value * 1e3
			case "mhz":
				freqHz = value * 1e6
			case "ghz":
				freqHz = value * 1e9
			}

			wavelength := speedOfLight / freqHz
			return formatWavelength(wavelength), true
		}
	}
	return "", false
}

// handleWavelengthToFrequency converts wavelength to frequency
// Examples: "2 m to MHz", "70 cm in MHz", "20 meters to mhz"
func handleWavelengthToFrequency(expr, exprLower string) (string, bool) {
	re := regexp.MustCompile(`(?i)^([\d.]+)\s*(m|meters?|cm|centimeters?|mm)\s+(?:to|in|->)\s+(mhz|khz|ghz)$`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	value, _ := strconv.ParseFloat(matches[1], 64)
	lengthUnit := strings.ToLower(matches[2])
	freqUnit := strings.ToLower(matches[3])

	// Convert to meters
	var wavelengthM float64
	switch {
	case strings.HasPrefix(lengthUnit, "cm") || lengthUnit == "centimeters" || lengthUnit == "centimeter":
		wavelengthM = value / 100
	case lengthUnit == "mm":
		wavelengthM = value / 1000
	default:
		wavelengthM = value
	}

	freqHz := speedOfLight / wavelengthM

	// Convert to requested unit
	var result float64
	var unitStr string
	switch freqUnit {
	case "khz":
		result = freqHz / 1e3
		unitStr = "kHz"
	case "mhz":
		result = freqHz / 1e6
		unitStr = "MHz"
	case "ghz":
		result = freqHz / 1e9
		unitStr = "GHz"
	}

	return formatFrequency(result, unitStr), true
}

// handleDipoleAntenna calculates half-wave dipole antenna length
// Examples: "dipole for 14.2 MHz", "dipole 7.1 mhz", "half-wave dipole 146 MHz", "dipole for 2 m"
func handleDipoleAntenna(expr, exprLower string) (string, bool) {
	// Try frequency input first
	re := regexp.MustCompile(`(?i)^(?:half[- ]?wave\s+)?dipole\s+(?:for\s+|antenna\s+)?([\d.]+)\s*(mhz|khz|ghz)(?:\s+vf[= ]*([\d.]+))?$`)
	matches := re.FindStringSubmatch(expr)

	var freqMHz float64
	vf := 0.95 // default velocity factor

	if matches != nil {
		freq, _ := strconv.ParseFloat(matches[1], 64)
		unit := strings.ToLower(matches[2])

		if len(matches) > 3 && matches[3] != "" {
			vf, _ = strconv.ParseFloat(matches[3], 64)
		}

		switch unit {
		case "khz":
			freqMHz = freq / 1000
		case "mhz":
			freqMHz = freq
		case "ghz":
			freqMHz = freq * 1000
		}
	} else {
		// Try wavelength input (e.g., "dipole for 2 m", "dipole for 70 cm")
		re = regexp.MustCompile(`(?i)^(?:half[- ]?wave\s+)?dipole\s+(?:for\s+|antenna\s+)?([\d.]+)\s*(m|meters?|cm|centimeters?)$`)
		matches = re.FindStringSubmatch(expr)
		if matches == nil {
			return "", false
		}

		wavelength, _ := strconv.ParseFloat(matches[1], 64)
		unit := strings.ToLower(matches[2])

		// Convert to meters
		var wavelengthM float64
		if strings.HasPrefix(unit, "cm") || unit == "centimeters" || unit == "centimeter" {
			wavelengthM = wavelength / 100
		} else {
			wavelengthM = wavelength
		}

		// Convert wavelength to frequency: f = c / λ
		freqMHz = (speedOfLight / wavelengthM) / 1e6
	}

	// Half-wave dipole formula: L(feet) = 468 / f(MHz) or L(m) = 142.65 / f(MHz)
	lengthM := (142.65 / freqMHz) * vf
	lengthFt := (468 / freqMHz) * vf
	eachLegM := lengthM / 2
	eachLegFt := lengthFt / 2

	return fmt.Sprintf("\n> Half-wave dipole for %.3f MHz\n> Total length: %.2f m (%.2f ft)\n> Each leg: %.2f m (%.2f ft)",
		freqMHz, lengthM, lengthFt, eachLegM, eachLegFt), true
}

// handleQuarterWaveVertical calculates quarter-wave vertical antenna length
// Examples: "quarter wave for 14.2 MHz", "1/4 wave 146 mhz", "λ/4 7.1 MHz", "quarter wave for 2 m"
func handleQuarterWaveVertical(expr, exprLower string) (string, bool) {
	// Try frequency input first
	re := regexp.MustCompile(`(?i)^(?:quarter[- ]?wave|1/4\s*wave|λ/4)\s+(?:vertical\s+)?(?:for\s+|antenna\s+)?([\d.]+)\s*(mhz|khz|ghz)(?:\s+vf[= ]*([\d.]+))?$`)
	matches := re.FindStringSubmatch(expr)

	var freqMHz float64
	vf := 0.95 // default velocity factor

	if matches != nil {
		freq, _ := strconv.ParseFloat(matches[1], 64)
		unit := strings.ToLower(matches[2])

		if len(matches) > 3 && matches[3] != "" {
			vf, _ = strconv.ParseFloat(matches[3], 64)
		}

		switch unit {
		case "khz":
			freqMHz = freq / 1000
		case "mhz":
			freqMHz = freq
		case "ghz":
			freqMHz = freq * 1000
		}
	} else {
		// Try wavelength input (e.g., "quarter wave for 2 m", "1/4 wave 70 cm")
		re = regexp.MustCompile(`(?i)^(?:quarter[- ]?wave|1/4\s*wave|λ/4)\s+(?:vertical\s+)?(?:for\s+|antenna\s+)?([\d.]+)\s*(m|meters?|cm|centimeters?)$`)
		matches = re.FindStringSubmatch(expr)
		if matches == nil {
			return "", false
		}

		wavelength, _ := strconv.ParseFloat(matches[1], 64)
		unit := strings.ToLower(matches[2])

		var wavelengthM float64
		if strings.HasPrefix(unit, "cm") || unit == "centimeters" || unit == "centimeter" {
			wavelengthM = wavelength / 100
		} else {
			wavelengthM = wavelength
		}

		freqMHz = (speedOfLight / wavelengthM) / 1e6
	}

	// Quarter-wave formula: L(feet) = 234 / f(MHz) or L(m) = 71.325 / f(MHz)
	lengthM := (71.325 / freqMHz) * vf
	lengthFt := (234 / freqMHz) * vf

	return fmt.Sprintf("\n> Quarter-wave vertical for %.3f MHz\n> Length: %.2f m (%.2f ft)",
		freqMHz, lengthM, lengthFt), true
}

// handleYagiElements calculates Yagi antenna element lengths
// Examples: "yagi for 144 MHz", "yagi 14.2 mhz 3 elements"
func handleYagiElements(expr, exprLower string) (string, bool) {
	re := regexp.MustCompile(`(?i)^yagi\s+(?:for\s+|antenna\s+)?([\d.]+)\s*(mhz|khz|ghz)(?:\s+(\d+)\s*elements?)?$`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	freq, _ := strconv.ParseFloat(matches[1], 64)
	unit := strings.ToLower(matches[2])

	// Convert to MHz
	var freqMHz float64
	switch unit {
	case "khz":
		freqMHz = freq / 1000
	case "mhz":
		freqMHz = freq
	case "ghz":
		freqMHz = freq * 1000
	}

	// Wavelength in meters
	wavelengthM := 299.792458 / freqMHz

	// Standard Yagi element lengths (as fraction of wavelength)
	reflector := wavelengthM * 0.495  // ~0.495λ
	drivenElem := wavelengthM * 0.473 // ~0.473λ (half-wave dipole)
	director := wavelengthM * 0.440   // ~0.440λ

	return fmt.Sprintf("\n> Yagi antenna for %.3f MHz (λ = %.2f m)\n> Reflector: %.2f m (%.2f ft)\n> Driven element: %.2f m (%.2f ft)\n> Director: %.2f m (%.2f ft)",
		freqMHz, wavelengthM,
		reflector, reflector*3.28084,
		drivenElem, drivenElem*3.28084,
		director, director*3.28084), true
}

// handleFreeToCable converts free space wavelength to cable wavelength with velocity factor
// Examples: "10m vf=0.66", "2m cable vf 0.82"
func handleFreeToCable(expr, exprLower string) (string, bool) {
	re := regexp.MustCompile(`(?i)^([\d.]+)\s*(?:m|meters?)\s+(?:cable\s+)?(?:vf|velocity factor)[= ]\s*([\d.]+)$`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", false
	}

	freeSpace, _ := strconv.ParseFloat(matches[1], 64)
	vf, _ := strconv.ParseFloat(matches[2], 64)

	cableLength := freeSpace * vf

	return fmt.Sprintf("%.2f m free space × %.2f VF = %.2f m in cable", freeSpace, vf, cableLength), true
}

// handleSWR calculates SWR-related values
// Examples: "swr 50 75" (impedance mismatch), "swr 1.5 return loss"
func handleSWR(expr, exprLower string) (string, bool) {
	// SWR from two impedances
	re := regexp.MustCompile(`(?i)^(?:swr|vswr)\s+([\d.]+)\s*(?:ohm|Ω)?\s+([\d.]+)\s*(?:ohm|Ω)?$`)
	matches := re.FindStringSubmatch(expr)
	if matches != nil {
		z1, _ := strconv.ParseFloat(matches[1], 64)
		z2, _ := strconv.ParseFloat(matches[2], 64)

		// Ensure z1 >= z2
		if z1 < z2 {
			z1, z2 = z2, z1
		}

		swr := z1 / z2
		reflectionCoef := (swr - 1) / (swr + 1)
		returnLoss := -20 * math.Log10(reflectionCoef)
		powerReflected := reflectionCoef * reflectionCoef * 100

		return fmt.Sprintf("\n> SWR %.1f Ω / %.1f Ω = %.2f:1\n> Reflection coefficient: %.3f\n> Return loss: %.1f dB\n> Power reflected: %.1f%%",
			z1, z2, swr, reflectionCoef, returnLoss, powerReflected), true
	}

	// SWR to return loss and other values
	re = regexp.MustCompile(`(?i)^(?:swr|vswr)\s+([\d.]+)(?::1)?$`)
	matches = re.FindStringSubmatch(expr)
	if matches != nil {
		swr, _ := strconv.ParseFloat(matches[1], 64)
		if swr < 1 {
			return "", false
		}

		reflectionCoef := (swr - 1) / (swr + 1)
		returnLoss := -20 * math.Log10(reflectionCoef)
		powerReflected := reflectionCoef * reflectionCoef * 100

		return fmt.Sprintf("\n> SWR %.2f:1\n> Reflection coefficient: %.3f\n> Return loss: %.1f dB\n> Power reflected: %.1f%%",
			swr, reflectionCoef, returnLoss, powerReflected), true
	}

	return "", false
}

// handleDecibelConversion converts between dB and linear ratios
// Examples: "3 db to times", "10 times to db", "6 db voltage"
func handleDecibelConversion(expr, exprLower string) (string, bool) {
	// dB to linear (power)
	re := regexp.MustCompile(`(?i)^([-\d.]+)\s*db\s+(?:to\s+)?(?:times|ratio|linear)(?:\s+power)?$`)
	matches := re.FindStringSubmatch(expr)
	if matches != nil {
		db, _ := strconv.ParseFloat(matches[1], 64)
		ratio := math.Pow(10, db/10)
		return fmt.Sprintf("%.1f dB = %.3f× (power ratio)", db, ratio), true
	}

	// dB to linear (voltage)
	re = regexp.MustCompile(`(?i)^([-\d.]+)\s*db\s+(?:to\s+)?(?:times|ratio|linear)\s+voltage$`)
	matches = re.FindStringSubmatch(expr)
	if matches != nil {
		db, _ := strconv.ParseFloat(matches[1], 64)
		ratio := math.Pow(10, db/20)
		return fmt.Sprintf("%.1f dB = %.3f× (voltage ratio)", db, ratio), true
	}

	// Linear to dB (power)
	re = regexp.MustCompile(`(?i)^([\d.]+)\s*(?:times|x|×)\s+(?:to\s+)?db(?:\s+power)?$`)
	matches = re.FindStringSubmatch(expr)
	if matches != nil {
		ratio, _ := strconv.ParseFloat(matches[1], 64)
		db := 10 * math.Log10(ratio)
		return fmt.Sprintf("%.3f× = %.1f dB (power)", ratio, db), true
	}

	return "", false
}

// handlePowerConversion converts between dBm and watts
// Examples: "30 dbm to watts", "1 watt to dbm", "100 mw to dbm"
func handlePowerConversion(expr, exprLower string) (string, bool) {
	// dBm to watts
	re := regexp.MustCompile(`(?i)^([-\d.]+)\s*dbm\s+(?:to|in)\s+(?:w|watts?|mw|milliwatts?)$`)
	matches := re.FindStringSubmatch(expr)
	if matches != nil {
		dbm, _ := strconv.ParseFloat(matches[1], 64)
		mw := math.Pow(10, dbm/10)
		w := mw / 1000

		if w >= 1 {
			return fmt.Sprintf("%.1f dBm = %.3f W", dbm, w), true
		}
		return fmt.Sprintf("%.1f dBm = %.3f mW", dbm, mw), true
	}

	// Watts to dBm
	re = regexp.MustCompile(`(?i)^([\d.]+)\s*(w|watts?|mw|milliwatts?)\s+(?:to|in)\s+dbm$`)
	matches = re.FindStringSubmatch(expr)
	if matches != nil {
		value, _ := strconv.ParseFloat(matches[1], 64)
		unit := strings.ToLower(matches[2])

		var mw float64
		if strings.HasPrefix(unit, "mw") || strings.HasPrefix(unit, "milliwatt") {
			mw = value
		} else {
			mw = value * 1000
		}

		dbm := 10 * math.Log10(mw)
		return fmt.Sprintf("%.3f %s = %.1f dBm", value, matches[2], dbm), true
	}

	return "", false
}

// Amateur radio band information
var hamBands = map[string]struct {
	name       string
	freqStart  float64
	freqEnd    float64
	wavelength string
}{
	"160m":  {"160 meters", 1.8, 2.0, "160m"},
	"80m":   {"80 meters", 3.5, 4.0, "80m"},
	"60m":   {"60 meters", 5.3305, 5.4065, "60m"},
	"40m":   {"40 meters", 7.0, 7.3, "40m"},
	"30m":   {"30 meters", 10.1, 10.15, "30m"},
	"20m":   {"20 meters", 14.0, 14.35, "20m"},
	"17m":   {"17 meters", 18.068, 18.168, "17m"},
	"15m":   {"15 meters", 21.0, 21.45, "15m"},
	"12m":   {"12 meters", 24.89, 24.99, "12m"},
	"10m":   {"10 meters", 28.0, 29.7, "10m"},
	"6m":    {"6 meters", 50.0, 54.0, "6m"},
	"2m":    {"2 meters", 144.0, 148.0, "2m"},
	"1.25m": {"1.25 meters", 222.0, 225.0, "1.25m"},
	"70cm":  {"70 centimeters", 420.0, 450.0, "70cm"},
	"33cm":  {"33 centimeters", 902.0, 928.0, "33cm"},
	"23cm":  {"23 centimeters", 1240.0, 1300.0, "23cm"},
}

// handleBandInfo provides information about amateur radio bands
// Examples: "ham band 14.2 MHz", "amateur band 146 MHz", "20m band"
func handleBandInfo(expr, exprLower string) (string, bool) {
	// By frequency
	re := regexp.MustCompile(`(?i)^(?:ham|amateur)\s+band\s+([\d.]+)\s*(mhz|khz|ghz)?$`)
	matches := re.FindStringSubmatch(expr)
	if matches != nil {
		freq, _ := strconv.ParseFloat(matches[1], 64)
		unit := "mhz"
		if len(matches) > 2 && matches[2] != "" {
			unit = strings.ToLower(matches[2])
		}

		// Convert to MHz
		var freqMHz float64
		switch unit {
		case "khz":
			freqMHz = freq / 1000
		case "mhz":
			freqMHz = freq
		case "ghz":
			freqMHz = freq * 1000
		}

		for _, band := range hamBands {
			if freqMHz >= band.freqStart && freqMHz <= band.freqEnd {
				return fmt.Sprintf("\n> %.3f MHz is in the %s band\n> Range: %.3f - %.3f MHz",
					freqMHz, band.name, band.freqStart, band.freqEnd), true
			}
		}

		return "not in a standard amateur radio band", true
	}

	// By band name: "20m band", "ham band 20m", "amateur band 2m"
	re = regexp.MustCompile(`(?i)^(?:(?:ham|amateur)\s+band\s+)?(\d+\.?\d*)\s*(m|cm)(?:\s+band)?$`)
	matches = re.FindStringSubmatch(expr)
	if matches != nil {
		bandName := strings.ToLower(matches[1] + matches[2])

		if band, ok := hamBands[bandName]; ok {
			return fmt.Sprintf("\n> %s band\n> Range: %.3f - %.3f MHz",
				band.name, band.freqStart, band.freqEnd), true
		}
	}

	return "", false
}

// handleOhmsLaw calculates electrical values using Ohm's Law and Power formulas
// V = I * R, P = V * I, P = I² * R, P = V² / R
// Examples: "12v 2a", "24v 100ohm", "5a 10ohm", "100w 50ohm"
func handleOhmsLaw(expr, exprLower string) (string, bool) {
	// Parse two electrical values
	// Patterns: "12v 2a", "12 volts 2 amps", "24v 100 ohm", "100w 50 ohm"

	var voltage, current, resistance, power float64
	var hasV, hasA, hasR, hasP bool

	// Voltage patterns
	reV := regexp.MustCompile(`(?i)([\d.]+)\s*(?:v|volts?)(?:\s|$)`)
	if matches := reV.FindStringSubmatch(expr); matches != nil {
		voltage, _ = strconv.ParseFloat(matches[1], 64)
		hasV = true
	}

	// Current patterns
	reA := regexp.MustCompile(`(?i)([\d.]+)\s*(?:a|amps?|amperes?)(?:\s|$)`)
	if matches := reA.FindStringSubmatch(expr); matches != nil {
		current, _ = strconv.ParseFloat(matches[1], 64)
		hasA = true
	}

	// Resistance patterns
	reR := regexp.MustCompile(`(?i)([\d.]+)\s*(?:ohms?|Ω)(?:\s|$)`)
	if matches := reR.FindStringSubmatch(expr); matches != nil {
		resistance, _ = strconv.ParseFloat(matches[1], 64)
		hasR = true
	}

	// Power patterns
	reP := regexp.MustCompile(`(?i)([\d.]+)\s*(?:w|watts?)(?:\s|$)`)
	if matches := reP.FindStringSubmatch(expr); matches != nil {
		power, _ = strconv.ParseFloat(matches[1], 64)
		hasP = true
	}

	// Count how many values we have
	count := 0
	if hasV {
		count++
	}
	if hasA {
		count++
	}
	if hasR {
		count++
	}
	if hasP {
		count++
	}

	// Need exactly 2 values to calculate the others
	if count != 2 {
		return "", false
	}

	// Calculate missing values using Ohm's Law and Power formulas
	if hasV && hasA {
		// V and I given: R = V/I, P = V*I
		resistance = voltage / current
		power = voltage * current
	} else if hasV && hasR {
		// V and R given: I = V/R, P = V²/R
		current = voltage / resistance
		power = (voltage * voltage) / resistance
	} else if hasV && hasP {
		// V and P given: I = P/V, R = V²/P
		current = power / voltage
		resistance = (voltage * voltage) / power
	} else if hasA && hasR {
		// I and R given: V = I*R, P = I²*R
		voltage = current * resistance
		power = current * current * resistance
	} else if hasA && hasP {
		// I and P given: V = P/I, R = P/I²
		voltage = power / current
		resistance = power / (current * current)
	} else if hasR && hasP {
		// R and P given: V = √(P*R), I = √(P/R)
		voltage = math.Sqrt(power * resistance)
		current = math.Sqrt(power / resistance)
	}

	return fmt.Sprintf("\n> Voltage: %.3f V\n> Current: %.3f A\n> Resistance: %.3f Ω\n> Power: %.3f W",
		voltage, current, resistance, power), true
}

// Helper functions

func formatWavelength(meters float64) string {
	if meters >= 1 {
		return fmt.Sprintf("%.3f m", meters)
	} else if meters >= 0.01 {
		return fmt.Sprintf("%.2f cm", meters*100)
	}
	return fmt.Sprintf("%.2f mm", meters*1000)
}

func formatFrequency(value float64, unit string) string {
	if value >= 1000 && unit == "kHz" {
		return fmt.Sprintf("%.3f MHz", value/1000)
	}
	if value >= 1000 && unit == "MHz" {
		return fmt.Sprintf("%.3f GHz", value/1000)
	}
	if value < 1 && unit == "MHz" {
		return fmt.Sprintf("%.3f kHz", value*1000)
	}
	if value < 1 && unit == "GHz" {
		return fmt.Sprintf("%.3f MHz", value*1000)
	}
	return fmt.Sprintf("%.3f %s", value, unit)
}

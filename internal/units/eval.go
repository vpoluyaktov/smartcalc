package units

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Handler defines the interface for unit conversion handlers.
type Handler interface {
	Handle(expr, exprLower string) (string, bool)
}

// HandlerFunc is an adapter to allow ordinary functions to be used as Handlers.
type HandlerFunc func(expr, exprLower string) (string, bool)

// Handle calls the underlying function.
func (f HandlerFunc) Handle(expr, exprLower string) (string, bool) {
	return f(expr, exprLower)
}

// handlerChain is the ordered list of handlers for unit conversions.
var handlerChain = []Handler{
	HandlerFunc(handleLengthConversion),
	HandlerFunc(handleWeightConversion),
	HandlerFunc(handleTemperatureConversion),
	HandlerFunc(handleVolumeConversion),
	HandlerFunc(handleDataConversion),
	HandlerFunc(handleSpeedConversion),
	HandlerFunc(handleAreaConversion),
}

// EvalUnits evaluates a unit conversion expression and returns the result.
func EvalUnits(expr string) (string, error) {
	expr = strings.TrimSpace(expr)
	exprLower := strings.ToLower(expr)

	for _, h := range handlerChain {
		if result, ok := h.Handle(expr, exprLower); ok {
			return result, nil
		}
	}

	return "", fmt.Errorf("unable to evaluate unit conversion: %s", expr)
}

// IsUnitExpression checks if an expression looks like a unit conversion.
func IsUnitExpression(expr string) bool {
	exprLower := strings.ToLower(expr)

	// Pattern: "number unit in/to unit"
	pattern := `[\d.]+\s*[a-z°]+\s+(?:in|to)\s+[a-z°]+`
	if matched, _ := regexp.MatchString(pattern, exprLower); matched {
		return true
	}

	// Check for unit keywords
	unitKeywords := []string{
		"miles", "km", "kilometers", "meters", "feet", "inches", "yards", "cm", "mm",
		"kg", "kilograms", "lbs", "pounds", "oz", "ounces", "grams", "tons",
		"celsius", "fahrenheit", "kelvin",
		"liters", "gallons", "ml", "cups", "pints", "quarts",
		"bytes", "kb", "mb", "gb", "tb", "pb", "kib", "mib", "gib", "tib", "pib",
		"mph", "kph", "m/s",
		"acres", "hectares", "sqft", "sqm",
	}

	for _, kw := range unitKeywords {
		if strings.Contains(exprLower, kw) && (strings.Contains(exprLower, " in ") || strings.Contains(exprLower, " to ")) {
			return true
		}
	}

	// Temperature patterns
	if matched, _ := regexp.MatchString(`\d+\s*°?[cfk]\s+(?:in|to)\s+`, exprLower); matched {
		return true
	}

	return false
}

// Length conversion factors to meters
var lengthToMeters = map[string]float64{
	"m": 1, "meter": 1, "meters": 1, "metre": 1, "metres": 1,
	"km": 1000, "kilometer": 1000, "kilometers": 1000, "kilometre": 1000, "kilometres": 1000,
	"cm": 0.01, "centimeter": 0.01, "centimeters": 0.01, "centimetre": 0.01, "centimetres": 0.01,
	"mm": 0.001, "millimeter": 0.001, "millimeters": 0.001, "millimetre": 0.001, "millimetres": 0.001,
	"mi": 1609.344, "mile": 1609.344, "miles": 1609.344,
	"yd": 0.9144, "yard": 0.9144, "yards": 0.9144,
	"ft": 0.3048, "foot": 0.3048, "feet": 0.3048,
	"in": 0.0254, "inch": 0.0254, "inches": 0.0254,
	"nm": 1852, "nautical mile": 1852, "nautical miles": 1852,
}

// Weight conversion factors to grams
var weightToGrams = map[string]float64{
	"g": 1, "gram": 1, "grams": 1,
	"kg": 1000, "kilogram": 1000, "kilograms": 1000, "kilo": 1000, "kilos": 1000,
	"mg": 0.001, "milligram": 0.001, "milligrams": 0.001,
	"lb": 453.592, "lbs": 453.592, "pound": 453.592, "pounds": 453.592,
	"oz": 28.3495, "ounce": 28.3495, "ounces": 28.3495,
	"ton": 907185, "tons": 907185, "short ton": 907185, "short tons": 907185,
	"tonne": 1000000, "tonnes": 1000000, "metric ton": 1000000, "metric tons": 1000000,
	"st": 6350.29, "stone": 6350.29, "stones": 6350.29,
}

// Volume conversion factors to liters
var volumeToLiters = map[string]float64{
	"l": 1, "liter": 1, "liters": 1, "litre": 1, "litres": 1,
	"ml": 0.001, "milliliter": 0.001, "milliliters": 0.001, "millilitre": 0.001, "millilitres": 0.001,
	"gal": 3.78541, "gallon": 3.78541, "gallons": 3.78541,
	"qt": 0.946353, "quart": 0.946353, "quarts": 0.946353,
	"pt": 0.473176, "pint": 0.473176, "pints": 0.473176,
	"cup": 0.236588, "cups": 0.236588,
	"floz": 0.0295735, "fl oz": 0.0295735, "fluid ounce": 0.0295735, "fluid ounces": 0.0295735,
	"tbsp": 0.0147868, "tablespoon": 0.0147868, "tablespoons": 0.0147868,
	"tsp": 0.00492892, "teaspoon": 0.00492892, "teaspoons": 0.00492892,
}

// Data conversion factors to bytes
// SI/Decimal units (base 1000): KB, MB, GB, TB, PB
// IEC/Binary units (base 1024): KiB, MiB, GiB, TiB, PiB
var dataToBytes = map[string]float64{
	// Bytes
	"b": 1, "byte": 1, "bytes": 1,

	// SI/Decimal units (base 1000)
	"kb": 1000, "kilobyte": 1000, "kilobytes": 1000,
	"mb": 1000 * 1000, "megabyte": 1000 * 1000, "megabytes": 1000 * 1000,
	"gb": 1000 * 1000 * 1000, "gigabyte": 1000 * 1000 * 1000, "gigabytes": 1000 * 1000 * 1000,
	"tb": 1000 * 1000 * 1000 * 1000, "terabyte": 1000 * 1000 * 1000 * 1000, "terabytes": 1000 * 1000 * 1000 * 1000,
	"pb": 1000 * 1000 * 1000 * 1000 * 1000, "petabyte": 1000 * 1000 * 1000 * 1000 * 1000, "petabytes": 1000 * 1000 * 1000 * 1000 * 1000,

	// IEC/Binary units (base 1024)
	"kib": 1024, "kibibyte": 1024, "kibibytes": 1024,
	"mib": 1024 * 1024, "mebibyte": 1024 * 1024, "mebibytes": 1024 * 1024,
	"gib": 1024 * 1024 * 1024, "gibibyte": 1024 * 1024 * 1024, "gibibytes": 1024 * 1024 * 1024,
	"tib": 1024 * 1024 * 1024 * 1024, "tebibyte": 1024 * 1024 * 1024 * 1024, "tebibytes": 1024 * 1024 * 1024 * 1024,
	"pib": 1024 * 1024 * 1024 * 1024 * 1024, "pebibyte": 1024 * 1024 * 1024 * 1024 * 1024, "pebibytes": 1024 * 1024 * 1024 * 1024 * 1024,
}

// Speed conversion factors to m/s
var speedToMPS = map[string]float64{
	"m/s": 1, "mps": 1, "meters per second": 1,
	"km/h": 0.277778, "kph": 0.277778, "kmh": 0.277778, "kilometers per hour": 0.277778,
	"mph": 0.44704, "miles per hour": 0.44704,
	"knot": 0.514444, "knots": 0.514444, "kn": 0.514444,
	"ft/s": 0.3048, "fps": 0.3048, "feet per second": 0.3048,
}

// Area conversion factors to square meters
var areaToSqMeters = map[string]float64{
	"sqm": 1, "m2": 1, "m²": 1, "square meter": 1, "square meters": 1, "sq m": 1,
	"sqkm": 1000000, "km2": 1000000, "km²": 1000000, "square kilometer": 1000000, "square kilometers": 1000000,
	"sqft": 0.092903, "ft2": 0.092903, "ft²": 0.092903, "square foot": 0.092903, "square feet": 0.092903, "sq ft": 0.092903,
	"sqmi": 2589988, "mi2": 2589988, "mi²": 2589988, "square mile": 2589988, "square miles": 2589988,
	"acre": 4046.86, "acres": 4046.86,
	"ha": 10000, "hectare": 10000, "hectares": 10000,
	"sqyd": 0.836127, "yd2": 0.836127, "square yard": 0.836127, "square yards": 0.836127,
	"sqin": 0.00064516, "in2": 0.00064516, "square inch": 0.00064516, "square inches": 0.00064516,
}

func handleLengthConversion(expr, exprLower string) (string, bool) {
	re := regexp.MustCompile(`^([\d.]+)\s*([a-z]+(?:\s+[a-z]+)?)\s+(?:in|to)\s+([a-z]+(?:\s+[a-z]+)?)$`)
	matches := re.FindStringSubmatch(exprLower)
	if matches == nil {
		return "", false
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return "", false
	}

	fromUnit := strings.TrimSpace(matches[2])
	toUnit := strings.TrimSpace(matches[3])

	fromFactor, fromOk := lengthToMeters[fromUnit]
	toFactor, toOk := lengthToMeters[toUnit]

	if !fromOk || !toOk {
		return "", false
	}

	result := value * fromFactor / toFactor
	return formatResult(result, toUnit), true
}

func handleWeightConversion(expr, exprLower string) (string, bool) {
	re := regexp.MustCompile(`^([\d.]+)\s*([a-z]+(?:\s+[a-z]+)?)\s+(?:in|to)\s+([a-z]+(?:\s+[a-z]+)?)$`)
	matches := re.FindStringSubmatch(exprLower)
	if matches == nil {
		return "", false
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return "", false
	}

	fromUnit := strings.TrimSpace(matches[2])
	toUnit := strings.TrimSpace(matches[3])

	fromFactor, fromOk := weightToGrams[fromUnit]
	toFactor, toOk := weightToGrams[toUnit]

	if !fromOk || !toOk {
		return "", false
	}

	result := value * fromFactor / toFactor
	return formatResult(result, toUnit), true
}

func handleTemperatureConversion(expr, exprLower string) (string, bool) {
	// Pattern: "100°F in Celsius" or "25 C to F" or "100 fahrenheit to celsius"
	re := regexp.MustCompile(`^([\d.-]+)\s*°?\s*([cfk]|celsius|fahrenheit|kelvin)\s+(?:in|to)\s+°?\s*([cfk]|celsius|fahrenheit|kelvin)$`)
	matches := re.FindStringSubmatch(exprLower)
	if matches == nil {
		return "", false
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return "", false
	}

	fromUnit := normalizeTemperatureUnit(matches[2])
	toUnit := normalizeTemperatureUnit(matches[3])

	if fromUnit == "" || toUnit == "" {
		return "", false
	}

	result := convertTemperature(value, fromUnit, toUnit)
	return formatTemperatureResult(result, toUnit), true
}

func normalizeTemperatureUnit(unit string) string {
	switch strings.ToLower(unit) {
	case "c", "celsius":
		return "C"
	case "f", "fahrenheit":
		return "F"
	case "k", "kelvin":
		return "K"
	}
	return ""
}

func convertTemperature(value float64, from, to string) float64 {
	// Convert to Celsius first
	var celsius float64
	switch from {
	case "C":
		celsius = value
	case "F":
		celsius = (value - 32) * 5 / 9
	case "K":
		celsius = value - 273.15
	}

	// Convert from Celsius to target
	switch to {
	case "C":
		return celsius
	case "F":
		return celsius*9/5 + 32
	case "K":
		return celsius + 273.15
	}
	return 0
}

func formatTemperatureResult(value float64, unit string) string {
	if value == float64(int(value)) {
		return fmt.Sprintf("%.0f°%s", value, unit)
	}
	return fmt.Sprintf("%.2f°%s", value, unit)
}

func handleVolumeConversion(expr, exprLower string) (string, bool) {
	re := regexp.MustCompile(`^([\d.]+)\s*([a-z]+(?:\s+[a-z]+)?)\s+(?:in|to)\s+([a-z]+(?:\s+[a-z]+)?)$`)
	matches := re.FindStringSubmatch(exprLower)
	if matches == nil {
		return "", false
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return "", false
	}

	fromUnit := strings.TrimSpace(matches[2])
	toUnit := strings.TrimSpace(matches[3])

	fromFactor, fromOk := volumeToLiters[fromUnit]
	toFactor, toOk := volumeToLiters[toUnit]

	if !fromOk || !toOk {
		return "", false
	}

	result := value * fromFactor / toFactor
	return formatResult(result, toUnit), true
}

func handleDataConversion(expr, exprLower string) (string, bool) {
	re := regexp.MustCompile(`^([\d.]+)\s*([a-z]+)\s+(?:in|to)\s+([a-z]+)$`)
	matches := re.FindStringSubmatch(exprLower)
	if matches == nil {
		return "", false
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return "", false
	}

	fromUnit := strings.TrimSpace(matches[2])
	toUnit := strings.TrimSpace(matches[3])

	fromFactor, fromOk := dataToBytes[fromUnit]
	toFactor, toOk := dataToBytes[toUnit]

	if !fromOk || !toOk {
		return "", false
	}

	result := value * fromFactor / toFactor
	return formatResult(result, strings.ToUpper(toUnit)), true
}

func handleSpeedConversion(expr, exprLower string) (string, bool) {
	re := regexp.MustCompile(`^([\d.]+)\s*([a-z/]+(?:\s+[a-z]+)*)\s+(?:in|to)\s+([a-z/]+(?:\s+[a-z]+)*)$`)
	matches := re.FindStringSubmatch(exprLower)
	if matches == nil {
		return "", false
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return "", false
	}

	fromUnit := strings.TrimSpace(matches[2])
	toUnit := strings.TrimSpace(matches[3])

	fromFactor, fromOk := speedToMPS[fromUnit]
	toFactor, toOk := speedToMPS[toUnit]

	if !fromOk || !toOk {
		return "", false
	}

	result := value * fromFactor / toFactor
	return formatResult(result, toUnit), true
}

func handleAreaConversion(expr, exprLower string) (string, bool) {
	re := regexp.MustCompile(`^([\d.]+)\s*([a-z²]+(?:\s+[a-z]+)*)\s+(?:in|to)\s+([a-z²]+(?:\s+[a-z]+)*)$`)
	matches := re.FindStringSubmatch(exprLower)
	if matches == nil {
		return "", false
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return "", false
	}

	fromUnit := strings.TrimSpace(matches[2])
	toUnit := strings.TrimSpace(matches[3])

	fromFactor, fromOk := areaToSqMeters[fromUnit]
	toFactor, toOk := areaToSqMeters[toUnit]

	if !fromOk || !toOk {
		return "", false
	}

	result := value * fromFactor / toFactor
	return formatResult(result, toUnit), true
}

func formatResult(value float64, unit string) string {
	if value == float64(int64(value)) && value < 1e15 {
		return fmt.Sprintf("%.0f %s", value, unit)
	}
	if value >= 1000000 || value < 0.001 {
		return fmt.Sprintf("%.6g %s", value, unit)
	}
	if value == float64(int(value*100))/100 {
		return fmt.Sprintf("%.2f %s", value, unit)
	}
	return fmt.Sprintf("%.4f %s", value, unit)
}

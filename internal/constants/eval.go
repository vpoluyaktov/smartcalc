package constants

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

// Constant represents a physical or mathematical constant
type Constant struct {
	Value       float64
	Unit        string
	Description string
}

// constants maps names to their values
var constants = map[string]Constant{
	// Mathematical constants
	"pi":           {math.Pi, "", "ratio of circumference to diameter"},
	"π":            {math.Pi, "", "ratio of circumference to diameter"},
	"e":            {math.E, "", "Euler's number"},
	"phi":          {1.6180339887498948, "", "golden ratio"},
	"φ":            {1.6180339887498948, "", "golden ratio"},
	"golden ratio": {1.6180339887498948, "", "golden ratio"},
	"sqrt2":        {math.Sqrt2, "", "square root of 2"},
	"sqrt3":        {1.7320508075688772, "", "square root of 3"},
	"ln2":          {math.Ln2, "", "natural log of 2"},
	"ln10":         {math.Ln10, "", "natural log of 10"},

	// Physical constants
	"speed of light":      {299792458, "m/s", "speed of light in vacuum"},
	"c":                   {299792458, "m/s", "speed of light in vacuum"},
	"gravity":             {9.80665, "m/s²", "standard gravity"},
	"g":                   {9.80665, "m/s²", "standard gravity"},
	"planck":              {6.62607015e-34, "J·s", "Planck constant"},
	"h":                   {6.62607015e-34, "J·s", "Planck constant"},
	"avogadro":            {6.02214076e23, "mol⁻¹", "Avogadro constant"},
	"na":                  {6.02214076e23, "mol⁻¹", "Avogadro constant"},
	"boltzmann":           {1.380649e-23, "J/K", "Boltzmann constant"},
	"kb":                  {1.380649e-23, "J/K", "Boltzmann constant"},
	"electron mass":       {9.1093837015e-31, "kg", "electron mass"},
	"proton mass":         {1.67262192369e-27, "kg", "proton mass"},
	"elementary charge":   {1.602176634e-19, "C", "elementary charge"},
	"vacuum permittivity": {8.8541878128e-12, "F/m", "vacuum permittivity"},
	"vacuum permeability": {1.25663706212e-6, "H/m", "vacuum permeability"},
	"gas constant":        {8.314462618, "J/(mol·K)", "ideal gas constant"},
	"r":                   {8.314462618, "J/(mol·K)", "ideal gas constant"},
	"stefan boltzmann":    {5.670374419e-8, "W/(m²·K⁴)", "Stefan-Boltzmann constant"},
	"gravitational":       {6.67430e-11, "m³/(kg·s²)", "gravitational constant"},
	"big g":               {6.67430e-11, "m³/(kg·s²)", "gravitational constant"},

	// Astronomical constants
	"earth mass":        {5.972e24, "kg", "mass of Earth"},
	"earth radius":      {6.371e6, "m", "mean radius of Earth"},
	"sun mass":          {1.989e30, "kg", "mass of Sun"},
	"moon mass":         {7.342e22, "kg", "mass of Moon"},
	"au":                {1.495978707e11, "m", "astronomical unit"},
	"astronomical unit": {1.495978707e11, "m", "astronomical unit"},
	"light year":        {9.4607e15, "m", "light year"},
	"parsec":            {3.0857e16, "m", "parsec"},
}

// Handler defines the interface for constant handlers.
type Handler interface {
	Handle(expr, exprLower string) (string, bool)
}

// HandlerFunc is an adapter to allow ordinary functions to be used as Handlers.
type HandlerFunc func(expr, exprLower string) (string, bool)

// Handle calls the underlying function.
func (f HandlerFunc) Handle(expr, exprLower string) (string, bool) {
	return f(expr, exprLower)
}

// handlerChain is the ordered list of handlers for constants.
var handlerChain = []Handler{
	HandlerFunc(handleConstantLookup),
}

// EvalConstants evaluates a constant expression and returns the result.
func EvalConstants(expr string) (string, error) {
	expr = strings.TrimSpace(expr)
	exprLower := strings.ToLower(expr)

	for _, h := range handlerChain {
		if result, ok := h.Handle(expr, exprLower); ok {
			return result, nil
		}
	}

	return "", fmt.Errorf("unable to evaluate constant: %s", expr)
}

// IsConstantExpression checks if an expression looks like a constant lookup.
func IsConstantExpression(expr string) bool {
	exprLower := strings.ToLower(strings.TrimSpace(expr))

	// Check for exact matches
	if _, ok := constants[exprLower]; ok {
		return true
	}

	// Check for "value of X" pattern
	if strings.HasPrefix(exprLower, "value of ") {
		name := strings.TrimPrefix(exprLower, "value of ")
		if _, ok := constants[name]; ok {
			return true
		}
	}

	return false
}

func handleConstantLookup(expr, exprLower string) (string, bool) {
	// Try direct lookup
	if c, ok := constants[exprLower]; ok {
		return formatConstant(c), true
	}

	// Try "value of X" pattern
	re := regexp.MustCompile(`^value\s+of\s+(.+)$`)
	if matches := re.FindStringSubmatch(exprLower); matches != nil {
		name := strings.TrimSpace(matches[1])
		if c, ok := constants[name]; ok {
			return formatConstant(c), true
		}
	}

	return "", false
}

func formatConstant(c Constant) string {
	if c.Unit != "" {
		return fmt.Sprintf("%g %s", c.Value, c.Unit)
	}
	// For mathematical constants, show more precision
	if c.Value == float64(int(c.Value)) {
		return fmt.Sprintf("%.0f", c.Value)
	}
	return fmt.Sprintf("%.10g", c.Value)
}

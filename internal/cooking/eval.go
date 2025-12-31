package cooking

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Volume units in milliliters (base unit)
var volumeUnits = map[string]float64{
	// US Customary
	"tsp":          4.92892,
	"teaspoon":     4.92892,
	"teaspoons":    4.92892,
	"tbsp":         14.7868,
	"tablespoon":   14.7868,
	"tablespoons":  14.7868,
	"floz":         29.5735,
	"fl oz":        29.5735,
	"fluid oz":     29.5735,
	"fluid ounce":  29.5735,
	"fluid ounces": 29.5735,
	"cup":          236.588,
	"cups":         236.588,
	"pint":         473.176,
	"pints":        473.176,
	"pt":           473.176,
	"quart":        946.353,
	"quarts":       946.353,
	"qt":           946.353,
	"gallon":       3785.41,
	"gallons":      3785.41,
	"gal":          3785.41,

	// Metric
	"ml":          1,
	"milliliter":  1,
	"milliliters": 1,
	"millilitre":  1,
	"millilitres": 1,
	"cl":          10,
	"centiliter":  10,
	"centiliters": 10,
	"centilitre":  10,
	"centilitres": 10,
	"dl":          100,
	"deciliter":   100,
	"deciliters":  100,
	"decilitre":   100,
	"decilitres":  100,
	"l":           1000,
	"liter":       1000,
	"liters":      1000,
	"litre":       1000,
	"litres":      1000,
}

// Weight units in grams (base unit)
var weightUnits = map[string]float64{
	// US Customary
	"oz":     28.3495,
	"ounce":  28.3495,
	"ounces": 28.3495,
	"lb":     453.592,
	"lbs":    453.592,
	"pound":  453.592,
	"pounds": 453.592,

	// Metric
	"mg":         0.001,
	"milligram":  0.001,
	"milligrams": 0.001,
	"g":          1,
	"gram":       1,
	"grams":      1,
	"kg":         1000,
	"kilogram":   1000,
	"kilograms":  1000,
	"kilo":       1000,
	"kilos":      1000,
}

// Special cooking ingredients with their conversions
// Maps ingredient name to grams per cup (or per unit for items like "stick")
var ingredientDensities = map[string]float64{
	// Fats
	"butter":        227, // grams per cup
	"margarine":     227,
	"oil":           218,
	"vegetable oil": 218,
	"olive oil":     216,
	"coconut oil":   218,

	// Flours
	"flour":             125,
	"all-purpose flour": 125,
	"ap flour":          125,
	"bread flour":       127,
	"cake flour":        114,
	"whole wheat flour": 120,
	"almond flour":      96,
	"coconut flour":     112,

	// Sugars
	"sugar":               200,
	"granulated sugar":    200,
	"white sugar":         200,
	"brown sugar":         220,
	"powdered sugar":      120,
	"confectioners sugar": 120,
	"icing sugar":         120,
	"honey":               340,
	"maple syrup":         322,
	"molasses":            328,
	"corn syrup":          328,

	// Dairy
	"milk":         244,
	"whole milk":   244,
	"skim milk":    245,
	"cream":        238,
	"heavy cream":  238,
	"sour cream":   242,
	"yogurt":       245,
	"cream cheese": 232,

	// Grains
	"rice":        185,
	"white rice":  185,
	"brown rice":  190,
	"oats":        80,
	"rolled oats": 80,
	"quinoa":      170,

	// Nuts & Seeds
	"almonds":         143,
	"walnuts":         120,
	"pecans":          109,
	"peanuts":         146,
	"cashews":         137,
	"sunflower seeds": 140,
	"chia seeds":      170,
	"flax seeds":      150,

	// Other
	"cocoa powder":  85,
	"cocoa":         85,
	"cornstarch":    128,
	"baking powder": 230,
	"baking soda":   220,
	"salt":          288,
	"yeast":         128,
}

// Special unit conversions (not volume or weight based)
var specialUnits = map[string]struct {
	toGrams float64
	toML    float64
	desc    string
}{
	"stick":           {113.4, 118.3, "butter"}, // 1 stick butter = 113.4g = 1/2 cup
	"sticks":          {113.4, 118.3, "butter"},
	"stick butter":    {113.4, 118.3, "butter"},
	"stick of butter": {113.4, 118.3, "butter"},
	"cube":            {113.4, 118.3, "butter"}, // Same as stick
	"cube butter":     {113.4, 118.3, "butter"},
	"pat":             {4.5, 4.7, "butter"}, // Small pat of butter
	"pat butter":      {4.5, 4.7, "butter"},
	"knob":            {14, 14.6, "butter"}, // Knob of butter ~1 tbsp
	"knob butter":     {14, 14.6, "butter"},
}

// Temperature conversions are handled separately
var tempPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)^(\d+(?:\.\d+)?)\s*(?:°|degrees?)?\s*(?:f|fahrenheit)\s+(?:to|in|as)\s+(?:°|degrees?)?\s*(?:c|celsius|centigrade)$`),
	regexp.MustCompile(`(?i)^(\d+(?:\.\d+)?)\s*(?:°|degrees?)?\s*(?:c|celsius|centigrade)\s+(?:to|in|as)\s+(?:°|degrees?)?\s*(?:f|fahrenheit)$`),
	regexp.MustCompile(`(?i)^(\d+(?:\.\d+)?)\s*(?:°|degrees?)?\s*(?:f|fahrenheit)\s+(?:to|in|as)\s+gas\s*(?:mark)?$`),
	regexp.MustCompile(`(?i)^gas\s*(?:mark)?\s*(\d+)\s+(?:to|in|as)\s+(?:°|degrees?)?\s*(?:f|fahrenheit|c|celsius)$`),
}

// IsCookingExpression checks if an expression is a cooking conversion
func IsCookingExpression(expr string) bool {
	expr = strings.TrimSpace(strings.ToLower(expr))

	// Check for temperature patterns
	for _, p := range tempPatterns {
		if p.MatchString(expr) {
			return true
		}
	}

	// Check for special units (stick, pat, etc.)
	for unit := range specialUnits {
		pattern := fmt.Sprintf(`(?i)^\d+(?:\.\d+)?\s*%s`, regexp.QuoteMeta(unit))
		if matched, _ := regexp.MatchString(pattern, expr); matched {
			return true
		}
	}

	// Check for ingredient-based conversions
	// e.g., "1 cup flour to grams", "200g butter to cups"
	ingredientPattern := `(?i)^\d+(?:\.\d+)?\s*(?:cups?|tbsp|tablespoons?|tsp|teaspoons?|g|grams?|oz|ounces?|ml|kg)\s+(?:of\s+)?(\w+(?:\s+\w+)?)\s+(?:to|in|as)\s+`
	if matched, _ := regexp.MatchString(ingredientPattern, expr); matched {
		return true
	}

	// Check for volume-to-volume cooking conversions
	// Only match cooking-specific patterns to avoid conflicts with units package
	cookingVolumePattern := `(?i)^\d+(?:\.\d+)?\s*(?:cups?|tbsp|tablespoons?|tsp|teaspoons?|fl\s*oz|fluid\s*oz|pints?|quarts?)\s+(?:to|in|as)\s+(?:cups?|tbsp|tablespoons?|tsp|teaspoons?|fl\s*oz|fluid\s*oz|pints?|quarts?|ml|l|liters?|litres?)$`
	if matched, _ := regexp.MatchString(cookingVolumePattern, expr); matched {
		return true
	}

	// Check for gas mark
	if matched, _ := regexp.MatchString(`(?i)gas\s*mark`, expr); matched {
		return true
	}

	return false
}

// EvalCooking evaluates a cooking conversion expression
func EvalCooking(expr string) (string, error) {
	expr = strings.TrimSpace(expr)
	exprLower := strings.ToLower(expr)

	// Handle temperature conversions
	for i, p := range tempPatterns {
		if matches := p.FindStringSubmatch(expr); matches != nil {
			return handleTempConversion(matches, i)
		}
	}

	// Handle gas mark conversions
	if matched, _ := regexp.MatchString(`(?i)gas\s*mark`, exprLower); matched {
		return handleGasMark(expr)
	}

	// Handle special units (stick, pat, etc.)
	for unit, conv := range specialUnits {
		pattern := regexp.MustCompile(fmt.Sprintf(`(?i)^(\d+(?:\.\d+)?)\s*%s(?:\s+(?:of\s+)?butter)?\s*(?:(?:to|in|as)\s+(\w+))?$`, regexp.QuoteMeta(unit)))
		if matches := pattern.FindStringSubmatch(expr); matches != nil {
			return handleSpecialUnit(matches, unit, conv)
		}
	}

	// Handle ingredient-based conversions
	ingredientPattern := regexp.MustCompile(`(?i)^(\d+(?:\.\d+)?)\s*(cups?|tbsp|tablespoons?|tsp|teaspoons?|g|grams?|oz|ounces?|ml|kg)\s+(?:of\s+)?(\w+(?:\s+\w+)?(?:\s+\w+)?)\s+(?:to|in|as)\s+(\w+)$`)
	if matches := ingredientPattern.FindStringSubmatch(expr); matches != nil {
		return handleIngredientConversion(matches)
	}

	// Handle volume-to-volume conversions
	volumePattern := regexp.MustCompile(`(?i)^(\d+(?:\.\d+)?)\s*(\w+(?:\s+\w+)?)\s+(?:to|in|as)\s+(\w+(?:\s+\w+)?)$`)
	if matches := volumePattern.FindStringSubmatch(expr); matches != nil {
		return handleVolumeConversion(matches)
	}

	return "", fmt.Errorf("unrecognized cooking expression")
}

func handleTempConversion(matches []string, patternIndex int) (string, error) {
	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return "", err
	}

	switch patternIndex {
	case 0: // F to C
		celsius := (value - 32) * 5 / 9
		return fmt.Sprintf("%.0f°F = %.0f°C", value, celsius), nil
	case 1: // C to F
		fahrenheit := value*9/5 + 32
		return fmt.Sprintf("%.0f°C = %.0f°F", value, fahrenheit), nil
	case 2: // F to Gas Mark
		gasMark := fahrenheitToGasMark(value)
		return fmt.Sprintf("%.0f°F = Gas Mark %s", value, gasMark), nil
	case 3: // Gas Mark to F/C
		return handleGasMarkToTemp(matches)
	}

	return "", fmt.Errorf("unknown temperature conversion")
}

func fahrenheitToGasMark(f float64) string {
	// Gas mark conversion table
	gasMarks := []struct {
		mark string
		f    float64
	}{
		{"1/4", 225},
		{"1/2", 250},
		{"1", 275},
		{"2", 300},
		{"3", 325},
		{"4", 350},
		{"5", 375},
		{"6", 400},
		{"7", 425},
		{"8", 450},
		{"9", 475},
		{"10", 500},
	}

	// Find closest gas mark
	closest := gasMarks[0]
	minDiff := abs(f - closest.f)
	for _, gm := range gasMarks[1:] {
		diff := abs(f - gm.f)
		if diff < minDiff {
			minDiff = diff
			closest = gm
		}
	}
	return closest.mark
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func handleGasMark(expr string) (string, error) {
	// Parse gas mark number
	pattern := regexp.MustCompile(`(?i)gas\s*mark\s*(\d+(?:/\d+)?)\s*(?:(?:to|in|as)\s+(\w+))?`)
	matches := pattern.FindStringSubmatch(expr)
	if matches == nil {
		return "", fmt.Errorf("invalid gas mark expression")
	}

	markStr := matches[1]
	targetUnit := strings.ToLower(matches[2])

	// Gas mark to temperature table
	gasMarkToF := map[string]float64{
		"1/4": 225, "1/2": 250,
		"1": 275, "2": 300, "3": 325, "4": 350,
		"5": 375, "6": 400, "7": 425, "8": 450,
		"9": 475, "10": 500,
	}

	f, ok := gasMarkToF[markStr]
	if !ok {
		return "", fmt.Errorf("unknown gas mark: %s", markStr)
	}

	c := (f - 32) * 5 / 9

	if targetUnit == "" || targetUnit == "f" || targetUnit == "fahrenheit" {
		return fmt.Sprintf("Gas Mark %s = %.0f°F (%.0f°C)", markStr, f, c), nil
	} else if targetUnit == "c" || targetUnit == "celsius" {
		return fmt.Sprintf("Gas Mark %s = %.0f°C (%.0f°F)", markStr, c, f), nil
	}

	return fmt.Sprintf("Gas Mark %s = %.0f°F (%.0f°C)", markStr, f, c), nil
}

func handleGasMarkToTemp(matches []string) (string, error) {
	markStr := matches[1]
	gasMarkToF := map[string]float64{
		"1": 275, "2": 300, "3": 325, "4": 350,
		"5": 375, "6": 400, "7": 425, "8": 450,
		"9": 475, "10": 500,
	}

	f, ok := gasMarkToF[markStr]
	if !ok {
		return "", fmt.Errorf("unknown gas mark: %s", markStr)
	}

	c := (f - 32) * 5 / 9
	return fmt.Sprintf("Gas Mark %s = %.0f°F (%.0f°C)", markStr, f, c), nil
}

func handleSpecialUnit(matches []string, unit string, conv struct {
	toGrams float64
	toML    float64
	desc    string
}) (string, error) {
	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return "", err
	}

	targetUnit := strings.ToLower(matches[2])
	grams := value * conv.toGrams
	ml := value * conv.toML
	cups := ml / 236.588
	tbsp := ml / 14.7868

	unitName := unit
	if value != 1 && !strings.HasSuffix(unit, "s") {
		unitName = unit + "s"
	}

	if targetUnit == "" {
		// Show comprehensive conversion
		if value == 1 {
			return fmt.Sprintf("1 %s %s = %.1fg = %.0f ml = %.2f cups = %.1f tbsp",
				unit, conv.desc, grams, ml, cups, tbsp), nil
		}
		return fmt.Sprintf("%.0f %s %s = %.1fg = %.0f ml = %.2f cups = %.1f tbsp",
			value, unitName, conv.desc, grams, ml, cups, tbsp), nil
	}

	// Convert to specific unit
	switch targetUnit {
	case "g", "grams", "gram":
		return fmt.Sprintf("%.1fg", grams), nil
	case "oz", "ounces", "ounce":
		return fmt.Sprintf("%.2f oz", grams/28.3495), nil
	case "ml", "milliliters":
		return fmt.Sprintf("%.0f ml", ml), nil
	case "cups", "cup":
		return fmt.Sprintf("%.2f cups", cups), nil
	case "tbsp", "tablespoons", "tablespoon":
		return fmt.Sprintf("%.1f tbsp", tbsp), nil
	}

	return fmt.Sprintf("%.0f %s = %.1fg", value, unitName, grams), nil
}

func handleIngredientConversion(matches []string) (string, error) {
	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return "", err
	}

	fromUnit := strings.ToLower(matches[2])
	ingredient := strings.ToLower(strings.TrimSpace(matches[3]))
	toUnit := strings.ToLower(matches[4])

	// Get ingredient density (grams per cup)
	density, ok := ingredientDensities[ingredient]
	if !ok {
		// Try partial match
		for ing, d := range ingredientDensities {
			if strings.Contains(ing, ingredient) || strings.Contains(ingredient, ing) {
				density = d
				ok = true
				break
			}
		}
	}
	if !ok {
		return "", fmt.Errorf("unknown ingredient: %s", ingredient)
	}

	// Convert from source unit to grams
	var grams float64
	switch {
	case strings.HasPrefix(fromUnit, "cup"):
		grams = value * density
	case fromUnit == "tbsp" || strings.HasPrefix(fromUnit, "tablespoon"):
		grams = value * density / 16 // 16 tbsp per cup
	case fromUnit == "tsp" || strings.HasPrefix(fromUnit, "teaspoon"):
		grams = value * density / 48 // 48 tsp per cup
	case fromUnit == "g" || strings.HasPrefix(fromUnit, "gram"):
		grams = value
	case fromUnit == "kg":
		grams = value * 1000
	case fromUnit == "oz" || strings.HasPrefix(fromUnit, "ounce"):
		grams = value * 28.3495
	case fromUnit == "ml":
		// Approximate: use density ratio
		grams = value * density / 236.588
	default:
		return "", fmt.Errorf("unknown unit: %s", fromUnit)
	}

	// Convert grams to target unit
	switch {
	case strings.HasPrefix(toUnit, "cup"):
		cups := grams / density
		return fmt.Sprintf("%.2f cups", cups), nil
	case toUnit == "tbsp" || strings.HasPrefix(toUnit, "tablespoon"):
		tbsp := grams / density * 16
		return fmt.Sprintf("%.1f tbsp", tbsp), nil
	case toUnit == "tsp" || strings.HasPrefix(toUnit, "teaspoon"):
		tsp := grams / density * 48
		return fmt.Sprintf("%.1f tsp", tsp), nil
	case toUnit == "g" || strings.HasPrefix(toUnit, "gram"):
		return fmt.Sprintf("%.1fg", grams), nil
	case toUnit == "kg":
		return fmt.Sprintf("%.3f kg", grams/1000), nil
	case toUnit == "oz" || strings.HasPrefix(toUnit, "ounce"):
		return fmt.Sprintf("%.2f oz", grams/28.3495), nil
	case toUnit == "lb" || strings.HasPrefix(toUnit, "pound"):
		return fmt.Sprintf("%.2f lb", grams/453.592), nil
	case toUnit == "ml":
		ml := grams / density * 236.588
		return fmt.Sprintf("%.0f ml", ml), nil
	}

	return "", fmt.Errorf("unknown target unit: %s", toUnit)
}

func handleVolumeConversion(matches []string) (string, error) {
	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return "", err
	}

	fromUnit := normalizeUnit(strings.ToLower(matches[2]))
	toUnit := normalizeUnit(strings.ToLower(matches[3]))

	// Get conversion factors
	fromML, ok := volumeUnits[fromUnit]
	if !ok {
		return "", fmt.Errorf("unknown unit: %s", fromUnit)
	}

	toML, ok := volumeUnits[toUnit]
	if !ok {
		return "", fmt.Errorf("unknown unit: %s", toUnit)
	}

	// Convert
	ml := value * fromML
	result := ml / toML

	// Format output
	toUnitDisplay := toUnit
	if result != 1 && !strings.HasSuffix(toUnit, "s") && len(toUnit) > 2 {
		toUnitDisplay = toUnit + "s"
	}

	if result == float64(int(result)) {
		return fmt.Sprintf("%.0f %s", result, toUnitDisplay), nil
	}
	return fmt.Sprintf("%.2f %s", result, toUnitDisplay), nil
}

func normalizeUnit(unit string) string {
	// Handle multi-word units
	unit = strings.ReplaceAll(unit, " ", "")

	// Common aliases
	aliases := map[string]string{
		"tablespoon":  "tbsp",
		"tablespoons": "tbsp",
		"teaspoon":    "tsp",
		"teaspoons":   "tsp",
		"fluidoz":     "floz",
		"fluidounce":  "floz",
		"fluidounces": "floz",
		"milliliter":  "ml",
		"milliliters": "ml",
		"millilitre":  "ml",
		"millilitres": "ml",
		"liter":       "l",
		"liters":      "l",
		"litre":       "l",
		"litres":      "l",
	}

	if normalized, ok := aliases[unit]; ok {
		return normalized
	}
	return unit
}

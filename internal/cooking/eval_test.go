package cooking

import (
	"strings"
	"testing"
)

func TestIsCookingExpression(t *testing.T) {
	tests := []struct {
		expr     string
		expected bool
	}{
		// Volume conversions
		{"2 cups to tbsp", true},
		{"1 cup to tablespoons", true},
		{"3 tbsp to tsp", true},
		{"1 pint to cups", true},
		{"2 quarts to liters", true},

		// Special units
		{"1 stick butter", true},
		{"2 sticks to grams", true},
		{"1 pat butter", true},

		// Ingredient conversions
		{"1 cup flour to grams", true},
		{"200g butter to cups", true},
		{"1 cup sugar to grams", true},

		// Temperature
		{"350 f to c", true},
		{"180 c to f", true},
		{"350 fahrenheit to celsius", true},

		// Gas mark
		{"gas mark 4", true},
		{"gas mark 6 to f", true},

		// Non-cooking expressions
		{"2 + 2", false},
		{"hello world", false},
		{"100 meters to feet", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := IsCookingExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("IsCookingExpression(%q) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestVolumeConversions(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"2 cups to tbsp", "32"},
		{"1 cup to tablespoons", "16"},
		{"3 tbsp to tsp", "9"},
		{"1 tbsp to ml", "14"},
		{"1 cup to ml", "236"},
		{"1 pint to cups", "2"},
		{"1 quart to pints", "2"},
		{"1 gallon to quarts", "4"},
		{"500 ml to cups", "2.1"},
		{"1 liter to cups", "4.2"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalCooking(tt.expr)
			if err != nil {
				t.Errorf("EvalCooking(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("EvalCooking(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

func TestSpecialUnits(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"1 stick butter", "113"},
		{"2 sticks butter", "226"},
		{"1 stick to grams", "113"},
		{"1 pat butter", "4.5"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalCooking(tt.expr)
			if err != nil {
				t.Errorf("EvalCooking(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("EvalCooking(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

func TestIngredientConversions(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"1 cup flour to grams", "125"},
		{"1 cup sugar to grams", "200"},
		{"1 cup butter to grams", "227"},
		{"1 cup brown sugar to grams", "220"},
		{"1 cup honey to grams", "340"},
		{"1 cup rice to grams", "185"},
		{"1 cup oats to grams", "80"},
		{"250g flour to cups", "2"},
		{"100g butter to tbsp", "7"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalCooking(tt.expr)
			if err != nil {
				t.Errorf("EvalCooking(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("EvalCooking(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

func TestTemperatureConversions(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"350 f to c", "177"},
		{"180 c to f", "356"},
		{"400 fahrenheit to celsius", "204"},
		{"200 celsius to fahrenheit", "392"},
		{"32 f to c", "0"},
		{"100 c to f", "212"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalCooking(tt.expr)
			if err != nil {
				t.Errorf("EvalCooking(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("EvalCooking(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

func TestGasMarkConversions(t *testing.T) {
	tests := []struct {
		expr     string
		contains string
	}{
		{"gas mark 4", "350"},
		{"gas mark 6", "400"},
		{"gas mark 4 to f", "350"},
		{"gas mark 4 to c", "177"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := EvalCooking(tt.expr)
			if err != nil {
				t.Errorf("EvalCooking(%q) error: %v", tt.expr, err)
				return
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("EvalCooking(%q) = %q, want to contain %q", tt.expr, result, tt.contains)
			}
		})
	}
}

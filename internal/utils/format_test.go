package utils

import (
	"testing"
)

func TestAddThousandsSeparators(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1000", "1,000"},
		{"1000000", "1,000,000"},
		{"100", "100"},
		{"12345678", "12,345,678"},
		{"1", "1"},
		{"12", "12"},
		{"123", "123"},
		{"1234", "1,234"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := addThousandsSeparators(tt.input)
			if result != tt.expected {
				t.Errorf("addThousandsSeparators(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatNumberWithThousands(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{1000, "1,000"},
		{1000000, "1,000,000"},
		{1234.56, "1,234.56"},
		{100, "100"},
		{0.5, "0.5"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatNumberWithThousands(tt.input)
			if result != tt.expected {
				t.Errorf("formatNumberWithThousands(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatCurrency(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{100, "$100.00"},
		{1000, "$1,000.00"},
		{1234.56, "$1,234.56"},
		{1234.567, "$1,234.57"}, // rounds to 2 decimals
		{0.99, "$0.99"},
		{1000000, "$1,000,000.00"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatCurrency(tt.input)
			if result != tt.expected {
				t.Errorf("FormatCurrency(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatResult(t *testing.T) {
	tests := []struct {
		name       string
		isCurrency bool
		value      float64
		expected   string
	}{
		{"regular number", false, 1234, "1,234"},
		{"currency", true, 1234, "$1,234.00"},
		{"regular decimal", false, 1234.56, "1,234.56"},
		{"currency decimal", true, 1234.56, "$1,234.56"},
		{"small number", false, 5, "5"},
		{"small currency", true, 5, "$5.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatResult(tt.isCurrency, tt.value)
			if result != tt.expected {
				t.Errorf("FormatResult(%v, %v) = %q, want %q", tt.isCurrency, tt.value, result, tt.expected)
			}
		})
	}
}

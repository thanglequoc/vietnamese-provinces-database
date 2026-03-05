package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeString_ApostropheReplacement(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Replaces smart apostrophe with standard apostrophe",
			input:    "Xã Ea H'MLay",
			expected: "Xã Ea H'MLay",
		},
		{
			name:     "Multiple smart apostrophes",
			input:    "Ea H'MLay H'Lấk",
			expected: "Ea H'MLay H'Lấk",
		},
		{
			name:     "No apostrophes",
			input:    "Xã Bình Dương",
			expected: "Xã Bình Dương",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeString_NFCNormalization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normalizes decomposed characters to NFC",
			input:    "Xã", // decomposed: X + a + ̃
			expected: "Xã",  // precomposed
		},
		{
			name:     "Already in NFC form",
			input:    "Xã",
			expected: "Xã",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeString_ToneMarkCorrection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normalizes tone marks",
			input:    "Thành phố",
			expected: "Thành phố",
		},
		{
			name:     "Combined normalization",
			input:    "Xã Ea H'MLay",
			expected: "Xã Ea H'MLay",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoveUnitPrefix_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name            string
		fullName        string
		unitName        string
		expected        string
	}{
		{
			name:     "Removes Phường prefix (lowercase)",
			fullName: "phường Bắc Sơn",
			unitName: "Phường",
			expected: "Bắc Sơn",
		},
		{
			name:     "Removes Xã prefix (uppercase)",
			fullName: "XÃ BÌNH DƯƠNG",
			unitName: "Xã",
			expected: "BÌNH DƯƠNG",
		},
		{
			name:     "Removes Thị trấn prefix (mixed case)",
			fullName: "Thị trấn Châu Thành",
			unitName: "Thị trấn",
			expected: "Châu Thành",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeUnitPrefix(tt.fullName, tt.unitName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoveUnitPrefix_NoPrefix(t *testing.T) {
	tests := []struct {
		name     string
		fullName string
		unitName string
		expected string
	}{
		{
			name:     "No unit prefix present",
			fullName: "Bắc Sơn",
			unitName: "Phường",
			expected: "Bắc Sơn",
		},
		{
			name:     "Empty string",
			fullName: "",
			unitName: "Phường",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeUnitPrefix(tt.fullName, tt.unitName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCapitalizeWords_RuneAware(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Capitalizes first letter of each word",
			input:    "bắc sơn",
			expected: "Bắc Sơn",
		},
		{
			name:     "Single word",
			input:    "bình",
			expected: "Bình",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Multiple spaces",
			input:    "bắc  sơn",
			expected: "Bắc Sơn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := capitalizeWords(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCapitalizeWords_PreservesInternalCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Capitalizes first letter of each word - h'leo",
			input:    "h'leo",
			expected: "H'leo",
		},
		{
			name:     "Capitalizes first letter of each word - n'iêng",
			input:    "n'iêng",
			expected: "N'iêng",
		},
		{
			name:     "Capitalizes first letter of each word - bắc sôn",
			input:    "bắc sôn",
			expected: "Bắc Sôn",  // Note: normalizeString is called after capitalizeWords
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := capitalizeWords(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

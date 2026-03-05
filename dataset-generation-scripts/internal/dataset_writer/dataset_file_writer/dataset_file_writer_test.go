package dataset_writer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscapeSingleQuote(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Single quote becomes double quote",
			input:    "Ea H'MLay",
			expected: "Ea H''MLay",
		},
		{
			name:     "Multiple single quotes",
			input:    "H'Leo - N'Iêng",
			expected: "H''Leo - N''Iêng",
		},
		{
			name:     "No single quotes",
			input:    "Bắc Sơn",
			expected: "Bắc Sơn",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Single quote at start",
			input:    "'Start",
			expected: "''Start",
		},
		{
			name:     "Single quote at end",
			input:    "End'",
			expected: "End''",
		},
		{
			name:     "Only single quote",
			input:    "'",
			expected: "''",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeSingleQuote(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseEuropeanFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
		wantErr  bool
	}{
		{
			name:     "Standard European format with comma",
			input:    "1.234,56",
			expected: 1234.56,
			wantErr:  false,
		},
		{
			name:     "European format with thousands separator",
			input:    "12.345,67",
			expected: 12345.67,
			wantErr:  false,
		},
		{
			name:     "Small number",
			input:    "0,99",
			expected: 0.99,
			wantErr:  false,
		},
		{
			name:     "Integer without decimals",
			input:    "1.000,00",
			expected: 1000.0,
			wantErr:  false,
		},
		{
			name:     "No thousands separator",
			input:    "123,45",
			expected: 123.45,
			wantErr:  false,
		},
		{
			name:     "Very large number",
			input:    "1.000.000,00",
			expected: 1000000.0,
			wantErr:  false,
		},
		{
			name:     "Negative number",
			input:    "-1.234,56",
			expected: -1234.56,
			wantErr:  false,
		},
		{
			name:     "Zero",
			input:    "0,00",
			expected: 0.0,
			wantErr:  false,
		},
		{
			name:     "Dot as decimal (invalid - converts to large number)",
			input:    "1234.56",
			expected: 123456, // "1234.56" → "123456" after removing dot
			wantErr:  false,
		},
		{
			name:     "Invalid format - letters",
			input:    "abc",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseEuropeanFloat(tt.input)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, tt.expected, result, 0.01)
			}
		})
	}
}

func TestGetFileTimeSuffix(t *testing.T) {
	// Note: This test is somewhat fragile as it depends on current time
	// We're mainly testing that the format is correct
	result := getFileTimeSuffix()
	
	// Check that the result doesn't contain spaces or colons
	assert.NotContains(t, result, " ", "should not contain spaces")
	assert.NotContains(t, result, ":", "should not contain colons")
	
	// Check that it contains underscore separator
	assert.Contains(t, result, "__", "should contain double underscore as separator")
	
	// Check that it contains underscore for time separator
	assert.Contains(t, result, "_", "should contain underscore for time separator")
}
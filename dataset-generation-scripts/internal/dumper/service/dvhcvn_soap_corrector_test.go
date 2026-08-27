package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCorrectDvhcvnSoapData_UnitCodeCorrection(t *testing.T) {
	tests := []struct {
		name         string
		unitCode     string
		unitName     string
		expectedCode string
		expectedName string
	}{
		{
			name:         "No correction for unknown code",
			unitCode:     "99999",
			unitName:     "Xã Unknown",
			expectedCode: "99999",
			expectedName: "Xã Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := correctDvhcvnSoapData(tt.unitCode, tt.unitName)

			assert.Equal(t, tt.expectedCode, result.code, "unit code should match expected")
			assert.Equal(t, tt.expectedName, result.name, "unit name should match expected")
		})
	}
}

func TestCorrectDvhcvnSoapData_NoCorrectionNeeded(t *testing.T) {
	tests := []struct {
		name     string
		unitCode string
		unitName string
	}{
		{
			name:     "Unlisted code remains unchanged",
			unitCode: "12345",
			unitName: "Xã Bình Dương",
		},
		{
			name:     "Another unlisted code",
			unitCode: "54321",
			unitName: "Phường Tân Bình",
		},
		{
			name:     "Ba Chẽ ward code remains unchanged",
			unitCode: "06970",
			unitName: "Xã Ba Chẽ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := correctDvhcvnSoapData(tt.unitCode, tt.unitName)

			assert.Equal(t, tt.unitCode, result.code, "code should remain unchanged")
			assert.Equal(t, tt.unitName, result.name, "name should remain unchanged")
		})
	}
}

func TestGetEnglishNameOverride_AllOverrides(t *testing.T) {
	tests := []struct {
		name         string
		unitCode     string
		expectedName string
		expectedOk   bool
	}{
		{
			name:         "Hà Nội → Hanoi",
			unitCode:     "01",
			expectedName: "Hanoi",
			expectedOk:   true,
		},
		{
			name:         "Hải Phòng → Haiphong",
			unitCode:     "31",
			expectedName: "Haiphong",
			expectedOk:   true,
		},
		{
			name:         "Hoàng Sa → Paracel",
			unitCode:     "20333",
			expectedName: "Paracel",
			expectedOk:   true,
		},
		{
			name:         "Trường Sa → Spratly",
			unitCode:     "22736",
			expectedName: "Spratly",
			expectedOk:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, ok := getEnglishNameOverride(tt.unitCode)

			assert.Equal(t, tt.expectedOk, ok, "ok should match expected")
			assert.Equal(t, tt.expectedName, name, "name should match expected")
		})
	}
}

func TestGetEnglishNameOverride_UnknownCode(t *testing.T) {
	tests := []struct {
		name     string
		unitCode string
	}{
		{
			name:     "Unknown province code returns empty",
			unitCode: "99",
		},
		{
			name:     "Unknown ward code returns empty",
			unitCode: "99999",
		},
		{
			name:     "Empty code returns empty",
			unitCode: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, ok := getEnglishNameOverride(tt.unitCode)

			assert.False(t, ok, "unknown code should return ok=false")
			assert.Empty(t, name, "unknown code should return empty name")
		})
	}
}

func TestEnglishNameOverrides_NotEmpty(t *testing.T) {
	assert.NotEmpty(t, englishNameOverrides, "englishNameOverrides map should not be empty")
}
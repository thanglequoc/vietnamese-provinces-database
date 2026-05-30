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
			name:         "Corrects Ba Chẽ ward code from 06970 to 06978",
			unitCode:     "06970",
			unitName:     "Xã Ba Chẽ",
			expectedCode: "06978",
			expectedName: "Xã Ba Chẽ",
		},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := correctDvhcvnSoapData(tt.unitCode, tt.unitName)

			assert.Equal(t, tt.unitCode, result.code, "code should remain unchanged")
			assert.Equal(t, tt.unitName, result.name, "name should remain unchanged")
		})
	}
}

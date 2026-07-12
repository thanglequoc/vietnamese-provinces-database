package service

import (
	"fmt"
	"strings"
)

// correctionRule defines how to correct a specific unit's data
type correctionRule struct {
	search        string // text to search for in the name
	replace       string // replacement text for the name
	correctedCode string // if non-empty, use this as the corrected unit code
}

// unitCorrections maps unit codes to their specific correction rules
// This makes it easy to add more corrections as they are discovered
var unitCorrections = map[string]correctionRule{
	// Ref: https://luatvietnam.vn/hanh-chinh/quyet-dinh-19-2025-qd-ttg-cua-thu-tuong-chinh-phu-ban-hanh-bang-danh-muc-va-ma-so-cac-don-vi-hanh-chinh-viet-nam-404751-d1.html
	// "00745": {search: "Phố Bảng", replace: "Phó Bảng", correctedCode: ""},
	"06970": {search: "", replace: "", correctedCode: "06978"}, // Ba Chẽ ward - correct code from 06970 to 06978
}

// englishNameOverrides maps unit codes to their standard English short names.
// These override the auto-generated English name (which is produced by stripping
// Vietnamese tone marks) for units that have internationally recognized English names.
// Example: "Hà Nội" → "Hanoi" instead of "Ha Noi"
var englishNameOverrides = map[string]string{
	"01":    "Hanoi",
	"31":    "Haiphong",
	"20333": "Paracel",
	"22736": "Spratly",
}

// getEnglishNameOverride returns the standard English short name for a unit,
// or empty string and false if no override exists.
func getEnglishNameOverride(unitCode string) (string, bool) {
	name, exists := englishNameOverrides[unitCode]
	return name, exists
}

// correctionResult holds the corrected data for a unit
type correctionResult struct {
	code string
	name string
}

// correctDvhcvnSoapData applies corrections to unit codes and names from SOAP data
// Uses a map-based approach for efficient lookups and easy maintenance
// Returns both the corrected unit code and corrected name
func correctDvhcvnSoapData(unitCode string, name string) correctionResult {
	result := correctionResult{
		code: unitCode,
		name: name,
	}

	// Check if there's a correction rule for this unit code
	if rule, exists := unitCorrections[unitCode]; exists {
		hasChanges := false

		// Apply name correction if search and replace are specified
		if rule.search != "" && rule.replace != "" {
			result.name = strings.Replace(result.name, rule.search, rule.replace, 1)
			hasChanges = true
		}

		// Apply unit code correction if specified
		if rule.correctedCode != "" {
			result.code = rule.correctedCode
			hasChanges = true
		}

		// Log warning if any correction was applied
		if hasChanges {
			fmt.Printf("⚠️  Data correction applied: [%s] '%s' → [%s] '%s'\n",
				unitCode, name, result.code, result.name)
		}
	}

	return result
}
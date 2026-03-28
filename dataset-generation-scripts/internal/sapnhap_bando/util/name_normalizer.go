package util

import (
	"strings"

	"github.com/thanglequoc-vn-provinces/v2/internal/common/viet"
	"golang.org/x/text/unicode/norm"
)

// RemoveAdministrativeUnitPrefix removes common administrative unit prefixes from a name
// Based on the mapping from administrative_units table
// Examples:
// - "Thủ đô Hà Nội" → "Hà Nội"
// - "Thành phố Hà Nội" → "Hà Nội"
// - "Tỉnh Quảng Ninh" → "Quảng Ninh"
func RemoveAdministrativeUnitPrefix(name string) string {
	lowerName := strings.ToLower(strings.TrimSpace(name))
	
	// Define prefixes to remove (matching the administrative unit types)
	prefixes := []string{
		"thủ đô ",    // Thủ đô (Capital city)
		"tỉnh ",      // Tỉnh (Province)
		"thành phố ", // Thành phố (City)
		"đặc khu ",   // Đặc khu (Special zone)
		"thị xã ",     // Thị xã (Township)
		"thị trấn ",  // Thị trấn (Townlet)
		"quận ",       // Quận (District)
		"huyện ",      // Huyện (County)
		"phường ",     // Phường (Ward)
		"xã ",         // Xã (Commune)
	}
	
	for _, prefix := range prefixes {
		if strings.HasPrefix(lowerName, prefix) {
			// Remove the prefix
			result := strings.TrimSpace(name[len(prefix):])
			return result
		}
	}
	
	return name // Return original if no prefix found
}

// NormalizeForMatching normalizes Vietnamese names for comparison by:
// 1. Removing administrative unit prefixes (e.g., "Thủ đô Hà Nội" → "Hà Nội")
// 2. Normalizing tone marks (preserving differences, e.g., "Hà Nội" → "hà nội")
// 3. Converting to lowercase for case-insensitive matching
// 4. Trimming whitespace
// This preserves tone mark differences to distinguish between similar names
func NormalizeForMatching(name string) string {
	// First, remove administrative unit prefix
	normalized := RemoveAdministrativeUnitPrefix(name)
	
	// Normalize tone marks to standard form (preserves differences between similar names)
	normalized = viet.NormalizeToneMarks(normalized)
	
	// Convert to lowercase for case-insensitive matching
	normalized = strings.ToLower(normalized)
	normalized = strings.ReplaceAll(normalized, "’", "'")
	
	// Trim whitespace
	return strings.TrimSpace(normalized)
}

// NormalizeString handles all normalization steps for Vietnamese text
// 1. Replaces smart apostrophes with standard apostrophes
// 2. Normalizes to NFC form to handle decomposed Unicode characters
// This ensures consistent encoding across all data
func NormalizeString(s string) string {
	// Replace smart apostrophe with standard apostrophe
	result := strings.ReplaceAll(s, "’", "'")
	result = viet.NormalizeToneMarks(result)
	
	// Normalize to NFC form to handle decomposed characters
	// This ensures "X ̃" (decomposed) becomes "Xã" (precomposed)
	result = norm.NFC.String(result)
	
	return result
}

package service

import "strings"

// correctionRule defines how to correct a specific unit's name
type correctionRule struct {
	search  string
	replace string
}

/*
	The SOAP data source from DVHCHVN contains some incorrect unit names, possibly due to old-name reference, or typo errors.
	This mapping ensure these faulty names are corrected to match the actual names from the official decrees document.
*/
var unitCorrections = map[string]correctionRule{
	// Phó Bảng (Tuyên Quang)
	"00745": {search: "Phố Bảng", replace: "Phó Bảng"},
}

// correctDvhcvnSoapData applies corrections to unit names from the SOAP data
// Uses a map-based approach for efficient lookups and easy maintenance
func correctDvhcvnSoapData(unitCode string, name string) string {
	// Check if there's a correction rule for this unit code
	if rule, exists := unitCorrections[unitCode]; exists {
		name = strings.Replace(name, rule.search, rule.replace, 1)
	}
	
	return name
}

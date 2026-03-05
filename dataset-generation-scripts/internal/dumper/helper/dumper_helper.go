package helper

import (
	"strings"
	"github.com/thanglequoc-vn-provinces/v2/internal/common/viet"
)

/*
Determine the province administrative unit id from its name
*/
func GetAdministrativeUnit_ProvinceLevel(provinceFullName string) int {
	normalized := strings.ToLower(viet.RemoveVietToneMark(provinceFullName))
	if strings.HasPrefix(normalized, "thanh pho") {
		return 1
	}
	if strings.HasPrefix(normalized, "tinh") {
		return 2
	}
	panic("Unable to determine administrative unit name from province: " + provinceFullName)
}

/*
Determine the ward administrative unit id from its name
*/
func GetAdministrativeUnit_WardLevel(wardFullName string) int {
	// Normalize to handle precomposed vs decomposed characters
	normalized := strings.ToLower(viet.RemoveVietToneMark(wardFullName))
	
	if strings.HasPrefix(normalized, "phuong") {
		return 3
	}
	if strings.HasPrefix(normalized, "xa") {
		return 4
	}
	if strings.HasPrefix(normalized, "dac khu") {
		return 5
	}
	panic("Unable to determine administrative unit name from ward: " + wardFullName)
}

/*
Generate code name from the name
*/
func ToCodeName(shortName string) string {
	shortName = strings.ReplaceAll(shortName, " - ", " ")
	shortName = strings.ReplaceAll(shortName, "'", "") // to handle special name with single quote
	return strings.ToLower(strings.ReplaceAll(viet.RemoveVietToneMark(shortName), " ", "_"))
}

func RemoveWhiteSpaces(name string) string {
	return strings.Trim(strings.ReplaceAll(name, "  ", " "), " ")
}

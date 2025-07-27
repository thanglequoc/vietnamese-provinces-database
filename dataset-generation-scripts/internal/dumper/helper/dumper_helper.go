package helper

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

/*
Determine the province administrative unit id from its name
*/
func GetAdministrativeUnit_ProvinceLevel(provinceFullName string) int {
	if strings.HasPrefix(provinceFullName, "Thành phố") {
		return 1
	}
	if strings.HasPrefix(provinceFullName, "Tỉnh") {
		return 2
	}
	panic("Unable to determine administrative unit name from province: " + provinceFullName)
}

/*
Determine the ward administrative unit id from its name
*/
func GetAdministrativeUnit_WardLevel(wardFullName string) int {
	if strings.HasPrefix(wardFullName, "Phường") {
		return 3
	}
	if strings.HasPrefix(wardFullName, "Xã") {
		return 4
	}
	if strings.HasPrefix(wardFullName, "Đặc khu") {
		return 5
	}
	panic("Unable to determine administrative unit name from ward: " + wardFullName)
}

/*
Normalize string to remove Vietnamese special character and sign
*/
func NormalizeString(source string) string {
	trans := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(trans, source)
	result = strings.ReplaceAll(result, "đ", "d")
	result = strings.ReplaceAll(result, "Đ", "D")
	return result
}

/*
Generate code name from the name
*/
func ToCodeName(shortName string) string {
	shortName = strings.ReplaceAll(shortName, " - ", " ")
	shortName = strings.ReplaceAll(shortName, "'", "") // to handle special name with single quote
	return strings.ToLower(strings.ReplaceAll(NormalizeString(shortName), " ", "_"))
}

func RemoveWhiteSpaces(name string) string {
	return strings.Trim(strings.ReplaceAll(name, "  ", " "), " ")
}

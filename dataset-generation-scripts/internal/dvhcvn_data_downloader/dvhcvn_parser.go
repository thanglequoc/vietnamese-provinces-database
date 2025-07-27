package dvhcvn_data_downloader

import (
	"github.com/dlclark/regexp2"
	"regexp"
	"strings"
)

/*
Create the DvhcvnProvinceModel from the dvhcvn response object.
Input should be the dvhcvn content within the <TABLE> tag
E.g: <MaTinh>94</MaTinh><TenTinh>Tỉnh Sóc Trăng</TenTinh><LoaiHinh>Tỉnh</LoaiHinh>
*/
func toDvhcvnProvinceModel(s string) DvhcvnProvinceModel {
	proviceCode := extractRegexValue(`<MaTinh>(.+)</MaTinh>`, s)
	proviceName := extractRegexValue(`<TenTinh>(.+)</TenTinh>`, s)
	unit := extractRegexValue(`<LoaiHinh>(.+)</LoaiHinh>`, s)

	return DvhcvnProvinceModel{
		ProvinceCode: proviceCode,
		ProvinceName: convertStandardUnitName(proviceName),
		Unit:         unit,
	}
}

/*
Create the toDvhcvnDistrictModel from the dvhcvn response object.
Input should be the dvhcvn content within the <TABLE> tag
E.g: <MaTinh>01</MaTinh><TenTinh>Thành phố Hà Nội</TenTinh><MaQuanHuyen>003</MaQuanHuyen><TenQuanHuyen>Quận Tây Hồ</TenQuanHuyen><LoaiHinh>Quận</LoaiHinh>
*/
func toDvhcvnDistrictModel(s string) DvhcvnDistrictModel {
	proviceCode := extractRegexValue(`<MaTinh>(.+)</MaTinh>`, s)
	proviceName := extractRegexValue(`<TenTinh>(.+)</TenTinh>`, s)
	districtCode := extractRegexValue(`<MaQuanHuyen>(.+)</MaQuanHuyen>`, s)
	districtName := extractRegexValue(`<TenQuanHuyen>(.+)</TenQuanHuyen>`, s)
	unit := extractRegexValue(`<LoaiHinh>(.+)</LoaiHinh>`, s)

	return DvhcvnDistrictModel{
		ProvinceCode: proviceCode,
		ProvinceName: convertStandardUnitName(proviceName),
		DistrictCode: districtCode,
		DistrictName: convertStandardUnitName(districtName),
		Unit:         unit,
	}
}

/*
Create the DvhcvnWardModel from the dvhcvn response object.
Input should be the dvhcvn content within the <TABLE> tag
E.g: <MaTinh>74</MaTinh><TenTinh>Tỉnh Bình Dương</TenTinh><MaQuanHuyen>721</MaQuanHuyen><TenQuanHuyen>Thành phố Bến Cát</TenQuanHuyen><MaPhuongXa>25843</MaPhuongXa><TenPhuongXa>Phường An Tây</TenPhuongXa><LoaiHinh>Phường</LoaiHinh>
*/
func toDvhcvnWardModel(s string) DvhcvnWardModel {
	proviceCode := extractRegexValue(`<MaTinh>(.+)</MaTinh>`, s)
	proviceName := extractRegexValue(`<TenTinh>(.+)</TenTinh>`, s)
	districtCode := extractRegexValue(`<MaQuanHuyen>(.+)</MaQuanHuyen>`, s)
	districtName := extractRegexValue(`<TenQuanHuyen>(.+)</TenQuanHuyen>`, s)
	wardCode := extractRegexValue(`<MaPhuongXa>(.+)</MaPhuongXa>`, s)
	wardName := extractRegexValue(`<TenPhuongXa>(.+)</TenPhuongXa>`, s)

	unit := extractRegexValue(`<LoaiHinh>(.+)</LoaiHinh>`, s)

	return DvhcvnWardModel{
		ProvinceCode: proviceCode,
		ProvinceName: convertStandardUnitName(proviceName),
		DistrictCode: districtCode,
		DistrictName: convertStandardUnitName(districtName),
		WardCode:     wardCode,
		WardName:     convertStandardUnitName(wardName),
		Unit:         unit,
	}
}

/*
Standardize Adminstrative Unit name to follow correct Vietnamese typography
E.g: Thành phố, instead of Thành Phố
This method is put to use because some district/ward result from DVHCVN API does not follow the unify typography pattern
*/
func convertStandardUnitName(s string) string {
	if strings.HasPrefix(s, "Thành Phố") {
		return strings.Replace(s, "Thành Phố", "Thành phố", 1)
	}
	if strings.HasPrefix(s, "Thị Xã") {
		return strings.Replace(s, "Thị Xã", "Thị xã", 1)
	}
	if strings.HasPrefix(s, "Thị Trấn") {
		return strings.Replace(s, "Thị Trấn", "Thị trấn", 1)
	}
	return s
}

func regexp2FindAllString(re *regexp2.Regexp, s string) []string {
	var matches []string
	m, _ := re.FindStringMatch(s)
	for m != nil {
		matches = append(matches, m.String())
		m, _ = re.FindNextMatch(m)
	}
	return matches
}

func sanitizeString(s string) string {
	return strings.Trim(
		strings.ReplaceAll(s, "  ", " "), " ")
}

func extractRegexValue(pattern string, s string) string {
	regex := regexp.MustCompile(pattern)
	if match := regex.FindStringSubmatch(s); match != nil {
		return sanitizeString(match[1])
	}
	return ""
}

/*
Extract the XML response from DVHCVN api to DvhcvnProvinceModel
*/
func extractProvinceDvhcvnUnits(soapResponse string) []DvhcvnProvinceModel {
	regexPattern := regexp2.MustCompile(`(?<=<TABLE\b[^>]*>)([\s\S\n]*?)(?=<\/TABLE>)`, 0)
	dvhcvnUnitBlocks := regexp2FindAllString(regexPattern, soapResponse)

	var result []DvhcvnProvinceModel
	for _, unit := range dvhcvnUnitBlocks {
		result = append(result, toDvhcvnProvinceModel(unit))
	}
	return result
}

/*
Extract the XML response from DVHCVN api to DvhcvnDistrictModel
*/
func extractDistrictDvhcvnUnits(soapResponse string) []DvhcvnDistrictModel {
	regexPattern := regexp2.MustCompile(`(?<=<TABLE\b[^>]*>)([\s\S\n]*?)(?=<\/TABLE>)`, 0)
	dvhcvnUnitBlocks := regexp2FindAllString(regexPattern, soapResponse)

	var result []DvhcvnDistrictModel
	for _, unit := range dvhcvnUnitBlocks {
		result = append(result, toDvhcvnDistrictModel(unit))
	}
	return result
}

/*
Extract the XML response from DVHCVN api to DvhcvnWardModel
*/
func extractWardDvhcvnUnits(soapResponse string) []DvhcvnWardModel {
	regexPattern := regexp2.MustCompile(`(?<=<TABLE\b[^>]*>)([\s\S\n]*?)(?=<\/TABLE>)`, 0)
	dvhcvnUnitBlocks := regexp2FindAllString(regexPattern, soapResponse)

	var result []DvhcvnWardModel
	for _, unit := range dvhcvnUnitBlocks {
		result = append(result, toDvhcvnWardModel(unit))
	}
	return result
}

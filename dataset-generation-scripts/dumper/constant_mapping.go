package dumper


var AdministrativeUnitNames [5]string = [...]string{
	"Thành phố",
	"Tỉnh",
	"Phường",
	"Xã",
	"Đặc khu",
}

/*
Short name definition of Administrative Unit from top to bottom
- Municipality (Thành phố trực thuộc trung ương)
- Province (Tỉnh)
- Commune (Xã)
- Special administrative region (Đặc khu)
*/
var AdministrativeUnitNamesShortNameMap_vn = map[int]string{
	1: "Thành phố",
	2: "Tỉnh",
	3: "Phường",
	4: "Xã",
	5: "Đặc khu",
}
var AdministrativeUnitNamesShortNameMap_en = map[int]string{
	1: "City",
	2: "Province",
	3: "Ward",
	4: "Commune",
	5: "Special administrative region",
}

/*
Mapping definition for the province code and the region that it belongs to
Define as constant mapping since the province geographical data would likely to be never changed

NOTE: After the major province merge, ProvinceRegionMap is no longer accurate as one province may span across 3 regions now
E.g: Vĩnh Phúc, Phú Thọ, Hòa Bình => Phú Thọ
URL: https://vi.wikipedia.org/wiki/Sáp_nhập_tỉnh,_thành_Việt_Nam_2025
*/

var ProvinceRegionMap = map[string]int {
	"01": 3,
	"26": 3,
	"27": 3,
	"30": 3,
	"31": 3,
	"33": 3,
	"34": 3,
	"35": 3,
	"96": 8,
	"02": 1,
	"04": 1,
	"06": 1,
	"08": 1,
	"19": 1,
	"20": 1,
	"22": 1,
	"24": 1,
	"25": 1,
	"10": 2,
	"11": 2,
	"12": 2,
	"14": 2,
	"15": 2,
	"17": 2,
	"70": 7,
	"72": 7,
	"74": 7,
	"75": 7,
	"79": 7,
	"77": 7,
	"36": 3,
	"37": 3,
	"38": 4,
	"40": 4,
	"42": 4,
	"44": 4,
	"45": 4,
	"46": 4,
	"48": 5,
	"49": 5,
	"51": 5,
	"52": 5,
	"54": 5,
	"56": 5,
	"58": 5,
	"60": 5,
	"62": 6,
	"64": 6,
	"66": 6,
	"67": 6,
	"68": 6,
	"80": 8,
	"82": 8,
	"83": 8,
	"84": 8,
	"86": 8,
	"87": 8,
	"89": 8,
	"91": 8,
	"92": 8,
	"93": 8,
	"94": 8,
	"95": 8,
}

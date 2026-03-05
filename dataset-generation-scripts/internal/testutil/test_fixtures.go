package testutil

import (
	vn_provinces_tmp_model "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
)

// TestFixtureLoader provides common test data fixtures
type TestFixtureLoader struct{}

// NewTestFixtureLoader creates a new test fixture loader
func NewTestFixtureLoader() *TestFixtureLoader {
	return &TestFixtureLoader{}
}

// LoadProvinceTestData returns sample province test data
func (l *TestFixtureLoader) LoadProvinceTestData() []vn_provinces_tmp_model.Province {
	return []vn_provinces_tmp_model.Province{
		{
			Code:                 "01",
			Name:                 "Hà Nội",
			NameEn:               "Ha Noi",
			FullName:             "Thành phố Hà Nội",
			FullNameEn:           "Ha Noi",
			CodeName:             "ha_noi",
			AdministrativeUnitId: 1, // Thanh pho
		},
		{
			Code:                 "02",
			Name:                 "Hải Phòng",
			NameEn:               "Hai Phong",
			FullName:             "Thành phố Hải Phòng",
			FullNameEn:           "Hai Phong",
			CodeName:             "hai_phong",
			AdministrativeUnitId: 1, // Thanh pho
		},
		{
			Code:                 "10",
			Name:                 "Khánh Hòa",
			NameEn:               "Khanh Hoa",
			FullName:             "Tỉnh Khánh Hòa",
			FullNameEn:           "Khanh Hoa",
			CodeName:             "khanh_hoa",
			AdministrativeUnitId: 2, // Tinh
		},
	}
}

// LoadWardTestData returns sample ward test data
func (l *TestFixtureLoader) LoadWardTestData() []vn_provinces_tmp_model.Ward {
	return []vn_provinces_tmp_model.Ward{
		{
			Code:                 "001",
			Name:                 "Bắc Sơn",
			NameEn:               "Bac Son",
			FullName:             "Phường Bắc Sơn",
			FullNameEn:           "Bac Son Ward",
			CodeName:             "bac_son",
			AdministrativeUnitId: 3, // Phuong
			ProvinceCode:         "01",
		},
		{
			Code:                 "002",
			Name:                 "Tân Xã",
			NameEn:               "Tan Xa",
			FullName:             "Xã Tân Xã",
			FullNameEn:           "Tan Xa Commune",
			CodeName:             "tan_xa",
			AdministrativeUnitId: 4, // Xa
			ProvinceCode:         "01",
		},
		{
			Code:                 "003",
			Name:                 "1",
			NameEn:               "1",
			FullName:             "Phường 1",
			FullNameEn:           "Ward 1",
			CodeName:             "1",
			AdministrativeUnitId: 3, // Phuong
			ProvinceCode:         "02",
		},
		{
			Code:                 "004",
			Name:                 "Phó Bảng",
			NameEn:               "Pho Bang",
			FullName:             "Xã Phó Bảng",
			FullNameEn:           "Pho Bang Commune",
			CodeName:             "pho_bang",
			AdministrativeUnitId: 4, // Xa
			ProvinceCode:         "06",
		},
	}
}

// LoadAdministrativeRegionTestData returns sample administrative region test data
func (l *TestFixtureLoader) LoadAdministrativeRegionTestData() []vn_provinces_tmp_model.AdministrativeRegion {
	return []vn_provinces_tmp_model.AdministrativeRegion{
		{
			Id:        1,
			Name:      "Đông Bắc Bộ",
			NameEn:    "Northeast",
			CodeName:  "dong_bac_bo",
			CodeNameEn: "northeast",
		},
		{
			Id:        2,
			Name:      "Tây Bắc Bộ",
			NameEn:    "Northwest",
			CodeName:  "tay_bac_bo",
			CodeNameEn: "northwest",
		},
	}
}

// LoadAdministrativeUnitTestData returns sample administrative unit test data
func (l *TestFixtureLoader) LoadAdministrativeUnitTestData() []vn_provinces_tmp_model.AdministrativeUnit {
	return []vn_provinces_tmp_model.AdministrativeUnit{
		{
			Id:          1,
			FullName:    "Thành phố",
			FullNameEn:  "City",
			ShortName:   "TP",
			ShortNameEn: "City",
			CodeName:    "thanh_pho",
			CodeNameEn:  "city",
		},
		{
			Id:          2,
			FullName:    "Tỉnh",
			FullNameEn:  "Province",
			ShortName:   "Tỉnh",
			ShortNameEn: "Province",
			CodeName:    "tinh",
			CodeNameEn:  "province",
		},
		{
			Id:          3,
			FullName:    "Phường",
			FullNameEn:  "Ward",
			ShortName:   "Phường",
			ShortNameEn: "Ward",
			CodeName:    "phuong",
			CodeNameEn:  "ward",
		},
		{
			Id:          4,
			FullName:    "Xã",
			FullNameEn:  "Commune",
			ShortName:   "Xã",
			ShortNameEn: "Commune",
			CodeName:    "xa",
			CodeNameEn:  "commune",
		},
		{
			Id:          5,
			FullName:    "Đặc khu",
			FullNameEn:  "Special Zone",
			ShortName:   "Đặc khu",
			ShortNameEn: "Special Zone",
			CodeName:    "dac_khu",
			CodeNameEn:  "special_zone",
		},
	}
}

// LoadGISTestData returns sample GIS test data
func (l *TestFixtureLoader) LoadGISTestData() ([]vn_provinces_tmp_model.Province, []vn_provinces_tmp_model.Ward) {
	provinces := []vn_provinces_tmp_model.Province{
		{
			Code:                 "01",
			Name:                 "Hà Nội",
			NameEn:               "Ha Noi",
			FullName:             "Thành phố Hà Nội",
			FullNameEn:           "Ha Noi",
			CodeName:             "ha_noi",
			AdministrativeUnitId: 1,
		},
	}

	wards := []vn_provinces_tmp_model.Ward{
		{
			Code:                 "001",
			Name:                 "Bắc Sơn",
			NameEn:               "Bac Son",
			FullName:             "Phường Bắc Sơn",
			FullNameEn:           "Bac Son Ward",
			CodeName:             "bac_son",
			AdministrativeUnitId: 3,
			ProvinceCode:         "01",
		},
	}

	return provinces, wards
}
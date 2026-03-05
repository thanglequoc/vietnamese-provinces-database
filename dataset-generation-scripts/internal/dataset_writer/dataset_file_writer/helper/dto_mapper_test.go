package helper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	vn_provinces_tmp_model "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
)

func TestConvertToJsonProvinceModel(t *testing.T) {
	provinces := []vn_provinces_tmp_model.Province{
		{
			Code:                 "01",
			Name:                 "Hà Nội",
			NameEn:               "Ha Noi",
			FullName:             "Thành phố Hà Nội",
			FullNameEn:           "Ha Noi City",
			CodeName:             "ha_noi",
			AdministrativeUnitId:  1,
			AdministrativeUnit: vn_provinces_tmp_model.AdministrativeUnit{
				Id:          1,
				FullName:    "Thành phố",
				ShortName:   "TP",
				FullNameEn:  "City",
				ShortNameEn: "City",
			},
		},
	}
	
	result := ConvertToJsonProvinceModel(provinces)
	assert.Len(t, result, 1)
	assert.Equal(t, "province", result[0].Type)
	assert.Equal(t, "01", result[0].Code)
	assert.Equal(t, "Hà Nội", result[0].Name)
	assert.Equal(t, "Ha Noi", result[0].NameEn)
	assert.Equal(t, "Thành phố Hà Nội", result[0].FullName)
	assert.Equal(t, "ha_noi", result[0].CodeName)
	assert.Equal(t, "TP", result[0].AdministrativeUnitShortName)
	assert.Equal(t, "Thành phố", result[0].AdministrativeUnitFullName)
}

func TestConvertToJsonProvinceModel_WithWards(t *testing.T) {
	ward := &vn_provinces_tmp_model.Ward{
		Code:       "001",
		Name:       "Bắc Sơn",
		NameEn:     "Bac Son",
		FullName:   "Phường Bắc Sơn",
		FullNameEn: "Bac Son Ward",
		CodeName:   "bac_son",
		ProvinceCode: "01",
		AdministrativeUnit: vn_provinces_tmp_model.AdministrativeUnit{
			Id:          3,
			ShortName:   "Phường",
			ShortNameEn: "Ward",
		},
	}
	
	provinces := []vn_provinces_tmp_model.Province{
		{
			Code:                 "01",
			Name:                 "Hà Nội",
			NameEn:               "Ha Noi",
			FullName:             "Thành phố Hà Nội",
			FullNameEn:           "Ha Noi City",
			CodeName:             "ha_noi",
			AdministrativeUnitId:  1,
			Wards:                []*vn_provinces_tmp_model.Ward{ward},
		},
	}
	
	result := ConvertToJsonProvinceModel(provinces)
	assert.Len(t, result, 1)
	assert.Len(t, result[0].Wards, 1)
	assert.Equal(t, "ward", result[0].Wards[0].Type)
	assert.Equal(t, "001", result[0].Wards[0].Code)
	assert.Equal(t, "Bắc Sơn", result[0].Wards[0].Name)
	assert.Equal(t, "Phường", result[0].Wards[0].AdministrativeUnitShortName)
}

func TestConvertToJsonProvinceSimplifiedModel(t *testing.T) {
	provinces := []vn_provinces_tmp_model.Province{
		{
			Code:       "01",
			Name:       "Hà Nội",
			NameEn:     "Ha Noi",
			FullName:   "Thành phố Hà Nội",
			FullNameEn: "Ha Noi City",
			CodeName:   "ha_noi",
		},
	}
	
	result := ConvertToJsonProvinceSimplifiedModel(provinces)
	assert.Len(t, result, 1)
	assert.Equal(t, "01", result[0].Code)
	assert.Equal(t, "Hà Nội", result[0].Name)
	assert.Equal(t, "Ha Noi", result[0].NameEn)
	assert.Equal(t, "Thành phố Hà Nội", result[0].FullName)
	assert.Equal(t, "ha_noi", result[0].CodeName)
}

func TestConvertToJsonProvinceVNSimplifiedModel(t *testing.T) {
	provinces := []vn_provinces_tmp_model.Province{
		{
			Code:     "01",
			FullName: "Thành phố Hà Nội",
		},
	}
	
	result := ConvertToJsonProvinceVNSimplifiedModel(provinces)
	assert.Len(t, result, 1)
	assert.Equal(t, "01", result[0].Code)
	assert.Equal(t, "Thành phố Hà Nội", result[0].FullName)
}

func TestConvertToMongoProvinceModel(t *testing.T) {
	provinces := []vn_provinces_tmp_model.Province{
		{
			Code:                 "01",
			Name:                 "Hà Nội",
			NameEn:               "Ha Noi",
			FullName:             "Thành phố Hà Nội",
			FullNameEn:           "Ha Noi City",
			CodeName:             "ha_noi",
			AdministrativeUnitId:  1,
		},
	}
	
	result := ConvertToMongoProvinceModel(provinces)
	assert.Len(t, result, 1)
	assert.Equal(t, "province", result[0].Type)
	assert.Equal(t, "01", result[0].Code)
	assert.Equal(t, "Hà Nội", result[0].Name)
	assert.Equal(t, "Ha Noi", result[0].NameEn)
	assert.Equal(t, "ha_noi", result[0].CodeName)
}

func TestConvertToJsonWardModel(t *testing.T) {
	wards := []vn_provinces_tmp_model.Ward{
		{
			Code:       "001",
			Name:       "Bắc Sơn",
			NameEn:     "Bac Son",
			FullName:   "Phường Bắc Sơn",
			FullNameEn: "Bac Son Ward",
			CodeName:   "bac_son",
			ProvinceCode: "01",
			AdministrativeUnitId: 3,
			AdministrativeUnit: vn_provinces_tmp_model.AdministrativeUnit{
				Id:          3,
				ShortName:   "Phường",
				FullName:    "Phường",
				ShortNameEn: "Ward",
				FullNameEn:  "Ward",
			},
		},
	}
	
	result := ConvertToJsonWardModel(wards)
	assert.Len(t, result, 1)
	assert.Equal(t, "ward", result[0].Type)
	assert.Equal(t, "001", result[0].Code)
	assert.Equal(t, "Bắc Sơn", result[0].Name)
	assert.Equal(t, "Bac Son", result[0].NameEn)
	assert.Equal(t, "Phường Bắc Sơn", result[0].FullName)
	assert.Equal(t, "bac_son", result[0].CodeName)
	assert.Equal(t, "01", result[0].ProvinceCode)
	assert.Equal(t, "Phường", result[0].AdministrativeUnitShortName)
}

func TestConvertToJsonWardSimplifiedModel(t *testing.T) {
	wards := []vn_provinces_tmp_model.Ward{
		{
			Code:         "001",
			Name:         "Bắc Sơn",
			NameEn:       "Bac Son",
			FullName:     "Phường Bắc Sơn",
			FullNameEn:   "Bac Son Ward",
			CodeName:     "bac_son",
			ProvinceCode: "01",
		},
	}
	
	result := ConvertToJsonWardSimplifiedModel(wards)
	assert.Len(t, result, 1)
	assert.Equal(t, "001", result[0].Code)
	assert.Equal(t, "Bắc Sơn", result[0].Name)
	assert.Equal(t, "Bac Son", result[0].NameEn)
	assert.Equal(t, "Phường Bắc Sơn", result[0].FullName)
	assert.Equal(t, "bac_son", result[0].CodeName)
	assert.Equal(t, "01", result[0].ProvinceCode)
}

func TestConvertToJsonWardVNSimplifiedModel(t *testing.T) {
	wards := []vn_provinces_tmp_model.Ward{
		{
			Code:         "001",
			FullName:     "Phường Bắc Sơn",
			ProvinceCode: "01",
		},
	}
	
	result := ConvertToJsonWardVNSimplifiedModel(wards)
	assert.Len(t, result, 1)
	assert.Equal(t, "001", result[0].Code)
	assert.Equal(t, "Phường Bắc Sơn", result[0].FullName)
	assert.Equal(t, "01", result[0].ProvinceCode)
}

func TestConvertToMongoWardModel(t *testing.T) {
	wards := []vn_provinces_tmp_model.Ward{
		{
			Code:                 "001",
			Name:                 "Bắc Sơn",
			NameEn:               "Bac Son",
			FullName:             "Phường Bắc Sơn",
			FullNameEn:           "Bac Son Ward",
			CodeName:             "bac_son",
			ProvinceCode:         "01",
			AdministrativeUnitId:  3,
		},
	}
	
	result := ConvertToMongoWardModel(wards)
	assert.Len(t, result, 1)
	assert.Equal(t, "ward", result[0].Type)
	assert.Equal(t, "001", result[0].Code)
	assert.Equal(t, "Bắc Sơn", result[0].Name)
	assert.Equal(t, "Bac Son", result[0].NameEn)
	assert.Equal(t, "Phường Bắc Sơn", result[0].FullName)
	assert.Equal(t, "bac_son", result[0].CodeName)
	assert.Equal(t, "01", result[0].ProvinceCode)
}

func TestConvertToJsonProvinceModel_EmptySlice(t *testing.T) {
	provinces := []vn_provinces_tmp_model.Province{}
	result := ConvertToJsonProvinceModel(provinces)
	assert.Len(t, result, 0)
}

func TestConvertToJsonWardModel_EmptySlice(t *testing.T) {
	wards := []vn_provinces_tmp_model.Ward{}
	result := ConvertToJsonWardModel(wards)
	assert.Len(t, result, 0)
}
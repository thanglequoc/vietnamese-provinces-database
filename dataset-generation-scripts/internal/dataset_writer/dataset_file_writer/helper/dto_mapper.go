package helper

import (
	vn_common "github.com/thanglequoc-vn-provinces/v2/internal/common"
	dataset_file_writer_dto "github.com/thanglequoc-vn-provinces/v2/internal/dataset_writer/dataset_file_writer/dto"
)

func ConvertToJsonProvinceModel(provinces []vn_common.Province) []dataset_file_writer_dto.JsonProvinceModel {
	var result []dataset_file_writer_dto.JsonProvinceModel
	for _, province := range provinces {
		p := dataset_file_writer_dto.JsonProvinceModel{
			Type:       "province",
			Code:       province.Code,
			Name:       province.Name,
			NameEn:     province.NameEn,
			FullName:   province.FullName,
			FullNameEn: province.FullNameEn,
			CodeName:   province.CodeName,

			AdministrativeUnitId:          province.AdministrativeUnitId,
			AdministrativeUnitShortName:   province.AdministrativeUnit.ShortName,
			AdministrativeUnitFullName:    province.AdministrativeUnit.FullName,
			AdministrativeUnitShortNameEn: province.AdministrativeUnit.ShortNameEn,
			AdministrativeUnitFullNameEn:  province.AdministrativeUnit.FullNameEn,
		}

		if len(province.Wards) != 0 {
			wards := make([]vn_common.Ward, len(province.Wards))
			for i, w := range province.Wards {
				wards[i] = *w
			}
			p.Wards = ConvertToJsonWardModel(wards)
		}
		result = append(result, p)
	}
	return result
}

func ConvertToJsonProvinceSimplifiedModel(provinces []vn_common.Province) []dataset_file_writer_dto.JsonProvinceSimplifiedModel {
	var result []dataset_file_writer_dto.JsonProvinceSimplifiedModel
	for _, province := range provinces {
		p := dataset_file_writer_dto.JsonProvinceSimplifiedModel{
			Code:       province.Code,
			Name:       province.Name,
			NameEn:     province.NameEn,
			FullName:   province.FullName,
			FullNameEn: province.FullNameEn,
			CodeName:   province.CodeName,
		}

		if len(province.Wards) != 0 {
			wards := make([]vn_common.Ward, len(province.Wards))
			for i, w := range province.Wards {
				wards[i] = *w
			}
			p.Wards = ConvertToJsonWardSimplifiedModel(wards)
		}
		result = append(result, p)
	}
	return result
}

func ConvertToJsonProvinceVNSimplifiedModel(provinces []vn_common.Province) []dataset_file_writer_dto.JsonProvinceVNSimplifiedModel {
	var result []dataset_file_writer_dto.JsonProvinceVNSimplifiedModel
	for _, province := range provinces {
		p := dataset_file_writer_dto.JsonProvinceVNSimplifiedModel{
			Code:     province.Code,
			FullName: province.FullName,
		}

		if len(province.Wards) != 0 {
			wards := make([]vn_common.Ward, len(province.Wards))
			for i, w := range province.Wards {
				wards[i] = *w
			}
			p.Wards = ConvertToJsonWardVNSimplifiedModel(wards)
		}
		result = append(result, p)
	}
	return result
}

func ConvertToMongoProvinceModel(provinces []vn_common.Province) []dataset_file_writer_dto.MongoProvinceModel {
	var result []dataset_file_writer_dto.MongoProvinceModel
	for _, province := range provinces {
		p := dataset_file_writer_dto.MongoProvinceModel{
			Type:                 "province",
			Code:                 province.Code,
			Name:                 province.Name,
			NameEn:               province.NameEn,
			FullName:             province.FullName,
			FullNameEn:           province.FullNameEn,
			CodeName:             province.CodeName,
			AdministrativeUnitId: province.AdministrativeUnitId,
		}

		if len(province.Wards) != 0 {
			wards := make([]vn_common.Ward, len(province.Wards))
			for i, w := range province.Wards {
				wards[i] = *w
			}
			p.Wards = ConvertToMongoWardModel(wards)
		}
		result = append(result, p)
	}

	return result
}

func ConvertToJsonWardModel(wards []vn_common.Ward) []dataset_file_writer_dto.JsonWardModel {
	var result []dataset_file_writer_dto.JsonWardModel

	for _, ward := range wards {
		w := dataset_file_writer_dto.JsonWardModel{
			Type:         "ward",
			Code:         ward.Code,
			Name:         ward.Name,
			NameEn:       ward.NameEn,
			FullName:     ward.FullName,
			FullNameEn:   ward.FullNameEn,
			CodeName:     ward.CodeName,
			ProvinceCode: ward.ProvinceCode,

			AdministrativeUnitId:          ward.AdministrativeUnitId,
			AdministrativeUnitShortName:   ward.AdministrativeUnit.ShortName,
			AdministrativeUnitFullName:    ward.AdministrativeUnit.FullName,
			AdministrativeUnitShortNameEn: ward.AdministrativeUnit.ShortNameEn,
			AdministrativeUnitFullNameEn:  ward.AdministrativeUnit.FullNameEn,
		}
		result = append(result, w)
	}

	return result
}

func ConvertToJsonWardSimplifiedModel(wards []vn_common.Ward) []dataset_file_writer_dto.JsonWardSimplifiedModel {
	var result []dataset_file_writer_dto.JsonWardSimplifiedModel

	for _, ward := range wards {
		w := dataset_file_writer_dto.JsonWardSimplifiedModel{
			Code:         ward.Code,
			Name:         ward.Name,
			NameEn:       ward.NameEn,
			FullName:     ward.FullName,
			FullNameEn:   ward.FullNameEn,
			CodeName:     ward.CodeName,
			ProvinceCode: ward.ProvinceCode,
		}
		result = append(result, w)
	}

	return result
}

func ConvertToJsonWardVNSimplifiedModel(wards []vn_common.Ward) []dataset_file_writer_dto.JsonWardVNSimplifiedModel {
	var result []dataset_file_writer_dto.JsonWardVNSimplifiedModel
	for _, ward := range wards {
		w := dataset_file_writer_dto.JsonWardVNSimplifiedModel{
			Code:         ward.Code,
			FullName:     ward.FullName,
			ProvinceCode: ward.ProvinceCode,
		}
		result = append(result, w)
	}
	return result
}

func ConvertToMongoWardModel(wards []vn_common.Ward) []dataset_file_writer_dto.MongoWardModel {
	var result []dataset_file_writer_dto.MongoWardModel

	for _, ward := range wards {
		w := dataset_file_writer_dto.MongoWardModel{
			Type:                 "ward",
			Code:                 ward.Code,
			Name:                 ward.Name,
			NameEn:               ward.NameEn,
			FullName:             ward.FullName,
			FullNameEn:           ward.FullNameEn,
			CodeName:             ward.CodeName,
			ProvinceCode:         ward.ProvinceCode,
			AdministrativeUnitId: ward.AdministrativeUnitId,
		}
		result = append(result, w)
	}

	return result
}

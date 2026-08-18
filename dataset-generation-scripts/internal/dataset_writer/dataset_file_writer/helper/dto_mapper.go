package helper

import (
	"strings"

	"github.com/thanglequoc-vn-provinces/v2/internal/common/viet"
	dataset_file_writer_dto "github.com/thanglequoc-vn-provinces/v2/internal/dataset_writer/dataset_file_writer/dto"
	"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
)

func ConvertToJsonProvinceModel(provinces []model.Province) []dataset_file_writer_dto.JsonProvinceModel {
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
			PostalCodePrefix: province.PostalCodePrefix,

			AdministrativeUnitId:          province.AdministrativeUnitId,
			AdministrativeUnitShortName:   province.AdministrativeUnit.ShortName,
			AdministrativeUnitFullName:    province.AdministrativeUnit.FullName,
			AdministrativeUnitShortNameEn: province.AdministrativeUnit.ShortNameEn,
			AdministrativeUnitFullNameEn:  province.AdministrativeUnit.FullNameEn,
		}

		if len(province.Wards) != 0 {
			wards := make([]model.Ward, len(province.Wards))
			for i, w := range province.Wards {
				wards[i] = *w
			}
			p.Wards = ConvertToJsonWardModel(wards)
		}
		result = append(result, p)
	}
	return result
}

func ConvertToJsonProvinceSimplifiedModel(provinces []model.Province) []dataset_file_writer_dto.JsonProvinceSimplifiedModel {
	var result []dataset_file_writer_dto.JsonProvinceSimplifiedModel
	for _, province := range provinces {
		p := dataset_file_writer_dto.JsonProvinceSimplifiedModel{
			Code:       province.Code,
			Name:       province.Name,
			NameEn:     province.NameEn,
			FullName:   province.FullName,
			FullNameEn: province.FullNameEn,
			CodeName:   province.CodeName,
			PostalCodePrefix: province.PostalCodePrefix,
		}

		if len(province.Wards) != 0 {
			wards := make([]model.Ward, len(province.Wards))
			for i, w := range province.Wards {
				wards[i] = *w
			}
			p.Wards = ConvertToJsonWardSimplifiedModel(wards)
		}
		result = append(result, p)
	}
	return result
}

func ConvertToJsonProvinceVNSimplifiedModel(provinces []model.Province) []dataset_file_writer_dto.JsonProvinceVNSimplifiedModel {
	var result []dataset_file_writer_dto.JsonProvinceVNSimplifiedModel
	for _, province := range provinces {
		p := dataset_file_writer_dto.JsonProvinceVNSimplifiedModel{
			Code:     province.Code,
			FullName: province.FullName,
			PostalCodePrefix: province.PostalCodePrefix,
		}

		if len(province.Wards) != 0 {
			wards := make([]model.Ward, len(province.Wards))
			for i, w := range province.Wards {
				wards[i] = *w
			}
			p.Wards = ConvertToJsonWardVNSimplifiedModel(wards)
		}
		result = append(result, p)
	}
	return result
}

func ConvertToMongoProvinceModel(provinces []model.Province) []dataset_file_writer_dto.MongoProvinceModel {
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
			PostalCodePrefix:     province.PostalCodePrefix,
		}

		if len(province.Wards) != 0 {
			wards := make([]model.Ward, len(province.Wards))
			for i, w := range province.Wards {
				wards[i] = *w
			}
			p.Wards = ConvertToMongoWardModel(wards)
		}
		result = append(result, p)
	}

	return result
}

func ConvertToJsonWardModel(wards []model.Ward) []dataset_file_writer_dto.JsonWardModel {
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
			PostalCode:   ward.PostalCode,

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

func ConvertToJsonWardSimplifiedModel(wards []model.Ward) []dataset_file_writer_dto.JsonWardSimplifiedModel {
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
			PostalCode:   ward.PostalCode,
		}
		result = append(result, w)
	}

	return result
}

func ConvertToJsonWardVNSimplifiedModel(wards []model.Ward) []dataset_file_writer_dto.JsonWardVNSimplifiedModel {
	var result []dataset_file_writer_dto.JsonWardVNSimplifiedModel
	for _, ward := range wards {
		w := dataset_file_writer_dto.JsonWardVNSimplifiedModel{
			Code:         ward.Code,
			FullName:     ward.FullName,
			ProvinceCode: ward.ProvinceCode,
			PostalCode:   ward.PostalCode,
		}
		result = append(result, w)
	}
	return result
}

func ConvertToMongoWardModel(wards []model.Ward) []dataset_file_writer_dto.MongoWardModel {
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
			PostalCode:           ward.PostalCode,
		}
		result = append(result, w)
	}

	return result
}

// GenerateSearchKeywords builds a deduplicated keyword array for Elasticsearch
// autocomplete/search. The array contains: code, tone-stripped lowercase name,
// lowercase English name, and codeName.
func GenerateSearchKeywords(code, name, nameEn, codeName string) []string {
	keywords := []string{
		code,
		strings.ToLower(viet.RemoveVietToneMark(name)),
		strings.ToLower(nameEn),
		codeName,
	}
	return deduplicate(keywords)
}

// deduplicate removes duplicate strings from a slice while preserving order.
func deduplicate(items []string) []string {
	seen := make(map[string]bool, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

// ConvertToElasticsearchProvinceModel converts Province domain models to
// Elasticsearch province documents with embedded wards and SearchKeywords.
func ConvertToElasticsearchProvinceModel(provinces []model.Province) []dataset_file_writer_dto.ElasticsearchProvinceDocument {
	var result []dataset_file_writer_dto.ElasticsearchProvinceDocument
	for _, province := range provinces {
		p := dataset_file_writer_dto.ElasticsearchProvinceDocument{
			Code:               province.Code,
			Name:               province.Name,
			NameEn:             province.NameEn,
			FullName:           province.FullName,
			FullNameEn:         province.FullNameEn,
			CodeName:           province.CodeName,
			PostalCodePrefix:   province.PostalCodePrefix,
			AdministrativeUnit: convertToElasticsearchAdministrativeUnit(province.AdministrativeUnit),
			SearchKeywords:     GenerateSearchKeywords(province.Code, province.Name, province.NameEn, province.CodeName),
		}

		if len(province.Wards) != 0 {
			wards := make([]model.Ward, len(province.Wards))
			for i, w := range province.Wards {
				wards[i] = *w
			}
			p.Wards = convertToElasticsearchWardDocuments(wards)
		}
		result = append(result, p)
	}
	return result
}

func convertToElasticsearchWardDocuments(wards []model.Ward) []dataset_file_writer_dto.ElasticsearchWardDocument {
	var result []dataset_file_writer_dto.ElasticsearchWardDocument
	for _, ward := range wards {
		w := dataset_file_writer_dto.ElasticsearchWardDocument{
			Code:               ward.Code,
			Name:               ward.Name,
			NameEn:             ward.NameEn,
			FullName:           ward.FullName,
			FullNameEn:         ward.FullNameEn,
			CodeName:           ward.CodeName,
			PostalCode:         ward.PostalCode,
			AdministrativeUnit: convertToElasticsearchAdministrativeUnit(ward.AdministrativeUnit),
			SearchKeywords:     GenerateSearchKeywords(ward.Code, ward.Name, ward.NameEn, ward.CodeName),
		}
		result = append(result, w)
	}
	return result
}

func convertToElasticsearchAdministrativeUnit(au model.AdministrativeUnit) dataset_file_writer_dto.ElasticsearchAdministrativeUnit {
	return dataset_file_writer_dto.ElasticsearchAdministrativeUnit{
		Id:          au.Id,
		FullName:    au.FullName,
		FullNameEn:  au.FullNameEn,
		ShortName:   au.ShortName,
		ShortNameEn: au.ShortNameEn,
		CodeName:    au.CodeName,
		CodeNameEn:  au.CodeNameEn,
	}
}

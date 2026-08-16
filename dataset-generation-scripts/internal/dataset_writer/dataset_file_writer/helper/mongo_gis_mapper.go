package helper

import (
	"encoding/json"
	"fmt"

	dataset_file_writer_dto "github.com/thanglequoc-vn-provinces/v2/internal/dataset_writer/dataset_file_writer/dto"
	sapnhapbandomodel "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
	"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
)

// ConvertToMongoGISProvinceDocuments converts SapNhapSiteGeoUnit provinces
// to MongoDB GIS province documents.
func ConvertToMongoGISProvinceDocuments(
	geoProvinces []*sapnhapbandomodel.SapNhapSiteGeoUnit,
) []dataset_file_writer_dto.MongoGISProvinceDocument {
	var docs []dataset_file_writer_dto.MongoGISProvinceDocument
	for _, geoProvince := range geoProvinces {
		province := geoProvince.VNProvince
		doc := dataset_file_writer_dto.MongoGISProvinceDocument{
			Code:               province.Code,
			Name:               province.Name,
			NameEn:             province.NameEn,
			FullName:           province.FullName,
			FullNameEn:         province.FullNameEn,
			CodeName:           province.CodeName,
			AdministrativeUnit: convertToMongoAdministrativeUnit(province.AdministrativeUnit),
			SearchKeywords:     GenerateSearchKeywords(province.Code, province.Name, province.NameEn, province.CodeName),
		}

		// Add province GIS
		provinceProps := &dataset_file_writer_dto.MongoGISProperties{
			Code:             province.Code,
			Name:             province.Name,
			NameEn:           province.NameEn,
			FullName:         province.FullName,
			FullNameEn:       province.FullNameEn,
			CodeName:         province.CodeName,
			PostalCodePrefix: province.PostalCodePrefix,
			GisServerId:      geoProvince.MaLK,
			AreaKm2:          geoProvince.DienTichKM2,
		}
		if gis, err := sapnhapGeoUnitToMongoGIS(*geoProvince, provinceProps); err == nil {
			doc.GIS = gis
		}

		docs = append(docs, doc)
	}
	return docs
}

// ConvertToMongoGISWardDocuments converts SapNhapSiteGeoUnit wards
// to MongoDB GIS ward documents.
func ConvertToMongoGISWardDocuments(
	geoWards []*sapnhapbandomodel.SapNhapSiteGeoUnit,
) []dataset_file_writer_dto.MongoGISWardDocument {
	var docs []dataset_file_writer_dto.MongoGISWardDocument
	for _, geoWard := range geoWards {
		ward := geoWard.VNWard
		doc := dataset_file_writer_dto.MongoGISWardDocument{
			Code:               ward.Code,
			Name:               ward.Name,
			NameEn:             ward.NameEn,
			FullName:           ward.FullName,
			FullNameEn:         ward.FullNameEn,
			CodeName:           ward.CodeName,
			ProvinceCode:       geoWard.VNDSProvinceCode,
			AdministrativeUnit: convertToMongoAdministrativeUnit(ward.AdministrativeUnit),
			SearchKeywords:     GenerateSearchKeywords(ward.Code, ward.Name, ward.NameEn, ward.CodeName),
		}

		// Add ward GIS
		wardProps := &dataset_file_writer_dto.MongoGISProperties{
			Code:        ward.Code,
			Name:        ward.Name,
			NameEn:      ward.NameEn,
			FullName:    ward.FullName,
			FullNameEn:  ward.FullNameEn,
			CodeName:    ward.CodeName,
			PostalCode:  ward.PostalCode,
			GisServerId: geoWard.MaLK,
			AreaKm2:     geoWard.DienTichKM2,
		}
		if gis, err := sapnhapGeoUnitToMongoGIS(*geoWard, wardProps); err == nil {
			doc.GIS = gis
		}

		docs = append(docs, doc)
	}
	return docs
}

// sapnhapGeoUnitToMongoGIS converts a SapNhapSiteGeoUnit's BBoxGeoJSON and
// GeomGeoJSON into a MongoGIS struct.
func sapnhapGeoUnitToMongoGIS(unit sapnhapbandomodel.SapNhapSiteGeoUnit, properties *dataset_file_writer_dto.MongoGISProperties) (*dataset_file_writer_dto.MongoGIS, error) {
	bbox, center, err := parseBBoxForMongo(unit.BBoxGeoJSON)
	if err != nil {
		return nil, err
	}
	return &dataset_file_writer_dto.MongoGIS{
		Properties:  properties,
		Center:      center,
		BoundingBox: bbox,
		Geometry:    unit.GeomGeoJSON,
	}, nil
}

// parseBBoxForMongo parses a BBoxGeoJSON array [xmin, ymin, xmax, ymax]
// into MongoBoundingBox and MongoGeoPoint (center as GeoJSON Point).
func parseBBoxForMongo(bboxGeoJSON json.RawMessage) (dataset_file_writer_dto.MongoBoundingBox, dataset_file_writer_dto.MongoGeoPoint, error) {
	var coords []float64
	if err := json.Unmarshal(bboxGeoJSON, &coords); err != nil {
		return dataset_file_writer_dto.MongoBoundingBox{}, dataset_file_writer_dto.MongoGeoPoint{}, fmt.Errorf("parse bbox geojson: %w", err)
	}
	if len(coords) != 4 {
		return dataset_file_writer_dto.MongoBoundingBox{}, dataset_file_writer_dto.MongoGeoPoint{}, fmt.Errorf("expected 4 bbox coordinates, got %d", len(coords))
	}
	xmin, ymin, xmax, ymax := coords[0], coords[1], coords[2], coords[3]
	bbox := dataset_file_writer_dto.MongoBoundingBox{
		MinLongitude: xmin,
		MinLatitude:  ymin,
		MaxLongitude: xmax,
		MaxLatitude:  ymax,
	}
	center := dataset_file_writer_dto.MongoGeoPoint{
		Type:        "Point",
		Coordinates: [2]float64{(xmin + xmax) / 2, (ymin + ymax) / 2}, // [lon, lat]
	}
	return bbox, center, nil
}

// convertToMongoAdministrativeUnit converts a model.AdministrativeUnit to
// the MongoDB DTO.
func convertToMongoAdministrativeUnit(au model.AdministrativeUnit) dataset_file_writer_dto.MongoAdministrativeUnit {
	return dataset_file_writer_dto.MongoAdministrativeUnit{
		Id:          au.Id,
		FullName:    au.FullName,
		FullNameEn:  au.FullNameEn,
		ShortName:   au.ShortName,
		ShortNameEn: au.ShortNameEn,
		CodeName:    au.CodeName,
		CodeNameEn:  au.CodeNameEn,
	}
}
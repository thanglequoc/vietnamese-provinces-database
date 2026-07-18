package dto

import "encoding/json"

// ElasticsearchProvinceDocument represents a province document in the
// provinces or provinces-gis Elasticsearch index.
type ElasticsearchProvinceDocument struct {
	Code               string                          `json:"Code"`
	Name               string                          `json:"Name"`
	NameEn             string                          `json:"NameEn"`
	FullName           string                          `json:"FullName"`
	FullNameEn         string                          `json:"FullNameEn"`
	CodeName           string                          `json:"CodeName"`
	AdministrativeUnit ElasticsearchAdministrativeUnit `json:"AdministrativeUnit"`
	SearchKeywords     []string                        `json:"SearchKeywords"`
	Wards              []ElasticsearchWardDocument     `json:"Wards"`
	GIS                *ElasticsearchGIS               `json:"GIS,omitempty"`
	Meta               *ElasticsearchMeta              `json:"Meta,omitempty"`
}

// ElasticsearchWardDocument represents an embedded ward inside a province document.
type ElasticsearchWardDocument struct {
	Code               string                          `json:"Code"`
	Name               string                          `json:"Name"`
	NameEn             string                          `json:"NameEn"`
	FullName           string                          `json:"FullName"`
	FullNameEn         string                          `json:"FullNameEn"`
	CodeName           string                          `json:"CodeName"`
	AdministrativeUnit ElasticsearchAdministrativeUnit `json:"AdministrativeUnit"`
	SearchKeywords     []string                        `json:"SearchKeywords"`
	GIS                *ElasticsearchGIS               `json:"GIS,omitempty"`
}

// ElasticsearchAdministrativeUnit is the embedded administrative unit object.
type ElasticsearchAdministrativeUnit struct {
	Id          int    `json:"Id"`
	FullName    string `json:"FullName"`
	FullNameEn  string `json:"FullNameEn"`
	ShortName   string `json:"ShortName"`
	ShortNameEn string `json:"ShortNameEn"`
	CodeName    string `json:"CodeName"`
	CodeNameEn  string `json:"CodeNameEn"`
}

// ElasticsearchGIS holds optional GIS data for the provinces-gis index.
type ElasticsearchGIS struct {
	Center      ElasticsearchGeoPoint   `json:"Center"`
	BoundingBox ElasticsearchBoundingBox `json:"BoundingBox"`
	Geometry    json.RawMessage         `json:"Geometry"`
}

// ElasticsearchGeoPoint is a lat/lon point for Elasticsearch geo_point mapping.
type ElasticsearchGeoPoint struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// ElasticsearchBoundingBox holds the bounding box coordinates.
type ElasticsearchBoundingBox struct {
	MinLongitude float64 `json:"MinLongitude"`
	MinLatitude  float64 `json:"MinLatitude"`
	MaxLongitude float64 `json:"MaxLongitude"`
	MaxLatitude  float64 `json:"MaxLatitude"`
}

// ElasticsearchMeta holds dataset version metadata.
type ElasticsearchMeta struct {
	DatasetVersion         string `json:"DatasetVersion"`
	AdministrativeRevision string `json:"AdministrativeRevision"`
	GeneratedAt            string `json:"GeneratedAt"`
}
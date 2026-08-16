package dto

import "encoding/json"

// MongoGISProvinceDocument represents a province document in the
// provinces-gis MongoDB collection.
type MongoGISProvinceDocument struct {
	Code               string                  `json:"Code"`
	Name               string                  `json:"Name"`
	NameEn             string                  `json:"NameEn"`
	FullName           string                  `json:"FullName"`
	FullNameEn         string                  `json:"FullNameEn"`
	CodeName           string                  `json:"CodeName"`
	AdministrativeUnit MongoAdministrativeUnit `json:"AdministrativeUnit"`
	SearchKeywords     []string                `json:"SearchKeywords"`
	GIS                *MongoGIS               `json:"GIS,omitempty"`
}

// MongoGISWardDocument represents a ward document in the
// wards-gis MongoDB collection.
type MongoGISWardDocument struct {
	Code               string                  `json:"Code"`
	Name               string                  `json:"Name"`
	NameEn             string                  `json:"NameEn"`
	FullName           string                  `json:"FullName"`
	FullNameEn         string                  `json:"FullNameEn"`
	CodeName           string                  `json:"CodeName"`
	ProvinceCode       string                  `json:"ProvinceCode"`
	AdministrativeUnit MongoAdministrativeUnit `json:"AdministrativeUnit"`
	SearchKeywords     []string                `json:"SearchKeywords"`
	GIS                *MongoGIS               `json:"GIS,omitempty"`
}

// MongoGIS holds GIS data for the provinces-gis and wards-gis collections.
type MongoGIS struct {
	Center      MongoGeoPoint       `json:"Center"`
	BoundingBox MongoBoundingBox    `json:"BoundingBox"`
	Geometry    json.RawMessage     `json:"Geometry"`
	Properties  *MongoGISProperties `json:"Properties,omitempty"`
}

// MongoGeoPoint is a MongoDB-native GeoJSON Point for 2dsphere indexing.
type MongoGeoPoint struct {
	Type        string     `json:"type"`
	Coordinates [2]float64 `json:"coordinates"` // [lon, lat]
}

// MongoBoundingBox holds the bounding box coordinates.
type MongoBoundingBox struct {
	MinLongitude float64 `json:"MinLongitude"`
	MinLatitude  float64 `json:"MinLatitude"`
	MaxLongitude float64 `json:"MaxLongitude"`
	MaxLatitude  float64 `json:"MaxLatitude"`
}

// MongoGISProperties holds administrative metadata inside the GIS object.
type MongoGISProperties struct {
	Code             string  `json:"Code"`
	Name             string  `json:"Name"`
	NameEn           string  `json:"NameEn"`
	FullName         string  `json:"FullName"`
	FullNameEn       string  `json:"FullNameEn"`
	CodeName         string  `json:"CodeName"`
	PostalCode       string  `json:"PostalCode"`
	PostalCodePrefix string  `json:"PostalCodePrefix"`
	GisServerId      string  `json:"GisServerId"`
	AreaKm2          float64 `json:"AreaKm2"`
}

// MongoAdministrativeUnit is the embedded administrative unit object.
type MongoAdministrativeUnit struct {
	Id          int    `json:"Id"`
	FullName    string `json:"FullName"`
	FullNameEn  string `json:"FullNameEn"`
	ShortName   string `json:"ShortName"`
	ShortNameEn string `json:"ShortNameEn"`
	CodeName    string `json:"CodeName"`
	CodeNameEn  string `json:"CodeNameEn"`
}

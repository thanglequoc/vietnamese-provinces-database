package dto

import "encoding/json"

type GeoJSONFeatureCollection struct {
	Type     string           `json:"type"`
	BBox     json.RawMessage  `json:"bbox"`
	Features []GeoJSONFeature `json:"features"`
}

type GeoJSONFeature struct {
	Type       string                   `json:"type"`
	ID         string                   `json:"id"`
	BBox       json.RawMessage          `json:"bbox"`
	Geometry   json.RawMessage          `json:"geometry"`
	Properties GeoJSONFeatureProperties `json:"properties"`
}

type GeoJSONFeatureProperties struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	NameEn      string  `json:"nameEn"`
	FullName    string  `json:"fullName"`
	FullNameEn  string  `json:"fullNameEn"`
	CodeName    string  `json:"codeName"`
	GISServerID string  `json:"gisServerId"`
	AreaKm2     float64 `json:"areaKm2"`
}

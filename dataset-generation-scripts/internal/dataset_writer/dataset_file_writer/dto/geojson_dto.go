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
	UnitName     string  `json:"unit_name"`
	UnitCode     string  `json:"unit_code"`
	UnitCodeName string  `json:"unit_code_name"`
	GISServerID  string  `json:"gis_server_id"`
	AreaKm2      float64 `json:"area_km2"`
}

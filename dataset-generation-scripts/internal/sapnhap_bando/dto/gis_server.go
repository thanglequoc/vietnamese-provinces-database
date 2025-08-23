package dto

import "encoding/json"

// GISLocationResponse represents the detailed response from the GIS server
type GISLocationResponse struct {
	Type     string               `json:"type"`
	Bbox     []float64            `json:"bbox"`
	Features []GISLocationFeature `json:"features"`
}

// GISLocationFeature represents a feature in the detailed GIS response
type GISLocationFeature struct {
	Bbox     []float64   `json:"bbox"`
	Geometry GISGeometry `json:"geometry"`
	Type     string      `json:"type"`
}

// GISGeometry represents the geometry data with coordinates
type GISGeometry struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"` // Use RawMessage for flexible coordinate parsing
}

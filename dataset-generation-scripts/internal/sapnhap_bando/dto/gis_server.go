package dto

import (
	gisModel "github.com/thanglequoc-vn-provinces/v2/internal/gis/model"
)

// GISLocationResponse represents the detailed response from the GIS server
type GISLocationResponse struct {
	Type     string               `json:"type"`
	BBox     gisModel.BBox        `json:"bbox"`
	Features []GISLocationFeature `json:"features"`
}

// GISLocationFeature represents a feature in the detailed GIS response
type GISLocationFeature struct {
	BBox       gisModel.BBox `json:"bbox"`
	Geometry   GISGeometry   `json:"geometry"`
	ID         string        `json:"id"`
	Properties interface{}   `json:"properties"`
	Type       string        `json:"type"`
}

// GISGeometry represents the geometry data with coordinates
type GISGeometry struct {
	Type        string                               `json:"type"`
	Coordinates [][]gisModel.GISLinearRingCoordinate `json:"coordinates"`
}

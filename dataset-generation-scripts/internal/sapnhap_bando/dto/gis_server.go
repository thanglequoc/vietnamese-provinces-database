package dto

import (
	"fmt"
	"strings"

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

func (g GISGeometry) ToWKTCoordinate() string {
	if (g.Type != "MultiPolygon") {
		fmt.Println("Detected Geometry type anomaly")
		panic(fmt.Sprintf("Unsupported Geometry type: %s", g.Type))
	}
	var sb strings.Builder
	sb.WriteString(strings.ToUpper(g.Type))
	sb.WriteString("(")

	var polygons []string

	for _, polygon := range g.Coordinates {
		var polygonBuilder strings.Builder
		polygonBuilder.WriteString("(")
		for i, ring := range polygon {
			polygonBuilder.WriteString(ring.ToCoordinateRingString())
			if (i < len(polygon) - 1) {
				polygonBuilder.WriteString(",")
			}
		}
		
		polygonBuilder.WriteString(")")
		polygons = append(polygons, polygonBuilder.String())
	}

	sb.WriteString(strings.Join(polygons, ","))	
	sb.WriteString(")")
	return sb.String()
}
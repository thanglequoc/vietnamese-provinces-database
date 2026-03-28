package dto

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// GeoJSONFeatureCollection represents a standard GeoJSON FeatureCollection
type GeoJSONFeatureCollection struct {
	Type     string           `json:"type"`
	BBox     [4]float64       `json:"bbox"` // [minLng, minLat, maxLng, maxLat]
	Features []GeoJSONFeature `json:"features"`
}

// GeoJSONFeature represents a standard GeoJSON Feature
type GeoJSONFeature struct {
	BBox     [4]float64      `json:"bbox"` // [minLng, minLat, maxLng, maxLat]
	Geometry GeoJSONGeometry `json:"geometry"`
	ID       string          `json:"id"` // GIS server ID (e.g., "tinh34.7", "xa3321.3285")
}

// GeoJSONGeometry represents a standard GeoJSON Geometry
type GeoJSONGeometry struct {
	Type        string          `json:"type"` // "MultiPolygon"
	Coordinates [][][][2]float64 `json:"coordinates"` // MultiPolygon format
}

// ToWKBboxPolygon converts GeoJSON bbox to WKT POLYGON format
func (g *GeoJSONFeature) ToWKBboxPolygon() string {
	// GeoJSON bbox: [minLng, minLat, maxLng, maxLat]
	minLng, minLat, maxLng, maxLat := g.BBox[0], g.BBox[1], g.BBox[2], g.BBox[3]

	return fmt.Sprintf(
		"POLYGON((%f %f, %f %f, %f %f, %f %f, %f %f))",
		minLng, minLat, // bottom-left
		minLng, maxLat, // top-left
		maxLng, maxLat, // top-right
		maxLng, minLat, // bottom-right
		minLng, minLat, // close ring
	)
}

// ToWKTMultiPolygon converts GeoJSON MultiPolygon to WKT MULTIPOLYGON format
func (g *GeoJSONGeometry) ToWKTMultiPolygon() string {
	if g.Type != "MultiPolygon" {
		panic(fmt.Sprintf("Unsupported geometry type: %s", g.Type))
	}

	var sb strings.Builder
	sb.WriteString("MULTIPOLYGON(")

	var polygons []string
	for _, polygon := range g.Coordinates {
		var polygonBuilder strings.Builder
		polygonBuilder.WriteString("(")

		var rings []string
		for _, ring := range polygon {
			var ringBuilder strings.Builder
			ringBuilder.WriteString("(")

			for i, coord := range ring {
				ringBuilder.WriteString(fmt.Sprintf("%f %f", coord[0], coord[1]))
				if i < len(ring)-1 {
					ringBuilder.WriteString(", ")
				}
			}
			// Close the ring by repeating first point
			if len(ring) > 0 {
				ringBuilder.WriteString(fmt.Sprintf(", %f %f", ring[0][0], ring[0][1]))
			}

			ringBuilder.WriteString(")")
			rings = append(rings, ringBuilder.String())
		}

		polygonBuilder.WriteString(strings.Join(rings, ", "))
		polygonBuilder.WriteString(")")
		polygons = append(polygons, polygonBuilder.String())
	}

	sb.WriteString(strings.Join(polygons, ", "))
	sb.WriteString(")")

	return sb.String()
}

// LoadGeoJSONFile loads a GeoJSON file from disk
func LoadGeoJSONFile(filePath string) (*GeoJSONFeatureCollection, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var collection GeoJSONFeatureCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, err
	}

	return &collection, nil
}

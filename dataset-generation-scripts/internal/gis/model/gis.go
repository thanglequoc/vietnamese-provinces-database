package model

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type LngLat struct {
	Longitude float64
	Latitude  float64
}

// GISLinearRingCoordinate represents a closed ring of coordinates.
// In GIS, a LinearRing is used to define the boundary of a Polygon.
// The first and last points should be the same to form a closed loop.
type GISLinearRingCoordinate struct {
	GISPoints []LngLat
}

// BBox represents an axis-aligned bounding box of a geometry.
// It is defined by its four corner coordinates in longitude/latitude (WGS84).
//
// In GIS, a bounding box is the smallest rectangle that fully contains
// a geometry. While typically expressed by two corners (minX, minY, maxX, maxY),
// this struct stores all four corners explicitly for convenience.
type BBox struct {
	BottomLeft  LngLat
	TopLeft     LngLat
	TopRight    LngLat
	BottomRight LngLat
}

func (g *GISLinearRingCoordinate) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		return nil
	}
	var points []LngLat

	var trimmedData = strings.ReplaceAll(strings.ReplaceAll(string(data), "\n", ""), " ", "")
	if err := json.Unmarshal([]byte(trimmedData), &points); err != nil {
		log.Default().Println(err)
		log.Default().Println("Unable to unmarshal data: " + trimmedData)
		return err
	}

	g.GISPoints = points
	return nil
}

func (l *LngLat) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		return nil
	}

	var trimmedData = strings.ReplaceAll(strings.ReplaceAll(string(data), "\n", ""), " ", "")
	var latlng [2]float64
	if err := json.Unmarshal([]byte(trimmedData), &latlng); err != nil {
		log.Default().Println("Unable to unmarshal values: " + trimmedData)
		return err
	}

	l.Longitude = latlng[0]
	l.Latitude = latlng[1]
	return nil
}

// ToWKTPoint returns "lon lat" string for PostGIS WKT usage.
func (p LngLat) ToWKTCoordinatePoint() string {
	return fmt.Sprintf("%f %f", p.Longitude, p.Latitude)
}

func (b *BBox) UnmarshalJSON(data []byte) error {
	var bboxArray [4]float64
	if err := json.Unmarshal(data, &bboxArray); err != nil {
		return err
	}

	// @thangle: The bbox array format from the GIS Server response is:
	// [minLng, minLat, maxLng, maxLat]
	// - (minLng, minLat) represents the bottom-left corner
	// - (maxLng, maxLat) represents the top-right corner
	// The other two corners (top-left, bottom-right) are derived from these values.
	minLng := bboxArray[0]
	minLat := bboxArray[1]
	maxLng := bboxArray[2]
	maxLat := bboxArray[3]

	b.BottomLeft = LngLat{Longitude: minLng, Latitude: minLat}
	b.TopLeft = LngLat{Longitude: minLng, Latitude: maxLat}
	b.TopRight = LngLat{Longitude: maxLng, Latitude: maxLat}
	b.BottomRight = LngLat{Longitude: maxLng, Latitude: minLat}

	return nil
}

// ToWKTPolygon returns a WKT POLYGON string for the bbox
func (b BBox) ToWKTPolygon() string {
	return fmt.Sprintf(
		"POLYGON((%s, %s, %s, %s, %s))",
		b.BottomLeft.ToWKTCoordinatePoint(),
		b.TopLeft.ToWKTCoordinatePoint(),
		b.TopRight.ToWKTCoordinatePoint(),
		b.BottomRight.ToWKTCoordinatePoint(),
		b.BottomLeft.ToWKTCoordinatePoint(), // repeat to close the ring
	)
}

// Helper methods for easier access to coordinates
func (b *BBox) MinLongitude() float64 {
	return b.BottomLeft.Longitude
}

func (b *BBox) MinLatitude() float64 {
	return b.BottomLeft.Latitude
}

func (b *BBox) MaxLongitude() float64 {
	return b.TopRight.Longitude
}

func (b *BBox) MaxLatitude() float64 {
	return b.TopRight.Latitude
}

// Center returns the center point of the bounding box
func (b *BBox) Center() LngLat {
	return LngLat{
		Longitude: (b.MinLongitude() + b.MaxLongitude()) / 2,
		Latitude:  (b.MinLatitude() + b.MaxLatitude()) / 2,
	}
}

// Contains checks if a point is within the bounding box
func (b *BBox) Contains(point LngLat) bool {
	return point.Longitude >= b.MinLongitude() &&
		point.Longitude <= b.MaxLongitude() &&
		point.Latitude >= b.MinLatitude() &&
		point.Latitude <= b.MaxLatitude()
}

// ToArray returns the bbox in [minLng, minLat, maxLng, maxLat] format
func (b *BBox) ToArray() [4]float64 {
	return [4]float64{
		b.MinLongitude(),
		b.MinLatitude(),
		b.MaxLongitude(),
		b.MaxLatitude(),
	}
}

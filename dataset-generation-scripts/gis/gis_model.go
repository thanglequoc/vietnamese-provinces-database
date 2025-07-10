package gis

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type LngLat struct {
	Longitude float64
	Latitude float64
}

type GisLinearRingCoordinate struct {
	GisPoints []LngLat
}

type ProvinceGIS struct {
	LevelId string `json:"level1_id"`
	Name string `json:"name"`
	Districts []DistrictGIS `json:"level2s"`
	Coordinates [][] GisLinearRingCoordinate `json:"coordinates"`
	BBox BBox `json:"bbox"`
	Type string `json:"type"`
}

type DistrictGIS struct {
	LevelId string `json:"level2_id"`
	Name string `json:"name"`
	Coordinates [][] GisLinearRingCoordinate `json:"coordinates"`
	BBox BBox `json:"bbox"`
	Type string `json:"type"`
}

type BBox struct {
	BottomLeftLng float64
	BottomLeftLat float64

	TopLeftLng float64
	TopLeftLat float64

	TopRightLng float64
	TopRightLat float64

	BottomRightLng float64
	BottomRightLat float64
}

func (g *GisLinearRingCoordinate) UnmarshalJSON(data []byte) error {
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

	g.GisPoints = points
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

func (b *BBox) UnmarshalJSON(data []byte) error {
	var bboxArray [4]float64
	if err := json.Unmarshal(data, &bboxArray); err != nil {
			return err
	}

	// @thangle: The bbox array format from daohoangson gis resource is 
	// 2 first digits: BottomLeft coordinate LatLng
	// 2 last digits: TopRight coordinate LatLng
	b.BottomLeftLng = bboxArray[0]
	b.BottomLeftLat = bboxArray[1]

	b.TopLeftLng = bboxArray[2]
	b.TopLeftLat = bboxArray[1]
	
	b.TopRightLng = bboxArray[2]
	b.TopRightLat = bboxArray[3]

	b.BottomRightLng = bboxArray[0]
	b.BottomRightLat = bboxArray[3]
	return nil
}

func ToCoordinateStr(lng float64, lat float64) string {
	return fmt.Sprintf("%.15f %.15f", lng, lat) 
}

func ToCoordinateStrLatLng(lng float64, lat float64) string {
	return fmt.Sprintf("%.15f %.15f", lat, lng) 
}

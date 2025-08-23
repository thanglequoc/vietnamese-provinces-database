package dto

import (
	"encoding/json"
	"fmt"
)

type BanDoGISProvince struct {
	STT               string `json:"stt"`
	Ten               string `json:"ten"`
	TruocSN           string `json:"truocsn"`
	GISServerResponse string `json:"gisServerResponse"`
}

type BanDoGISWard struct {
	STT               string `json:"stt"`
	Ten               string `json:"ten"`
	TruocSN           string `json:"truocsn"`
	ProvinceName      string `json:"provinceName"`
	GISServerResponse string `json:"gisServerResponse"`
}

type BandoGISFeature struct {
	Geometry   interface{}            `json:"geometry"`
	ID         string                 `json:"id"`
	Properties map[string]interface{} `json:"properties"`
	Type       string                 `json:"type"`
}

type BandoGISServerResponse struct {
	Features []BandoGISFeature `json:"features"`
	Type     string            `json:"type"`
}

func (p *BanDoGISProvince) GetGISResponse() (*BandoGISServerResponse, error) {
	if p.GISServerResponse == "" {
		return nil, fmt.Errorf("no GIS server response data")
	}

	var gisResponse BandoGISServerResponse
	err := json.Unmarshal([]byte(p.GISServerResponse), &gisResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal GIS server response: %w", err)
	}

	return &gisResponse, nil
}

func (w *BanDoGISWard) GetGISResponse() (*BandoGISServerResponse, error) {
	if w.GISServerResponse == "" {
		return nil, fmt.Errorf("no GIS server response data")
	}

	var gisResponse BandoGISServerResponse
	err := json.Unmarshal([]byte(w.GISServerResponse), &gisResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal GIS server response: %w", err)
	}

	return &gisResponse, nil
}

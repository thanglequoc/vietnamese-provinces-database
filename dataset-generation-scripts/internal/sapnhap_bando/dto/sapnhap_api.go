package dto

import (
	"encoding/json"
	"fmt"
)

// DTO for the ptinh API response
type SapNhapProvinceData struct {
	ID          int     `json:"id"`
	MaHC        int     `json:"mahc"`
	TenTinh     string  `json:"tentinh"`
	DienTichKm2 string  `json:"dientichkm2"`
	DanSoNguoi  string  `json:"dansonguoi"`
	TrungTamHC  string  `json:"trungtamhc"`
	KinhDo      float64 `json:"kinhdo"`
	ViDo        float64 `json:"vido"`
	TruocSN     string  `json:"truocsapnhap"`
	Con         string  `json:"con"`
}

type SapNhapWardData struct {
	ID           int     `json:"id"`
	Matinh       int     `json:"matinh"`
	Ma           string  `json:"ma"`
	TenTinh      string  `json:"tentinh"`
	Loai         string  `json:"loai"`
	TenHC        string  `json:"tenhc"`
	Cay          string  `json:"cay"`
	DienTichKm2  float64 `json:"dientichkm2"`
	DanSoNguoi   string  `json:"dansonguoi"`
	TrungTamHC   string  `json:"trungtamhc"`
	KinhDo       float64 `json:"kinhdo"`
	ViDo         float64 `json:"vido"`
	TruocSapNhap string  `json:"truocsapnhap"`
	MaXa         int     `json:"maxa"`
	Khoa         string  `json:"khoa"`
}

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

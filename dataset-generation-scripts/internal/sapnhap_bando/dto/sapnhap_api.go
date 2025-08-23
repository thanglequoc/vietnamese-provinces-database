package dto

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

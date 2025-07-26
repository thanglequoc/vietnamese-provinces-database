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

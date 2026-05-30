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

type SapNhapGeoObjectMetadata struct {
	ID           int     `json:"id"`
	DienTichKM2  string  `json:"dientichkm2"`
	DanSoNguoi   string  `json:"dansonguoi"`
	TrungTamHC   string  `json:"trungtamhc"`
	TruocSapNhap string  `json:"truocsapnhap"`
	Con          string  `json:"con"`
	Ma           string  `json:"ma"`
	Ten          string  `json:"ten"`
	MaGoc        string  `json:"magoc"`
	Malk         string  `json:"malk"`
	Stat         int     `json:"stat"`
	DiaChi       string  `json:"diachi"`
	DienThoai    string  `json:"dthoai"`
	CanCu        string  `json:"cancu"`
	TenTinh      *string `json:"tentinh"`
	Link         string  `json:"link"`
}

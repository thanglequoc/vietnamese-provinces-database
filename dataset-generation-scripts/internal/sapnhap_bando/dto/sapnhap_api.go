package dto

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

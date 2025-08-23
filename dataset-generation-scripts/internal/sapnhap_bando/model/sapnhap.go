package model

import (
	"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
	"github.com/uptrace/bun"
)

type SapNhapSiteProvince struct {
	bun.BaseModel `bun:"table:sapnhap_provinces,alias:sp"`

	ID             int            `bun:"id,pk"`
	MaHC           int            `bun:"mahc"`
	TenTinh        string         `bun:"tentinh,notnull"`
	DienTichKm2    string         `bun:"dientichkm2"`
	DanSoNguoi     string         `bun:"dansonguoi"`
	TrungTamHC     string         `bun:"trungtamhc"`
	KinhDo         float64        `bun:"kinhdo"`
	ViDo           float64        `bun:"vido"`
	TruocSN        string         `bun:"truocsapnhap"`
	Con            string         `bun:"con"`
	VNProvinceCode string         `bun:"vn_province_code,notnull"`
	Province       model.Province `bun:"rel:belongs-to,join:vn_province_code=code"`
}

type SapNhapSiteWard struct {
	bun.BaseModel `bun:"table:sapnhap_wards,alias:sw"`

	ID                  int                 `json:"id" bun:"id,pk"`
	MaTinh              int                 `json:"matinh" bun:"matinh"`
	Ma                  string              `json:"ma" bun:"ma"`
	TenTinh             string              `json:"tentinh" bun:"tentinh"`
	Loai                string              `json:"loai" bun:"loai"`
	TenHC               string              `json:"tenhc" bun:"tenhc"`
	Cay                 string              `json:"cay" bun:"cay"`
	DienTichKm2         float64             `json:"dientichkm2" bun:"dientichkm2"`
	DanSoNguoi          string              `json:"dansonguoi" bun:"dansonguoi"`
	TrungTamHC          string              `json:"trungtamhc" bun:"trungtamhc"`
	KinhDo              float64             `json:"kinhdo" bun:"kinhdo"`
	ViDo                float64             `json:"vido" bun:"vido"`
	TruocSapNhap        string              `json:"truocsapnhap" bun:"truocsapnhap"`
	MaXa                int                 `json:"maxa" bun:"maxa"`
	Khoa                string              `json:"khoa" bun:"khoa"`
	SapNhapSiteProvince SapNhapSiteProvince `bun:"rel:belongs-to,join:matinh=mahc"`
	VNWardCode          string              `bun:"vn_ward_code,notnull"`
	Ward                model.Ward          `bun:"rel:belongs-to,join:vn_ward_code=code"`
}

type SapNhapProvinceGis struct {
	bun.BaseModel `bun:"table:sapnhap_provinces_gis,alias:spg"`

	Stt                   int                 `json:"stt" bun:"stt"`
	Ten                   string              `json:"ten" bun:"ten"`
	TruocSapNhap          string              `json:"truocSapNhap" bun:"truocsn"`
	GisServerID           string              `json:"gisServerID" bun:"gis_server_id"`
	SapNhapProvinceMaTinh int                 `json:"sapNhapProvinceMaTinh" bun:"sapnhap_province_matinh"`
	SapNhapSiteProvince   SapNhapSiteProvince `bun:"rel:belongs-to,join:sapnhap_province_matinh=mahc"`
}

type SapNhapWardGis struct {
	bun.BaseModel `bun:"table:sapnhap_wards_gis,alias:swg"`

	Stt             int             `json:"stt" bun:"stt"`
	Ten             string          `json:"ten" bun:"ten"`
	TruocSapNhap    string          `json:"truocSapNhap" bun:"truocsn"`
	GisServerID     string          `json:"gisServerID" bun:"gis_server_id"`
	SapNhapWardMaXa int             `json:"sapNhapWardMaXa" bun:"sapnhap_ward_maxa"`
	SapNhapSiteWard SapNhapSiteWard `bun:"rel:belongs-to,join:sapnhap_ward_maxa=maxa"`
}

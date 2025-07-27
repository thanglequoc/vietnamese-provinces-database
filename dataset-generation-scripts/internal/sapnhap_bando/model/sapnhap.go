package model

import (
	"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
	"github.com/uptrace/bun"
)

type SapNhapSiteProvince struct {
	bun.BaseModel `bun:"table:sapnhap_provinces,alias:sp"`

	ID             int             `bun:"id,pk"`
	MaHC           int             `bun:"mahc"`
	TenTinh        string          `bun:"tentinh,notnull"`
	DienTichKm2    string          `bun:"dientichkm2"`
	DanSoNguoi     string          `bun:"dansonguoi"`
	TrungTamHC     string          `bun:"trungtamhc"`
	KinhDo         float64         `bun:"kinhdo"`
	ViDo           float64         `bun:"vido"`
	TruocSN        string          `bun:"truocsapnhap"`
	Con            string          `bun:"con"`
	VNProvinceCode string          `bun:"vn_province_code,notnull"`
	Province       model.Province `bun:"rel:belongs-to,join:vn_province_code=code"`
}

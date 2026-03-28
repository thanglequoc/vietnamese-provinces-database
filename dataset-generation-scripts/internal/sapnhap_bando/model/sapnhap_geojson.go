package model

import (
	"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
	"github.com/uptrace/bun"
)

type SapNhapSiteGeoUnit struct {
	bun.BaseModel `bun:"table:sapnhap_geojson_objects,alias:sp"`

	Ma               string `bun:"ma,pk,notnull"`
	Ten              string `bun:"ten,notnull"`
	MaGoc            string `bun:"magoc"`
	MaLK             string `bun:"malk"`
	TruocSapNhap     string `bun:"truocsapnhap"`
	VNDSProvinceCode string `bun:"vn_ds_province_code"`
	VNDSWardCode     string `bun:"vn_ds_ward_code"`
	BBoxWKT          string `bun:"bbox_wkt"`
	GeomWKT          string `bun:"geom_wkt"`

	BBoxWKTLatLng string `bun:"bbox_wkt_lat_lng,scanonly"`
	GeomWKTLatLng string `bun:"geom_wkt_lat_lng,scanonly"`

	// Relationships
	Parent     *SapNhapSiteGeoUnit `bun:"rel:belongs-to,join:magoc=ma"`
	VNProvince model.Province      `bun:"rel:belongs-to,join:vn_ds_province_code=code"`
	VNWard     model.Ward          `bun:"rel:belongs-to,join:vn_ds_ward_code=code"`
}

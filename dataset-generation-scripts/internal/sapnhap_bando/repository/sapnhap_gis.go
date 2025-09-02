package repository

import (
	"context"

	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
	"github.com/uptrace/bun"
)

type SapNhapGISRepository struct {
	db *bun.DB
}

func NewSapNhapGISRepository(db *bun.DB) *SapNhapGISRepository {
	return &SapNhapGISRepository{
		db: db,
	}
}

func (r *SapNhapGISRepository) InsertSapNhapProvinceGIS(provinceGIS *model.SapNhapProvinceGIS) error {
	_, err := r.db.NewInsert().
		Model(provinceGIS).
		Exec(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (r *SapNhapGISRepository) InsertSapNhapWardGIS(wardGIS *model.SapNhapWardGIS) error {
	_, err := r.db.NewInsert().
		Model(wardGIS).
		Exec(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (r *SapNhapGISRepository) GetAllSapNhapProvinceGIS() ([]model.SapNhapProvinceGIS, error) {
	var provinceGISList []model.SapNhapProvinceGIS
	err := r.db.NewSelect().
		Model(&provinceGISList).Relation("SapNhapSiteProvince").
		ColumnExpr("stt, ten, truocsn, gis_server_id, sapnhap_province_matinh, bbox_wkt, geom_wkt").
		ColumnExpr("ST_AsText(ST_FlipCoordinates(bbox)) AS bbox_wkt_lat_lng").
		ColumnExpr("ST_AsText(ST_FlipCoordinates(geom)) AS geom_wkt_lat_lng").
		Scan(context.Background())
	if err != nil {
		return nil, err
	}
	return provinceGISList, nil
}

func (r *SapNhapGISRepository) GetAllSapNhapWardGIS() ([]model.SapNhapWardGIS, error) {
	var wardGISList []model.SapNhapWardGIS
	err := r.db.NewSelect().
		Model(&wardGISList).Relation("SapNhapSiteWard").
		ColumnExpr("stt, ten, truocsn, gis_server_id, sapnhap_ward_maxa, bbox_wkt, geom_wkt").
		ColumnExpr("ST_AsText(ST_FlipCoordinates(bbox)) AS bbox_wkt_lat_lng").
		ColumnExpr("ST_AsText(ST_FlipCoordinates(geom)) AS geom_wkt_lat_lng").
		Scan(context.Background())
	if err != nil {
		return nil, err
	}
	return wardGISList, nil
}

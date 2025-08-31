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
		Value("bbox", "ST_GeomFromText(?, 4326)", provinceGIS.BBox). // This converts string → geometry
		Value("gis_geom", "ST_GeomFromText(?, 4326)", provinceGIS.GISGeom).
		Exec(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (r *SapNhapGISRepository) InsertSapNhapWardGIS(wardGIS *model.SapNhapWardGIS) error {
	_, err := r.db.NewInsert().
		Model(wardGIS).
		Value("bbox", "ST_GeomFromText(?, 4326)", wardGIS.BBox).
		Value("gis_geom", "ST_GeomFromText(?, 4326)", wardGIS.GISGeom).
		Exec(context.Background())
	if err != nil {
		return err
	}
	return nil
}

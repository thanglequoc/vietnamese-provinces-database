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

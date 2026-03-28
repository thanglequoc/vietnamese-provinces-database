package repository

import (
	"context"
	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
	"github.com/uptrace/bun"
)

type SapNhapRepository struct {
	db *bun.DB
}

func NewSapNhapRepository(db *bun.DB) *SapNhapRepository {
	return &SapNhapRepository{
		db: db,
	}
}

/* Get all sapnhap province data */
func (r *SapNhapRepository) GetAllSapNhapSiteProvinces() ([]model.SapNhapSiteProvince, error) {
	var provinces []model.SapNhapSiteProvince
	err := r.db.NewSelect().
		Model(&provinces).Relation("Province").Relation("SapNhapGIS").
		Scan(context.Background())
	if err != nil {
		return nil, err
	}
	return provinces, nil
}

func (r *SapNhapRepository) GetAllSapNhapSiteWards() ([]model.SapNhapSiteWard, error) {
	var wards []model.SapNhapSiteWard
	err := r.db.NewSelect().
		Model(&wards).Relation("Ward").Relation("Ward").Relation("SapNhapGIS").
		Scan(context.Background())
	if err != nil {
		return nil, err
	}
	return wards, nil
}

// Insert new sapnhap province data
func (r *SapNhapRepository) InsertSapNhapSiteProvince(province *model.SapNhapSiteProvince) error {
	_, err := r.db.NewInsert().
		Model(province).
		Exec(context.Background())
	if err != nil {
		return err
	}
	return nil
}

// Insert new sapnhap ward data
func (r *SapNhapRepository) InsertSapNhapSiteWard(ward *model.SapNhapSiteWard) error {
	_, err := r.db.NewInsert().
		Model(ward).
		Exec(context.Background())
	if err != nil {
		return err
	}
	return nil
}

// FindSapNhapSiteProvinceByMaHC finds a province by its MaHC (administrative code)
func (r *SapNhapRepository) FindSapNhapSiteProvinceByMaHC(ctx context.Context, maHC int) (*model.SapNhapSiteProvince, error) {
	province := &model.SapNhapSiteProvince{}
	err := r.db.NewSelect().
		Model(province).
		Where("mahc = ?", maHC).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return province, nil
}

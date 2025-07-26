package repository

import (
	"context"
	"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
	"github.com/uptrace/bun"
	"log"
)

type VnProvincesTmpRepository struct {
	db *bun.DB
}

func NewVnProvincesTmpRepository(db *bun.DB) *VnProvincesTmpRepository {
	return &VnProvincesTmpRepository{
		db: db,
	}
}

// Get all administrative units from the tmp database
func (r *VnProvincesTmpRepository) GetAllAdministrativeUnits() []model.AdministrativeUnit {
	var result []model.AdministrativeUnit
	ctx := context.Background()
	r.db.NewSelect().Model(&result).Scan(ctx)
	return result
}

func (r *VnProvincesTmpRepository) GetAllAdministrativeRegions() []model.AdministrativeRegion {
	var result []model.AdministrativeRegion
	ctx := context.Background()
	err := r.db.NewSelect().Model(&result).Scan(ctx)
	if err != nil {
		log.Fatal("Unable to query administrative regions", err)
	}
	return result
}

func (r *VnProvincesTmpRepository) GetAllProvinces() []model.Province {
	var result []model.Province
	ctx := context.Background()
	err := r.db.NewSelect().Model(&result).Relation("AdministrativeUnit").Relation("Wards").Relation("Wards.AdministrativeUnit").Scan(ctx)
	if err != nil {
		log.Fatal("Unable to query provinces", err)
	}
	return result
}

// method to get all wards
func (r *VnProvincesTmpRepository) GetAllWards() []model.Ward {
	var result []model.Ward
	ctx := context.Background()
	err := r.db.NewSelect().Model(&result).Scan(ctx)
	if err != nil {
		log.Fatal("Unable to query wards", err)
	}
	return result
}

func (r *VnProvincesTmpRepository) InsertWard(ctx context.Context, ward *model.Ward) error {
	_, err := r.db.NewInsert().
		Model(ward).
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *VnProvincesTmpRepository) InsertProvince(ctx context.Context, province *model.Province) error {
	_, err := r.db.NewInsert().
		Model(province).
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

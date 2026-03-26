package repository

import (
	"context"

	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
	"github.com/uptrace/bun"
)

type SapNhapGeoJSONObjectRepository struct {
	db *bun.DB
}

func NewSapNhapGeoJSONObjectRepository(db *bun.DB) *SapNhapGeoJSONObjectRepository {
	return &SapNhapGeoJSONObjectRepository{
		db: db,
	}
}

func (r *SapNhapGeoJSONObjectRepository) GetAllSapNhapGeoJSONObjects() ([]*model.SapNhapSiteGeoUnit, error) {
	var geoObjects []*model.SapNhapSiteGeoUnit
	
	err := r.db.NewSelect().
		Model(&geoObjects).
		Relation("Parent").
		Relation("VNProvince").
		Relation("VNWard").
		Scan(context.Background())
	
	if err != nil {
		return nil, err
	}
	
	return geoObjects, nil
}

// UpdateSapNhapGeoJSONObjectWKT updates the WKT geometry fields (bbox_wkt and geom_wkt) for a single record.
// The geometry columns (bbox and geom) are PostgreSQL generated columns that will automatically
// update from these WKT text fields via ST_GeomFromText().
func (r *SapNhapGeoJSONObjectRepository) UpdateSapNhapGeoJSONObjectWKT(ctx context.Context, ma string, bboxWKT, geomWKT string) error {
	_, err := r.db.NewUpdate().
		Model((*model.SapNhapSiteGeoUnit)(nil)).
		Set("bbox_wkt = ?", bboxWKT).
		Set("geom_wkt = ?", geomWKT).
		Where("ma = ?", ma).
		Exec(ctx)
	
	return err
}

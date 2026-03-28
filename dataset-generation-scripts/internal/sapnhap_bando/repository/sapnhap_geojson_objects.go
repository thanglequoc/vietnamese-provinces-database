package repository

import (
	"context"
	"fmt"
	"strings"

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

// CodeUpdate represents a single code update for batch processing
type CodeUpdate struct {
	Ma               string
	VNDSProvinceCode string
	VNDSWardCode     string
}

// BatchUpdateProvinceAndWardCodes updates multiple records with province and ward codes in a single batch operation.
// This method uses a CASE statement to efficiently update multiple records at once.
func (r *SapNhapGeoJSONObjectRepository) BatchUpdateProvinceAndWardCodes(ctx context.Context, updates []CodeUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	// Build CASE statement for vn_ds_province_code
	provinceCase := "CASE ma"
	for _, update := range updates {
		if update.VNDSProvinceCode != "" {
			provinceCase += fmt.Sprintf(" WHEN '%s' THEN '%s'", update.Ma, update.VNDSProvinceCode)
		}
	}
	provinceCase += " ELSE vn_ds_province_code END"

	// Build CASE statement for vn_ds_ward_code
	wardCase := "CASE ma"
	for _, update := range updates {
		if update.VNDSWardCode != "" {
			wardCase += fmt.Sprintf(" WHEN '%s' THEN '%s'", update.Ma, update.VNDSWardCode)
		}
	}
	wardCase += " ELSE vn_ds_ward_code END"

	// Build IN clause for WHERE clause
	maValues := make([]string, len(updates))
	for i, update := range updates {
		maValues[i] = fmt.Sprintf("'%s'", update.Ma)
	}
	whereClause := fmt.Sprintf("ma IN (%s)", strings.Join(maValues, ", "))

	// Execute batch update
	query := fmt.Sprintf(
		"UPDATE sapnhap_geojson_objects SET vn_ds_province_code = %s, vn_ds_ward_code = %s WHERE %s",
		provinceCase, wardCase, whereClause,
	)

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to execute batch update: %w", err)
	}

	return nil
}

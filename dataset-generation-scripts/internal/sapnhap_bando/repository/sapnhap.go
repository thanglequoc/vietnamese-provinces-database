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
		Model(&provinces).Relation("Province").Relation("SapNhapGIS", func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.
			Column("stt", "ten", "truocsn", "gis_server_id", "sapnhap_province_matinh").
			// Note: no table alias before bbox/gis_geom in the expr,
			// and alias result as RelationName__FieldName
			ColumnExpr(`ST_AsText(bbox) AS "SapNhapGIS__bbox"`).
			ColumnExpr(`ST_AsText(gis_geom) AS "SapNhapGIS__gis_geom"`)
	}).
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

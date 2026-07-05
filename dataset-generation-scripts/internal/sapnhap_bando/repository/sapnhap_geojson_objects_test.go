package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	dbpkg "github.com/thanglequoc-vn-provinces/v2/internal/database"
	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
	vnmodel "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
	"github.com/uptrace/bun"
)

func TestUpdateSapNhapGeoJSONObjectGeomWKT_DerivesBBoxFromGeom(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, tx.Rollback())
	})

	repo := NewSapNhapGeoJSONObjectRepository(tx)

	row := &model.SapNhapSiteGeoUnit{
		Ma:      "test_geom_bbox_derived",
		Ten:     "Test Geom Derived BBox",
		MaLK:    "test_geom_bbox_derived_malk",
		BBoxWKT: "POLYGON((105.1 20.1,105.1 20.2,105.2 20.2,105.2 20.1,105.1 20.1))",
		GeomWKT: "MULTIPOLYGON(((105.1 20.1,105.1 20.2,105.2 20.2,105.2 20.1,105.1 20.1)))",
	}

	insertTestGeoObject(t, tx, row)

	geomWKT := "MULTIPOLYGON(((105.59755 20.880711, 105.59755 20.955745, 105.677417 20.955745, 105.677417 20.880711, 105.59755 20.880711)))"

	err = repo.UpdateSapNhapGeoJSONObjectGeomWKT(ctx, row.Ma, geomWKT)
	require.NoError(t, err)

	updatedRow := getTestGeoObjectByMa(t, tx, row.Ma)
	assert.Equal(t, geomWKT, updatedRow.GeomWKT)
	assert.Equal(t,
		"POLYGON((105.59755 20.880711,105.59755 20.955745,105.677417 20.955745,105.677417 20.880711,105.59755 20.880711))",
		updatedRow.BBoxWKT,
	)
	assert.Equal(t, 0, countBBoxMismatchesForMaList(t, tx, row.Ma))
}

func TestGetAllSapNhapGeoJSONObjects_IncludeGeoJSONFragments(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, tx.Rollback())
	})

	_, err = tx.ExecContext(ctx, `INSERT INTO administrative_units (id, full_name, full_name_en, short_name, short_name_en, code_name, code_name_en)
		VALUES (2, 'Tỉnh', 'Province', 'Tỉnh', 'Province', 'tinh', 'province')
		ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	_, err = tx.NewInsert().Model(&vnmodel.Province{
		Code:                 "test_p1",
		Name:                 "Test Province",
		NameEn:               "Test Province",
		FullName:             "Test Province",
		FullNameEn:           "Test Province",
		CodeName:             "test_province",
		AdministrativeUnitId: 2,
	}).Exec(ctx)
	require.NoError(t, err)

	_, err = tx.ExecContext(ctx, `INSERT INTO administrative_units (id, full_name, full_name_en, short_name, short_name_en, code_name, code_name_en)
		VALUES (3, 'Phường', 'Ward', 'Phường', 'Ward', 'phuong', 'ward')
		ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	_, err = tx.NewInsert().Model(&vnmodel.Ward{
		Code:                 "test_w1",
		Name:                 "Test Ward",
		NameEn:               "Test Ward",
		FullName:             "Test Ward",
		FullNameEn:           "Test Ward",
		CodeName:             "test_ward",
		ProvinceCode:         "test_p1",
		AdministrativeUnitId: 3,
	}).Exec(ctx)
	require.NoError(t, err)

	insertTestGeoObject(t, tx, &model.SapNhapSiteGeoUnit{
		Ma:               "test_province_geojson_01",
		Ten:              "Test Province",
		MaLK:             "test_province_malk",
		DienTichKM2:      3359.84,
		VNDSProvinceCode: "test_p1",
		BBoxWKT:          "POLYGON((102.1 20.1,102.1 21.3,103.2 21.3,103.2 20.1,102.1 20.1))",
		GeomWKT:          "MULTIPOLYGON(((102.1 20.1,103.2 20.1,103.2 21.3,102.1 21.3,102.1 20.1)))",
	})

	insertTestGeoObject(t, tx, &model.SapNhapSiteGeoUnit{
		Ma:               "test_ward_geojson_01",
		Ten:              "Test Ward",
		MaLK:             "test_ward_malk",
		DienTichKM2:      2.97,
		VNDSProvinceCode: "test_p1",
		VNDSWardCode:     "test_w1",
		BBoxWKT:          "POLYGON((105.1 20.1,105.1 20.2,105.2 20.2,105.2 20.1,105.1 20.1))",
		GeomWKT:          "MULTIPOLYGON(((105.1 20.1,105.2 20.1,105.2 20.2,105.1 20.2,105.1 20.1)))",
	})

	repo := NewSapNhapGeoJSONObjectRepository(tx)

	provinces, err := repo.GetAllSapNhapGeoJSONProvinces(ctx)
	require.NoError(t, err)
	var province *model.SapNhapSiteGeoUnit
	for _, candidate := range provinces {
		if candidate.VNDSProvinceCode == "test_p1" {
			province = candidate
			break
		}
	}
	require.NotNil(t, province)
	assert.NotEmpty(t, province.BBoxGeoJSON)
	assert.NotEmpty(t, province.GeomGeoJSON)
	assert.True(t, json.Valid(province.BBoxGeoJSON))
	assert.True(t, json.Valid(province.GeomGeoJSON))
	assert.Equal(t, "test_province", province.VNProvince.CodeName)

	wards, err := repo.GetAllSapNhapGeoJSONWards(ctx)
	require.NoError(t, err)
	var ward *model.SapNhapSiteGeoUnit
	for _, candidate := range wards {
		if candidate.VNDSWardCode == "test_w1" {
			ward = candidate
			break
		}
	}
	require.NotNil(t, ward)
	assert.NotEmpty(t, ward.BBoxGeoJSON)
	assert.NotEmpty(t, ward.GeomGeoJSON)
	assert.True(t, json.Valid(ward.BBoxGeoJSON))
	assert.True(t, json.Valid(ward.GeomGeoJSON))
	assert.Equal(t, "test_ward", ward.VNWard.CodeName)
}

func setupTestDB(t *testing.T) *bun.DB {
	t.Helper()
	originalWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir("../../../"))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(originalWD))
	})
	require.NoError(t, godotenv.Load(".env"))

	db := dbpkg.GetPostgresDBConnection()
	ctx := context.Background()

	// Initialize PostGIS extension (for CI environments)
	_, err = db.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS postgis;")
	if err != nil {
		t.Logf("Warning: Could not create postgis extension: %v", err)
	}

	// Initialize sapnhap_geojson_objects table if it doesn't exist
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS sapnhap_geojson_objects (
			ma VARCHAR(50) PRIMARY KEY,
			ten VARCHAR(255) NOT NULL,
			magoc VARCHAR(50),
			malk VARCHAR(255),
			dientichkm2 FLOAT8,
			truocsapnhap VARCHAR(255),
			vn_ds_province_code VARCHAR(20),
			vn_ds_ward_code VARCHAR(20),
			bbox_wkt TEXT,
			geom_wkt TEXT,
			bbox GEOMETRY(POLYGON, 4326),
			geom GEOMETRY(MULTIPOLYGON, 4326)
		)
	`)
	if err != nil {
		t.Logf("Warning: Could not create sapnhap_geojson_objects table: %v", err)
	}

	return db
}

func insertTestGeoObject(t *testing.T, db bun.IDB, geoObject *model.SapNhapSiteGeoUnit) {
	t.Helper()
	columns := []string{"ma", "ten", "malk", "dientichkm2", "bbox_wkt", "geom_wkt"}
	if geoObject.VNDSProvinceCode != "" {
		columns = append(columns, "vn_ds_province_code")
	}
	if geoObject.VNDSWardCode != "" {
		columns = append(columns, "vn_ds_ward_code")
	}

	_, err := db.NewInsert().
		Model(geoObject).
		Column(columns...).
		Exec(context.Background())
	require.NoError(t, err)
}

func getTestGeoObjectByMa(t *testing.T, db bun.IDB, ma string) *model.SapNhapSiteGeoUnit {
	t.Helper()

	var geoObject model.SapNhapSiteGeoUnit
	err := db.NewSelect().
		Model(&geoObject).
		Where("ma = ?", ma).
		Scan(context.Background())
	require.NoError(t, err)

	return &geoObject
}

func countBBoxMismatchesForMaList(t *testing.T, db bun.IDB, maList ...string) int {
	t.Helper()

	var count int
	err := db.NewSelect().
		Model((*model.SapNhapSiteGeoUnit)(nil)).
		ColumnExpr("COUNT(*)").
		Where("ma IN (?)", bun.In(maList)).
		Where("geom IS NOT NULL").
		Where("(bbox IS NULL OR NOT ST_Equals(bbox, ST_Envelope(geom)))").
		Scan(context.Background(), &count)
	require.NoError(t, err)

	return count
}

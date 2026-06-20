package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	dbpkg "github.com/thanglequoc-vn-provinces/v2/internal/database"
	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
	"github.com/uptrace/bun"
)

func TestCorrectMismatchedBBoxWKTFromGeom(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, tx.Rollback())
	})

	repo := NewSapNhapGeoJSONObjectRepository(tx)

	badRow := &model.SapNhapSiteGeoUnit{
		Ma:      "test_bbox_bad",
		Ten:     "Test BBox Bad",
		MaLK:    "test_malk_bad",
		BBoxWKT: "POLYGON((-1 -1, -1 20.955745, 105.677417 20.955745, 105.677417 -1, -1 -1))",
		GeomWKT: "MULTIPOLYGON(((105.59755 20.880711, 105.59755 20.955745, 105.677417 20.955745, 105.677417 20.880711, 105.59755 20.880711)))",
	}
	goodRow := &model.SapNhapSiteGeoUnit{
		Ma:      "test_bbox_good",
		Ten:     "Test BBox Good",
		MaLK:    "test_malk_good",
		BBoxWKT: "POLYGON((105.573681 20.935366, 105.573681 21.013438, 105.636725 21.013438, 105.636725 20.935366, 105.573681 20.935366))",
		GeomWKT: "MULTIPOLYGON(((105.573681 20.935366, 105.573681 21.013438, 105.636725 21.013438, 105.636725 20.935366, 105.573681 20.935366)))",
	}

	insertTestGeoObject(t, tx, badRow)
	insertTestGeoObject(t, tx, goodRow)

	updatedCount, err := repo.CorrectMismatchedBBoxWKTFromGeom(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, updatedCount, 1)

	updatedBadRow := getTestGeoObjectByMa(t, tx, badRow.Ma)
	updatedGoodRow := getTestGeoObjectByMa(t, tx, goodRow.Ma)

	assert.Equal(t,
		"POLYGON((105.59755 20.880711,105.59755 20.955745,105.677417 20.955745,105.677417 20.880711,105.59755 20.880711))",
		updatedBadRow.BBoxWKT,
	)
	assert.Equal(t, goodRow.BBoxWKT, updatedGoodRow.BBoxWKT)
	assert.Equal(t, 0, countBBoxMismatchesForMaList(t, tx, badRow.Ma, goodRow.Ma))
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
	return dbpkg.GetPostgresDBConnection()
}

func insertTestGeoObject(t *testing.T, db bun.IDB, geoObject *model.SapNhapSiteGeoUnit) {
	t.Helper()
	_, err := db.NewInsert().
		Model(geoObject).
		Column("ma", "ten", "malk", "bbox_wkt", "geom_wkt").
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

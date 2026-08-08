package repository

import (
	"context"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	dbpkg "github.com/thanglequoc-vn-provinces/v2/internal/database"
	"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
	"github.com/uptrace/bun"
)

func setupTestDB(t *testing.T) *bun.DB {
	t.Helper()
	originalWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir("../../../"))
	t.Cleanup(func() { require.NoError(t, os.Chdir(originalWD)) })
	require.NoError(t, godotenv.Load(".env"))
	return dbpkg.GetPostgresDBConnection()
}

func TestUpdateProvincePostalCodePrefix(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	_, err := db.ExecContext(ctx, "DELETE FROM wards_tmp WHERE province_code = 'ZZ'")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "DELETE FROM provinces_tmp WHERE code = 'ZZ'")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "INSERT INTO provinces_tmp(code, name, full_name, administrative_unit_id) VALUES ('ZZ','Test Prov','Tỉnh Test Prov',2)")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM wards_tmp WHERE province_code = 'ZZ'")
		_, _ = db.ExecContext(ctx, "DELETE FROM provinces_tmp WHERE code = 'ZZ'")
	})

	repo := NewPostalCodeRepository(db)
	err = repo.UpdateProvincePostalCodePrefix(ctx, "ZZ", "10, 11")
	require.NoError(t, err)

	var prefix string
	err = db.NewSelect().Column("postal_code_prefix").Model((*model.Province)(nil)).Where("code = 'ZZ'").Scan(ctx, &prefix)
	require.NoError(t, err)
	assert.Equal(t, "10, 11", prefix)
}

func TestUpdateWardPostalCode(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	_, err := db.ExecContext(ctx, "DELETE FROM wards_tmp WHERE province_code = 'ZZ'")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "DELETE FROM provinces_tmp WHERE code = 'ZZ'")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "INSERT INTO provinces_tmp(code, name, full_name, administrative_unit_id) VALUES ('ZZ','Test Prov','Tỉnh Test Prov',2)")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "INSERT INTO wards_tmp(code, name, province_code, administrative_unit_id) VALUES ('99999','Test Ward','ZZ',4)")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM wards_tmp WHERE province_code = 'ZZ'")
		_, _ = db.ExecContext(ctx, "DELETE FROM provinces_tmp WHERE code = 'ZZ'")
	})

	repo := NewPostalCodeRepository(db)
	err = repo.UpdateWardPostalCode(ctx, "99999", "11024")
	require.NoError(t, err)

	var postal string
	err = db.NewSelect().Column("postal_code").Model((*model.Ward)(nil)).Where("code = '99999'").Scan(ctx, &postal)
	require.NoError(t, err)
	assert.Equal(t, "11024", postal)
}

func TestCountMissing(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	repo := NewPostalCodeRepository(db)
	missingProvinces, err := repo.CountProvincesMissingPostalPrefix(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, missingProvinces, 0)

	missingWards, err := repo.CountWardsMissingPostalCode(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, missingWards, 0)
}

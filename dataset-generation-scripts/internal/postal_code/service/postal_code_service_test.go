package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dbpkg "github.com/thanglequoc-vn-provinces/v2/internal/database"
	"github.com/thanglequoc-vn-provinces/v2/internal/postal_code/repository"
	"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
	vnRepo "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/repository"
)

func setupServiceTestDB(t *testing.T) {
	t.Helper()
	originalWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir("../../../"))
	t.Cleanup(func() { require.NoError(t, os.Chdir(originalWD)) })
}

func TestImportPostalCodes(t *testing.T) {
	setupServiceTestDB(t)

	// Seed data: temp files instead of committed seeds, so the test is hermetic.
	seedDir := t.TempDir()
	provSeed := filepath.Join(seedDir, "province_postal_code_prefixes.json")
	wardSeed := filepath.Join(seedDir, "ward_postal_codes.json")
	require.NoError(t, os.WriteFile(provSeed, []byte(`[{"code":"ZZ","postal_code_prefix":"10, 11"}]`), 0644))
	require.NoError(t, os.WriteFile(wardSeed, []byte(`[{"province_code":"ZZ","name":"Test Ward","postal_code":"11024"}]`), 0644))

	ctx := context.Background()
	db := dbpkg.GetPostgresDBConnection()

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

	vnTmpRepo := vnRepo.NewVnProvincesTmpRepository(db)
	postalRepo := repository.NewPostalCodeRepository(db)
	svc := NewPostalCodeService(vnTmpRepo, postalRepo, seedDir)

	err = svc.ImportPostalCodes(ctx)
	require.NoError(t, err)

	var prefix string
	require.NoError(t, db.NewSelect().Column("postal_code_prefix").Model((*model.Province)(nil)).Where("code = 'ZZ'").Scan(ctx, &prefix))
	assert.Equal(t, "10, 11", prefix)

	var postal string
	require.NoError(t, db.NewSelect().Column("postal_code").Model((*model.Ward)(nil)).Where("code = '99999'").Scan(ctx, &postal))
	assert.Equal(t, "11024", postal)
}

func TestImportPostalCodes_ToneVariantMatch(t *testing.T) {
	setupServiceTestDB(t)

	seedDir := t.TempDir()
	provSeed := filepath.Join(seedDir, "province_postal_code_prefixes.json")
	wardSeed := filepath.Join(seedDir, "ward_postal_codes.json")
	require.NoError(t, os.WriteFile(provSeed, []byte(`[{"code":"ZZ","postal_code_prefix":"10, 11"}]`), 0644))
	// Seed uses "Hòa" (decree spelling); DB stores "Hoà" (convention). Must still match.
	require.NoError(t, os.WriteFile(wardSeed, []byte(`[{"province_code":"ZZ","name":"Bình Hòa","postal_code":"90908"}]`), 0644))

	ctx := context.Background()
	db := dbpkg.GetPostgresDBConnection()

	_, err := db.ExecContext(ctx, "DELETE FROM wards_tmp WHERE province_code = 'ZZ'")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "DELETE FROM provinces_tmp WHERE code = 'ZZ'")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "INSERT INTO provinces_tmp(code, name, full_name, administrative_unit_id) VALUES ('ZZ','Test Prov','Tỉnh Test Prov',2)")
	require.NoError(t, err)
	// DB name uses "Hoà"; name_en is tone-stripped "Binh Hoa".
	_, err = db.ExecContext(ctx, "INSERT INTO wards_tmp(code, name, name_en, province_code, administrative_unit_id) VALUES ('99998','Bình Hoà','Binh Hoa','ZZ',4)")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM wards_tmp WHERE province_code = 'ZZ'")
		_, _ = db.ExecContext(ctx, "DELETE FROM provinces_tmp WHERE code = 'ZZ'")
	})

	vnTmpRepo := vnRepo.NewVnProvincesTmpRepository(db)
	postalRepo := repository.NewPostalCodeRepository(db)
	svc := NewPostalCodeService(vnTmpRepo, postalRepo, seedDir)

	require.NoError(t, svc.ImportPostalCodes(ctx))

	var postal string
	require.NoError(t, db.NewSelect().Column("postal_code").Model((*model.Ward)(nil)).Where("code = '99998'").Scan(ctx, &postal))
	assert.Equal(t, "90908", postal)
}

func TestImportPostalCodes_CollisionDisambiguatesByExactName(t *testing.T) {
	setupServiceTestDB(t)

	seedDir := t.TempDir()
	provSeed := filepath.Join(seedDir, "province_postal_code_prefixes.json")
	wardSeed := filepath.Join(seedDir, "ward_postal_codes.json")
	require.NoError(t, os.WriteFile(provSeed, []byte(`[{"code":"ZZ","postal_code_prefix":"10, 11"}]`), 0644))
	// Two seeds in the same province with tone-distinct names; exact match must win.
	require.NoError(t, os.WriteFile(wardSeed, []byte(`[{"province_code":"ZZ","name":"Văn Lang","postal_code":"23811"},{"province_code":"ZZ","name":"Văn Lăng","postal_code":"24211"}]`), 0644))

	ctx := context.Background()
	db := dbpkg.GetPostgresDBConnection()

	_, err := db.ExecContext(ctx, "DELETE FROM wards_tmp WHERE province_code = 'ZZ'")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "DELETE FROM provinces_tmp WHERE code = 'ZZ'")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "INSERT INTO provinces_tmp(code, name, full_name, administrative_unit_id) VALUES ('ZZ','Test Prov','Tỉnh Test Prov',2)")
	require.NoError(t, err)
	// Both wards share the same stripped name_en "Van Lang".
	_, err = db.ExecContext(ctx, "INSERT INTO wards_tmp(code, name, name_en, province_code, administrative_unit_id) VALUES ('99001','Văn Lang','Van Lang','ZZ',4)")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "INSERT INTO wards_tmp(code, name, name_en, province_code, administrative_unit_id) VALUES ('99002','Văn Lăng','Van Lang','ZZ',4)")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM wards_tmp WHERE province_code = 'ZZ'")
		_, _ = db.ExecContext(ctx, "DELETE FROM provinces_tmp WHERE code = 'ZZ'")
	})

	vnTmpRepo := vnRepo.NewVnProvincesTmpRepository(db)
	postalRepo := repository.NewPostalCodeRepository(db)
	svc := NewPostalCodeService(vnTmpRepo, postalRepo, seedDir)

	require.NoError(t, svc.ImportPostalCodes(ctx))

	var langPostal, langPostal2 string
	require.NoError(t, db.NewSelect().Column("postal_code").Model((*model.Ward)(nil)).Where("code = '99001'").Scan(ctx, &langPostal))
	require.NoError(t, db.NewSelect().Column("postal_code").Model((*model.Ward)(nil)).Where("code = '99002'").Scan(ctx, &langPostal2))
	assert.Equal(t, "23811", langPostal)
	assert.Equal(t, "24211", langPostal2)
}

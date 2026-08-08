# Add Postal Codes to Vietnamese Provinces Database — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add national postal codes (Quyết định 2334/QĐ-BKHCN) as `postal_code_prefix` on provinces and `postal_code` on wards, populated from committed seed files, and include them in every generated dataset output.

**Architecture:** New `internal/postal_code/` package (entrypoint → service → repository) imports seed JSON from `resources/postal/` into the temporary Postgres DB after the DVHCVN dump and before the dataset writers. The `Province`/`Ward` Bun models gain the two columns, so all existing writers pick them up. Each writer's DTO/mapper/template is extended, plus GIS outputs (GeoJSON properties, `provinces-gis`/`wards-gis` docs).

**Tech Stack:** Go 1.24, Bun ORM + pgdialect, Testify, `internal/common/viet` (Vietnamese normalization), Python 3 (seed generation).

## Global Constraints

- Go module path: `github.com/thanglequoc-vn-provinces/v2`
- DB: PostgreSQL/PostGIS in Docker container `vn_provinces_postgres_container`, db `vn_provinces_tmp`, port `15432`
- Both new columns are **nullable**: `provinces_tmp.postal_code_prefix varchar(255) NULL`, `wards_tmp.postal_code varchar(20) NULL`
- No new table, no new indexes, no changes to `sapnhap_geojson_objects` or SQL GIS tables (`gis_provinces`/`gis_wards`)
- All commits must be on branch `134_AddPostalCode`
- Tests requiring Docker: `go test -v ./...` from `dataset-generation-scripts/`

---

### Task 1: Generate postal code seed files

**Files:**
- Create: `development/134_AddPostalCode/generate_postal_seeds.py`
- Create (generated): `dataset-generation-scripts/resources/postal/province_postal_code_prefixes.json`
- Create (generated): `dataset-generation-scripts/resources/postal/ward_postal_codes.json`

**Interfaces:**
- Consumes: `development/134_AddPostalCode/postal_codes.md`, `development/134_AddPostalCode/db_wards.tsv`
- Produces: two JSON seed files consumed by Task 4

- [ ] **Step 1: Write the seed generation script**

Create `development/134_AddPostalCode/generate_postal_seeds.py`:

```python
#!/usr/bin/env python3
"""Generate postal code seed JSON files from postal_codes.md and db_wards.tsv.

Outputs:
  ../../dataset-generation-scripts/resources/postal/province_postal_code_prefixes.json
  ../../dataset-generation-scripts/resources/postal/ward_postal_codes.json

The province summary table gives 2-digit prefix strings per province. The per-
province tables give 5-digit postal codes per ward. Province codes are resolved
by normalizing province names from db_wards.tsv.
"""
import json
import os
import re
import unicodedata

POSTAL_MD = "postal_codes.md"
DB_TSV = "db_wards.tsv"
OUT_DIR = "../../dataset-generation-scripts/resources/postal"


def strip_tone(s):
    s = unicodedata.normalize("NFD", s)
    s = "".join(c for c in s if not unicodedata.combining(c))
    s = s.replace("Đ", "D").replace("đ", "d")
    return s


def norm(s):
    s = s.strip().lower()
    s = s.replace("tp. ", "").replace("tỉnh ", "")
    s = re.sub(r"\s+", "", s)
    s = strip_tone(s)
    s = re.sub(r"['''-]", "", s)
    return s


def cells(line):
    inner = line.strip()
    if inner.startswith("|"):
        inner = inner[1:]
    if inner.endswith("|"):
        inner = inner[:-1]
    return [c.strip() for c in inner.split("|")]


def strip_unit_prefix(name):
    name = re.sub(r"^(X\.|P\.|TT\.|TX\.)\s*", "", name.strip()).strip()
    return re.sub(r"^Đặc\s+khu\s+", "", name).strip()


def load_province_code_map():
    code_by_norm = {}
    with open(DB_TSV, encoding="utf-8") as f:
        for line in f:
            if line.startswith(" code ") or line.startswith("-+-"):
                continue
            parts = [p.strip() for p in line.split("|")]
            if len(parts) < 6:
                continue
            try:
                int(parts[0])
            except ValueError:
                continue
            prov_name = parts[5].strip()
            prov_code = parts[3].strip()
            key = norm(prov_name)
            if key not in code_by_norm:
                code_by_norm[key] = prov_code
    return code_by_norm


def parse_province_prefixes(lines, code_by_norm):
    """Parse the 'Mã bưu chính cấp tỉnh' summary table."""
    result = []
    in_summary = False
    for line in lines:
        if line.strip().startswith("## "):
            in_summary = False
            continue
        c = cells(line)
        if any("Tên tỉnh" in x for x in c):
            in_summary = True
            continue
        if not in_summary:
            continue
        if len(c) < 3:
            continue
        try:
            int(c[0])
        except ValueError:
            continue
        name = c[1]
        prefix = c[2]
        if not name:
            continue
        code = code_by_norm.get(norm(name))
        if code is None:
            raise SystemExit(f"UNMATCHED province in summary: {name}")
        result.append({"code": code, "postal_code_prefix": prefix})
    return result


def parse_ward_postal_codes(lines, code_by_norm):
    result = []
    cur_prov_code = None
    for line in lines:
        stripped = line.strip()
        m = re.match(r"^##\s+(TỈNH|TP\.)\s+(.+)$", stripped)
        if m:
            prov_name = f"{m.group(1)} {m.group(2)}"
            cur_prov_code = code_by_norm.get(norm(prov_name))
            if cur_prov_code is None:
                raise SystemExit(f"UNMATCHED province header: {prov_name}")
            continue
        c = cells(line)
        if len(c) < 3 or cur_prov_code is None:
            continue
        try:
            int(c[0])
        except ValueError:
            continue
        name_raw = c[1]
        postal = c[2]
        if not re.fullmatch(r"\d{5}", postal):
            continue
        name = strip_unit_prefix(name_raw)
        if not name:
            continue
        result.append({"province_code": cur_prov_code, "name": name, "postal_code": postal})
    return result


def main():
    lines = open(POSTAL_MD, encoding="utf-8").read().splitlines()
    code_by_norm = load_province_code_map()

    prefixes = parse_province_prefixes(lines, code_by_norm)
    wards = parse_ward_postal_codes(lines, code_by_norm)

    os.makedirs(OUT_DIR, exist_ok=True)
    with open(os.path.join(OUT_DIR, "province_postal_code_prefixes.json"), "w", encoding="utf-8") as f:
        json.dump(prefixes, f, ensure_ascii=False, indent=2)
    with open(os.path.join(OUT_DIR, "ward_postal_codes.json"), "w", encoding="utf-8") as f:
        json.dump(wards, f, ensure_ascii=False, indent=2)

    print(f"province prefixes: {len(prefixes)} (expected 34)")
    print(f"ward postal codes: {len(wards)} (expected 3321)")


if __name__ == "__main__":
    main()
```

- [ ] **Step 2: Run the script and verify counts**

Run: `python3 generate_postal_seeds.py`
Expected: prints `province prefixes: 34` and `ward postal codes: 3321`.

- [ ] **Step 3: Sanity-check the generated files**

```bash
head -5 ../../dataset-generation-scripts/resources/postal/province_postal_code_prefixes.json
head -5 ../../dataset-generation-scripts/resources/postal/ward_postal_codes.json
```

Verify `province_postal_code_prefixes.json` first entry is `{"code": "91", "postal_code_prefix": "90, 91, 92"}` (An Giang) and `ward_postal_codes.json` contains `{"province_code": "01", "name": "Hoàn Kiếm", "postal_code": "11024"}`.

- [ ] **Step 4: Commit**

```bash
git add development/134_AddPostalCode/generate_postal_seeds.py dataset-generation-scripts/resources/postal/
git commit -m "feat: add postal code seed files (Quyết định 2334/QĐ-BKHCN)"
```

---

### Task 2: Schema and model changes

**Files:**
- Modify: `dataset-generation-scripts/resources/db_table_init.sql`
- Modify: `postgresql/postgres_CreateTables_vn_units.sql`
- Modify: `mysql/mysql_CreateTables_vn_units.sql`
- Modify: `sqlserver/mssql_CreateTables_vn_units.sql`
- Modify: `oracle/oracle_CreateTables_vn_units.sql`
- Modify: `dataset-generation-scripts/internal/vn_provinces_tmp/model/vn_provinces_tmp_model.go`

**Interfaces:**
- Consumes: nothing
- Produces: `Province.PostalCodePrefix`, `Ward.PostalCode` fields (used by Tasks 4, 6-11)

- [ ] **Step 1: Add columns to `resources/db_table_init.sql`**

Add `postal_code_prefix varchar(255) NULL,` after the `code_name` line in the `provinces_tmp` CREATE TABLE (after line 31 `code_name varchar(255) NULL,`), and `postal_code varchar(20) NULL,` after the `code_name` line in the `wards_tmp` CREATE TABLE (after line 46 `code_name varchar(255) NULL,`).

- [ ] **Step 2: Add columns to the four published CreateTables files**

In each of `postgresql/postgres_CreateTables_vn_units.sql`, `mysql/mysql_CreateTables_vn_units.sql`, and `oracle/oracle_CreateTables_vn_units.sql`, add `postal_code_prefix varchar(255) NULL,` after the `code_name varchar(255) NULL,` line in the `provinces` table, and `postal_code varchar(20) NULL,` after the `code_name varchar(255) NULL,` line in the `wards` table.

In `sqlserver/mssql_CreateTables_vn_units.sql`, use `nvarchar`: add `postal_code_prefix nvarchar(255) NULL,` and `postal_code nvarchar(20) NULL,` at the same positions.

- [ ] **Step 3: Add fields to the Bun models**

In `dataset-generation-scripts/internal/vn_provinces_tmp/model/vn_provinces_tmp_model.go`, add to the `Province` struct (after `CodeName string \`bun:"code_name"\``):

```go
	PostalCodePrefix string `bun:"postal_code_prefix"`
```

Add to the `Ward` struct (after `CodeName string \`bun:"code_name"\``):

```go
	PostalCode string `bun:"postal_code"`
```

- [ ] **Step 4: Build and bootstrap the temporary DB**

Run: `go build ./...`
Expected: compiles.

Then refresh the temporary DB so the columns exist:

```bash
docker compose -f docker/docker-compose.yaml up -d
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "ALTER TABLE provinces_tmp ADD COLUMN IF NOT EXISTS postal_code_prefix varchar(255) NULL; ALTER TABLE wards_tmp ADD COLUMN IF NOT EXISTS postal_code varchar(20) NULL;"
```

Verify: `docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "\d provinces_tmp"` shows `postal_code_prefix`.

- [ ] **Step 5: Commit**

```bash
git add dataset-generation-scripts/resources/db_table_init.sql postgresql/ mysql/ sqlserver/ oracle/ dataset-generation-scripts/internal/vn_provinces_tmp/model/vn_provinces_tmp_model.go
git commit -m "feat: add postal_code/postal_code_prefix columns to schema and models"
```

---

### Task 3: postal_code repository

**Files:**
- Create: `dataset-generation-scripts/internal/postal_code/repository/postal_code_repository.go`
- Create: `dataset-generation-scripts/internal/postal_code/repository/postal_code_repository_test.go`

**Interfaces:**
- Consumes: `database.GetPostgresDBConnection()` (`internal/database`), Bun
- Produces:
  - `NewPostalCodeRepository(db bun.IDB) *PostalCodeRepository`
  - `(r *PostalCodeRepository) UpdateProvincePostalCodePrefix(ctx context.Context, code, prefix string) error`
  - `(r *PostalCodeRepository) UpdateWardPostalCode(ctx context.Context, wardCode, postalCode string) error`
  - `(r *PostalCodeRepository) CountProvincesMissingPostalPrefix(ctx context.Context) (int, error)`
  - `(r *PostalCodeRepository) CountWardsMissingPostalCode(ctx context.Context) (int, error)`

- [ ] **Step 1: Write the failing test**

Create `dataset-generation-scripts/internal/postal_code/repository/postal_code_repository_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./internal/postal_code/repository/...`
Expected: FAIL with `undefined: NewPostalCodeRepository`.

- [ ] **Step 3: Write minimal implementation**

Create `dataset-generation-scripts/internal/postal_code/repository/postal_code_repository.go`:

```go
package repository

import (
	"context"

	"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
	"github.com/uptrace/bun"
)

type PostalCodeRepository struct {
	db bun.IDB
}

func NewPostalCodeRepository(db bun.IDB) *PostalCodeRepository {
	return &PostalCodeRepository{db: db}
}

func (r *PostalCodeRepository) UpdateProvincePostalCodePrefix(ctx context.Context, code, prefix string) error {
	_, err := r.db.NewUpdate().
		Model((*model.Province)(nil)).
		Set("postal_code_prefix = ?", prefix).
		Where("code = ?", code).
		Exec(ctx)
	return err
}

func (r *PostalCodeRepository) UpdateWardPostalCode(ctx context.Context, wardCode, postalCode string) error {
	_, err := r.db.NewUpdate().
		Model((*model.Ward)(nil)).
		Set("postal_code = ?", postalCode).
		Where("code = ?", wardCode).
		Exec(ctx)
	return err
}

func (r *PostalCodeRepository) CountProvincesMissingPostalPrefix(ctx context.Context) (int, error) {
	return r.db.NewSelect().
		Model((*model.Province)(nil)).
		Where("postal_code_prefix IS NULL OR postal_code_prefix = ''").
		Count(ctx)
}

func (r *PostalCodeRepository) CountWardsMissingPostalCode(ctx context.Context) (int, error) {
	return r.db.NewSelect().
		Model((*model.Ward)(nil)).
		Where("postal_code IS NULL OR postal_code = ''").
		Count(ctx)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -v ./internal/postal_code/repository/...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/postal_code/repository/
git commit -m "feat: add postal code repository"
```

---

### Task 4: postal_code service

**Files:**
- Create: `dataset-generation-scripts/internal/postal_code/service/postal_code_service.go`
- Create: `dataset-generation-scripts/internal/postal_code/service/postal_code_service_test.go`

**Interfaces:**
- Consumes: `PostalCodeRepository` (Task 3), `VnProvincesTmpRepository` (`internal/vn_provinces_tmp/repository`), seed JSON files (`resources/postal/`)
- Produces:
  - `NewPostalCodeService(vnRepo *vnRepo.VnProvincesTmpRepository, postalRepo *repository.PostalCodeRepository, seedDir string) *PostalCodeService`
  - `(s *PostalCodeService) ImportPostalCodes(ctx context.Context) error`

- [ ] **Step 1: Write the failing test**

Create `dataset-generation-scripts/internal/postal_code/service/postal_code_service_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./internal/postal_code/service/...`
Expected: FAIL with `undefined: NewPostalCodeService`.

- [ ] **Step 3: Write minimal implementation**

Create `dataset-generation-scripts/internal/postal_code/service/postal_code_service.go`:

```go
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/thanglequoc-vn-provinces/v2/internal/common/viet"
	"github.com/thanglequoc-vn-provinces/v2/internal/postal_code/repository"
	vnRepo "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/repository"
)

type provincePrefixSeed struct {
	Code             string `json:"code"`
	PostalCodePrefix string `json:"postal_code_prefix"`
}

type wardPostalCodeSeed struct {
	ProvinceCode string `json:"province_code"`
	Name         string `json:"name"`
	PostalCode   string `json:"postal_code"`
}

type PostalCodeService struct {
	vnProvinceTmpRepo *vnRepo.VnProvincesTmpRepository
	postalCodeRepo    *repository.PostalCodeRepository
	seedDir           string
}

func NewPostalCodeService(
	vnRepo *vnRepo.VnProvincesTmpRepository,
	postalRepo *repository.PostalCodeRepository,
	seedDir string,
) *PostalCodeService {
	return &PostalCodeService{
		vnProvinceTmpRepo: vnRepo,
		postalCodeRepo:    postalRepo,
		seedDir:           seedDir,
	}
}

// normalizeExact produces a tone-preserving NFC key: lowercased, NFC-normalized,
// with curly apostrophes folded to ASCII and whitespace collapsed.
func normalizeExact(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "’", "'")
	s = strings.ReplaceAll(s, "‘", "'")
	s = viet.NormalizeToneMarks(s)
	return strings.Join(strings.Fields(s), " ")
}

// normalizeStripped removes tone marks entirely and collapses whitespace,
// apostrophes, and hyphens. This mirrors how the DB stores `name_en`, and is
// used as a fallback when the tone-preserving exact match fails (e.g. the
// source decree writes "Hòa" while the DB normalizes to "Hoà").
func normalizeStripped(s string) string {
	s = normalizeExact(s)
	s = viet.RemoveVietToneMark(s)
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, "-", "")
	return strings.Join(strings.Fields(s), "")
}

func (s *PostalCodeService) loadProvincePrefixes() ([]provincePrefixSeed, error) {
	data, err := os.ReadFile(filepath.Join(s.seedDir, "province_postal_code_prefixes.json"))
	if err != nil {
		return nil, fmt.Errorf("read province prefixes seed: %w", err)
	}
	var seeds []provincePrefixSeed
	if err := json.Unmarshal(data, &seeds); err != nil {
		return nil, fmt.Errorf("parse province prefixes seed: %w", err)
	}
	return seeds, nil
}

func (s *PostalCodeService) loadWardPostalCodes() ([]wardPostalCodeSeed, error) {
	data, err := os.ReadFile(filepath.Join(s.seedDir, "ward_postal_codes.json"))
	if err != nil {
		return nil, fmt.Errorf("read ward postal codes seed: %w", err)
	}
	var seeds []wardPostalCodeSeed
	if err := json.Unmarshal(data, &seeds); err != nil {
		return nil, fmt.Errorf("parse ward postal codes seed: %w", err)
	}
	return seeds, nil
}

// ImportPostalCodes populates postal_code_prefix on provinces and postal_code on
// wards. It verifies a 100% match rate: every ward in wards_tmp must receive a
// postal code, and the number of updated wards must equal the seed count.
func (s *PostalCodeService) ImportPostalCodes(ctx context.Context) error {
	prefixSeeds, err := s.loadProvincePrefixes()
	if err != nil {
		return err
	}
	wardSeeds, err := s.loadWardPostalCodes()
	if err != nil {
		return err
	}
	log.Printf("ℹ️ Importing postal codes: %d province prefixes, %d ward postal codes", len(prefixSeeds), len(wardSeeds))

	for _, seed := range prefixSeeds {
		if err := s.postalCodeRepo.UpdateProvincePostalCodePrefix(ctx, seed.Code, seed.PostalCodePrefix); err != nil {
			return fmt.Errorf("update province %s postal prefix: %w", seed.Code, err)
		}
	}

	provinces := s.vnProvinceTmpRepo.GetAllProvinces()

	// Validate that every seed province_code exists in provinces_tmp.
	validProvinceCodes := make(map[string]bool, len(provinces))
	for _, p := range provinces {
		validProvinceCodes[p.Code] = true
	}
	for _, seed := range wardSeeds {
		if !validProvinceCodes[seed.ProvinceCode] {
			return fmt.Errorf("seed province_code %q not found in provinces_tmp", seed.ProvinceCode)
		}
	}

	wards := s.vnProvinceTmpRepo.GetAllWards()

	// Tier 1: exact tone-preserving match, scoped by province code.
	exactKeyToCode := make(map[string]string, len(wards))
	// Tier 2: tone-stripped match against name_en, scoped by province code.
	// Multiple wards in the same province may share a stripped key only if they
	// differ by tone mark (e.g. "Văn Lang" vs "Văn Lăng"); for those we require
	// the parent province to be identical before disambiguating by exact name.
	strippedKeyToCodes := make(map[string][]string, len(wards))
	provinceOfWard := make(map[string]string, len(wards))
	for _, w := range wards {
		exactKey := w.ProvinceCode + "|" + normalizeExact(w.Name)
		if _, exists := exactKeyToCode[exactKey]; !exists {
			exactKeyToCode[exactKey] = w.Code
		}
		strippedKey := w.ProvinceCode + "|" + normalizeStripped(w.NameEn)
		strippedKeyToCodes[strippedKey] = append(strippedKeyToCodes[strippedKey], w.Code)
		provinceOfWard[w.Code] = w.ProvinceCode
	}

	matched := 0
	var unmatched []wardPostalCodeSeed
	for _, seed := range wardSeeds {
		code := ""
		// Tier 1: exact name match within the seed's province.
		exactKey := seed.ProvinceCode + "|" + normalizeExact(seed.Name)
		if c, ok := exactKeyToCode[exactKey]; ok {
			code = c
		} else {
			// Tier 2: tone-stripped match against name_en. Only accept when the
			// parent province is identical and the match is unambiguous.
			strippedKey := seed.ProvinceCode + "|" + normalizeStripped(seed.Name)
			candidates := strippedKeyToCodes[strippedKey]
			if len(candidates) == 1 {
				code = candidates[0]
			} else if len(candidates) > 1 {
				// Disambiguate by exact name among the candidates that share the
				// same parent province.
				for _, cand := range candidates {
					if provinceOfWard[cand] != seed.ProvinceCode {
						continue
					}
					candExactKey := seed.ProvinceCode + "|" + normalizeExact(seed.Name)
					if c2, ok := exactKeyToCode[candExactKey]; ok && c2 == cand {
						code = cand
						break
					}
				}
			}
		}
		if code == "" {
			unmatched = append(unmatched, seed)
			continue
		}
		if err := s.postalCodeRepo.UpdateWardPostalCode(ctx, code, seed.PostalCode); err != nil {
			return fmt.Errorf("update ward %s postal code: %w", code, err)
		}
		matched++
	}

	if len(unmatched) > 0 {
		return fmt.Errorf("postal code import failed: %d unmatched ward(s), first: %+v", len(unmatched), unmatched[0])
	}

	missingProvinces, err := s.postalCodeRepo.CountProvincesMissingPostalPrefix(ctx)
	if err != nil {
		return err
	}
	missingWards, err := s.postalCodeRepo.CountWardsMissingPostalCode(ctx)
	if err != nil {
		return err
	}
	log.Printf("✅ Postal code import complete: %d/%d wards matched, %d provinces and %d wards still missing",
		matched, len(wardSeeds), missingProvinces, missingWards)
	if matched != len(wardSeeds) {
		return fmt.Errorf("postal code import verification failed: matched=%d/%d",
			matched, len(wardSeeds))
	}
	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -v ./internal/postal_code/service/...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/postal_code/service/
git commit -m "feat: add postal code import service"
```

---

### Task 5: postal_code entrypoint and main.go wiring

**Files:**
- Create: `dataset-generation-scripts/internal/postal_code/postal_code.go`
- Modify: `dataset-generation-scripts/main.go`

**Interfaces:**
- Consumes: `PostalCodeService` (Task 4), `database.GetPostgresDBConnection()`
- Produces: `ImportPostalCodes()` (called from `main.go` after the dumper)

- [ ] **Step 1: Write the entrypoint**

Create `dataset-generation-scripts/internal/postal_code/postal_code.go`:

```go
package postal_code

import (
	"context"
	"log"

	db "github.com/thanglequoc-vn-provinces/v2/internal/database"
	"github.com/thanglequoc-vn-provinces/v2/internal/postal_code/repository"
	"github.com/thanglequoc-vn-provinces/v2/internal/postal_code/service"
	vnRepo "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/repository"
)

const seedDir = "./resources/postal"

// ImportPostalCodes reads the postal code seed files and populates
// postal_code_prefix / postal_code on provinces_tmp and wards_tmp.
// Must run after BeginDumpingDataWithDvhcvnDirectSource() and before
// ReadAndGenerateSQLDatasets().
func ImportPostalCodes() {
	postgresDB := db.GetPostgresDBConnection()
	vnTmpRepo := vnRepo.NewVnProvincesTmpRepository(postgresDB)
	postalRepo := repository.NewPostalCodeRepository(postgresDB)
	postalService := service.NewPostalCodeService(vnTmpRepo, postalRepo, seedDir)

	ctx := context.Background()
	if err := postalService.ImportPostalCodes(ctx); err != nil {
		log.Fatalf("❌ Failed to import postal codes: %v", err)
		panic(err)
	}
	log.Println("✅ Postal codes imported successfully")
}
```

- [ ] **Step 2: Wire into main.go**

In `dataset-generation-scripts/main.go`, add the import `postal_code "github.com/thanglequoc-vn-provinces/v2/internal/postal_code"` and insert `postal_code.ImportPostalCodes()` between the dumper call and the writer call:

```go
	dumper.BeginDumpingDataWithDvhcvnDirectSource()
	postal_code.ImportPostalCodes()
	dataset_writer.ReadAndGenerateSQLDatasets()
```

- [ ] **Step 3: Build**

Run: `go build ./...`
Expected: compiles.

- [ ] **Step 4: Run a full generation (verifies the real seeds import 100%)**

Run: `go run main.go`
Expected: logs `✅ Postal code import complete: 3321 wards matched` and `✅ Postal codes imported successfully` (before the writer section). GIS steps still run.

Then verify in the DB:

```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "SELECT COUNT(*) AS missing_provinces FROM provinces_tmp WHERE postal_code_prefix IS NULL OR postal_code_prefix = ''; SELECT COUNT(*) AS missing_wards FROM wards_tmp WHERE postal_code IS NULL OR postal_code = '';"
```

Expected: `0` for both. Spot-check: `SELECT code,name,postal_code FROM wards_tmp WHERE name IN ('Hoàn Kiếm','An Phú');` shows `11024` / `90456`.

- [ ] **Step 5: Commit**

```bash
git add internal/postal_code/ main.go
git commit -m "feat: wire postal code import into generation pipeline"
```

---

### Task 6: SQL writers (Postgres/MySQL, MSSQL, Oracle)

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/postgres_mysql_dataset_file_writer.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/mssql_dataset_file_writer.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/oracle_dataset_file_writer.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/postgres_mysql_dataset_file_writer_test.go`

**Interfaces:**
- Consumes: `Province.PostalCodePrefix`, `Ward.PostalCode` model fields
- Produces: SQL INSERT statements including the two new columns

- [ ] **Step 1: Write the failing tests**

In `postgres_mysql_dataset_file_writer_test.go`, add `PostalCodePrefix: "10, 11, 12, 13, 14"` to the province fixtures in `TestPostgresMySQLDatasetFileWriter_WriteToFile_Provinces` (both provinces), `PostalCode: "11024"` to the ward fixtures in `TestPostgresMySQLDatasetFileWriter_WriteToFile_Wards`, and add assertions:

```go
	assert.Contains(t, contentStr, "INSERT INTO provinces(code,name,name_en,full_name,full_name_en,code_name,administrative_unit_id,postal_code_prefix)")
	assert.Contains(t, contentStr, "('01','Hà Nội','Ha Noi','Thành phố Hà Nội','Ha Noi City','ha_noi',1,'10, 11, 12, 13, 14')")
	assert.Contains(t, contentStr, "INSERT INTO wards(code,name,name_en,full_name,full_name_en,code_name,province_code,administrative_unit_id,postal_code)")
	assert.Contains(t, contentStr, "('001','Bắc Sơn','Bac Son','Phường Bắc Sơn','Bac Son Ward','bac_son','01',3,'11024')")
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -v ./internal/dataset_writer/dataset_file_writer/ -run TestPostgresMySQLDatasetFileWriter`
Expected: FAIL (templates lack the new columns).

- [ ] **Step 3: Update the Postgres/MySQL writer templates**

In `postgres_mysql_dataset_file_writer.go`, change the constants:

```go
const insertProvinceTemplate string = "INSERT INTO provinces(code,name,name_en,full_name,full_name_en,code_name,administrative_unit_id,postal_code_prefix) VALUES"
const insertProvinceValueTemplate string = "('%s','%s','%s','%s','%s','%s',%d,'%s')"
const insertWardTemplate string = "INSERT INTO wards(code,name,name_en,full_name,full_name_en,code_name,province_code,administrative_unit_id,postal_code) VALUES"
const insertDistrictWardValueTemplate string = "('%s','%s','%s','%s','%s','%s','%s',%d,'%s')"
```

Then add a nullable-string helper in `dataset_file_writer.go` (next to `escapeSingleQuote`):

```go
// nullableSQLString returns 'value' escaped, or NULL when value is empty.
func nullableSQLString(s string) string {
	if s == "" {
		return "NULL"
	}
	return "'" + escapeSingleQuote(s) + "'"
}
```

Update the province loop's `fmt.Sprintf` call (keep the existing `dataWriter.WriteString(...)` wrapper):

```go
		dataWriter.WriteString(
			fmt.Sprintf(insertProvinceValueTemplate, p.Code, escapeSingleQuote(p.Name), escapeSingleQuote(p.NameEn), escapeSingleQuote(p.FullName),
				escapeSingleQuote(p.FullNameEn), p.CodeName, p.AdministrativeUnitId, nullableSQLString(p.PostalCodePrefix)))
```

Update the ward loop's `fmt.Sprintf` call (keep the existing `dataWriter.WriteString(...)` wrapper):

```go
		dataWriter.WriteString(
			fmt.Sprintf(insertDistrictWardValueTemplate, w.Code, escapeSingleQuote(w.Name), escapeSingleQuote(w.NameEn), escapeSingleQuote(w.FullName),
				escapeSingleQuote(w.FullNameEn), w.CodeName, w.ProvinceCode, w.AdministrativeUnitId, nullableSQLString(w.PostalCode)))
```

- [ ] **Step 4: Update the MSSQL writer templates**

In `mssql_dataset_file_writer.go`, change the constants:

```go
const insertProvinceValueMsSqlTemplate string = "('%s',N'%s',N'%s',N'%s',N'%s','%s',%d,%s)"
const insertProvinceWardValueMsSqlTemplate string = "('%s',N'%s',N'%s',N'%s',N'%s','%s','%s',%d,%s)"
```

Add a helper in `dataset_file_writer.go`:

```go
// nullableNString returns N'value' escaped, or NULL when value is empty.
func nullableNString(s string) string {
	if s == "" {
		return "NULL"
	}
	return "N'" + escapeSingleQuote(s) + "'"
}
```

Update the province `fmt.Sprintf` call to append `nullableNString(p.PostalCodePrefix)` and the ward call to append `nullableNString(w.PostalCode)` as the last argument.

- [ ] **Step 5: Update the Oracle writer templates**

In `oracle_dataset_file_writer.go`, change:

```go
const insertProvinceOracleTemplate string = "\tINTO provinces(code,name,name_en,full_name,full_name_en,code_name,administrative_unit_id,postal_code_prefix) VALUES('%s','%s','%s','%s','%s','%s',%d,%s)"
```

and the ward template:

```go
	const insertWardOracleTemplate string = "\tINTO wards(code,name,name_en,full_name,full_name_en,code_name,province_code,administrative_unit_id,postal_code) VALUES('%s','%s','%s','%s','%s','%s','%s',%d,%s)"
```

Update the province `fmt.Sprintf` call to append `nullableSQLString(p.PostalCodePrefix)` and the ward call to append `nullableSQLString(d.PostalCode)` as the last argument.

- [ ] **Step 6: Run all SQL writer tests**

Run: `go test -v ./internal/dataset_writer/dataset_file_writer/ -run "Postgres|Mssql|Oracle"`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/
git commit -m "feat: include postal codes in SQL dataset writers"
```

---

### Task 7: JSON writer

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/dto/json_dto.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/helper/dto_mapper.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/json_file_writer_test.go`

**Interfaces:**
- Consumes: `Province.PostalCodePrefix`, `Ward.PostalCode`
- Produces: `JsonProvinceModel.PostalCodePrefix`, `JsonWardModel.PostalCode` (+ simplified/VN-only variants)

- [ ] **Step 1: Write the failing test**

In `json_file_writer_test.go`, add:

```go
func TestJSONDatasetFileWriter_WriteToFile_PostalCodes(t *testing.T) {
	tmpDir := t.TempDir()

	writer := &JSONDatasetFileWriter{OutputFolderPath: tmpDir}

	provinces := []vn_provinces_tmp_model.Province{
		{
			Code: "01", Name: "Hà Nội", NameEn: "Ha Noi",
			FullName: "Thành phố Hà Nội", FullNameEn: "Ha Noi City",
			CodeName: "ha_noi", AdministrativeUnitId: 1,
			PostalCodePrefix: "10, 11, 12, 13, 14",
			Wards: []*vn_provinces_tmp_model.Ward{
				{
					Code: "00070", Name: "Hoàn Kiếm", NameEn: "Hoan Kiem",
					FullName: "Phường Hoàn Kiếm", FullNameEn: "Hoan Kiem Ward",
					CodeName: "hoan_kiem", ProvinceCode: "01", AdministrativeUnitId: 3,
					PostalCode: "11024",
				},
			},
		},
	}

	err := writer.WriteToFile(nil, nil, provinces, nil)
	assert.NoError(t, err)

	files, _ := os.ReadDir(tmpDir)
	var fullContent []byte
	for _, f := range files {
		if len(f.Name()) >= 5 && f.Name()[:5] == "full_" {
			fullContent, _ = os.ReadFile(filepath.Join(tmpDir, f.Name()))
		}
	}
	assert.Contains(t, string(fullContent), "11024")
	assert.Contains(t, string(fullContent), "10, 11, 12, 13, 14")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./internal/dataset_writer/dataset_file_writer/ -run TestJSONDatasetFileWriter_WriteToFile_PostalCodes`
Expected: FAIL (fields not emitted).

- [ ] **Step 3: Add fields to JSON DTOs**

In `dto/json_dto.go`, add `PostalCodePrefix string` to `JsonProvinceModel`, `JsonProvinceSimplifiedModel`, `JsonProvinceVNSimplifiedModel`; add `PostalCode string` to `JsonWardModel`, `JsonWardSimplifiedModel`, `JsonWardVNSimplifiedModel`.

- [ ] **Step 4: Populate fields in the mappers**

In `helper/dto_mapper.go`, add `PostalCodePrefix: province.PostalCodePrefix` to the three province mappers (`ConvertToJsonProvinceModel`, `ConvertToJsonProvinceSimplifiedModel`, `ConvertToJsonProvinceVNSimplifiedModel`) and `PostalCode: ward.PostalCode` to the three ward mappers (`ConvertToJsonWardModel`, `ConvertToJsonWardSimplifiedModel`, `ConvertToJsonWardVNSimplifiedModel`).

- [ ] **Step 5: Run test to verify it passes**

Run: `go test -v ./internal/dataset_writer/dataset_file_writer/ -run "JSONDatasetFileWriter"`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/dto/json_dto.go internal/dataset_writer/dataset_file_writer/helper/dto_mapper.go internal/dataset_writer/dataset_file_writer/json_file_writer_test.go
git commit -m "feat: include postal codes in JSON dataset writer"
```

---

### Task 8: MongoDB writer

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/dto/mongo_dto.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/helper/dto_mapper.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/mongodb_file_writer_test.go`

**Interfaces:**
- Consumes: `Province.PostalCodePrefix`, `Ward.PostalCode`
- Produces: `MongoProvinceModel.PostalCodePrefix`, `MongoWardModel.PostalCode`

- [ ] **Step 1: Write the failing test**

In `mongodb_file_writer_test.go`, add:

```go
func TestMongoDBDatasetFileWriter_WriteToFile_PostalCodes(t *testing.T) {
	tmpDir := t.TempDir()

	writer := &MongoDBDatasetFileWriter{OutputFolderPath: tmpDir}

	provinces := []vn_provinces_tmp_model.Province{
		{
			Code: "01", Name: "Hà Nội", NameEn: "Ha Noi",
			FullName: "Thành phố Hà Nội", FullNameEn: "Ha Noi City",
			CodeName: "ha_noi", AdministrativeUnitId: 1, PostalCodePrefix: "10, 11, 12, 13, 14",
			Wards: []*vn_provinces_tmp_model.Ward{
				{
					Code: "00070", Name: "Hoàn Kiếm", NameEn: "Hoan Kiem",
					FullName: "Phường Hoàn Kiếm", FullNameEn: "Hoan Kiem Ward",
					CodeName: "hoan_kiem", ProvinceCode: "01", AdministrativeUnitId: 3, PostalCode: "11024",
				},
			},
		},
	}

	err := writer.WriteToFile(nil, nil, provinces, []vn_provinces_tmp_model.Ward{})
	assert.NoError(t, err)

	files, _ := os.ReadDir(tmpDir)
	var mongoContent []byte
	for _, f := range files {
		if len(f.Name()) >= 18 && f.Name()[:18] == "mongo_data_vn_unit" {
			mongoContent, _ = os.ReadFile(filepath.Join(tmpDir, f.Name()))
		}
	}
	assert.Contains(t, string(mongoContent), "11024")
	assert.Contains(t, string(mongoContent), "10, 11, 12, 13, 14")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./internal/dataset_writer/dataset_file_writer/ -run TestMongoDBDatasetFileWriter_WriteToFile_PostalCodes`
Expected: FAIL.

- [ ] **Step 3: Add fields to MongoDB DTOs**

In `dto/mongo_dto.go`, add `PostalCodePrefix string` to `MongoProvinceModel` and `PostalCode string` to `MongoWardModel`.

- [ ] **Step 4: Populate fields in the mappers**

In `helper/dto_mapper.go`, add `PostalCodePrefix: province.PostalCodePrefix` to `ConvertToMongoProvinceModel` and `PostalCode: ward.PostalCode` to `ConvertToMongoWardModel`.

- [ ] **Step 5: Run test to verify it passes**

Run: `go test -v ./internal/dataset_writer/dataset_file_writer/ -run "MongoDBDatasetFileWriter"`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/dto/mongo_dto.go internal/dataset_writer/dataset_file_writer/helper/dto_mapper.go internal/dataset_writer/dataset_file_writer/mongodb_file_writer_test.go
git commit -m "feat: include postal codes in MongoDB dataset writer"
```

---

### Task 9: Redis writer

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/redis_file_writer.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/redis_file_writer_test.go`

**Interfaces:**
- Consumes: `Province.PostalCodePrefix`, `Ward.PostalCode`
- Produces: Redis `HSET province:` includes `postalCodePrefix`, `HSET ward:` includes `postalCode`

- [ ] **Step 1: Write the failing test**

In `redis_file_writer_test.go`, add `PostalCodePrefix: "10, 11, 12, 13, 14"` to the province fixture in `TestRedisDatasetFileWriter_WriteToFile_Provinces` and `PostalCode: "11024"` to the ward fixture in `TestRedisDatasetFileWriter_WriteToFile_Wards`, then add assertions:

```go
	assert.Contains(t, contentStr, `postalCodePrefix "10, 11, 12, 13, 14"`)
	assert.Contains(t, contentStr, `postalCode "11024"`)
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -v ./internal/dataset_writer/dataset_file_writer/ -run TestRedisDatasetFileWriter`
Expected: FAIL.

- [ ] **Step 3: Update Redis templates and generators**

In `redis_file_writer.go`, change the templates:

```go
const hsetProvinceTemplate string = "HSET province:%s code \"%s\" name \"%s\" nameEn \"%s\" fullName \"%s\" fullNameEn \"%s\" codeName \"%s\" postalCodePrefix \"%s\" administrativeUnitId %d \n"
const hsetWardTemplate string = "HSET ward:%s code \"%s\" name \"%s\" nameEn \"%s\" fullName \"%s\" fullNameEn \"%s\" codeName \"%s\" postalCode \"%s\" administrativeUnitId %d districtCode \"%s\" \n"
```

Update `generateProvinceRecord`:

```go
func generateProvinceRecord(p model.Province) string {
	return fmt.Sprintf(hsetProvinceTemplate, p.Code, p.Code, p.Name, p.NameEn, p.FullName, p.FullNameEn, p.CodeName, p.PostalCodePrefix, p.AdministrativeUnitId)
}
```

Update `generateWardRecord`:

```go
func generateWardRecord(w model.Ward) string {
	return fmt.Sprintf(hsetWardTemplate, w.Code, w.Code, w.Name, w.NameEn, w.FullName, w.FullNameEn, w.CodeName, w.PostalCode, w.AdministrativeUnitId, w.ProvinceCode)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -v ./internal/dataset_writer/dataset_file_writer/ -run "RedisDatasetFileWriter"`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/redis_file_writer.go internal/dataset_writer/dataset_file_writer/redis_file_writer_test.go
git commit -m "feat: include postal codes in Redis dataset writer"
```

---

### Task 10: Elasticsearch writer

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/dto/elasticsearch_dto.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/helper/dto_mapper.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer_test.go`

**Interfaces:**
- Consumes: `Province.PostalCodePrefix`, `Ward.PostalCode`
- Produces: `ElasticsearchProvinceDocument.PostalCodePrefix`, `ElasticsearchWardDocument.PostalCode`, mappings updated with `keyword` fields

- [ ] **Step 1: Write the failing test**

In `elasticsearch_file_writer_test.go`, add `PostalCodePrefix: "10, 11, 12, 13, 14"` to the province fixture and `PostalCode: "11024"` to the ward fixture in `TestWriteToFile_NonGIS`, then add assertions after reading `provinces.ndjson`:

```go
	assert.Contains(t, string(data), `"PostalCodePrefix":"10, 11, 12, 13, 14"`)
	assert.Contains(t, string(data), `"PostalCode":"11024"`)
```

Also add a test that the generated mapping contains the new fields:

```go
func TestWriteToFile_MappingContainsPostalFields(t *testing.T) {
	tmpDir := t.TempDir()
	writer := ElasticsearchDatasetFileWriter{OutputFolderPath: tmpDir}
	err := writer.WriteToFile(nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("WriteToFile failed: %v", err)
	}
	mapping, err := os.ReadFile(filepath.Join(tmpDir, "mappings", "provinces.json"))
	if err != nil {
		t.Fatalf("read mapping: %v", err)
	}
	assert.Contains(t, string(mapping), "PostalCode")
	assert.Contains(t, string(mapping), "PostalCodePrefix")
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -v ./internal/dataset_writer/dataset_file_writer/ -run "WriteToFile"`
Expected: FAIL.

- [ ] **Step 3: Add fields to ES DTOs**

In `dto/elasticsearch_dto.go`:
- Add `PostalCodePrefix string \`json:"PostalCodePrefix"\`` to `ElasticsearchProvinceDocument`
- Add `PostalCode string \`json:"PostalCode"\`` to `ElasticsearchWardDocument`
- Add `PostalCode string \`json:"PostalCode"\`` and `PostalCodePrefix string \`json:"PostalCodePrefix"\`` to `ElasticsearchGISProperties`

- [ ] **Step 4: Populate fields in the mappers**

In `helper/dto_mapper.go`:
- In `ConvertToElasticsearchProvinceModel`, add `PostalCodePrefix: province.PostalCodePrefix`
- In `convertToElasticsearchWardDocuments`, add `PostalCode: ward.PostalCode`

In `elasticsearch_file_writer.go`, in `WriteElasticsearchGISDataToFile`:
- In the `provinceProps` literal, add `PostalCodePrefix: province.PostalCodePrefix`
- In the `wardProps` literal, add `PostalCode: ward.PostalCode`

- [ ] **Step 5: Update the ES mappings**

In `elasticsearch_file_writer.go`:
- In `writeProvincesMapping`, add to the top-level `properties` map: `"PostalCodePrefix": map[string]string{"type": "keyword"},` and to the `Wards` nested `properties`: `"PostalCode": map[string]string{"type": "keyword"},`
- In `writeProvincesGISMapping`, add `"PostalCodePrefix": map[string]string{"type": "keyword"},` to the top-level `properties`, `"PostalCode": map[string]string{"type": "keyword"},` to the `Wards` nested `properties`, and both `"PostalCode"`/`"PostalCodePrefix"` to the `GIS.Properties` `properties` map.

- [ ] **Step 6: Run tests to verify they pass**

Run: `go test -v ./internal/dataset_writer/dataset_file_writer/ -run "WriteToFile|Elasticsearch"`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/dto/elasticsearch_dto.go internal/dataset_writer/dataset_file_writer/helper/dto_mapper.go internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer.go internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer_test.go
git commit -m "feat: include postal codes in Elasticsearch dataset writer"
```

---

### Task 11: GIS outputs (GeoJSON, MongoDB GIS)

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/dto/geojson_dto.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/geojson_file_writer.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/dto/mongo_gis_dto.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/helper/mongo_gis_mapper.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/json_file_writer_test.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/mongodb_gis_file_writer_test.go`

**Interfaces:**
- Consumes: `VNProvince.PostalCodePrefix`, `VNWard.PostalCode` (loaded via `SapNhapSiteGeoUnit` relations)
- Produces: GeoJSON `properties.postalCodePrefix`/`properties.postalCode`; Mongo/ES GIS `Properties.PostalCode`/`PostalCodePrefix`

- [ ] **Step 1: Write the failing tests**

In `json_file_writer_test.go`, extend `TestJSONDatasetFileWriter_WriteGISGeoJSONToFile`:
- Add `PostalCodePrefix: "10, 11, 12, 13, 14"` to the province fixture's `VNProvince` literal (after `CodeName: "ha_noi",`)
- Add `PostalCode: "11024"` to the ward fixture's `VNWard` literal (after `CodeName: "ba_dinh",`)
- Add an assertion after the existing `assert.Equal(t, 3359.84, properties["areaKm2"])`:

```go
	assert.Equal(t, "10, 11, 12, 13, 14", properties["postalCodePrefix"])
```

- Add assertions after the existing `wardFeature` unmarshal block (after `assert.Equal(t, wardJSON["bbox"], wardFeature["bbox"])`):

```go
	wardProperties := wardFeature["properties"].(map[string]any)
	assert.Equal(t, "11024", wardProperties["postalCode"])
```

In `mongodb_gis_file_writer_test.go`, extend `TestWriteMongoGISDataToFile`:
- Add `PostalCodePrefix: "10, 11, 12, 13, 14"` to the `VNProvince` literal (after `CodeName: "ha_noi",`)
- Add `PostalCode: "11024"` to the `VNWard` literal (after `CodeName: "ba_dinh",`)
- Add assertions after the `provinceDocs[0].GIS.Center.Type` check:

```go
	if provinceDocs[0].GIS.Properties == nil {
		t.Fatal("expected province GIS Properties to be populated")
	}
	if provinceDocs[0].GIS.Properties.PostalCodePrefix != "10, 11, 12, 13, 14" {
		t.Errorf("expected PostalCodePrefix '10, 11, 12, 13, 14', got %q", provinceDocs[0].GIS.Properties.PostalCodePrefix)
	}
```

- Add assertions after the `wardDocs[0].GIS == nil` check:

```go
	if wardDocs[0].GIS.Properties == nil {
		t.Fatal("expected ward GIS Properties to be populated")
	}
	if wardDocs[0].GIS.Properties.PostalCode != "11024" {
		t.Errorf("expected PostalCode '11024', got %q", wardDocs[0].GIS.Properties.PostalCode)
	}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -v ./internal/dataset_writer/dataset_file_writer/ -run "GIS|GeoJSON"`
Expected: FAIL.

- [ ] **Step 3: Add fields to GeoJSON DTO**

In `dto/geojson_dto.go`, add to `GeoJSONFeatureProperties`:

```go
	PostalCode       string  `json:"postalCode"`
	PostalCodePrefix string  `json:"postalCodePrefix"`
```

- [ ] **Step 4: Populate GeoJSON properties**

In `geojson_file_writer.go`:
- In `writeProvinceGeoJSON`, add `PostalCodePrefix: province.VNProvince.PostalCodePrefix` to the `GeoJSONFeatureProperties` literal
- In `writeWardGeoJSON`, add `PostalCode: ward.VNWard.PostalCode` to the `GeoJSONFeatureProperties` literal

- [ ] **Step 5: Add fields to MongoDB GIS DTO**

In `dto/mongo_gis_dto.go`, add to `MongoGISProperties`:

```go
	PostalCode       string `json:"PostalCode"`
	PostalCodePrefix string `json:"PostalCodePrefix"`
```

- [ ] **Step 6: Populate MongoDB GIS mapper**

In `helper/mongo_gis_mapper.go`:
- In `ConvertToMongoGISProvinceDocuments`, add `PostalCodePrefix: province.PostalCodePrefix` to the `provinceProps` literal
- In `ConvertToMongoGISWardDocuments`, add `PostalCode: ward.PostalCode` to the `wardProps` literal

- [ ] **Step 7: Run tests to verify they pass**

Run: `go test -v ./internal/dataset_writer/dataset_file_writer/ -run "GIS|GeoJSON|MongoGIS"`
Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/dto/geojson_dto.go internal/dataset_writer/dataset_file_writer/geojson_file_writer.go internal/dataset_writer/dataset_file_writer/dto/mongo_gis_dto.go internal/dataset_writer/dataset_file_writer/helper/mongo_gis_mapper.go internal/dataset_writer/dataset_file_writer/json_file_writer_test.go internal/dataset_writer/dataset_file_writer/mongodb_gis_file_writer_test.go
git commit -m "feat: include postal codes in GIS outputs"
```

---

### Task 12: Published artifacts, docs, and final verification

**Files:**
- Modify: `AGENTS.md`
- Modify: `dataset-generation-scripts/README.md`
- Modify: `elasticsearch/mappings/provinces.json`, `elasticsearch/mappings/provinces-gis.json` (committed copies)
- Regenerated: top-level `postgresql/`, `mysql/`, `sqlserver/`, `oracle/`, `json/`, `mongodb/`, `redis/`, `elasticsearch/` outputs

**Interfaces:**
- Consumes: all prior tasks; the generated `output/` directory from a full `go run main.go`
- Produces: updated published datasets + docs

- [ ] **Step 1: Re-run full generation**

Run: `go run main.go` (from `dataset-generation-scripts/`).
Expected: completes with postal code import (3321 wards matched) and all writers succeed.

- [ ] **Step 2: Update the committed Elasticsearch mapping copies**

Edit `elasticsearch/mappings/provinces.json` and `elasticsearch/mappings/provinces-gis.json` to mirror the Go-generated mappings (add `PostalCode`/`PostalCodePrefix` keyword fields at the same locations as in Task 10 Step 5). Alternatively, copy the freshly generated files from `dataset-generation-scripts/output/elasticsearch/mappings/`.

- [ ] **Step 3: Update the published dataset folders**

Copy the freshly generated files from `dataset-generation-scripts/output/` into the top-level published folders, following the existing release flow:
- `output/postgresql/postgres_ImportData_vn_units.sql` → `postgresql/`
- `output/mysql/mysql_ImportData_vn_units.sql` → `mysql/`
- `output/sqlserver/mssql_ImportData_vn_units.sql` → `sqlserver/`
- `output/oracle/oracle_ImportData_vn_units.sql` → `oracle/`
- `output/json/*.json` → `json/` (full + simplified + vn_only, dropping old timestamped files)
- `output/mongodb/*.json` → `mongodb/`
- `output/redis/*.redis` → `redis/redis_vn_provinces_dataset.redis`
- `output/elasticsearch/*` → `elasticsearch/` (ndjson + mappings)
- GIS: copy `output/json/geojson/` → `json/geojson/`, `output/mongodb/gis/` → `mongodb/gis/`, `output/postgresql/gis/` + `output/mysql/gis/` + `output/sqlserver/gis/` → respective `gis/` folders

- [ ] **Step 4: Verify the published SQL files contain postal columns**

```bash
grep -c "postal_code" postgresql/postgres_ImportData_vn_units.sql mysql/mysql_ImportData_vn_units.sql sqlserver/mssql_ImportData_vn_units.sql oracle/oracle_ImportData_vn_units.sql
```

Expected: non-zero for each (provinces + wards inserts both carry the columns).

- [ ] **Step 5: Update AGENTS.md**

In the **Key Columns** section for `provinces_tmp`, add: `postal_code_prefix` (comma-separated 2-digit postal prefixes, e.g. `'10, 11, 12, 13, 14'`). For `wards_tmp`, add: `postal_code` (5-digit national postal code). Also update the **Current generation flow** section in AGENTS.md (and `dataset-generation-scripts/CLAUDE.md` if it documents the flow) to include `postal_code.ImportPostalCodes()` after the DVHCVN dump.

- [ ] **Step 6: Update dataset-generation-scripts README**

In `dataset-generation-scripts/README.md`, add a line to the output structure section noting the generated SQL/JSON/Mongo/Redis/ES outputs now include `postal_code` / `postal_code_prefix`.

- [ ] **Step 7: Run the full test suite**

Run: `go test -v ./...`
Expected: all tests pass (Docker must be running).

- [ ] **Step 8: Final DB verification**

```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "SELECT (SELECT COUNT(*) FROM provinces_tmp WHERE postal_code_prefix IS NULL OR postal_code_prefix = '') AS missing_provinces, (SELECT COUNT(*) FROM wards_tmp WHERE postal_code IS NULL OR postal_code = '') AS missing_wards;"
```

Expected: `0 | 0`.

- [ ] **Step 9: Commit**

```bash
git add AGENTS.md dataset-generation-scripts/README.md elasticsearch/ postgresql/ mysql/ sqlserver/ oracle/ json/ mongodb/ redis/
git commit -m "docs: publish postal code dataset and update documentation"
```

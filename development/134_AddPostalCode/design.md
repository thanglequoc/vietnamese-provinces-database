# Design: Add Postal Codes to Vietnamese Provinces Database

> **Status**: Approved (design phase) — waiting for implementation plan
> **Date**: 2026-08-08
> **Source decree**: Quyết định số 2334/QĐ-BKHCN ngày 24/8/2025 (Bộ Khoa học và Công nghệ)

## 1. Objective

Integrate the national postal codes from Quyết định 2334/QĐ-BKHCN into the
existing Vietnamese provinces database so that **every dataset output carries
postal codes by default**.

Two levels of postal data are captured:

- **Ward-level**: 5-digit postal code per ward (3,321 wards)
- **Province-level**: 2-digit prefix string per province, e.g. `'90, 91, 92'`

Storage decision (confirmed): postal codes are **columns on the existing
wards/provinces tables** — no separate `postal_codes` table. Both columns are
**nullable** for future-proofing against decree revisions that introduce wards
without assigned postal codes.

## 2. Data source & seed files

**Source**: `development/134_AddPostalCode/postal_codes.md` — parsed from the
decree by `parse_postal_codes_docx.py`. Already validated at **100% match**
(0 unmatched, 0 multi-match, 0 STT problems) against `wards_tmp`.

**New seed files** (committed under `dataset-generation-scripts/resources/postal/`):

| File | Rows | Shape |
|------|------|-------|
| `province_postal_code_prefixes.json` | 34 | `[{"code": "01", "postal_code_prefix": "10, 11, 12, 13, 14"}]` |
| `ward_postal_codes.json` | 3,321 | `[{"province_code": "01", "name": "Hoàn Kiếm", "postal_code": "11024"}]` |

- Province entries keyed by **province code** (stable, matches `provinces_tmp.code`).
- Ward entries keyed by **(province_code, ward `name`)** — `name` is the
  stripped unit name (no `X.`/`P.` prefix), matching `wards_tmp.name`.
- Seed files are generated from `postal_codes.md` by the Python tooling in
  `development/134_AddPostalCode/` so they remain reproducible.

## 3. Schema & model changes

### Temporary database (`dataset-generation-scripts/resources/db_table_init.sql`)

- `provinces_tmp`: add nullable `postal_code_prefix varchar(255) NULL`
- `wards_tmp`: add nullable `postal_code varchar(20) NULL`

### Published `*_CreateTables_vn_units.sql`

- `provinces` tables (postgresql, mysql, sqlserver, oracle): add `postal_code_prefix`
- `wards` tables: add `postal_code`
- Use each engine's string type (`varchar` / `nvarchar` per engine convention).
- Both nullable.

### Bun models (`internal/vn_provinces_tmp/model/vn_provinces_tmp_model.go`)

- `Province`: add `PostalCodePrefix string \`bun:"postal_code_prefix"\``
- `Ward`: add `PostalCode string \`bun:"postal_code"\``

### Impact on existing flows

- DVHCVN dumper inserts without these fields → rows get `NULL` (correct; filled
  later by the postal import step).
- `sapnhap_geojson_objects` backfill matches by name only — unaffected.
- `fresh_cleanup.sql` drops whole tables — unchanged.

## 4. Postal code import step

**New package**: `internal/postal_code/`

```
internal/postal_code/
├── postal_code.go            # Entry point: ImportPostalCodes()
├── service/                  # Parse seed JSON, resolve matches
│   └── postal_code_service.go
└── repository/               # DB operations
    └── postal_code_repository.go
```

**Pipeline position** in `main.go`:

```
BootstrapTemporaryDatasetStructure()
BeginDumpingDataWithDvhcvnDirectSource()
postal_code.ImportPostalCodes()      // NEW — after dump, before writers
ReadAndGenerateSQLDatasets()
... GIS phase (unchanged)
```

**Matching logic**:

1. Read both seed JSON files from `resources/postal/`.
2. **Provinces**: `UPDATE provinces_tmp SET postal_code_prefix = ? WHERE code = ?`
   (direct code lookup).
3. **Wards**: for each seed row, `UPDATE wards_tmp SET postal_code = ?
   WHERE province_code = ? AND name = ?` — normalize both sides with
   `common/viet` (NFC + tone-mark normalization). Ward names are unique within
   a province, so each match is deterministic.
4. **Verification gate**: after import, assert all 3,321 wards and all 34
   provinces were matched. Any unmatched row fails loudly (mirrors the
   project's 100% GIS ID match standard) — never silently ship incomplete data.

## 5. Dataset writer changes

All writers keep their current `WriteToFile` signatures — no new params. The
new model fields flow through automatically.

| Writer | Change |
|--------|--------|
| **SQL** (postgres/mysql, mssql, oracle) | Province INSERT: add `postal_code_prefix`; Ward INSERT: add `postal_code` |
| **JSON** (`json_dto.go`, `dto_mapper.go`, `json_file_writer.go`) | `JsonProvinceModel.PostalCodePrefix`, `JsonWardModel.PostalCode`; add to simplified + VN-only variants |
| **MongoDB** (`mongo_dto.go`, `dto_mapper.go`) | `MongoProvinceModel.PostalCodePrefix`, `MongoWardModel.PostalCode` |
| **Redis** (`redis_file_writer.go`) | `HSET province:` gains `postalCodePrefix`; `HSET ward:` gains `postalCode` |
| **Elasticsearch** (`elasticsearch_dto.go`, `dto_mapper.go`, `elasticsearch_file_writer.go`) | `ElasticsearchProvinceDocument.PostalCodePrefix`, `ElasticsearchWardDocument.PostalCode`; update mappings (`provinces.json`, `provinces-gis.json`) as `keyword`; GIS Properties DTOs gain fields |
| **GeoJSON** (`geojson_dto.go`, `geojson_file_writer.go`) | `GeoJSONFeatureProperties` gains `PostalCode` (wards) / `PostalCodePrefix` (provinces) |
| **MongoDB GIS** (`mongo_gis_dto.go`, `mongo_gis_mapper.go`) | `MongoGISProperties` gains fields; mapper reads `VNProvince.PostalCodePrefix` / `VNWard.PostalCode` |
| **Elasticsearch GIS** | `ElasticsearchGISProperties` gains fields |

**Tests**: update each writer's `*_test.go` to assert new fields; add seed
parsing + matching unit tests in `internal/postal_code/`.

## 6. Published artifacts & docs

- Copy freshly generated `output/` files into published folders
  (`postgresql/`, `mysql/`, `sqlserver/`, `oracle/`, `json/`, `mongodb/`,
  `redis/`, `elasticsearch/`) — same flow as existing GIS releases.
- Update curated `*_CreateTables_vn_units.sql` (see Section 3).
- **Docs**: update `AGENTS.md` Key Columns tables (`postal_code` /
  `postal_code_prefix`), and `dataset-generation-scripts/README.md` output
  structure/flow.

## 7. Verification after generation

1. `SELECT COUNT(*) FROM wards_tmp WHERE postal_code IS NULL` → 0
2. `SELECT COUNT(*) FROM provinces_tmp WHERE postal_code_prefix IS NULL` → 0
3. Spot-check known values: P. Hoàn Kiếm = `11024`, X. An Phú = `90456`
4. Confirm published SQL files reference the new columns and pass engine
   syntax checks.

## 8. Out of scope (YAGNI)

- No standalone `postal_codes` table.
- No district-level postal codes.
- No changes to `sapnhap_geojson_objects`.
- No new indexes.

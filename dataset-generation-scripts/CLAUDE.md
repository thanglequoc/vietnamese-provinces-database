# Vietnamese Provinces Database - Go Service Instructions

**See [../AGENTS.md](../AGENTS.md) for root-level project context**, architecture overview, and multi-database export details.

This file provides subsystem-specific guidance for the Go services in `dataset-generation-scripts/`.

---

## When to Use Database Queries

See [../AGENTS.md#database-query-skill](../AGENTS.md) for the universal db-query skill — auto-trigger rules, connection details, schema/table reference, and query patterns. Apply these rules automatically; do not wait for explicit invocation.

## Database Access

See [../AGENTS.md#database-query-skill](../AGENTS.md). Quick reminder: the database is accessible via:
```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "QUERY"
```
Container: `vn_provinces_postgres_container`, Database: `vn_provinces_tmp`, Port: `15432`

## Data Migration History

### SAPNhap API Deprecation (March 2026)

The upstream data site `sapnhap.bando.com.vn` deprecated their API endpoints:
- `/pcotinh` (provinces list)
- `/ptracuu` (wards lookup)

**Solution:** Migrated to local JSON and GeoJSON files:
- `./resources/gis/geojson_11Mar2026/` - GeoJSON geometry files (used for An Giang manual patch)

> **July 2026 Update:** The `bando_gisserver/` local JSON metadata files (provinces.json, wards.json) were removed.
> The code path that loaded them was dead — never invoked from `main.go`.
> The live GIS pipeline uses HTTP API calls to `sapnhap.bando.com.vn` directly.
> See: `development/cleanup_old_reference/remove_bando_gisserver_references.md`

**Implementation:**
- Replaced API calls with file-based loading
- Implemented GIS server ID matching for data integrity
- All 3,355 records verified with 100% GIS ID match rate
- Sequential ID generation for wards to prevent duplication

**See:** `development/adapt_the_removal_of_sapnhap_api.md` for full documentation

## Project Structure

```
dataset-generation-scripts/
├── internal/
│   ├── common/
│   │   └── viet/              # Vietnamese text handling (tone marks, normalization)
│   ├── dvhcvn_data_downloader/ # DVHCVN SOAP data ingestion (parser, model, downloader)
│   ├── sapnhap_bando/         # Geographic data service
│   │   ├── fetcher/           # Data fetching from HTTP API & local files
│   │   ├── service/           # Business logic (GIS fetch, backfill, metadata)
│   │   ├── repository/        # Database operations (sapnhap, gis, geojson_objects)
│   │   ├── model/             # Database models (sapnhap, geojson)
│   │   ├── dto/               # Data transfer objects (geojson_file, gis_server, sapnhap_api)
│   │   └── util/              # Name normalization utilities
│   ├── dumper/                # Admin data ingestion to DB
│   │   ├── config/            # Constant mappings
│   │   ├── helper/            # Helper functions
│   │   ├── model/             # Domain models
│   │   ├── repository/        # Seed data repository
│   │   └── service/           # DVHCVN SOAP seed dumper, corrector, manual seed dumper
│   ├── dataset_writer/        # Multi-format dataset generation
│   │   └── dataset_file_writer/ # Per-format writers (postgres/mysql, mssql, oracle, json, mongodb, redis, geojson)
│   │       ├── dto/           # Output DTOs
│   │       └── helper/        # DTO mappers
│   ├── vn_provinces_tmp/      # Core VN provinces data layer (model + repository)
│   ├── gis/                   # GIS models
│   ├── testutil/              # Test fixtures/helpers
│   └── database/              # Postgres connection + bootstrap/SQL execution
├── resources/
│   ├── db_table_init.sql      # Core table schema
│   ├── db_region_administrative_unit.sql
│   ├── fresh_cleanup.sql
│   ├── gis/
│   │   ├── geojson_11Mar2026/ # GeoJSON geometry files (from deprecated API)
│   │   ├── sapnhapbando_geojson/ # Auxiliary GIS GeoJSON resources (3,355 files)
│   │   ├── sapnhap_bando_tables.sql
│   │   ├── sapnhapbando_init_geo_json_objects_tbl.sql
│   │   └── sapnhapbando_geo_objects.sql
│   ├── manual_seeds/          # Manual fallback seed data
│   └── rules/                 # Vietnamese text convention rules
├── docker/
│   └── docker-compose.yaml    # Postgres/PostGIS container (port 15432)
├── memory/                    # AI agent memory & feedback
├── output/                    # Generated artifacts (gitignored)
└── main.go                    # Entry point
```

## Development Workflow

1. **Database Operations:** Use the `/db-query` skill to run SQL queries
2. **Running Scripts:** `go run .` from dataset-generation-scripts directory
3. **Database Migration:** Scripts handle schema and data seeding
4. **GIS Data:** All geometry stored in WKT format (PostGIS)

## Feature Planning (AI-Assisted Development)

**MANDATORY:** For any new feature, save a detailed plan in the `../development/` folder following the template:
- **Filename:** Short descriptive summary (e.g., `backfill-province-codes.md`)
- **Content:** Objectives, affected components, step-by-step logic, edge cases, assumptions
- **Purpose:** Document all AI-assisted work for future reference and team collaboration

**See:** [../AGENTS.md#development-workflow](../AGENTS.md#development-workflow) for full requirements and examples.

### Recent Feature Examples
- `adapt_the_removal_of_sapnhap_api.md` — Migration from deprecated API to local file-based data (March 2026)

---

## Internal Package Guide

| Package | Purpose | Key Files |
|---------|---------|-----------|
| `common/viet/` | Vietnamese text processing | Tone mark removal, Y-normalization, Unicode handling |
| `dvhcvn_data_downloader/` | Direct DVHCVN admin data ingestion | SOAP parser (`dvhcvn_parser.go`), data downloader, models |
| `sapnhap_bando/` | Geographic data service (HTTP API) | `fetcher/` loads data, `service/` orchestrates GIS fetch + backfill, `repository/` queries DB, `util/` normalizes names |
| `dumper/` | Reads admin data, persists to DB | `service/` has DVHCVN SOAP dumper + corrector, manual seed dumper; `config/` has constant mappings; `repository/` for seed data |
| `dataset_writer/` | Generates SQL/JSON/NoSQL exports | `dataset_file_writer/` has per-format writers (postgres/mysql, mssql, oracle, json, mongodb, redis, geojson) + `dto/` + `helper/` |
| `vn_provinces_tmp/` | Core VN administrative data layer | `model/` has Bun ORM models (Province, Ward, AdministrativeUnit, AdministrativeRegion); `repository/` for queries |
| `gis/` | GIS shared models | Geometry data structures |
| `database/` | Postgres connection + DB bootstrap | `postgres_connector.go` (.env-based DSN), `vn_province_db_service.go` (schema bootstrap, SQL script execution) |
| `testutil/` | Test helpers | Database seeding, fixtures |

---

## Key Modules & Dependencies

| Module | Purpose | Usage |
|--------|---------|-------|
| `uptrace/bun` + `pgdialect` + `pgdriver` | PostgreSQL ORM | CRUD operations, query building |
| `joho/godotenv` | `.env` configuration | Load DB credentials |
| `stretchr/testify` | Testing assertions | Unit & integration tests |
| `golang.org/x/text` | Unicode normalization | Province/ward name cleaning |
| `dlclark/regexp2` | Advanced regex | Pattern matching |

---

## Important Configuration

| File | Purpose |
|------|---------|
| `.env` | Database credentials — 5 env vars: `POSTGRES_DB_USERNAME`, `POSTGRES_DB_PSWD`, `POSTGRES_DB_HOST`, `POSTGRES_DB_PORT`, `POSTGRES_TMP_DB_NAME`. The DSN is constructed programmatically in `internal/database/postgres_connector.go` |
| `.env.example` | Template for `.env` (copy to `.env` and fill in credentials) |
| `docker/docker-compose.yaml` | Docker Postgres/PostGIS service (port 15432→5432) |
| `go.mod` | Go module definition & dependencies (Go 1.24.0) |
| `resources/gis/geojson_11Mar2026/` | GeoJSON geometry files from deprecated API |
| `resources/gis/sapnhapbando_geojson/` | 3,355 auxiliary GIS GeoJSON resource files |
| `main.go` | Entry point for dataset generation |

## GIS Server ID Matching

Critical for data integrity: GIS server IDs from JSON metadata must match GeoJSON feature IDs:
- Provinces: Format `tinh34.7` (from `tinh` prefix)
- Wards: Format `xa3321.3285` (from `xa` prefix)

Verification query:
```sql
SELECT
  pg.gis_server_id,
  pg.ten,
  pg.sapnhap_province_matinh
FROM sapnhap_provinces_gis pg
JOIN sapnhap_provinces sp ON pg.sapnhap_province_matinh = sp.mahc;
```

## Common Tasks

### Check Data Completeness
```sql
SELECT
  (SELECT COUNT(*) FROM provinces_tmp) as provinces,
  (SELECT COUNT(*) FROM wards_tmp) as wards,
  (SELECT COUNT(*) FROM sapnhap_provinces) as sapnhap_provinces,
  (SELECT COUNT(*) FROM sapnhap_wards) as sapnhap_wards,
  (SELECT COUNT(*) FROM sapnhap_provinces_gis) as provinces_gis,
  (SELECT COUNT(*) FROM sapnhap_wards_gis) as wards_gis;
```

### Verify GIS Geometry
```sql
SELECT
  COUNT(*) as total,
  COUNT(bbox_wkt) as with_bbox,
  COUNT(geom_wkt) as with_geom
FROM sapnhap_wards_gis;
```

### Find Orphaned Records
```sql
-- Wards without VN ward code
SELECT COUNT(*) FROM sapnhap_wards WHERE vn_ward_code IS NULL;

-- Provinces without VN province code
SELECT COUNT(*) FROM sapnhap_provinces WHERE vn_province_code IS NULL;
```

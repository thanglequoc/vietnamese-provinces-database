# AI Agent Instructions — Vietnamese Provinces Database

A comprehensive database of Vietnamese administrative units (34 provinces, districts, wards) maintained through automated Go scripts that generate SQL, JSON, and GIS data for multiple database engines.

---

## Quick Start for AI Agents

### Building & Running

```bash
cd dataset-generation-scripts

# Run the generation scripts
go run main.go

# Start Postgres/PostGIS for integration tests
docker compose -f docker/docker-compose.yaml up -d

# Run tests
go test -v ./...

# Check output in: ./output/
```

**Testing note**: Some tests connect to the temporary Postgres/PostGIS database on `localhost:15432`. If Docker is not running, `go test -v ./...` will fail in integration-style packages.

### Database Connection (Docker)

```bash
# For queries:
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "QUERY"

# Key tables:
# - provinces_tmp (34 records)
# - wards_tmp (3,321 records)  
# - sapnhap_provinces, sapnhap_wards (with geometry)
# - sapnhap_geojson_objects (3,355 records)
```

**Note**: When working with database queries, proactively use them without waiting for explicit `/db-query` invocation if the task involves data verification, counts, searching, or GIS data.

---

## Project Structure

```
vietnamese-provinces-database/
├── AGENTS.md                          # This file
├── README.md / README_vi.md           # Dataset usage documentation
├── dataset-generation-scripts/        # Core Go automation service
│   ├── CLAUDE.md                      # Detailed subsystem context ← START HERE for code work
│   ├── main.go                        # Entry point
│   ├── go.mod / go.sum               # Go dependencies (Bun, postgres driver)
│   ├── .env                          # Database credentials
│   ├── docker/
│   │   └── docker-compose.yaml       # Local Postgres/PostGIS container
│   ├── internal/
│   │   ├── common/                   # Shared helpers
│   │   │   └── viet/                 # Vietnamese text handling (tone marks, normalization)
│   │   ├── dvhcvn_data_downloader/   # Direct DVHCVN administrative data ingestion (SOAP parser)
│   │   ├── sapnhap_bando/           # Geographic data service (HTTP API to sapnhap.bando.com.vn)
│   │   │   ├── fetcher/             # Loads JSON metadata & GeoJSON files
│   │   │   ├── service/             # Business logic (GIS fetch, backfill, metadata)
│   │   │   ├── repository/          # Database operations (sapnhap, gis, geojson_objects)
│   │   │   ├── model/               # Domain models (sapnhap, geojson)
│   │   │   ├── dto/                 # Data transfer objects (geojson_file, gis_server, sapnhap_api)
│   │   │   └── util/                # Name normalization utilities
│   │   ├── dumper/                  # Reads admin data, persists to DB
│   │   │   ├── config/              # Constant mappings
│   │   │   ├── helper/              # Dumper helper functions
│   │   │   ├── model/               # Dumper domain models
│   │   │   ├── repository/          # Seed data repository
│   │   │   └── service/             # DVHCVN SOAP seed dumper, corrector, manual seed dumper
│   │   ├── dataset_writer/          # Generates SQL/JSON/NoSQL output
│   │   │   └── dataset_file_writer/ # Per-format file writers (postgres/mysql, mssql, oracle, json, mongodb, redis, geojson)
│   │   │       ├── dto/             # Output DTOs (geojson, json, mongo)
│   │   │       └── helper/          # DTO mappers
│   │   ├── vn_provinces_tmp/        # Core VN provinces data layer (provinces_tmp, wards_tmp tables)
│   │   │   ├── model/               # Bun ORM models (Province, Ward, AdministrativeUnit, AdministrativeRegion)
│   │   │   └── repository/          # Repository queries
│   │   ├── gis/                     # GIS models and shared GIS logic
│   │   ├── testutil/                # Test fixtures/helpers
│   │   └── database/                # Postgres connection pool + bootstrap/SQL script execution
│   ├── resources/
│   │   ├── db_table_init.sql         # Core table schema (provinces_tmp, wards_tmp, etc.)
│   │   ├── db_region_administrative_unit.sql  # Region & administrative unit seed data
│   │   ├── fresh_cleanup.sql         # DB cleanup script (run before each generation)
│   │   ├── gis/
│   │   │   ├── geojson_11Mar2026/    # ← GeoJSON geometry (from deprecated API)
│   │   │   ├── sapnhapbando_geojson/ # Auxiliary GIS GeoJSON resources (3,355 files)
│   │   │   ├── sapnhap_bando_tables.sql          # GIS table schema
│   │   │   ├── sapnhapbando_init_geo_json_objects_tbl.sql  # Geo objects table init
│   │   │   └── sapnhapbando_geo_objects.sql       # Geo objects seed data
│   │   ├── manual_seeds/             # Manual fallback seed data (provinces_seed.sql + wards/)
│   │   └── rules/                    # Vietnamese text convention rules
│   │       └── vn_tone_mark_convention.md
│   ├── bando_co_dvch.sql             # Raw SAPNhap administrative unit SQL dump (497KB)
│   ├── sapnhap-bando-crawler/        # Historical/auxiliary crawler tooling (Python scripts)
│   ├── memory/
│   │   ├── MEMORY.md                # Memory index
│   │   └── feedback_*.md            # User preferences & learnings
│   ├── output/                       # Generated artifacts (gitignored, staging area)
│   └── tmp/                          # Temporary working directory
├── development/                       # Feature documentation & planning artifacts
│   ├── adapt_the_removal_of_sapnhap_api.md  # Context: API → file-based migration
│   ├── cleanup_old_reference/         # Completed cleanup plans (e.g., remove_bando_gisserver_references.md)
│   └── include_geojson_export/        # GeoJSON export feature planning
├── docs/
│   └── gis/                          # User-facing GIS documentation (gis_readme.md, gis_readme_vi.md, gis_example_query.md)
├── json/, mysql/, postgresql/, oracle/, sqlserver/, mongodb/, redis/
│   └── Generated dataset exports in various formats
└── .github/workflows/                # CI/CD pipelines (test-go.yml runs Go tests with PostGIS)
```

---

## Key Conventions

### Code Style
- **Language**: Go 1.24.0
- **ORM**: Bun with PostgreSQL dialect
- **Naming**: Database=`snake_case`, Go structs=`PascalCase`, JSON=`snake_case`
- **Error Handling**: Always wrap with context using `fmt.Errorf`

### Development Workflow

1. **Feature Planning**: Save detailed plans in `development/` folder (mandatory for AI-assisted work)
   - Example filename: `backfill-province-codes-from-tmp-tables.md`
   - Include: objectives, affected components, step-by-step logic, edge cases, assumptions

2. **Database Operations**: Use `docker exec` or `/db-query` skill for:
   - Data counts, totals, statistics
   - Data verification or integrity checks
   - Finding/searching specific records
   - GIS or geometry queries
   - Schema/table information

3. **Testing**: Use standard Go test patterns with Testify assertions

4. **Git**: Project uses Git LFS for large GIS files
   ```bash
   git lfs install
   git lfs pull
   ```

### Important Relationships

Current generation flow:
- **Administrative source of truth**: `main.go` defaults to `USE_DIRECT_DVHCVN_SOURCE = true`, so the dumper normally ingests administrative-unit data from the direct DVHCVN source before writing exports.
- **Fallback/manual path**: Manual seed data lives under `dataset-generation-scripts/resources/manual_seeds/` and is used only when the direct source path is disabled.
- **GIS enrichment**: GIS metadata and geometries are joined after the main administrative dump succeeds.

Geographic data migration context (March 2026):
- **Before**: GIS metadata fetched from SAPNhap API (`/pcotinh`, `/ptracuu`)
- **After (March 2026)**: GIS metadata and geometry load from local files:
  - `./resources/gis/geojson_11Mar2026/*.geojson`
- **After (July 2026)**: `bando_gisserver/` local JSON files removed (dead code cleanup). Live GIS uses HTTP API (`pread_json`, `p.co_dvhc_id`) + An Giang manual patch.
- **Key IDs**: `mahc` (province)→`sapnhap_province_matinh` (GIS), `maxa` (ward)→`sapnhap_ward_maxa` (GIS)
- **Status**: All 3,355 records verified with 100% GIS ID match rate

---

## Decision Reference

### When to Query vs When to Generate

| Task | Approach | Tools |
|------|----------|-------|
| Check province/ward counts | Query | `docker exec psql` or `/db-query` |
| Verify data integrity (duplicates, orphans) | Query | SQL verification scripts |
| Generate new SQL dumps | Execute script | `go run main.go` |
| Migrate new government decree | Query → Plan → Execute | Read decree, find affected records, generate patch |

### Government Decree Workflow

Vietnamese government issues administrative change decrees (e.g., `30/2026/QH16`, `19/2025/QĐ-TTg`). 

**When new decree arrives:**
1. Check `patch/` directory for similar historical decrees
2. Read decree document for specific changes (ward merges, promotions, reclassifications)
3. Query `provinces_tmp` and `wards_tmp` to find affected records
4. Document changes in `development/` as a feature plan
5. Update dumper logic if systematic changes required
6. Generate and validate patches for all database formats

**See**: `development/adapt_the_removal_of_sapnhap_api.md` for a recent complex migration example.

---

## Multi-Database Output

The project generates compatible SQL/data for:

- **PostgreSQL** (`postgresql/`) — Primary data source, with PostGIS
- **MySQL / MariaDB** (`mysql/`)
- **Microsoft SQL Server** (`sqlserver/`)
- **Oracle** (`oracle/`)
- **MongoDB** (`mongodb/`)
- **Redis** (`redis/`)

Output locations:
- `dataset-generation-scripts/output/` is the generator's immediate output/staging area when `go run main.go` is executed.
- Top-level folders such as `postgresql/`, `mysql/`, `sqlserver/`, `oracle/`, `mongodb/`, `redis/`, and `json/` contain the repository's published/exported dataset artifacts.

Typical contents in `dataset-generation-scripts/output/` include:
- timestamped import/export artifacts such as `postgresql_mysql_generated_ImportData_vn_units_*.sql`
- JSON, MongoDB, and Redis exports in format-specific subdirectories
- GIS import scripts in `output/gis/`

---

## Testing & Validation

- Unit tests in each `internal/*/` package (run with `go test -v ./...`)
- Historical patch verification in `patch/` directory
- GIS server ID matching (100% validation before release)

### CI/CD Pipeline (`.github/workflows/test-go.yml`)

The CI pipeline runs on pull requests to `main`/`master` and on manual dispatch:

1. **Service setup**: Spins up `postgis/postgis:15-3.3` as a service container on port `15432`
2. **Schema initialization**: Runs SQL scripts in order:
   - `resources/db_table_init.sql` — core tables
   - `resources/db_region_administrative_unit.sql` — region/unit seed data
   - `resources/gis/sapnhap_bando_tables.sql` — GIS tables
   - `resources/gis/sapnhapbando_init_geo_json_objects_tbl.sql` — geo objects table
3. **Environment variables** (required for tests):
   - `POSTGRES_DB_HOST=localhost`, `POSTGRES_DB_PORT=15432`
   - `POSTGRES_DB_USERNAME=postgres`, `POSTGRES_DB_PSWD=root`
   - `POSTGRES_TMP_DB_NAME=vn_provinces_tmp`
4. **Test execution**: `go test -v ./...` from `dataset-generation-scripts/`

> **Note**: `first-workflow.yml` is a trivial hello-world stub, not functional CI.

---

## Subsystem Deep Dive

For detailed code-level context on Go services, database models, and implementation details, see:

👉 **[dataset-generation-scripts/CLAUDE.md](dataset-generation-scripts/CLAUDE.md)** — Comprehensive guide to:
- Automatic database query triggers
- Docker PostgreSQL setup
- Database schema and relationships
- Project structure and component descriptions
- Code conventions and error handling

---

## Memory & Feedback

This project maintains persistent learnings:
- [Automatic Skill Invocation](dataset-generation-scripts/memory/feedback_auto_db_query.md) — Skills should auto-trigger based on context, not require manual `/skill-name` invocation

---

## Quick Links

| Resource | Purpose |
|----------|---------|
| [README.md](README.md) | User guide — dataset installation & usage |
| [CLAUDE.md](dataset-generation-scripts/CLAUDE.md) | Code agent guide — detailed subsystem context |
| [dataset-generation-scripts/README.md](dataset-generation-scripts/README.md) | Maintainer guide — how to run generation scripts |
| [docs/gis/](docs/gis/) | User-facing GIS documentation (readme, example queries) |
| [development/](development/) | Feature documentation & planning artifacts |
| [patch/](patch/) | Historical decree patches & changelog |
| [resources/gis/](dataset-generation-scripts/resources/gis/) | GeoJSON & metadata source files |
| [resources/rules/](dataset-generation-scripts/resources/rules/) | Vietnamese text convention rules |

---

## When to Escalate

| Scenario | Action |
|----------|--------|
| New government decree needs incorporation | Create plan doc in `development/`, verify GIS ID matches before release |
| Database schema needs changes | Query current state first, document migration plan, test on temporary DB |
| Performance issue in data generation | Profile with `pprof`, check `memory/` for previous optimization notes |
| New database format requested | Study existing implementations (e.g., `dataset_writer/`), add new dialect support |

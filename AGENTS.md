# AI Agent Instructions — Vietnamese Provinces Database

A comprehensive database of Vietnamese administrative units (34 provinces, districts, wards) maintained through automated Go scripts that generate SQL, JSON, and GIS data for multiple database engines.

---

## Quick Start for AI Agents

### Building & Running

```bash
cd dataset-generation-scripts

# Run the generation scripts
go run main.go

# Run tests
go test -v ./...

# Check output in: ./output/
```

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
│   ├── internal/
│   │   ├── sapnhap_bando/           # Geographic data service (formerly API, now file-based)
│   │   │   ├── fetcher/             # Loads JSON metadata & GeoJSON files
│   │   │   ├── service/             # Business logic
│   │   │   ├── repository/          # Database operations
│   │   │   ├── model/               # Domain models
│   │   │   └── dto/                 # Data transfer objects
│   │   ├── dumper/                  # Reads admin data, persists to DB
│   │   ├── dataset_writer/          # Generates SQL/JSON/NoSQL output
│   │   ├── vn_provinces_tmp/        # Core VN provinces data layer
│   │   ├── gis_comparator/          # GIS data validation
│   │   ├── spatial_gis_comparator/  # Advanced spatial analysis
│   │   ├── geojson_fetcher/         # GeoJSON handling
│   │   └── database/                # Postgres connection pool
│   ├── resources/
│   │   └── gis/
│   │       ├── bando_gisserver/     # ← Provinces/wards JSON metadata
│   │       └── geojson_11Mar2026/   # ← GeoJSON geometry (from deprecated API)
│   ├── memory/
│   │   ├── MEMORY.md                # Memory index
│   │   └── feedback_*.md            # User preferences & learnings
│   ├── output/                       # Generated artifacts
│   └── patch/                        # Historical decree patches
├── development/                       # Feature documentation
│   ├── adapt_the_removal_of_sapnhap_api.md  # Context: API → file-based migration
│   └── [phase-plans].md              # Plans for new features
├── json/, mysql/, postgresql/, oracle/, sqlserver/, mongodb/, redis/
│   └── Generated dataset exports in various formats
└── .github/workflows/                # CI/CD pipelines
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

Geographic data migration context (March 2026):
- **Before**: Data fetched from SAPNhap API (`/pcotinh`, `/ptracuu`)
- **Now**: Loads from local files:
  - `./resources/gis/bando_gisserver/provinces.json`
  - `./resources/gis/bando_gisserver/wards.json`
  - `./resources/gis/geojson_11Mar2026/*.geojson`
- **Key IDs**: `mahc` (province)→`sapnhap_province_matinh` (GIS), `maxa` (ward)→`sapnhap_ward_maxa` (GIS)
- **Status**: All 3,355 records verified with 100% GIS ID match rate

---

## Decision Reference

### When to Query vs When to Generate

| Task | Approach | Tools |
|------|----------|-------|
| Check province/ward counts | Query | `docker exec psql` or `/db-query` |
| Verify data integrity (duplicates, orphans) | Query | SQL verification scripts |
| Compare old vs new GIS data | Query + analysis | `gis_comparator`, `spatial_gis_comparator` |
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

Each format in `{output/}` includes:
- `*_CreateTables_vn_units.sql` — Schema
- `*_ImportData_vn_units.sql` — Data insert statements
- GIS-specific files in `gis/` subdirectories

---

## Testing & Validation

- Unit tests in each `internal/*/` package (run with `go test -v ./...`)
- Data validation via `gis_comparator` for geometry matching
- Historical patch verification in `patch/` directory
- GIS server ID matching (100% validation before release)

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
- [Database Skill Preference](dataset-generation-scripts/memory/feedback_database_skill.md) — Use reusable skills for recurring tasks
- [Automatic Skill Invocation](dataset-generation-scripts/memory/feedback_auto_db_query.md) — Skills should auto-trigger based on context

---

## Quick Links

| Resource | Purpose |
|----------|---------|
| [README.md](README.md) | User guide — dataset installation & usage |
| [CLAUDE.md](dataset-generation-scripts/CLAUDE.md) | Code agent guide — detailed subsystem context |
| [development/](development/) | Feature documentation & planning artifacts |
| [patch/](patch/) | Historical decree patches & changelog |
| [resources/gis/](dataset-generation-scripts/resources/gis/) | GeoJSON & metadata source files |

---

## When to Escalate

| Scenario | Action |
|----------|--------|
| New government decree needs incorporation | Create plan doc in `development/`, verify GIS ID matches before release |
| Database schema needs changes | Query current state first, document migration plan, test on temporary DB |
| Performance issue in data generation | Profile with `pprof`, check `memory/` for previous optimization notes |
| New database format requested | Study existing implementations (e.g., `dataset_writer/`), add new dialect support |

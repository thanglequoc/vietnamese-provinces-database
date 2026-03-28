# Vietnamese Provinces Database - Claude Code Instructions

This file contains project-specific context for Claude Code to work effectively with this repository.

## When to Use Database Queries

**AUTOMATICALLY use database queries when the user asks about:**
- Data counts, totals, or statistics (e.g., "how many", "count", "total")
- Data verification or integrity checks (e.g., "check", "verify", "missing", "orphaned")
- Finding or searching for specific records (e.g., "find", "search", "show", "list")
- Database schema or table information (e.g., "what tables", "schema", "columns")
- GIS or geometry data (e.g., "geometry", "bbox", "geom", "GIS data")
- Any question about provinces, wards, or administrative data
- Data relationships or joins between tables
- **Direct database read requests** (e.g., "Read from the vn_provinces_tmp db", "Query from database", "Get from [table_name]")

**Do NOT wait for explicit `/db-query` invocation** - proactively use database queries when the context suggests the user needs information from the database.

**Examples of automatic triggers:**
- "How many wards are in Hà Nội?" → Run query immediately
- "Check if there are any missing GIS data" → Run verification query
- "Show me provinces without codes" → Query and display results
- "Verify the data integrity" → Run verification queries
- "Read from the vn_provinces_tmp db" → Execute database query
- "Get data from sapnhap_wards table" → Query the specified table

## Database Access

This project uses a PostgreSQL database with PostGIS extension running in Docker.

### Quick Database Access

The database is accessible via Docker container:
- Container: `vn_provinces_postgres_container`
- Database: `vn_provinces_tmp`
- Connection: `localhost:15432`

To run queries, use:
```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "QUERY"
```

### Database Schema

**Key Tables:**
- `provinces_tmp` (34 records) - Vietnam provinces with codes
- `wards_tmp` (3,321 records) - Vietnam wards with codes
- `sapnhap_provinces` (34 records) - Province metadata from SAPNhap site
- `sapnhap_wards` (3,321 records) - Ward metadata from SAPNhap site
- `sapnhap_provinces_gis` (34 records) - Province geometry (bbox, geom WKT)
- `sapnhap_wards_gis` (3,321 records) - Ward geometry (bbox, geom WKT)
- `sapnhap_geojson_objects` (3,355 records) - Combined geo objects

**Important Relationships:**
- `sapnhap_provinces.vn_province_code` → `provinces_tmp.code`
- `sapnhap_wards.vn_ward_code` → `wards_tmp.code`
- `sapnhap_provinces.mahc` → `sapnhap_provinces_gis.sapnhap_province_matinh`
- `sapnhap_wards.maxa` → `sapnhap_wards_gis.sapnhap_ward_maxa`

## Data Migration History

### SAPNhap API Deprecation (March 2026)

The upstream data site `sapnhap.bando.com.vn` deprecated their API endpoints:
- `/pcotinh` (provinces list)
- `/ptracuu` (wards lookup)

**Solution:** Migrated to local JSON and GeoJSON files:
- `./resources/gis/bando_gisserver/provinces.json` - Province metadata
- `./resources/gis/bando_gisserver/wards.json` - Ward metadata
- `./resources/gis/geojson_11Mar2026/` - GeoJSON geometry files

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
│   ├── sapnhap_bando/
│   │   ├── fetcher/      # Data fetching from local files
│   │   ├── service/      # Business logic
│   │   ├── repository/   # Database operations
│   │   ├── model/        # Database models
│   │   └── dto/          # Data transfer objects
│   ├── vn_provinces_tmp/ # VN provinces data layer
│   └── database/         # Database connection
├── resources/
│   └── gis/
│       ├── bando_gisserver/  # JSON metadata files
│       └── geojson_11Mar2026/ # GeoJSON geometry files
└── main.go
```

## Development Workflow

1. **Database Operations:** Use the `/db-query` skill to run SQL queries
2. **Running Scripts:** `go run .` from dataset-generation-scripts directory
3. **Database Migration:** Scripts handle schema and data seeding
4. **GIS Data:** All geometry stored in WKT format (PostGIS)

## AI Agent Planning Convention

**MANDATORY:** For any new feature developed with an AI agent, the final detailed plan must be saved as a Markdown file under the project's `development/` folder.

### Requirements

1. **Filename:** Use a short, descriptive summary of the feature (e.g., `backfill-province-codes-from-tmp-tables.md`)
2. **Content:** The plan must be sufficiently detailed to serve as a standalone implementation reference, covering:
   - **Objectives:** What the feature aims to achieve
   - **Affected Components:** Which files, tables, or systems are impacted
   - **Step-by-Step Logic:** Detailed implementation steps
   - **Edge Cases:** Potential issues and how to handle them
   - **Assumptions:** Any preconditions or assumptions made

### Example

```
development/
├── adapt_the_removal_of_sapnhap_api.md
└── your-new-feature-plan.md
```

This ensures that:
- All AI-assisted development is properly documented
- Future work can reference previous planning decisions
- Team members can understand the rationale behind implementation choices
- Complex features have a written specification before implementation

## Code Conventions

- **ORM:** Bun (uptrace/bun)
- **Database:** PostgreSQL with PostGIS
- **Language:** Go 1.23+
- **Naming:**
  - Database: snake_case
  - Go struct: PascalCase
  - JSON: snake_case
- **Error Handling:** Always wrap errors with context using `fmt.Errorf`

## Important File Paths

- `.env` - Database credentials and configuration
- `docker-compose.yml` - Docker services (if exists)
- `main.go` - Entry point for dataset generation
- `development/` - Development documentation

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

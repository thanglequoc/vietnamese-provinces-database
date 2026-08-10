> ⚠️ **Heads-up!**
> This section is for maintainers who need to regenerate the dataset. If you're looking to use the dataset, see the [root README](../README.md).

# Vietnamese Provinces Database Dataset Automation Scripts

The Vietnamese Government periodically issues decrees that change administrative units — merging wards, promoting them to higher units, etc. These automation scripts ingest the latest administrative data, enrich it with GIS geometry, and generate import scripts across multiple database formats.

## How it works

[![architecture diagram](https://i.postimg.cc/BnY9f29C/image.png)](https://postimg.cc/dhyS8k07)

The scripts operate in two phases:

- **Dumper** (`internal/dumper/`): Reads administrative data from the DVHCVN SOAP source, transforms the records, and inserts them into a temporary Postgres database
- **Dataset Writer** (`internal/dataset_writer/`): Reads from the temporary Postgres database and generates import scripts for multiple databases (PostgreSQL/MySQL, SQL Server, Oracle), plus JSON, MongoDB, and Redis exports

See [CLAUDE.md](CLAUDE.md) for detailed subsystem context.

## Prerequisites

- **Go 1.24+** (matches `go.mod`)
- **Docker** (for the temporary Postgres/PostGIS database)

All required data files (including GeoJSON geometry) are committed in the repository — no extra downloads or extraction needed.

## Setup

### 1. Clone the repository

```bash
git clone git@github.com:thanglequoc/vietnamese-provinces-database.git
cd vietnamese-provinces-database
```

### 2. Start the temporary Postgres database

```bash
cd dataset-generation-scripts
docker compose -f docker/docker-compose.yaml up -d
```

This starts a Postgres/PostGIS container named `vn_provinces_postgres_container` on port `15432` with database `vn_provinces_tmp`.

### 3. Configure the `.env` file

```bash
cp .env.example .env
```

The default values in `.env.example` match the Docker container — no edits needed for the standard setup. If you customized the Docker port or credentials, update `.env` accordingly.

## Run

```bash
go run main.go
```

Results land in the `output/` directory. By default, the script generates:

- SQL import scripts for PostgreSQL/MySQL, SQL Server, Oracle
- JSON, MongoDB, and Redis exports
- Elasticsearch NDJSON + mappings
- GIS SQL scripts and GeoJSON files

All exported formats include national postal codes: `postal_code_prefix` on
provinces and `postal_code` on wards (sourced from Quyết định 2334/QĐ-BKHCN via
`resources/postal/` seed files).

**Skipping GIS**: The `INCLUDE_GIS` constant in `main.go` defaults to `true`. Set it to `false` for a faster, admin-only run that skips GIS data fetching and geometry output — no internet connection required.

## Output structure

After a successful run, the `output/` directory contains:

```
output/
├── postgresql_mysql_generated_ImportData_vn_units_*.sql   # PostgreSQL & MySQL import
├── mssql_generated_ImportData_vn_units_*.sql               # SQL Server import
├── oracle_generated_ImportData_vn_units_*.sql              # Oracle import
├── json/
│   ├── full_json_generated_data_vn_units_*.json            # Full dataset (provinces + wards + districts)
│   ├── simplified_json_generated_data_vn_units_*.json      # Simplified names
│   └── vn_only_simplified_json_generated_data_vn_units_*.json  # Vietnamese-only simplified
├── mongodb/
│   ├── administrative_regions_*.json
│   ├── administrative_units_*.json
│   └── mongo_data_vn_unit_*.json                           # Full MongoDB import
├── redis/
│   └── redis_vn_provinces_dataset_*.redis                   # Redis commands
└── gis/                                                     # (only if INCLUDE_GIS=true)
    ├── *_ImportData_gis_*.sql                               # GIS SQL imports per engine
    ├── *_ImportData_gis_*.sql.zip                           # Compressed versions
    ├── vn_provinces_wards_geojson_*.zip                     # Combined GeoJSON archive
    └── geojson/                                             # Per-province GeoJSON
        ├── README.md
        ├── 01_ha_noi/
        │   ├── 01_ha_noi.geojson                            # Province boundary
        │   └── wards/                                       # Per-ward boundaries
        │       ├── 00004_ba_dinh.geojson
        │       └── ...
        ├── 04_cao_bang/
        └── ...
```

## Verify success

Check the expected record counts in the temporary database:

```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp \
  -c "SELECT COUNT(*) FROM provinces_tmp; SELECT COUNT(*) FROM wards_tmp;"
```

Expected: 34 provinces, 3,321 wards _(counts may change with new government decrees)_.

## Tests

```bash
# From dataset-generation-scripts/
go test -v ./...
```

**Important**: Docker must be running — most tests connect to the temporary Postgres database. Packages that require the database include `internal/sapnhap_bando/...` and `internal/dumper/...`. Pure unit tests (e.g., `internal/common/viet/`) will pass without Docker.

## Troubleshooting

| Symptom | Likely cause | Fix |
|---------|-------------|-----|
| `connect: connection refused` on port 15432 | Docker not running | `docker compose -f docker/docker-compose.yaml up -d` |
| Port 15432 already in use | Another Postgres instance on that port | Change the host port in `docker/docker-compose.yaml` and update `POSTGRES_DB_PORT` in `.env` |
| GIS pipeline errors or missing geometry | GeoJSON files missing from `resources/gis/geojson_11Mar2026/` | Ensure the repository was cloned correctly; the GeoJSON files are committed in Git |
| `package ... is not in GOROOT` | Go module dependencies not downloaded | Run `go mod download` from `dataset-generation-scripts/` |
| Tests fail with database connection errors | Database container not started | Verify with `docker ps`, check `.env` values match the Docker config |
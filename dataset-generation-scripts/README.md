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

### Dataset version metadata

The maintainer-controlled version source is `version.txt` (`dataset_version` +
`latest_decree`). Each run stamps it with a UTC `generated_at` and writes a separate
`vn_provinces_metadata` entity to every format:

| Format | Metadata location |
|--------|-------------------|
| PostgreSQL / MySQL / SQL Server / Oracle | `vn_provinces_metadata` table (`INSERT` in the base import script) |
| JSON | `json/vn_provinces_metadata.json` |
| MongoDB | `mongodb/mongo_data_vn_provinces_metadata.json` |
| Redis | `vnProvincesMetadata` hash |
| Elasticsearch | `vn_provinces_metadata` index (`vn_provinces_metadata.ndjson` + `mappings/vn_provinces_metadata.json`) |

**Skipping GIS**: The `INCLUDE_GIS` constant in `main.go` defaults to `true`. Set it to `false` for a faster, admin-only run that skips GIS data fetching and geometry output — no internet connection required.

## Automated decree detection (GitHub Actions)

The [`.github/workflows/decree-automation.yml`](../.github/workflows/decree-automation.yml)
workflow checks the GSO decree listing at
<https://danhmuchanhchinh.nso.gov.vn/NghiDinh.aspx> on a schedule
(Mon/Wed/Fri 08:00 ICT and Sat 22:00 ICT). When a decree that is already
effective is newer than `version.txt`'s `latest_decree`, it regenerates the
dataset and opens a PR against `master`.

How the decision is made (`internal/decree_check` + `cmd/decreecheck`):

1. Fetch the GSO page and parse the decree grid (decree number, issue date,
   effective date, content).
2. Pick the first row whose effective date is on or before today in
   `Asia/Ho_Chi_Minh`. The GSO list is newest-*published*-first, so
   future-effective decrees sitting at the top are skipped.
3. Compare it with `latest_decree` in `version.txt`. If it differs (and does not
   appear older), the workflow proceeds.

The `generate` job then:

1. Writes `dataset-generation-scripts/.env` for the PostGIS service container.
2. Runs `go run ./cmd/decreecheck --apply --decree "<n>"` — this minor-bumps
   `dataset_version` (middle digit, `v5.1.0 → v5.2.0`) and sets `latest_decree`.
3. Runs `go run main.go` and `./copy-datasets-to-repo.sh`.
4. Opens or updates the PR `auto/decree-<slug>` via
   `peter-evans/create-pull-request`.

**Required secret**: `DECREE_AUTOMATION_PAT` — a PAT (classic: `repo` +
`workflow`; fine-grained: Contents RW + Pull requests RW). A PAT is required so
the generated PR triggers `Test Go Code`; PRs opened with the default
`GITHUB_TOKEN` do not trigger `pull_request` workflows.

**Manual run** via `workflow_dispatch`:

- `force: true` — generate even when no new decree is detected.
- `dry_run: true` — generate but do not open the PR.

**Local dry run**:

```bash
cd dataset-generation-scripts
go run ./cmd/decreecheck \
  --html-file ./internal/decree_check/testdata/nghidinh_sample.html \
  --today 2026-09-21
```

## Output structure

After a successful run, the `output/` directory contains:

```
output/
├── postgresql_mysql_generated_ImportData_vn_units_*.sql   # PostgreSQL & MySQL import
├── mssql_generated_ImportData_vn_units_*.sql               # SQL Server import
├── oracle_generated_ImportData_vn_units_*.sql              # Oracle import
├── json/
│   ├── README.md                                           # Generated dataset README with bold timestamp
│   ├── full_json_generated_data_vn_units.json              # Full dataset (provinces + wards + districts)
│   ├── simplified_json_generated_data_vn_units.json        # Simplified names
│   ├── simplified_json_generated_data_vn_units_minified.json  # Simplified names (minified)
│   ├── vn_only_simplified_json_generated_data_vn_units.json  # Vietnamese-only simplified
│   ├── vn_only_simplified_json_generated_data_vn_units_minified.json  # Vietnamese-only simplified (minified)
│   ├── vn_provinces_wards_geojson.zip                      # Combined GeoJSON archive
│   └── geojson/                                            # Per-province GeoJSON
│       ├── 01_ha_noi/
│       │   ├── 01_ha_noi.geojson                           # Province boundary
│       │   └── wards/                                      # Per-ward boundaries
│       │       ├── 00004_ba_dinh.geojson
│       │       └── ...
│       ├── 04_cao_bang/
│       └── ...
├── mongodb/
│   ├── README.md                                           # Generated dataset README with bold timestamp
│   ├── administrative_units.json
│   ├── administrative_regions.json
│   ├── mongo_data_vn_unit.json                             # Full MongoDB import
│   └── gis/                                                # provinces-gis / wards-gis collections
└── redis/
    ├── README.md                                           # Generated dataset README with bold timestamp
    └── redis_vn_provinces_dataset.redis                     # Redis commands
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
| Decree check fails with `markup may have changed` | The GSO page markup changed | Update the patterns in `internal/decree_check/decree_check.go` and refresh `internal/decree_check/testdata/nghidinh_sample.html` |
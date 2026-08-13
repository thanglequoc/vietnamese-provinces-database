# Design: Per-Dataset READMEs and Deterministic GIS Filenames

**Date:** 2026-08-14
**Branch:** `134_AddPostalCode`
**Status:** Approved

## Objective

1. Generate a `README.md` for **every** dataset type from its own Go writer, so
   the published folders are self-describing. Each README includes a bold
   generation timestamp, the file list with sizes, the data structure, and
   sample queries in the native client syntax.
2. Remove the datetime suffix from all GIS output filenames (postgres, mysql,
   mssql, mongodb gis) to prevent duplicate timestamp variants accumulating in
   the published folders.

## Requirements (confirmed with user)

1. README generation is implemented in each dataset writer (not in the copy
   script). Applies to **all** dataset types: postgresql, mysql, sqlserver,
   oracle, mongodb, redis, elasticsearch, json.
2. Existing generated READMEs (json, elasticsearch, mongodb/gis) are refactored
   to the same shared format for consistency.
3. Remove the datetime suffix from GIS files in the postgres/mysql/mssql/mongodb
   gis writers. Elasticsearch part files already have no suffix — unchanged.
4. The copy script must prune old timestamped GIS variants from the published
   folders so no duplicates remain.

## Changes

### 1. Shared README helper (`dataset_file_writer` package)

New shared helper, e.g. in `dataset_file_writer.go`:

```go
type DatasetReadmeFile struct {
    Name        string
    Description string
}

// writeDatasetReadme renders README.md at outputFolderPath with a bold
// generation timestamp, a "Files" list with per-file sizes, and
// dataset-specific markdown sections appended verbatim.
func writeDatasetReadme(
    outputFolderPath string,
    title string,
    intro string,
    files []DatasetReadmeFile,
    sections []string,
) error
```

Rendered structure:

```
# <title>

**Generated at: <time.Now().Format(time.RFC1123Z)>**

<intro>

## Files

- `<name>` — <description> (<size>)
...

## Data Structure
<dataset-specific>

## Sample Queries
<dataset-specific, native client syntax>

## GIS / GeoJSON
<where applicable>
```

`formatFileSize(size int64) string` (already in `json_file_writer.go`) moves to
the shared helper so all writers reuse it.

### 2. Deterministic GIS filenames (writers)

- `postgres_mysql_dataset_file_writer.go`: remove `fileTimeSuffix` from
  `postgresGISFilePath` / `mysqlGISFilePath` →
  `postgresql_ImportData_gis.sql` / `mysql_ImportData_gis.sql`. Chunked output
  becomes `postgresql_ImportData_gis-part-NN.sql` + `.sql.manifest`.
- `mssql_dataset_file_writer.go`: `mssql_ImportData_gis.sql` → chunked
  `mssql_ImportData_gis-part-NN.sql` + `.manifest`.
- `mongodb_gis_file_writer.go`: `mongo_data_vn_province_gis.json`,
  `mongo_data_vn_ward_gis.json` → chunked `mongo_data_vn_ward_gis_part_NN.json`
  + `.json.manifest`.
- Remove the now-unused `fileTimeSuffix` locals in these three functions.
- Elasticsearch: unchanged (already deterministic).

### 3. README generation per writer

At the end of `WriteToFile`, each writer calls `writeDatasetReadme`:

| Writer | Output path(s) | Files listed |
|--------|----------------|--------------|
| `PostgresMySQLDatasetFileWriter` | `output/postgresql/README.md`, `output/mysql/README.md` | `postgres_ImportData_vn_units.sql` / `mysql_ImportData_vn_units.sql`, `gis/` subfolder |
| `MssqlDatasetFileWriter` | `output/sqlserver/README.md` | `mssql_ImportData_vn_units.sql`, `gis/` |
| `OracleDatasetFileWriter` | `output/oracle/README.md` | `oracle_ImportData_vn_units.sql` |
| `MongoDBDatasetFileWriter` | `output/mongodb/README.md` | `administrative_units.json`, `administrative_regions.json`, `mongo_data_vn_unit.json`, `gis/` |
| `RedisDatasetFileWriter` | `output/redis/README.md` | `redis_vn_provinces_dataset.redis` |
| `ElasticsearchDatasetFileWriter` | `output/elasticsearch/README.md` | refactor existing `writeESReadme` to shared format (NDJSON parts, mappings) |
| `JSONDatasetFileWriter` | `output/json/README.md` | refactor existing `writeJSONDatasetReadme` to shared format + add structure/queries sections |

Each README contains:
- A 1-2 line intro naming the dataset type.
- `## Data Structure` — brief description of the exported tables/documents/keys,
  covering **both** the regular dataset and the GIS subfolder (e.g. the
  `gis_provinces`/`gis_wards` tables for SQL, the `provinces-gis`/`wards-gis`
  collections for MongoDB).
- `## Sample Queries` — 2-3 queries in the native client (SQL for the SQL
  engines, `mongosh`, `redis-cli`, `curl` for Elasticsearch, plain JSON read for
  json), again covering both regular and GIS data.
- `## GIS / GeoJSON` section where a `gis/` or `geojson/` subfolder exists,
  describing its files.

### 3b. Remove gis-subfolder README writers

The main dataset README is the single source of truth. Remove the
gis-subfolder README generation:

- `mongodb_gis_file_writer.go`: remove the `writeMongoGISReadme` call and
  function — no more `output/mongodb/gis/README.md`.
- `geojson_file_writer.go`: remove `writeGeoJSONReadme` — no more
  `output/json/geojson/README.md`.
- The geojson zip archive no longer embeds `geojson/README.md` — it contains
  only the geojson data files.

### 4. Copy script (`copy-datasets-to-repo.sh`)

After copying each SQL/mongo dataset, prune old timestamped GIS variants in the
published folders:

- `postgresql/gis/postgresql_ImportData_gis_*.sql*`
- `mysql/gis/mysql_ImportData_gis_*.sql*`
- `sqlserver/gis/mssql_ImportData_gis_*.sql*`
- `mongodb/gis/mongo_data_vn_province_gis_*.json`
- `mongodb/gis/mongo_data_vn_ward_gis_*.json*`
- `json/geojson/README.md` (removed so the stale gis-subfolder README does not
  linger in the published folder)

(glob patterns that match the datetime-suffixed variants but not the new
fixed-name files). The generated READMEs copy over naturally via `cp -R`.

### 5. Tests

- Update chunk/manifest filename assertions in `postgres_mysql_dataset_file_writer_test.go`,
  `mssql_dataset_file_writer_test.go`, `mongodb_gis_file_writer_test.go` to the
  new deterministic names.
- Add/update README assertions for each writer (bold timestamp, file list,
  structure, sample queries covering regular + GIS).
- JSON README test updated for the shared format.
- Elasticsearch README test updated if the shared format changes its content.
- Geojson test: remove the `geojson/README.md`-in-zip assertion and the README
  file assertions in the geojson output folder.
- MongoDB GIS test: assert `README.md` is no longer written to the gis folder.

### 6. Docs

- Update `dataset-generation-scripts/README.md` output tree to the deterministic
  GIS filenames.
- Update `docs/gis/gis_readme.md` / `gis_readme_vi.md` if they reference
  `_<ts>` part-file naming.

## Edge Cases

- **Duplicate GIS variants**: because output names are deterministic, each run
  overwrites the previous output. The copy script prunes old timestamped
  variants from the published folders so only one set remains.
- **README timing**: READMEs are written at the end of each writer's
  `WriteToFile` (non-GIS phase) so they always exist, even on admin-only runs.
  GIS subfolder contents are described in the README's GIS section, not listed
  with sizes.
- **Existing static READMEs**: mongodb/redis/elasticsearch top-level READMEs are
  replaced by generated ones on the next copy; content is derived from the
  writer output, not hand-maintained.
- **No gis-subfolder READMEs**: `mongodb/gis/README.md` and
  `json/geojson/README.md` no longer exist; the parent dataset README covers the
  GIS content. The geojson zip contains only data files.

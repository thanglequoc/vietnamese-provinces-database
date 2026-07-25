# Design: Archive GIS SQL Files as .zip During Dataset Generation

**Date**: 2026-07-25
**Status**: Draft
**Related**: GIS dataset generation pipeline (`GenerateGISSQLDatasets()`)

---

## Objective

When the GIS dataset generation pipeline produces SQL files for PostgreSQL, MySQL, and MSSQL, it should also automatically archive each file as an individually zipped `.sql.zip` archive — matching the existing pattern used for GeoJSON exports.

---

## Current State

The `GenerateGISSQLDatasets()` function in `dataset_writer.go` produces three raw `.sql` files in `output/gis/`:

| Engine | Output File |
|--------|-------------|
| PostgreSQL | `output/gis/postgresql_ImportData_gis_{timestamp}.sql` |
| MySQL | `output/gis/mysql_ImportData_gis_{timestamp}.sql` |
| MSSQL | `output/gis/mssql_ImportData_gis_{timestamp}.sql` |

The GeoJSON export (`geojson_file_writer.go`) already produces a `vn_provinces_wards_geojson_{timestamp}.zip` archive via `archiveGeoJSONDirectory()`. The SQL GIS files are not zipped.

The published output folders (`postgresql/gis/`, `mysql/gis/`, `sqlserver/gis/`) already contain `.sql.zip` files from past runs (placed manually or via another process), confirming this is the expected deliverable format.

## Desired State

After generation, the `output/gis/` folder should contain:

```
output/gis/
├── geojson/                                         # existing GeoJSON export
├── postgresql_ImportData_gis_{timestamp}.sql         # raw SQL (keep)
├── postgresql_ImportData_gis_{timestamp}.sql.zip     # NEW — compressed archive
├── mysql_ImportData_gis_{timestamp}.sql              # raw SQL (keep)
├── mysql_ImportData_gis_{timestamp}.sql.zip          # NEW — compressed archive
├── mssql_ImportData_gis_{timestamp}.sql              # raw SQL (keep)
├── mssql_ImportData_gis_{timestamp}.sql.zip          # NEW — compressed archive
└── vn_provinces_wards_geojson_{timestamp}.zip        # existing GeoJSON archive
```

**Key decisions**:
- Individual `.zip` per SQL file (user preference)
- Raw `.sql` files are retained alongside the `.zip` archives
- Follows the same compression approach as the existing GeoJSON archiver (`archive/zip` + `compress/flate` with `BestCompression`)

---

## Approach: Shared `zipFile` Utility + Per-Writer Calls

### Why This Approach

- **DRY**: One `zipFile` helper, reused in both writers
- **Mirrors existing pattern**: `archiveGeoJSONDirectory()` already lives alongside GeoJSON writing code in `geojson_file_writer.go`
- **Minimal change surface**: each writer just adds a single post-flush call

### New Shared Utility

**Location**: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/dataset_file_writer.go`

This file already contains shared package-level helpers (`getFileTimeSuffix`, `escapeSingleQuote`, `parseEuropeanFloat`). The new helper follows the same pattern.

```go
// zipFile compresses a single file to <sourcePath>.zip using best compression.
// Returns nil on success. If the source file cannot be read or the archive
// cannot be created, returns a wrapped error.
func zipFile(sourcePath string) error
```

**Implementation sketch** (based on `archiveGeoJSONDirectory` in `geojson_file_writer.go`, simplified for a single file):

1. Open source file for reading
2. Create destination file at `sourcePath + ".zip"`
3. Create `zip.NewWriter(dest)` with `flate.BestCompression`
4. Create zip entry using `FileInfoHeader`
5. Copy source contents into the zip entry
6. Close both files and the zip writer

### Modifications to Existing Writers

**`PostgresMySQLDatasetFileWriter.WriteGISDataToFile()`** (`postgres_mysql_dataset_file_writer.go`):
- After `postgresScriptDataWriter.Flush()`, add: `_ = zipFile(postgresGISFilePath)`
- After `mysqlScriptDataWriter.Flush()`, add: `_ = zipFile(mysqlGISFilePath)`

**`MssqlDatasetFileWriter.WriteGISDataToFile()`** (`mssql_dataset_file_writer.go`):
- After `mssqlScriptDataWriter.Flush()`, add: `_ = zipFile(mssqlGISFilePath)`

### Error Handling

Zip failures are **non-fatal** — they log a warning via `fmt.Printf` (consistent with other print-style logging in the codebase, e.g., `fmt.Println("✅ ...")` and `fmt.Fprintf(os.Stderr, ...)`) but do not abort generation. Rationale:

- The raw `.sql` file is the canonical output and is already successfully written
- A zip failure (e.g., disk full, permissions) should not invalidate an otherwise successful run
- This matches `GenerateGISSQLDatasets()`, which treats each writer independently (failure in MSSQL doesn't affect PostgreSQL/MySQL)

**Signature choice**: `_ = zipFile(...)` discards the error at the call site and logs inside the function. Alternatively, return from the function only if the underlying SQL write succeeded. The spec chooses the former as it's simpler and the zip is strictly additive.

---

## Changes Summary

| File | Change |
|------|--------|
| `internal/dataset_writer/dataset_file_writer/dataset_file_writer.go` | Add `zipFile(sourcePath string) error` helper |
| `internal/dataset_writer/dataset_file_writer/postgres_mysql_dataset_file_writer.go` | Call `zipFile()` after flushing postgres and mysql GIS files |
| `internal/dataset_writer/dataset_file_writer/mssql_dataset_file_writer.go` | Call `zipFile()` after flushing mssql GIS file |

**No other files change.** `GenerateGISSQLDatasets()` in `dataset_writer.go` is unaffected — it only orchestrates the writer calls.

---

## Testing

- **Unit test**: `dataset_file_writer_test.go` — test `zipFile()` creates a valid `.zip` archive containing the expected file with correct content. Pass a temp file with known content, verify the archive extracts correctly.
- **Integration**: Run `go run main.go`, inspect `output/gis/` for the three `.zip` files, verify each contains exactly one `.sql` file and extracts cleanly.

---

## Dependencies

- `archive/zip` — already imported in `geojson_file_writer.go`
- `compress/flate` — already imported in `geojson_file_writer.go`
- `io` — standard library
- `os` — standard library

No new external dependencies.

---

## Out of Scope

- Publishing/copying the `.zip` files to the top-level output directories (`postgresql/gis/`, `mysql/gis/`, `sqlserver/gis/`) — that is a separate copy step, likely a CI/CD concern or existing manual process
- Archiving non-GIS SQL files (e.g., `*_ImportData_vn_units_*.sql`) — this is limited to GIS output only
- Adding a combined all-in-one archive of all SQL files

---

## Rollout

1. Add `zipFile` helper and its unit test
2. Add calls in both writers
3. Run full generation `go run main.go`
4. Verify `output/gis/` contains all three `.zip` files
5. Spot-check one archive: `unzip -l output/gis/postgresql_ImportData_gis_*.sql.zip`
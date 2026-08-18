# Chunk GIS SQL Output for PostgreSQL / MySQL / MSSQL

Date: 2026-08-10
Status: Approved design (ready for implementation plan)

## Objective

Replace the single-file GIS SQL dataset output for **PostgreSQL**, **MySQL**, and
**MSSQL (SQL Server)** with chunked output. Currently each engine writes one
~150 MB `.sql` file, then zips it to ~44 MB so it stays under GitHub's 50 MB
file limit. The new generation should:

- Emit **chunk files**, each **under 40 MB** (matching the Elasticsearch
  `maxNDJSONChunkSize` limit), so no file exceeds GitHub's 50 MB warning.
- **Never emit a single-file** GIS `.sql` (always chunked, even if only one
  `-part-01.sql` is produced).
- **Remove the `.sql.zip` output** entirely (chunks are already under the limit).
- Follow the existing Elasticsearch chunking implementation for naming,
  manifest format, and size-limit semantics.

## Background / Current State

| Aspect | Current behavior |
|--------|------------------|
| Postgres writer | `internal/dataset_writer/dataset_file_writer/postgres_mysql_dataset_file_writer.go` → `WriteGISDataToFile` writes `./output/postgresql/gis/postgresql_ImportData_gis_<ts>.sql`, then `zipFile(...)`. |
| MySQL writer | Same file/struct, writes `./output/mysql/gis/mysql_ImportData_gis_<ts>.sql`, then zips. |
| MSSQL writer | `internal/dataset_writer/dataset_file_writer/mssql_dataset_file_writer.go` → `WriteGISDataToFile` writes `./output/sqlserver/gis/mssql_ImportData_gis_<ts>.sql`, then zips. |
| Zip helper | `zipFile` in `dataset_file_writer.go` — only callers are the two GIS writers (plus its own tests). |
| File sizes | ~150 MB raw `.sql`, ~44 MB `.sql.zip` (Aug-2026 run). |
| ES reference | `elasticsearch_file_writer.go` → `writeChunkedNDJSON` (40 MB limit, `-part-NN` naming, `<path>.manifest`, single-file short-circuit). |
| Mongo reference | `mongodb_gis_file_writer.go` → `writeChunkedMongoJSON` (50 MB limit, `_part_NN` naming, `<path>.manifest`). |
| Published artifacts | Top-level `postgresql/gis/`, `mysql/gis/`, `sqlserver/gis/` contain static `*_CreateGISTables.sql` + one `.sql.zip` each (dated 2026-08-08). |

Timestamps: `getFileTimeSuffix()` produces e.g. `2026-08-10__21_55_01`
(`time.DateTime` with `:`→`_`, ` `→`__`).

## SQL Statement Granularity

GIS SQL files are `INSERT`-only data scripts (tables are created by the separate
`*_CreateGISTables.sql` bootstrap scripts). Content layout:

1. Header comment block (engine banner, created-at, reference URL).
2. `-- DATA for gis_provinces --` section — one `INSERT` statement per province
   (34 statements).
3. Separator comment + `-- DATA for gis_wards --` section — ward rows are written
   as batched multi-VALUES `INSERT` statements, `batchInsertItemSize = 50` rows
   per statement.
   - Postgres/MySQL: each batch ends with `;\n\n`.
   - MSSQL: each batch ends with `;\n\nGO\n\n` (batch separator for `sqlcmd`).
4. Footer comment `-- END OF SCRIPT FILE --`.

**Chunk boundaries are SQL-statement boundaries.** A chunk is a contiguous,
ordered run of complete statements. Each chunk is individually valid, runnable
SQL and is **self-describing**: it starts with its own SQL header comment
(timestamp, repository link, and "part X of N"), so users can tell what a chunk
is and where it sits in the sequence without opening the manifest. (Because every
chunk carries its own header, byte-exact concatenation of parts no longer
reproduces the old single-file bytes — instead the ordered statements within the
parts, read via the manifest, fully reconstruct the dataset.)

Ward batches are ~2 MB per statement; provinces are ~1–3 MB each — well under
the 40 MB limit, so no single block needs splitting.

## Design

### 1. Shared chunk helper (new file `gis_sql_chunk_writer.go`)

Same package `dataset_writer` (file `dataset_file_writer/`). Reuses the existing
private helpers `filepathDir`, `filepathBase`, `filepathExt`, `stringsJoin`.

```go
// maxSQLGISChunkSize is the maximum size of a single GIS SQL chunk file.
// Matches the Elasticsearch maxNDJSONChunkSize (40 MB) and stays safely under
// GitHub's 50 MB file warning. var (not const) so tests can override it.
var maxSQLGISChunkSize = 40 * 1024 * 1024 // 40 MB

// chunkHeaderInfo carries the fixed text rendered into every chunk's leading
// SQL comment. The part/total numbers are interpolated per chunk at write time.
type chunkHeaderInfo struct {
    Banner     string // e.g. "Add-on GIS Dataset for PostgreSQL of Vietnamese Provinces Database"
    CreatedAt  string // e.g. time.Now().Format(time.RFC1123Z)
    Repository string // e.g. "https://github.com/thanglequoc/vietnamese-provinces-database"
}

// writeChunkedSQLFile writes complete-SQL blocks as chunk files, each under
// maxSQLGISChunkSize. It always emits chunk files — never a single file — and
// writes a manifest at path + ".manifest" listing chunk filenames in order.
// Chunks are named <base>-part-NN<ext> (e.g. postgresql_ImportData_gis_<ts>-part-01.sql),
// matching the Elasticsearch naming convention.
func writeChunkedSQLFile(path string, blocks [][]byte, header chunkHeaderInfo) error
```

Per-chunk leading header comment (rendered on **every** chunk, not just the
first), following the existing banner style:

```sql
/* === <Banner> === */
/* Part 1 of N */
/* Created at:  <CreatedAt> */
/* Reference: <Repository> */
/* =============================================== */
```

(`Part 1 of N` is computed from the chunk index + total chunk count, which are
only known after greedy packing completes.)

Behavior:

- A `block` is an atomic byte slice (one complete statement, or a comment /
  separator block that must stay with the surrounding statements). Blocks are
  never split.
- Greedy packing: iterate blocks in order; when
  `currentSize + len(block) > maxSQLGISChunkSize` and the current chunk is
  non-empty, close the chunk and start a new one. Then append the block.
- Degenerate case: a single block larger than the limit (not expected) goes into
  its own oversized chunk — same semantics as the ES chunker.
- Write each chunk to `dir/<nameNoExt>-part-%02d<ext>` with `os.O_CREATE|os.O_WRONLY|os.O_TRUNC`.
- Every chunk file starts with the per-chunk header comment (banner, `Part X of
  N`, created-at, repository link) followed by that chunk's blocks.
- Write manifest `path + ".manifest"` = `strings.Join(chunkNames, "\n") + "\n"`.
- Log a summary line (chunk count, per-chunk MB) like the ES/Mongo chunkers.
- Return wrapped errors (`fmt.Errorf` with context) per project convention.

Naming example (single engine):

```
output/postgresql/gis/postgresql_ImportData_gis_2026-08-10__21_55_01-part-01.sql
output/postgresql/gis/postgresql_ImportData_gis_2026-08-10__21_55_01-part-02.sql
...
output/postgresql/gis/postgresql_ImportData_gis_2026-08-10__21_55_01.sql.manifest
```

The `path` argument is the would-be single-file path; its directory is used for
chunks and manifest. There is no `.zip` anywhere in this output.

### 2. Refactor `PostgresMySQLDatasetFileWriter.WriteGISDataToFile`

- Keep the existing loop structure and SQL template logic (Postgres OGC lng-lat
  order vs MySQL EPSG lat-lng order).
- Instead of opening two files and streaming with `bufio.Writer`, build two
  `[][]byte` block lists (`postgresBlocks`, `mysqlBlocks`) while looping over
  provinces and ward batches. Each block is the fully-formatted statement bytes
  (including the trailing `;\n\n` / `,\n` separators as today).
- Build a `chunkHeaderInfo` per engine with the existing banner text
  (`/* === Add-on GIS Dataset for PostgreSQL ... === */` / `... for MySQL ...`),
  `time.Now().Format(time.RFC1123Z)` created-at, and the repository URL
  (`https://github.com/thanglequoc/vietnamese-provinces-database` — lowercase
  `thanglequoc`).
- Call `writeChunkedSQLFile(postgresGISFilePath, postgresBlocks, postgresHeader)`
  and `writeChunkedSQLFile(mysqlGISFilePath, mysqlBlocks, mysqlHeader)`.
- Remove `zipFile` calls and the single-file writes.
- Keep the hardcoded output directories (`./output/postgresql/gis`,
  `./output/mysql/gis`) and timestamped filenames unchanged, to avoid touching
  `dataset_writer.go` wiring and to keep the existing `os.Chdir`-based tests'
  glob pattern working.

### 3. Refactor `MssqlDatasetFileWriter.WriteGISDataToFile`

- Same approach: build an `[][]byte` block list using the existing MSSQL
  templates (`geometry::STGeomFromText(...)`, `GO` batch separators), plus a
  `chunkHeaderInfo` with the MSSQL banner text.
- Call `writeChunkedSQLFile(mssqlGISFilePath, blocks, header)`.
- Remove `zipFile` call and single-file write.

### 4. Remove `zipFile`

- Delete `zipFile` from `dataset_file_writer.go` (it becomes dead code).
- Remove the three `zipFile` tests in `dataset_file_writer_test.go`
  (`TestZipFile_CreatesValidArchive`, `TestZipFile_SourceFileNotFound`,
  `TestZipFile_EmptyFile`).
- Keep the `archive/zip` / `compress/flate` / `io` imports only if still needed
  elsewhere in the file (they are not, after removal).

### 5. Tests

- **New** `TestWriteChunkedSQLFile` cases (in a new or existing test file):
  - Single part: small blocks produce exactly `-part-01.sql` + manifest listing
    just that one part.
  - Multiple parts: override `maxSQLGISChunkSize` (e.g. 100 bytes) so small
    blocks split; assert ≥2 parts, manifest lists all parts in order, each part
    ≤ limit, and every part starts/ends at a block boundary (no partial block).
  - **Header on every chunk:** assert each part's leading comment contains the
    banner, the repository URL, a timestamp, and the correct `Part X of Y` line
    (e.g. `Part 1 of 3` on part-01). Stripping the header from each part and
    concatenating the bodies in manifest order equals the original block
    sequence.
  - Per-part size logging path.
- **Update** the two existing Postgres/MySQL GIS tests
  (`TestPostgresMySQLDatasetFileWriter_WriteGISDataToFile_MySQLWardBatchUsesCommaWithinBatch`,
  `TestPostgresMySQLDatasetFileWriter_WriteGISDataToFile_MySQLWardBatchSplitsAtBatchSize`):
  - They currently read `output/mysql/gis/mysql_ImportData_gis_*.sql` (single
    file). Change to read the manifest, then read/concatenate the `-part-NN.sql`
    files (or read the first part directly), keeping the existing content
    assertions (batch comma/separator behavior, 2 INSERT statements at batch
    size, etc.).
  - The test `readGeneratedGISFile` helper must be updated accordingly.
- **New** MSSQL GIS test (none exists today) exercising the mssql writer with a
  small dataset: assert chunk + manifest output, per-chunk header comment, and
  correct `GO` handling.
- Verify existing `zipFile` tests are removed and nothing else references
  `zipFile`.

### 6. Documentation updates

- `docs/gis/gis_readme.md` and `docs/gis/gis_readme_vi.md`:
  - PG/MySQL/MSSQL "Import GIS Data" steps: replace single-file
    `postgresql_ImportData_gis_*.sql` references with instructions to import all
    `*-part-NN.sql` files in order (read the `.manifest` for the ordered list).
  - Update the `[gis_dataset_{postgresql,mysql,sqlserver}_bucket_url]` link
    definitions at the bottom (docs lines ~490–500) to point to the `.manifest`
    file (e.g. `.../GISDataSet/postgresql_ImportData_gis_<ts>.sql.manifest`) and
    note that the data is split into `< 40 MB` parts listed by the manifest.
- `dataset-generation-scripts/README.md` output-structure listing
  (lines ~90–92): replace `*_ImportData_gis_*.sql` + `*_ImportData_gis_*.sql.zip`
  with `*_ImportData_gis_*-part-*.sql` + `*.manifest`.
- **Repository URL casing:** the existing header comments hardcode
  `https://github.com/ThangLeQuoc/vietnamese-provinces-database` in
  `postgres_mysql_dataset_file_writer.go` (lines 62, 172, 178),
  `mssql_dataset_file_writer.go` (lines 56, 153), and
  `oracle_dataset_file_writer.go` (line 42), plus a clone URL in
  `dataset-generation-scripts/README.md` (line 31). Update all to lowercase
  `thanglequoc` so generated output is consistent.
- `AGENTS.md` (line 731) references `output/postgresql/gis/` only as the staging
  area — still accurate, no change needed. `dataset-generation-scripts/CLAUDE.md`
  has no matching references.

### 7. Published artifacts (top-level folders)

- Delete tracked zips:
  - `postgresql/gis/postgresql_ImportData_gis_2026-08-08__21_32_14.sql.zip`
  - `mysql/gis/mysql_ImportData_gis_2026-08-08__21_32_14.sql.zip`
  - `sqlserver/gis/mssql_ImportData_gis_2026-08-08__21_33_02.sql.zip`
- Regenerate the dataset via `go run main.go` (requires the local Postgres/PostGIS
  Docker container on port 15432 and network access to sapnhap.bando.com.vn).
- Copy the freshly generated chunk files + manifests from
  `dataset-generation-scripts/output/{postgresql,mysql,sqlserver}/gis/` into the
  matching top-level `*/gis/` folders.
- Commit the new artifacts (plain files, no LFS needed; each under 40 MB).
- Keep the static `*_CreateGISTables.sql` files unchanged.

## Out of Scope

- MongoDB and Elasticsearch output (already chunked).
- Non-GIS SQL datasets (`postgres_ImportData_vn_units.sql` etc.) — unchanged.
- The GeoJSON archive (`vn_provinces_wards_geojson_*.zip`) — unchanged.
- Changing the hardcoded output directories to honor `w.OutputFilePath` for the
  GIS writers (existing inconsistency; kept as-is to limit blast radius).
- Adding a new integration-test script for importing the chunked SQL parts.

## Edge Cases / Assumptions

- Block sizes are all well under 40 MB, so no block needs further splitting.
- `blocks` is never empty (header block always present). Guard anyway: if no
  blocks, return early/nil without writing a manifest.
- Chunk boundaries land exactly between statements, so each chunk is valid SQL
  on its own.
- Every chunk carries its own self-describing header comment (banner, `Part X of
  Y`, created-at, repository link), so a chunk is identifiable on its own.
- The manifest is the source of truth for import order; it lists only the chunk
  filenames (one per line), matching the ES/Mongo manifest format.
- After regeneration, published artifacts use the new chunk naming; any external
  links to single-file URLs in the docs are updated as part of this change.

## Verification

1. `go test -v ./internal/dataset_writer/...` passes (chunk helper + updated
   Postgres/MySQL/MSSQL writer tests; zip tests removed).
2. `go run main.go` completes and produces chunk files + manifests in
   `output/{postgresql,mysql,sqlserver}/gis/`, with every chunk ≤ 40 MB and no
   `.sql.zip` / single `.sql` left behind.
3. Each chunk starts with its header comment (banner, `Part X of Y`, created-at,
   repo link); stripping headers and concatenating the bodies in manifest order
   reproduces the expected SQL statement sequence (verified by tests).
4. Top-level published folders contain chunk files + manifests and no `.sql.zip`.
5. Docs reference the new chunked naming.

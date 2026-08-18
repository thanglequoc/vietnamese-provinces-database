# Chunk GIS SQL Output (Postgres/MySQL/MSSQL) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the single ~150 MB GIS SQL files (and their `.sql.zip` workarounds) for PostgreSQL, MySQL, and MSSQL with chunked `-part-NN.sql` files under 40 MB each, plus an ordered `.manifest`, following the existing Elasticsearch chunking pattern.

**Architecture:** A shared `writeChunkedSQLFile(path, blocks, header)` helper (package `dataset_writer`) packs complete-SQL byte blocks greedily into chunks under `maxSQLGISChunkSize` (40 MB), always emitting `-part-NN` files plus a `<path>.manifest`, with a self-describing SQL header comment (`/* === Banner === */`, `/* Part X of N */`, `/* Created at: */`, `/* Reference: */`) on every chunk. The Postgres/MySQL and MSSQL GIS writers are refactored to build block lists and delegate to this helper; `zipFile` and the single-file output are removed.

**Tech Stack:** Go 1.24.0, standard library only (no new dependencies), Testify for tests.

**Design doc:** `development/chunk-gis-sql-output.md`

## Global Constraints

- Chunk size limit: `maxSQLGISChunkSize = 40 * 1024 * 1024` (40 MB), declared as a `var` so tests can override it. Matches ES `maxNDJSONChunkSize`.
- **Always chunk** — never write a single-file `.sql`. Even one part is named `-part-01.sql`. A `.manifest` is always written.
- No `.sql.zip` output anywhere. Remove `zipFile` entirely (dead code).
- Chunk boundaries only ever fall **between** complete SQL statements (blocks are atomic).
- Every chunk starts with the header comment block; `Part X of N` uses the chunk's 1-based index and the total chunk count.
- Repository URL is **lowercase**: `https://github.com/thanglequoc/vietnamese-provinces-database`. Fix all existing `ThangLeQuoc` occurrences.
- Manifest format: one chunk filename per line, trailing newline (identical to ES/Mongo manifests).
- Chunk file naming: `<base>-part-%02d<ext>` (e.g. `postgresql_ImportData_gis_<ts>-part-01.sql`).
- Keep hardcoded output dirs (`./output/{postgresql,mysql,sqlserver}/gis`) and `getFileTimeSuffix()` timestamps unchanged.
- Tests use Testify (`assert`). Run tests from `dataset-generation-scripts/` with `go test -v ./...`.

---
---

### Task 1: Shared chunk helper `writeChunkedSQLFile` + `chunkHeaderInfo`

**Files:**
- Create: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/gis_sql_chunk_writer.go`
- Test: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/gis_sql_chunk_writer_test.go`

**Interfaces:**
- Consumes: package-private helpers already defined in `elasticsearch_file_writer.go` (same package): `filepathDir(path string) string`, `filepathBase(path string) string`, `filepathExt(path string) string`, `stringsJoin(strs []string, sep string) string`.
- Produces:
  - `var maxSQLGISChunkSize = 40 * 1024 * 1024`
  - `type chunkHeaderInfo struct { Banner, CreatedAt, Repository string }`
  - `func writeChunkedSQLFile(path string, blocks [][]byte, header chunkHeaderInfo) error`

- [ ] **Step 1: Write the failing test**

Create `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/gis_sql_chunk_writer_test.go`:

```go
package dataset_writer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteChunkedSQLFile_SinglePart(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "mysql_ImportData_gis_2026-08-10__21_55_01.sql")

	header := chunkHeaderInfo{
		Banner:     "Add-on GIS Dataset for MySQL of Vietnamese Provinces Database",
		CreatedAt:  "Mon, 10 Aug 2026 21:55:01 +0700",
		Repository: "https://github.com/thanglequoc/vietnamese-provinces-database",
	}
	blocks := [][]byte{
		[]byte("-- DATA for gis_provinces --\n"),
		[]byte("INSERT INTO gis_provinces(province_code, gis_server_id) VALUES ('01','x');\n"),
	}

	err := writeChunkedSQLFile(path, blocks, header)
	assert.NoError(t, err)

	// Exactly one part + manifest, header rendered on the part
	partPath := filepath.Join(tmpDir, "mysql_ImportData_gis_2026-08-10__21_55_01-part-01.sql")
	content, err := os.ReadFile(partPath)
	assert.NoError(t, err)
	contentStr := string(content)
	assert.Contains(t, contentStr, "/* === Add-on GIS Dataset for MySQL of Vietnamese Provinces Database === */")
	assert.Contains(t, contentStr, "/* Part 1 of 1 */")
	assert.Contains(t, contentStr, "/* Created at:  Mon, 10 Aug 2026 21:55:01 +0700 */")
	assert.Contains(t, contentStr, "/* Reference: https://github.com/thanglequoc/vietnamese-provinces-database */")
	assert.Contains(t, contentStr, "INSERT INTO gis_provinces(province_code, gis_server_id) VALUES ('01','x');")

	// No single file at path
	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err), "single file at path should NOT exist")

	// Manifest lists the single part
	manifestData, err := os.ReadFile(path + ".manifest")
	assert.NoError(t, err)
	assert.Equal(t, "mysql_ImportData_gis_2026-08-10__21_55_01-part-01.sql\n", string(manifestData))
}

func TestWriteChunkedSQLFile_MultipleParts(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "postgresql_ImportData_gis_2026-08-10__21_55_01.sql")

	originalMax := maxSQLGISChunkSize
	maxSQLGISChunkSize = 5000
	defer func() { maxSQLGISChunkSize = originalMax }()

	header := chunkHeaderInfo{
		Banner:     "Add-on GIS Dataset for PostgreSQL of Vietnamese Provinces Database",
		CreatedAt:  "Mon, 10 Aug 2026 21:55:01 +0700",
		Repository: "https://github.com/thanglequoc/vietnamese-provinces-database",
	}

	// 10 blocks of ~1.8 KB each; with a 5000-byte limit, blocks pack 2-per-chunk
	var blocks [][]byte
	var expected string
	for i := 0; i < 10; i++ {
		block := []byte(strings.Repeat("INSERT INTO gis_wards(ward_code) VALUES ('x');\n", 40))
		blocks = append(blocks, block)
		expected += string(block)
	}

	err := writeChunkedSQLFile(path, blocks, header)
	assert.NoError(t, err)

	// No single file at path
	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err), "single file at path should NOT exist")

	manifestData, err := os.ReadFile(path + ".manifest")
	assert.NoError(t, err)
	manifestLines := strings.Split(strings.TrimSpace(string(manifestData)), "\n")
	assert.True(t, len(manifestLines) >= 2, "expected at least 2 parts, got %d", len(manifestLines))

	var body strings.Builder
	blockCount := 0
	for i, name := range manifestLines {
		partData, err := os.ReadFile(filepath.Join(tmpDir, name))
		assert.NoError(t, err)
		assert.LessOrEqual(t, len(partData), maxSQLGISChunkSize, "part %s exceeds chunk limit", name)

		partStr := string(partData)
		assert.Contains(t, partStr, fmt.Sprintf("/* Part %d of %d */", i+1, len(manifestLines)))
		assert.Contains(t, partStr, "/* Reference: https://github.com/thanglequoc/vietnamese-provinces-database */")

		// Strip the header (everything up to the first blank line) and collect the body
		idx := strings.Index(partStr, "\n\n")
		assert.NotEqual(t, -1, idx, "part %s missing header/body separator", name)
		body.WriteString(partStr[idx+2:])
		blockCount += strings.Count(partStr[idx+2:], "INSERT INTO gis_wards(ward_code)")
	}

	assert.Equal(t, expected, body.String())
	assert.Equal(t, len(blocks), blockCount)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd dataset-generation-scripts && go test ./internal/dataset_writer/dataset_file_writer/ -run TestWriteChunkedSQLFile -v`
Expected: FAIL — compile error `undefined: writeChunkedSQLFile`.

- [ ] **Step 3: Write minimal implementation**

Create `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/gis_sql_chunk_writer.go`:

```go
package dataset_writer

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

// maxSQLGISChunkSize is the maximum size of a single GIS SQL chunk file.
// Matches the Elasticsearch maxNDJSONChunkSize (40 MB) and stays safely under
// GitHub's 50 MB file warning. var (not const) so tests can override it.
var maxSQLGISChunkSize = 40 * 1024 * 1024 // 40 MB

// chunkHeaderInfo carries the fixed text rendered into every chunk's leading
// SQL comment. The part/total numbers are interpolated per chunk at write time.
type chunkHeaderInfo struct {
	Banner     string
	CreatedAt  string
	Repository string
}

// writeChunkedSQLFile writes complete-SQL blocks as chunk files, each under
// maxSQLGISChunkSize. It always emits chunk files — never a single file — and
// writes a manifest at path + ".manifest" listing chunk filenames in order.
// Chunks are named <base>-part-NN<ext> (e.g. postgresql_ImportData_gis_<ts>-part-01.sql),
// matching the Elasticsearch naming convention.
//
// Every chunk starts with a self-describing SQL header comment containing the
// banner, "Part X of N", created-at timestamp, and repository link.
func writeChunkedSQLFile(path string, blocks [][]byte, header chunkHeaderInfo) error {
	if len(blocks) == 0 {
		return nil
	}

	// Split into chunks greedily at block boundaries (blocks are atomic).
	var chunks [][][]byte
	currentChunk := [][]byte{}
	currentSize := 0
	for _, b := range blocks {
		if currentSize+len(b) > maxSQLGISChunkSize && len(currentChunk) > 0 {
			chunks = append(chunks, currentChunk)
			currentChunk = [][]byte{}
			currentSize = 0
		}
		currentChunk = append(currentChunk, b)
		currentSize += len(b)
	}
	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	dir := filepathDir(path)
	base := filepathBase(path)
	ext := filepathExt(base)
	nameNoExt := base[:len(base)-len(ext)]

	log.Printf("📦 [GIS SQL] chunked: %d blocks → %d files (max %d MB each)",
		len(blocks), len(chunks), maxSQLGISChunkSize/1024/1024)

	var chunkNames []string
	for i, chunk := range chunks {
		chunkName := fmt.Sprintf("%s-part-%02d%s", nameNoExt, i+1, ext)
		chunkPath := fmt.Sprintf("%s/%s", dir, chunkName)
		chunkNames = append(chunkNames, chunkName)

		file, err := os.OpenFile(chunkPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("create chunk file %s: %w", chunkPath, err)
		}

		writer := bufio.NewWriter(file)
		headerLines := []string{
			fmt.Sprintf("/* === %s === */\n", header.Banner),
			fmt.Sprintf("/* Part %d of %d */\n", i+1, len(chunks)),
			fmt.Sprintf("/* Created at:  %s */\n", header.CreatedAt),
			fmt.Sprintf("/* Reference: %s */\n", header.Repository),
			"/* =============================================== */\n\n",
		}
		for _, line := range headerLines {
			if _, err := writer.WriteString(line); err != nil {
				file.Close()
				return fmt.Errorf("write header line: %w", err)
			}
		}
		for _, b := range chunk {
			if _, err := writer.Write(b); err != nil {
				file.Close()
				return fmt.Errorf("write block: %w", err)
			}
		}
		if err := writer.Flush(); err != nil {
			file.Close()
			return fmt.Errorf("flush chunk file %s: %w", chunkPath, err)
		}
		file.Close()

		chunkSize := 0
		for _, b := range chunk {
			chunkSize += len(b)
		}
		log.Printf("   %s: %.1f MB, %d blocks", chunkName, float64(chunkSize)/1024/1024, len(chunk))
	}

	manifestPath := path + ".manifest"
	manifestContent := stringsJoin(chunkNames, "\n") + "\n"
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		return fmt.Errorf("write manifest file: %w", err)
	}
	log.Printf("   Manifest: %s", filepathBase(manifestPath))

	return nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd dataset-generation-scripts && go test ./internal/dataset_writer/dataset_file_writer/ -run TestWriteChunkedSQLFile -v`
Expected: PASS for `TestWriteChunkedSQLFile_SinglePart` and `TestWriteChunkedSQLFile_MultipleParts`.

- [ ] **Step 5: Commit**

```bash
git add dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/gis_sql_chunk_writer.go dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/gis_sql_chunk_writer_test.go
git commit -m "feat: add chunked GIS SQL writer helper"
```

---
---

### Task 2: Refactor Postgres/MySQL GIS writer to chunked output

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/postgres_mysql_dataset_file_writer.go` — replace `WriteGISDataToFile` (lines ~140–245); remove `log` from imports.
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/postgres_mysql_dataset_file_writer_test.go` — update `readGeneratedGISFile` helper.

**Interfaces:**
- Consumes: `writeChunkedSQLFile(path string, blocks [][]byte, header chunkHeaderInfo) error` (Task 1); existing templates `insertProvinceGISTemplate`, `insertWardGISTemplate`, `insertWardGISValueTemplate`, `batchInsertItemSize`.
- Produces: refactored `func (w *PostgresMySQLDatasetFileWriter) WriteGISDataToFile(sapNhapProvincesGIS []*sapnhapmodels.SapNhapSiteGeoUnit, sapNhapWardsGIS []*sapnhapmodels.SapNhapSiteGeoUnit) error` — outputs chunk files + manifests under `./output/postgresql/gis` and `./output/mysql/gis`, no single `.sql`, no `.zip`.

- [ ] **Step 1: Update the test helper to read via manifest (red)**

In `postgres_mysql_dataset_file_writer_test.go`, replace the `readGeneratedGISFile` function (currently lines ~409–422) with:

```go
func readGeneratedGISFile(t *testing.T, rootDir, pattern string) string {
	t.Helper()

	// Find the manifest (pattern is e.g. "mysql_ImportData_gis_*.sql")
	manifestMatches, err := filepath.Glob(filepath.Join(rootDir, "output", "mysql", "gis", pattern+".manifest"))
	assert.NoError(t, err)
	if !assert.Len(t, manifestMatches, 1, "should have created one GIS manifest file") {
		return ""
	}

	manifestData, err := os.ReadFile(manifestMatches[0])
	assert.NoError(t, err)

	var sb strings.Builder
	for _, name := range strings.Split(strings.TrimSpace(string(manifestData)), "\n") {
		content, err := os.ReadFile(filepath.Join(rootDir, "output", "mysql", "gis", name))
		assert.NoError(t, err)
		sb.Write(content)
	}
	return sb.String()
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd dataset-generation-scripts && go test ./internal/dataset_writer/dataset_file_writer/ -run TestPostgresMySQLDatasetFileWriter_WriteGISDataToFile -v`
Expected: FAIL on `readGeneratedGISFile` (`should have created one GIS manifest file`) because the writer still emits a single `.sql` + `.zip` and no manifest.

- [ ] **Step 3: Refactor `WriteGISDataToFile`**

In `postgres_mysql_dataset_file_writer.go`, replace the entire `WriteGISDataToFile` function (from `func (w *PostgresMySQLDatasetFileWriter) WriteGISDataToFile(` through its closing `}`) with:

```go
func (w *PostgresMySQLDatasetFileWriter) WriteGISDataToFile(sapNhapProvincesGIS []*sapnhapmodels.SapNhapSiteGeoUnit, sapNhapWardsGIS []*sapnhapmodels.SapNhapSiteGeoUnit) error {
	fileTimeSuffix := getFileTimeSuffix()

	postgresGISDir := filepath.Join("./output/postgresql", "gis")
	mysqlGISDir := filepath.Join("./output/mysql", "gis")

	_ = os.MkdirAll(postgresGISDir, os.ModePerm)
	_ = os.MkdirAll(mysqlGISDir, os.ModePerm)

	postgresGISFilePath := filepath.Join(postgresGISDir, fmt.Sprintf("postgresql_ImportData_gis_%s.sql", fileTimeSuffix))
	mysqlGISFilePath := filepath.Join(mysqlGISDir, fmt.Sprintf("mysql_ImportData_gis_%s.sql", fileTimeSuffix))

	createdAt := time.Now().Format(time.RFC1123Z)
	postgresHeader := chunkHeaderInfo{
		Banner:     "Add-on GIS Dataset for PostgreSQL of Vietnamese Provinces Database",
		CreatedAt:  createdAt,
		Repository: "https://github.com/thanglequoc/vietnamese-provinces-database",
	}
	mysqlHeader := chunkHeaderInfo{
		Banner:     "Add-on GIS Dataset for MySQL of Vietnamese Provinces Database",
		CreatedAt:  createdAt,
		Repository: "https://github.com/thanglequoc/vietnamese-provinces-database",
	}

	var postgresBlocks [][]byte
	var mysqlBlocks [][]byte

	postgresBlocks = append(postgresBlocks, []byte("-- DATA for gis_provinces --\n"))
	mysqlBlocks = append(mysqlBlocks, []byte("-- DATA for gis_provinces --\n"))

	for _, p := range sapNhapProvincesGIS {
		vnProvinceCode := p.VNDSProvinceCode

		// Postgres - PostGIS use OGC (Open Geospatial Consortium) standard (lng - lat)
		postgresInsertLine := fmt.Sprintf(insertProvinceGISTemplate+"\n",
			vnProvinceCode, p.MaLK, p.DienTichKM2, p.BBoxWKT, p.GeomWKT)
		postgresBlocks = append(postgresBlocks, []byte(postgresInsertLine))

		// MySQL uses official EPSG standard (lat - lng)
		mysqlInsertLine := fmt.Sprintf(insertProvinceGISTemplate+"\n",
			vnProvinceCode, p.MaLK, p.DienTichKM2, p.BBoxWKTLatLng, p.GeomWKTLatLng)
		mysqlBlocks = append(mysqlBlocks, []byte(mysqlInsertLine))
	}

	postgresBlocks = append(postgresBlocks, []byte("-- ----------------------------------\n\n-- DATA for gis_wards --\n"))
	mysqlBlocks = append(mysqlBlocks, []byte("-- ----------------------------------\n\n-- DATA for gis_wards --\n"))

	counter := 0
	postgresBatch := strings.Builder{}
	mysqlBatch := strings.Builder{}
	for i, w := range sapNhapWardsGIS {
		if counter == 0 {
			postgresBatch.WriteString(insertWardGISTemplate + "\n")
			mysqlBatch.WriteString(insertWardGISTemplate + "\n")
		}

		vnWardCode := w.VNDSWardCode
		postgresInsertLine := fmt.Sprintf(insertWardGISValueTemplate+"\n",
			vnWardCode, w.MaLK, w.DienTichKM2, w.BBoxWKT, w.GeomWKT)
		postgresBatch.WriteString(postgresInsertLine)

		mysqlInsertLine := fmt.Sprintf(insertWardGISValueTemplate+"\n",
			vnWardCode, w.MaLK, w.DienTichKM2, w.BBoxWKTLatLng, w.GeomWKTLatLng)
		mysqlBatch.WriteString(mysqlInsertLine)

		counter++
		if counter == batchInsertItemSize || i == len(sapNhapWardsGIS)-1 {
			postgresBatch.WriteString(";\n\n")
			mysqlBatch.WriteString(";\n\n")
			postgresBlocks = append(postgresBlocks, []byte(postgresBatch.String()))
			mysqlBlocks = append(mysqlBlocks, []byte(mysqlBatch.String()))
			counter = 0
			postgresBatch.Reset()
			mysqlBatch.Reset()
		} else {
			postgresBatch.WriteString(",\n")
			mysqlBatch.WriteString(",\n")
		}
	}

	postgresBlocks = append(postgresBlocks, []byte("-- ----------------------------------\n\n-- END OF SCRIPT FILE --\n"))
	mysqlBlocks = append(mysqlBlocks, []byte("-- ----------------------------------\n\n-- END OF SCRIPT FILE --\n"))

	if err := writeChunkedSQLFile(postgresGISFilePath, postgresBlocks, postgresHeader); err != nil {
		return fmt.Errorf("write postgres GIS chunks: %w", err)
	}
	if err := writeChunkedSQLFile(mysqlGISFilePath, mysqlBlocks, mysqlHeader); err != nil {
		return fmt.Errorf("write mysql GIS chunks: %w", err)
	}
	return nil
}
```

Then update the import block of the same file to remove `log` (it is no longer used):

```go
import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	sapnhapmodels "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"

	"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
)
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd dataset-generation-scripts && go test ./internal/dataset_writer/dataset_file_writer/ -run TestPostgresMySQLDatasetFileWriter_WriteGISDataToFile -v`
Expected: PASS for both `..._MySQLWardBatchUsesCommaWithinBatch` and `..._MySQLWardBatchSplitsAtBatchSize`.

- [ ] **Step 5: Commit**

```bash
git add dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/postgres_mysql_dataset_file_writer.go dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/postgres_mysql_dataset_file_writer_test.go
git commit -m "feat: chunk postgres/mysql GIS SQL output with manifest"
```

---
---

### Task 3: Refactor MSSQL GIS writer to chunked output + add test

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/mssql_dataset_file_writer.go` — replace `WriteGISDataToFile` (lines ~132–196); remove `log` from imports.
- Create: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/mssql_dataset_file_writer_test.go`

**Interfaces:**
- Consumes: `writeChunkedSQLFile` (Task 1); existing templates `insertMssqlGISProvinceTemplate`, `insertMssqlGISWardTemplate`, `insertMssqlGISWardValueTemplate`, `batchInsertItemSize`.
- Produces: refactored `func (w *MssqlDatasetFileWriter) WriteGISDataToFile(...) error` writing chunk files + manifest under `./output/sqlserver/gis`, no single `.sql`, no `.zip`.

- [ ] **Step 1: Write the failing test**

Create `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/mssql_dataset_file_writer_test.go`:

```go
package dataset_writer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	sapnhapmodels "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
)

func TestMssqlDatasetFileWriter_WriteGISDataToFile_Chunked(t *testing.T) {
	tmpDir := t.TempDir()
	originalWD, err := os.Getwd()
	assert.NoError(t, err)

	err = os.Chdir(tmpDir)
	assert.NoError(t, err)
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	writer := &MssqlDatasetFileWriter{}
	provinces := []*sapnhapmodels.SapNhapSiteGeoUnit{
		{
			VNDSProvinceCode: "01",
			MaLK:             "tinh.1",
			DienTichKM2:      3359.84,
			BBoxWKT:          "POLYGON((105 20,106 20,106 21,105 21,105 20))",
			GeomWKT:          "MULTIPOLYGON(((105 20,106 20,106 21,105 21,105 20)))",
		},
	}
	wards := []*sapnhapmodels.SapNhapSiteGeoUnit{
		{
			VNDSWardCode: "00001",
			MaLK:         "xa.1",
			DienTichKM2:  5.23,
			BBoxWKT:      "POLYGON((105.8 21,105.9 21,105.9 21.1,105.8 21.1,105.8 21))",
			GeomWKT:      "POLYGON((105.8 21,105.9 21,105.9 21.1,105.8 21.1,105.8 21))",
		},
		{
			VNDSWardCode: "00002",
			MaLK:         "xa.2",
			DienTichKM2:  3.14,
			BBoxWKT:      "POLYGON((106 21,106.1 21,106.1 21.1,106 21.1,106 21))",
			GeomWKT:      "POLYGON((106 21,106.1 21,106.1 21.1,106 21.1,106 21))",
		},
	}

	err = writer.WriteGISDataToFile(provinces, wards)
	assert.NoError(t, err)

	content := readGeneratedMssqlGISFile(t, tmpDir)

	assert.Contains(t, content, "/* === Add-on GIS Dataset for Microsoft SQL Server of Vietnamese Provinces Database === */")
	assert.Contains(t, content, "/* Part 1 of 1 */")
	assert.Contains(t, content, "/* Reference: https://github.com/thanglequoc/vietnamese-provinces-database */")
	assert.Contains(t, content, "INSERT INTO gis_provinces(province_code, gis_server_id, area_km2, bbox, geom) VALUES")
	assert.Contains(t, content, "geometry::STGeomFromText('POLYGON((105 20,106 20,106 21,105 21,105 20))', 4326)")
	assert.Contains(t, content, "('00001','xa.1',5.230000")
	assert.Contains(t, content, "('00002','xa.2',3.140000")
	assert.Contains(t, content, ",\n")
	assert.Contains(t, content, ";\n\nGO\n\n")
	assert.Contains(t, content, "-- END OF SCRIPT FILE --")
}

func readGeneratedMssqlGISFile(t *testing.T, rootDir string) string {
	t.Helper()

	manifestMatches, err := filepath.Glob(filepath.Join(rootDir, "output", "sqlserver", "gis", "mssql_ImportData_gis_*.sql.manifest"))
	assert.NoError(t, err)
	if !assert.Len(t, manifestMatches, 1, "should have created one MSSQL GIS manifest file") {
		return ""
	}

	manifestData, err := os.ReadFile(manifestMatches[0])
	assert.NoError(t, err)

	var sb strings.Builder
	for _, name := range strings.Split(strings.TrimSpace(string(manifestData)), "\n") {
		content, err := os.ReadFile(filepath.Join(rootDir, "output", "sqlserver", "gis", name))
		assert.NoError(t, err)
		sb.Write(content)
	}
	return sb.String()
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd dataset-generation-scripts && go test ./internal/dataset_writer/dataset_file_writer/ -run TestMssqlDatasetFileWriter_WriteGISDataToFile_Chunked -v`
Expected: FAIL on `readGeneratedMssqlGISFile` (`should have created one MSSQL GIS manifest file`) because the writer still emits a single `.sql` + `.zip` and no manifest.

- [ ] **Step 3: Refactor `WriteGISDataToFile`**

In `mssql_dataset_file_writer.go`, replace the entire `WriteGISDataToFile` function (from `func (w *MssqlDatasetFileWriter) WriteGISDataToFile(` through its closing `}`) with:

```go
func (w *MssqlDatasetFileWriter) WriteGISDataToFile(sapNhapProvincesGIS []*sapnhapmodels.SapNhapSiteGeoUnit, sapNhapWardsGIS []*sapnhapmodels.SapNhapSiteGeoUnit) error {
	fileTimeSuffix := getFileTimeSuffix()

	gisOutputFolderPath := "./output/sqlserver/gis"
	if err := os.MkdirAll(gisOutputFolderPath, os.ModePerm); err != nil {
		return fmt.Errorf("create output folder %s: %w", gisOutputFolderPath, err)
	}

	mssqlGISFilePath := fmt.Sprintf(gisOutputFolderPath+"/mssql_ImportData_gis_%s.sql", fileTimeSuffix)

	header := chunkHeaderInfo{
		Banner:     "Add-on GIS Dataset for Microsoft SQL Server of Vietnamese Provinces Database",
		CreatedAt:  time.Now().Format(time.RFC1123Z),
		Repository: "https://github.com/thanglequoc/vietnamese-provinces-database",
	}

	var blocks [][]byte
	blocks = append(blocks, []byte("-- DATA for gis_provinces --\n"))
	for _, p := range sapNhapProvincesGIS {
		vnProvinceCode := p.VNDSProvinceCode
		mssqlInsertLine := fmt.Sprintf(insertMssqlGISProvinceTemplate+"\n",
			vnProvinceCode, p.MaLK, p.DienTichKM2, p.BBoxWKT, p.GeomWKT)
		blocks = append(blocks, []byte(mssqlInsertLine))
	}
	blocks = append(blocks, []byte("-- ----------------------------------\n\n-- DATA for gis_wards --\n"))

	counter := 0
	batch := strings.Builder{}
	for i, w := range sapNhapWardsGIS {
		if counter == 0 {
			batch.WriteString(insertMssqlGISWardTemplate + "\n")
		}

		vnWardCode := w.VNDSWardCode
		mssqlInsertLine := fmt.Sprintf(insertMssqlGISWardValueTemplate,
			vnWardCode, w.MaLK, w.DienTichKM2, w.BBoxWKT, w.GeomWKT)
		batch.WriteString(mssqlInsertLine)

		counter++
		if counter == batchInsertItemSize || i == len(sapNhapWardsGIS)-1 {
			batch.WriteString(";\n\nGO\n\n")
			blocks = append(blocks, []byte(batch.String()))
			counter = 0
			batch.Reset()
		} else {
			batch.WriteString(",\n")
		}
	}

	blocks = append(blocks, []byte("-- ----------------------------------\n\n-- END OF SCRIPT FILE --\n"))

	if err := writeChunkedSQLFile(mssqlGISFilePath, blocks, header); err != nil {
		return fmt.Errorf("write mssql GIS chunks: %w", err)
	}
	return nil
}
```

Then update the import block of the same file to remove `log` (no longer used):

```go
import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	sapnhapmodels "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
	"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
)
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd dataset-generation-scripts && go test ./internal/dataset_writer/dataset_file_writer/ -run TestMssqlDatasetFileWriter_WriteGISDataToFile_Chunked -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/mssql_dataset_file_writer.go dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/mssql_dataset_file_writer_test.go
git commit -m "feat: chunk mssql GIS SQL output with manifest"
```

---
---

### Task 4: Remove `zipFile` and its tests

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/dataset_file_writer.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/dataset_file_writer_test.go`

**Interfaces:**
- Consumes: nothing new.
- Produces: no changes to public API; `zipFile` and its three tests are gone.

- [ ] **Step 1: Delete `zipFile` and clean imports in `dataset_file_writer.go`**

Remove the `zipFile` function (currently lines ~66–118, the comment `// zipFile compresses a single file...` through its closing `}`).

Replace the import block with:

```go
import (
	"strconv"
	"strings"
	"time"

	sapnhapmodels "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
	"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
)
```

(Dropped: `archive/zip`, `compress/flate`, `fmt`, `io`, `log`, `os` — all only used by `zipFile`.)

- [ ] **Step 2: Delete the `zipFile` tests in `dataset_file_writer_test.go`**

Remove these three test functions:
- `TestZipFile_CreatesValidArchive`
- `TestZipFile_SourceFileNotFound`
- `TestZipFile_EmptyFile`

Replace the import block with:

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)
```

(Dropped: `archive/zip`, `io` — only used by the removed tests.)

- [ ] **Step 3: Run tests to verify the package still compiles and passes**

Run: `cd dataset-generation-scripts && go test ./internal/dataset_writer/dataset_file_writer/ -v`
Expected: PASS (all remaining tests, including the new chunk tests).

- [ ] **Step 4: Verify no other references to `zipFile`**

Run: `cd dataset-generation-scripts && rg -n "zipFile" .`
Expected: no matches.

- [ ] **Step 5: Commit**

```bash
git add dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/dataset_file_writer.go dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/dataset_file_writer_test.go
git commit -m "refactor: remove unused zipFile helper and its tests"
```

---
---

### Task 5: Fix repository URL casing `ThangLeQuoc` → `thanglequoc`

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/postgres_mysql_dataset_file_writer.go` (line ~62, non-GIS `WriteToFile` header)
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/mssql_dataset_file_writer.go` (line ~56, non-GIS `WriteToFile` header)
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/oracle_dataset_file_writer.go` (line ~42)
- Modify: `dataset-generation-scripts/README.md` (line ~31)

**Interfaces:** none.

- [ ] **Step 1: Replace occurrences**

In each of the four files, replace every `https://github.com/ThangLeQuoc/vietnamese-provinces-database` with `https://github.com/thanglequoc/vietnamese-provinces-database`, and in `dataset-generation-scripts/README.md` replace `git@github.com:ThangLeQuoc/vietnamese-provinces-database.git` with `git@github.com:thanglequoc/vietnamese-provinces-database.git`.

Verify with:
Run: `cd dataset-generation-scripts && rg -n "ThangLeQuoc" ..`
Expected: no matches.

- [ ] **Step 2: Run the full test suite**

Run: `cd dataset-generation-scripts && go test ./...`
Expected: PASS (or note that integration-test packages that need Docker may skip/fail as documented).

- [ ] **Step 3: Commit**

```bash
git add dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/postgres_mysql_dataset_file_writer.go dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/mssql_dataset_file_writer.go dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/oracle_dataset_file_writer.go dataset-generation-scripts/README.md
git commit -m "chore: use lowercase thanglequoc in repository references"
```

---
---

### Task 6: Update documentation

**Files:**
- Modify: `docs/gis/gis_readme.md`
- Modify: `docs/gis/gis_readme_vi.md`
- Modify: `dataset-generation-scripts/README.md` (output-structure listing, lines ~90–92)

**Interfaces:** none.

- [ ] **Step 1: Update `docs/gis/gis_readme.md` PostgreSQL section (Step 3, lines ~183–192)**

Replace:

````md
#### Step 3: Import GIS Data

Unzip the Postg GIS data archive at [postgresql/gis](../../postgresql/gis/)
Or download the raw GIS dataset file directly from [vn-province bucket][gis_dataset_postgresql_bucket_url]

Execute the data import script to populate boundaries:

```bash
psql -U <username> -d <database_name> -f postgresql/gis/postgresql_ImportData_gis_2026-06-20__12_32_01.sql
```
````

with:

````md
#### Step 3: Import GIS Data

Download the chunked GIS dataset from [postgresql/gis](../../postgresql/gis/). The
data is split into parts smaller than 40 MB, named
`postgresql_ImportData_gis_<timestamp>-part-NN.sql`. Import every part **in
order**, following the list in the accompanying `.manifest` file. For example:

```bash
psql -U <username> -d <database_name> -f postgresql/gis/postgresql_ImportData_gis_2026-06-20__12_32_01-part-01.sql
psql -U <username> -d <database_name> -f postgresql/gis/postgresql_ImportData_gis_2026-06-20__12_32_01-part-02.sql
# ... one psql invocation per part, in manifest order
```
````

- [ ] **Step 2: Update `docs/gis/gis_readme.md` MySQL section (Step 2, lines ~211–220)**

Replace the "Unzip the MySQL GIS data archive ... raw GIS dataset file ..." paragraph and the single-file `mysql` command block with the chunked equivalent:

````md
#### Step 2: Import GIS Data

Download the chunked GIS dataset from [mysql/gis](../../mysql/gis/). The data is
split into parts smaller than 40 MB, named
`mysql_ImportData_gis_<timestamp>-part-NN.sql`. Import every part **in order**,
following the list in the accompanying `.manifest` file. For example:

```bash
mysql -u <username> -p <database_name> < mysql/gis/mysql_ImportData_gis_2026-06-20__12_32_01-part-01.sql
mysql -u <username> -p <database_name> < mysql/gis/mysql_ImportData_gis_2026-06-20__12_32_01-part-02.sql
# ... one mysql invocation per part, in manifest order
```
````

- [ ] **Step 3: Update `docs/gis/gis_readme.md` SQL Server section (Step 2, lines ~247–256)**

Replace the "Unzip the SQL Server GIS data archive ... raw GIS dataset file ..." paragraph and the single-file `sqlcmd` command block with the chunked equivalent:

````md
#### Step 2: Import GIS Data

Download the chunked GIS dataset from [sqlserver/gis](../../sqlserver/gis/). The
data is split into parts smaller than 40 MB, named
`mssql_ImportData_gis_<timestamp>-part-NN.sql`. Import every part **in order**,
following the list in the accompanying `.manifest` file. For example:

```cmd
sqlcmd -S <server_name> -d <database_name> -U <username> -P <password> -i sqlserver/gis/mssql_ImportData_gis_2026-06-20__12_32_02-part-01.sql
sqlcmd -S <server_name> -d <database_name> -U <username> -P <password> -i sqlserver/gis/mssql_ImportData_gis_2026-06-20__12_32_02-part-02.sql
# ... one sqlcmd invocation per part, in manifest order
```
````

- [ ] **Step 4: Update `docs/gis/gis_readme.md` bucket URL definitions (lines ~497–499)**

Replace:

````md
[gis_dataset_postgresql_bucket_url]: https://vn-provinces-ds.thanglequoc.xyz/v4.1.0/GISDataSet/postgresql_ImportData_gis_2026-07-12__19_50_50.sql
[gis_dataset_mysql_bucket_url]: https://vn-provinces-ds.thanglequoc.xyz/v4.1.0/GISDataSet/mysql_ImportData_gis_2026-07-12__19_50_50.sql
[gis_dataset_sqlserver_bucket_url]: https://vn-provinces-ds.thanglequoc.xyz/v4.1.0/GISDataSet/mssql_ImportData_gis_2026-07-12__19_50_51.sql
````

with (pointing to the manifests; the bucket hosts the `.manifest` and its `-part-NN.sql` files):

````md
[gis_dataset_postgresql_bucket_url]: https://vn-provinces-ds.thanglequoc.xyz/v4.1.0/GISDataSet/postgresql_ImportData_gis_2026-07-12__19_50_50.sql.manifest
[gis_dataset_mysql_bucket_url]: https://vn-provinces-ds.thanglequoc.xyz/v4.1.0/GISDataSet/mysql_ImportData_gis_2026-07-12__19_50_50.sql.manifest
[gis_dataset_sqlserver_bucket_url]: https://vn-provinces-ds.thanglequoc.xyz/v4.1.0/GISDataSet/mssql_ImportData_gis_2026-07-12__19_50_51.sql.manifest
````

- [ ] **Step 5: Mirror Steps 1–4 in `docs/gis/gis_readme_vi.md`**

Apply the identical replacements to `docs/gis/gis_readme_vi.md` (same section structure at lines ~190, ~219, ~256, ~492–494; keep its existing `v4.0.0` version prefix in the bucket URLs).

- [ ] **Step 6: Update `dataset-generation-scripts/README.md` output-structure listing (lines ~90–92)**

Replace:

````md
└── gis/                                                     # (only if INCLUDE_GIS=true)
    ├── *_ImportData_gis_*.sql                               # GIS SQL imports per engine
    ├── *_ImportData_gis_*.sql.zip                           # Compressed versions
````

with:

````md
└── gis/                                                     # (only if INCLUDE_GIS=true)
    ├── *_ImportData_gis_*-part-*.sql                        # GIS SQL import chunks per engine (each < 40 MB)
    ├── *_ImportData_gis_*.sql.manifest                      # Ordered chunk list per engine
````

- [ ] **Step 7: Commit**

```bash
git add docs/gis/gis_readme.md docs/gis/gis_readme_vi.md dataset-generation-scripts/README.md
git commit -m "docs: document chunked GIS SQL import"
```

---
---

### Task 7: Regenerate dataset and publish chunked artifacts

**Files:**
- Delete: `postgresql/gis/postgresql_ImportData_gis_2026-08-08__21_32_14.sql.zip`
- Delete: `mysql/gis/mysql_ImportData_gis_2026-08-08__21_32_14.sql.zip`
- Delete: `sqlserver/gis/mssql_ImportData_gis_2026-08-08__21_33_02.sql.zip`
- Add: regenerated `*-part-NN.sql` + `.manifest` files under `postgresql/gis/`, `mysql/gis/`, `sqlserver/gis/`

**Prerequisites:** Docker with the Postgres/PostGIS container, and network access to `sapnhap.bando.com.vn` (the live GIS fetch). If the GIS fetch fails, stop and report; do not commit partial data.

- [ ] **Step 1: Ensure the Postgres/PostGIS container is running**

Run (from `dataset-generation-scripts/`):
```bash
docker compose -f docker/docker-compose.yaml up -d
```

- [ ] **Step 2: Regenerate the full dataset**

Run:
```bash
cd dataset-generation-scripts && go run main.go
```
Expected: completes with no fatal errors; log output teed to `output/generation-log.txt` (check it if console output is truncated).

- [ ] **Step 3: Verify the generated GIS output**

Run:
```bash
for d in postgresql mysql sqlserver; do
  echo "== $d =="
  ls -la output/$d/gis/ | head -20
done
```

Verify for each of `output/postgresql/gis/`, `output/mysql/gis/`, `output/sqlserver/gis/`:
- Exactly one `*.sql.manifest` listing `*-part-NN.sql` files.
- No `.sql.zip` and no bare `*_ImportData_gis_*.sql` single file.
- Every part is ≤ 40 MB (use `ls -l` and confirm no file > 41943040 bytes).

- [ ] **Step 4: Remove the old published zips**

```bash
git rm postgresql/gis/postgresql_ImportData_gis_2026-08-08__21_32_14.sql.zip \
       mysql/gis/mysql_ImportData_gis_2026-08-08__21_32_14.sql.zip \
       sqlserver/gis/mssql_ImportData_gis_2026-08-08__21_33_02.sql.zip
```

- [ ] **Step 5: Copy the generated chunks into the top-level folders**

```bash
cp dataset-generation-scripts/output/postgresql/gis/*-part-*.sql dataset-generation-scripts/output/postgresql/gis/*.manifest postgresql/gis/
cp dataset-generation-scripts/output/mysql/gis/*-part-*.sql dataset-generation-scripts/output/mysql/gis/*.manifest mysql/gis/
cp dataset-generation-scripts/output/sqlserver/gis/*-part-*.sql dataset-generation-scripts/output/sqlserver/gis/*.manifest sqlserver/gis/
```

- [ ] **Step 6: Verify committed artifacts are all under 40 MB**

Run:
```bash
find postgresql/gis mysql/gis sqlserver/gis -type f -name '*.sql' -o -name '*.manifest' | xargs ls -l | awk '$5 > 41943040 {print "TOO LARGE: " $0}'
```
Expected: no output (no file exceeds 40 MB).

- [ ] **Step 7: Run the test suite once more**

Run: `cd dataset-generation-scripts && go test ./...`
Expected: PASS (or documented Docker-dependent skips).

- [ ] **Step 8: Commit**

```bash
git add postgresql/gis mysql/gis sqlserver/gis
git commit -m "feat: publish chunked GIS SQL datasets, remove old zips"
```

---
---

## Self-Review Notes

- **Spec coverage:** objective (chunked < 40 MB, never single-file, no zips) → Tasks 1–4; per-chunk header (timestamp/repo/part) → Task 1 (helper) + Tasks 2–3 (banner/createdAt); URL casing → Tasks 2/3 (headers use lowercase) + Task 5 (remaining occurrences); docs → Task 6; published artifacts → Task 7.
- **Type consistency:** `writeChunkedSQLFile(path string, blocks [][]byte, header chunkHeaderInfo) error`, `chunkHeaderInfo{Banner, CreatedAt, Repository string}`, and `maxSQLGISChunkSize` are defined in Task 1 and consumed identically in Tasks 2–3. Manifest naming (`path + ".manifest"`, one filename per line) is consistent with the ES/Mongo convention. Test helpers `readGeneratedGISFile` / `readGeneratedMssqlGISFile` use the same manifest-glob pattern.
- **Placeholder scan:** all steps contain concrete code, commands, and expected outcomes.

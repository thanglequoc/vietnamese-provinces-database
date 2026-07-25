# GIS SQL Zip Archive Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Automatically archive PostgreSQL, MySQL, and MSSQL GIS SQL files as individual `.zip` archives during dataset generation.

**Architecture:** A single shared `zipFile` helper in `dataset_file_writer.go` is called after each GIS SQL file is flushed — once for postgres, once for mysql, once for mssql. Mirrors the existing `archiveGeoJSONDirectory` pattern from `geojson_file_writer.go`.

**Tech Stack:** Go 1.24.0, `archive/zip`, `compress/flate`, `testify/assert`

## Global Constraints

- Individual `.zip` per SQL file (not a combined archive)
- Raw `.sql` files are retained alongside the `.zip` archives
- Compression uses `flate.BestCompression` (matching existing GeoJSON archiver)
- Zip failures are non-fatal: logged via `log.Println`, error returned but discarded at call site
- Archive file name is `{originalPath}.zip` (e.g., `postgresql_ImportData_gis_2026-07-25__18_03_14.sql.zip`)
- Follows existing test patterns: `t.TempDir()`, table-driven, `testify/assert`

---

### Task 1: Write tests for `zipFile` function

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/dataset_file_writer_test.go`

**Interfaces:**
- Produces: `func zipFile(sourcePath string) error` (defined in Task 2, tested here)

- [ ] **Step 1: Add three test cases for `zipFile` in `dataset_file_writer_test.go`**

Append the following test functions at the end of `dataset_file_writer_test.go`:

```go
func TestZipFile_CreatesValidArchive(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a source SQL file with known content
	sourcePath := filepath.Join(tmpDir, "test_ImportData_gis.sql")
	content := []byte("INSERT INTO gis_provinces(province_code, gis_server_id) VALUES ('01','prov.1');\n")
	err := os.WriteFile(sourcePath, content, 0644)
	assert.NoError(t, err)

	// Run zipFile
	err = zipFile(sourcePath)
	assert.NoError(t, err)

	// Verify archive exists
	archivePath := sourcePath + ".zip"
	assert.FileExists(t, archivePath)

	// Open archive and verify contents
	archive, err := zip.OpenReader(archivePath)
	assert.NoError(t, err)
	defer archive.Close()

	assert.Len(t, archive.File, 1, "archive should contain exactly one file")

	// Verify the zipped file name matches the source file name
	assert.Equal(t, filepath.Base(sourcePath), archive.File[0].Name)

	// Verify content
	rc, err := archive.File[0].Open()
	assert.NoError(t, err)
	defer rc.Close()

	extracted, err := io.ReadAll(rc)
	assert.NoError(t, err)
	assert.Equal(t, content, extracted)
}

func TestZipFile_SourceFileNotFound(t *testing.T) {
	err := zipFile("/nonexistent/path/to/file.sql")
	assert.Error(t, err)
}

func TestZipFile_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()

	sourcePath := filepath.Join(tmpDir, "empty.sql")
	err := os.WriteFile(sourcePath, []byte{}, 0644)
	assert.NoError(t, err)

	err = zipFile(sourcePath)
	assert.NoError(t, err)

	archivePath := sourcePath + ".zip"
	assert.FileExists(t, archivePath)

	archive, err := zip.OpenReader(archivePath)
	assert.NoError(t, err)
	defer archive.Close()

	assert.Len(t, archive.File, 1)
	rc, err := archive.File[0].Open()
	assert.NoError(t, err)
	defer rc.Close()

	extracted, err := io.ReadAll(rc)
	assert.NoError(t, err)
	assert.Empty(t, extracted)
}
```

Add the required imports to the top of `dataset_file_writer_test.go`. The file already imports `os`, `path/filepath`, `strings`, `testing`, `testify/assert`, and model packages. Add `archive/zip` and `io`:

```go
import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	sapnhapmodels "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
	vn_provinces_tmp_model "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
)
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd dataset-generation-scripts && go test -v ./internal/dataset_writer/dataset_file_writer/ -run TestZipFile`
Expected: FAIL — `undefined: zipFile`

- [ ] **Step 3: Commit**

```bash
cd dataset-generation-scripts && git add internal/dataset_writer/dataset_file_writer/dataset_file_writer_test.go && git commit -m "test: add failing tests for zipFile helper"
```

---

### Task 2: Implement `zipFile` helper

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/dataset_file_writer.go`

**Interfaces:**
- Produces: `func zipFile(sourcePath string) error`

- [ ] **Step 1: Add `zipFile` function to `dataset_file_writer.go`**

Replace the import block in `dataset_file_writer.go`:

```go
import (
	"archive/zip"
	"compress/flate"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	sapnhapmodels "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
	"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
)
```

Add the `zipFile` function at the end of the file (after `parseEuropeanFloat`):

```go
// zipFile compresses a single file to <sourcePath>.zip using best compression.
// On failure, logs a warning and returns the error. The caller may discard
// the error if zip failure should be non-fatal.
func zipFile(sourcePath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		log.Printf("[WARN] Unable to open source file for zip archive %s: %v", sourcePath, err)
		return fmt.Errorf("open source file %s for zipping: %w", sourcePath, err)
	}
	defer source.Close()

	archivePath := sourcePath + ".zip"
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		log.Printf("[WARN] Unable to create zip archive %s: %v", archivePath, err)
		return fmt.Errorf("create zip archive %s: %w", archivePath, err)
	}
	defer archiveFile.Close()

	zipWriter := zip.NewWriter(archiveFile)
	defer zipWriter.Close()

	zipWriter.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, flate.BestCompression)
	})

	sourceInfo, err := source.Stat()
	if err != nil {
		log.Printf("[WARN] Unable to stat source file %s: %v", sourcePath, err)
		return fmt.Errorf("stat source file %s: %w", sourcePath, err)
	}

	header, err := zip.FileInfoHeader(sourceInfo)
	if err != nil {
		log.Printf("[WARN] Unable to create zip header for %s: %v", sourcePath, err)
		return fmt.Errorf("create zip header for %s: %w", sourcePath, err)
	}
	header.Name = sourceInfo.Name()
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		log.Printf("[WARN] Unable to create zip entry for %s: %v", sourcePath, err)
		return fmt.Errorf("create zip entry for %s: %w", sourcePath, err)
	}

	if _, err := io.Copy(writer, source); err != nil {
		log.Printf("[WARN] Unable to write content to zip archive %s: %v", archivePath, err)
		return fmt.Errorf("copy source content into zip archive %s: %w", archivePath, err)
	}

	return nil
}
```

- [ ] **Step 2: Run the tests to verify they pass**

Run: `cd dataset-generation-scripts && go test -v ./internal/dataset_writer/dataset_file_writer/ -run TestZipFile`
Expected: PASS — all three `TestZipFile` tests pass

- [ ] **Step 3: Commit**

```bash
cd dataset-generation-scripts && git add internal/dataset_writer/dataset_file_writer/dataset_file_writer.go && git commit -m "feat: add zipFile helper for single-file zip archiving"
```

---

### Task 3: Add zip calls in `PostgresMySQLDatasetFileWriter.WriteGISDataToFile`

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/postgres_mysql_dataset_file_writer.go`

**Interfaces:**
- Consumes: `func zipFile(sourcePath string) error` (from Task 2)
- No signature change to `WriteGISDataToFile`

- [ ] **Step 1: Add `zipFile` calls after each flush**

In `postgres_mysql_dataset_file_writer.go`, the function `WriteGISDataToFile` ends with:

```go
	postgresScriptDataWriter.Flush()
	mysqlScriptDataWriter.Flush()
```

Replace with:

```go
	postgresScriptDataWriter.Flush()
	_ = zipFile(postgresGISFilePath)

	mysqlScriptDataWriter.Flush()
	_ = zipFile(mysqlGISFilePath)
```

- [ ] **Step 2: Run the existing GIS writer tests to verify no regressions**

Run: `cd dataset-generation-scripts && go test -v ./internal/dataset_writer/dataset_file_writer/ -run TestPostgresMySQLDatasetFileWriter_WriteGISDataToFile`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
cd dataset-generation-scripts && git add internal/dataset_writer/dataset_file_writer/postgres_mysql_dataset_file_writer.go && git commit -m "feat: zip postgres and mysql GIS SQL files after generation"
```

---

### Task 4: Add zip call in `MssqlDatasetFileWriter.WriteGISDataToFile`

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/mssql_dataset_file_writer.go`

**Interfaces:**
- Consumes: `func zipFile(sourcePath string) error` (from Task 2)
- No signature change to `WriteGISDataToFile`

- [ ] **Step 1: Add `zipFile` call after mssql flush**

In `mssql_dataset_file_writer.go`, the function `WriteGISDataToFile` has `defer mssqlGISFile.Close()` at line 143. Add the zip call after the flush at line 187. The flush is followed by `return nil` at line 189.

Replace:

```go
	mssqlScriptDataWriter.Flush()

	return nil
```

With:

```go
	mssqlScriptDataWriter.Flush()
	_ = zipFile(mssqlGISFilePath)

	return nil
```

- [ ] **Step 2: Run the test suite for the writer package**

Run: `cd dataset-generation-scripts && go test -v ./internal/dataset_writer/...`
Expected: All tests PASS

- [ ] **Step 3: Commit**

```bash
cd dataset-generation-scripts && git add internal/dataset_writer/dataset_file_writer/mssql_dataset_file_writer.go && git commit -m "feat: zip mssql GIS SQL file after generation"
```

---

### Task 5: Integration verification

**Files:**
- None (verification only)

- [ ] **Step 1: Run the full generation pipeline**

Run: `cd dataset-generation-scripts && go run main.go`
Expected: Generation completes with success messages including GIS generation

- [ ] **Step 2: Verify the output directory**

```bash
ls -la dataset-generation-scripts/output/gis/
```

Expected: Should see three `.sql.zip` files alongside the `.sql` files:
- `postgresql_ImportData_gis_{timestamp}.sql`
- `postgresql_ImportData_gis_{timestamp}.sql.zip`
- `mysql_ImportData_gis_{timestamp}.sql`
- `mysql_ImportData_gis_{timestamp}.sql.zip`
- `mssql_ImportData_gis_{timestamp}.sql`
- `mssql_ImportData_gis_{timestamp}.sql.zip`

- [ ] **Step 3: Verify one archive is valid**

```bash
unzip -l dataset-generation-scripts/output/gis/postgresql_ImportData_gis_*.sql.zip
```

Expected: Lists exactly one `.sql` file with non-zero compressed size

- [ ] **Step 4: Commit (if any generated output needs tracking)**

No changes to commit — verification only.

---
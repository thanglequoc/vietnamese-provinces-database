# Per-Dataset READMEs and Deterministic GIS Filenames — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Generate a self-describing `README.md` from every dataset writer (bold timestamp + file sizes + data structure + sample queries covering regular and GIS), and remove datetime suffixes from all GIS/mongodb/redis output files so published folders never accumulate duplicate timestamp variants.

**Architecture:** Add a shared `writeDatasetReadme` helper (boilerplate: title, bold timestamp, Files-with-sizes) in `dataset_file_writer.go`; each writer passes its dataset-specific sections. Strip `getFileTimeSuffix()` from postgres/mysql/mssql GIS chunk bases, mongodb GIS + base files, and redis. Remove the gis-subfolder README writers (mongodb gis, geojson). Update the copy script to prune old timestamped variants, then docs.

**Tech Stack:** Go 1.24, stdlib `encoding/json`/`os`/`path/filepath`/`strings`/`time`, Testify.

## Global Constraints

- Module root: `dataset-generation-scripts/`; tests run from there: `go test ./internal/dataset_writer/dataset_file_writer/... -v`.
- Final deterministic output filenames:
  - `postgresql_ImportData_gis-part-NN.sql` + `postgresql_ImportData_gis.sql.manifest`
  - `mysql_ImportData_gis-part-NN.sql` + `mysql_ImportData_gis.sql.manifest`
  - `mssql_ImportData_gis-part-NN.sql` + `mssql_ImportData_gis.sql.manifest`
  - `mongo_data_vn_province_gis.json`, `mongo_data_vn_ward_gis.json`, `mongo_data_vn_ward_gis_part_NN.json` + `.json.manifest`
  - `administrative_units.json`, `administrative_regions.json`, `mongo_data_vn_unit.json` (mongodb base)
  - `redis_vn_provinces_dataset.redis` (redis base)
- Every dataset output folder gets a `README.md`; gis-subfolder READMEs are removed (no `mongodb/gis/README.md`, no `json/geojson/README.md`).
- The `writeDatasetReadme` helper renders `**Generated at: <RFC1123Z>**` (bold) as the timestamp line.
- No commit of generated artifacts in source tasks; only source/tests/docs.

---

### Task 1: Shared README helper

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/dataset_file_writer.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/dataset_file_writer_test.go`

**Interfaces:**
- Produces:
  - `type DatasetReadmeFile struct { Name, Description string }`
  - `writeDatasetReadme(outputFolderPath, title, intro string, files []DatasetReadmeFile, sections []string) error`
  - `renderReadmeFilesSection(outputFolderPath string, files []DatasetReadmeFile) []string`
  - `formatFileSize(size int64) string` (moved here from `json_file_writer.go`)

- [ ] **Step 1: Write the failing test**

Append to `dataset_file_writer_test.go`:

```go
func TestWriteDatasetReadme(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "data.sql"), []byte("-- test\n"), 0644)
	assert.NoError(t, err)

	err = writeDatasetReadme(tmpDir,
		"PostgreSQL Dataset — Vietnamese Provinces Database",
		"Import script for the Vietnamese Provinces Database.",
		[]DatasetReadmeFile{{Name: "data.sql", Description: "Sample file"}},
		[]string{"## Data Structure\n\nsample", "## Sample Queries\n\nSELECT 1;"})
	assert.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "README.md"))
	assert.NoError(t, err)
	s := string(content)
	assert.Contains(t, s, "# PostgreSQL Dataset")
	assert.Contains(t, s, "**Generated at:")
	assert.Contains(t, s, "`data.sql` — Sample file")
	assert.Contains(t, s, "## Data Structure")
	assert.Contains(t, s, "## Sample Queries")

	// Missing files are skipped without error
	err = writeDatasetReadme(tmpDir, "T", "I", []DatasetReadmeFile{{Name: "nope.sql", Description: "missing"}}, nil)
	assert.NoError(t, err)
}
```

Update `dataset_file_writer_test.go` imports to add `"os"` and `"path/filepath"`:

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestWriteDatasetReadme -v`
Expected: FAIL — `writeDatasetReadme` undefined.

- [ ] **Step 3: Implement the helper**

Add to `dataset_file_writer.go` (update imports to add `"fmt"`, `"os"`, `"path/filepath"`):

```go
// DatasetReadmeFile describes one generated file for the dataset README.
type DatasetReadmeFile struct {
	Name        string
	Description string
}

// writeDatasetReadme writes README.md at outputFolderPath: a bold generation
// timestamp, a "Files" list with per-file sizes, then dataset-specific sections.
func writeDatasetReadme(outputFolderPath, title, intro string, files []DatasetReadmeFile, sections []string) error {
	lines := []string{
		"# " + title,
		"",
		fmt.Sprintf("**Generated at: %s**", time.Now().Format(time.RFC1123Z)),
		"",
		intro,
		"",
	}
	lines = append(lines, renderReadmeFilesSection(outputFolderPath, files)...)
	if len(sections) > 0 {
		lines = append(lines, "", strings.Join(sections, "\n"))
	}
	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(filepath.Join(outputFolderPath, "README.md"), []byte(content), 0644)
}

// renderReadmeFilesSection renders a "## Files" markdown block with per-file sizes.
// Files that do not exist yet are skipped.
func renderReadmeFilesSection(outputFolderPath string, files []DatasetReadmeFile) []string {
	lines := []string{"## Files", ""}
	for _, f := range files {
		filePath := filepath.Join(outputFolderPath, f.Name)
		info, err := os.Stat(filePath)
		if err != nil {
			continue
		}
		lines = append(lines, fmt.Sprintf("- `%s` — %s (%s)", f.Name, f.Description, formatFileSize(info.Size())))
	}
	return lines
}

func formatFileSize(size int64) string {
	const kb = 1024
	switch {
	case size >= kb*kb:
		return fmt.Sprintf("%.2f MB", float64(size)/(kb*kb))
	case size >= kb:
		return fmt.Sprintf("%.2f KB", float64(size)/kb)
	default:
		return fmt.Sprintf("%d B", size)
	}
}
```

- [ ] **Step 4: Remove the now-duplicated `formatFileSize` from `json_file_writer.go`**

Delete from `json_file_writer.go`:

```go
func formatFileSize(size int64) string {
	const kb = 1024
	switch {
	case size >= kb*kb:
		return fmt.Sprintf("%.2f MB", float64(size)/(kb*kb))
	case size >= kb:
		return fmt.Sprintf("%.2f KB", float64(size)/kb)
	default:
		return fmt.Sprintf("%d B", size)
	}
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestWriteDatasetReadme -v`
Expected: PASS. (Note: existing package tests must still compile — `json_file_writer.go` keeps its other functions.)

- [ ] **Step 6: Run the full package test suite**

Run: `go test ./internal/dataset_writer/dataset_file_writer/...`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/dataset_file_writer.go internal/dataset_writer/dataset_file_writer/dataset_file_writer_test.go internal/dataset_writer/dataset_file_writer/json_file_writer.go
git commit -m "feat: add shared dataset README writer helper"
```

---

### Task 2: Postgres/MySQL — deterministic GIS chunks + READMEs

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/postgres_mysql_dataset_file_writer.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/postgres_mysql_dataset_file_writer_test.go`

**Interfaces:**
- Consumes: `writeDatasetReadme`, `DatasetReadmeFile` from Task 1.

- [ ] **Step 1: Update the failing tests**

In `postgres_mysql_dataset_file_writer_test.go`:
- Line 358: change `"mysql_ImportData_gis_*.sql"` → `"mysql_ImportData_gis.sql"`.
- Line 401: change `"mysql_ImportData_gis_*.sql"` → `"mysql_ImportData_gis.sql"`.
- Line 412 comment stays; `readGeneratedGISFile` globs `pattern+".manifest"` = `mysql_ImportData_gis.sql.manifest`.

Append a README test:

```go
func TestPostgresMySQLDatasetFileWriter_WriteToFile_README(t *testing.T) {
	tmpDir := t.TempDir()
	writer := &PostgresMySQLDatasetFileWriter{
		OutputFilePath: filepath.Join(tmpDir, "postgres_ImportData_vn_units.sql"),
	}
	provinces := []vn_provinces_tmp_model.Province{{Code: "01", Name: "Hà Nội", AdministrativeUnitId: 1}}

	err := writer.WriteToFile(nil, nil, provinces, nil)
	assert.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "README.md"))
	assert.NoError(t, err)
	s := string(content)
	assert.Contains(t, s, "**Generated at:")
	assert.Contains(t, s, "postgres_ImportData_vn_units.sql")
	assert.Contains(t, s, "## Data Structure")
	assert.Contains(t, s, "## Sample Queries")
	assert.Contains(t, s, "gis/")
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run "TestPostgresMySQLDatasetFileWriter_WriteGISDataToFile|TestPostgresMySQLDatasetFileWriter_WriteToFile_README" -v`
Expected: FAIL — glob `mysql_ImportData_gis.sql.manifest` not found; `README.md` not found.

- [ ] **Step 3: Implement**

In `WriteGISDataToFile` (`postgres_mysql_dataset_file_writer.go`), delete line 141 `fileTimeSuffix := getFileTimeSuffix()` and change lines 149-150 to:

```go
	postgresGISFilePath := filepath.Join(postgresGISDir, "postgresql_ImportData_gis.sql")
	mysqlGISFilePath := filepath.Join(mysqlGISDir, "mysql_ImportData_gis.sql")
```

At the end of `WriteToFile` (replace `file.Close()\n\treturn nil` at lines 136-137 with):

```go
	dataWriter.Flush()
	file.Close()

	return writePostgresMySQLReadme(filepath.Dir(outputFilePath), outputFilePath)
}
```

Append these functions to the file:

```go
// writePostgresMySQLReadme writes the dataset README for the PostgreSQL or
// MySQL export. isPostgres is inferred from the output file path.
func writePostgresMySQLReadme(outputFolderPath, outputFilePath string) error {
	isPostgres := strings.Contains(filepath.Base(outputFilePath), "postgres")
	if isPostgres {
		return writeDatasetReadme(outputFolderPath,
			"PostgreSQL Dataset — Vietnamese Provinces Database",
			"Import script for the Vietnamese Provinces Database on PostgreSQL/PostGIS.",
			[]DatasetReadmeFile{
				{Name: "postgres_ImportData_vn_units.sql", Description: "INSERT statements for regions, units, provinces, and wards"},
			},
			postgresMySQLReadmeSections("PostgreSQL/PostGIS", true))
	}
	return writeDatasetReadme(outputFolderPath,
		"MySQL Dataset — Vietnamese Provinces Database",
		"Import script for the Vietnamese Provinces Database on MySQL/MariaDB.",
		[]DatasetReadmeFile{
			{Name: "mysql_ImportData_vn_units.sql", Description: "INSERT statements for regions, units, provinces, and wards"},
		},
		postgresMySQLReadmeSections("MySQL", false))
}

func postgresMySQLReadmeSections(engine string, isPostgres bool) []string {
	pointLiteral := "ST_SetSRID(ST_MakePoint(105.8542, 21.0285), 4326)"
	if !isPostgres {
		pointLiteral = "ST_GeomFromText('POINT(105.8542 21.0285)', 4326)"
	}
	return []string{
		"## Data Structure",
		"",
		"The import script populates: `administrative_regions` (8), `administrative_units` (8), `provinces` (34), and `wards` (3,321). Each province and ward carries postal code fields (`postal_code_prefix` / `postal_code`).",
		"",
		"GIS geometry (in `gis/`) populates `gis_provinces` and `gis_wards` with `bbox` and `geom` spatial columns.",
		"",
		"## Sample Queries",
		"",
		"```sql",
		"SELECT COUNT(*) FROM provinces;",
		"",
		"SELECT w.code, w.name FROM wards w WHERE w.province_code = '01' ORDER BY w.name;",
		"",
		"-- GIS: province containing a point",
		"SELECT p.code, p.name",
		"FROM provinces p",
		"JOIN gis_provinces g ON p.code = g.province_code",
		"WHERE ST_Within(" + pointLiteral + ", g.geom);",
		"```",
		"",
		"## GIS / GeoJSON",
		"",
		"The `gis/` subfolder contains chunked GIS import scripts (`" + engine + "` part files `*-part-NN.sql`) plus a `.manifest` file listing the chunks in order. Import each chunk in order after importing the base script.",
	}
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run "TestPostgresMySQLDatasetFileWriter" -v`
Expected: PASS.

- [ ] **Step 5: Run the full package test suite**

Run: `go test ./internal/dataset_writer/dataset_file_writer/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/postgres_mysql_dataset_file_writer.go internal/dataset_writer/dataset_file_writer/postgres_mysql_dataset_file_writer_test.go
git commit -m "feat: deterministic postgres/mysql GIS chunks and generated READMEs"
```

---

### Task 3: Mssql — deterministic GIS chunks + README

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/mssql_dataset_file_writer.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/mssql_dataset_file_writer_test.go`

- [ ] **Step 1: Update the failing tests**

In `mssql_dataset_file_writer_test.go` line 71, change:

```go
	manifestMatches, err := filepath.Glob(filepath.Join(rootDir, "output", "sqlserver", "gis", "mssql_ImportData_gis_*.sql.manifest"))
```

to:

```go
	manifestMatches, err := filepath.Glob(filepath.Join(rootDir, "output", "sqlserver", "gis", "mssql_ImportData_gis.sql.manifest"))
```

Append a README test:

```go
func TestMssqlDatasetFileWriter_WriteToFile_README(t *testing.T) {
	tmpDir := t.TempDir()
	writer := &MssqlDatasetFileWriter{
		OutputFilePath: filepath.Join(tmpDir, "mssql_ImportData_vn_units.sql"),
	}
	provinces := []vn_provinces_tmp_model.Province{{Code: "01", Name: "Hà Nội", AdministrativeUnitId: 1}}

	err := writer.WriteToFile(nil, nil, provinces, nil)
	assert.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "README.md"))
	assert.NoError(t, err)
	s := string(content)
	assert.Contains(t, s, "**Generated at:")
	assert.Contains(t, s, "mssql_ImportData_vn_units.sql")
	assert.Contains(t, s, "## Sample Queries")
	assert.Contains(t, s, "gis/")
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run "TestMssqlDatasetFileWriter" -v`
Expected: FAIL — manifest glob not found; README not found.

- [ ] **Step 3: Implement**

In `mssql_dataset_file_writer.go` `WriteGISDataToFile`, delete line 133 `fileTimeSuffix := getFileTimeSuffix()` and change line 140 to:

```go
	mssqlGISFilePath := gisOutputFolderPath + "/mssql_ImportData_gis.sql"
```

At the end of `WriteToFile`, replace `fileMsSql.Close()\n\n\treturn nil` with:

```go
	fileMsSql.Close()

	return writeMssqlReadme(filepath.Dir(outputFilePath))
}
```

Append:

```go
func writeMssqlReadme(outputFolderPath string) error {
	return writeDatasetReadme(outputFolderPath,
		"Microsoft SQL Server Dataset — Vietnamese Provinces Database",
		"Import script for the Vietnamese Provinces Database on Microsoft SQL Server.",
		[]DatasetReadmeFile{
			{Name: "mssql_ImportData_vn_units.sql", Description: "INSERT statements for regions, units, provinces, and wards"},
		},
		[]string{
			"## Data Structure",
			"",
			"The import script populates: `administrative_regions` (8), `administrative_units` (8), `provinces` (34), and `wards` (3,321). Each province and ward carries postal code fields (`postal_code_prefix` / `postal_code`).",
			"",
			"GIS geometry (in `gis/`) populates `gis_provinces` and `gis_wards` with `bbox` and `geom` geometry columns.",
			"",
			"## Sample Queries",
			"",
			"```sql",
			"SELECT COUNT(*) FROM provinces;",
			"",
			"SELECT w.code, w.name FROM wards w WHERE w.province_code = '01' ORDER BY w.name;",
			"",
			"-- GIS: province containing a point",
			"SELECT p.code, p.name",
			"FROM provinces p",
			"JOIN gis_provinces g ON p.code = g.province_code",
			"WHERE g.geom.STContains(geometry::STGeomFromText('POINT(105.8542 21.0285)', 4326)) = 1;",
			"```",
			"",
			"## GIS / GeoJSON",
			"",
			"The `gis/` subfolder contains chunked SQL Server GIS import scripts (`mssql_ImportData_gis-part-NN.sql`) plus a `.manifest` file listing the chunks in order.",
		})
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestMssqlDatasetFileWriter -v`
Expected: PASS.

- [ ] **Step 5: Run the full package test suite**

Run: `go test ./internal/dataset_writer/dataset_file_writer/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/mssql_dataset_file_writer.go internal/dataset_writer/dataset_file_writer/mssql_dataset_file_writer_test.go
git commit -m "feat: deterministic mssql GIS chunks and generated README"
```

---

### Task 4: Oracle — README

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/oracle_dataset_file_writer.go`
- Create: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/oracle_dataset_file_writer_test.go`

- [ ] **Step 1: Write the failing test**

Create `oracle_dataset_file_writer_test.go`:

```go
package dataset_writer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	vn_provinces_tmp_model "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
)

func TestOracleDatasetFileWriter_WriteToFile_README(t *testing.T) {
	tmpDir := t.TempDir()
	writer := &OracleDatasetFileWriter{
		OutputFilePath: filepath.Join(tmpDir, "oracle_ImportData_vn_units.sql"),
	}
	provinces := []vn_provinces_tmp_model.Province{{Code: "01", Name: "Hà Nội", AdministrativeUnitId: 1}}

	err := writer.WriteToFile(nil, nil, provinces, nil)
	assert.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "README.md"))
	assert.NoError(t, err)
	s := string(content)
	assert.Contains(t, s, "**Generated at:")
	assert.Contains(t, s, "oracle_ImportData_vn_units.sql")
	assert.Contains(t, s, "## Data Structure")
	assert.Contains(t, s, "## Sample Queries")
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestOracleDatasetFileWriter_WriteToFile_README -v`
Expected: FAIL — README.md not found.

- [ ] **Step 3: Implement**

In `oracle_dataset_file_writer.go`, replace the end of `WriteToFile`:

```go
	dataWriter.WriteString("-- END OF SCRIPT FILE --\n")
	dataWriter.Flush()
	file.Close()
	return nil
}
```

with:

```go
	dataWriter.WriteString("-- END OF SCRIPT FILE --\n")
	dataWriter.Flush()
	file.Close()

	return writeOracleReadme(filepath.Dir(outputFilePath))
}

func writeOracleReadme(outputFolderPath string) error {
	return writeDatasetReadme(outputFolderPath,
		"Oracle Dataset — Vietnamese Provinces Database",
		"Import script for the Vietnamese Provinces Database on Oracle.",
		[]DatasetReadmeFile{
			{Name: "oracle_ImportData_vn_units.sql", Description: "INSERT ALL statements for regions, units, provinces, and wards"},
		},
		[]string{
			"## Data Structure",
			"",
			"The import script populates: `administrative_regions` (8), `administrative_units` (8), `provinces` (34), and `wards` (3,321). Each province and ward carries postal code fields (`postal_code_prefix` / `postal_code`).",
			"",
			"## Sample Queries",
			"",
			"```sql",
			"SELECT COUNT(*) FROM provinces;",
			"",
			"SELECT w.code, w.name FROM wards w WHERE w.province_code = '01' ORDER BY w.name;",
			"```",
		})
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestOracleDatasetFileWriter_WriteToFile_README -v`
Expected: PASS.

- [ ] **Step 5: Run the full package test suite**

Run: `go test ./internal/dataset_writer/dataset_file_writer/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/oracle_dataset_file_writer.go internal/dataset_writer/dataset_file_writer/oracle_dataset_file_writer_test.go
git commit -m "feat: generate Oracle dataset README"
```

---

### Task 5: MongoDB — base ts removal, base README, GIS ts removal, drop gis README

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/mongodb_file_writer.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/mongodb_gis_file_writer.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/mongodb_gis_file_writer_test.go`

- [ ] **Step 1: Update the failing tests**

In `mongodb_gis_file_writer_test.go`, replace the README-exists block (lines 138-142):

```go
	// Verify README.md exists
	readmePath := filepath.Join(tmpDir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		t.Fatal("README.md not found")
	}
```

with:

```go
	// GIS subfolder no longer carries its own README
	readmePath := filepath.Join(tmpDir, "README.md")
	if _, err := os.Stat(readmePath); !os.IsNotExist(err) {
		t.Fatal("README.md should NOT be written in the mongodb gis folder")
	}
```

Append a base-writer README test to `mongodb_file_writer_test.go`:

```go
func TestMongoDBDatasetFileWriter_WriteToFile_README(t *testing.T) {
	tmpDir := t.TempDir()
	writer := MongoDBDatasetFileWriter{OutputFolderPath: tmpDir}
	provinces := []vn_provinces_tmp_model.Province{{Code: "01", Name: "Hà Nội", AdministrativeUnitId: 1}}

	err := writer.WriteToFile(nil, nil, provinces, nil)
	assert.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "README.md"))
	assert.NoError(t, err)
	s := string(content)
	assert.Contains(t, s, "**Generated at:")
	assert.Contains(t, s, "mongo_data_vn_unit.json")
	assert.Contains(t, s, "## Sample Queries")
	assert.Contains(t, s, "gis/")
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run "TestMongoDBDatasetFileWriter_WriteToFile_README|TestWriteMongoGISDataToFile" -v`
Expected: FAIL — base README missing; gis README still present.

- [ ] **Step 3: Implement — mongodb base writer**

In `mongodb_file_writer.go`:
- Delete line 23 `fileTimeSuffix := getFileTimeSuffix()`.
- Lines 26, 39, 52: change the three `_%s` file paths to fixed names:

```go
	administrativeUnitsFilePath := fmt.Sprintf("%s/administrative_units.json", w.OutputFolderPath)
```

```go
	administrativeRegionsFilePath := fmt.Sprintf("%s/administrative_regions.json", w.OutputFolderPath)
```

```go
	dataProvinceMongoPath := fmt.Sprintf("%s/mongo_data_vn_unit.json", w.OutputFolderPath)
```

- Replace the end of `WriteToFile`:

```go
	dataProvinceMongoFile.Close()
	return writeMongoReadme(w.OutputFolderPath)
}

func writeMongoReadme(outputFolderPath string) error {
	return writeDatasetReadme(outputFolderPath,
		"MongoDB Dataset — Vietnamese Provinces Database",
		"MongoDB documents for Vietnamese provinces with embedded wards.",
		[]DatasetReadmeFile{
			{Name: "administrative_units.json", Description: "Array of 8 administrative unit types"},
			{Name: "administrative_regions.json", Description: "Array of 8 regions"},
			{Name: "mongo_data_vn_unit.json", Description: "Array of 34 province documents, each embedding its Wards array"},
		},
		[]string{
			"## Data Structure",
			"",
			"- `administrative_units.json` — array of 8 administrative unit types",
			"- `administrative_regions.json` — array of 8 regions",
			"- `mongo_data_vn_unit.json` — the `provinces` collection: 34 province documents, each with an embedded `Wards` array",
			"",
			"## Sample Queries",
			"",
			"```javascript",
			"// Count provinces",
			"db.getCollection('provinces').countDocuments();",
			"",
			"// Wards of a province",
			"db.getCollection('provinces').findOne({Code: '01'}, {Name: 1, Wards: 1});",
			"```",
			"",
			"## GIS / GeoJSON",
			"",
			"The `gis/` subfolder contains the `provinces-gis` (34) and `wards-gis` (3,321) collections (`mongo_data_vn_province_gis.json`, `mongo_data_vn_ward_gis[_part_NN].json`), the `create_indexes.js` index script, and a `.manifest`. Import them, run `create_indexes.js`, then query with `$geoIntersects`.",
		})
}
```

- [ ] **Step 4: Implement — mongodb gis writer**

In `mongodb_gis_file_writer.go` `WriteMongoGISDataToFile`:
- Delete line 29 `fileTimeSuffix := getFileTimeSuffix()`.
- Lines 41, 47: change to fixed names:

```go
	provincePath := fmt.Sprintf("%s/mongo_data_vn_province_gis.json", w.OutputFolderPath)
```

```go
	wardPath := fmt.Sprintf("%s/mongo_data_vn_ward_gis.json", w.OutputFolderPath)
```

- Delete the README block (lines 58-62):

```go
	// Write README.md
	readmePath := fmt.Sprintf("%s/README.md", w.OutputFolderPath)
	if err := writeMongoGISReadme(readmePath); err != nil {
		return fmt.Errorf("write README: %w", err)
	}
```

- Delete the entire `writeMongoGISReadme` function (lines 195 through the end of its `os.WriteFile` return, before `// writeChunkedMongoJSON`'s siblings). After removal, verify the file still compiles (`time` import is still used by `generatedAt := time.Now().UTC()` at line 30).

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestMongoDBDatasetFileWriter -v`
Expected: PASS.

- [ ] **Step 6: Run the full package test suite**

Run: `go test ./internal/dataset_writer/dataset_file_writer/...`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/mongodb_file_writer.go internal/dataset_writer/dataset_file_writer/mongodb_gis_file_writer.go internal/dataset_writer/dataset_file_writer/mongodb_gis_file_writer_test.go internal/dataset_writer/dataset_file_writer/mongodb_file_writer_test.go
git commit -m "feat: deterministic mongodb filenames, base README, drop gis-subfolder README"
```

---

### Task 6: Redis — deterministic filename + README

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/redis_file_writer.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/redis_file_writer_test.go`

- [ ] **Step 1: Update the failing tests**

In `redis_file_writer_test.go`, every occurrence of:

```go
	assert.Len(t, files, 1)
```

(2 occurrences, lines 38 and 187) becomes:

```go
	assert.Len(t, files, 2)
```

and every read of `files[0].Name()` (lines 40, 76, 109, 146, 189) must target the `.redis` file instead. Replace each occurrence of:

```go
	content, err := os.ReadFile(filepath.Join(tmpDir, files[0].Name()))
```

with:

```go
	var datasetName string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".redis") {
			datasetName = f.Name()
			break
		}
	}
	require.NotEmpty(t, datasetName, "expected a .redis dataset file")
	content, err := os.ReadFile(filepath.Join(tmpDir, datasetName))
```

Add `"github.com/stretchr/testify/require"` to the test file imports (keep `assert`). Add `"strings"` if not already imported (it is, per line 5 of the current imports).

Append a README test:

```go
func TestRedisDatasetFileWriter_WriteToFile_README(t *testing.T) {
	tmpDir := t.TempDir()
	writer := RedisDatasetFileWriter{OutputFolderPath: tmpDir}
	provinces := []vn_provinces_tmp_model.Province{{Code: "01", Name: "Hà Nội", AdministrativeUnitId: 1}}

	err := writer.WriteToFile(nil, nil, provinces, nil)
	assert.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "README.md"))
	assert.NoError(t, err)
	s := string(content)
	assert.Contains(t, s, "**Generated at:")
	assert.Contains(t, s, "redis_vn_provinces_dataset.redis")
	assert.Contains(t, s, "## Data Structure")
	assert.Contains(t, s, "## Sample Queries")
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestRedisDatasetFileWriter -v`
Expected: FAIL — README missing; `Len(t, files, 1)` sees 2 files; `files[0]` is README.md.

- [ ] **Step 3: Implement**

In `redis_file_writer.go`:
- Delete line 31 `fileTimeSuffix := getFileTimeSuffix()`.
- Change line 33 to:

```go
	redisDatasetFilePath := fmt.Sprintf("%s/redis_vn_provinces_dataset.redis", w.OutputFolderPath)
```

- Replace the end of `WriteToFile`:

```go
	dataWriter.Flush()
	redisDatasetFile.Close()

	return writeRedisReadme(w.OutputFolderPath)
}

func writeRedisReadme(outputFolderPath string) error {
	return writeDatasetReadme(outputFolderPath,
		"Redis Dataset — Vietnamese Provinces Database",
		"Redis commands loading all Vietnamese provinces, wards, regions, and administrative units.",
		[]DatasetReadmeFile{
			{Name: "redis_vn_provinces_dataset.redis", Description: "Redis HSET/SADD commands"},
		},
		[]string{
			"## Data Structure",
			"",
			"- `province:<code>` — province hash (name, nameEn, fullName, codeName, postalCodePrefix, administrativeUnitId)",
			"- `ward:<code>` — ward hash (name, fullName, codeName, postalCode, administrativeUnitId, districtCode)",
			"- `administrativeUnit:<id>` — unit type hash",
			"- `region:<id>` — region hash",
			"- `province:<code>:wards` — SET of ward codes",
			"- `province:<code>:wards:vn` / `province:<code>:wards:en` — ward code → name hashes",
			"",
			"## Sample Queries",
			"",
			"```bash",
			"redis-cli HGETALL province:01",
			"redis-cli SMEMBERS province:01:wards",
			"redis-cli HGET ward:00004 fullName",
			"```",
		})
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestRedisDatasetFileWriter -v`
Expected: PASS.

- [ ] **Step 5: Run the full package test suite**

Run: `go test ./internal/dataset_writer/dataset_file_writer/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/redis_file_writer.go internal/dataset_writer/dataset_file_writer/redis_file_writer_test.go
git commit -m "feat: deterministic redis filename and generated README"
```

---

### Task 7: Elasticsearch — shared-format README header + files section

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer.go`

**Interfaces:**
- Consumes: `writeDatasetReadme`, `DatasetReadmeFile` from Task 1.

- [ ] **Step 1: Update the existing README test to assert the shared format**

In `elasticsearch_file_writer_test.go`, the README check at lines 126-133 currently only asserts non-empty. Extend it (inside the existing test that reads `readmePath`) by appending after the `t.Fatal("README.md is empty")` block:

```go
	if !bytes.Contains(readme, []byte("**Generated at:")) {
		t.Fatal("README.md missing bold Generated at header")
	}
	if !bytes.Contains(readme, []byte("## Files")) {
		t.Fatal("README.md missing Files section")
	}
	if !bytes.Contains(readme, []byte("## Sample Queries")) {
		t.Fatal("README.md missing Sample Queries section")
	}
```

Add `"bytes"` to the test file imports if not present.

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestWriteToFile_NonGIS -v`
Expected: FAIL — the README currently has `Created at:` (not bold) and no `## Files`/`## Sample Queries`.

- [ ] **Step 3: Implement**

Rewrite `writeESReadme` (currently `elasticsearch_file_writer.go:687-1003`) to build its rich body and hand it to the shared helper. Replace the function body with:

```go
// writeESReadme writes the README.md for the Elasticsearch dataset.
func writeESReadme(path string) error {
	outputFolderPath := filepath.Dir(path)

	sections := []string{
		"## Overview",
		"",
		"This dataset provides Vietnamese provinces and wards in Elasticsearch document format with two indices:",
		"",
		"| Index | Description |",
		"|-------|-------------|",
		"| `provinces` | Provincial metadata with embedded wards, search keywords, and administrative unit data (no GIS geometry) |",
		"| `provinces-gis` | Same structure plus GIS geometry for both provinces and wards (bounding boxes + GeoJSON polygons) |",
		"",
		"## Data Structure",
		"",
		"Each province is a single denormalized document with:",
		"",
		"- **Core fields**: Code, Name, NameEn, FullName, FullNameEn, CodeName",
		"- **`AdministrativeUnit`**: Embedded administrative unit object (Id, FullName, ShortName, etc.)",
		"- **`SearchKeywords`**: Pre-computed autocomplete keywords (code, tone-stripped name, English name, codeName)",
		"- **`Wards`**: Array of nested ward documents with the same structure",
		"- **`GIS`**: (provinces-gis only) Center (geo_point), BoundingBox, Geometry (geo_shape)",
		"- **`Meta`**: Dataset version metadata (DatasetVersion, AdministrativeRevision, GeneratedAt)",
		"",
		"## Sample Queries",
		"",
		"```bash",
		"# Count documents",
		`curl "localhost:9200/provinces/_count"`,
		`curl "localhost:9200/provinces-gis/_count"`,
		"",
		"# Autocomplete search",
		`curl "localhost:9200/provinces/_search" -H 'Content-Type: application/json' -d '{"query": {"terms": {"SearchKeywords": ["ha noi"]}}, "_source": ["Code", "Name", "NameEn"]}'`,
		"",
		"# GIS: find province covering a point",
		`curl "localhost:9200/provinces-gis/_search" -H 'Content-Type: application/json' -d '{"query": {"geo_shape": {"GIS.Geometry": {"shape": {"type": "point", "coordinates": [105.8542, 21.0285]}, "relation": "intersects"}}}, "_source": ["Code", "Name"]}'`,
		"```",
		"",
		"## Quick Start",
		"",
		"1. Create the indices with the mappings in `mappings/`.",
		"2. Bulk import `provinces.ndjson` (and the `provinces-gis-part-*.ndjson` chunks in order, per `provinces-gis.ndjson.manifest`).",
		"3. Verify: 34 documents in each index.",
		"",
		"## Notes",
		"",
		"- Field names use **PascalCase** (consistent with MongoDB/JSON exports).",
		"- The `Meta` field is named without an underscore prefix — Elasticsearch reserves `_`-prefixed field names.",
		"- NDJSON files use the Elasticsearch Bulk API format.",
	}

	return writeDatasetReadme(outputFolderPath,
		"Elasticsearch Dataset — Vietnamese Provinces Database",
		"Provinces and wards as Elasticsearch documents in two indices: `provinces` (no geometry) and `provinces-gis` (with GIS geometry).",
		[]DatasetReadmeFile{
			{Name: "provinces.ndjson", Description: "Bulk API NDJSON for the provinces index"},
			{Name: "mappings/provinces.json", Description: "Index mapping for provinces"},
		},
		sections)
}
```

This drops the old 300-line body in favor of the shared skeleton. The rich content (overview, structure, queries, quick start, notes) is preserved in condensed form.

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run "TestWriteToFile_NonGIS|TestWriteElasticsearchGISDataToFile_GIS" -v`
Expected: PASS.

- [ ] **Step 5: Run the full package test suite**

Run: `go test ./internal/dataset_writer/dataset_file_writer/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer.go internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer_test.go
git commit -m "feat: use shared README format for elasticsearch dataset"
```

---

### Task 8: JSON — shared-format README; geojson — drop README from folder and zip

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/json_file_writer.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/geojson_file_writer.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/json_file_writer_test.go`

- [ ] **Step 1: Update the failing tests**

In `json_file_writer_test.go`:
- `TestJSONDatasetFileWriter_WriteToFile_README` stays (asserts `**Generated at:`, filenames, `geojson/`, `vn_provinces_wards_geojson.zip`). Add assertions for the new sections:

```go
	assert.Contains(t, contentStr, "## Data Structure")
	assert.Contains(t, contentStr, "## Sample Queries")
```

- In `TestJSONDatasetFileWriter_WriteGISGeoJSONToFile`, replace the README assertions (lines 336-343):

```go
	readmeContent, err := os.ReadFile(filepath.Join(tmpDir, "README.md"))
	require.NoError(t, err)
	assert.Contains(t, string(readmeContent), "Created at:")
	assert.Contains(t, string(readmeContent), "geojson.io")
	assert.Contains(t, string(readmeContent), "{province_code}_{province_code_name}")
```

with:

```go
	_, err = os.Stat(filepath.Join(tmpDir, "README.md"))
	require.Error(t, err, "geojson subfolder should no longer have a README.md")
```

- Replace the zip entry assertion (line 352):

```go
	assert.Contains(t, names, "geojson/README.md")
```

with:

```go
	assert.NotContains(t, names, "geojson/README.md")
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run "TestJSONDatasetFileWriter_WriteToFile_README|TestJSONDatasetFileWriter_WriteGISGeoJSONToFile" -v`
Expected: FAIL — JSON README lacks `## Data Structure`; geojson README still exists and is still in the zip.

- [ ] **Step 3: Implement — JSON README via shared helper**

In `json_file_writer.go`, replace `writeJSONDatasetReadme` (currently lines 73-111) and remove `formatFileSize` (already removed in Task 1). New implementation:

```go
// writeJSONDatasetReadme writes a README describing the JSON dataset files and the
// optional geojson artifacts, with a bold generation timestamp at the top.
func writeJSONDatasetReadme(outputFolderPath string) error {
	return writeDatasetReadme(outputFolderPath,
		"JSON Dataset — Vietnamese Provinces Database",
		"Administrative unit JSON data for Vietnam: provinces with embedded wards.",
		[]DatasetReadmeFile{
			{Name: "full_json_generated_data_vn_units.json", Description: "Full dataset (provinces + wards + administrative info)"},
			{Name: "simplified_json_generated_data_vn_units.json", Description: "Simplified dataset (pretty-printed)"},
			{Name: "simplified_json_generated_data_vn_units_minified.json", Description: "Simplified dataset (minified)"},
			{Name: "vn_only_simplified_json_generated_data_vn_units.json", Description: "Vietnamese-only simplified (pretty-printed)"},
			{Name: "vn_only_simplified_json_generated_data_vn_units_minified.json", Description: "Vietnamese-only simplified (minified)"},
		},
		[]string{
			"## Data Structure",
			"",
			"Each entry is a province object with `code`, `name`, `nameEn`, `fullName`, `fullNameEn`, `codeName`, `administrativeUnit*`, `postalCodePrefix`, and a `wards` array of ward objects.",
			"",
			"## Sample Queries",
			"",
			"```js",
			"const dataset = require('./full_json_generated_data_vn_units.json');",
			"dataset.find(p => p.code === '01');",
			"dataset.flatMap(p => p.wards).filter(w => w.postalCode === '11024');",
			"```",
			"",
			"## GIS / GeoJSON",
			"",
			"The `geojson/` subfolder contains per-province and per-ward GeoJSON boundary exports, and `vn_provinces_wards_geojson.zip` is the combined archive of those files. These artifacts are present when the GIS generation step runs.",
		})
}
```

- [ ] **Step 4: Implement — remove geojson subfolder README**

In `geojson_file_writer.go`:
- Remove the `writeGeoJSONReadme` call in `WriteGISGeoJSONToFile` (currently lines 48-50):

```go
	if err := writeGeoJSONReadme(outputFolderPath, executionTime); err != nil {
		return err
	}
```

- Remove the now-unused `executionTime` local (line 23):

```go
	executionTime := time.Now()
```

- Delete the entire `writeGeoJSONReadme` function (currently lines 200-243).
- Remove the now-unused `"strings"` and `"time"` imports from the import block (verified: `strings` was only used by `writeGeoJSONReadme`; `time` only by `executionTime` and `writeGeoJSONReadme`).
- The zip archive no longer includes a README because the folder no longer contains one (the archive walks the folder).

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run "TestJSONDatasetFileWriter" -v`
Expected: PASS.

- [ ] **Step 6: Run the full package test suite**

Run: `go test ./internal/dataset_writer/dataset_file_writer/...`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/json_file_writer.go internal/dataset_writer/dataset_file_writer/geojson_file_writer.go internal/dataset_writer/dataset_file_writer/json_file_writer_test.go
git commit -m "feat: shared JSON README format; drop geojson subfolder README from folder and zip"
```

---

### Task 9: Copy script — prune timestamped variants, drop strip logic

**Files:**
- Modify: `dataset-generation-scripts/copy-datasets-to-repo.sh`

- [ ] **Step 1: Update the script**

Replace the whole script with:

```bash
#!/usr/bin/env bash
set -euo pipefail

# One-shot copy of generated dataset outputs (dataset-generation-scripts/output/)
# into the repository's published folders (json/, postgresql/, ...).
#
# Run this AFTER regenerating the dataset (go run main.go) so the output
# contains the freshly generated files.
#
# Usage:
#   ./copy-datasets-to-repo.sh            # copy everything
#   ./copy-datasets-to-repo.sh --dry-run  # preview what would be copied
#   ./copy-datasets-to-repo.sh json       # copy only one dataset

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
OUTPUT_DIR="$SCRIPT_DIR/output"

DRY_RUN=0
DATASETS=()
for arg in "$@"; do
  case "$arg" in
    --dry-run|-n) DRY_RUN=1 ;;
    *) DATASETS+=("$arg") ;;
  esac
done

run() {
  if [[ "$DRY_RUN" -eq 1 ]]; then
    echo "[dry-run] $*"
  else
    "$@"
  fi
}

# copy_datasets <output_subdir> <repo_target_dir>
#   Copies every top-level entry (files + subdirs) from output/<output_subdir>
#   into <repo_target_dir>, overwriting existing files.
copy_datasets() {
  local src="$OUTPUT_DIR/$1"
  local dst="$REPO_ROOT/$2"
  if [[ ! -d "$src" ]]; then
    echo "skip: $1 not found in output/"
    return 0
  fi
  echo "== $1 -> $2 =="
  run mkdir -p "$dst"
  run cp -R "$src/." "$dst/"
}

# prune_datetime_variants <repo_target_dir> <glob>
#   Removes timestamped artifact variants (matching <glob>) from <repo_target_dir>.
prune_datetime_variants() {
  local target="$REPO_ROOT/$1"
  local stale=("$target"/$2)
  if [[ -f "${stale[0]:-}" ]]; then
    for f in "${stale[@]}"; do
      echo "remove stale: $f"
      run rm "$f"
    done
  fi
}

# Prune stale timestamped geojson archives from json/ so only the fixed-name
# zip (vn_provinces_wards_geojson.zip) remains after copying.
prune_stale_geojson_zips() {
  local stale_zips=("$REPO_ROOT"/json/vn_provinces_wards_geojson_*.zip)
  if [[ -f "${stale_zips[0]:-}" ]]; then
    for z in "${stale_zips[@]}"; do
      [[ "$(basename "$z")" == "vn_provinces_wards_geojson.zip" ]] && continue
      echo "remove stale: $z"
      run rm "$z"
    done
  fi
}

copy_all() {
  copy_datasets json json
  prune_stale_geojson_zips

  copy_datasets postgresql postgresql
  prune_datetime_variants postgresql/gis 'postgresql_ImportData_gis_*.sql*'

  copy_datasets mysql mysql
  prune_datetime_variants mysql/gis 'mysql_ImportData_gis_*.sql*'

  copy_datasets sqlserver sqlserver
  prune_datetime_variants sqlserver/gis 'mssql_ImportData_gis_*.sql*'

  copy_datasets oracle oracle

  copy_datasets mongodb mongodb
  prune_datetime_variants mongodb 'administrative_units_*.json'
  prune_datetime_variants mongodb 'administrative_regions_*.json'
  prune_datetime_variants mongodb 'mongo_data_vn_unit_*.json'
  prune_datetime_variants mongodb/gis 'mongo_data_vn_province_gis_*.json'
  prune_datetime_variants mongodb/gis 'mongo_data_vn_ward_gis_2*.json*'

  copy_datasets redis redis
  prune_datetime_variants redis 'redis_vn_provinces_dataset_*.redis'

  copy_datasets elasticsearch elasticsearch
}

if [[ "${#DATASETS[@]}" -eq 0 ]]; then
  copy_all
else
  case "${DATASETS[0]}" in
    json)          copy_datasets json json; prune_stale_geojson_zips ;;
    postgresql)    copy_datasets postgresql postgresql; prune_datetime_variants postgresql/gis 'postgresql_ImportData_gis_*.sql*' ;;
    mysql)         copy_datasets mysql mysql; prune_datetime_variants mysql/gis 'mysql_ImportData_gis_*.sql*' ;;
    sqlserver)     copy_datasets sqlserver sqlserver; prune_datetime_variants sqlserver/gis 'mssql_ImportData_gis_*.sql*' ;;
    oracle)        copy_datasets oracle oracle ;;
    mongodb)       copy_datasets mongodb mongodb; prune_datetime_variants mongodb 'administrative_units_*.json'; prune_datetime_variants mongodb 'administrative_regions_*.json'; prune_datetime_variants mongodb 'mongo_data_vn_unit_*.json'; prune_datetime_variants mongodb/gis 'mongo_data_vn_province_gis_*.json'; prune_datetime_variants mongodb/gis 'mongo_data_vn_ward_gis_2*.json*' ;;
    redis)         copy_datasets redis redis; prune_datetime_variants redis 'redis_vn_provinces_dataset_*.redis' ;;
    elasticsearch) copy_datasets elasticsearch elasticsearch ;;
    *) echo "unknown dataset: ${DATASETS[0]}" >&2; exit 1 ;;
  esac
fi

echo
echo "Done."
```

- [ ] **Step 2: Verify syntax and dry-run**

Run: `bash -n copy-datasets-to-repo.sh && bash copy-datasets-to-repo.sh --dry-run 2>&1 | tail -25`
Expected: `syntax OK`; dry-run shows direct `cp -R` for every dataset (no `--strip-datetime`), and the prune lines match the old timestamped variants.

- [ ] **Step 3: Commit**

```bash
git add copy-datasets-to-repo.sh
git commit -m "feat: prune timestamped dataset variants in copy script"
```

---

### Task 10: Update docs

**Files:**
- Modify: `dataset-generation-scripts/README.md`
- Modify: `docs/gis/gis_readme.md`
- Modify: `docs/gis/gis_readme_vi.md`

- [ ] **Step 1: Update `dataset-generation-scripts/README.md`**

In the output-structure tree (json section already updated in the previous feature), update the GIS/mongodb/redis lines to deterministic names. Change:

```
│   ├── vn_provinces_wards_geojson.zip                      # Combined GeoJSON archive
```

(keep) and update the mongodb lines:

```
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

Remove any remaining `_*.json`/`_*.redis` suffix patterns for mongodb/redis base files.

- [ ] **Step 2: Update `docs/gis/gis_readme.md` and `docs/gis/gis_readme_vi.md`**

Replace the `<timestamp>` part-file patterns and example commands:
- `postgresql_ImportData_gis_<timestamp>-part-NN.sql` → `postgresql_ImportData_gis-part-NN.sql`
- `mysql_ImportData_gis_<timestamp>-part-NN.sql` → `mysql_ImportData_gis-part-NN.sql`
- `mssql_ImportData_gis_<timestamp>-part-NN.sql` → `mssql_ImportData_gis-part-NN.sql`
- Example commands like `...-f postgresql/gis/postgresql_ImportData_gis_2026-06-20__12_32_01-part-01.sql` → `...-f postgresql/gis/postgresql_ImportData_gis-part-01.sql` (same for `-part-02.sql`, mysql, mssql).
- Leave the `[gis_dataset_*_bucket_url]` download links at the bottom unchanged (they reference immutable historical v4.0.0 artifacts).

- [ ] **Step 3: Verify with git diff**

Run: `git diff --stat`
Expected: the 3 doc files listed.

- [ ] **Step 4: Commit**

```bash
git add dataset-generation-scripts/README.md docs/gis/gis_readme.md docs/gis/gis_readme_vi.md
git commit -m "docs: document deterministic GIS/mongodb/redis filenames and generated READMEs"
```

---

## Final Verification

Run from `dataset-generation-scripts/`:

```bash
go test ./internal/dataset_writer/dataset_file_writer/... -v
```

Expected: all tests pass. Then:

```bash
bash -n copy-datasets-to-repo.sh
```

Optional full-generation smoke check (requires Docker Postgres up on `localhost:15432` and internet for GIS): `go run main.go`, then `bash copy-datasets-to-repo.sh --dry-run`, inspect `output/*/README.md` files, and manually delete the now-stale `json/geojson/README.md` and `mongodb/gis/README.md` from the published folders. Then run `go run main.go` + `bash copy-datasets-to-repo.sh` to publish and regenerate all dataset artifacts.

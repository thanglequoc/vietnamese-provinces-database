# Stable JSON Output Filenames, Minified Variants, and Combined README — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the JSON dataset generation output copy-paste ready for the published `json/` folder by removing datetime suffixes, generating `_minified` variants, and adding a combined README with a bold timestamp.

**Architecture:** Modify the `JSONDatasetFileWriter.WriteToFile` (pretty + minified JSON + README generation) and the geojson archive writer (zip naming), update the writer tests, then update repo docs. Output filenames become deterministic so the user can `cp` them straight over the `json/` folder.

**Tech Stack:** Go 1.24, stdlib `encoding/json`/`os`/`path/filepath`/`strings`/`time`, Testify (`assert`/`require`) for tests.

## Global Constraints

- Module root: `dataset-generation-scripts/`; module path `github.com/thanglequoc-vn-provinces/v2`.
- Tests run from `dataset-generation-scripts/`: `go test ./internal/dataset_writer/dataset_file_writer/... -v`.
- Pretty JSON keeps the existing single-space indent (`json.MarshalIndent(payload, "", " ")`) so generated file bytes stay compatible with the published dataset.
- Minified JSON uses compact `json.Marshal(payload)` (no indentation, no newlines).
- Final output filenames (exact):
  - `full_json_generated_data_vn_units.json`
  - `simplified_json_generated_data_vn_units.json`
  - `simplified_json_generated_data_vn_units_minified.json`
  - `vn_only_simplified_json_generated_data_vn_units.json`
  - `vn_only_simplified_json_generated_data_vn_units_minified.json`
  - `README.md` (in the json output root)
  - `vn_provinces_wards_geojson.zip` (in the json output root)
- No commit of generated dataset artifacts; only source/tests/docs are committed.

---

### Task 1: Deterministic admin JSON filenames (remove datetime suffix)

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/json_file_writer.go` (whole file rewrite)
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/json_file_writer_test.go` (Tests: `TestJSONDatasetFileWriter_WriteToFile_FullJSON`, `TestJSONDatasetFileWriter_WriteToFile_EmptyDataset`, `TestJSONDatasetFileWriter_WriteToFile_MultipleProvinces`, `TestJSONDatasetFileWriter_WriteToFile_PostalCodes`)

**Interfaces:**
- Produces: helper `writePrettyJSON(filePath string, payload any) error` — marshals with single-space indent and writes with `os.WriteFile(path, data, 0644)`.

- [ ] **Step 1: Update the failing tests**

Replace `TestJSONDatasetFileWriter_WriteToFile_FullJSON` (currently lines 16-52) with:

```go
func TestJSONDatasetFileWriter_WriteToFile_FullJSON(t *testing.T) {
	tmpDir := t.TempDir()

	writer := &JSONDatasetFileWriter{
		OutputFolderPath: tmpDir,
	}

	provinces := []vn_provinces_tmp_model.Province{
		{
			Code:                 "01",
			Name:                 "Hà Nội",
			NameEn:               "Ha Noi",
			FullName:             "Thành phố Hà Nội",
			FullNameEn:           "Ha Noi City",
			CodeName:             "ha_noi",
			AdministrativeUnitId: 1,
		},
	}

	err := writer.WriteToFile(nil, nil, provinces, nil)
	assert.NoError(t, err)

	files, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)
	assert.Len(t, files, 3, "should create 3 JSON files (full, simplified, vn_only)")

	// Deterministic filename — no datetime suffix
	fullContent, err := os.ReadFile(filepath.Join(tmpDir, "full_json_generated_data_vn_units.json"))
	assert.NoError(t, err)

	var data interface{}
	err = json.Unmarshal(fullContent, &data)
	assert.NoError(t, err, "should produce valid JSON")

	contentStr := string(fullContent)
	assert.Contains(t, contentStr, "Hà Nội")
}
```

Replace `TestJSONDatasetFileWriter_WriteToFile_EmptyDataset` (currently lines 54-77) with:

```go
func TestJSONDatasetFileWriter_WriteToFile_EmptyDataset(t *testing.T) {
	tmpDir := t.TempDir()

	writer := &JSONDatasetFileWriter{
		OutputFolderPath: tmpDir,
	}

	err := writer.WriteToFile(nil, nil, nil, nil)
	assert.NoError(t, err)

	files, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)
	assert.Len(t, files, 3, "should create 3 JSON files even with empty data")

	// Verify files are created and valid JSON
	for _, f := range files {
		content, err := os.ReadFile(filepath.Join(tmpDir, f.Name()))
		assert.NoError(t, err)

		var data interface{}
		err = json.Unmarshal(content, &data)
		assert.NoError(t, err, f.Name()+" should be valid JSON")
	}
}
```

Replace `TestJSONDatasetFileWriter_WriteToFile_MultipleProvinces` (currently lines 79-129) with:

```go
func TestJSONDatasetFileWriter_WriteToFile_MultipleProvinces(t *testing.T) {
	tmpDir := t.TempDir()

	writer := &JSONDatasetFileWriter{
		OutputFolderPath: tmpDir,
	}

	provinces := []vn_provinces_tmp_model.Province{
		{
			Code:                 "01",
			Name:                 "Hà Nội",
			NameEn:               "Ha Noi",
			FullName:             "Thành phố Hà Nội",
			FullNameEn:           "Ha Noi City",
			CodeName:             "ha_noi",
			AdministrativeUnitId: 1,
		},
		{
			Code:                 "02",
			Name:                 "Hải Phòng",
			NameEn:               "Hai Phong",
			FullName:             "Thành phố Hải Phòng",
			FullNameEn:           "Hai Phong City",
			CodeName:             "hai_phong",
			AdministrativeUnitId: 1,
		},
		{
			Code:                 "10",
			Name:                 "Khánh Hòa",
			NameEn:               "Khanh Hoa",
			FullName:             "Tỉnh Khánh Hòa",
			FullNameEn:           "Khanh Hoa Province",
			CodeName:             "khanh_hoa",
			AdministrativeUnitId: 2,
		},
	}

	err := writer.WriteToFile(nil, nil, provinces, nil)
	assert.NoError(t, err)

	files, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)
	assert.Len(t, files, 3)

	// Verify full JSON contains all provinces
	fullContent, _ := os.ReadFile(filepath.Join(tmpDir, "full_json_generated_data_vn_units.json"))
	contentStr := string(fullContent)
	assert.Contains(t, contentStr, "Hà Nội")
	assert.Contains(t, contentStr, "Hải Phòng")
	assert.Contains(t, contentStr, "Khánh Hòa")
}
```

Replace `TestJSONDatasetFileWriter_WriteToFile_PostalCodes` (currently lines 131-165) with:

```go
func TestJSONDatasetFileWriter_WriteToFile_PostalCodes(t *testing.T) {
	tmpDir := t.TempDir()

	writer := &JSONDatasetFileWriter{OutputFolderPath: tmpDir}

	provinces := []vn_provinces_tmp_model.Province{
		{
			Code: "01", Name: "Hà Nội", NameEn: "Ha Noi",
			FullName: "Thành phố Hà Nội", FullNameEn: "Ha Noi City",
			CodeName: "ha_noi", AdministrativeUnitId: 1,
			PostalCodePrefix: "10, 11, 12, 13, 14",
			Wards: []*vn_provinces_tmp_model.Ward{
				{
					Code: "00070", Name: "Hoàn Kiếm", NameEn: "Hoan Kiem",
					FullName: "Phường Hoàn Kiếm", FullNameEn: "Hoan Kiem Ward",
					CodeName: "hoan_kiem", ProvinceCode: "01", AdministrativeUnitId: 3,
					PostalCode: "11024",
				},
			},
		},
	}

	err := writer.WriteToFile(nil, nil, provinces, nil)
	assert.NoError(t, err)

	fullContent, err := os.ReadFile(filepath.Join(tmpDir, "full_json_generated_data_vn_units.json"))
	assert.NoError(t, err)
	assert.Contains(t, string(fullContent), "11024")
	assert.Contains(t, string(fullContent), "10, 11, 12, 13, 14")
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestJSONDatasetFileWriter_WriteToFile -v`
Expected: FAIL — `json_file_writer.go` still writes `full_json_generated_data_vn_units_<suffix>.json`, so the exact-filename read fails with "no such file or directory".

- [ ] **Step 3: Rewrite `json_file_writer.go`**

Replace the entire file contents with:

```go
package dataset_writer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
	file_writer_helper "github.com/thanglequoc-vn-provinces/v2/internal/dataset_writer/dataset_file_writer/helper"
)

type JSONDatasetFileWriter struct {
	OutputFolderPath string
}

func (w *JSONDatasetFileWriter) WriteToFile(
	regions []model.AdministrativeRegion,
	administrativeUnits []model.AdministrativeUnit,
	provinces []model.Province,
	wards []model.Ward) error {

	os.MkdirAll(w.OutputFolderPath, 0746)

	provinceData := file_writer_helper.ConvertToJsonProvinceModel(provinces)
	if err := writePrettyJSON(filepath.Join(w.OutputFolderPath, "full_json_generated_data_vn_units.json"), provinceData); err != nil {
		return err
	}

	// JSON Simplified file
	provinceSimplifiedData := file_writer_helper.ConvertToJsonProvinceSimplifiedModel(provinces)
	if err := writePrettyJSON(filepath.Join(w.OutputFolderPath, "simplified_json_generated_data_vn_units.json"), provinceSimplifiedData); err != nil {
		return err
	}

	// VN only JSON Simplified file
	provinceVNSimplifiedData := file_writer_helper.ConvertToJsonProvinceVNSimplifiedModel(provinces)
	if err := writePrettyJSON(filepath.Join(w.OutputFolderPath, "vn_only_simplified_json_generated_data_vn_units.json"), provinceVNSimplifiedData); err != nil {
		return err
	}

	return nil
}

// writePrettyJSON marshals payload with a single-space indent and writes it to filePath.
func writePrettyJSON(filePath string, payload any) error {
	data, err := json.MarshalIndent(payload, "", " ")
	if err != nil {
		return fmt.Errorf("marshal pretty json payload for %s: %w", filePath, err)
	}
	return os.WriteFile(filePath, data, 0644)
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestJSONDatasetFileWriter_WriteToFile -v`
Expected: PASS for all 4 tests.

- [ ] **Step 5: Run the full package test suite**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -v`
Expected: PASS (all existing tests incl. geojson writer tests).

- [ ] **Step 6: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/json_file_writer.go internal/dataset_writer/dataset_file_writer/json_file_writer_test.go
git commit -m "feat: remove datetime suffix from JSON dataset filenames"
```

---

### Task 2: Generate minified JSON variants

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/json_file_writer.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/json_file_writer_test.go`

**Interfaces:**
- Consumes: `writePrettyJSON(filePath string, payload any) error` from Task 1.
- Produces: helper `writeMinifiedJSON(filePath string, payload any) error` — compact `json.Marshal`, writes with `os.WriteFile(path, data, 0644)`.

- [ ] **Step 1: Add the failing minified test**

Append this test to `json_file_writer_test.go`:

```go
func TestJSONDatasetFileWriter_WriteToFile_Minified(t *testing.T) {
	tmpDir := t.TempDir()

	writer := &JSONDatasetFileWriter{OutputFolderPath: tmpDir}

	provinces := []vn_provinces_tmp_model.Province{
		{
			Code:                 "01",
			Name:                 "Hà Nội",
			NameEn:               "Ha Noi",
			FullName:             "Thành phố Hà Nội",
			FullNameEn:           "Ha Noi City",
			CodeName:             "ha_noi",
			AdministrativeUnitId: 1,
		},
	}

	err := writer.WriteToFile(nil, nil, provinces, nil)
	require.NoError(t, err)

	files, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	assert.Len(t, files, 5, "should create 5 JSON files (full, simplified + minified, vn_only + minified)")

	minifiedFiles := []string{
		"simplified_json_generated_data_vn_units_minified.json",
		"vn_only_simplified_json_generated_data_vn_units_minified.json",
	}
	for _, name := range minifiedFiles {
		path := filepath.Join(tmpDir, name)
		content, err := os.ReadFile(path)
		require.NoError(t, err, name+" should exist")

		var data interface{}
		require.NoError(t, json.Unmarshal(content, &data), name+" should be valid JSON")
		assert.NotContains(t, string(content), "\n", name+" should be a single line")
		assert.Contains(t, string(content), "Hà Nội")
	}

	// Minified simplified must be smaller than its pretty counterpart
	prettyInfo, err := os.Stat(filepath.Join(tmpDir, "simplified_json_generated_data_vn_units.json"))
	require.NoError(t, err)
	minifiedInfo, err := os.Stat(filepath.Join(tmpDir, "simplified_json_generated_data_vn_units_minified.json"))
	require.NoError(t, err)
	assert.Less(t, minifiedInfo.Size(), prettyInfo.Size())
}
```

- [ ] **Step 2: Run the tests to verify the new test fails**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestJSONDatasetFileWriter_WriteToFile_Minified -v`
Expected: FAIL — no minified files created; `assert.Len` sees 3 files, and reads of `_minified` paths fail.

- [ ] **Step 3: Add minified output and helper to `json_file_writer.go`**

In `WriteToFile`, after each simplified file is written, write its minified counterpart. The full method becomes:

```go
func (w *JSONDatasetFileWriter) WriteToFile(
	regions []model.AdministrativeRegion,
	administrativeUnits []model.AdministrativeUnit,
	provinces []model.Province,
	wards []model.Ward) error {

	os.MkdirAll(w.OutputFolderPath, 0746)

	provinceData := file_writer_helper.ConvertToJsonProvinceModel(provinces)
	if err := writePrettyJSON(filepath.Join(w.OutputFolderPath, "full_json_generated_data_vn_units.json"), provinceData); err != nil {
		return err
	}

	// JSON Simplified file
	provinceSimplifiedData := file_writer_helper.ConvertToJsonProvinceSimplifiedModel(provinces)
	if err := writePrettyJSON(filepath.Join(w.OutputFolderPath, "simplified_json_generated_data_vn_units.json"), provinceSimplifiedData); err != nil {
		return err
	}
	// JSON Simplified minified file
	if err := writeMinifiedJSON(filepath.Join(w.OutputFolderPath, "simplified_json_generated_data_vn_units_minified.json"), provinceSimplifiedData); err != nil {
		return err
	}

	// VN only JSON Simplified file
	provinceVNSimplifiedData := file_writer_helper.ConvertToJsonProvinceVNSimplifiedModel(provinces)
	if err := writePrettyJSON(filepath.Join(w.OutputFolderPath, "vn_only_simplified_json_generated_data_vn_units.json"), provinceVNSimplifiedData); err != nil {
		return err
	}
	// VN only JSON Simplified minified file
	if err := writeMinifiedJSON(filepath.Join(w.OutputFolderPath, "vn_only_simplified_json_generated_data_vn_units_minified.json"), provinceVNSimplifiedData); err != nil {
		return err
	}

	return nil
}
```

Append the helper after `writePrettyJSON`:

```go
// writeMinifiedJSON marshals payload compactly (no whitespace) and writes it to filePath.
func writeMinifiedJSON(filePath string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal minified json payload for %s: %w", filePath, err)
	}
	return os.WriteFile(filePath, data, 0644)
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestJSONDatasetFileWriter_WriteToFile -v`
Expected: `TestJSONDatasetFileWriter_WriteToFile_Minified` PASSES; `TestJSONDatasetFileWriter_WriteToFile_EmptyDataset` and `TestJSONDatasetFileWriter_WriteToFile_MultipleProvinces` FAIL on file-count assertions (they still expect 3 files, but 5 now exist). The two count assertions are fixed in Step 5.

- [ ] **Step 5: Update the remaining file-count assertions**

In `TestJSONDatasetFileWriter_WriteToFile_EmptyDataset`, change `assert.Len(t, files, 3, ...)` to `assert.Len(t, files, 5, ...)`.
In `TestJSONDatasetFileWriter_WriteToFile_MultipleProvinces`, change `assert.Len(t, files, 3)` to `assert.Len(t, files, 5)`.

- [ ] **Step 6: Run the full package test suite**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -v`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/json_file_writer.go internal/dataset_writer/dataset_file_writer/json_file_writer_test.go
git commit -m "feat: generate minified JSON variants for simplified datasets"
```

---

### Task 3: Generate combined README with bold timestamp

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/json_file_writer.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/json_file_writer_test.go`

**Interfaces:**
- Consumes: `writePrettyJSON`, `writeMinifiedJSON` from Tasks 1-2.
- Produces: helper `writeJSONDatasetReadme(outputFolderPath string) error` — writes `README.md` at the json output root; helper `formatFileSize(size int64) string` — renders `B`/`KB`/`MB`.

- [ ] **Step 1: Add the failing README test**

Append this test to `json_file_writer_test.go`:

```go
func TestJSONDatasetFileWriter_WriteToFile_README(t *testing.T) {
	tmpDir := t.TempDir()

	writer := &JSONDatasetFileWriter{OutputFolderPath: tmpDir}

	err := writer.WriteToFile(nil, nil, nil, nil)
	require.NoError(t, err)

	files, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	assert.Len(t, files, 6, "should create 5 JSON files + README.md")

	readmeContent, err := os.ReadFile(filepath.Join(tmpDir, "README.md"))
	require.NoError(t, err)

	contentStr := string(readmeContent)
	assert.Contains(t, contentStr, "**Generated at:")
	assert.Contains(t, contentStr, "full_json_generated_data_vn_units.json")
	assert.Contains(t, contentStr, "simplified_json_generated_data_vn_units_minified.json")
	assert.Contains(t, contentStr, "vn_only_simplified_json_generated_data_vn_units_minified.json")
	assert.Contains(t, contentStr, "geojson/")
	assert.Contains(t, contentStr, "vn_provinces_wards_geojson.zip")
}
```

- [ ] **Step 2: Run the tests to verify the new test fails**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestJSONDatasetFileWriter_WriteToFile_README -v`
Expected: FAIL — README.md not created; `assert.Len` sees 5 files and the read fails.

- [ ] **Step 3: Add README generation to `json_file_writer.go`**

Update the `WriteToFile` signature's final return and append the readme write. Change the end of `WriteToFile` from:

```go
	// VN only JSON Simplified minified file
	if err := writeMinifiedJSON(filepath.Join(w.OutputFolderPath, "vn_only_simplified_json_generated_data_vn_units_minified.json"), provinceVNSimplifiedData); err != nil {
		return err
	}

	return nil
}
```

to:

```go
	// VN only JSON Simplified minified file
	if err := writeMinifiedJSON(filepath.Join(w.OutputFolderPath, "vn_only_simplified_json_generated_data_vn_units_minified.json"), provinceVNSimplifiedData); err != nil {
		return err
	}

	return writeJSONDatasetReadme(w.OutputFolderPath)
}
```

Append these functions at the end of `json_file_writer.go` (after `writeMinifiedJSON`):

```go
// writeJSONDatasetReadme writes a README describing the JSON dataset files and the
// optional geojson artifacts, with a bold generation timestamp at the top.
func writeJSONDatasetReadme(outputFolderPath string) error {
	readmePath := filepath.Join(outputFolderPath, "README.md")
	lines := []string{
		"# Vietnamese Provinces JSON Dataset",
		"",
		fmt.Sprintf("**Generated at: %s**", time.Now().Format(time.RFC1123Z)),
		"",
		"This folder contains the administrative unit JSON data for Vietnam, generated by the dataset generation script.",
		"",
		"## Files",
		"",
	}
	files := []string{
		"full_json_generated_data_vn_units.json",
		"simplified_json_generated_data_vn_units.json",
		"simplified_json_generated_data_vn_units_minified.json",
		"vn_only_simplified_json_generated_data_vn_units.json",
		"vn_only_simplified_json_generated_data_vn_units_minified.json",
	}
	for _, name := range files {
		filePath := filepath.Join(outputFolderPath, name)
		info, err := os.Stat(filePath)
		if err != nil {
			continue
		}
		lines = append(lines, fmt.Sprintf("- `%s` — %s", name, formatFileSize(info.Size())))
	}
	lines = append(lines,
		"",
		"## GIS / GeoJSON",
		"",
		"The `geojson/` subfolder contains per-province and per-ward GeoJSON boundary exports, and `vn_provinces_wards_geojson.zip` is the combined archive of those files. These artifacts are present when the GIS generation step runs.",
		"",
	)
	content := strings.Join(lines, "\n")

	return os.WriteFile(readmePath, []byte(content), 0644)
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

Add `"strings"` and `"time"` to the imports of `json_file_writer.go` (keep the existing `encoding/json`, `fmt`, `os`, `path/filepath`).

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -v`
Expected: `TestJSONDatasetFileWriter_WriteToFile_README` PASSES; `TestJSONDatasetFileWriter_WriteToFile_EmptyDataset` and `TestJSONDatasetFileWriter_WriteToFile_MultipleProvinces` FAIL on file-count assertions (they still expect 5 files, but 6 now exist). The two count assertions are fixed in Step 5.

- [ ] **Step 5: Update the remaining file-count assertions**

In `TestJSONDatasetFileWriter_WriteToFile_EmptyDataset`, change `assert.Len(t, files, 5, ...)` to `assert.Len(t, files, 6, ...)`.
In `TestJSONDatasetFileWriter_WriteToFile_MultipleProvinces`, change `assert.Len(t, files, 5)` to `assert.Len(t, files, 6)`.

- [ ] **Step 6: Run the full package test suite**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -v`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/json_file_writer.go internal/dataset_writer/dataset_file_writer/json_file_writer_test.go
git commit -m "feat: generate combined JSON dataset README with bold timestamp"
```

---

### Task 4: Remove datetime suffix from the geojson zip

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/geojson_file_writer.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/json_file_writer_test.go` (in `TestJSONDatasetFileWriter_WriteGISGeoJSONToFile`)

- [ ] **Step 1: Update the failing test**

In `TestJSONDatasetFileWriter_WriteGISGeoJSONToFile` (json_file_writer_test.go), replace lines 274-280:

```go
	zipMatches, err := filepath.Glob(filepath.Join(rootDir, "vn_provinces_wards_geojson_*.zip"))
	require.NoError(t, err)
	require.Len(t, zipMatches, 1)

	zipReader, err := zip.OpenReader(zipMatches[0])
	require.NoError(t, err)
	defer zipReader.Close()
```

with:

```go
	zipReader, err := zip.OpenReader(filepath.Join(rootDir, "vn_provinces_wards_geojson.zip"))
	require.NoError(t, err)
	defer zipReader.Close()
```

- [ ] **Step 2: Run the tests to verify the test fails**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestJSONDatasetFileWriter_WriteGISGeoJSONToFile -v`
Expected: FAIL — zip still named `vn_provinces_wards_geojson_<suffix>.zip`, so the exact path read fails.

- [ ] **Step 3: Update the archive writer and remove the now-unused helper**

In `geojson_file_writer.go`:
- Line 75: change `archiveGeoJSONDirectory(outputFolderPath, executionTime)` to `archiveGeoJSONDirectory(outputFolderPath)`.
- In `archiveGeoJSONDirectory`, change the signature from `func archiveGeoJSONDirectory(outputFolderPath string, executionTime time.Time) error {` to `func archiveGeoJSONDirectory(outputFolderPath string) error {`.
- Replace line 247:

```go
	archiveName := fmt.Sprintf("vn_provinces_wards_geojson_%s.zip", formatFileTimeSuffix(executionTime))
```

with:

```go
	archiveName := "vn_provinces_wards_geojson.zip"
```

- Delete the now-unused `formatFileTimeSuffix` function (currently lines 304-306):

```go
func formatFileTimeSuffix(t time.Time) string {
	return strings.ReplaceAll(strings.ReplaceAll(t.Format(time.DateTime), ":", "_"), " ", "__")
}
```

- Verify imports: `geojson_file_writer.go` still uses `strings` (in `writeGeoJSONReadme` via `strings.Join`) and `time` (in `WriteGISGeoJSONToFile` via `time.Now()` and in `writeGeoJSONReadme`). No import changes needed.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -v`
Expected: PASS (the geojson test, incl. the `geojson/README.md` zip entry assertion at line 286, unchanged).

- [ ] **Step 5: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/geojson_file_writer.go internal/dataset_writer/dataset_file_writer/json_file_writer_test.go
git commit -m "feat: use fixed geojson zip filename without datetime suffix"
```

---

### Task 5: Update repository docs

**Files:**
- Modify: `dataset-generation-scripts/README.md` (output structure section, lines 80-103)
- Modify: `docs/gis/gis_readme.md` (line 130)
- Modify: `docs/gis/gis_readme_vi.md` (line 128)

- [ ] **Step 1: Update `dataset-generation-scripts/README.md`**

Replace the current `json/` + `gis/` output-tree block (lines 80-103) with:

```
├── json/
│   ├── README.md                                           # Generated dataset README with bold timestamp
│   ├── full_json_generated_data_vn_units.json              # Full dataset (provinces + wards + districts)
│   ├── simplified_json_generated_data_vn_units.json        # Simplified names
│   ├── simplified_json_generated_data_vn_units_minified.json  # Simplified names (minified)
│   ├── vn_only_simplified_json_generated_data_vn_units.json  # Vietnamese-only simplified
│   ├── vn_only_simplified_json_generated_data_vn_units_minified.json  # Vietnamese-only simplified (minified)
│   ├── vn_provinces_wards_geojson.zip                      # Combined GeoJSON archive
│   └── geojson/                                            # Per-province GeoJSON
│       ├── README.md
│       ├── 01_ha_noi/
│       │   ├── 01_ha_noi.geojson                           # Province boundary
│       │   └── wards/                                      # Per-ward boundaries
│       │       ├── 00004_ba_dinh.geojson
│       │       └── ...
│       ├── 04_cao_bang/
│       └── ...
├── mongodb/
│   ├── administrative_regions_*.json
│   ├── administrative_units_*.json
│   └── mongo_data_vn_unit_*.json                           # Full MongoDB import
└── redis/
    └── redis_vn_provinces_dataset_*.redis                   # Redis commands
```

(Adjust the leading tree connectors as needed so the `mongodb/` and `redis/` entries remain correctly nested under `output/`.)

- [ ] **Step 2: Update `docs/gis/gis_readme.md` and `docs/gis/gis_readme_vi.md`**

In both files, change the geojson archive reference from `vn_provinces_wards_geojson_<datetime>.zip` to `vn_provinces_wards_geojson.zip`.

- [ ] **Step 3: Verify docs render**

Run: `git diff --stat`
Expected: the 3 doc files listed with diffs. No code build step needed for markdown.

- [ ] **Step 4: Commit**

```bash
git add dataset-generation-scripts/README.md docs/gis/gis_readme.md docs/gis/gis_readme_vi.md
git commit -m "docs: document stable JSON filenames, minified variants, and fixed geojson zip name"
```

---

## Verification (after all tasks)

Run from `dataset-generation-scripts/`:

```bash
go test ./internal/dataset_writer/dataset_file_writer/... -v
```

Expected: all tests pass.

Optional full-generation smoke check (requires Docker Postgres up on `localhost:15432` and internet for GIS): `go run main.go`, then confirm `output/json/` contains the 5 JSON files + `README.md` + `vn_provinces_wards_geojson.zip` (no datetime suffixes) and `output/json/geojson/README.md` is unchanged. Then copy `output/json/*` over `json/` to publish.

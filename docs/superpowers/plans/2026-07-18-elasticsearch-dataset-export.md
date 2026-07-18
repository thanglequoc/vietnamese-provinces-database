# Elasticsearch Dataset Export — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Elasticsearch as a new dataset export format, producing two indices (`provinces` and `provinces-gis`) with denormalized province documents, embedded wards, SearchKeywords, embedded AdministrativeUnit, optional GIS geometry, mapping files, and Bulk API-compatible NDJSON.

**Architecture:** New `ElasticsearchDatasetFileWriter` implementing the existing `DatasetFileWriter` interface pattern. Non-GIS index generated from `provinces`/`wards` models in `ReadAndGenerateSQLDatasets()`. GIS index generated from `SapNhapSiteGeoUnit` records in `GenerateGISSQLDatasets()`. New DTOs in `dto/elasticsearch_dto.go`, new mapper functions in `helper/dto_mapper.go`.

**Tech Stack:** Go 1.24, Bun ORM, `encoding/json`, `bufio`. No external Elasticsearch client library — the project generates files, not live connections.

**Branch:** `165_IncludeDataForElasticSearch`
**Spec:** `docs/superpowers/specs/2026-07-18-elasticsearch-dataset-export-design.md`

## Global Constraints

- No database schema or data changes
- No changes to existing exporters (mongodb, redis, json, sql writers)
- No Elasticsearch client library dependency
- No Elasticsearch container in CI/CD (NDJSON validated structurally)
- Field naming uses PascalCase (consistent with MongoDB/JSON exports)
- The `Meta` field is named without underscore prefix (Elasticsearch reserves `_`-prefixed fields)

---

### Task 1: Create Elasticsearch DTOs

**Files:**
- Create: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/dto/elasticsearch_dto.go`

**Interfaces:**
- Produces: ES document structs used by the writer and mapper
- Depends on: `encoding/json` (for `json.RawMessage` in GIS fields)

- [ ] **Step 1: Create `dto/elasticsearch_dto.go`**

Define the following structs with PascalCase JSON tags:

```go
package dto

import "encoding/json"

// ElasticsearchProvinceDocument represents a province document in the
// provinces or provinces-gis Elasticsearch index.
type ElasticsearchProvinceDocument struct {
    Code               string                        `json:"Code"`
    Name               string                        `json:"Name"`
    NameEn             string                        `json:"NameEn"`
    FullName           string                        `json:"FullName"`
    FullNameEn         string                        `json:"FullNameEn"`
    CodeName           string                        `json:"CodeName"`
    AdministrativeUnit ElasticsearchAdministrativeUnit `json:"AdministrativeUnit"`
    SearchKeywords     []string                      `json:"SearchKeywords"`
    Wards              []ElasticsearchWardDocument   `json:"Wards"`
    GIS                *ElasticsearchGIS             `json:"GIS,omitempty"`
    Meta               *ElasticsearchMeta            `json:"Meta,omitempty"`
}

// ElasticsearchWardDocument represents an embedded ward inside a province document.
type ElasticsearchWardDocument struct {
    Code               string                        `json:"Code"`
    Name               string                        `json:"Name"`
    NameEn             string                        `json:"NameEn"`
    FullName           string                        `json:"FullName"`
    FullNameEn         string                        `json:"FullNameEn"`
    CodeName           string                        `json:"CodeName"`
    AdministrativeUnit ElasticsearchAdministrativeUnit `json:"AdministrativeUnit"`
    SearchKeywords     []string                      `json:"SearchKeywords"`
    GIS                *ElasticsearchGIS             `json:"GIS,omitempty"`
}

// ElasticsearchAdministrativeUnit is the embedded administrative unit object.
type ElasticsearchAdministrativeUnit struct {
    Id          int    `json:"Id"`
    FullName    string `json:"FullName"`
    FullNameEn  string `json:"FullNameEn"`
    ShortName   string `json:"ShortName"`
    ShortNameEn string `json:"ShortNameEn"`
    CodeName    string `json:"CodeName"`
    CodeNameEn  string `json:"CodeNameEn"`
}

// ElasticsearchGIS holds optional GIS data for the provinces-gis index.
type ElasticsearchGIS struct {
    Center      ElasticsearchGeoPoint   `json:"Center"`
    BoundingBox ElasticsearchBoundingBox `json:"BoundingBox"`
    Geometry    json.RawMessage         `json:"Geometry"`
}

// ElasticsearchGeoPoint is a lat/lon point for Elasticsearch geo_point mapping.
type ElasticsearchGeoPoint struct {
    Lat float64 `json:"lat"`
    Lon float64 `json:"lon"`
}

// ElasticsearchBoundingBox holds the bounding box coordinates.
type ElasticsearchBoundingBox struct {
    MinLongitude float64 `json:"MinLongitude"`
    MinLatitude  float64 `json:"MinLatitude"`
    MaxLongitude float64 `json:"MaxLongitude"`
    MaxLatitude  float64 `json:"MaxLatitude"`
}

// ElasticsearchMeta holds dataset version metadata.
type ElasticsearchMeta struct {
    DatasetVersion         string `json:"DatasetVersion"`
    AdministrativeRevision string `json:"AdministrativeRevision"`
    GeneratedAt            string `json:"GeneratedAt"`
}
```

- [ ] **Step 2: Verify the file compiles**

```bash
cd dataset-generation-scripts && go build ./internal/dataset_writer/dataset_file_writer/dto/
```

Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/dto/elasticsearch_dto.go
git commit -m "feat(165): add Elasticsearch DTOs for province/ward documents"
```

---

### Task 2: Add SearchKeywords and DTO mapper functions

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/helper/dto_mapper.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/helper/dto_mapper_test.go`

**Interfaces:**
- Produces: `GenerateSearchKeywords(code, name, nameEn, codeName string) []string` and `ConvertToElasticsearchProvinceModel(provinces []model.Province) []dto.ElasticsearchProvinceDocument`
- Depends on: `internal/common/viet` (for `RemoveVietToneMark`), `model.Province`/`model.Ward`/`model.AdministrativeUnit`

- [ ] **Step 1: Add `GenerateSearchKeywords` function to `helper/dto_mapper.go`**

Add the import for the `viet` package and the function:

```go
import (
    "strings"
    "github.com/thanglequoc-vn-provinces/v2/internal/common/viet"
    // ... existing imports
)

// GenerateSearchKeywords builds a deduplicated keyword array for Elasticsearch
// autocomplete/search. The array contains: code, tone-stripped lowercase name,
// lowercase English name, and codeName.
func GenerateSearchKeywords(code, name, nameEn, codeName string) []string {
    keywords := []string{
        code,
        strings.ToLower(viet.RemoveVietToneMark(name)),
        strings.ToLower(nameEn),
        codeName,
    }
    return deduplicate(keywords)
}

// deduplicate removes duplicate strings from a slice while preserving order.
func deduplicate(items []string) []string {
    seen := make(map[string]bool, len(items))
    result := make([]string, 0, len(items))
    for _, item := range items {
        if !seen[item] {
            seen[item] = true
            result = append(result, item)
        }
    }
    return result
}
```

- [ ] **Step 2: Add `ConvertToElasticsearchProvinceModel` function to `helper/dto_mapper.go`**

```go
func ConvertToElasticsearchProvinceModel(provinces []model.Province) []dataset_file_writer_dto.ElasticsearchProvinceDocument {
    var result []dataset_file_writer_dto.ElasticsearchProvinceDocument
    for _, province := range provinces {
        p := dataset_file_writer_dto.ElasticsearchProvinceDocument{
            Code:     province.Code,
            Name:     province.Name,
            NameEn:   province.NameEn,
            FullName: province.FullName,
            FullNameEn: province.FullNameEn,
            CodeName: province.CodeName,
            AdministrativeUnit: convertToElasticsearchAdministrativeUnit(province.AdministrativeUnit),
            SearchKeywords: GenerateSearchKeywords(province.Code, province.Name, province.NameEn, province.CodeName),
        }

        if len(province.Wards) != 0 {
            wards := make([]model.Ward, len(province.Wards))
            for i, w := range province.Wards {
                wards[i] = *w
            }
            p.Wards = convertToElasticsearchWardDocuments(wards)
        }
        result = append(result, p)
    }
    return result
}

func convertToElasticsearchWardDocuments(wards []model.Ward) []dataset_file_writer_dto.ElasticsearchWardDocument {
    var result []dataset_file_writer_dto.ElasticsearchWardDocument
    for _, ward := range wards {
        w := dataset_file_writer_dto.ElasticsearchWardDocument{
            Code:               ward.Code,
            Name:               ward.Name,
            NameEn:             ward.NameEn,
            FullName:           ward.FullName,
            FullNameEn:         ward.FullNameEn,
            CodeName:           ward.CodeName,
            AdministrativeUnit: convertToElasticsearchAdministrativeUnit(ward.AdministrativeUnit),
            SearchKeywords:     GenerateSearchKeywords(ward.Code, ward.Name, ward.NameEn, ward.CodeName),
        }
        result = append(result, w)
    }
    return result
}

func convertToElasticsearchAdministrativeUnit(au model.AdministrativeUnit) dataset_file_writer_dto.ElasticsearchAdministrativeUnit {
    return dataset_file_writer_dto.ElasticsearchAdministrativeUnit{
        Id:          au.Id,
        FullName:    au.FullName,
        FullNameEn:  au.FullNameEn,
        ShortName:   au.ShortName,
        ShortNameEn: au.ShortNameEn,
        CodeName:    au.CodeName,
        CodeNameEn:  au.CodeNameEn,
    }
}
```

- [ ] **Step 3: Add unit tests to `helper/dto_mapper_test.go`**

Test cases:
- `GenerateSearchKeywords` with Hà Nội: `["01", "ha noi", "hanoi", "ha_noi"]`
- `GenerateSearchKeywords` deduplication: when `NameEn` lowercased equals tone-stripped `Name`
- `ConvertToElasticsearchProvinceModel` with a province that has wards: verify all fields populated
- `ConvertToElasticsearchProvinceModel` with a province that has no wards: verify `Wards` is nil/empty

- [ ] **Step 4: Run tests**

```bash
cd dataset-generation-scripts && go test -v ./internal/dataset_writer/dataset_file_writer/helper/
```

Expected: All tests pass.

- [ ] **Step 5: Commit**

```bash
git add dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/helper/dto_mapper.go
git add dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/helper/dto_mapper_test.go
git commit -m "feat(165): add SearchKeywords generation and ES DTO mapper functions"
```

---

### Task 3: Create the Elasticsearch file writer

**Files:**
- Create: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer.go`

**Interfaces:**
- Produces: `ElasticsearchDatasetFileWriter` struct with `WriteToFile` (non-GIS) and `WriteElasticsearchGISDataToFile` (GIS) methods
- Depends on: Tasks 1 & 2 (DTOs and mapper functions), `SapNhapSiteGeoUnit` model, `json.RawMessage` for GIS geometry

- [ ] **Step 1: Create `elasticsearch_file_writer.go`**

The writer struct:

```go
type ElasticsearchDatasetFileWriter struct {
    OutputFolderPath string
}
```

**Constants:**

```go
const (
    esIndexName          = "provinces"
    esGISIndexName       = "provinces-gis"
    esDatasetVersion     = "2026.07.01"
    esAdminRevision      = "2026-04-30"
)
```

**`WriteToFile` method** (non-GIS index):

1. Create `OutputFolderPath` and `OutputFolderPath/mappings/` directories
2. Call `helper.ConvertToElasticsearchProvinceModel(provinces)` to get documents
3. Add `Meta` to each province document with `time.Now().UTC().Format(time.RFC3339)`
4. Write `provinces.ndjson` — for each document, write two lines:
   - `{"index": {"_index": "provinces", "_id": "<Code>"}}`
   - JSON-marshaled document (compact, no indent)
5. Write `mappings/provinces.json` — static mapping (see spec section 7)
6. Write `README.md` with usage instructions

**`WriteElasticsearchGISDataToFile` method** (GIS index):

1. Reuse `OutputFolderPath` (same `elasticsearch/` folder)
2. Group wards by `VNDSProvinceCode`
3. For each province GIS unit:
   - Build an `ElasticsearchProvinceDocument` from `VNProvince` relation
   - Add `GIS` field: parse `BBoxGeoJSON` array `[xmin, ymin, xmax, ymax]` → `BoundingBox` + `Center`; pass `GeomGeoJSON` as `Geometry`
   - Embed matching ward GIS units as `ElasticsearchWardDocument` with their `GIS` fields
   - Add `Meta`
4. Write `provinces-gis.ndjson` with `_index: "provinces-gis"` and `_id: <province code>`
5. Write `mappings/provinces-gis.json` — static mapping (provinces mapping + GIS fields)

**GIS bbox parsing helper:**

```go
// parseBBox parses a BBoxGeoJSON array [xmin, ymin, xmax, ymax] into
// ElasticsearchBoundingBox and ElasticsearchGeoPoint (center).
func parseBBox(bboxGeoJSON json.RawMessage) (dataset_file_writer_dto.ElasticsearchBoundingBox, dataset_file_writer_dto.ElasticsearchGeoPoint, error) {
    var coords []float64
    if err := json.Unmarshal(bboxGeoJSON, &coords); err != nil {
        return ..., ..., fmt.Errorf("parse bbox geojson: %w", err)
    }
    if len(coords) != 4 {
        return ..., ..., fmt.Errorf("expected 4 bbox coordinates, got %d", len(coords))
    }
    xmin, ymin, xmax, ymax := coords[0], coords[1], coords[2], coords[3]
    bbox := dto.ElasticsearchBoundingBox{
        MinLongitude: xmin,
        MinLatitude:  ymin,
        MaxLongitude: xmax,
        MaxLatitude:  ymax,
    }
    center := dto.ElasticsearchGeoPoint{
        Lat: (ymin + ymax) / 2,
        Lon: (xmin + xmax) / 2,
    }
    return bbox, center, nil
}
```

- [ ] **Step 2: Verify the file compiles**

```bash
cd dataset-generation-scripts && go build ./internal/dataset_writer/dataset_file_writer/
```

Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer.go
git commit -m "feat(165): add ElasticsearchDatasetFileWriter with NDJSON and mapping generation"
```

---

### Task 4: Add unit tests for the Elasticsearch writer

**Files:**
- Create: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer_test.go`

**Interfaces:**
- Consumes: Task 3 (writer implementation)
- Produces: Test coverage for NDJSON format, mapping generation, GIS bbox parsing, README generation

- [ ] **Step 1: Create `elasticsearch_file_writer_test.go`**

Test cases:

1. **`TestWriteToFile_NonGIS`**: Write a small set of provinces/wards to a temp dir. Verify:
   - `provinces.ndjson` exists and has 2 lines per province (index line + doc line)
   - Each index line has `_index: "provinces"` and `_id` matching the province code
   - `mappings/provinces.json` exists and is valid JSON with expected field types
   - `README.md` exists and is non-empty

2. **`TestWriteElasticsearchGISDataToFile_GIS`**: Write a small set of GIS units. Verify:
   - `provinces-gis.ndjson` exists with `_index: "provinces-gis"`
   - Documents contain `GIS` field with `Center`, `BoundingBox`, `Geometry`
   - `mappings/provinces-gis.json` has `geo_point` and `geo_shape` types

3. **`TestParseBBox`**: Verify bbox parsing from `[xmin, ymin, xmax, ymax]` array:
   - Correct `BoundingBox` fields
   - Correct `Center` (midpoint calculation)
   - Error on malformed input (non-array, wrong length)

4. **`TestGenerateSearchKeywords`** (if not already covered in Task 2): Verify keyword generation and deduplication

- [ ] **Step 2: Run tests**

```bash
cd dataset-generation-scripts && go test -v ./internal/dataset_writer/dataset_file_writer/
```

Expected: All tests pass.

- [ ] **Step 3: Commit**

```bash
git add dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer_test.go
git commit -m "test(165): add unit tests for Elasticsearch writer, NDJSON, and GIS bbox parsing"
```

---

### Task 5: Wire Elasticsearch writer into dataset_writer.go

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_writer.go`

**Interfaces:**
- Consumes: Tasks 1-4 (DTOs, mapper, writer, tests)
- Produces: Elasticsearch export integrated into the main generation pipeline

- [ ] **Step 1: Add ES writer to `ReadAndGenerateSQLDatasets()`**

After the Redis writer block (after line 95), add:

```go
// Elasticsearch
elasticsearchDatasetFileWriter := datasetfilewriter.ElasticsearchDatasetFileWriter{
    OutputFolderPath: "./output/elasticsearch",
}
err = elasticsearchDatasetFileWriter.WriteToFile(regions, administrativeUnits, provinces, wards)
if err != nil {
    log.Fatal("Unable to generate Elasticsearch Dataset", err)
} else {
    fmt.Println("✅ Elasticsearch Dataset successfully generated")
}
```

- [ ] **Step 2: Add ES GIS writer to `GenerateGISSQLDatasets()`**

After the GeoJSON writer block (after line 139), add:

```go
// Elasticsearch GIS
elasticsearchGISFileWriter := datasetfilewriter.ElasticsearchDatasetFileWriter{
    OutputFolderPath: "./output/elasticsearch",
}
err = elasticsearchGISFileWriter.WriteElasticsearchGISDataToFile(sapNhapGeoProvinces, sapNhapGeoWards)
if err != nil {
    log.Fatal("Unable to generate Elasticsearch GIS Dataset", err)
} else {
    fmt.Println("✅ Elasticsearch GIS Dataset successfully generated")
}
```

- [ ] **Step 3: Verify compilation**

```bash
cd dataset-generation-scripts && go build ./...
```

Expected: No errors.

- [ ] **Step 4: Commit**

```bash
git add dataset-generation-scripts/internal/dataset_writer/dataset_writer.go
git commit -m "feat(165): wire Elasticsearch writer into dataset generation pipeline"
```

---

### Task 6: Create the top-level `elasticsearch/` output directory and README

**Files:**
- Create: `elasticsearch/README.md`

**Interfaces:**
- Consumes: Tasks 1-5 (the generation pipeline produces `output/elasticsearch/`)
- Produces: Published output directory with user-facing documentation

- [ ] **Step 1: Create `elasticsearch/README.md`**

The README should cover:
- Overview of the Elasticsearch dataset
- Index descriptions (`provinces` vs `provinces-gis`)
- Document structure (fields, nested wards, SearchKeywords, Meta)
- How to create indices with the mapping files:
  ```bash
  curl -X PUT "localhost:9200/provinces" -H 'Content-Type: application/json' -d @mappings/provinces.json
  curl -X PUT "localhost:9200/provinces-gis" -H 'Content-Type: application/json' -d @mappings/provinces-gis.json
  ```
- How to bulk-import NDJSON:
  ```bash
  curl -X POST "localhost:9200/_bulk" -H 'Content-Type: application/x-ndjson' --data-binary @provinces.ndjson
  curl -X POST "localhost:9200/_bulk" -H 'Content-Type: application/x-ndjson' --data-binary @provinces-gis.ndjson
  ```
- Example queries (province dropdown, ward search, GIS spatial query)
- Note about `Meta` field naming (no underscore prefix)

- [ ] **Step 2: Commit**

```bash
git add elasticsearch/README.md
git commit -m "docs(165): add Elasticsearch dataset README with usage instructions"
```

---

### Task 7: End-to-end verification

**Files:**
- Check: `dataset-generation-scripts/output/elasticsearch/` (generated artifacts)
- Check: `elasticsearch/` (published directory)

- [ ] **Step 1: Start the Postgres/PostGIS container**

```bash
cd dataset-generation-scripts && docker compose -f docker/docker-compose.yaml up -d
```

- [ ] **Step 2: Run the full generation pipeline**

```bash
cd dataset-generation-scripts && go run main.go
```

Expected output includes:
```
✅ Elasticsearch Dataset successfully generated
✅ Elasticsearch GIS Dataset successfully generated
```

- [ ] **Step 3: Verify generated files exist**

```bash
ls -la dataset-generation-scripts/output/elasticsearch/
ls -la dataset-generation-scripts/output/elasticsearch/mappings/
```

Expected files:
- `provinces.ndjson`
- `provinces-gis.ndjson`
- `mappings/provinces.json`
- `mappings/provinces-gis.json`
- `README.md`

- [ ] **Step 4: Verify NDJSON format is valid**

```bash
# Check first 4 lines of provinces.ndjson (should be 2 index+doc pairs)
head -4 dataset-generation-scripts/output/elasticsearch/provinces.ndjson

# Verify each line is valid JSON
jq -c . dataset-generation-scripts/output/elasticsearch/provinces.ndjson > /dev/null && echo "NDJSON valid" || echo "NDJSON invalid"
```

- [ ] **Step 5: Verify document count matches province count**

```bash
# Count index lines (should be 34 for 34 provinces)
grep -c '"index"' dataset-generation-scripts/output/elasticsearch/provinces.ndjson
```

Expected: `34`

- [ ] **Step 6: Verify GIS index has GIS fields**

```bash
# Check that provinces-gis documents contain GIS field
grep -o '"GIS"' dataset-generation-scripts/output/elasticsearch/provinces-gis.ndjson | head -1
```

Expected: `"GIS"` appears at least once.

- [ ] **Step 7: Run all tests**

```bash
cd dataset-generation-scripts && go test -v ./...
```

Expected: All tests pass.

- [ ] **Step 8: Copy generated artifacts to top-level `elasticsearch/` directory**

```bash
mkdir -p elasticsearch/mappings
cp dataset-generation-scripts/output/elasticsearch/provinces.ndjson elasticsearch/
cp dataset-generation-scripts/output/elasticsearch/provinces-gis.ndjson elasticsearch/
cp dataset-generation-scripts/output/elasticsearch/mappings/provinces.json elasticsearch/mappings/
cp dataset-generation-scripts/output/elasticsearch/mappings/provinces-gis.json elasticsearch/mappings/
```

- [ ] **Step 9: Commit published artifacts**

```bash
git add elasticsearch/
git commit -m "feat(165): publish Elasticsearch dataset artifacts (NDJSON + mappings)"
```

- [ ] **Step 10: Final git status check**

```bash
git status
```

Expected: Working tree clean (or only untracked files unrelated to this feature).
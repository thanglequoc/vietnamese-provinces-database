# Design: Elasticsearch Dataset Export

**Date:** 2026-07-18
**Status:** Approved
**Approach:** A — Dedicated `ElasticsearchDatasetFileWriter` implementing `DatasetFileWriter` interface
**Branch:** `165_IncludeDataForElasticSearch`
**Source Instruction:** `development/165_SupportElasticSearch/base-instruction.md`

---

## Objective

Extend the Vietnamese Provinces Database to support **Elasticsearch** as an additional export format. The Elasticsearch dataset follows document-oriented design principles: each province is a single denormalized document with all wards embedded as nested objects. This optimizes for the most common use cases — province/ward dropdowns, search, full-text search, and optional GIS spatial queries.

The export produces two indices:
1. **`provinces`** — Province metadata + embedded wards, no GIS geometry
2. **`provinces-gis`** — Same structure + GIS geometry for both provinces and wards

---

## Background

The project currently generates exports for PostgreSQL/MySQL, MSSQL, Oracle, JSON, MongoDB, and Redis. Each format has a dedicated file writer in `internal/dataset_writer/dataset_file_writer/` implementing the `DatasetFileWriter` interface. The MongoDB export (`mongodb_file_writer.go`) is the closest analog: it produces denormalized province documents with embedded wards.

The GIS data pipeline (`GenerateGISSQLDatasets()` in `dataset_writer.go`) fetches `SapNhapSiteGeoUnit` records from the `sapnhap_geojson_objects` table. These records already carry:
- `BBoxGeoJSON` — bounding box as a JSON array `[xmin, ymin, xmax, ymax]` (PostGIS `ST_XMin`/`ST_YMin`/`ST_XMax`/`ST_YMax`)
- `GeomGeoJSON` — geometry as GeoJSON (PostGIS `ST_AsGeoJSON`)
- `VNProvince` / `VNWard` — loaded Bun relations with full administrative metadata

This means the GIS index can be generated from the same data source as the existing GeoJSON export, with no additional database queries.

---

## Design

### 1. New Files

| File | Purpose |
|------|---------|
| `dto/elasticsearch_dto.go` | ES document DTOs: `ElasticsearchProvinceDocument`, `ElasticsearchWardDocument`, `ElasticsearchAdministrativeUnit`, `ElasticsearchGIS`, `ElasticsearchMeta` |
| `elasticsearch_file_writer.go` | `ElasticsearchDatasetFileWriter` struct with `WriteToFile` + `WriteElasticsearchGISDataToFile` methods |
| `elasticsearch_file_writer_test.go` | Unit tests for NDJSON format, mapping generation, GIS bbox parsing |
| `helper/dto_mapper.go` (modify) | Add `ConvertToElasticsearchProvinceModel` and `GenerateSearchKeywords` functions |

### 2. Document Structure

Field naming uses **PascalCase**, consistent with the existing MongoDB and JSON exports. JSON tags use PascalCase to match.

#### Province Document (`provinces` index)

```json
{
  "Code": "01",
  "Name": "Hà Nội",
  "NameEn": "Hanoi",
  "FullName": "Thành phố Hà Nội",
  "FullNameEn": "Hanoi City",
  "CodeName": "ha_noi",
  "AdministrativeUnit": {
    "Id": 1,
    "FullName": "Thành phố",
    "FullNameEn": "City",
    "ShortName": "TP.",
    "ShortNameEn": "City",
    "CodeName": "thanh_pho",
    "CodeNameEn": "city"
  },
  "SearchKeywords": ["01", "ha noi", "hanoi", "ha_noi"],
  "Wards": [
    {
      "Code": "00004",
      "Name": "Ba Đình",
      "NameEn": "Ba Dinh",
      "FullName": "Phường Ba Đình",
      "FullNameEn": "Ba Dinh Ward",
      "CodeName": "ba_dinh",
      "AdministrativeUnit": {
        "Id": 3,
        "FullName": "Phường",
        "FullNameEn": "Ward",
        "ShortName": "P.",
        "ShortNameEn": "Ward",
        "CodeName": "phuong",
        "CodeNameEn": "ward"
      },
      "SearchKeywords": ["00004", "ba dinh", "ba_dinh"]
    }
  ],
  "Meta": {
    "DatasetVersion": "2026.07.01",
    "AdministrativeRevision": "2026-04-30",
    "GeneratedAt": "2026-07-18T08:00:00Z"
  }
}
```

#### Ward Document (embedded, same in both indices)

Same fields as the province document minus `Wards` and `Meta`. The `SearchKeywords` for wards follows the same logic.

#### GIS Fields (`provinces-gis` index only)

Both province and ward documents gain an optional `GIS` object:

```json
"GIS": {
  "Center": {
    "lat": 21.0285,
    "lon": 105.8542
  },
  "BoundingBox": {
    "MinLongitude": 105.285,
    "MinLatitude": 20.845,
    "MaxLongitude": 106.024,
    "MaxLatitude": 21.212
  },
  "Geometry": {
    "type": "MultiPolygon",
    "coordinates": [...]
  }
}
```

- **Center**: Derived from bbox center: `lat = (ymin + ymax) / 2`, `lon = (xmin + xmax) / 2`
- **BoundingBox**: Parsed from the existing `BBoxGeoJSON` array `[xmin, ymin, xmax, ymax]`
- **Geometry**: The existing `GeomGeoJSON` (already GeoJSON format from PostGIS `ST_AsGeoJSON`), passed through as-is via `json.RawMessage`

### 3. SearchKeywords Logic

Each province and ward document includes a `SearchKeywords` array containing:

| Position | Source | Transformation | Example (Hà Nội) |
|----------|--------|----------------|-------------------|
| 1 | `Code` | As-is | `"01"` |
| 2 | `Name` (Vietnamese) | `RemoveVietToneMark(name)` → lowercase | `"ha noi"` |
| 3 | `NameEn` | lowercase | `"hanoi"` |
| 4 | `CodeName` | As-is | `"ha_noi"` |

The array is **deduplicated** (preserving order) to handle cases where variants collide (e.g., a ward where `NameEn` lowercased equals the tone-stripped `Name`).

Uses the existing `viet.RemoveVietToneMark()` utility from `internal/common/viet/remove_tone_mark.go`.

### 4. AdministrativeUnit Embedding

Per the base instruction, the complete `AdministrativeUnit` object is embedded (not just the `AdministrativeUnitId`). This makes every document self-contained. The `AdministrativeUnit` model (`vn_provinces_tmp_model.go`) already has all fields: `Id`, `FullName`, `FullNameEn`, `ShortName`, `ShortNameEn`, `CodeName`, `CodeNameEn`.

### 5. Meta Field

**Deviation from base instruction:** The base instruction specifies a `_meta` field. However, Elasticsearch reserves field names starting with `_` and will reject them in document sources. The field is named `Meta` (no underscore prefix) instead. The mapping documents this as the dataset metadata field.

The `Meta` object contains:
- `DatasetVersion`: Semantic version string (e.g., `"2026.07.01"`)
- `AdministrativeRevision`: Date of the latest administrative decree in effect (e.g., `"2026-04-30"`)
- `GeneratedAt`: ISO 8601 timestamp of generation (e.g., `"2026-07-18T08:00:00Z"`)

These values are set at generation time. The `DatasetVersion` and `AdministrativeRevision` are constants defined in the writer; `GeneratedAt` uses `time.Now().UTC()`.

### 6. NDJSON Bulk API Format

The NDJSON files are directly compatible with the Elasticsearch Bulk API. Each document consists of two lines:

```
{"index": {"_index": "provinces", "_id": "01"}}
{"Code": "01", "Name": "Hà Nội", ...}
{"index": {"_index": "provinces", "_id": "02"}}
{"Code": "02", ...}
```

For the GIS index, `_index` is `"provinces-gis"`. The `_id` is the province code.

### 7. Mapping Files

Static JSON mapping files are generated by the writer.

#### `mappings/provinces.json`

| Field | Type | Notes |
|-------|------|-------|
| `Code` | keyword | Exact match, aggregation |
| `CodeName` | keyword | Exact match |
| `SearchKeywords` | keyword | Array of keywords for autocomplete |
| `Name` | text + keyword subfield | Full-text + exact |
| `NameEn` | text + keyword subfield | Full-text + exact |
| `FullName` | text | Full-text only |
| `FullNameEn` | text | Full-text only |
| `AdministrativeUnit` | object | Embedded object with its own properties |
| `AdministrativeUnit.Id` | integer | |
| `AdministrativeUnit.FullName` | keyword | |
| `AdministrativeUnit.ShortName` | keyword | |
| `Wards` | nested | Nested type for ward-level queries |
| `Wards.Code` | keyword | |
| `Wards.CodeName` | keyword | |
| `Wards.SearchKeywords` | keyword | |
| `Wards.Name` | text + keyword | |
| `Wards.NameEn` | text + keyword | |
| `Wards.FullName` | text | |
| `Wards.FullNameEn` | text | |
| `Wards.AdministrativeUnit` | object | |
| `Meta` | object | |
| `Meta.DatasetVersion` | keyword | |
| `Meta.AdministrativeRevision` | keyword | |
| `Meta.GeneratedAt` | date | |

#### `mappings/provinces-gis.json`

Same as above, plus:

| Field | Type | Notes |
|-------|------|-------|
| `GIS` | object | |
| `GIS.Center` | geo_point | Lat/lon point |
| `GIS.BoundingBox` | object | Min/Max longitude/latitude |
| `GIS.Geometry` | geo_shape | Full GeoJSON geometry |
| `Wards.GIS` | object | Same structure as province GIS |
| `Wards.GIS.Center` | geo_point | |
| `Wards.GIS.BoundingBox` | object | |
| `Wards.GIS.Geometry` | geo_shape | |

### 8. Writer Integration

#### `ReadAndGenerateSQLDatasets()` (in `dataset_writer.go`)

Add after the Redis writer block:

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

This generates the `provinces` index (NDJSON + mapping + README).

#### `GenerateGISSQLDatasets()` (in `dataset_writer.go`)

Add after the GeoJSON writer block:

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

This generates the `provinces-gis` index (NDJSON + mapping). It reuses the same `OutputFolderPath` so both indices land in the same `elasticsearch/` folder.

### 9. Output Structure

```
elasticsearch/
├── README.md
├── mappings/
│   ├── provinces.json
│   └── provinces-gis.json
├── provinces.ndjson
└── provinces-gis.ndjson
```

The `README.md` explains the dataset, how to create indices with the mappings, and how to bulk-import the NDJSON files.

### 10. Published Output

After generation, the artifacts are copied to the top-level `elasticsearch/` directory (consistent with `mongodb/`, `redis/`, `json/`, etc.). The generation script's `output/elasticsearch/` is the staging area; the top-level folder is the published artifact.

---

## Files Affected

| File | Action | Details |
|------|--------|---------|
| `internal/dataset_writer/dataset_file_writer/dto/elasticsearch_dto.go` | Create | ES document DTOs |
| `internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer.go` | Create | Writer implementation |
| `internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer_test.go` | Create | Unit tests |
| `internal/dataset_writer/dataset_file_writer/helper/dto_mapper.go` | Modify | Add `ConvertToElasticsearchProvinceModel` + `GenerateSearchKeywords` |
| `internal/dataset_writer/dataset_file_writer/helper/dto_mapper_test.go` | Modify | Add tests for ES mapper |
| `internal/dataset_writer/dataset_writer.go` | Modify | Wire ES writer into both generation functions |
| `elasticsearch/` (top-level) | Create | Published output directory |

---

## Risks & Mitigations

| Risk | Severity | Mitigation |
|------|----------|------------|
| NDJSON file size with GIS geometry | Medium | GIS geometry can be large; NDJSON is line-delimited so it streams well. The `provinces-gis.ndjson` may be large but is consistent with existing GeoJSON zip exports. |
| `Meta` field name deviation from base instruction | Low | Documented in spec and README. Elasticsearch rejects `_`-prefixed fields in document sources. |
| SearchKeywords deduplication edge cases | Low | Dedup preserves order; collisions (e.g., `NameEn` == tone-stripped `Name`) are handled. |
| GIS data missing for some wards | Low | The GIS writer already validates that all 3,355 records have bbox and geom. The ES writer will skip GIS fields if `BBoxGeoJSON` or `GeomGeoJSON` is empty (defensive). |
| Mapping file correctness | Low | Mappings are static JSON generated by the writer; validated by unit tests. |

---

## Out of Scope

- No Elasticsearch container in CI/CD (NDJSON is validated structurally, not by live ES)
- No changes to the database schema or data
- No changes to existing exporters
- No Elasticsearch client library dependency (the project generates files, not live connections)
- No Kibana dashboards or index templates beyond the mapping files
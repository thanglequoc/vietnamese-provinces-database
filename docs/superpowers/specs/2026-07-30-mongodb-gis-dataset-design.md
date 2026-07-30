# MongoDB GIS Dataset Support — Design Spec

**Date:** 2026-07-30  
**Branch:** `178_IncludeGISDataMongoDB` (from `165_IncludeDataForElasticSearch`)  
**Status:** Draft — awaiting review

---

## 1. Objective

Implement a MongoDB GIS dataset that mirrors the Elasticsearch GIS dataset architecture. The standard MongoDB dataset (administrative provinces/wards without geometry) must remain separate from the GIS dataset, following the same separation principle used for Elasticsearch (`provinces` vs `provinces-gis` indices).

## 2. Background

The Elasticsearch GIS implementation provides:
- A dedicated `provinces-gis` index with full GeoJSON geometry
- Version metadata (`Meta`)
- A GIS document generation pipeline that takes `SapNhapSiteGeoUnit` data
- GeoJSON serialization for geometry and bounding box
- Separation between the standard dataset and the GIS dataset

The existing MongoDB writer (`MongoDBDatasetFileWriter`) currently:
- Has only `WriteToFile()` — no GIS method
- Writes 3 files: `administrative_units.json`, `administrative_regions.json`, `mongo_data_vn_unit.json`
- Uses simple DTOs (`MongoProvinceModel`/`MongoWardModel`) with no JSON tags, no SearchKeywords, no Meta, no GIS
- Is not called in `GenerateGISSQLDatasets()`

## 3. MongoDB Collection Structure

| Collection | Documents | Description |
|------------|-----------|-------------|
| `provinces` | 34 | Standard province documents with embedded wards (no GIS) — unchanged |
| `provinces-gis` | 34 | Same structure + GIS geometry for provinces and wards — NEW |

The GIS dataset is a **separate file** (`mongo_data_vn_unit_gis.json`), following the same separation principle as Elasticsearch.

## 4. GIS Document Schema

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
    "FullName": "Thành phố trực thuộc trung ương",
    "FullNameEn": "Municipality",
    "ShortName": "Thành phố",
    "ShortNameEn": "City",
    "CodeName": "thanh_pho_truc_thuoc_trung_uong",
    "CodeNameEn": "municipality"
  },
  "SearchKeywords": ["01", "ha noi", "hanoi", "ha_noi"],
  "GIS": {
    "Center": { "type": "Point", "coordinates": [105.8542, 21.0285] },
    "BoundingBox": {
      "MinLongitude": 105.2859,
      "MinLatitude": 20.4863,
      "MaxLongitude": 106.0617,
      "MaxLatitude": 21.3851
    },
    "Geometry": {
      "type": "MultiPolygon",
      "coordinates": [[[[105.2859, 21.3851], ...]]]
    },
    "Properties": {
      "Code": "01",
      "Name": "Hà Nội",
      "NameEn": "Hanoi",
      "FullName": "Thành phố Hà Nội",
      "FullNameEn": "Hanoi City",
      "CodeName": "ha_noi",
      "GisServerId": "diaphanhanhchinhcaptinh_sn.108",
      "AreaKm2": 3359.84
    }
  },
  "Wards": [
    {
      "Code": "00004",
      "Name": "Ba Đình",
      "NameEn": "Ba Dinh",
      "FullName": "Phường Ba Đình",
      "FullNameEn": "Ba Dinh Ward",
      "CodeName": "ba_dinh",
      "AdministrativeUnit": { "Id": 3, "FullName": "Phường", ... },
      "SearchKeywords": ["00004", "ba dinh", "ba_dinh"],
      "GIS": {
        "Center": { "type": "Point", "coordinates": [105.8231, 21.0347] },
        "BoundingBox": { "MinLongitude": 105.8115, ... },
        "Geometry": { "type": "Polygon", "coordinates": [[[105.8115, 21.0433], ...]] },
        "Properties": {
          "Code": "00004",
          "Name": "Ba Đình",
          "GisServerId": "diaphanhanhchinhphuong_sn.456",
          "AreaKm2": 5.23
        }
      }
    }
  ],
  "Meta": {
    "DatasetVersion": "2026.07.01",
    "AdministrativeRevision": "2026-04-30",
    "GeneratedAt": "2026-07-25T13:55:11Z"
  }
}
```

### Key Design Decisions

- **Center**: MongoDB-native GeoJSON Point `{ "type": "Point", "coordinates": [lon, lat] }` — enables 2dsphere indexing. This differs from Elasticsearch's `{ "lat": ..., "lon": ... }` format because MongoDB requires GeoJSON format for 2dsphere indexes.
- **Geometry**: Already GeoJSON from the `SapNhapSiteGeoUnit.GeomGeoJSON` data source — stored directly as-is (`json.RawMessage`).
- **BoundingBox**: Stored as a structured object with named fields (same as ES) for readability. Not indexed with 2dsphere (it's a convenience field; spatial queries use Geometry).
- **Field names**: PascalCase (consistent with Elasticsearch and existing MongoDB exports).
- **SearchKeywords**: Reuse existing `GenerateSearchKeywords()` helper from `helper/dto_mapper.go`.
- **Meta**: Same version metadata structure as Elasticsearch.

## 5. Required Geospatial Indexes

A `create_indexes.js` script will be generated alongside the data file:

```javascript
// provinces-gis collection indexes
db.provinces_gis.createIndex({ "Code": 1 }, { unique: true });
db.provinces_gis.createIndex({ "GIS.Geometry": "2dsphere" });
db.provinces_gis.createIndex({ "GIS.Center": "2dsphere" });
db.provinces_gis.createIndex({ "SearchKeywords": 1 });
db.provinces_gis.createIndex({ "Wards.GIS.Geometry": "2dsphere" });
db.provinces_gis.createIndex({ "Wards.Code": 1 });
```

These indexes enable:
- Unique province lookup by Code
- Spatial queries: `$geoIntersects`, `$geoWithin` on province and ward geometries
- Proximity searches: `$near`, `$nearSphere` on province centers
- Keyword-based autocomplete search
- Ward lookup by code within provinces

## 6. Dataset Generation Workflow

```
GenerateGISSQLDatasets() in dataset_writer.go
  ├── Postgres/MySQL GIS (existing)
  ├── MSSQL GIS (existing)
  ├── GeoJSON (existing)
  ├── Elasticsearch GIS (existing)
  └── MongoDB GIS Writer (NEW)
       ├── Groups wards by province code
       ├── Builds MongoGISProvinceDocument with embedded wards + GIS
       ├── Writes mongo_data_vn_unit_gis.json
       ├── Writes create_indexes.js
       └── Writes README.md
```

## 7. Reuse Opportunities

| Component | Source | Reuse Strategy |
|-----------|--------|----------------|
| `parseBBox()` | `elasticsearch_file_writer.go` | Extract to shared helper or duplicate (small function) |
| `GenerateSearchKeywords()` | `helper/dto_mapper.go` | Direct reuse — already shared |
| Admin unit conversion | `helper/dto_mapper.go` | New `convertToMongoAdministrativeUnit()` following ES pattern |
| GIS document assembly loop | `elasticsearch_file_writer.go` | Pattern reuse — same ward-grouping-by-province logic |
| `SapNhapSiteGeoUnit` model | `sapnhap_bando/model` | Direct reuse — same data source |

**Note on `parseBBox()`:** This function is currently a private function in `elasticsearch_file_writer.go`. To avoid duplication, we will either:
- **Option A (recommended):** Duplicate the small function in the MongoDB GIS writer (it's ~15 lines, keeps writers independent)
- **Option B:** Extract to a shared helper package (more refactoring, affects ES writer)

We recommend Option A for simplicity and to avoid touching the existing ES writer.

## 8. Required Code Changes

### New Files

1. **`dto/mongo_gis_dto.go`** — MongoDB GIS document DTOs:
   - `MongoGISProvinceDocument` (Code, Name, NameEn, FullName, FullNameEn, CodeName, AdministrativeUnit, SearchKeywords, Wards, GIS, Meta)
   - `MongoGISWardDocument` (same minus Meta)
   - `MongoGIS` (Center, BoundingBox, Geometry, Properties)
   - `MongoGeoPoint` (type + coordinates — GeoJSON Point)
   - `MongoBoundingBox` (MinLongitude, MinLatitude, MaxLongitude, MaxLatitude)
   - `MongoGISProperties` (Code, Name, NameEn, FullName, FullNameEn, CodeName, GisServerId, AreaKm2)
   - `MongoAdministrativeUnit` (Id, FullName, FullNameEn, ShortName, ShortNameEn, CodeName, CodeNameEn)
   - `MongoMeta` (DatasetVersion, AdministrativeRevision, GeneratedAt)

2. **`helper/mongo_gis_mapper.go`** — Conversion helpers:
   - `ConvertToMongoGISProvinceDocument()` — converts SapNhapSiteGeoUnit provinces + wards to GIS documents
   - `sapnhapGeoUnitToMongoGIS()` — converts a SapNhapSiteGeoUnit's BBoxGeoJSON and GeomGeoJSON into MongoGIS
   - `parseBBoxForMongo()` — parses bbox GeoJSON array into MongoBoundingBox + MongoGeoPoint (center)
   - `convertToMongoAdministrativeUnit()` — converts model.AdministrativeUnit to MongoAdministrativeUnit

3. **`mongodb_gis_file_writer.go`** — GIS writer:
   - `WriteMongoGISDataToFile(sapNhapGeoProvinces, sapNhapGeoWards []*SapNhapSiteGeoUnit) error`
   - Generates `mongo_data_vn_unit_gis.json` (JSON array of province documents)
   - Generates `create_indexes.js` (index creation script)
   - Generates `README.md` (usage documentation)

### Modified Files

1. **`dataset_writer.go`** — Add MongoDB GIS writer call in `GenerateGISSQLDatasets()`:
   ```go
   // MongoDB GIS
   mongoDBGISFileWriter := datasetfilewriter.MongoDBDatasetFileWriter{
       OutputFolderPath: "./output/mongodb",
   }
   err = mongoDBGISFileWriter.WriteMongoGISDataToFile(sapNhapGeoProvinces, sapNhapGeoWards)
   ```

### Unchanged Files

- `mongodb_file_writer.go` — existing writer, backward compatible
- `dto/mongo_dto.go` — existing DTOs, untouched
- Database models, repositories — no changes

## 9. Backward Compatibility

- The standard MongoDB dataset (`mongo_data_vn_unit.json`) remains unchanged
- The GIS dataset is a **new separate file** (`mongo_data_vn_unit_gis.json`)
- No changes to existing `WriteToFile()` method signature
- New method `WriteMongoGISDataToFile()` follows the same pattern as `WriteElasticsearchGISDataToFile()`
- Existing tests remain unaffected

## 10. Performance & Storage Implications

- GIS JSON file will be large (~50-100MB) due to full polygon geometries for 34 provinces + 3,321 wards
- MongoDB's 16MB document size limit is per-document; each province document with all its ward geometries should stay under this limit (largest province ~5-8MB based on ES data)
- The `mongo_data_vn_unit_gis.json` file will be a JSON array of 34 province documents
- 2dsphere indexes on Geometry enable efficient spatial queries (`$geoIntersects`, `$geoWithin`)
- Import time: `mongoimport` can handle the file; for very large files, chunked import may be needed (same as ES NDJSON chunking)

## 11. Testing Strategy

- **Unit tests** for `helper/mongo_gis_mapper.go`: test `ConvertToMongoGISProvinceDocument()`, `sapnhapGeoUnitToMongoGIS()`, `parseBBoxForMongo()`
- **Unit tests** for `mongodb_gis_file_writer.go`: test `WriteMongoGISDataToFile()` with mock data, verify output file structure
- **Integration**: Run `go run main.go` and verify the generated `mongo_data_vn_unit_gis.json` file
- Follow existing test patterns from `elasticsearch_file_writer_test.go`

## 12. Implementation Steps (Ordered)

1. Create `dto/mongo_gis_dto.go` with MongoDB GIS document types
2. Create `helper/mongo_gis_mapper.go` with conversion functions
3. Create `mongodb_gis_file_writer.go` with `WriteMongoGISDataToFile()` method
4. Add MongoDB GIS writer call in `dataset_writer.go` `GenerateGISSQLDatasets()`
5. Write unit tests for the new GIS mapper and writer
6. Run `go test` to verify
7. Run `go run main.go` to generate the actual MongoDB GIS dataset
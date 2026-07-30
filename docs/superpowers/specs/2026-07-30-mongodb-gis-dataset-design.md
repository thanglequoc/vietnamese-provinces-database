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

MongoDB works best with smaller, granular documents. Unlike Elasticsearch's nested type, MongoDB's nested array 2dsphere indexes have limitations and large embedded arrays hurt query performance. Therefore, we use **separate collections** for province GIS and ward GIS.

| Collection | Documents | Description |
|------------|-----------|-------------|
| `provinces` | 34 | Standard province documents with embedded wards (no GIS) — unchanged |
| `provinces-gis` | 34 | Province documents with province-level GIS only (no embedded wards) — NEW |
| `wards-gis` | 3,321 | Standalone ward documents with ward-level GIS + `ProvinceCode` reference — NEW |

The GIS datasets are **separate files**, following the same separation principle as Elasticsearch.

### Design Rationale: Separate Collections (Option B)

- **No 16MB document size risk**: Hà Nội has 500+ wards with full polygon geometry. Embedding all ward geometries in a single province document risks exceeding MongoDB's 16MB limit.
- **Efficient ward-level spatial queries**: Dedicated `wards-gis` collection with its own 2dsphere index enables fast `$geoIntersects` queries (e.g., "which ward contains this point?").
- **MongoDB-idiomatic**: Normalized documents are the recommended pattern for MongoDB, unlike Elasticsearch's nested type.
- **Better index performance**: Indexes on top-level fields are more efficient than indexes on nested array fields.
- **Easier updates**: Individual ward documents can be updated without touching the parent province document.

## 4. GIS Document Schema

### Province GIS Document (`provinces-gis` collection)

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
  "Meta": {
    "DatasetVersion": "2026.07.01",
    "AdministrativeRevision": "2026-04-30",
    "GeneratedAt": "2026-07-25T13:55:11Z"
  }
}
```

### Ward GIS Document (`wards-gis` collection)

```json
{
  "Code": "00004",
  "Name": "Ba Đình",
  "NameEn": "Ba Dinh",
  "FullName": "Phường Ba Đình",
  "FullNameEn": "Ba Dinh Ward",
  "CodeName": "ba_dinh",
  "ProvinceCode": "01",
  "AdministrativeUnit": {
    "Id": 3,
    "FullName": "Phường",
    "FullNameEn": "Ward",
    "ShortName": "Phường",
    "ShortNameEn": "Ward",
    "CodeName": "phuong",
    "CodeNameEn": "ward"
  },
  "SearchKeywords": ["00004", "ba dinh", "ba_dinh"],
  "GIS": {
    "Center": { "type": "Point", "coordinates": [105.8231, 21.0347] },
    "BoundingBox": {
      "MinLongitude": 105.8115,
      "MinLatitude": 21.0261,
      "MaxLongitude": 105.8347,
      "MaxLatitude": 21.0433
    },
    "Geometry": {
      "type": "Polygon",
      "coordinates": [[[105.8115, 21.0433], ...]]
    },
    "Properties": {
      "Code": "00004",
      "Name": "Ba Đình",
      "NameEn": "Ba Dinh",
      "FullName": "Phường Ba Đình",
      "FullNameEn": "Ba Dinh Ward",
      "CodeName": "ba_dinh",
      "GisServerId": "diaphanhanhchinhphuong_sn.456",
      "AreaKm2": 5.23
    }
  },
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
- **ProvinceCode**: Ward GIS documents include a `ProvinceCode` field for cross-collection joins (`$lookup`).

## 5. Required Geospatial Indexes

A `create_indexes.js` script will be generated alongside the data files:

```javascript
// provinces-gis collection indexes
db.provinces_gis.createIndex({ "Code": 1 }, { unique: true });
db.provinces_gis.createIndex({ "GIS.Geometry": "2dsphere" });
db.provinces_gis.createIndex({ "GIS.Center": "2dsphere" });
db.provinces_gis.createIndex({ "SearchKeywords": 1 });

// wards-gis collection indexes
db.wards_gis.createIndex({ "Code": 1 }, { unique: true });
db.wards_gis.createIndex({ "ProvinceCode": 1 });
db.wards_gis.createIndex({ "GIS.Geometry": "2dsphere" });
db.wards_gis.createIndex({ "GIS.Center": "2dsphere" });
db.wards_gis.createIndex({ "SearchKeywords": 1 });
```

These indexes enable:
- Unique province/ward lookup by Code
- Spatial queries: `$geoIntersects`, `$geoWithin` on province and ward geometries
- Proximity searches: `$near`, `$nearSphere` on province/ward centers
- Keyword-based autocomplete search
- Ward lookup by province code (for `$lookup` joins)

## 6. Dataset Generation Workflow

```
GenerateGISSQLDatasets() in dataset_writer.go
  ├── Postgres/MySQL GIS (existing)
  ├── MSSQL GIS (existing)
  ├── GeoJSON (existing)
  ├── Elasticsearch GIS (existing)
  └── MongoDB GIS Writer (NEW)
       ├── Builds MongoGISProvinceDocument for each province (with GIS, no embedded wards)
       ├── Builds MongoGISWardDocument for each ward (with GIS + ProvinceCode)
       ├── Writes mongo_data_vn_province_gis.json (34 province documents)
       ├── Writes mongo_data_vn_ward_gis.json (3,321 ward documents)
       ├── Writes create_indexes.js
       └── Writes README.md
```

## 7. Reuse Opportunities

| Component | Source | Reuse Strategy |
|-----------|--------|----------------|
| `parseBBox()` | `elasticsearch_file_writer.go` | Duplicate the small function (Option A — keeps writers independent) |
| `GenerateSearchKeywords()` | `helper/dto_mapper.go` | Direct reuse — already shared |
| Admin unit conversion | `helper/dto_mapper.go` | New `convertToMongoAdministrativeUnit()` following ES pattern |
| GIS document assembly | `elasticsearch_file_writer.go` | Pattern reuse — similar geo unit → GIS conversion |
| `SapNhapSiteGeoUnit` model | `sapnhap_bando/model` | Direct reuse — same data source |

## 8. Required Code Changes

### New Files

1. **`dto/mongo_gis_dto.go`** — MongoDB GIS document DTOs:
   - `MongoGISProvinceDocument` (Code, Name, NameEn, FullName, FullNameEn, CodeName, AdministrativeUnit, SearchKeywords, GIS, Meta)
   - `MongoGISWardDocument` (Code, Name, NameEn, FullName, FullNameEn, CodeName, ProvinceCode, AdministrativeUnit, SearchKeywords, GIS, Meta)
   - `MongoGIS` (Center, BoundingBox, Geometry, Properties)
   - `MongoGeoPoint` (type + coordinates — GeoJSON Point)
   - `MongoBoundingBox` (MinLongitude, MinLatitude, MaxLongitude, MaxLatitude)
   - `MongoGISProperties` (Code, Name, NameEn, FullName, FullNameEn, CodeName, GisServerId, AreaKm2)
   - `MongoAdministrativeUnit` (Id, FullName, FullNameEn, ShortName, ShortNameEn, CodeName, CodeNameEn)
   - `MongoMeta` (DatasetVersion, AdministrativeRevision, GeneratedAt)

2. **`helper/mongo_gis_mapper.go`** — Conversion helpers:
   - `ConvertToMongoGISProvinceDocuments()` — converts SapNhapSiteGeoUnit provinces to GIS documents
   - `ConvertToMongoGISWardDocuments()` — converts SapNhapSiteGeoUnit wards to GIS documents
   - `sapnhapGeoUnitToMongoGIS()` — converts a SapNhapSiteGeoUnit's BBoxGeoJSON and GeomGeoJSON into MongoGIS
   - `parseBBoxForMongo()` — parses bbox GeoJSON array into MongoBoundingBox + MongoGeoPoint (center)
   - `convertToMongoAdministrativeUnit()` — converts model.AdministrativeUnit to MongoAdministrativeUnit

3. **`mongodb_gis_file_writer.go`** — GIS writer:
   - `WriteMongoGISDataToFile(sapNhapGeoProvinces, sapNhapGeoWards []*SapNhapSiteGeoUnit) error`
   - Generates `mongo_data_vn_province_gis.json` (JSON array of 34 province documents)
   - Generates `mongo_data_vn_ward_gis.json` (JSON array of 3,321 ward documents)
   - **Chunking**: If either file exceeds 50MB, splits into `*_part_01.json`, `*_part_02.json`, etc. + manifest file
   - Generates `create_indexes.js` (index creation script for both collections)
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
- The GIS datasets are **new separate files** (`mongo_data_vn_province_gis.json`, `mongo_data_vn_ward_gis.json`)
- No changes to existing `WriteToFile()` method signature
- New method `WriteMongoGISDataToFile()` follows the same pattern as `WriteElasticsearchGISDataToFile()`
- Existing tests remain unaffected

## 10. Performance & Storage Implications

- Province GIS file: ~10-20MB (34 documents with large MultiPolygon geometries)
- Ward GIS file: ~40-80MB (3,321 documents with polygon geometries)
- **Chunking**: Each output file must be kept under 50MB. If the ward GIS file exceeds 50MB, the writer will split into multiple chunk files (e.g., `mongo_data_vn_ward_gis_part_01.json`, `part_02.json`, etc.), each containing a JSON array of ward documents. A manifest file (`mongo_data_vn_ward_gis.manifest`) will list the chunk filenames in order. This follows the same pattern as the Elasticsearch `writeChunkedNDJSON()` function.
- Individual documents are small (well under MongoDB's 16MB limit) since wards are stored as separate documents
- 2dsphere indexes on Geometry enable efficient spatial queries (`$geoIntersects`, `$geoWithin`)
- Import time: `mongoimport` can handle each file; for chunked files, the manifest enables automated sequential import

## 11. Testing Strategy

- **Unit tests** for `helper/mongo_gis_mapper.go`: test `ConvertToMongoGISProvinceDocuments()`, `ConvertToMongoGISWardDocuments()`, `sapnhapGeoUnitToMongoGIS()`, `parseBBoxForMongo()`
- **Unit tests** for `mongodb_gis_file_writer.go`: test `WriteMongoGISDataToFile()` with mock data, verify output file structure and chunking
- **Integration**: Run `go run main.go` and verify the generated GIS JSON files
- Follow existing test patterns from `elasticsearch_file_writer_test.go`

## 12. Implementation Steps (Ordered)

1. Create `dto/mongo_gis_dto.go` with MongoDB GIS document types
2. Create `helper/mongo_gis_mapper.go` with conversion functions
3. Create `mongodb_gis_file_writer.go` with `WriteMongoGISDataToFile()` method (including chunking logic)
4. Add MongoDB GIS writer call in `dataset_writer.go` `GenerateGISSQLDatasets()`
5. Write unit tests for the new GIS mapper and writer
6. Run `go test` to verify
7. Run `go run main.go` to generate the actual MongoDB GIS dataset
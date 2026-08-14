# Elasticsearch Dataset — Vietnamese Provinces Database

**Generated at: Fri, 14 Aug 2026 09:21:21 +0700**

Provinces and wards as Elasticsearch documents in two indices: `provinces` (no geometry) and `provinces-gis` (with GIS geometry).

## Files

- `provinces.ndjson` — Bulk API NDJSON for the provinces index (1.18 MB)
- `mappings/provinces.json` — Index mapping for provinces (3.01 KB)

## Overview

This dataset provides Vietnamese provinces and wards in Elasticsearch document format with two indices:

| Index | Documents | Description |
|-------|-----------|-------------|
| `provinces` | 34 | Provincial metadata with embedded wards, search keywords, and administrative unit data (no GIS geometry) |
| `provinces-gis` | 34 | Same structure plus GIS geometry for both provinces and wards (bounding boxes + GeoJSON polygons) |

## Data Structure

Each province is a single denormalized document with:

- **Core fields**: `Code`, `Name`, `NameEn`, `FullName`, `FullNameEn`, `CodeName`
- **`AdministrativeUnit`**: embedded administrative unit object (Id, FullName, ShortName, CodeName, ...)
- **`SearchKeywords`**: pre-computed autocomplete keywords (code, tone-stripped name, English name, codeName)
- **`Wards`**: nested array of ward documents with the same field shape (plus `PostalCode`)
- **`Meta`**: `DatasetVersion`, `AdministrativeRevision`, `GeneratedAt`
- **`GIS`**: (provinces-gis only) `Center` (geo_point), `BoundingBox`, `Geometry` (geo_shape), `Properties`

## Sample Document

```json
{
  "Code": "01",
  "Name": "Hà Nội",
  "NameEn": "Ha Noi",
  "FullName": "Thành phố Hà Nội",
  "FullNameEn": "Ha Noi City",
  "CodeName": "ha_noi",
  "AdministrativeUnit": { "Id": 1, "FullName": "Thành phố trực thuộc trung ương", "ShortName": "Thành phố" },
  "SearchKeywords": ["01", "ha noi", "ha_noi"],
  "Wards": [
    { "Code": "00004", "Name": "Ba Đình", "FullName": "Phường Ba Đình", "PostalCode": "11120" }
  ]
}
```

## Quick Start

1. Create the indices with the mappings in `mappings/`.

```bash
curl -X PUT "localhost:9200/provinces" -H 'Content-Type: application/json' -d @mappings/provinces.json
curl -X PUT "localhost:9200/provinces-gis" -H 'Content-Type: application/json' -d @mappings/provinces-gis.json
```

2. Bulk import `provinces.ndjson`, and the `provinces-gis-part-*.ndjson` chunks in order (per `provinces-gis.ndjson.manifest`):

```bash
curl -X POST "localhost:9200/_bulk" -H 'Content-Type: application/x-ndjson' --data-binary @provinces.ndjson
curl -X POST "localhost:9200/_bulk" -H 'Content-Type: application/x-ndjson' --data-binary @provinces-gis-part-01.ndjson
```

3. Verify: 34 documents in each index.

## Sample Queries

```json
// Count documents
POST /provinces/_count

// Autocomplete search
POST /provinces/_search
{ "query": { "terms": { "SearchKeywords": ["ha noi"] } }, "_source": ["Code", "Name", "NameEn"] }

// Search a ward and return the matched nested document only
POST /provinces/_search
{ "_source": false, "query": { "nested": { "path": "Wards", "query": { "match": { "Wards.CodeName": "ba_dinh" } }, "inner_hits": {} } } }

// GIS: find province covering a point
POST /provinces-gis/_search
{ "query": { "geo_shape": { "GIS.Geometry": { "shape": { "type": "point", "coordinates": [105.8542, 21.0285] }, "relation": "intersects" } } }, "_source": ["Code", "Name"] }
```

## Notes

- Field names use **PascalCase** (consistent with MongoDB/JSON exports).
- The `Meta` field is named without an underscore prefix — Elasticsearch reserves `_`-prefixed field names.
- NDJSON files use the Elasticsearch Bulk API format.

# Elasticsearch Dataset — Vietnamese Provinces Database

**Generated at: Sun, 20 Sep 2026 13:13:17 +0000**

Provinces and wards as Elasticsearch documents in two indices: `provinces` (no geometry) and `provinces-gis` (with GIS geometry).

## Files

- `provinces.ndjson` — Bulk API NDJSON for the provinces index (1.18 MB)
- `vn_provinces_metadata.ndjson` — Bulk API NDJSON for the vn_provinces_metadata index (153 B)
- `mappings/provinces.json` — Index mapping for provinces (2.72 KB)
- `mappings/vn_provinces_metadata.json` — Index mapping for vn_provinces_metadata (248 B)

## Overview

This dataset provides Vietnamese provinces and wards in Elasticsearch document format with two indices:

| Index | Documents | Description |
|-------|-----------|-------------|
| `provinces` | 34 | Provincial metadata with embedded wards, search keywords, and administrative unit data (no GIS geometry) |
| `provinces-gis` | 34 | Same structure plus GIS geometry for both provinces and wards (bounding boxes + GeoJSON polygons) |
| `vn_provinces_metadata` | 1 | Dataset version, latest decree, and generation timestamp |

## Data Structure

Each province is a single denormalized document with:

- **Core fields**: `Code`, `Name`, `NameEn`, `FullName`, `FullNameEn`, `CodeName`
- **`AdministrativeUnit`**: embedded administrative unit object (Id, FullName, ShortName, CodeName, ...)
- **`SearchKeywords`**: pre-computed autocomplete keywords (code, tone-stripped name, English name, codeName)
- **`Wards`**: nested array of ward documents with the same field shape (plus `PostalCode`)
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
curl -X PUT "localhost:9200/vn_provinces_metadata" -H 'Content-Type: application/json' -d @mappings/vn_provinces_metadata.json
```

2. Bulk import `provinces.ndjson`, `vn_provinces_metadata.ndjson`, and the `provinces-gis-part-*.ndjson` chunks in order (per `provinces-gis.ndjson.manifest`):

```bash
curl -X POST "localhost:9200/_bulk" -H 'Content-Type: application/x-ndjson' --data-binary @provinces.ndjson
curl -X POST "localhost:9200/_bulk" -H 'Content-Type: application/x-ndjson' --data-binary @vn_provinces_metadata.ndjson
curl -X POST "localhost:9200/_bulk" -H 'Content-Type: application/x-ndjson' --data-binary @provinces-gis-part-01.ndjson
```

3. Verify: 34 documents in each province index, 1 in `vn_provinces_metadata`.

## Sample Queries

```json
// Count documents
POST /provinces/_count

// Dataset version and latest decree
GET /vn_provinces_metadata/_doc/1

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
- NDJSON files use the Elasticsearch Bulk API format.

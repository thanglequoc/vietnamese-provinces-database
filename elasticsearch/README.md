# Vietnamese Provinces Database — Elasticsearch Dataset

## Overview

This dataset provides Vietnamese provinces and wards in Elasticsearch document format
with two indices:

| Index | Description |
|-------|-------------|
| `provinces` | Provincial metadata with embedded wards, search keywords, and administrative unit data (no GIS geometry) |
| `provinces-gis` | Same structure plus GIS geometry for both provinces and wards (bounding boxes + GeoJSON polygons) |

## Document Structure

Each province is a single denormalized document with:

- **Core fields**: `Code`, `Name`, `NameEn`, `FullName`, `FullNameEn`, `CodeName`
- **`AdministrativeUnit`**: Embedded administrative unit object (`Id`, `FullName`, `ShortName`, etc.)
- **`SearchKeywords`**: Pre-computed autocomplete keywords (code, tone-stripped Vietnamese name, English name, codeName)
- **`Wards`**: Array of nested ward documents with the same structure as provinces
- **`GIS`**: (provinces-gis only) `Center` (geo_point), `BoundingBox`, `Geometry` (geo_shape)
- **`Meta`**: Dataset version metadata (`DatasetVersion`, `AdministrativeRevision`, `GeneratedAt`)

## Quick Start

### 1. Create the Indices with Mappings

```bash
# Create the provinces index
curl -X PUT "localhost:9200/provinces" \
  -H 'Content-Type: application/json' \
  -d @mappings/provinces.json

# Create the provinces-gis index (with GIS geometry support)
curl -X PUT "localhost:9200/provinces-gis" \
  -H 'Content-Type: application/json' \
  -d @mappings/provinces-gis.json
```

### 2. Bulk Import the Data

```bash
# Import province data (non-GIS)
curl -X POST "localhost:9200/_bulk" \
  -H 'Content-Type: application/x-ndjson' \
  --data-binary @provinces.ndjson

# Import province data (with GIS)
curl -X POST "localhost:9200/_bulk" \
  -H 'Content-Type: application/x-ndjson' \
  --data-binary @provinces-gis.ndjson
```

### 3. Verify Import

```bash
curl "localhost:9200/provinces/_count"
curl "localhost:9200/provinces-gis/_count"
```

Expected: 34 documents in each index (one per province).

## Example Queries

### Province dropdown (sorted by code)

```json
POST /provinces/_search
{
  "size": 34,
  "sort": [{"Code": "asc"}],
  "_source": ["Code", "Name", "NameEn"]
}
```

### Search wards within a province

```json
POST /provinces/_search
{
  "query": {
    "nested": {
      "path": "Wards",
      "query": {
        "bool": {
          "must": [
            {"match": {"Wards.FullName": "Ba Đình"}}
          ]
        }
      },
      "inner_hits": {}
    }
  },
  "_source": ["Code", "Name"]
}
```

### Autocomplete search

```json
POST /provinces/_search
{
  "query": {
    "terms": {"SearchKeywords": ["ha noi"]}
  },
  "_source": ["Code", "Name", "NameEn"]
}
```

### GIS: Find province covering a point

```json
POST /provinces-gis/_search
{
  "query": {
    "geo_shape": {
      "GIS.Geometry": {
        "shape": {
          "type": "point",
          "coordinates": [105.8542, 21.0285]
        },
        "relation": "intersects"
      }
    }
  },
  "_source": ["Code", "Name"]
}
```

## File Listing

| File | Description |
|------|-------------|
| `provinces.ndjson` | Bulk API NDJSON for the provinces index |
| `provinces-gis.ndjson` | Bulk API NDJSON for the provinces-gis index |
| `mappings/provinces.json` | Index mapping for provinces |
| `mappings/provinces-gis.json` | Index mapping for provinces-gis |

## Notes

- Field names use **PascalCase** (consistent with MongoDB/JSON exports)
- The `Meta` field is named without underscore prefix — Elasticsearch reserves `_`-prefixed field names
- The dataset version and administrative revision are set at generation time
- NDJSON files use the Elasticsearch Bulk API format (each document = index action line + document line)
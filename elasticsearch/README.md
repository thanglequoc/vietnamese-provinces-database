# Vietnamese Provinces Database — Elasticsearch Dataset

Created at:  Sat, 25 Jul 2026 20:49:07 +0700

## Overview

This dataset provides Vietnamese provinces and wards in Elasticsearch document format
with two indices:

| Index | Description |
|-------|-------------|
| `provinces` | Provincial metadata with embedded wards, search keywords, and administrative unit data (no GIS geometry) |
| `provinces-gis` | Same structure plus GIS geometry for both provinces and wards (bounding boxes + GeoJSON polygons) |

## Document Structure

Each province is a single denormalized document with:

- **Core fields**: Code, Name, NameEn, FullName, FullNameEn, CodeName
- **`AdministrativeUnit`**: Embedded administrative unit object (Id, FullName, ShortName, etc.)
- **`SearchKeywords`**: Pre-computed autocomplete keywords (code, tone-stripped name, English name, codeName)
- **`Wards`**: Array of nested ward documents with the same structure
- **`GIS`**: (provinces-gis only) Center (geo_point), BoundingBox, Geometry (geo_shape)
- **`Meta`**: Dataset version metadata (DatasetVersion, AdministrativeRevision, GeneratedAt)

## Example Preview Document

Below is a sample province document (Hà Nội) with two of its wards:

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
        "ShortName": "Phường",
        "ShortNameEn": "Ward",
        "CodeName": "phuong",
        "CodeNameEn": "ward"
      },
      "SearchKeywords": ["00004", "ba dinh", "ba_dinh"]
    },
    {
      "Code": "00070",
      "Name": "Hoàn Kiếm",
      "NameEn": "Hoan Kiem",
      "FullName": "Phường Hoàn Kiếm",
      "FullNameEn": "Hoan Kiem Ward",
      "CodeName": "hoan_kiem",
      "AdministrativeUnit": {
        "Id": 3,
        "FullName": "Phường",
        "FullNameEn": "Ward",
        "ShortName": "Phường",
        "ShortNameEn": "Ward",
        "CodeName": "phuong",
        "CodeNameEn": "ward"
      },
      "SearchKeywords": ["00070", "hoan kiem", "hoan_kiem"]
    }
  ],
  "Meta": {
    "DatasetVersion": "2026.07.01",
    "AdministrativeRevision": "2026-04-30",
    "GeneratedAt": "2026-07-25T03:00:43Z"
  }
}
```

The `provinces-gis` index extends this same structure with a `GIS` object at both the province and ward level:

```json
{
  "Code": "01",
  "Name": "Hà Nội",
  "FullName": "Thành phố Hà Nội",
  "CodeName": "ha_noi",
  "AdministrativeUnit": {
    "Id": 1,
    "FullName": "Thành phố trực thuộc trung ương",
    "ShortName": "Thành phố"
  },
  "SearchKeywords": ["01", "ha noi", "hanoi", "ha_noi"],
  "GIS": {
    "Center": { "Lat": 21.0285, "Lon": 105.8542 },
    "BoundingBox": {
      "MinLongitude": 105.2859,
      "MinLatitude": 20.4863,
      "MaxLongitude": 106.0617,
      "MaxLatitude": 21.3851
    },
    "Geometry": {
      "type": "MultiPolygon",
      "coordinates": [[[[105.2859, 21.3851], [106.0617, 21.3851], ...]]]
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
      "FullName": "Phường Ba Đình",
      "CodeName": "ba_dinh",
      "AdministrativeUnit": { "Id": 3, "ShortName": "Phường" },
      "SearchKeywords": ["00004", "ba dinh", "ba_dinh"],
      "GIS": {
        "Center": { "Lat": 21.0347, "Lon": 105.8231 },
        "BoundingBox": {
          "MinLongitude": 105.8115, "MinLatitude": 21.0261,
          "MaxLongitude": 105.8347, "MaxLatitude": 21.0433
        },
        "Geometry": { "type": "Polygon", "coordinates": [[[105.8115, 21.0433], ...]] },
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
      }
    }
  ],
  "Meta": {
    "DatasetVersion": "2026.07.01",
    "AdministrativeRevision": "2026-04-30",
    "GeneratedAt": "2026-07-25T03:00:43Z"
  }
}
```

> **Note**: The `Geometry` field contains full GeoJSON polygons/multipolygons. The example above uses `...` to abbreviate the coordinate arrays for readability. Actual geometries for provinces are MultiPolygon with thousands of coordinate pairs.

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

### Return only the matched ward (no parent province)

Use `_source: false` on the parent document so only the nested ward hit is returned:

```json
POST /provinces/_search
{
  "_source": false,
  "query": {
    "nested": {
      "path": "Wards",
      "query": { "match": { "Wards.CodeName": "truong_sa" } },
      "inner_hits": { "name": "ward", "_source": true }
    }
  }
}
```

The ward document (with GIS data if present) is available at `.hits.hits[0].inner_hits.ward.hits.hits[0]._source`.

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

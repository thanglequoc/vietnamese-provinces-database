# Elasticsearch Dataset — Vietnamese Provinces Database

**Generated at: Fri, 14 Aug 2026 08:42:28 +0700**

Provinces and wards as Elasticsearch documents in two indices: `provinces` (no geometry) and `provinces-gis` (with GIS geometry).

## Files

- `provinces.ndjson` — Bulk API NDJSON for the provinces index (1.18 MB)
- `mappings/provinces.json` — Index mapping for provinces (3.01 KB)

## Overview

This dataset provides Vietnamese provinces and wards in Elasticsearch document format with two indices:

| Index | Description |
|-------|-------------|
| `provinces` | Provincial metadata with embedded wards, search keywords, and administrative unit data (no GIS geometry) |
| `provinces-gis` | Same structure plus GIS geometry for both provinces and wards (bounding boxes + GeoJSON polygons) |

## Data Structure

Each province is a single denormalized document with:

- **Core fields**: Code, Name, NameEn, FullName, FullNameEn, CodeName
- **`AdministrativeUnit`**: Embedded administrative unit object (Id, FullName, ShortName, etc.)
- **`SearchKeywords`**: Pre-computed autocomplete keywords (code, tone-stripped name, English name, codeName)
- **`Wards`**: Array of nested ward documents with the same structure
- **`GIS`**: (provinces-gis only) Center (geo_point), BoundingBox, Geometry (geo_shape)
- **`Meta`**: Dataset version metadata (DatasetVersion, AdministrativeRevision, GeneratedAt)

## Sample Queries

```bash
# Count documents
curl "localhost:9200/provinces/_count"
curl "localhost:9200/provinces-gis/_count"

# Autocomplete search
curl "localhost:9200/provinces/_search" -H 'Content-Type: application/json' -d '{"query": {"terms": {"SearchKeywords": ["ha noi"]}}, "_source": ["Code", "Name", "NameEn"]}'

# GIS: find province covering a point
curl "localhost:9200/provinces-gis/_search" -H 'Content-Type: application/json' -d '{"query": {"geo_shape": {"GIS.Geometry": {"shape": {"type": "point", "coordinates": [105.8542, 21.0285]}, "relation": "intersects"}}}, "_source": ["Code", "Name"]}'
```

## Quick Start

1. Create the indices with the mappings in `mappings/`.
2. Bulk import `provinces.ndjson` (and the `provinces-gis-part-*.ndjson` chunks in order, per `provinces-gis.ndjson.manifest`).
3. Verify: 34 documents in each index.

## Notes

- Field names use **PascalCase** (consistent with MongoDB/JSON exports).
- The `Meta` field is named without an underscore prefix — Elasticsearch reserves `_`-prefixed field names.
- NDJSON files use the Elasticsearch Bulk API format.

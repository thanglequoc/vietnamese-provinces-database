# Vietnamese Provinces Database — MongoDB GIS Dataset

Created at:  Sat, 08 Aug 2026 21:34:00 +0700

## Overview

This dataset provides Vietnamese provinces and wards in MongoDB document format
with two GIS collections:

| Collection | Documents | Description |
|------------|-----------|-------------|
| `provinces-gis` | 34 | Province documents with GIS geometry (bounding boxes + GeoJSON polygons) |
| `wards-gis` | 3,321 | Standalone ward documents with GIS geometry + ProvinceCode reference |

## Document Structure

### Province GIS Document

- **Core fields**: Code, Name, NameEn, FullName, FullNameEn, CodeName
- **`AdministrativeUnit`**: Embedded administrative unit object
- **`SearchKeywords`**: Pre-computed autocomplete keywords
- **`GIS`**: Center (GeoJSON Point), BoundingBox, Geometry (GeoJSON MultiPolygon), Properties
- **`Meta`**: Dataset version metadata

### Ward GIS Document

- Same structure as province, plus **`ProvinceCode`** for cross-collection joins

## Quick Start

### 1. Import the Data

```bash
# Import province GIS data
mongoimport --db vn_provinces --collection provinces_gis --file mongo_data_vn_province_gis_*.json --jsonArray

# Import ward GIS data (may be chunked — import each part sequentially)
mongoimport --db vn_provinces --collection wards_gis --file mongo_data_vn_ward_gis_*.json --jsonArray
```

> **Note**: If the ward GIS file was chunked, you'll see multiple files like
> `mongo_data_vn_ward_gis_*_part_01.json`, `part_02.json`, etc.
> Import each part individually, or use a script that reads the manifest file
> (`*.manifest`) for automated sequential import.

### 2. Create Indexes

```bash
mongosh vn_provinces create_indexes.js
```

### 3. Example Queries

```javascript
// Find province containing a point
db.getCollection('provinces-gis').findOne({
  "GIS.Geometry": {
    $geoIntersects: {
      $geometry: { type: "Point", coordinates: [105.8542, 21.0285] }
    }
  }
})

// Find all wards in a province
db.getCollection('wards-gis').find({ ProvinceCode: "01" })

// Find ward containing a point
db.getCollection('wards-gis').findOne({
  "GIS.Geometry": {
    $geoIntersects: {
      $geometry: { type: "Point", coordinates: [105.8231, 21.0347] }
    }
  }
})

// Find provinces near a point (within 50km)
db.getCollection('provinces-gis').find({
  "GIS.Center": {
    $near: {
      $geometry: { type: "Point", coordinates: [105.8542, 21.0285] },
      $maxDistance: 50000
    }
  }
})

// Join wards with provinces using $lookup
db.getCollection('wards-gis').aggregate([
  { $match: { ProvinceCode: "01" } },
  { $lookup: {
      from: "provinces-gis",
      localField: "ProvinceCode",
      foreignField: "Code",
      as: "Province"
  }}
])
```

## File Listing

| File | Description |
|------|-------------|
| `mongo_data_vn_province_gis_*.json` | Province GIS documents (JSON array) |
| `mongo_data_vn_ward_gis_*.json` | Ward GIS documents (JSON array, may be chunked) |
| `create_indexes.js` | Index creation script for both collections |

# MongoDB Dataset — Vietnamese Provinces Database

**Generated at: Fri, 14 Aug 2026 08:42:28 +0700**

MongoDB documents for Vietnamese provinces with embedded wards.

## Files

- `administrative_units.json` — Array of 8 administrative unit types (1016 B)
- `administrative_regions.json` — Array of 8 regions (1.15 KB)
- `mongo_data_vn_unit.json` — Array of 34 province documents, each embedding its Wards array (953.51 KB)

## Data Structure

- `administrative_units.json` — array of 8 administrative unit types
- `administrative_regions.json` — array of 8 regions
- `mongo_data_vn_unit.json` — the `provinces` collection: 34 province documents, each with an embedded `Wards` array

## Sample Queries

```javascript
// Count provinces
db.getCollection('provinces').countDocuments();

// Wards of a province
db.getCollection('provinces').findOne({Code: '01'}, {Name: 1, Wards: 1});
```

## GIS / GeoJSON

The `gis/` subfolder contains the `provinces-gis` (34) and `wards-gis` (3,321) collections (`mongo_data_vn_province_gis.json`, `mongo_data_vn_ward_gis[_part_NN].json`), the `create_indexes.js` index script, and a `.manifest`. Import them, run `create_indexes.js`, then query with `$geoIntersects`.

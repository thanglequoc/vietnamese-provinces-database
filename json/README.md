# JSON Dataset — Vietnamese Provinces Database

**Generated at: Fri, 14 Aug 2026 08:42:28 +0700**

Administrative unit JSON data for Vietnam: provinces with embedded wards.

## Files

- `full_json_generated_data_vn_units.json` — Full dataset (provinces + wards + administrative info) (1.51 MB)
- `simplified_json_generated_data_vn_units.json` — Simplified dataset (pretty-printed) (786.42 KB)
- `simplified_json_generated_data_vn_units_minified.json` — Simplified dataset (minified) (603.50 KB)
- `vn_only_simplified_json_generated_data_vn_units.json` — Vietnamese-only simplified (pretty-printed) (393.83 KB)
- `vn_only_simplified_json_generated_data_vn_units_minified.json` — Vietnamese-only simplified (minified) (289.28 KB)

## Data Structure

Each entry is a province object with `code`, `name`, `nameEn`, `fullName`, `fullNameEn`, `codeName`, `administrativeUnit*`, `postalCodePrefix`, and a `wards` array of ward objects.

## Sample Queries

```js
const dataset = require('./full_json_generated_data_vn_units.json');
dataset.find(p => p.code === '01');
dataset.flatMap(p => p.wards).filter(w => w.postalCode === '11024');
```

## GIS / GeoJSON

The `geojson/` subfolder contains per-province and per-ward GeoJSON boundary exports, and `vn_provinces_wards_geojson.zip` is the combined archive of those files. These artifacts are present when the GIS generation step runs.

# JSON Dataset — Vietnamese Provinces Database

**Generated at: Sun, 20 Sep 2026 13:13:17 +0000**

Administrative unit JSON data for Vietnam: provinces with embedded wards, in full and simplified forms.

## Files

- `full_json_generated_data_vn_units.json` — Full dataset (provinces + wards + administrative info) (1.51 MB)
- `simplified_json_generated_data_vn_units.json` — Simplified dataset (pretty-printed) (786.46 KB)
- `simplified_json_generated_data_vn_units_minified.json` — Simplified dataset (minified) (603.54 KB)
- `vn_only_simplified_json_generated_data_vn_units.json` — Vietnamese-only simplified (pretty-printed) (393.91 KB)
- `vn_only_simplified_json_generated_data_vn_units_minified.json` — Vietnamese-only simplified (minified) (289.36 KB)
- `vn_provinces_metadata.json` — Dataset version, latest decree, and generation timestamp (107 B)

## Overview

| File | Contents |
|------|----------|
| `full_json_generated_data_vn_units.json` | 34 province objects with full administrative info and embedded wards |
| `simplified_json_generated_data_vn_units.json` | Same structure, fewer fields (pretty-printed) |
| `simplified_json_generated_data_vn_units_minified.json` | Simplified, minified (no whitespace) |
| `vn_only_simplified_json_generated_data_vn_units.json` | Vietnamese-only fields (pretty-printed) |
| `vn_only_simplified_json_generated_data_vn_units_minified.json` | Vietnamese-only fields (minified) |
| `vn_provinces_metadata.json` | Dataset version, latest decree, and generation timestamp |

## Data Structure

Each entry is a province object:

- **`code`** — province code (`01`)
- **`name` / `nameEn`** — province name
- **`fullName` / `fullNameEn`** — full name with unit prefix
- **`codeName`** — code name (`ha_noi`)
- **`administrativeUnitId` / `administrativeUnitShortName` / `administrativeUnitFullName`** — unit type
- **`postalCodePrefix`** — comma-separated 2-digit postal prefixes
- **`wards`** — array of ward objects (`code`, `name`, `nameEn`, `fullName`, `fullNameEn`, `codeName`, `provinceCode`, `postalCode`, unit fields)

`vn_provinces_metadata.json` is a separate single object describing the dataset release:

- **`DatasetVersion`** — dataset release version (e.g. `v5.1.0`)
- **`LatestDecree`** — latest government decree reflected in the data (e.g. `30/2026/QH16`)
- **`GeneratedAt`** — dataset generation timestamp (UTC, RFC 3339)

## Sample Document

```json
[
  {
    "code": "01",
    "name": "Hà Nội",
    "nameEn": "Ha Noi",
    "fullName": "Thành phố Hà Nội",
    "fullNameEn": "Ha Noi City",
    "codeName": "ha_noi",
    "postalCodePrefix": "10, 11, 12, 13, 14",
    "wards": [
      { "code": "00004", "name": "Ba Đình", "nameEn": "Ba Dinh", "postalCode": "11120" }
    ]
  }
]
```

## Quick Start

```js
const dataset = require('./full_json_generated_data_vn_units.json');
```

```python
import json
with open('full_json_generated_data_vn_units.json', encoding='utf-8') as f:
    dataset = json.load(f)
```

## Sample Queries

```js
// Find Hà Nội
dataset.find(p => p.code === '01');

// Wards with postal code 11024
dataset.flatMap(p => p.wards).filter(w => w.postalCode === '11024');

// All province names
dataset.map(p => p.name);
```

## GIS / GeoJSON

The `geojson/` subfolder contains per-province and per-ward GeoJSON boundary exports, and `vn_provinces_wards_geojson.zip` is the combined archive of those files. These artifacts are present when the GIS generation step runs.

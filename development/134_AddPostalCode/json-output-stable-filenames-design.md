# Design: Stable JSON Output Filenames, Minified Variants, and Combined README

**Date:** 2026-08-14
**Branch:** `134_AddPostalCode`
**Status:** Approved

## Objective

Make the JSON dataset generation output easier to publish to the top-level `json/`
data source folder:

1. Remove the datetime suffix from all generated JSON files so output filenames
   exactly match the published `json/` folder, making copy-over trivial.
2. Generate `_minified` variants (compact JSON, no whitespace/newlines) for the
   simplified and VN-only simplified datasets.
3. Move the README out of the geojson folder to the `json/` root, updated to
   cover both administrative JSON data and GIS/geojson data, with a bold
   generation timestamp at the top.

## Requirements (confirmed with user)

1. Remove the datetime suffix from the 3 admin JSON files **and** the geojson zip
   file (`vn_provinces_wards_geojson.zip`).
2. Minified versions are generated only for `simplified_json_generated_data_vn_units`
   and `vn_only_simplified_json_generated_data_vn_units` — **not** for
   `full_json_generated_data_vn_units`.
3. Minified files are written by the Go script into `output/json/` alongside the
   pretty files.
4. The combined `README.md` is generated in the **JSON writer (non-GIS phase)**,
   written to the `json/` output root, timestamped with the JSON generation time.
   It exists even on admin-only runs (INCLUDE_GIS=false).
5. The `geojson/` subfolder keeps its own geojson-focused README unchanged
   (still included inside the zip as `geojson/README.md`).

## Changes

### 1. `internal/dataset_writer/dataset_file_writer/json_file_writer.go`

- Drop `getFileTimeSuffix()` from all 3 filenames:
  - `full_json_generated_data_vn_units.json`
  - `simplified_json_generated_data_vn_units.json`
  - `vn_only_simplified_json_generated_data_vn_units.json`
- Add compact output (`json.Marshal`, no indent) for:
  - `simplified_json_generated_data_vn_units_minified.json`
  - `vn_only_simplified_json_generated_data_vn_units_minified.json`
- Add a small helper `writeMinifiedJSON(filePath, payload)` to avoid duplicating
  the compact marshal + write logic.
- Write combined `README.md` to the output folder root
  (`output/json/README.md`) with:
  - A bold `**Generated at: <RFC1123Z timestamp>**` header on top.
  - A description of the 5 JSON data files (full, simplified, VN-only simplified,
    and their minified variants).
  - A description of the `geojson/` subfolder and `vn_provinces_wards_geojson.zip`.
  - File size stats for each artifact.
  - Note that geojson/GIS artifacts are present when the GIS generation step ran.

### 2. `internal/dataset_writer/dataset_file_writer/geojson_file_writer.go`

- `archiveGeoJSONDirectory`: change archive name from
  `vn_provinces_wards_geojson_<suffix>.zip` → `vn_provinces_wards_geojson.zip`.
- `writeGeoJSONReadme` unchanged — still writes into the `geojson/` subfolder and
  is still included in the zip.

### 3. Tests `internal/dataset_writer/dataset_file_writer/json_file_writer_test.go`

- File counts in `WriteToFile` tests change from `3` → `6`
  (5 JSON files + `README.md`).
- Filenames are now deterministic, so assertions reference exact names instead of
  index/suffix slicing.
- New assertions:
  - Minified files parse as valid JSON.
  - Minified files contain no `\n` (single line).
  - Minified content equals the compact form (no whitespace between tokens).
- README assertions: bold `Generated at:` present, mentions geojson + zip.
- `TestJSONDatasetFileWriter_WriteGISGeoJSONToFile`: update expected zip name to
  `vn_provinces_wards_geojson.zip`; the `geojson/README.md` entry assertion stays.

### 4. Docs

- Update `dataset-generation-scripts/README.md` output-structure tree to the new
  fixed filenames + minified entries.

## Edge Cases

- **Deterministic filenames**: because the suffix is removed, each run overwrites
  the previous output. This is intended — the `output/` folder is a staging area
  and the top-level `json/` folder is the published artifact.
- **Admin-only runs**: the combined README is generated in the JSON writer, so it
  exists even when `INCLUDE_GIS=false`. It describes geojson artifacts as present
  "when the GIS step runs".
- **Zip naming**: removing the suffix means only the latest zip survives in
  `output/json/` — matches the published folder which keeps one zip.

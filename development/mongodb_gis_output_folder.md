# MongoDB GIS Output Folder Refactor — Implementation Plan

**Goal:** Make the MongoDB GIS writer place all GIS-related files under `dataset-generation-scripts/output/mongodb/gis/` (mirroring `postgresql/gis`, `mysql/gis`, `sqlserver/gis`, `json/geojson`), leaving only non-GIS files in `output/mongodb/`.

**Architecture:** The MongoDB GIS writer (`WriteMongoGISDataToFile`) already routes every file it produces through `w.OutputFolderPath` (province/ward JSON, chunk parts + manifest, `create_indexes.js`, `README.md`). So the refactor is a single path change at the call site in `dataset_writer.go` — the same pattern GeoJSON uses (`./output/json/geojson`). No writer logic changes. Two downstream consumers (the import/verify shell script and AGENTS.md) get their paths updated to the new `gis/` location.

**Tech Stack:** Go 1.24, Bash, Markdown.

## Global Constraints
- Generator output only — do **not** touch the published top-level `mongodb/` folder.
- All 4 artifact types produced by `WriteMongoGISDataToFile` move under `mongodb/gis/`: province GIS JSON, ward GIS JSON (+ `_part_NN.json` + `.manifest`), `create_indexes.js`, `README.md`.
- Non-GIS Mongo files (`administrative_regions_*.json`, `administrative_units_*.json`, `mongo_data_vn_unit_*.json`) stay in `output/mongodb/`.

---

### Task 1: Point the MongoDB GIS writer at `output/mongodb/gis`

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_writer.go:173-182`

- [ ] **Step 1: Edit the call site**

Change `dataset_writer.go` line 175:

```go
	// MongoDB GIS
	mongoDBGISFileWriter := datasetfilewriter.MongoDBDatasetFileWriter{
		OutputFolderPath: "./output/mongodb/gis",
	}
```

The GIS writer's `os.MkdirAll(w.OutputFolderPath, 0746)` already creates the `gis/` dir (and its parent `./output/mongodb`, created earlier by the non-GIS writer at line 86), so no writer code changes are needed.

- [ ] **Step 2: Build & vet**

Run: `cd dataset-generation-scripts && go build ./... && go vet ./internal/dataset_writer/...`
Expected: no errors.

- [ ] **Step 3: Run the Mongo writer unit tests**

Run: `cd dataset-generation-scripts && go test -v ./internal/dataset_writer/...`
Expected: PASS. (`mongodb_gis_file_writer_test.go` writes into `t.TempDir()` via `OutputFolderPath`, so it validates the writer still emits all 4 artifacts into whatever path it's given.)

- [ ] **Step 4: Commit**

```bash
git add dataset-generation-scripts/internal/dataset_writer/dataset_writer.go
git commit -m "refactor: output MongoDB GIS files under output/mongodb/gis"
```

---

### Task 2: Update the MongoDB import/verify script

**Files:**
- Modify: `dataset-generation-scripts/integration-test/import_and_verify_mongodb.sh`

The script globs `$DATA_DIR/output/mongodb` for both non-GIS and GIS files. Add a `GIS_DATA_DIR` and point GIS-only lookups there.

- [ ] **Step 1: Add GIS dir + update GIS paths**

After line 12 add:

```bash
GIS_DATA_DIR="$DATA_DIR/gis"
```

Then update these GIS-specific references:

| Line | From | To |
|------|------|-----|
| 74 | `MANIFEST_FILE=$(ls "$DATA_DIR"/*ward_gis*.json.manifest ...` | `MANIFEST_FILE=$(ls "$GIS_DATA_DIR"/*ward_gis*.json.manifest ...` |
| 84 | `"$DATA_DIR"/mongo_data_vn_province_gis_*.json` | `"$GIS_DATA_DIR"/mongo_data_vn_province_gis_*.json` |
| 86 | `"$DATA_DIR/create_indexes.js"` | `"$GIS_DATA_DIR/create_indexes.js"` |
| 92 | `PROV_GIS_FILE=$(ls "$DATA_DIR"/mongo_data_vn_province_gis_*.json ...` | `PROV_GIS_FILE=$(ls "$GIS_DATA_DIR"/mongo_data_vn_province_gis_*.json ...` |
| 94 | `"$DATA_DIR/create_indexes.js"` | `"$GIS_DATA_DIR/create_indexes.js"` |
| 104 | `WARD_PARTS+=("$DATA_DIR/$line")` | `WARD_PARTS+=("$GIS_DATA_DIR/$line")` |
| 200 | `mongosh "$CONN_STRING" --quiet --file "$DATA_DIR/create_indexes.js"` | `mongosh "$CONN_STRING" --quiet --file "$GIS_DATA_DIR/create_indexes.js"` |

Non-GIS lookups (lines 80-83, 89-91) keep `$DATA_DIR`.

- [ ] **Step 2: Syntax-check the script**

Run: `bash -n dataset-generation-scripts/integration-test/import_and_verify_mongodb.sh`
Expected: no output (exit 0).

- [ ] **Step 3: Commit**

```bash
git add dataset-generation-scripts/integration-test/import_and_verify_mongodb.sh
git commit -m "refactor: import MongoDB GIS files from output/mongodb/gis"
```

---

### Task 3: Update AGENTS.md Mongo import documentation

**Files:**
- Modify: `AGENTS.md:431-459`

- [ ] **Step 1: Update the Mongo output-path docs**

- Line 431: change `output to \`dataset-generation-scripts/output/mongodb/\`` → `output to \`dataset-generation-scripts/output/mongodb/\` (GIS files under \`output/mongodb/gis/\`)`.
- Lines 440, 447, 453, 459: prefix the referenced files with `gis/`:
  - `dataset-generation-scripts/output/mongodb/gis/mongo_data_vn_province_gis.json`
  - `dataset-generation-scripts/output/mongodb/gis/mongo_data_vn_ward_gis.json`
  - `dataset-generation-scripts/output/mongodb/gis/mongo_data_vn_ward_gis_part_01.json`
  - `dataset-generation-scripts/output/mongodb/gis/create_indexes.js`

- [ ] **Step 2: Verify no stale references**

Run: `rg -n "output/mongodb" --glob '!output/**' --glob '!mongodb/**' --glob '!dataset-generation-scripts/output/**' .`
Expected: only the updated `gis/` paths plus the non-GIS ones in `dataset_writer.go:86` and the script's `DATA_DIR`.

- [ ] **Step 3: Commit**

```bash
git add AGENTS.md
git commit -m "docs: reference MongoDB GIS files under output/mongodb/gis"
```

---

### Task 4: End-to-end verification

- [ ] **Step 1: Regenerate the dataset**

With the Postgres/PostGIS container up (`docker compose -f dataset-generation-scripts/docker/docker-compose.yaml up -d`) and `INCLUDE_GIS` enabled, run:
`cd dataset-generation-scripts && go run main.go`

- [ ] **Step 2: Assert the new output tree**

```bash
ls dataset-generation-scripts/output/mongodb/          # only administrative_regions_*.json, administrative_units_*.json, mongo_data_vn_unit_*.json
ls dataset-generation-scripts/output/mongodb/gis/      # mongo_data_vn_province_gis_*.json, mongo_data_vn_ward_gis_* (+ parts + .manifest), create_indexes.js, README.md
```

Expected: `gis/` holds all 4 GIS artifact types; root holds only the 3 non-GIS files.

- [ ] **Step 3: Commit any leftover doc/source fixes (if verification surfaced none, skip)**

---

## Self-review notes
- All references to `output/mongodb` were located (grep above): `dataset_writer.go:86` (keep), `dataset_writer.go:175` (change), `integration-test/import_and_verify_mongodb.sh:12` (keep base, add `gis`), `AGENTS.md:431/440/447/453/459` (update). No other code/doc references exist; `CLAUDE.md`, `docs/gis/`, and CI workflows are unaffected.
- Writer unit tests need no changes — behavior is driven entirely by `OutputFolderPath`.

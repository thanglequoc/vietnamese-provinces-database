# Remove Meta Dataset Version from ES & MongoDB Documents — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove the `Meta` dataset-version object from Elasticsearch and MongoDB GIS documents, keeping the version constants for a future dedicated index.

**Architecture:** Remove `Meta` from the document DTOs, stop attaching it in the writers/mappers, drop it from the ES mappings, and update READMEs/tests/docs. Then regenerate and publish.

**Tech Stack:** Go 1.24, Testify. Tests from `dataset-generation-scripts/`.

## Global Constraints

- Module root: `dataset-generation-scripts/`; tests: `go test ./internal/dataset_writer/dataset_file_writer/... -v`.
- Keep package-level constants `esDatasetVer`, `esAdminRev`, `mongoDatasetVer`, `mongoAdminRev` (unused is legal; kept for the future index).
- Output documents must contain no `Meta` object.

---

### Task 1: Remove Meta from Elasticsearch

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/dto/elasticsearch_dto.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer_test.go`

- [ ] **Step 1: Update the failing test**

In `elasticsearch_file_writer_test.go`, `TestWriteToFile_NonGIS` (lines 92-97) asserts `doc.Meta` where `doc` is a typed `dataset_file_writer_dto.ElasticsearchProvinceDocument`. Once the `Meta` field is removed from the struct, these lines no longer compile. Delete this block:

```go
	if doc.Meta == nil {
		t.Fatal("expected Meta to be set")
	}
	if doc.Meta.DatasetVersion != esDatasetVer {
		t.Errorf("expected DatasetVersion %q, got %q", esDatasetVer, doc.Meta.DatasetVersion)
	}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestWriteToFile_NonGIS -v`
Expected: FAIL (build failure — `doc.Meta` undefined).

- [ ] **Step 3: Remove Meta from the DTO**

In `dto/elasticsearch_dto.go`:
- Delete the `Meta *ElasticsearchMeta \`json:"Meta,omitempty"\`` field from `ElasticsearchProvinceDocument` (line ~19) and from `ElasticsearchWardDocument` (if present).
- Delete the `ElasticsearchMeta` struct (lines ~84-90).

- [ ] **Step 4: Remove Meta from the writer**

In `elasticsearch_file_writer.go`:
- `WriteToFile`: delete the `generatedAt := time.Now().UTC().Format(time.RFC3339)` local (line 52) and the Meta-attach loop (lines 54-60).
- `WriteElasticsearchGISDataToFile`: delete `generatedAt` (line 100) and the per-doc `Meta: &dataset_file_writer_dto.ElasticsearchMeta{...}` block (lines ~115-120).
- Delete the `Meta` mapping blocks from the two mapping JSON strings (the `"Meta": map[string]interface{}{...}` entries, lines ~535 and ~674). Keep the rest of each mapping intact.

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run "TestWriteToFile_NonGIS|TestWriteElasticsearchGISDataToFile_GIS" -v`
Expected: PASS.

- [ ] **Step 6: Run the full package test suite**

Run: `go test ./internal/dataset_writer/dataset_file_writer/...`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/dto/elasticsearch_dto.go internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer.go internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer_test.go
git commit -m "feat: remove Meta dataset version from elasticsearch documents"
```

---

### Task 2: Remove Meta from MongoDB GIS

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/dto/mongo_gis_dto.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/helper/mongo_gis_mapper.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/helper/mongo_gis_mapper_test.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/mongodb_gis_file_writer.go`

- [ ] **Step 1: Update the failing tests**

In `helper/mongo_gis_mapper_test.go`, remove the `doc.Meta` assertions (lines ~118-122):

```go
	if doc.Meta == nil {
		t.Fatal("expected Meta to be populated")
	}
	if doc.Meta.DatasetVersion != "2026.07.01" {
		t.Errorf("expected DatasetVersion '2026.07.01', got %q", doc.Meta.DatasetVersion)
	}
```

Update the `ConvertToMongoGISProvinceDocuments` / `ConvertToMongoGISWardDocuments` call sites in the test to drop the `(version, adminRev, generatedAt)` arguments — the new signature takes only the geo units:

```go
	docs := ConvertToMongoGISProvinceDocuments(geoProvinces)
```

and

```go
	docs := ConvertToMongoGISWardDocuments(geoWards)
```

Add a check that the marshalled document has no `Meta`:

```go
	raw, err := json.Marshal(docs[0])
	if err != nil {
		t.Fatalf("marshal doc: %v", err)
	}
	if bytes.Contains(raw, []byte("Meta")) {
		t.Fatal("document should not contain a Meta object")
	}
```

(Add `"bytes"` to the test imports if not present.)

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/dataset_writer/dataset_file_writer/helper/... -run TestConvertToMongoGIS -v`
Expected: FAIL — mapper still sets Meta and/or signature mismatch.

- [ ] **Step 3: Remove Meta from the DTO**

In `dto/mongo_gis_dto.go`:
- Delete the `Meta *MongoMeta \`json:"Meta,omitempty"\`` field from `MongoGISProvinceDocument` (line ~17) and `MongoGISWardDocument` (line ~33).
- Delete the `MongoMeta` struct (lines ~83-90).

- [ ] **Step 4: Update the mapper**

In `helper/mongo_gis_mapper.go`, change the signatures of
`ConvertToMongoGISProvinceDocuments` and `ConvertToMongoGISWardDocuments` to
drop the `(datasetVersion, adminRevision, generatedAt string)` parameters, and
delete the `Meta: &dataset_file_writer_dto.MongoMeta{...}` assignments (lines
~30-33 and ~77-80).

- [ ] **Step 5: Update the writer**

In `mongodb_gis_file_writer.go` `WriteMongoGISDataToFile`:
- Delete `generatedAt := time.Now().UTC().Format(time.RFC3339)` (line 29).
- Update the two mapper calls to the new signatures:

```go
	provinceDocs := file_writer_helper.ConvertToMongoGISProvinceDocuments(sapNhapGeoProvinces)
```

```go
	wardDocs := file_writer_helper.ConvertToMongoGISWardDocuments(sapNhapGeoWards)
```

- Keep the `mongoDatasetVer` / `mongoAdminRev` constants.

- [ ] **Step 6: Run the tests to verify they pass**

Run: `go test ./internal/dataset_writer/dataset_file_writer/...`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/dto/mongo_gis_dto.go internal/dataset_writer/dataset_file_writer/helper/mongo_gis_mapper.go internal/dataset_writer/dataset_file_writer/helper/mongo_gis_mapper_test.go internal/dataset_writer/dataset_file_writer/mongodb_gis_file_writer.go
git commit -m "feat: remove Meta dataset version from mongodb GIS documents"
```

---

### Task 3: Update READMEs, tests, and AGENTS.md

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer.go` (README sections in `writeESReadme`)
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/mongodb_file_writer.go` (README section in `writeMongoReadme`)
- Modify: `AGENTS.md`

- [ ] **Step 1: Remove the Meta bullet from the ES README content**

In `writeESReadme` sections:
- Delete `"- **\`Meta\`**: \`DatasetVersion\`, \`AdministrativeRevision\`, \`GeneratedAt\`",` (the Data Structure bullet).
- Delete `"- The \`Meta\` field is named without an underscore prefix — Elasticsearch reserves \`_\`-prefixed field names.",` (the Notes line).

- [ ] **Step 2: Remove the Meta bullet from the Mongo README content**

In `writeMongoReadme` sections, delete `"- **\`Meta\`** — dataset version metadata",`.

- [ ] **Step 3: Update AGENTS.md**

- In the Elasticsearch schema section, remove the `Meta` field row from the key-fields table (the `| Meta | object | ... |` row).
- In the MongoDB section, remove any `Meta` references in the document-structure lists if present.

- [ ] **Step 4: Run the full package test suite**

Run: `go test ./internal/dataset_writer/dataset_file_writer/...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer.go internal/dataset_writer/dataset_file_writer/mongodb_file_writer.go AGENTS.md
git commit -m "docs: drop Meta references from generated READMEs and AGENTS.md"
```

---

### Task 4: Regenerate and publish

**Files:** none (generated data)

- [ ] **Step 1: Run the full test suite + vet**

Run from `dataset-generation-scripts/`: `go test ./... 2>&1 | grep -E 'FAIL|ok '` and `go vet ./internal/dataset_writer/...`
Expected: all `ok`, vet clean.

- [ ] **Step 2: Regenerate and publish**

Ensure Docker Postgres is up, then:

```bash
go run main.go
bash copy-datasets-to-repo.sh
```

- [ ] **Step 3: Verify no Meta remains**

Run: `rg -l '"Meta"' elasticsearch/*.ndjson elasticsearch/provinces-gis-part-*.ndjson mongodb/gis/*.json` and `rg -c '"GeneratedAt"' elasticsearch/*.ndjson mongodb/gis/*.json`
Expected: no files contain `Meta` / `GeneratedAt`.

- [ ] **Step 4: Commit the regenerated data**

```bash
git add elasticsearch/ mongodb/
git commit -m "data: regenerate ES and MongoDB datasets without Meta object"
```

(Leave `docs/release_notes_v4.1.0.md` and `docs/release_notes_v4.2.0.md` untracked.)

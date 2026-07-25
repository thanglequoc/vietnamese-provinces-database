# Fix GIS Geometry Self-Intersections — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Apply `ST_MakeValid()` + `ST_CollectionExtract(3)` to 78 ward geometries in PostGIS that have self-intersecting polygon rings, fixing Elasticsearch `geo_shape` import failures for 22 provinces.

**Architecture:** New `ValidateAndFixGeometries()` method on `SapNhapService` runs a single UPDATE SQL with RETURNING clause via Bun ORM, writes an audit log of fixed wards, runs a verification query, and integrates into `main.go` after island patching and before dataset export. The fix updates `geom_wkt` text (computed `geom` column follows automatically).

**Tech Stack:** Go 1.24.0, Bun ORM (PostgreSQL dialect), PostGIS 3.x

## Global Constraints

- Must not UPDATE the `geom` column directly (it's a GENERATED ALWAYS computed column from `geom_wkt`)
- Must preserve `MultiPolygon` type — fix chain uses `ST_CollectionExtract(geom, 3)` to discard non-polygon artifacts
- Must write audit log to `output/gis_geometry_fix_log_<timestamp>.txt` with ward codes for cross-verification
- Must be idempotent — running twice produces zero change (WHERE clause filters only invalid rows)
- Must verify zero invalid geometries remain after fix
- Must preserve area and coordinates (verified: 0 m² change, 0 coordinate drift)

---
````

### File Structure

```
dataset-generation-scripts/
├── internal/sapnhap_bando/service/
│   └── geometry_fixer.go          ← CREATE: ValidateAndFixGeometries method
├── internal/sapnhap_bando/
│   └── sapnhap_bando.go           ← MODIFY: add ValidateAndFixGeometries() export
├── main.go                        ← MODIFY: add call after PatchIslandProvincesGeometry
└── output/
    └── gis_geometry_fix_log_*.txt ← GENERATED at runtime (gitignored)
```

### Task Decomposition

---

### Task 1: Create geometry_fixer.go with ValidateAndFixGeometries method

**Files:**
- Create: `dataset-generation-scripts/internal/sapnhap_bando/service/geometry_fixer.go`

**Interfaces:**
- Consumes: `SapNhapService.db` (field `*bun.DB` — already exists on struct)
- Produces: `func (s *SapNhapService) ValidateAndFixGeometries(ctx context.Context) error`

- [ ] **Step 1: Write the file with the fix method**

Create `dataset-generation-scripts/internal/sapnhap_bando/service/geometry_fixer.go`:

```go
package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/uptrace/bun"
)

// FixedWardRecord represents a single ward geometry that was corrected
type FixedWardRecord struct {
	Ma           string `bun:"ma"`
	Ten          string `bun:"ten"`
	ProvinceCode string `bun:"vn_ds_province_code"`
	WardCode     string `bun:"vn_ds_ward_code"`
}

// ValidateAndFixGeometries finds and fixes all invalid (self-intersecting) ward
// geometries in sapnhap_geojson_objects. It updates geom_wkt (WKT text), and
// the computed geom column follows automatically.
//
// Fix chain: ST_CollectionExtract(ST_MakeValid(ST_GeomFromText(geom_wkt, 4326)), 3)
//
// An audit log is written to output/gis_geometry_fix_log_<timestamp>.txt with
// the codes and names of every fixed ward for cross-verification.
func (s *SapNhapService) ValidateAndFixGeometries(ctx context.Context) error {
	// Step 1: Count total records for the audit log
	totalCount, err := s.db.NewSelect().
		Model((*FixedWardRecord)(nil)).
		TableExpr("sapnhap_geojson_objects").
		Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to count total records: %w", err)
	}

	// Step 2: Count invalid records before fix
	invalidCount, err := s.db.NewSelect().
		TableExpr("sapnhap_geojson_objects").
		Where("NOT ST_IsValid(ST_GeomFromText(geom_wkt, 4326))").
		Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to count invalid records: %w", err)
	}

	log.Printf("ℹ️ GIS Geometry Check: %d total, %d invalid", totalCount, invalidCount)

	if invalidCount == 0 {
		log.Println("✅ All GIS geometries are valid — no fixes needed")
		return nil
	}

	// Step 3: Execute the fix UPDATE with RETURNING to get affected rows
	fixSQL := `UPDATE sapnhap_geojson_objects
SET geom_wkt = ST_AsText(
    ST_CollectionExtract(
        ST_MakeValid(ST_GeomFromText(geom_wkt, 4326)),
        3
    )
)
WHERE NOT ST_IsValid(ST_GeomFromText(geom_wkt, 4326))
RETURNING ma, ten, vn_ds_province_code, vn_ds_ward_code`

	var fixedWards []FixedWardRecord
	if err := s.db.NewRaw(fixSQL).Scan(ctx, &fixedWards); err != nil {
		return fmt.Errorf("failed to fix geometries: %w", err)
	}

	log.Printf("✅ Fixed %d ward geometries", len(fixedWards))

	// Step 4: Run verification query
	remainingInvalid, err := s.db.NewSelect().
		TableExpr("sapnhap_geojson_objects").
		Where("NOT ST_IsValid(ST_GeomFromText(geom_wkt, 4326))").
		Count(ctx)
	if err != nil {
		return fmt.Errorf("verification query failed: %w", err)
	}

	if remainingInvalid > 0 {
		return fmt.Errorf("verification failed: %d invalid geometries remain after fix", remainingInvalid)
	}
	log.Println("✅ Verification passed: 0 invalid geometries remain")

	// Step 5: Write audit log
	if err := writeAuditLog(totalCount, invalidCount, fixedWards); err != nil {
		log.Printf("⚠️ Failed to write audit log: %v", err)
		// Don't fail the pipeline — fix was already applied successfully
	}

	return nil
}

// writeAuditLog writes a detailed fix log to output/gis_geometry_fix_log_<timestamp>.txt
func writeAuditLog(totalCount, invalidCount int, fixedWards []FixedWardRecord) error {
	// Ensure output directory exists
	outputDir := "output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate timestamped filename
	timestamp := time.Now().Format("2006-01-02__15_04_05")
	filename := filepath.Join(outputDir, fmt.Sprintf("gis_geometry_fix_log_%s.txt", timestamp))

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create audit log file: %w", err)
	}
	defer f.Close()

	// Count unique provinces
	provinceSet := make(map[string]struct{})
	for _, w := range fixedWards {
		if w.ProvinceCode != "" {
			provinceSet[w.ProvinceCode] = struct{}{}
		}
	}

	// Write header
	fmt.Fprintf(f, "GIS Geometry Fix Audit Log\n")
	fmt.Fprintf(f, "Generated: %s ICT\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(f, "============================================================\n")
	fmt.Fprintf(f, "Total records in sapnhap_geojson_objects: %d\n", totalCount)
	fmt.Fprintf(f, "Records checked (invalid): %d\n", invalidCount)
	fmt.Fprintf(f, "Records fixed: %d\n", len(fixedWards))
	fmt.Fprintf(f, "Provinces affected: %d\n", len(provinceSet))
	fmt.Fprintf(f, "\n--- Fixed Wards ---\n")
	fmt.Fprintf(f, "%-8s  %-28s  %-10s  %-10s\n", "ma", "ten", "prov_code", "ward_code")
	fmt.Fprintf(f, "%-8s  %-28s  %-10s  %-10s\n", "--------", "----------------------------", "----------", "----------")

	for _, w := range fixedWards {
		fmt.Fprintf(f, "%-8s  %-28s  %-10s  %-10s\n", w.Ma, w.Ten, w.ProvinceCode, w.WardCode)
	}

	fmt.Fprintf(f, "\n--- Verification ---\n")
	fmt.Fprintf(f, "Remaining invalid geometries: 0\n")
	fmt.Fprintf(f, "Fix command: ST_CollectionExtract(ST_MakeValid(ST_GeomFromText(geom_wkt, 4326)), 3)\n")
	fmt.Fprintf(f, "Applied to: geom_wkt column (computed geom column follows automatically)\n")
	fmt.Fprintf(f, "Safety: Idempotent — re-running produces zero changes\n")

	log.Printf("📄 Audit log written to %s", filename)
	return nil
}
```

- [ ] **Step 2: Verify the file compiles**

```bash
cd dataset-generation-scripts && go build ./...
```
Expected: Compiles successfully (but won't be called yet — no-op until main.go integration)

- [ ] **Step 3: Commit**

```bash
git add dataset-generation-scripts/internal/sapnhap_bando/service/geometry_fixer.go
git commit -m "feat: add ValidateAndFixGeometries to fix self-intersecting ward polygons"
```

---

### Task 2: Export ValidateAndFixGeometries from sapnhap_bando package

**Files:**
- Modify: `dataset-generation-scripts/internal/sapnhap_bando/sapnhap_bando.go` (add after `PatchIslandProvincesGeometry`)

**Interfaces:**
- Consumes: `SapNhapService.ValidateAndFixGeometries(ctx context.Context) error` (from Task 1)
- Produces: Public function `ValidateAndFixGeometries()` (no args, no return — follows existing pattern of `log.Fatalf` on error)

- [ ] **Step 1: Add the public export function**

Add this function to `dataset-generation-scripts/internal/sapnhap_bando/sapnhap_bando.go`, after the `PatchIslandProvincesGeometry` function (after line 81):

```go
// ValidateAndFixGeometries checks all ward geometries in sapnhap_geojson_objects
// for self-intersections and fixes them using ST_MakeValid + ST_CollectionExtract.
// An audit log of all fixed wards is written to output/gis_geometry_fix_log_<timestamp>.txt.
//
// This should be called after PatchIslandProvincesGeometry() and before
// GenerateGISSQLDatasets() in the generation flow.
func ValidateAndFixGeometries() {
	postgresDB := db.GetPostgresDBConnection()
	sapNhapGeoJSONObjectRepository := sapNhapR.NewSapNhapGeoJSONObjectRepository(postgresDB)
	vnRepo := vnRepo.NewVnProvincesTmpRepository(postgresDB)
	sapNhapService := sapNhapService.NewSapNhapService(vnRepo, sapNhapGeoJSONObjectRepository, postgresDB)

	log.Println("ℹ️ Validating and fixing GIS geometries (self-intersection repair)...")
	if err := sapNhapService.ValidateAndFixGeometries(context.Background()); err != nil {
		log.Fatalf("Failed to fix GIS geometries: %v", err)
		panic(err)
	}
	log.Println("✅ GIS geometry validation and fix completed successfully")
}
```

The new file should look like this at the end (showing all 4 public functions):

```go
package sapnhapbando

import (
	"context"
	"log"

	db "github.com/thanglequoc-vn-provinces/v2/internal/database"

	sapNhapR "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/repository"
	sapNhapService "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/service"
	vnRepo "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/repository"
)

func FetchGISDataFromSapNhapBando() {
	// ... (existing code, unchanged)
}

func BackfillProvinceAndWardCodesInSapNhapGeojsonObjects() {
	// ... (existing code, unchanged)
}

func PatchIslandProvincesGeometry() {
	// ... (existing code, unchanged)
}

// ValidateAndFixGeometries checks all ward geometries in sapnhap_geojson_objects
// for self-intersections and fixes them using ST_MakeValid + ST_CollectionExtract.
// An audit log of all fixed wards is written to output/gis_geometry_fix_log_<timestamp>.txt.
//
// This should be called after PatchIslandProvincesGeometry() and before
// GenerateGISSQLDatasets() in the generation flow.
func ValidateAndFixGeometries() {
	postgresDB := db.GetPostgresDBConnection()
	sapNhapGeoJSONObjectRepository := sapNhapR.NewSapNhapGeoJSONObjectRepository(postgresDB)
	vnRepo := vnRepo.NewVnProvincesTmpRepository(postgresDB)
	sapNhapService := sapNhapService.NewSapNhapService(vnRepo, sapNhapGeoJSONObjectRepository, postgresDB)

	log.Println("ℹ️ Validating and fixing GIS geometries (self-intersection repair)...")
	if err := sapNhapService.ValidateAndFixGeometries(context.Background()); err != nil {
		log.Fatalf("Failed to fix GIS geometries: %v", err)
		panic(err)
	}
	log.Println("✅ GIS geometry validation and fix completed successfully")
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd dataset-generation-scripts && go build ./...
```
Expected: Compiles successfully

- [ ] **Step 3: Commit**

```bash
git add dataset-generation-scripts/internal/sapnhap_bando/sapnhap_bando.go
git commit -m "feat: expose ValidateAndFixGeometries as public pipeline step"
```

---

### Task 3: Integrate fix call into main.go pipeline

**Files:**
- Modify: `dataset-generation-scripts/main.go` (line 25 area)

**Interfaces:**
- Consumes: `sapnhap.ValidateAndFixGeometries()` (from Task 2)
- Produces: N/A — pipeline integration only

- [ ] **Step 1: Add the call in main.go**

In `dataset-generation-scripts/main.go`, add `sapnhap.ValidateAndFixGeometries()` after line 24 (`sapnhap.PatchIslandProvincesGeometry()`) and before line 25 (`dataset_writer.GenerateGISSQLDatasets()`):

The file should change from:

```go
func main() {
	// pre-run
	// Refresh temporary dataset, import existing dataset
	db.BootstrapTemporaryDatasetStructure()

	dumper.BeginDumpingDataWithDvhcvnDirectSource()
	dataset_writer.ReadAndGenerateSQLDatasets()

	if (INCLUDE_GIS) {
		db.BootstrapGISDataStructure()
		sapnhap.BackfillProvinceAndWardCodesInSapNhapGeojsonObjects()
		sapnhap.FetchGISDataFromSapNhapBando()
		sapnhap.PatchIslandProvincesGeometry()
		dataset_writer.GenerateGISSQLDatasets()
	}
}
```

To:

```go
func main() {
	// pre-run
	// Refresh temporary dataset, import existing dataset
	db.BootstrapTemporaryDatasetStructure()

	dumper.BeginDumpingDataWithDvhcvnDirectSource()
	dataset_writer.ReadAndGenerateSQLDatasets()

	if (INCLUDE_GIS) {
		db.BootstrapGISDataStructure()
		sapnhap.BackfillProvinceAndWardCodesInSapNhapGeojsonObjects()
		sapnhap.FetchGISDataFromSapNhapBando()
		sapnhap.PatchIslandProvincesGeometry()
		sapnhap.ValidateAndFixGeometries()
		dataset_writer.GenerateGISSQLDatasets()
	}
}
```

- [ ] **Step 2: Verify full compilation**

```bash
cd dataset-generation-scripts && go build -o /tmp/vn-provinces-test ./...
```
Expected: Builds successfully to `/tmp/vn-provinces-test`

- [ ] **Step 3: Commit**

```bash
git add dataset-generation-scripts/main.go
git commit -m "feat: integrate GIS geometry validation/fix into pipeline"
```

---

### Task 4: Test the fix end-to-end

**Files:**
- No file changes — verification-only task

**Prerequisites:** Docker PostgreSQL must be running with `sapnhap_geojson_objects` populated from the previous `go run main.go` run.

- [ ] **Step 1: Verify pre-fix state — 78 invalid records**

```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c \
  "SELECT COUNT(*) FROM sapnhap_geojson_objects WHERE NOT ST_IsValid(geom);"
```
Expected: `78`

- [ ] **Step 2: Run the pipeline with the fix**

```bash
cd dataset-generation-scripts && go run main.go
```
Expected output should include:
```
ℹ️ Validating and fixing GIS geometries (self-intersection repair)...
ℹ️ GIS Geometry Check: 3355 total, 78 invalid
✅ Fixed 78 ward geometries
✅ Verification passed: 0 invalid geometries remain
✅ GIS geometry validation and fix completed successfully
```

- [ ] **Step 3: Verify post-fix state — 0 invalid records**

```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c \
  "SELECT COUNT(*) FROM sapnhap_geojson_objects WHERE NOT ST_IsValid(geom);"
```
Expected: `0`

- [ ] **Step 4: Verify all geometries are still MultiPolygon**

```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c \
  "SELECT DISTINCT ST_GeometryType(geom) FROM sapnhap_geojson_objects;"
```
Expected: single row: `ST_MultiPolygon`

- [ ] **Step 5: Verify audit log exists and has 78 records**

```bash
ls -la dataset-generation-scripts/output/gis_geometry_fix_log_*.txt
```
Expected: One file exists

```bash
grep -c "^[0-9]" dataset-generation-scripts/output/gis_geometry_fix_log_*.txt
```
Expected: `78` (or 78 + 1 for the header divider line — count should be ~79-80 lines total in the fixed wards section)

- [ ] **Step 6: Test idempotency — run again, expect zero fixes**

```bash
cd dataset-generation-scripts && go run main.go
```
Expected output should include:
```
ℹ️ GIS Geometry Check: 3355 total, 0 invalid
✅ All GIS geometries are valid — no fixes needed
```

- [ ] **Step 7: Verify ES import — all 34 provinces**

Re-import all 34 provinces into Elasticsearch:

```bash
python3 << 'PYEOF'
import json, subprocess

with open('dataset-generation-scripts/output/elasticsearch/provinces-gis.ndjson', 'r') as f:
    lines = f.readlines()

# Check that all 34 index lines point to the right _id
docs = []
for i in range(34):
    index_line = json.loads(lines[i*2])
    docs.append(index_line['index']['_id'])

# Import all at once via curl (now that geometries are fixed, this might work)
result = subprocess.run([
    'curl', '-s', '--max-time', '300', '-X', 'POST',
    'localhost:9200/_bulk',
    '-H', 'Content-Type: application/x-ndjson',
    '--data-binary', '@dataset-generation-scripts/output/elasticsearch/provinces-gis.ndjson'
], capture_output=True, text=True, timeout=310)

if result.stdout:
    r = json.loads(result.stdout)
    errors = r.get('errors', True)
    items = r.get('items', [])
    success = sum(1 for item in items if item.get('index', {}).get('status') == 201)
    print(f"Imported: {success}/{len(items)} (errors={errors})")
else:
    print(f"No response (curl exit={result.returncode})")
PYEOF
```
Expected: `Imported: 34/34 (errors=False)` (Lâm Đồng may still fail on document size — that's the separate issue)

- [ ] **Step 8: Commit test verification notes (optional)**

```bash
git add dataset-generation-scripts/output/gis_geometry_fix_log_*.txt
git commit -m "test: add GIS fix audit log after validation run"
```

---

*Plan written 2026-07-20 — based on design spec at `docs/superpowers/specs/2026-07-20-gis-geometry-fix-design.md` and diagnostic report at `development/corrupted_gis_ward_data/report.md`*
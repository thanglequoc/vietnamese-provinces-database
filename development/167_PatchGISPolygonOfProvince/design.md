# Design: Patch GIS Polygon of Province — Merge Island Territories

## Problem

The province-level GIS boundaries for **Da Nang (province_code = 48)** and **Khanh Hoa (province_code = 56)** do not include their respective island territories:

- **Hoàng Sa (Paracel Islands, ward_code = 20333)** — belongs to Da Nang
- **Trường Sa (Spratly Islands, ward_code = 22736)** — belongs to Khanh Hoa

The ward-level geometries are correctly fetched from the `sapnhap.bando.com.vn` API and stored in `sapnhap_geojson_objects`. However, the province-level API response excludes the island polygons entirely.

### Evidence (from database query)

| Province | Province bbox | Island ward | Island bbox | `ST_Contains` | `ST_Intersects` |
|----------|--------------|-------------|-------------|---------------|-----------------|
| Da Nang (ma=48) | `[107.21, 14.95, 108.74, 16.23]` | Hoàng Sa (ma=20333) | `[111.19, 15.69, 112.74, 17.12]` | `false` | `false` |
| Khanh Hoa (ma=56) | `[108.55, 11.31, 109.46, 12.87]` | Trường Sa (ma=22736) | `[109.47, 7.18, 117.83, 11.53]` | `false` | `false` |

All other wards in both provinces spatially intersect their parent province polygon (93/94 for Da Nang, 64/65 for Khanh Hoa). The only wards that don't intersect at all are the two island territories — confirming this is an upstream GIS data defect, not a boundary precision issue.

### Root Cause

The upstream `sapnhap.bando.com.vn` API returns province geometries that exclude island territories. The ward-level geometries for the islands are fetched via separate API calls (using their own `malk` IDs) and are correct. The province-level API response simply doesn't include those polygons.

## Approach: Post-Fetch SQL Patch Step (Approach A)

Add a new step in the generation flow, executed **after** `FetchGISDataFromSapNhapBando()` and **before** `GenerateGISSQLDatasets()`. This step merges the island ward geometries into their parent province geometries using PostGIS `ST_Union`.

### Why this approach?

- **Leverages existing data:** The island ward geometries are already in the database from the API fetch — no additional API calls or local files needed
- **Clean separation:** The patch is a distinct step, easy to audit and reason about
- **Automatic:** Integrated into the generation flow, so the fix is applied every time the dataset is regenerated
- **Follows existing patterns:** Similar to how An Giang has a special-case patch in `processGeoJSONObject()`, but cleaner because it's a separate step rather than inline logic
- **PostGIS-native:** Uses `ST_Union` and `ST_Envelope` which are the correct tools for geometry merging and bbox recalculation

### Why not the alternatives?

- **Approach B (Go code special-case in `processGeoJSONObject`):** Would require either Go-level geometry manipulation (complex, no Go GIS library in project) or a DB round-trip mid-processing. Mixes concerns — the fetch step should fetch, the patch step should patch.
- **Approach C (standalone SQL patch script):** Not integrated into the automated generation flow. Users would need to manually run it, and generated datasets wouldn't automatically reflect the fix.

## Implementation

### 1. New Go function to execute the patch

**File:** `dataset-generation-scripts/internal/sapnhap_bando/service/sapnhap.go`

Add a new method to `SapNhapService`:

```go
// PatchIslandProvincesGeometry merges island ward geometries into their parent
// province geometries. This fixes an upstream GIS data defect where the province-level
// API response from sapnhap.bando.com.vn excludes island territories (Hoàng Sa and
// Trường Sa) that are present at the ward level.
//
// Affected provinces:
//   - Da Nang (ma=48) ← Hoàng Sa (ma=20333)
//   - Khanh Hoa (ma=56) ← Trường Sa (ma=22736)
func (s *SapNhapService) PatchIslandProvincesGeometry(ctx context.Context) error {
    // Merge Hoàng Sa into Da Nang
    err := s.mergeWardGeometryIntoProvince(ctx, "48", "20333")
    if err != nil {
        return fmt.Errorf("failed to patch Da Nang with Hoàng Sa geometry: %w", err)
    }
    log.Println("✅ Patched Da Nang (48) with Hoàng Sa (20333) island geometry")

    // Merge Trường Sa into Khanh Hoa
    err = s.mergeWardGeometryIntoProvince(ctx, "56", "22736")
    if err != nil {
        return fmt.Errorf("failed to patch Khanh Hoa with Trường Sa geometry: %w", err)
    }
    log.Println("✅ Patched Khanh Hoa (56) with Trường Sa (22736) island geometry")

    return nil
}

// mergeWardGeometryIntoProvince merges the geometry of a ward (identified by wardMa)
// into the geometry of a province (identified by provinceMa) using ST_Union.
// Both geom_wkt and bbox_wkt are updated for the province record.
func (s *SapNhapService) mergeWardGeometryIntoProvince(ctx context.Context, provinceMa, wardMa string) error {
    _, err := s.db.ExecContext(ctx, `
        UPDATE sapnhap_geojson_objects
        SET geom_wkt = ST_AsText(ST_Union(geom, (SELECT geom FROM sapnhap_geojson_objects WHERE ma = ?))),
            bbox_wkt = ST_AsText(ST_Envelope(ST_Union(geom, (SELECT geom FROM sapnhap_geojson_objects WHERE ma = ?))))
        WHERE ma = ?`,
        wardMa, wardMa, provinceMa)
    return err
}
```

### 2. New entry point function

**File:** `dataset-generation-scripts/internal/sapnhap_bando/sapnhap_bando.go`

Add a new exported function:

```go
// PatchIslandProvincesGeometry merges island ward geometries (Hoàng Sa, Trường Sa)
// into their parent province geometries (Da Nang, Khanh Hoa). This fixes an upstream
// GIS data defect where the province-level API response excludes island territories.
func PatchIslandProvincesGeometry() {
    postgresDB := db.GetPostgresDBConnection()
    sapNhapGeoJSONObjectRepository := sapNhapR.NewSapNhapGeoJSONObjectRepository(postgresDB)
    vnRepo := vnRepo.NewVnProvincesTmpRepository(postgresDB)
    sapNhapService := sapNhapService.NewSapNhapService(vnRepo, sapNhapGeoJSONObjectRepository, postgresDB)

    log.Println("ℹ️ Patching island province geometries (Hoàng Sa → Da Nang, Trường Sa → Khanh Hoa)...")
    if err := sapNhapService.PatchIslandProvincesGeometry(context.Background()); err != nil {
        log.Fatalf("Failed to patch island province geometries: %v", err)
        panic(err)
    }
    log.Println("✅ Island province geometry patching completed successfully")
}
```

### 3. Integration into main.go

**File:** `dataset-generation-scripts/main.go`

Add the patch step after `FetchGISDataFromSapNhapBando()` and before `GenerateGISSQLDatasets()`:

```go
if (INCLUDE_GIS) {
    db.BootstrapGISDataStructure()
    sapnhap.BackfillProvinceAndWardCodesInSapNhapGeojsonObjects()
    sapnhap.FetchGISDataFromSapNhapBando()
    sapnhap.PatchIslandProvincesGeometry()  // ← NEW: merge islands into provinces
    dataset_writer.GenerateGISSQLDatasets()
}
```

## Expected Result

After the patch:

| Province | Before bbox | After bbox (expected) | `ST_Contains(island)` |
|----------|------------|----------------------|----------------------|
| Da Nang (48) | `[107.21, 14.95, 108.74, 16.23]` | `[107.21, 14.95, 112.74, 17.12]` | `true` |
| Khanh Hoa (56) | `[108.55, 11.31, 109.46, 12.87]` | `[108.55, 7.18, 117.83, 12.87]` | `true` |

The province geometries will spatially contain all their administrative subdivisions, including the Paracel and Spratly Islands.

## Testing

### Manual verification (SQL queries)

After running the generation script, verify with:

```sql
-- Verify containment after patch
SELECT 
  'Hoang Sa within Da Nang' as check_name,
  ST_Contains(
    (SELECT geom FROM sapnhap_geojson_objects WHERE ma = '48'),
    (SELECT geom FROM sapnhap_geojson_objects WHERE ma = '20333')
  ) as is_contained
UNION ALL
SELECT
  'Truong Sa within Khanh Hoa',
  ST_Contains(
    (SELECT geom FROM sapnhap_geojson_objects WHERE ma = '56'),
    (SELECT geom FROM sapnhap_geojson_objects WHERE ma = '22736')
  );
-- Expected: both true

-- Verify bbox expansion
SELECT ma, ten,
  ST_XMin(bbox) as xmin, ST_YMin(bbox) as ymin,
  ST_XMax(bbox) as xmax, ST_YMax(bbox) as ymax
FROM sapnhap_geojson_objects
WHERE ma IN ('48', '56');
-- Expected: Da Nang xmax ~112.74, Khanh Hoa xmin ~109.47, ymin ~7.18, xmax ~117.83
```

### Unit test

Add a test in `dataset-generation-scripts/internal/sapnhap_bando/service/sapnhap_test.go` that verifies:
1. After `PatchIslandProvincesGeometry()`, `ST_Contains` returns `true` for both island-province pairs
2. The province bbox has expanded to include the island territories

## Scope

- **In scope:** Merge Hoàng Sa geometry into Da Nang, merge Trường Sa geometry into Khanh Hoa
- **Out of scope:** Fixing boundary precision issues for other wards (those are `ST_Contains` false but `ST_Intersects` true — a different, minor issue not addressed by this patch)
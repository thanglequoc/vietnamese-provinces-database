# Design: Fix GIS Geometry Self-Intersections for Elasticsearch Compatibility

**Date:** 2026-07-20
**Status:** Draft
**Approach:** Shared pipeline PostGIS fix (ST_MakeValid + ST_CollectionExtract)

---

## Objective

Fix 78 ward-level GIS geometries in `sapnhap_geojson_objects` that have self-intersecting polygon rings, causing Elasticsearch `geo_shape` import failures for 22 provinces. Apply `ST_MakeValid()` + `ST_CollectionExtract(3)` as a one-time pipeline step after GIS data ingestion, before any dataset export writer runs.

## Background

### Problem

During Elasticsearch `provinces-gis.ndjson` import testing, 13/34 provinces failed with:
- 11 provinces: `Polygon self-intersection at lat=... lon=...`
- 1 province (Hồ Chí Minh): `at least three non-collinear points required`
- 1 province (Lâm Đồng): document too large (separate issue)

PostGIS `ST_IsValid()` cross-validation confirmed **78 ward geometries** are invalid across **22 provinces** — all self-intersections. Zero province-level geometries are invalid. The root cause is coordinate precision rounding (~11cm at 6 decimal places) creating microscopic "bow-tie" intersections in polygon rings.

### Why Fix in Shared Pipeline

- `ST_MakeValid()` + `ST_CollectionExtract(3)` fixes all 78 wards with **zero area change, zero coordinate change**
- Applied once benefits all downstream consumers: Elasticsearch, PostGIS exports, MySQL GIS, MongoDB GIS, GeoJSON exports
- Safer than fixing in individual writers (ES-only fix leaves other consumers with invalid data)
- `ST_CollectionExtract(..., 3)` prevents the `GeometryCollection` type issue — output is always `MultiPolygon`

### Key Constraint

The `geom` column in `sapnhap_geojson_objects` is a **GENERATED ALWAYS** computed column:
```sql
geom geometry(Multipolygon, 4326) GENERATED ALWAYS AS (ST_GeomFromText(geom_wkt, 4326)) STORED
```
We cannot UPDATE `geom` directly. The fix must update `geom_wkt` (plain text), and the computed column follows automatically.

## Design

### New File: `internal/sapnhap_bando/service/geometry_fixer.go`

A single function:

```go
func (s *SapNhapService) ValidateAndFixGeometries(ctx context.Context) error
```

**Responsibilities:**
1. Execute the fix SQL with RETURNING clause to get affected rows
2. Write an audit log file to `output/gis_geometry_fix_log_<timestamp>.txt`
3. Run a verification query to confirm zero remaining invalid geometries
4. Return error if verification fails

### Fix SQL

```sql
UPDATE sapnhap_geojson_objects
SET geom_wkt = ST_AsText(
    ST_CollectionExtract(
        ST_MakeValid(ST_GeomFromText(geom_wkt, 4326)),
        3
    )
)
WHERE NOT ST_IsValid(ST_GeomFromText(geom_wkt, 4326))
RETURNING ma, ten, vn_ds_province_code, vn_ds_ward_code;
```

**How it works:**
1. `ST_GeomFromText(geom_wkt, 4326)` — parse the WKT text into a PostGIS geometry
2. `ST_MakeValid(...)` — untangle self-intersections (may produce GeometryCollection with polygon + ephemeral line/point)
3. `ST_CollectionExtract(..., 3)` — extract only polygons, discard linestrings/points, returns pure MultiPolygon
4. `ST_AsText(...)` — convert back to WKT text for storage
5. `WHERE NOT ST_IsValid(...)` — only fix invalid geometries

### Pipeline Integration

In `main.go`, after `sapnhap.FetchGISData()` and before any dataset writer:

```go
// Existing flow
dumper.DumpAdministrativeData()
sapnhap.FetchGISData()

// NEW: Fix invalid geometries before export
if err := sapnhapService.ValidateAndFixGeometries(ctx); err != nil {
    log.Fatalf("Failed to fix GIS geometries: %v", err)
}

// Existing flow
dataset_writer.WriteAllFormats()
```

### Audit Log Format

File: `output/gis_geometry_fix_log_2026-07-20__22_48_00.txt`

```
GIS Geometry Fix Audit Log
Generated: 2026-07-20 22:48:00 ICT
============================================================
Total records in sapnhap_geojson_objects: 3,355
Records checked (invalid): 78
Records fixed: 78
Provinces affected: 22

--- Fixed Wards ---
ma       ten                          prov_code  ward_code
00820    Xã Yên Minh                  08         00820
00832    Xã Bạch Đích                 08         00832
00958    Xã Thượng Sơn                08         00958
01075    Xã Nậm Dịch                  08         01075
01096    Xã Pà Vầy Sủ                 08         01096
... (all 78 records)

--- Verification ---
Remaining invalid geometries: 0
All geometry types: ST_MultiPolygon
Area preserved: confirmed (0 m² difference)

--- Fix Summary ---
Fix command: ST_CollectionExtract(ST_MakeValid(geom), 3)
Applied to: geom_wkt column (computed geom column follows automatically)
Safety: Idempotent — re-running produces zero changes
```

### Error Handling

| Scenario | Behavior |
|----------|----------|
| Fix SQL fails | Return error, abort pipeline (no partial fix) |
| Verification query returns invalid > 0 | Log warning, return error |
| Audit log file write fails | Log warning to stderr, continue (fix already applied) |
| Zero invalid records found | Log "No fixes needed", skip audit log, continue normally |
| Re-run (idempotent) | WHERE clause finds 0 rows, logs "No fixes needed" |

## Verification Queries (Manual Cross-Check)

After the fix runs, the user can verify:

```sql
-- 1. Zero invalid geometries remain
SELECT COUNT(*) FROM sapnhap_geojson_objects WHERE NOT ST_IsValid(geom);
-- Expected: 0

-- 2. All are MultiPolygon (no GeometryCollection leakage)
SELECT DISTINCT ST_GeometryType(geom) FROM sapnhap_geojson_objects;
-- Expected: ST_MultiPolygon (single row)

-- 3. Area preservation (requires backup)
SELECT a.ma, a.ten, 
  ROUND(ST_Area(a.geom::geography) - ST_Area(b.geom::geography), 2) AS diff_m2
FROM sapnhap_geojson_objects a
JOIN sapnhap_geojson_objects_backup b ON a.ma = b.ma
WHERE ABS(ST_Area(a.geom::geography) - ST_Area(b.geom::geography)) > 0.01;
-- Expected: 0 rows
```

## Files Affected

| File | Action | Details |
|------|--------|---------|
| `internal/sapnhap_bando/service/geometry_fixer.go` | **NEW** | `ValidateAndFixGeometries` function |
| `main.go` | **MODIFY** | Add one line call after `FetchGISData` |
| `output/gis_geometry_fix_log_*.txt` | **GENERATED** | Audit log (gitignored) |

## Risks & Mitigations

| Risk | Severity | Mitigation |
|------|----------|------------|
| Type change (MultiPolygon → GeometryCollection) | Medium | `ST_CollectionExtract(..., 3)` ensures output is always MultiPolygon |
| Coordinate drift from MakeValid | Low | Verified: 0 m² area change, 0 coordinate change across all 78 wards |
| Fix breaks other consumers (MySQL, Mongo) | Low | Fix in shared pipeline — all consumers get corrected data simultaneously |
| Generated column constraint blocks UPDATE | Low | Fix updates `geom_wkt` text, not the computed `geom` column |
| Idempotency | Low | WHERE clause ensures only invalid rows are touched; re-running does nothing |

## Out of Scope

- Lâm Đồng document size issue (12MB ES bulk limit — separate fix via ES config or split import)
- Province-level geometry fixes (none needed — all 34 province geometries are valid)
- Source GeoJSON file modification (fix is at database level, source files unchanged)
- GeoJSON export format changes (WKT → GeoJSON conversion already exists in writers)

## Testing Plan

1. **Pre-fix**: Query `SELECT COUNT(*) FROM sapnhap_geojson_objects WHERE NOT ST_IsValid(geom)` → expect 78
2. **Run fix**: `go run main.go` with the new step
3. **Post-fix**: Query again → expect 0
4. **ES import**: Re-import `provinces-gis.ndjson` → expect 34/34
5. **Audit log**: Verify 78 records logged with correct ward codes
6. **Re-run**: Run fix again → expect "No fixes needed" (idempotent)

---

*Design based on diagnostic report at `development/corrupted_gis_ward_data/report.md`*
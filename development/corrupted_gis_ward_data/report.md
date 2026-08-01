# Corrupted GIS Ward Data — Diagnostic Report

**Date:** 2026-07-18
**Updated:** 2026-07-19 (corrected via live ES import testing)
**Status:** Draft

---

## Executive Summary

During Elasticsearch dataset import of `provinces-gis.ndjson`, **13 out of 34 provinces** (38%) failed to index. Root cause investigation via direct ES import testing identified **three distinct geometry problems** that PostGIS accepts but Elasticsearch `geo_shape` rejects:

| Failure Type | Provinces Affected | Root Cause |
|---|---|---|
| **Self-intersecting polygon** | 11 | Ward boundary rings have "bow-tie" topology (edges cross themselves) |
| **Collinear polygon ring** | 1 | Dầu Tiếng ward (Hồ Chí Minh) has a polygon ring with all points on a line |
| **Document too large** | 1 | Lâm Đồng's 12MB document exceeds single-node ES bulk queue limit |

**Note:** An earlier offline scan incorrectly identified Kiên Lương ward (An Giang) as having collinear rings. Live ES import testing proved this was a **false positive** — Kiên Lương's geometry imports successfully. The offline collinearity tolerance was too strict.

---

## Problem Explanation (Non-GIS Audience)

### 1. Self-Intersecting Polygons (aka "Bow-Tie" Polygons)

A valid polygon's boundary should be a simple closed loop — like drawing a shape without lifting the pen. A self-intersecting polygon has edges that cross each other, creating a "bow-tie" or "figure-eight" shape:

```
Valid Polygon:               Self-Intersecting (Bow-Tie):

  ┌─────────┐                  A─────B
  │         │                  │╲   ╱│
  │  Area   │                  │ ╲ ╱ │
  │         │                  │  ╳  │   ← edges cross here
  └─────────┘                  │ ╱ ╲ │
                               │╱   ╲│
                               D─────C
```

**Why Elasticsearch rejects this:** The OGC Simple Features specification says polygon rings must not self-intersect. ES validates this at index time. PostGIS is more lenient — it stores self-intersecting polygons and only flags them as invalid when you explicitly call `ST_IsValid()`.

**How widespread:** 11/14 failing provinces (79%) fail due to self-intersections. These are the dominant problem.

### 2. Collinear Polygon Rings

```
Valid Triangle:              Collinear Ring:

    A                           A──B──C──D──E
   ╱ ╲                         All 5 points on
  ╱   ╲                        one straight line.
 B─────C                       Area = 0. Invalid.
```

**Why Elasticsearch rejects this:** `geo_shape` requires at least 3 non-collinear points to define a polygon with non-zero area. PostGIS accepts zero-area collinear rings.

**How widespread:** Only 1/14 failing provinces — Đầu Tiếng ward in Hồ Chí Minh. The collinear ring is a tiny 5-point, ~1-meter fragment alongside a valid 1,133-point main polygon.

### 3. Document Too Large

Lâm Đồng's province document is 12MB — the largest in the dataset due to extensive ward geometry. Single-node Elasticsearch has a default memory queue limit that rejects this. Multi-node clusters or split imports solve this.

---

## Full Test Results (Live ES Import)

Each province was individually imported via `curl -X POST localhost:9200/_bulk` with the exact NDJSON from `dataset-generation-scripts/output/elasticsearch/provinces-gis.ndjson`.

### Self-Intersecting Polygon Failures (11 provinces)

| Province Code | Province | Self-Intersection Location |
|---|---|---|
| 08 | Tuyên Quang | lat=23.135° lon=105.043° |
| 11 | Điện Biên | lat=21.284° lon=103.013° |
| 12 | Lai Châu | lat=22.429° lon=103.483° |
| 15 | Lào Cai | lat=21.899° lon=104.386° |
| 19 | Thái Nguyên | lat=22.003° lon=105.751° |
| 33 | Hưng Yên | lat=20.439° lon=106.301° |
| 38 | Thanh Hoá | lat=19.753° lon=105.442° |
| 44 | Quảng Trị | lat=17.742° lon=106.418° |
| 52 | Gia Lai | lat=13.946° lon=109.110° |
| 80 | Tây Ninh | lat=11.391° lon=106.020° |
| 92 | Cần Thơ | lat=10.109° lon=105.439° |

All errors report: `Polygon self-intersection at lat=... lon=...`

### Collinear Ring Failure (1 province)

| Province Code | Province | Ward Affected |
|---|---|---|
| 79 | Hồ Chí Minh | 25777 Dầu Tiếng (ward polygon ring with <3 non-collinear points) |

Error: `at least three non-collinear points required`

### Document Too Large (1 province)

| Province Code | Province | Doc Size |
|---|---|---|
| 68 | Lâm Đồng | ~12MB (exceeds single-node ES bulk queue) |

Error: `es_rejected_execution_exception` (coordinating_and_primary_bytes limit reached)

### False Positive — Corrected (1 province)

| Province Code | Province | Original Scan | Live ES Import |
|---|---|---|---|
| 91 | An Giang | Flagged as collinear | ✅ **Imports successfully** |

Kiên Lương ward (30787) in An Giang was incorrectly flagged by an offline collinearity scan with an overly strict tolerance (`1e-10` cross-product). Live ES import shows the geometry passes `geo_shape` validation. The ward's two largest rings (1,302 and 1,524 points) contain sufficient non-collinear points to satisfy ES.

---

## Affected Wards — Detailed (Verified via Live ES)

### Self-Intersection Failures

The self-intersection errors occur at the **ward level** (Wards[].GIS.Geometry). ES does not reveal which specific ward caused the error — it reports the first self-intersection found in the entire province document. To identify the exact ward, each ward's geometry would need to be individually extracted and tested, similar to the Kiên Lương test methodology.

The self-intersection coordinates are scattered across Vietnam geographically — from Lào Cai (north, 22.4°N) to Cần Thơ (south, 10.1°N) — confirming this is a systematic data quality issue, not a regional artifact.

### Collinear Ring Failure — Ward 25777 Dầu Tiếng

| Field | Value |
|-------|-------|
| **Code** | 25777 |
| **Full Name (VN)** | Xã Dầu Tiếng |
| **Full Name (EN)** | Dau Tieng Commune |
| **Province** | 79 — Hồ Chí Minh |
| **Geometry** | MultiPolygon, 3 polygons |

| Polygon | Points | Unique | Non-Collinear | Status |
|---------|--------|--------|---------------|--------|
| 0 | 5 | 4 | 0 | ⚠️ COLLINEAR |
| 1 | 7 | 6 | 0 | ⚠️ COLLINEAR |
| 2 | 1,133 | 1,122 | 1,055 | ✅ Valid |

Both collinear rings span ~1 meter. The main ring (1,133 pts) is fully valid. Deleting the two tiny collinear polygons would make the document import successfully.

---

## Impact Assessment

### Elasticsearch `provinces-gis` Index

| Category | Count | Percentage |
|----------|-------|------------|
| Total provinces | 34 | 100% |
| Successfully indexed | 21 | 62% |
| Failed — self-intersections | 11 | 32% |
| Failed — collinear ring | 1 | 3% |
| Failed — document too large | 1 | 3% |

### Other Output Formats

**No impact.** PostgreSQL, MySQL, MSSQL, Oracle, MongoDB, Redis, raw JSON, and raw GeoJSON all accept these geometries without validation errors. This is exclusively an Elasticsearch `geo_shape` strictness issue.

---

## Root Cause Analysis

### Why does PostGIS accept these but Elasticsearch doesn't?

| Validation Rule | PostGIS | Elasticsearch geo_shape |
|---|---|---|
| Self-intersecting rings | Accepted (stored; `ST_IsValid` returns false) | **Rejected at index time** |
| Collinear rings (zero area) | Accepted (stored; `ST_Area` returns 0) | **Rejected at index time** |
| OGC Simple Features compliance | Optional (via `ST_MakeValid`) | **Mandatory** |

Both PostGIS and ES are correct under different interpretations of the OGC standard:

- **PostGIS** follows the "store everything, validate on demand" philosophy — useful for ingestion of real-world data with known quality issues
- **Elasticsearch** enforces OGC validity at index time — useful for ensuring query-time correctness of spatial operations

### Where do the self-intersections come from?

The geometries originate from the **sapnhap.bando.com.vn** GIS data source, loaded and transformed by `dataset-generation-scripts/internal/sapnhap_bando/`. Self-intersections commonly arise from:

1. **Coordinate precision rounding**: Complex boundaries with many decimal places, when rounded to 6 decimal places (~11cm precision), can create microscopic self-intersections at narrow vertices
2. **Topology simplification**: Douglas-Peucker or similar simplification algorithms can introduce self-intersections at sharp corners
3. **Source data artifacts**: The upstream SAPNhap data may contain self-intersecting geometries natively

---

## Recommended Correction Strategy

### Three Distinct Fixes Required

| Problem | Fix Location | Approach |
|---------|-------------|----------|
| Self-intersecting polygons (11 provs) | `internal/dataset_writer/` ES writer | Apply `ST_MakeValid()` equivalent in Go (or buffer(0) trick) to fix bow-tie topology |
| Collinear rings (1 prov) | `internal/dataset_writer/` ES writer | Drop polygon rings with <3 non-collinear unique points before writing |
| Document too large (1 prov) | ES import process | Split Lâm Đồng across multiple bulk requests or increase ES heap |

### Primary Fix: Geometry Sanitizer in ES Writer

Add a pre-write validation/normalization step in the Go Elasticsearch dataset writer:

1. **For each `GIS.Geometry`** (province and ward level), iterate over MultiPolygon rings
2. **Detect self-intersections**: Cross-product check on adjacent ring segments; if intersection found, apply a small buffer (e.g., 0.000001°) to untangle the self-intersection
3. **Detect collinear rings**: Count non-collinear points; drop rings with <3 non-collinear unique points
4. **Drop, don't corrupt**: If normalization fails, drop the problematic polygon ring rather than the entire ward/province

### Alternatives Considered

| Approach | Pros | Cons |
|----------|------|------|
| **Fix in ES writer only** (recommended) | Minimal blast radius; only affects ES output | Fix is in Go, not SQL |
| **Fix in PostGIS pipeline** via `ST_MakeValid()` | Fixes all downstream consumers at once | Broader impact; needs regression across all output formats |
| **Fix source GeoJSON** before ingestion | Root cause fix; clean data | Requires modifying upstream data; fragile to re-imports |
| **Accept partial import** (21/34) | Zero code change | 38% data loss in GIS index |

---

## Files for Reference

| File | Description |
|------|-------------|
| `dataset-generation-scripts/output/elasticsearch/provinces-gis.ndjson` | Full export with problematic geometries |
| `dataset-generation-scripts/resources/gis/sapnhapbando_geojson/` | Source GeoJSON files |
| `dataset-generation-scripts/internal/sapnhap_bando/` | GIS data ingestion pipeline |
| `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/` | Export writers (ES, JSON, SQL) |

---

*Report initially generated 2026-07-18 by offline geometry analysis; corrected 2026-07-19 via live Elasticsearch import testing on `machine.thanglequoc.xyz`*
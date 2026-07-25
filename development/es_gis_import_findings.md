# Elasticsearch provinces-gis NDJSON Import: Findings & Struggles

**Date:** 2026-07-24  
**Status:** ✅ RESOLVED — All 34/34 provinces imported successfully  
**ES Node:** Docker on OVHCloud VPS, accessible via SSH tunnel on `localhost:9200`

---

## 1. File Size Profile

| Metric | Value |
|--------|-------|
| File | `provinces-gis.ndjson` |
| Total size | ~158 MB (165,859,390 bytes) |
| Lines | 68 (34 index + 34 data lines) |
| Doc count | 34 province-level documents |
| Per-doc range | 1.8 MB – 11.6 MB |
| Average doc size | 4.7 MB |
| Largest doc | Code `68` (Lâm Đồng), 11.6 MB |

Top 5 largest:
| Province | Code | Size | Wards | Heap cost during index |
|----------|------|------|-------|------------------------|
| Lâm Đồng | 68 | 11.6 MB | 124 | ~48.6 MB (4× inflation) |
| Gia Lai | 52 | 9.2 MB | ~100+ | ~37 MB |
| Đắk Lắk | 66 | 9.0 MB | ~100+ | ~36 MB |
| Tuyên Quang | 08 | 8.2 MB | ~100+ | ~33 MB |
| Đồng Nai | 75 | 8.2 MB | ~100+ | ~33 MB |

> Note: Provinces with large MultiPolygon geometries (offshore islands, complex mountain boundaries) have the largest docs regardless of ward count.

## 2. Import Attempts & Results

### Attempt 1: Full bulk import (158 MB) — ❌ Failed
```
POST /_bulk with entire NDJSON
→ HTTP 413 / timeout — exceeds http.max_content_length of 100 MB
```

### Attempt 2: Chunked bulk import (3 chunks, ~55/46/64 MB) — ❌ Failed
```
POST /_bulk with 12 docs (55.5 MB)
→ 429 es_rejected_execution_exception on ALL docs
  Reason: primary_operation_bytes accumulated across bulk batch
  exceeded max_primary_bytes=53,687,091 (10% of 512 MB heap)
```

### Attempt 3: One-at-a-time import (34 individual curl calls) — ⚠️ Partial
```
Result: 33/34 succeeded ✅, 1 failed ❌ (doc 68 — Lâm Đồng)

Failed doc details:
  Code: 68 (Lâm Đồng)
  Raw JSON size: 12,143,142 bytes (12.1 MB)
  Inflates to during indexing: 48,576,168 bytes (4x expansion)
  Shard write limit: 53,687,091 bytes (10% of 512 MB heap)
  Margin: 5,110,923 bytes (9.5% remaining — too tight)
```

### Attempt 4: Retry doc 68 after _flush + _refresh — ❌ Failed
```
→ Same 429 error — the breaker accumulates across multiple write attempts
  within the same indexing request cycle
```

### Attempt 5: Relax circuit breakers — ❌ Failed
```
Set indices.breaker.request.limit to 95%
Set indices.breaker.total.limit to 90%
Set number_of_replicas=0, refresh_interval=-1
→ Same 429 error — the issue is the shard-level write limit
  (max_primary_bytes), not the cluster-level breaker
```

### Attempt 6: Coordinate precision reduction — ❌ Negligible
```
Reduced coordinates to 6 decimal places, whitespace compaction
→ Savings: 13 bytes (0.0%) — coordinates already at optimal precision
  The issue is coordinate count (530,106 pairs), not formatting
```

### Attempt 7: Increase ES Heap 512 MB → 2 GB + Chunked Bulk — ✅ SUCCESS

**Root cause fix:** Increased JVM heap from `-Xms512m -Xmx512m` to `-Xms2g -Xmx2g` on the VPS Docker container. This raised the `max_primary_bytes` limit from ~54 MB to ~214 MB, comfortably accommodating the 48.6 MB inflation of doc 68.

**Import strategy:**
1. Updated `ES_JAVA_OPTS` in docker-compose on VPS
2. Restarted ES container (data volume persisted)
3. Deleted and re-created `provinces-gis` index with mapping
4. Set `number_of_replicas=0`, `refresh_interval=-1` for bulk speed
5. Split 158 MB NDJSON into 2 chunks (87.8 MB + 70.4 MB), both under 100 MB `http.max_content_length`
6. Bulk-imported each chunk — **0 errors on both**
7. Restored `refresh_interval=1s`

**Results:**
| Chunk | Size | Docs | Took | Errors |
|-------|------|------|------|--------|
| 1 | 87.81 MB | 22 | 13.8s | 0 |
| 2 | 70.37 MB | 12 | 10.4s | 0 |
| **Total** | **158 MB** | **34** | **24.2s** | **0** |

## 3. Root Cause Analysis

### The JVM Heap Constraint (Original Problem)

The ES node had **512 MB heap** (JVM `-Xmx512m`). Elasticsearch reserves portions:

| Component | Limit (512 MB heap) | Limit (2 GB heap) |
|-----------|---------------------|-------------------|
| Total heap | 536,870,912 bytes | 2,147,483,648 bytes |
| Indexing write buffer (10%) | 53,687,091 bytes | 214,748,365 bytes |
| Doc 68 inflates to | 48,576,168 (90.5% of limit) | 22.6% of limit ✅ |
| **Remaining margin** | 5,110,923 (9.5% — too tight) | 166 MB (77.4% — safe) |

### Why 4× Inflation?

The 12.1 MB JSON doc expands during indexing because:
1. **Nested fields**: `Wards` is type `nested` → each ward becomes a separate internal Lucene document
2. **geo_shape indexing**: The Geometry MultiPolygon is parsed into a spatial index structure (BKD tree) which is more memory-intensive than raw JSON
3. **Text analysis**: All `text` fields are tokenized and stored with position data
4. **Coord overhead**: 530,106 coordinate pairs × ~16 bytes each (in-memory representation) ≈ 8.5 MB just for coordinates

```
Raw JSON:   12.1 MB
  ↓ parse into objects
In-memory:  ~48.6 MB  (4×)
  ↓ serialize + store
On-disk:    ~20 MB per doc (estimated from 343 MB / 34 docs)
```

## 4. Solution Applied

### Increase ES Heap (The Fix That Worked)

```yaml
# /home/thanglequoc/docker/database/vietnamese-provinces-docker-compose.yaml
elasticsearch:
  image: docker.elastic.co/elasticsearch/elasticsearch:9.4.3
  environment:
    ES_JAVA_OPTS: "-Xms2g -Xmx2g"  # was -Xms512m -Xmx512m
```

With 1 GB heap (tested and verified):
- `max_primary_bytes` = ~107 MB (10% of 1 GB heap)
- Doc 68 inflation (48.6 MB) = only 45.3% of limit — comfortable margin
- VPS has 12 GB RAM, so 1 GB heap is well within capacity
- Heap usage after full import: 63% (680 MB) — healthy, no GC pressure
- Import time: 32.6s (vs 24.2s with 2 GB — slightly slower due to more frequent GC)

### 1 GB vs 2 GB Comparison

| Metric | 1 GB heap | 2 GB heap |
|--------|-----------|-----------|
| Chunk 1 (22 docs) | 18.6s, 0 errors | 13.8s, 0 errors |
| Chunk 2 (12 docs) | 14.0s, 0 errors | 10.4s, 0 errors |
| Total import time | 32.6s | 24.2s |
| Post-import heap usage | 63% (680 MB) | 77% (1.67 GB) |
| Doc 68 (Lâm Đồng) | ✅ Success | ✅ Success |
| Spatial queries | ✅ Working | ✅ Working |

> **Conclusion:** 1 GB heap is sufficient for the provinces-gis import. 2 GB is faster but not required. The current deployment uses 1 GB.

### Chunked Bulk Import (For http.max_content_length)

The 158 MB NDJSON exceeds ES's default `http.max_content_length` of 100 MB. Splitting into 2 chunks under 90 MB solves this **without changing any ES config** — `http.max_content_length` remains at its default value of 100 MB (104,857,600 bytes).

## 5. Final State

| Metric | Value |
|--------|-------|
| Index | `provinces-gis` |
| Provinces imported | **34/34** ✅ |
| Store size | 343.4 MB |
| Nested doc count (internal) | 3,355 (wards × nested type) |
| Heap | 1 GB (63% used after import) |
| Spatial queries | ✅ Working (verified Hà Nội + Lâm Đồng) |
| Index health | green |

## 6. Generator-Produced Chunked NDJSON

The generator (`elasticsearch_file_writer.go`) now automatically splits the GIS NDJSON
into chunk files (~70 MB each) when the total size exceeds 70 MB. This eliminates the
need for any import scripts — users simply `curl` each chunk file to the ES `_bulk` API.

**Output files when chunked:**
- `provinces-gis-part-01.ndjson` — first batch of province docs
- `provinces-gis-part-02.ndjson` — second batch (if needed)
- `provinces-gis.ndjson.manifest` — lists chunk filenames in order

**Import via direct curl:**
```bash
curl -X POST "localhost:9200/_bulk" \
  -H 'Content-Type: application/x-ndjson' \
  --data-binary @provinces-gis-part-01.ndjson

curl -X POST "localhost:9200/_bulk" \
  -H 'Content-Type: application/x-ndjson' \
  --data-binary @provinces-gis-part-02.ndjson
```

If the total size is under 70 MB, a single `provinces-gis.ndjson` is produced (no chunks).

## 7. Recommendations

### For the Generator (dataset-generation-scripts)

1. **Emit chunked NDJSON**: The generator should split the output into multiple `.ndjson` files (e.g., `provinces-gis-part-01.ndjson`, `provinces-gis-part-02.ndjson`), each under 90 MB. This avoids the `http.max_content_length` issue entirely.

2. **Size warning**: The generator should emit a warning when a single document exceeds 8 MB, flagging which provinces may need special handling during import.

3. **README update**: The `elasticsearch/README.md` should mention:
   - Minimum heap requirement: 1 GB (2 GB recommended for provinces-gis)
   - The chunked import strategy
   - Reference to the import helper script

### For the VPS Deployment

1. **Heap size**: `ES_JAVA_OPTS="-Xms1g -Xmx1g"` is sufficient for reliable provinces-gis imports (tested and verified). The VPS has 12 GB RAM, so 1 GB heap is safe. Use 2 GB if faster import times are desired.

2. **http.max_content_length**: Left at default (100 MB). The NDJSON is split into chunks under 90 MB during import — no ES config change needed. This is the preferred approach to keep ES changes minimal.

### For the Data Model

1. **Nested wards cost**: The `nested` type for `Wards` causes each ward to become a separate Lucene document internally (3,355 internal docs for 34 provinces). This is the primary driver of heap inflation. If spatial queries on wards are rarely needed, consider a separate `wards-gis` index instead.

2. **Geometry precision**: Coordinates are already at 6 decimal places (~11 cm accuracy). Further reduction has negligible impact since the inflation comes from BKD tree indexing, not JSON size.
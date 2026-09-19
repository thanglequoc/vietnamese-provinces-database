# 223 — Store GIS Result for Local Reuse (Local GIS Mock Server)

**Issue:** [#223 Store the GIS result for local reuse](https://github.com/thanglequoc/vietnamese-provinces-database/issues/223)
**Status:** Implemented
**Author:** AI-assisted
**Date:** 2026-09-19

---

## 1. Objective

Every run of the dataset generation script (`go run main.go` with `INCLUDE_GIS = true`)
issues **two live HTTP calls per geo object** to the government GIS server
`https://sapnhap.bando.com.vn`, for all **3,355** objects (~6,710 requests per run).
GIS data changes very rarely, so repeated test runs put an unnecessary load on the
external server.

Goals:

1. **Capture** the raw responses of the external GIS server into a local (gitignored) cache.
2. Provide a **minimal local mock server** that replays those cached responses so the
   dataset generation pipeline can run fully offline.
3. Make it **trivial to switch** between the real GIS server and the local mock server.

The mock server is local-only and **ignores authentication**.

---

## 2. Current Behaviour (baseline)

### 2.1 External endpoints used

Defined in `dataset-generation-scripts/internal/sapnhap_bando/fetcher/fetcher.go`:

| Endpoint | Method | Form field | Response | Consumed by |
|----------|--------|-----------|----------|-------------|
| `https://sapnhap.bando.com.vn/pread_json` | POST | `id=<malk>` | `dto.GISLocationResponse` (GeoJSON `FeatureCollection`, 1 feature, `MultiPolygon`) | `FetchGISDataFromSapNhapBando()` → geometry WKT |
| `https://sapnhap.bando.com.vn/p.co_dvhc_id` | POST | `malk=<malk>` | `[]dto.SapNhapGeoObjectMetadata` | `FillMetaDataForGeoJSONObjects()` → only `dientichkm2` is persisted |

### 2.2 Call sites (exact order in `main.go`)

1. `BootstrapGISDataStructure()` — creates `sapnhap_geojson_objects` and seeds 3,355 rows
   from `resources/gis/sapnhapbando_geo_objects.sql` (no `dientichkm2` values).
2. `sapnhap.BackfillProvinceAndWardCodesInSapNhapGeojsonObjects()`
   → `FillMetaDataForGeoJSONObjects()` → **3,355 × `p.co_dvhc_id`** calls.
3. `sapnhap.FetchGISDataFromSapNhapBando()` → **3,355 × `pread_json`** calls
   (10 concurrent workers, 5 retries).
4. `PatchIslandProvincesGeometry()` / `ValidateAndFixGeometries()` — PostGIS only.
5. `GenerateGISSQLDatasets()`.

### 2.3 Key facts verified during exploration

- `malk` is unique across all 3,355 objects and is filesystem-safe
  (only `[A-Za-z0-9._-]`; two prefixes: `diaphanhanhchinhcaptinh_sn.` ×34,
  `diaphanhanhchinhcapxa_.` ×3321).
- Every existing `pread_json` response has exactly **1 feature**, geometry type
  **`MultiPolygon`** (`GISGeometry.ToWKTCoordinate()` panics on any other type).
- `GetMetadataOfSapNhapGeoObject()` indexes `metadata[0]` **without a length check** —
  an empty array response would panic. The cache must only contain non-empty arrays.
- **An Giang (`ma = 91`)** is special-cased: its geometry is loaded from
  `resources/gis/geojson_11Mar2026/32_tinh_an_giang/province.geojson` and `pread_json`
  is **not** called for it. Its `p.co_dvhc_id` metadata **is** still fetched.
- Gzip compresses the geometry responses **~15×** (measured 1.0 GB raw → ~70–100 MB),
  which makes committing the cache to git practical.
- Existing local GIS data is stale and/or unused:
  - `resources/gis/sapnhapbando_geojson/` (1.0 GB, 3,355 files) — **not referenced by any
    Go code**; only the legacy Python crawler.
  - `resources/gis/geojson_11Mar2026/` (607 MB) — only the An Giang patch file is used.
  - **Per decision, both directories are left untouched** by this work. The new capture
    is the source of truth for the mock; the old data is not reused.

---

## 3. Decisions (confirmed)

| Topic | Decision |
|-------|----------|
| Cache format | Gzip per-`malk` files: `<malk>.json.gz` |
| Cache in git | **Gitignored** (large, ~63 MB / 6,710 files); pulled on demand with `./pull-gis-cache.sh` |
| Existing stale GIS dirs | **Keep everything as-is** (no cleanup) |
| Mock run model | **Standalone command** (`cmd/mockgis`) + `GIS_SERVER_BASE_URL` env var |

---

## 4. Target Cache Layout

```
dataset-generation-scripts/resources/gis/gis_server_cache/
├── manifest.json                                   # capture metadata (not compressed)
├── pread_json/
│   ├── diaphanhanhchinhcaptinh_sn.108.json.gz      # Hà Nội province geometry
│   ├── diaphanhanhchinhcapxa_2025.3264.json.gz     # Phường Cầu Giấy geometry
│   └── ...                                         # 3,355 files
└── co_dvhc_id/
    ├── diaphanhanhchinhcaptinh_sn.108.json.gz      # Hà Nội metadata
    └── ...                                         # 3,355 files
```

- File name = exact `malk` + `.json.gz`. No sanitization needed (verified safe).
- The mock also accepts plain `.json` files to allow hand-authored overrides.
- `manifest.json` shape:

```json
{
  "source_base_url": "https://sapnhap.bando.com.vn",
  "captured_at": "2026-09-19T10:00:00Z",
  "dataset_version": "v5.1.0",
  "latest_decree": "30/2026/QH16",
  "endpoints": {
    "pread_json": { "expected": 3355, "captured": 3355, "missing": [] },
    "co_dvhc_id": { "expected": 3355, "captured": 3355, "missing": [] }
  }
}
```

---

## 5. Affected Components

### 5.1 Modified

| File | Change |
|------|--------|
| `internal/sapnhap_bando/fetcher/fetcher.go` | Replace hardcoded URLs with a configurable base URL read from `GIS_SERVER_BASE_URL` (default real server). Log the active base URL. |
| `dataset-generation-scripts/.env.example` | Add documented `GIS_SERVER_BASE_URL` (real default + commented mock example). |
| `dataset-generation-scripts/README.md` | New "Local GIS mock server" section. |
| `dataset-generation-scripts/CLAUDE.md` | Mention `resources/gis/gis_server_cache/` + mock commands. |
| `AGENTS.md` | Update GIS resource tree + quick-start notes. |

### 5.2 New

| File | Purpose |
|------|---------|
| `internal/sapnhap_bando/capture/capture.go` | Capture logic (fetch raw, gzip-write, manifest, retries). |
| `internal/sapnhap_bando/capture/capture_test.go` | Unit tests (naming, skip/force, gzip round-trip via `httptest`). |
| `cmd/giscapture/main.go` | CLI entry point for capturing the cache. |
| `internal/mock_gis_server/cache.go` | Load cache index (`malk` → file path). |
| `internal/mock_gis_server/server.go` | HTTP handlers for `/pread_json` and `/p.co_dvhc_id`. |
| `internal/mock_gis_server/server_test.go` | Handler tests (both endpoints, 404, 405, gz + plain). |
| `cmd/mockgis/main.go` | CLI entry point for the mock server. |
| `resources/gis/gis_server_cache/**` | The captured cache (gitignored). |

---

## 6. Detailed Design

### 6.1 Fetcher — configurable base URL

In `internal/sapnhap_bando/fetcher/fetcher.go`:

```go
const defaultGISServerBaseURL = "https://sapnhap.bando.com.vn"

func baseURL() string {
    if v := strings.TrimSpace(os.Getenv("GIS_SERVER_BASE_URL")); v != "" {
        return strings.TrimRight(v, "/")
    }
    return defaultGISServerBaseURL
}
```

- `GetGISLocationCoordinates` → `baseURL() + "/pread_json"`.
- `GetMetadataOfSapNhapGeoObject` → `baseURL() + "/p.co_dvhc_id"`.
- Remove the now-unused `GET_METADATA_FROM_MALK_URL` constant.
- Read env **lazily at request time** so `.env` (loaded by `GetPostgresDBConnection` →
  `godotenv.Load`) is honoured even though the fetcher package initializes earlier.
- Log `"Using GIS server base URL: <url>"` once (sync.Once) to make mock/real obvious.

> Behaviour is unchanged when `GIS_SERVER_BASE_URL` is unset.

### 6.2 Capture command (`cmd/giscapture` + `internal/sapnhap_bando/capture`)

**Flags**

| Flag | Default | Meaning |
|------|---------|---------|
| `--out` | `./resources/gis/gis_server_cache` | Output cache directory |
| `--base-url` | `https://sapnhap.bando.com.vn` | Source server (explicit, never inherited from env, to avoid capturing from the mock) |
| `--workers` | `10` | Concurrent workers |
| `--force` | `false` | Overwrite existing files |
| `--source` | `db` | `db` (query `sapnhap_geojson_objects`) or `json` |
| `--json` | `./sapnhap-bando-crawler/donvi_tinhthanh.json` | Fallback malk source |
| `--only` | *(empty)* | Restrict to `pread_json` or `co_dvhc_id` |

**Algorithm**

1. Resolve the malk list:
   - `db`: `repository.GetAllSapNhapGeoJSONObjects()` → `[]malk` (authoritative list the
     pipeline uses). Requires the PostGIS container to be up.
   - `json`: read `[].malk` from the JSON file.
2. For each `malk`, for each selected endpoint:
   - Skip when `<out>/<endpoint>/<malk>.json.gz` exists and `--force` is not set.
   - `POST` with the correct form field (`id` / `malk`), 30s timeout, retry with the same
     backoff pattern as the fetcher (5 retries, 300ms → 5s).
   - Read the **raw body** and validate:
     - `pread_json`: valid JSON object with `features` length ≥ 1.
     - `co_dvhc_id`: valid JSON array with length ≥ 1.
   - Gzip-write the raw body (`gzip.BestCompression`, deterministic header: zero `ModTime`).
3. Write `manifest.json` (source, timestamp, `version.txt` values, counts, missing list).
4. Print a summary; **exit non-zero** if any object is still missing.

**Why raw bytes:** decode + re-encode through the DTOs would lose fields
(`id`, `properties`) and risks changing float precision. The mock must be a faithful
replica, so the exact upstream bytes are stored.

### 6.3 Mock server (`cmd/mockgis` + `internal/mock_gis_server`)

**Cache index** (`cache.go`)

```go
type Cache struct {
    PreadJSON map[string]string // malk -> absolute file path
    CoDvhcID  map[string]string
}
func LoadCache(dir string) (*Cache, error) // scans pread_json/ and co_dvhc_id/
```

- Accepts `.json.gz` and `.json`; key = filename without the extension(s).
- Fails fast with a clear message if a directory is missing/empty.

**Handlers** (`server.go`)

| Route | Form field | Behaviour |
|-------|-----------|-----------|
| `POST /pread_json` | `id` | Look up `PreadJSON[id]`, return decompressed bytes, `Content-Type: application/json` |
| `POST /p.co_dvhc_id` | `malk` | Look up `CoDvhcID[malk]`, return decompressed bytes |
| other paths | — | `404` |
| wrong method | — | `405` |
| unknown key | — | `404` with a JSON error body naming the missing key |

- **No authentication** (local-only).
- **Path traversal impossible**: only keys present in the in-memory index are served; the
  request value is never joined to a path.
- Logs each request as `method path key -> status` (verbose flag controls per-request logs).

**CLI** (`cmd/mockgis/main.go`)

| Flag | Default | Meaning |
|------|---------|---------|
| `--addr` | `127.0.0.1:18080` | Listen address |
| `--data` | `./resources/gis/gis_server_cache` | Cache directory |
| `--verbose` | `false` | Log every request |

- On startup: load cache, print endpoint counts, warn if `manifest.json`'s
  `dataset_version` differs from `version.txt` (stale-cache hint), then serve.

### 6.4 Switching real ↔ mock

Default (real):

```bash
go run main.go
```

Mock:

```bash
# terminal 1
go run ./cmd/mockgis --data ./resources/gis/gis_server_cache --addr 127.0.0.1:18080

# terminal 2
GIS_SERVER_BASE_URL=http://127.0.0.1:18080 go run main.go
```

Or set it persistently in `.env`:

```
GIS_SERVER_BASE_URL=http://127.0.0.1:18080
```

Refreshing the cache (one-time per machine, hits the real server; cache is gitignored):

```bash
./pull-gis-cache.sh                # uses the committed malk list (no DB needed)
./pull-gis-cache.sh --source db    # authoritative list from sapnhap_geojson_objects
```

---

## 7. Tests

| Test | Coverage |
|------|----------|
| `internal/mock_gis_server/server_test.go` | Both endpoints return cached bytes; unknown key → 404; wrong method → 405; `.json` and `.json.gz` both load |
| `internal/sapnhap_bando/capture/capture_test.go` | malk → filename mapping; skip-if-exists vs `--force`; gzip round-trip; validation rejects empty `features` / empty array |
| `internal/sapnhap_bando/fetcher/fetcher_test.go` (new/extended) | `GIS_SERVER_BASE_URL` override is used by both request builders (via `httptest`) |

Existing `go test -v ./...` must continue to pass.

---

## 8. Execution Order

1. Refactor `fetcher.go` for the configurable base URL (default = real; no behaviour change).
2. Implement `internal/sapnhap_bando/capture` + `cmd/giscapture`.
3. **Run the capture against the real server** and commit
   `resources/gis/gis_server_cache/` (verify 3,355 / 3,355, no missing).
4. Implement `internal/mock_gis_server` + `cmd/mockgis` (+ tests).
5. Update `.env.example`, `README.md`, `CLAUDE.md`, `AGENTS.md`.
6. Verify end-to-end against the mock.
7. Run `go build ./...` and `go test -v ./...`.

---

## 9. Verification

1. `go build ./...` and `go test -v ./...` pass.
2. `manifest.json` reports `captured == expected == 3355` for both endpoints, `missing: []`.
3. Start `cmd/mockgis`; `curl` a known `malk` against both routes and diff the output
   against the cached file (byte-identical after decompression).
4. `GIS_SERVER_BASE_URL=http://127.0.0.1:18080 go run main.go` with `INCLUDE_GIS = true`:
   - `Processing complete. Success: 3355, Errors: 0`.
   - GIS outputs generated under `output/`.
5. (Optional strong check) Compare generated GIS SQL checksums from a mock run vs a real
   run; they should match.

---

## 10. Edge Cases & Risks

| Risk / case | Handling |
|-------------|----------|
| Upstream `malk` IDs changed since seed | Capture records 404s / id mismatches in `manifest.json.missing`; exits non-zero for manual follow-up. |
| Rate limiting / transient failures | 10 workers + 5 retries with backoff; capture is resumable (skips existing). |
| Empty metadata array | Capture rejects it (would panic `metadata[0]`); mock returns 404 instead of `[]`. |
| An Giang (`ma = 91`) | No `pread_json` call in the pipeline; capture may still record it harmlessly. Its `co_dvhc_id` **must** be captured. |
| Stale cache after a new decree | `manifest.json` stores `dataset_version` / `latest_decree`; mock warns when it differs from `version.txt`. Re-run capture to refresh. |
| Cache file size | Gzip (~15×) keeps the cache at ~63 MB / 6,710 files; **gitignored** and pulled on demand. |
| `.env` not yet loaded when fetcher runs | Base URL read lazily per request; `godotenv.Load()` already runs during DB bootstrap. |
| Accidentally capturing from the mock | `cmd/giscapture --base-url` defaults to the real server and never reads `GIS_SERVER_BASE_URL`. |

---

## 11. Assumptions

- The government server's response shape stays as-is; the cache stores verbatim bytes, so
  the mock is automatically faithful to whatever was captured.
- The `malk` list in `sapnhap_geojson_objects` (seeded from
  `resources/gis/sapnhapbando_geo_objects.sql`) remains the authoritative key set.
- Local mock usage is developer/CI convenience only; production/release runs still use the
  real server by default.
- Existing `resources/gis/sapnhapbando_geojson/` and `geojson_11Mar2026/` are intentionally
  left in place and are not used as capture input or by the mock.

---

## 12. Out of Scope

- Removing/consolidating the legacy GIS directories.
- Any change to GIS geometry output formats or the island-patch logic.
- Authentication/authorization on the mock server (explicitly not needed).
- Automating periodic cache refresh in CI.

---

## 13. Implementation Outcome (2026-09-19)

Implemented as planned.

| Item | Result |
|------|--------|
| Fetcher base URL | `GISServerBaseURL()` reads `GIS_SERVER_BASE_URL` (default government server). `LogActiveGISServer()` logs `🌐 GIS server base URL: <url> (government server\|custom/mock server)` once per process, called at the start of the GIS phase in `main.go`. `GET_GIS_COORDINATES_URL`/`GET_METADATA_FROM_MALK_URL` constants replaced by `PREAD_JSON_PATH`/`CO_DVHC_ID_PATH` + base. |
| Capture | `internal/sapnhap_bando/capture` + `cmd/giscapture`. Raw bodies stored verbatim, gzip (`BestCompression`, zero `ModTime`), atomic temp-file writes, resumable, manifest. |
| Fresh cache | Captured 2026-09-19 from `https://sapnhap.bando.com.vn`: **3,355/3,355** for both endpoints, 0 missing. Total size **63 MB** on disk (6,710 files). |
| Distribution | Cache is **gitignored** (`.gitignore`); `dataset-generation-scripts/pull-gis-cache.sh` performs the one-time pull (`--source json` by default, `--source db` optional). |
| Data freshness check | Hà Nội area changed `3.359,80` (stale `sapnhapbando_geojson/ti1.geojson`) → `3.359,84` (fresh cache), confirming a fresh dump was warranted. |
| Mock server | `internal/mock_gis_server` + `cmd/mockgis`; serves `/pread_json` (form `id`) and `/p.co_dvhc_id` (form `malk`); 404 unknown, 405 wrong method, no auth, no path traversal. |
| Tests | New tests for fetcher URL resolution, capture (naming/validation/gzip/run/resume/missing), and mock handlers. `go build`, `go vet`, and `go test ./...` all pass. |
| End-to-end | `GIS_SERVER_BASE_URL=http://127.0.0.1:18080 go run main.go` → `Processing complete. Success: 3355, Errors: 0`; DB `sapnhap_geojson_objects` has 3,355 geom/bbox/area; GIS outputs generated. |
| Docs | Updated `.env.example`, `dataset-generation-scripts/README.md`, `CLAUDE.md`, `AGENTS.md`. |

The stale `resources/gis/sapnhapbando_geojson/` and `resources/gis/geojson_11Mar2026/`
directories were intentionally left in place (per decision) and are not used by the
mock or capture.


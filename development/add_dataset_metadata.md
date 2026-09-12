# Add dataset metadata (version / latest decree / generated_at)

## Objective

Let consumers of every exported dataset discover which release of the Vietnamese
Provinces Database they have installed, which government decree it reflects, and when it
was generated — so they can detect and apply updates correctly.

## Source of truth

A new `dataset-generation-scripts/version.txt` holds the two maintainer-controlled values:

```
dataset_version=v5.1.0
latest_decree=30/2026/QH16
```

The generation timestamp is captured at run time (UTC) and is **not** stored in the file.

## Metadata shape

A **separate, single-row** `dataset_metadata` entity. No existing table/collection/index
structure is modified.

| Column | Type | Notes |
|--------|------|-------|
| `dataset_version` | varchar(50) | e.g. `v5.1.0` |
| `latest_decree` | varchar(100) | nullable, e.g. `30/2026/QH16` |
| `generated_at` | timestamp | UTC generation time |

## Affected components

| Component | Change |
|-----------|--------|
| `version.txt` | New version source file |
| `internal/dataset_metadata/` | New loader package (`Parse`, `LoadFromFile`, `Load`) + tests |
| `resources/db_table_init.sql` | Add `dataset_metadata` DDL (tmp Postgres) |
| `resources/fresh_cleanup.sql` | Drop `dataset_metadata` |
| `internal/database/vn_province_db_service.go` | `BootstrapDatasetMetadata()` inserts the row |
| `main.go` | Call `BootstrapDatasetMetadata()` after structure bootstrap |
| `postgresql/`, `mysql/`, `sqlserver/`, `oracle/` `*_CreateTables_vn_units.sql` | Add engine-native `dataset_metadata` DDL |
| `postgres_mysql_dataset_file_writer.go` | `Metadata` field → `INSERT` in base import file |
| `mssql_dataset_file_writer.go` | `Metadata` field → `INSERT` (N'' literals) |
| `oracle_dataset_file_writer.go` | `Metadata` field → `INSERT` (`TO_TIMESTAMP`) |
| `json_file_writer.go` | `Metadata` field → `metadata.json` |
| `mongodb_file_writer.go` | `Metadata` field → `mongo_data_vn_metadata.json` |
| `redis_file_writer.go` | `Metadata` field → `HSET datasetMetadata ...` |
| `elasticsearch_file_writer.go` | `Metadata` field → `dataset_metadata.ndjson` + mapping (index `dataset_metadata`) |
| `dataset_writer.go` | Load metadata once, pass to every writer |
| `.opencode/skills/vn-provinces-patch/` | Ignore `dataset_metadata` when diffing |
| `README.md`, `README_vi.md`, `AGENTS.md`, `CLAUDE.md` | Documentation |

## Implementation notes

- The loader captures `GeneratedAt = time.Now().UTC()` at parse time so all formats share
  one timestamp per run.
- Writers expose a `Metadata` struct field rather than new method parameters, keeping
  existing `WriteToFile` signatures and tests unchanged. When the field is the zero value,
  no metadata artifact is emitted (existing tests remain valid).
- Unused legacy constants `esDatasetVer` / `esAdminRev` / `mongoDatasetVer` /
  `mongoAdminRev` were removed and superseded by `version.txt`.
- The metadata section is written before the `-- END OF SCRIPT FILE --` marker in the SQL
  import files.

## Edge cases

- Missing/invalid `version.txt` aborts generation with a clear error.
- Empty `latest_decree` is written as `NULL` (SQL) / empty string (JSON/Redis/ES).
- Timestamps are always UTC (`YYYY-MM-DD HH:MM:SS` for SQL, RFC 3339 elsewhere).
- `dataset_metadata` changes on every regeneration (timestamp), so the patch skill must
  ignore it; otherwise every patch would be non-empty.

## Upstream GIS version decision

The GIS source `sapnhap.bando.com.vn` exposes **no dataset version number** (verified:
homepage footer, `p.co_dvhc_id`, `pread_json`, WMS `GetCapabilities`, tile layer names).
It only publishes static publication identifiers (ĐKXB/ISBN), year-tagged tile layers
(`vietnam_2026`), and a per-object legal basis (`cancu`). Therefore no separate GIS version
field was added; `generated_at` is the GIS as-of timestamp and `latest_decree` is the legal
basis.

## Assumptions

- The dataset is regenerated (and published) whenever a new decree or dataset version is
  prepared.
- Consumers run the provided `*_CreateTables_*` script before importing, so the
  `dataset_metadata` table exists.
- `dataset_metadata` is intentionally excluded from data-only upgrade patches.

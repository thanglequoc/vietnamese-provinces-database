# Design: Enrich Generated Dataset READMEs

**Date:** 2026-08-14
**Branch:** `134_AddPostalCode`
**Status:** Approved

## Objective

Replace the terse, repetitive generated dataset READMEs (from the previous
"per-dataset READMEs" feature) with genuinely useful documentation for every
dataset type — mirroring and extending the quality of the removed `mongodb/gis`
and original Elasticsearch READMEs.

## Problem

The current generated READMEs are too thin:

- **MongoDB/JSON**: `## Data Structure` simply re-lists the `## Files` entries
  with a one-line description; no field-level docs, no sample document, no
  import instructions.
- **PostgreSQL/MySQL/MSSQL/Oracle**: a one-line Data Structure, no table schema,
  no "create tables first" import steps, only 2-3 queries.
- **Redis**: no load command, no sample data lines.
- **Elasticsearch**: better, but condensed relative to the original (lost the
  JSON document preview and the fuller query set).

## Requirements (confirmed with user)

1. Approach A: static rich content embedded as Go string slices in each writer
   (no template files, no data access). The shared `writeDatasetReadme` skeleton
   is unchanged.
2. Every dataset README follows six sections: `## Overview`, `## Data Structure`,
   `## Sample Document`, `## Quick Start`, `## Sample Queries`,
   `## GIS / GeoJSON`.
3. Restore-and-extend: the removed `mongodb/gis` README and the original
   Elasticsearch README content are restored into the new main READMEs and
   extended with the base (non-GIS) dataset coverage.

## Changes

### 1. Shared skeleton (unchanged)

`writeDatasetReadme` continues to render `# <title>`, bold `**Generated at:**`,
intro, `## Files` (with sizes), then the dataset's `sections` verbatim.

### 2. Section content per dataset

| Section | Content |
|---------|---------|
| `## Overview` | What the dataset is + markdown table of tables/collections/keys with counts (8 regions, 8 units, 34 provinces, 3,321 wards; GIS tables/collections where present) |
| `## Data Structure` | Field/column-level documentation (SQL: per-table column lists incl. GIS tables; MongoDB/ES/JSON: document fields — Code, Name, NameEn, FullName, FullNameEn, CodeName, AdministrativeUnit, SearchKeywords, Wards, Meta, GIS; Redis: key patterns + fields) |
| `## Sample Document` | Concrete example (MongoDB/ES/JSON: full JSON document preview with 1-2 wards, restored from old content; SQL: sample `INSERT`/row per table; Redis: sample `HSET`/`SADD` lines) |
| `## Quick Start` | Actual load commands (SQL: `*_CreateTables_vn_units.sql` → import data → `gis/` parts in manifest order; MongoDB: `mongoimport` + `mongosh create_indexes.js`; Redis: `redis-cli --pipe`; ES: `curl -X PUT` mappings → `curl -X POST _bulk` NDJSON incl. GIS parts; JSON: `require`/`json.load`) |
| `## Sample Queries` | 4-8 queries per dataset, regular + GIS |
| `## GIS / GeoJSON` | The `gis/`/`geojson/` subfolder files + import steps |

### 3. Writers to update

| Writer | Function to rewrite | Notes |
|--------|--------------------|-------|
| `postgres_mysql_dataset_file_writer.go` | `postgresMySQLReadmeSections` (postgres + mysql branches) | Table schema for 4 base tables + GIS tables; sample rows; `postgres_CreateTables_vn_units.sql` / `mysql_CreateTables_vn_units.sql` quick start |
| `mssql_dataset_file_writer.go` | `writeMssqlReadme` | MSSQL syntax, `mssql_CreateTables_vn_units.sql` |
| `oracle_dataset_file_writer.go` | `writeOracleReadme` | Oracle syntax, `oracle_CreateTables_vn_units.sql`, no GIS |
| `mongodb_file_writer.go` | `writeMongoReadme` | Restore old gis README content + base collections |
| `redis_file_writer.go` | `writeRedisReadme` | `redis-cli --pipe` quick start, sample HSET lines |
| `elasticsearch_file_writer.go` | `writeESReadme` | Restore original rich body + shared Files section |
| `json_file_writer.go` | `writeJSONDatasetReadme` | Field docs, sample document, Node/Python usage |

### 4. Content source of truth

All content is static Go string slices; counts are literal values. No database or
data access at README-generation time.

### 5. Tests

Update the per-writer README tests to assert the richer sections
(`## Overview`, `## Data Structure`, `## Sample Document`, `## Quick Start`,
`## Sample Queries`, `## GIS / GeoJSON`) plus per-dataset content markers:

- MongoDB: `$geoIntersects`, `mongoimport`, `create_indexes.js`
- Elasticsearch: `_bulk`, nested `Wards` search
- PostgreSQL/MySQL: `CreateTables_vn_units.sql`
- MSSQL: `STContains`
- Oracle: `INSERT ALL`
- Redis: `--pipe`
- JSON: `require(`
- GIS subfolder section: `## GIS / GeoJSON`

## Edge Cases

- **Content stability**: the README bodies are static; only timestamp and file
  sizes are dynamic. Tests assert section headers and markers, not full content.
- **SQL create-tables scripts** exist only in the published folders, not in
  `output/`; the README Quick Start references them by name (they are copied
  alongside in the published folder).
- **Counts drift**: 34/3,321/8 are documented as literals; they change only when
  a new government decree is applied (out of scope here).

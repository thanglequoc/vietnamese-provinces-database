# Design: ElasticSearch Query Skill in AGENTS.md

**Date:** 2026-07-18
**Status:** Draft
**Approach:** A — Fully Inline in AGENTS.md (follows existing Postgres db-query pattern)

---

## Objective

Add a new `## ElasticSearch Query Skill` section to the root `AGENTS.md` that gives any AI agent (Cline, Cursor, Windsurf, etc.) automatic Elasticsearch query capabilities — auto-trigger rules, SSH tunnel management, schema reference, and query patterns — targeting an ElasticSearch instance on OVHCloud VPS accessible via bastion host.

## Background

The project has a successfully proven pattern for database skills: the existing `## Database Query Skill` section in `AGENTS.md` auto-triggers based on keywords, connects to a local Docker Postgres container via `docker exec`, and includes schema reference and query patterns. This design replicates that pattern for Elasticsearch.

The user's ElasticSearch runs in Docker on an OVHCloud VPS (`machine.thanglequoc.xyz`). Access requires an SSH tunnel forwarding port 9200. The ES instance has no authentication (open on localhost once the tunnel is established).

## Connection Architecture

```
AI Agent (Cline)                    Bastion (OVHCloud VPS)
┌──────────────┐      SSH tunnel      ┌──────────────────┐
│              │  ──port 9200────▶   │  Docker           │
│  curl        │                      │  ┌─────────────┐ │
│  localhost:  │                      │  │ ElasticSearch│ │
│  9200        │                      │  │ :9200        │ │
│              │                      │  └─────────────┘ │
└──────────────┘                      └──────────────────┘
```

- **SSH command**: `ssh -f -N -L 9200:localhost:9200 -i ~/.ssh/id_ed25519 thanglequoc@machine.thanglequoc.xyz`
- `-f` backgrounds the tunnel so the agent can immediately proceed with queries
- `-N` means no remote command is executed
- Key path: `/Users/thanglequoc/.ssh/id_ed25519`
- No Elasticsearch authentication required

## Design

### 1. New AGENTS.md Section: "ElasticSearch Query Skill"

Add a comprehensive section to `AGENTS.md` after the existing "Database Query Skill" section. The new section contains:

#### Auto-Trigger Rules

An explicit list of trigger keywords signaling the agent to proactively query Elasticsearch:

| Category | Trigger Keywords |
|----------|-----------------|
| Search/find | "search", "find", "full-text", "autocomplete", "lookup in ES", "query ES" |
| Count/total | "how many in elastic", "es count", "count documents" |
| Management | "create index", "delete index", "put mapping", "bulk import", "cluster health", "index exists" |
| Data topics | "provinces index", "provinces-gis", "elasticsearch", "elastic search", "ndjson" |
| GIS queries | "geo shape", "geo point", "spatial query", "find province by point", "find by coordinates" |
| Direct requests | "query elasticsearch", "run ES", "curl ES", "search ES" |

Critical instruction: **Do NOT wait for explicit skill invocation — proactively query when context suggests Elasticsearch interaction is needed.**

#### Connection & Tunnel Management

The skill must auto-manage the SSH tunnel. The AGENTS.md provides:

1. **Tunnel check command**:
   ```bash
   curl -s -o /dev/null -w "%{http_code}" localhost:9200
   ```
   Returns `200` if reachable, otherwise assume tunnel is down.

2. **Tunnel start command**:
   ```bash
   ssh -f -N -L 9200:localhost:9200 -i ~/.ssh/id_ed25519 thanglequoc@machine.thanglequoc.xyz
   ```

3. **Base URL**: `http://localhost:9200` — no authentication headers needed.

4. **JSON formatting**: Pipe through `jq` for readability; fall back to `python3 -m json.tool` if `jq` unavailable.

#### Schema Reference

**Indices:**

| Index | Documents | Description |
|-------|-----------|-------------|
| `provinces` | 34 | Province documents with embedded wards, no GIS |
| `provinces-gis` | 34 | Same structure + GIS geometry (Center, BoundingBox, Geometry) |

**Key Fields (both indices):**

| Field | Type | Description |
|-------|------|-------------|
| `Code` | keyword | Province code (e.g., "01") |
| `Name` | text + keyword | Vietnamese name |
| `NameEn` | text + keyword | English name |
| `FullName` | text | Full Vietnamese name with prefix |
| `FullNameEn` | text | Full English name with prefix |
| `CodeName` | keyword | Snake-case code name (e.g., "ha_noi") |
| `AdministrativeUnit` | object (flattened) | Embedded unit: Id, FullName, FullNameEn, ShortName, ShortNameEn, CodeName, CodeNameEn |
| `SearchKeywords` | keyword[] | Pre-computed autocomplete array |
| `Wards` | nested | Array of ward documents (Code, Name, NameEn, FullName, FullNameEn, CodeName, AdministrativeUnit, SearchKeywords; + GIS in provinces-gis) |
| `Meta` | object | DatasetVersion, AdministrativeRevision, GeneratedAt |
| `GIS` | object | **(provinces-gis only)** Center (geo_point), BoundingBox (object with Min/Max lat/lon), Geometry (geo_shape, GeoJSON MultiPolygon) |

**Mapping files**: `elasticsearch/mappings/provinces.json` and `elasticsearch/mappings/provinces-gis.json`

#### Common Query Patterns

**Search / dataset queries:**

Count documents:
```bash
curl -s "localhost:9200/provinces/_count" | jq .
curl -s "localhost:9200/provinces-gis/_count" | jq .
```

Search by keyword (autocomplete):
```bash
curl -s "localhost:9200/provinces/_search" \
  -H 'Content-Type: application/json' \
  -d '{"query": {"terms": {"SearchKeywords": ["ha noi"]}}, "_source": ["Code", "Name", "NameEn"]}' | jq .
```

Nested ward search (find wards matching a name):
```bash
curl -s "localhost:9200/provinces/_search" \
  -H 'Content-Type: application/json' \
  -d '{"query": {"nested": {"path": "Wards", "query": {"match": {"Wards.FullName": "Ba Đình"}}, "inner_hits": {}}}, "_source": ["Code", "Name"]}' | jq .
```

Province dropdown (all provinces sorted):
```bash
curl -s "localhost:9200/provinces/_search" \
  -H 'Content-Type: application/json' \
  -d '{"size": 34, "sort": [{"Code": "asc"}], "_source": ["Code", "Name", "NameEn"]}' | jq .
```

GIS spatial query (find province containing a point):
```bash
curl -s "localhost:9200/provinces-gis/_search" \
  -H 'Content-Type: application/json' \
  -d '{"query": {"geo_shape": {"GIS.Geometry": {"shape": {"type": "point", "coordinates": [105.8542, 21.0285]}, "relation": "intersects"}}}, "_source": ["Code", "Name"]}' | jq .
```

**Management:**

Cluster health:
```bash
curl -s "localhost:9200/_cluster/health" | jq .
```

List all indices:
```bash
curl -s "localhost:9200/_cat/indices?v"
```

Create index with mapping:
```bash
curl -X PUT "localhost:9200/provinces" \
  -H 'Content-Type: application/json' \
  -d @elasticsearch/mappings/provinces.json
```

Bulk import NDJSON:
```bash
curl -X POST "localhost:9200/_bulk" \
  -H 'Content-Type: application/x-ndjson' \
  --data-binary @elasticsearch/provinces.ndjson
```

Delete index:
```bash
curl -X DELETE "localhost:9200/provinces"
```

Get index mapping:
```bash
curl -s "localhost:9200/provinces/_mapping" | jq .
```

#### Agent Instructions

When context triggers an Elasticsearch query:
1. Check tunnel reachability (`curl -s -o /dev/null -w "%{http_code}" localhost:9200`); start tunnel if returning non-200
2. Execute the curl command with `-s` (silent mode) and pipe through `jq` for readable output
3. Format results in tables, lists, or summaries as appropriate
4. Use `_count` before `_search` to confirm expected document counts
5. For large result sets, add `"size": 5` (or appropriate limit) to avoid overwhelming output
6. Fall back to `python3 -m json.tool` if `jq` is unavailable on the system
7. If Elasticsearch returns an error (non-2xx status), explain the issue and suggest a fix
8. When the SSH tunnel fails to connect, report the error clearly and suggest checking the bastion host

### 2. Scope Boundaries

**In scope:**
- Auto-trigger rules and connection management in AGENTS.md
- Dataset query patterns for `provinces` and `provinces-gis` indices
- General ES management (cluster health, index CRUD, bulk import)
- GIS spatial queries through the `provinces-gis` index

**Out of scope:**
- No changes to existing PostgreSQL db-query skill
- No changes to Go code or the generation pipeline
- No Elasticsearch client library installation
- No MCP server or new runtime dependencies
- No changes to CI/CD pipeline
- No Elasticsearch container management (start/stop Docker on VPS)

### 3. AGENTS.md Placement

The new section is inserted **after** the existing `## Database Query Skill` section and **before** the existing `## Quick Start for AI Agents` section. This keeps all data-access skills grouped together.

## Files Affected

| File | Action | Details |
|------|--------|---------|
| `AGENTS.md` | Modify | Add "ElasticSearch Query Skill" section (~120-150 lines) after the Database Query Skill section |

Single file change — consistent with the approach taken for the Postgres db-query skill consolidation.

## Risks & Mitigations

| Risk | Severity | Mitigation |
|------|----------|------------|
| SSH tunnel not established when agent tries to query | Medium | AGENTS.md instructs agent to check reachability first and start tunnel if needed |
| Agent fails to background SSH tunnel correctly | Low | `ssh -f -N` is standard and reliable; fallback instruction to run tunnel in separate terminal |
| AGENTS.md grows too large | Low | New section is ~120-150 lines; AGENTS.md currently ~200 lines; total ~350 lines remains manageable |
| Elasticsearch unavailable (VPS down, Docker stopped) | Low | Agent instructions include error handling — report failure clearly and suggest troubleshooting |
| Credentials/connection details change | Low | SSH details are in a single location, easy to update |
| Overlap with Postgres skill triggers | Low | ES-specific triggers use "elasticsearch", "es", "ndjson" keywords — distinct from Postgres triggers |

## Out of Scope

- No database schema or data changes
- No Go code changes
- No Elasticsearch client library dependency
- No MCP server creation
- No CI/CD pipeline changes
- No changes to `elasticsearch/` output directory or generated files
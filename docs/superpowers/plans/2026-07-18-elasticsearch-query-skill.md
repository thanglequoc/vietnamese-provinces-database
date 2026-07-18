# ElasticSearch Query Skill in AGENTS.md — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a new `## ElasticSearch Query Skill` section to the root `AGENTS.md` that gives any AI agent automatic Elasticsearch query capabilities — auto-trigger rules, SSH tunnel management, schema reference, and curl-based query patterns.

**Architecture:** A single Markdown section added to the root `AGENTS.md` file, inserted after the existing `## Database Query Skill` section (after line 163) and before `## Project Structure` (line 166). Follows the identical structure of the existing Postgres db-query skill: auto-trigger keywords table → connection details → schema/table reference → common query patterns → agent instructions. All queries use `curl` via an SSH tunnel.

**Tech Stack:** Markdown documentation. No code changes. Queries use `curl` + `jq` (or `python3 -m json.tool` fallback). SSH tunnel uses native `ssh` client.

**Spec:** `docs/superpowers/specs/2026-07-18-elasticsearch-query-skill-design.md`

## Global Constraints

- No database schema or data changes
- No Go code changes
- No Elasticsearch client library dependency
- No MCP server or new runtime dependencies
- No CI/CD pipeline changes
- No changes to existing PostgreSQL db-query skill section
- No Elasticsearch container management (start/stop Docker on VPS)
- No changes to `elasticsearch/` output directory or generated files
- Use exact SSH details: host `machine.thanglequoc.xyz`, port `22`, user `thanglequoc`, key `~/.ssh/id_ed25519`
- No Elasticsearch authentication required

---

### Task 1: Insert ElasticSearch Query Skill section into AGENTS.md

**Files:**
- Modify: `AGENTS.md` (insert between line 163 and line 166)

**Interfaces:**
- Consumes: (none — standalone documentation change)
- Produces: A new `## ElasticSearch Query Skill` section in AGENTS.md, containing auto-trigger rules, SSH tunnel management commands, schema reference for `provinces`/`provinces-gis` indices, curl-based query patterns, and agent instructions.

The section is inserted between:
- Line 163: `---` (end of Database Query Skill section)
- Line 166: `## Project Structure`

It must follow the same heading hierarchy and visual style as the existing Database Query Skill section.

- [ ] **Step 1: Add the new section to AGENTS.md**

Insert the following content after line 163 (`---`) and before line 166 (`## Project Structure`). The content is the complete ElasticSearch Query Skill section:

```markdown
## ElasticSearch Query Skill

**CRITICAL: Proactively query Elasticsearch whenever the task involves searching provinces/wards data, cluster management, bulk imports, or GIS spatial queries. Do NOT wait for explicit command invocation — the triggers below signal when to auto-query.**

### Auto-Trigger Rules

Execute Elasticsearch queries automatically when user asks about any of these:

| Category | Trigger Keywords |
|----------|-----------------|
| Search/find | "search", "find", "full-text", "autocomplete", "lookup in ES", "query ES" |
| Count/total | "how many in elastic", "es count", "count documents" |
| Management | "create index", "delete index", "put mapping", "bulk import", "cluster health", "index exists" |
| Data topics | "provinces index", "provinces-gis", "elasticsearch", "elastic search", "ndjson" |
| GIS queries | "geo shape", "geo point", "spatial query", "find province by point", "find by coordinates" |
| Direct requests | "query elasticsearch", "run ES", "curl ES", "search ES" |

**Examples:**
- "Search for provinces matching 'ha noi' in Elasticsearch" → Run search query immediately
- "How many documents are in the provinces index?" → Run `_count` query
- "Check ES cluster health" → Run `_cluster/health`
- "Bulk import the provinces NDJSON" → Run `_bulk` with the file
- "Find which province contains coordinates 105.85, 21.03" → Run geo_shape query on provinces-gis
- "Show me the mapping for the provinces index" → Run `_mapping`

### Connection & Tunnel Management

Elasticsearch runs in Docker on the OVHCloud VPS and is accessible via SSH tunnel:

```bash
# Check if Elasticsearch is reachable through the tunnel
curl -s -o /dev/null -w "%{http_code}" localhost:9200

# Start the SSH tunnel if unreachable (returns non-200)
ssh -f -N -L 9200:localhost:9200 -i ~/.ssh/id_ed25519 thanglequoc@machine.thanglequoc.xyz
```

| Detail | Value |
|--------|-------|
| Base URL | `http://localhost:9200` |
| Authentication | None (open on localhost via tunnel) |
| SSH Host | `machine.thanglequoc.xyz` |
| SSH Port | `22` |
| SSH User | `thanglequoc` |
| SSH Key | `~/.ssh/id_ed25519` |
| JSON Formatter | `jq` (fallback: `python3 -m json.tool`) |

### Schema Reference

**Indices:**

| Index | Documents | Description |
|-------|-----------|-------------|
| `provinces` | 34 | Province documents with embedded wards, administrative units, and search keywords (no GIS geometry) |
| `provinces-gis` | 34 | Same structure as provinces + GIS fields: Center (geo_point), BoundingBox, Geometry (geo_shape) |

**Key Fields (both indices):**

| Field | Type | Description |
|-------|------|-------------|
| `Code` | keyword | Province code (e.g., "01", "02") |
| `Name` | text + keyword | Vietnamese name |
| `NameEn` | text + keyword | English name |
| `FullName` | text | Full Vietnamese name with administrative prefix |
| `FullNameEn` | text | Full English name with administrative prefix |
| `CodeName` | keyword | Snake-case code name (e.g., "ha_noi", "ho_chi_minh") |
| `AdministrativeUnit` | object | Embedded unit: Id, FullName, FullNameEn, ShortName, ShortNameEn, CodeName, CodeNameEn |
| `SearchKeywords` | keyword[] | Pre-computed autocomplete array (code, tone-stripped Vietnamese name, English name, codeName) |
| `Wards` | nested | Array of ward documents (same fields as province: Code, Name, NameEn, FullName, FullNameEn, CodeName, AdministrativeUnit, SearchKeywords) |
| `Meta` | object | DatasetVersion, AdministrativeRevision, GeneratedAt |
| `GIS` | object | **(provinces-gis only)** Center (geo_point: lat/lon), BoundingBox (MinLongitude, MinLatitude, MaxLongitude, MaxLatitude), Geometry (geo_shape, GeoJSON MultiPolygon) |

**Ward fields** (inside `Wards[]`): `Code`, `Name`, `NameEn`, `FullName`, `FullNameEn`, `CodeName`, `AdministrativeUnit`, `SearchKeywords`; plus `GIS` in provinces-gis.

**Mapping files**: `elasticsearch/mappings/provinces.json` and `elasticsearch/mappings/provinces-gis.json`

### Common Query Patterns

**Dataset queries:**

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

Province dropdown (all provinces sorted by code):
```bash
curl -s "localhost:9200/provinces/_search" \
  -H 'Content-Type: application/json' \
  -d '{"size": 34, "sort": [{"Code": "asc"}], "_source": ["Code", "Name", "NameEn"]}' | jq .
```

Nested ward search (find which province contains a specific ward):
```bash
curl -s "localhost:9200/provinces/_search" \
  -H 'Content-Type: application/json' \
  -d '{"query": {"nested": {"path": "Wards", "query": {"match": {"Wards.FullName": "Ba Đình"}}, "inner_hits": {}}}, "_source": ["Code", "Name"]}' | jq .
```

Full-text search on province names:
```bash
curl -s "localhost:9200/provinces/_search" \
  -H 'Content-Type: application/json' \
  -d '{"query": {"multi_match": {"query": "hanoi", "fields": ["Name", "NameEn", "FullName", "FullNameEn"]}}, "_source": ["Code", "Name", "NameEn"]}' | jq .
```

Get a single province by code:
```bash
curl -s "localhost:9200/provinces/_doc/01" | jq .
```

**GIS spatial queries (provinces-gis index):**

Find province containing a point:
```bash
curl -s "localhost:9200/provinces-gis/_search" \
  -H 'Content-Type: application/json' \
  -d '{"query": {"geo_shape": {"GIS.Geometry": {"shape": {"type": "point", "coordinates": [105.8542, 21.0285]}, "relation": "intersects"}}}, "_source": ["Code", "Name"]}' | jq .
```

Find provinces within a bounding box:
```bash
curl -s "localhost:9200/provinces-gis/_search" \
  -H 'Content-Type: application/json' \
  -d '{"query": {"geo_bounding_box": {"GIS.Center": {"top_left": {"lat": 21.5, "lon": 105.5}, "bottom_right": {"lat": 20.5, "lon": 106.5}}}}, "_source": ["Code", "Name"]}' | jq .
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

Check if an index exists:
```bash
curl -s -o /dev/null -w "%{http_code}" "localhost:9200/provinces"
```

Create index from mapping file:
```bash
curl -X PUT "localhost:9200/provinces" \
  -H 'Content-Type: application/json' \
  -d @elasticsearch/mappings/provinces.json
```

Bulk import NDJSON data:
```bash
curl -X POST "localhost:9200/_bulk" \
  -H 'Content-Type: application/x-ndjson' \
  --data-binary @elasticsearch/provinces.ndjson
```

Delete an index:
```bash
curl -X DELETE "localhost:9200/provinces"
```

Get index mapping:
```bash
curl -s "localhost:9200/provinces/_mapping" | jq .
```

Refresh an index (make recent writes visible):
```bash
curl -X POST "localhost:9200/provinces/_refresh"
```

### Agent Instructions

When context triggers an Elasticsearch query:
1. Check tunnel reachability first: `curl -s -o /dev/null -w "%{http_code}" localhost:9200`
2. If return code is not 200, start the tunnel: `ssh -f -N -L 9200:localhost:9200 -i ~/.ssh/id_ed25519 thanglequoc@machine.thanglequoc.xyz`
3. Execute the curl command with `-s` (silent mode) and pipe through `jq` for readable JSON output
4. Format results in tables, lists, or key-value summaries — not raw JSON dumps
5. Use `_count` before `_search` when checking document totals to avoid returning full documents
6. For large result sets, add `"size": 5` (or appropriate small number) to avoid overwhelming output
7. Use `_source` filtering in search requests to return only needed fields
8. Fall back to `python3 -m json.tool` if `jq` is not installed on the system
9. If Elasticsearch returns an error (non-2xx status code), explain the issue clearly and suggest a fix
10. If the SSH tunnel fails to connect, report the error and suggest: verifying the bastion is online, checking SSH key permissions (`chmod 600 ~/.ssh/id_ed25519`), or running `ssh -v thanglequoc@machine.thanglequoc.xyz` to debug
```

- [ ] **Step 2: Verify the section was inserted correctly**

Run:
```bash
grep -n "## ElasticSearch Query Skill" AGENTS.md
```
Expected: Line number output showing the new section heading exists.

Run:
```bash
grep -n "## Project Structure" AGENTS.md
```
Expected: Line number greater than the ElasticSearch section heading, confirming it appears before Project Structure.

Run:
```bash
grep -c "^## " AGENTS.md
```
Expected: Should show one more `## ` heading than before (the new section).

- [ ] **Step 3: Verify key content is present**

Run:
```bash
grep -c "ssh -f -N -L 9200:localhost:9200" AGENTS.md
```
Expected: `1`

Run:
```bash
grep -c "machine.thanglequoc.xyz" AGENTS.md
```
Expected: At least `1`

Run:
```bash
grep -c "_cluster/health" AGENTS.md
```
Expected: At least `1`

Run:
```bash
grep -c "_bulk" AGENTS.md
```
Expected: At least `1`

Run:
```bash
grep -c "geo_shape" AGENTS.md
```
Expected: At least `1`

- [ ] **Step 4: Commit**

```bash
git add AGENTS.md
git commit -m "docs: add ElasticSearch query skill section to AGENTS.md"
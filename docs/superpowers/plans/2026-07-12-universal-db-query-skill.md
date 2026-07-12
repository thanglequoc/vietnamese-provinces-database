# Universal db-query Skill in AGENTS.md — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Consolidate the project's database query skill from a Claude Code-specific `.claude/skills/db-query.md` into a universal section within `AGENTS.md`, usable by any AI agent.

**Architecture:** Add a comprehensive "Database Query Skill" section to `AGENTS.md` containing auto-trigger rules, connection details, schema/table reference, and query patterns. Delete the old `.claude/skills/db-query.md` file. Trim duplicate content from `CLAUDE.md` with a pointer to AGENTS.md. Update the memory feedback file to reflect the new location.

**Tech Stack:** No code changes — documentation reorganization only.

## Global Constraints

- No database schema or data changes
- No Go code changes
- No changes to `settings.local.json`
- No changes to CI/CD pipeline
- All content must remain accurate against the current database schema (tables, columns, relationships)

---

### Task 1: Add "Database Query Skill" section to AGENTS.md

**Files:**
- Modify: `AGENTS.md` (replace lines 28-42)

**Interfaces:**
- Produces: A new `## Database Query Skill` section in AGENTS.md that all agents can read for auto-trigger rules, connection details, schema, and query patterns.
- This section replaces the existing `### Database Connection (Docker)` subsection (lines 28-42) with the expanded content.

- [ ] **Step 1: Replace the "Database Connection (Docker)" subsection in AGENTS.md with the full Database Query Skill section**

Replace lines 28-42 of `AGENTS.md`:

**Content to insert (replacing existing lines 28-42):**

```markdown
## Database Query Skill

**CRITICAL: Proactively query the database whenever the task involves data verification, counts, searching, or GIS data. Do NOT wait for explicit command invocation — the triggers below signal when to auto-query.**

### Auto-Trigger Rules

Execute database queries automatically when user asks about any of these:

| Category | Trigger Keywords |
|----------|-----------------|
| Count/total | "how many", "count", "total", "number of", "how much" |
| Search/find | "find", "search", "show", "list", "get", "lookup" |
| Verification | "check", "verify", "validate", "missing", "orphaned" |
| Data topics | "database", "table", "data", "records", "provinces", "wards", "schema" |
| GIS | "geometry", "bbox", "geom", "gis", "spatial", "geojson", "coordinates" |
| Direct requests | "query from", "read from database", "get from [table]", "run a query" |

**Examples:**
- "How many wards are in Hà Nội?" → Run query immediately
- "Check if there are any missing GIS data" → Run verification query
- "Show me provinces without codes" → Query and display results
- "List all tables" → Run `\dt`
- "Get data from sapnhap_geojson_objects" → Query the table

### Connection Details

The temporary database runs in Docker:

```bash
# Container: vn_provinces_postgres_container
# Database: vn_provinces_tmp
# Username: postgres
# Host:     localhost
# Port:     15432

# Single query
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "YOUR_SQL_QUERY_HERE"

# Multi-line query (heredoc)
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp <<'EOF'
YOUR_MULTI_LINE_SQL_QUERY_HERE
EOF
```

### Table Reference

| Table | Records | Description |
|-------|---------|-------------|
| `provinces_tmp` | 34 | Vietnam provinces (code, name, name_en, full_name, administrative_unit_id) |
| `wards_tmp` | 3,321 | Vietnam wards (code, name, name_en, province_code FK, administrative_unit_id) |
| `administrative_regions` | 8 | Region lookup (id, name, name_en) |
| `administrative_units` | 8 | Unit type lookup (id, full_name, short_name) |
| `sapnhap_geojson_objects` | 3,355 | Combined geo objects with PostGIS geometry (ma, ten, magoc, malk, truocsapnhap, dientichkm2, bbox_wkt, geom_wkt, vn_ds_province_code FK, vn_ds_ward_code FK) |

### Key Columns

**`provinces_tmp`**: `code` (PK, e.g. "01", "02"), `name`, `name_en`, `full_name`, `code_name`, `administrative_unit_id`

**`wards_tmp`**: `code` (PK), `name`, `name_en`, `full_name`, `province_code` (FK → `provinces_tmp.code`), `administrative_unit_id`

**`sapnhap_geojson_objects`**: `ma` (PK), `ten` (name), `magoc` (parent FK, self-ref), `truocsapnhap` (pre-merge name), `dientichkm2` (area in km²), `bbox_wkt` (WKT POLYGON), `geom_wkt` (WKT MULTIPOLYGON), `vn_ds_province_code` (FK → `provinces_tmp.code`), `vn_ds_ward_code` (FK → `wards_tmp.code`)

### Common Query Patterns

**Count records:**
```sql
SELECT COUNT(*) FROM provinces_tmp;
SELECT COUNT(*) FROM wards_tmp;
```

**Data completeness check:**
```sql
SELECT
  (SELECT COUNT(*) FROM provinces_tmp) as provinces,
  (SELECT COUNT(*) FROM wards_tmp) as wards,
  (SELECT COUNT(*) FROM sapnhap_geojson_objects) as geo_objects;
```

**Join provinces and wards:**
```sql
SELECT p.name as province, w.name as ward
FROM provinces_tmp p
JOIN wards_tmp w ON p.code = w.province_code
WHERE p.code = '01'
LIMIT 10;
```

**GIS geometry completeness:**
```sql
SELECT
  COUNT(*) as total,
  COUNT(bbox_wkt) as with_bbox,
  COUNT(geom_wkt) as with_geom
FROM sapnhap_geojson_objects;
```

**Find orphaned records:**
```sql
-- Geo objects without matching province
SELECT COUNT(*) FROM sapnhap_geojson_objects
WHERE vn_ds_province_code IS NULL;

-- Geo objects without matching ward
SELECT COUNT(*) FROM sapnhap_geojson_objects
WHERE vn_ds_ward_code IS NULL;
```

**List wards in a specific province:**
```sql
SELECT code, name, name_en FROM wards_tmp
WHERE province_code = '01'
ORDER BY name;
```

**PostGIS spatial queries:**
```sql
-- Get area of a province geometry
SELECT ten, ST_Area(geom::geography) / 1000000 as area_km2
FROM sapnhap_geojson_objects
WHERE vn_ds_province_code = '01';

-- Find objects within a bounding box (example: Hà Nội area)
SELECT ma, ten, ST_AsText(bbox) FROM sapnhap_geojson_objects
WHERE ST_Within(geom, ST_MakeEnvelope(105.5, 20.5, 106.0, 21.5, 4326));
```

### Agent Instructions

When context triggers a database query:
1. Execute the SQL query using the `docker exec` command pattern above
2. Format results in a readable way (tables, lists, summaries)
3. Provide insights or follow-up suggestions if relevant
4. If the query returns an error, explain the issue and suggest a fix
5. Use `LIMIT` for large result sets to avoid overwhelming output
6. Explore schema with: `\dt` (list tables), `\d table_name` (describe table)
```

- [ ] **Step 2: Verify the replacement was applied correctly**

Run: `grep -n "## Database Query Skill" AGENTS.md`
Expected: Line number output showing the new section exists.

Run: `grep -n "### Database Connection (Docker)" AGENTS.md`
Expected: No output (old subsection removed).

- [ ] **Step 3: Commit**

```bash
git add AGENTS.md
git commit -m "docs: add universal Database Query Skill section to AGENTS.md"
```

---

### Task 2: Delete .claude/skills/db-query.md

**Files:**
- Delete: `dataset-generation-scripts/.claude/skills/db-query.md`

**Interfaces:**
- Consumes: Task 1 (new AGENTS.md section replaces this content)
- Produces: Clean removal of Claude Code-specific skill file; skill is now solely in AGENTS.md

- [ ] **Step 1: Delete the old skill file**

```bash
rm dataset-generation-scripts/.claude/skills/db-query.md
```

- [ ] **Step 2: Verify deletion**

```bash
test -f dataset-generation-scripts/.claude/skills/db-query.md && echo "STILL EXISTS" || echo "DELETED"
```
Expected: `DELETED`

- [ ] **Step 3: Remove empty skills directory if empty**

```bash
if [ -d "dataset-generation-scripts/.claude/skills" ] && [ -z "$(ls -A dataset-generation-scripts/.claude/skills)" ]; then
  rmdir dataset-generation-scripts/.claude/skills
fi
```

- [ ] **Step 4: Commit**

```bash
git add dataset-generation-scripts/.claude/skills/db-query.md
git commit -m "refactor: remove Claude Code-specific db-query skill (consolidated into AGENTS.md)"
```

---

### Task 3: Update CLAUDE.md — trim DB sections, add pointer to AGENTS.md

**Files:**
- Modify: `dataset-generation-scripts/CLAUDE.md` (lines 9-56 replaced, lines 31-41 replaced)

**Interfaces:**
- Consumes: Task 1 (new AGENTS.md section to reference)
- Produces: CLAUDE.md with reduced duplicate content, pointing to AGENTS.md for db-query skill

- [ ] **Step 1: Replace the "When to Use Database Queries" section (lines 9-30) and the "Database Access" section (lines 31-56) with a pointer**

Replace lines 9-56 of `dataset-generation-scripts/CLAUDE.md`:

**SEARCH block** (existing lines 9-56):
```
## When to Use Database Queries

**AUTOMATICALLY use database queries when the user asks about:**
- Data counts, totals, or statistics (e.g., "how many", "count", "total")
- Data verification or integrity checks (e.g., "check", "verify", "missing", "orphaned")
...
```

**REPLACE block:**
```markdown
## When to Use Database Queries

See [../AGENTS.md#database-query-skill](../AGENTS.md) for the universal db-query skill — auto-trigger rules, connection details, schema/table reference, and query patterns. Apply these rules automatically; do not wait for explicit invocation.

## Database Access

See [../AGENTS.md#database-query-skill](../AGENTS.md). Quick reminder: the database is accessible via:
```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "QUERY"
```
Container: `vn_provinces_postgres_container`, Database: `vn_provinces_tmp`, Port: `15432`
```

- [ ] **Step 2: Verify the replacement**

Run: `head -20 dataset-generation-scripts/CLAUDE.md | grep -A5 "When to Use Database Queries"`
Expected: Shows the new pointer text referencing AGENTS.md.

Run: `grep "vn_provinces_postgres_container" dataset-generation-scripts/CLAUDE.md`
Expected: Still shows the connection info (quick reminder preserved).

- [ ] **Step 3: Commit**

```bash
git add dataset-generation-scripts/CLAUDE.md
git commit -m "docs: trim CLAUDE.md DB sections, point to AGENTS.md db-query skill"
```

---

### Task 4: Update memory/feedback_auto_db_query.md

**Files:**
- Modify: `dataset-generation-scripts/memory/feedback_auto_db_query.md`

**Interfaces:**
- Consumes: Tasks 1-3 (skill now in AGENTS.md)
- Produces: Updated memory reflecting new location

- [ ] **Step 1: Update the feedback file to reference AGENTS.md instead of .claude/skills/**

Replace the content of `dataset-generation-scripts/memory/feedback_auto_db_query.md`:

```markdown
---
name: feedback_auto_db_query
description: User wants automatic skill invocation without manual commands
type: feedback
---

**Rule:** Skills should automatically trigger based on context, not require manual invocation.

**Why:** User explicitly asked "How can I make claude code aware when to use the db-query skill... I don't want to invoke the skill manually every time"

**Current location:** `AGENTS.md` → `## Database Query Skill` section

**How to apply:**
- The AGENTS.md section contains explicit auto-trigger keywords and patterns
- Be proactive: if user asks about data, statistics, or verification, run queries immediately
- Don't wait for explicit commands
- Common triggers: "how many", "count", "check", "find", "show", "verify", "missing", "database", "table"

**Example:**
- User says: "How many wards are there?"
- Response: Immediately run `docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "SELECT COUNT(*) FROM wards_tmp"`
- Don't say: "Would you like me to query the database?"
```

- [ ] **Step 2: Verify update**

Run: `grep "AGENTS.md" dataset-generation-scripts/memory/feedback_auto_db_query.md`
Expected: Shows the new `Current location` line.

- [ ] **Step 3: Commit**

```bash
git add dataset-generation-scripts/memory/feedback_auto_db_query.md
git commit -m "docs: update memory to reference AGENTS.md-based db-query skill"
```

---

### Task 5: Final verification

**Files:**
- Check: `AGENTS.md` (new section exists)
- Check: `dataset-generation-scripts/.claude/skills/db-query.md` (deleted)
- Check: `dataset-generation-scripts/CLAUDE.md` (points to AGENTS.md)
- Check: `dataset-generation-scripts/memory/feedback_auto_db_query.md` (updated)

- [ ] **Step 1: Verify AGENTS.md has the new section**

```bash
grep -c "## Database Query Skill" AGENTS.md
```
Expected: `1` (or more if repeated across subsections)

- [ ] **Step 2: Verify old file is deleted**

```bash
test ! -f dataset-generation-scripts/.claude/skills/db-query.md && echo "OK: old file deleted" || echo "FAIL: old file still exists"
```
Expected: `OK: old file deleted`

- [ ] **Step 3: Verify CLAUDE.md references AGENTS.md**

```bash
grep "AGENTS.md#database-query-skill" dataset-generation-scripts/CLAUDE.md
```
Expected: Shows the reference

- [ ] **Step 4: Verify all changes are committed**

```bash
git status
```
Expected: Working tree clean

- [ ] **Step 5: Verify git log shows expected commits**

```bash
git log --oneline -5
```
Expected: Recent commits show the 4 changes from Tasks 1-4
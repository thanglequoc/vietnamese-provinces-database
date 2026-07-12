# Design: Universal db-query Skill in AGENTS.md

**Date:** 2026-07-12
**Status:** Approved
**Approach:** A — Fully Inline in AGENTS.md

---

## Objective

Consolidate the project's database query skill into a single, universal definition in `AGENTS.md` so that any AI agent (Cline, Cursor, Windsurf, Claude Code, etc.) reading AGENTS.md automatically gets full db-query capabilities — auto-trigger rules, connection details, schema reference, and query patterns — without relying on agent-specific skill formats.

## Background

The project previously had a Claude Code-specific skill at `dataset-generation-scripts/.claude/skills/db-query.md`. This file used Claude Code's `.claude/skills/` format with YAML frontmatter and was only discoverable by Claude Code. Other agents had no access to this skill unless they happened to read that specific file.

Additionally, `CLAUDE.md` contained duplicate "When to Use Database Queries" and "Database Access" sections that overlapped with the skill file, creating maintenance burden and potential inconsistency.

The user's memory feedback (`feedback_auto_db_query.md`) explicitly states: skills should auto-trigger based on context, not require manual invocation.

## Design

### 1. New AGENTS.md Section: "Database Query Skill"

Add a comprehensive section to `AGENTS.md` that replaces the existing brief "Database Connection (Docker)" subsection. The new section contains:

#### Auto-Trigger Rules

An explicit list of trigger keywords and phrases that signal the agent should proactively run database queries:
- Count/total: "how many", "count", "total", "number of"
- Search/find: "find", "search", "show", "list", "get"
- Verification: "check", "verify", "missing", "orphaned", "validate"
- Data topics: "database", "table", "data", "records", "provinces", "wards"
- GIS: "geometry", "bbox", "geom", "gis", "spatial"
- Direct requests: "query from", "read from database", "get from [table]"

Instruction: "Do NOT wait for explicit skill invocation — proactively query when context suggests database information is needed."

Examples of auto-triggers:
- "How many wards are in Hà Nội?" → Run query immediately
- "Check if there are any missing GIS data" → Run verification query
- "Show me provinces without codes" → Query and display results

#### Connection Details

- Container: `vn_provinces_postgres_container`
- Database: `vn_provinces_tmp`
- Username: `postgres`
- Host: `localhost`
- Port: `15432`
- Password: `root`

Command patterns:
```bash
# Single query
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "YOUR_SQL_QUERY_HERE"

# Multi-line query (heredoc)
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp <<'EOF'
YOUR_MULTI_LINE_SQL_QUERY_HERE
EOF
```

#### Table & Schema Reference

**Core Tables:**
| Table | Records | Description |
|-------|---------|-------------|
| `provinces_tmp` | 34 | Vietnam provinces with codes |
| `wards_tmp` | 3,321 | Vietnam wards with codes |
| `administrative_regions` | — | Administrative regions |
| `administrative_units` | — | Administrative unit types |

**GIS Table:**
| Table | Records | Description |
|-------|---------|-------------|
| `sapnhap_geojson_objects` | 3,355 | Combined geo objects with geometry |

**Key Columns:**

`provinces_tmp`:
- `code` (PK) — Province code (e.g., "01", "02")
- `name` — Province name (cleaned, no prefix)
- `name_en` — English name
- `full_name` — Full name with administrative prefix
- `administrative_unit_id` (FK) — Unit type

`wards_tmp`:
- `code` (PK) — Ward code
- `name` — Ward name (cleaned)
- `name_en` — English name
- `province_code` (FK) → `provinces_tmp.code`
- `administrative_unit_id` (FK) — Unit type

`sapnhap_geojson_objects`:
- `ma` (PK) — Object identifier
- `ten` — Name
- `magoc` — Parent reference (self-referential FK)
- `malk` — Link code
- `truocsapnhap` — Pre-merge name
- `dientichkm2` — Area in km²
- `vn_ds_province_code` (FK) → `provinces_tmp.code`
- `vn_ds_ward_code` (FK) → `wards_tmp.code`
- `bbox_wkt` — Bounding box in WKT POLYGON format
- `geom_wkt` — Geometry in WKT MULTIPOLYGON format
- `bbox` — PostGIS geometry (generated from `bbox_wkt`)
- `geom` — PostGIS geometry (generated from `geom_wkt`)

**Key Relationships:**
- `wards_tmp.province_code` → `provinces_tmp.code`
- `provinces_tmp.administrative_unit_id` → `administrative_units.id`
- `wards_tmp.administrative_unit_id` → `administrative_units.id`
- `sapnhap_geojson_objects.vn_ds_province_code` → `provinces_tmp.code`
- `sapnhap_geojson_objects.vn_ds_ward_code` → `wards_tmp.code`
- `sapnhap_geojson_objects.magoc` → `sapnhap_geojson_objects.ma` (self-reference)

#### Common Query Patterns

- Count records
- Join provinces and wards
- Check GIS data completeness
- Find orphaned/missing records
- PostGIS spatial queries (`ST_AsText()`, `ST_Area()`, `ST_Contains()`)

#### Agent Instructions

When the database context triggers:
1. Execute the SQL query using the docker exec command
2. Format results readably (tables, lists, etc.)
3. Provide insights or follow-up suggestions if relevant
4. If the query returns an error, explain the issue and suggest fixes
5. Use `LIMIT` to avoid large result sets
6. Use `\dt` to list all tables, `\d table_name` to see table structure

### 2. Remove `.claude/skills/db-query.md`

Delete the file entirely. The `.claude/skills/` directory may remain empty or be removed. The `settings.local.json` file remains unchanged (it contains unrelated permissions).

### 3. Update `CLAUDE.md`

Replace the following sections in `CLAUDE.md` with a brief pointer to AGENTS.md:
- "When to Use Database Queries" section → replaced with: "See [../AGENTS.md#database-query-skill](../AGENTS.md) for the universal db-query skill — auto-trigger rules, connection details, and query patterns."
- "Database Access" / "Quick Database Access" section → same pointer
- "Database Schema" / "Key Tables" section → same pointer

Keep `CLAUDE.md` focused on Go subsystem-specific details (project structure, internal packages, modules, configuration) that are not in AGENTS.md.

### 4. Update Memory

Update `dataset-generation-scripts/memory/feedback_auto_db_query.md` to reflect:
- The skill is now AGENTS.md-based (not `.claude/skills/`)
- Location: `AGENTS.md` → "Database Query Skill" section
- The auto-trigger principle remains the same

## Files Affected

| File | Action | Details |
|------|--------|---------|
| `AGENTS.md` | Modify | Add "Database Query Skill" section (~80-100 lines), replace brief "Database Connection (Docker)" subsection |
| `dataset-generation-scripts/.claude/skills/db-query.md` | Delete | Content consolidated into AGENTS.md |
| `dataset-generation-scripts/CLAUDE.md` | Modify | Trim DB sections, add pointer to AGENTS.md |
| `dataset-generation-scripts/memory/feedback_auto_db_query.md` | Modify | Update skill location reference |

## Risks & Mitigations

| Risk | Severity | Mitigation |
|------|----------|------------|
| AGENTS.md grows by ~80-100 lines | Low | Well-structured with clear section headers; table format for schema keeps it compact |
| Claude Code loses `.claude/skills/` auto-discovery | Low | Claude Code reads AGENTS.md too; auto-trigger keywords in AGENTS.md serve the same purpose |
| Duplicate content between AGENTS.md and CLAUDE.md | Low | Design explicitly trims CLAUDE.md to a pointer, eliminating duplication |
| Schema changes in future | Low | Table reference is concise and easy to update in one place |

## Out of Scope

- No database schema or data changes
- No Go code changes
- No changes to `settings.local.json` (unrelated permissions)
- No changes to CI/CD pipeline
- No new tooling or scripts
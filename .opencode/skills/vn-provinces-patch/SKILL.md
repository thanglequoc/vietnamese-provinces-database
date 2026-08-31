---
name: vn-provinces-patch
description: Generate a data-only SQL upgrade patch for the VN provinces database by diffing the current postgresql/postgres_ImportData_vn_units.sql against an earlier git commit/tag (default HEAD), excluding the generated-timestamp header and all GIS data, then verify it in the local PostGIS docker container and store it under patch/<release_version>/. Use when asked to "generate a patch", "create an upgrade/release patch", "diff the postgres province data", "compare provinces", "update vn provinces to the next release", or when a git commit/tag is given as the old baseline.
---

# VN Provinces Patch Generator

Generate an SQL upgrade patch for the **VN province unit data only** (no GIS) by comparing the
current `postgresql/postgres_ImportData_vn_units.sql` against a previous version of the same file,
then verify and publish it as a drop-in upgrade script for downstream users.

## Scope

- **Covered**: `administrative_regions`, `administrative_units`, `provinces`, `wards` rows only.
- **Excluded**: all GIS data, everything under `postgresql/gis/` or `resources/gis/`, and the
  generated-timestamp header comment (`/* Created at: ... */`) — that metadata line changes on every
  regeneration and must NOT produce a patch.
- **Only PostgreSQL** baseline (`postgresql/`). Other engines (mysql, mssql, oracle, mongodb, …) are out of scope.

## Inputs & baseline resolution

| Role | Source |
|------|--------|
| **NEW** (target release) | working-tree file `postgresql/postgres_ImportData_vn_units.sql` |
| **OLD** (baseline) | resolved in priority order below |

OLD baseline resolution:

1. **User-specified ref** — if the user summons the skill with a git commit hash or tag
   (e.g. `v4.2.0`, `abc1234`), treat that ref as the OLD version:
   `git show <ref>:postgresql/postgres_ImportData_vn_units.sql`
2. **Default** — no ref given: OLD = `HEAD` version, NEW = working-tree file.
3. **No diff vs HEAD** (file is clean/unchanged): list the file's history with
   `git log --oneline -- postgresql/postgres_ImportData_vn_units.sql`, ask the user to pick a
   baseline (commit or tag), and use that. If the user has nothing, report "no data changes" and stop.

**Important**: the patch is only meaningful when the NEW working-tree file differs from the OLD
baseline. If they are identical, tell the user there is nothing to patch.

## Workflow

### Step 1 — Verify inputs

- Confirm `postgresql/postgres_ImportData_vn_units.sql` exists and the repo is a git worktree.
- Resolve the OLD baseline per the rules above. Record the resolved ref for later use (also needed for verification).

### Step 2 — Run the diff engine

From the repo root, run:

```bash
python3 .opencode/skills/vn-provinces-patch/scripts/diff_postgres_units.py \
  --new postgresql/postgres_ImportData_vn_units.sql \
  --old-ref <OLD_REF> \
  --path postgresql/postgres_ImportData_vn_units.sql \
  --detect-renames \
  --output /tmp/vn_units_patch_draft.sql
```

- Use `--old-file <file>` instead of `--old-ref` if the OLD version came from a local file rather than git.
- The script prints a per-table summary (added / changed / deleted) and writes the draft patch.
- Exit code `2` means **no data changes** — stop and report.
- `--detect-renames` flags deleted+added rows that differ only by code; it reports them but does not
  emit them. Confirm each rename with the user and add `UPDATE <table> SET code='<new>' WHERE code='<old>';`
  yourself (see Step 3).

### Step 3 — Review & refine the patch

Read the draft (`/tmp/vn_units_patch_draft.sql`) and finalize:

- **Ordering**: keep inserts in FK-safe order — regions → units → provinces → wards; deletes reverse
  (wards → provinces). The script already does this; preserve it.
- **Renames**: convert confirmed code changes to `UPDATE ... SET code=...`; ask the user before doing so.
- **Deletes (merges/absorptions)**: keep the `DELETE`, and add a comment telling users to update any
  references to the removed code (e.g. `-- Ward 00008 merged into ...; update existing references`).
- **Deleted province with surviving wards**: stop and ask — a province delete requires its wards to be
  deleted/moved first.
- **Whole-table added/removed** (schema drift): warn the user; data-only patches cannot create/drop tables.
- Add a header block documenting the release/decree context if known (ask the user for the decree
  number and effective date if available).

### Step 4 — Ask for the release version & create the folder

Ask the user: **"Which release version should this patch target?"** (e.g. `v4.3.0`).

Create `patch/<release_version>/` and write:
- `patch/<release_version>/<release_version>_patch.sql` — the finalized patch (file name matches the
  existing `patch/*/*_patch.sql` convention).
- `patch/<release_version>/README.md` — per-table change summary, optional decree context, and apply instructions.

### Step 5 — Mandatory verification (PostGIS docker container)

Verify the patch reproduces the NEW data exactly. The container
`vn_provinces_postgres_container` (user `postgres`, password `root`, port `15432`) is defined in
`dataset-generation-scripts/docker/docker-compose.yaml`. Start it if it is down:

```bash
docker compose -f dataset-generation-scripts/docker/docker-compose.yaml up -d
```

Then run the bootstrap → patch → compare flow:

```bash
# 1. Create the scratch database (always recreate from scratch)
docker exec vn_provinces_postgres_container psql -U postgres -c "DROP DATABASE IF EXISTS vn_provinces_patch_verification;"
docker exec vn_provinces_postgres_container psql -U postgres -c "CREATE DATABASE vn_provinces_patch_verification;"

# 2. Bootstrap schema (published schema matching the import file)
docker exec -i vn_provinces_postgres_container psql -U postgres -d vn_provinces_patch_verification \
  < postgresql/postgres_CreateTables_vn_units.sql

# 3. Load OLD data (the resolved baseline)
git show <OLD_REF>:postgresql/postgres_ImportData_vn_units.sql \
  | docker exec -i vn_provinces_postgres_container psql -U postgres -d vn_provinces_patch_verification

# 4. Apply the generated patch
docker exec -i vn_provinces_postgres_container psql -U postgres -d vn_provinces_patch_verification \
  < patch/<release_version>/<release_version>_patch.sql

# 5. Load NEW data into a separate `new` schema as the expected reference
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_patch_verification \
  -c "CREATE SCHEMA new;"
{ echo "SET search_path TO new;"; cat postgresql/postgres_CreateTables_vn_units.sql; } \
  | docker exec -i vn_provinces_postgres_container psql -U postgres -d vn_provinces_patch_verification
{ echo "SET search_path TO new;"; cat postgresql/postgres_ImportData_vn_units.sql; } \
  | docker exec -i vn_provinces_postgres_container psql -U postgres -d vn_provinces_patch_verification
```

Run the comparison (all checks must return **0** mismatches). Note the `-i` flag so stdin is
connected to psql inside the container:

```bash
docker exec -i vn_provinces_postgres_container psql -U postgres -d vn_provinces_patch_verification <<'SQL'
SELECT 'province missing in patched' AS check_name, count(*) FROM new.provinces n
  LEFT JOIN provinces p ON p.code = n.code WHERE p.code IS NULL
UNION ALL SELECT 'province extra in patched', count(*) FROM provinces p
  LEFT JOIN new.provinces n ON n.code = p.code WHERE n.code IS NULL
UNION ALL SELECT 'province name mismatch', count(*) FROM provinces p
  JOIN new.provinces n ON p.code = n.code WHERE p."name" IS DISTINCT FROM n."name"
UNION ALL SELECT 'province name_en mismatch', count(*) FROM provinces p
  JOIN new.provinces n ON p.code = n.code WHERE p.name_en IS DISTINCT FROM n.name_en
UNION ALL SELECT 'ward missing in patched', count(*) FROM new.wards n
  LEFT JOIN wards p ON p.code = n.code WHERE p.code IS NULL
UNION ALL SELECT 'ward extra in patched', count(*) FROM wards p
  LEFT JOIN new.wards n ON n.code = p.code WHERE n.code IS NULL
UNION ALL SELECT 'ward name mismatch', count(*) FROM wards p
  JOIN new.wards n ON p.code = n.code WHERE p."name" IS DISTINCT FROM n."name"
UNION ALL SELECT 'ward name_en mismatch', count(*) FROM wards p
  JOIN new.wards n ON p.code = n.code WHERE p.name_en IS DISTINCT FROM n.name_en
UNION ALL SELECT 'region mismatch', count(*) FROM new.administrative_regions n
  FULL OUTER JOIN administrative_regions p ON p.id = n.id
  WHERE p.id IS NULL OR n.id IS NULL OR p."name" IS DISTINCT FROM n."name" OR p.name_en IS DISTINCT FROM n.name_en
UNION ALL SELECT 'unit mismatch', count(*) FROM new.administrative_units n
  FULL OUTER JOIN administrative_units p ON p.id = n.id
  WHERE p.id IS NULL OR n.id IS NULL OR p.full_name IS DISTINCT FROM n.full_name OR p.full_name_en IS DISTINCT FROM n.full_name_en
UNION ALL SELECT 'province row count', (SELECT count(*) FROM provinces)
UNION ALL SELECT 'ward row count', (SELECT count(*) FROM wards);
SQL
```

**Pass criterion**: every `mismatches`/count check row must equal **0** (the count rows verify parity).
On **failure**: do NOT finalize the patch — debug (re-check the diff, ordering, FK conflicts), fix,
re-apply, and re-verify.

On success, clean up (keep the DB only if the user asks to inspect it):

```bash
docker exec vn_provinces_postgres_container psql -U postgres -c "DROP DATABASE IF EXISTS vn_provinces_patch_verification;"
```

### Step 6 — Report

Summarize for the user:
- OLD baseline ref used, NEW file, and the release version folder.
- Per-table change counts (added / changed / deleted), renames, and any merged/absorbed codes.
- Verification result (all checks passed) and the patch file location.

## Patch generation rules (recap)

- Data only — never touch GIS, and never let the `Created at` comment cause a diff.
- Diff rows as keyed sets, so reordering rows in the file is NOT a change.
- `INSERT` new rows with the full column list; `UPDATE` only changed columns (`SET col=... WHERE <pk>=...`);
  `DELETE` removed rows with a reference-migration note.
- Escape `'` as `''` in string values.
- Keep FK-safe statement order (see Step 3).

## Cross-platform portability

This skill follows the open **Agent Skills** spec (`SKILL.md` with `name`/`description` frontmatter).
It lives in `.opencode/skills/` for opencode but the identical folder can be exposed to any agent
platform — see the skill `README.md` for symlinking / `skills.paths` wiring per platform.

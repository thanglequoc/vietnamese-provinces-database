# GIS Geometry Fix — Before/After Comparison Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Import both pre-fix (July 12) and post-fix (July 20) GIS data into local Docker PostGIS, run area/topology comparison queries on 78 fixed wards + 34 provinces, produce a ranked Markdown report confirming no data corruption from the `ST_MakeValid` fix.

**Architecture:** Sequential PostgreSQL import into ephemeral Docker PostGIS container: old data → snapshot → cleanup → new data. Then PostGIS comparison queries (`ST_Area` geodesic, `ST_Equals`, `ST_NPoints`, `ST_IsValid`) against both tables, with output written to a Markdown report file.

**Tech Stack:** Docker (PostGIS 15-3.3), PostgreSQL, psql CLI, Bash, shell redirection

## Global Constraints

- Local Docker PostGIS at `localhost:15432`, database `vn_provinces_tmp`, user `postgres`
- `fresh_cleanup.sql` only drops `sapnhap_geojson_objects` — the `_old` snapshot survives
- Area computed with `ST_Area(geom::geography)` for geodesic accuracy (WGS84)
- Report written to `development/corrupted_gis_ward_data/gis_comparison_report_2026-07-24.md`
- Thresholds: < 0.001% = 🟢 OK, 0.001%–0.01% = 🟡 WARN, > 0.01% = 🔴 ALARM

---

### Task 1: Start Docker PostGIS and Initialize Schema

**Files:**
- Read: `dataset-generation-scripts/docker/docker-compose.yaml`
- Execute via stdin: `dataset-generation-scripts/resources/db_table_init.sql`, `dataset-generation-scripts/resources/db_region_administrative_unit.sql`, `dataset-generation-scripts/resources/gis/sapnhap_bando_tables.sql`

**Interfaces:**
- Produces: Running PostGIS container at `localhost:15432` with empty tables `provinces_tmp`, `wards_tmp`, `administrative_regions`, `administrative_units`, `sapnhap_geojson_objects`, `sapnhap_wards_gis`, `sapnhap_provinces_gis`, `sapnhap_wards`, `sapnhap_provinces`

- [ ] **Step 1: Start Docker container**

```bash
docker compose -f dataset-generation-scripts/docker/docker-compose.yaml up -d
```

- [ ] **Step 2: Verify container is running**

```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "SELECT 1 AS alive;"
```

Expected output: row with `alive = 1`.

- [ ] **Step 3: Check if database has existing data — drop everything to start clean**

```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "DROP TABLE IF EXISTS sapnhap_geojson_objects, sapnhap_geojson_objects_old, sapnhap_wards_gis, sapnhap_provinces_gis, sapnhap_wards, sapnhap_provinces, wards_tmp, districts_tmp, provinces_tmp, wards_tmp_seed, provinces_tmp_seed, wards, districts, provinces, administrative_units, administrative_regions CASCADE;"
```

- [ ] **Step 4: Initialize core tables**

```bash
cat dataset-generation-scripts/resources/db_table_init.sql | docker exec -i vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp
```

Expected: `CREATE TABLE` messages for `provinces_tmp`, `wards_tmp`, `districts_tmp`, `administrative_regions`, `administrative_units`.

- [ ] **Step 5: Seed administrative regions and units**

```bash
cat dataset-generation-scripts/resources/db_region_administrative_unit.sql | docker exec -i vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp
```

Expected: `INSERT` messages for 8 regions and 8 units.

- [ ] **Step 6: Initialize GIS tables**

```bash
cat dataset-generation-scripts/resources/gis/sapnhap_bando_tables.sql | docker exec -i vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp
```

Expected: `CREATE TABLE` for `sapnhap_geojson_objects` and related GIS tables.

- [ ] **Step 7: Commit**

```bash
git add -A && git commit -m "chore: initialize PostGIS schema for GIS comparison"
```

Only commit if the CLI tool `git status` shows changes. If working directory was already clean, skip this commit — the database state is ephemeral.

---

### Task 2: Import Administrative Units

**Files:**
- Execute via stdin: `postgresql/postgres_ImportData_vn_units.sql`

**Interfaces:**
- Consumes: Empty `provinces_tmp`, `wards_tmp`, `districts_tmp` tables from Task 1
- Produces: Populated `provinces_tmp` (34 rows), `wards_tmp` (3,321 rows), `districts_tmp` (populated)

- [ ] **Step 1: Import administrative unit data**

```bash
cat postgresql/postgres_ImportData_vn_units.sql | docker exec -i vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp
```

This file contains INSERT statements. Expected: stream of `INSERT 0 1` messages.

- [ ] **Step 2: Verify province count**

```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "SELECT COUNT(*) AS province_count FROM provinces_tmp;"
```

Expected: `province_count = 34`.

- [ ] **Step 3: Verify ward count**

```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "SELECT COUNT(*) AS ward_count FROM wards_tmp;"
```

Expected: `ward_count = 3321`.

---

### Task 3: Import OLD GIS Data (July 12 — Pre-Fix) and Snapshot

**Files:**
- Read: `postgresql/gis/postgresql_ImportData_gis_2026-07-12__19_50_50.sql.zip`
- Produce (in DB): `sapnhap_geojson_objects` (3,355 rows), `sapnhap_geojson_objects_old` (snapshot)

**Interfaces:**
- Consumes: Populated `provinces_tmp`, `wards_tmp` from Task 2; empty `sapnhap_geojson_objects` from Task 1
- Produces: `sapnhap_geojson_objects` with pre-fix data; `sapnhap_geojson_objects_old` snapshot table

- [ ] **Step 1: Import OLD GIS data from zipped SQL**

```bash
unzip -p postgresql/gis/postgresql_ImportData_gis_2026-07-12__19_50_50.sql.zip | docker exec -i vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp
```

The `-p` flag pipes unzipped content to stdout without extracting to disk. Expected: stream of `INSERT 0 1` messages.

- [ ] **Step 2: Verify old data record count**

```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "SELECT COUNT(*) AS old_count FROM sapnhap_geojson_objects;"
```

Expected: `old_count = 3355`.

- [ ] **Step 3: Snapshot old data into `_old` table**

```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp <<'EOF'
CREATE TABLE sapnhap_geojson_objects_old AS SELECT * FROM sapnhap_geojson_objects;
EOF
```

Expected: `SELECT 3355`.

- [ ] **Step 4: Verify snapshot has same count**

```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "SELECT COUNT(*) AS snapshot_count FROM sapnhap_geojson_objects_old;"
```

Expected: `snapshot_count = 3355`.

---

### Task 4: Clean and Import NEW GIS Data (July 20 — Post-Fix)

**Files:**
- Execute: `dataset-generation-scripts/resources/fresh_cleanup.sql`
- Read: `dataset-generation-scripts/output/gis/postgresql_ImportData_gis_2026-07-20__23_14_35.sql` (145 MB)

**Interfaces:**
- Consumes: `sapnhap_geojson_objects_old` must survive cleanup; `provinces_tmp`, `wards_tmp` will be re-populated
- Produces: `sapnhap_geojson_objects` with post-fix data (3,355 rows)

- [ ] **Step 1: Run fresh_cleanup.sql (drops all tables EXCEPT `_old`)**

```bash
cat dataset-generation-scripts/resources/fresh_cleanup.sql | docker exec -i vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp
```

Expected: `DROP TABLE` messages for 13 tables. `sapnhap_geojson_objects_old` is NOT in the cleanup script, so it survives.

- [ ] **Step 2: Verify `_old` table survived cleanup**

```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "SELECT COUNT(*) AS survived_count FROM sapnhap_geojson_objects_old;"
```

Expected: `survived_count = 3355`. If this fails, the `_old` table was dropped — stop and investigate.

- [ ] **Step 3: Re-initialize schema (repeat Task 1, Steps 4-6)**

```bash
cat dataset-generation-scripts/resources/db_table_init.sql | docker exec -i vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp
cat dataset-generation-scripts/resources/db_region_administrative_unit.sql | docker exec -i vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp
cat dataset-generation-scripts/resources/gis/sapnhap_bando_tables.sql | docker exec -i vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp
```

- [ ] **Step 4: Re-import administrative units (repeat Task 2, Step 1)**

```bash
cat postgresql/postgres_ImportData_vn_units.sql | docker exec -i vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp
```

- [ ] **Step 5: Import NEW GIS data (145 MB file)**

```bash
docker exec -i vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp < dataset-generation-scripts/output/gis/postgresql_ImportData_gis_2026-07-20__23_14_35.sql
```

This is the largest file. Expected: stream of `INSERT 0 1` messages. May take 30-60 seconds.

- [ ] **Step 6: Verify new data record count**

```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "SELECT COUNT(*) AS new_count FROM sapnhap_geojson_objects;"
```

Expected: `new_count = 3355`.

- [ ] **Step 7: Confirm both tables coexist**

```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "SELECT 'old' AS dataset, COUNT(*) FROM sapnhap_geojson_objects_old UNION ALL SELECT 'new', COUNT(*) FROM sapnhap_geojson_objects;"
```

Expected:
```
 dataset | count
---------+-------
 old     |  3355
 new     |  3355
```

---

### Task 5: Run Tier 1 — Area Comparison (78 Fixed Wards)

**Files:**
- Read: `dataset-generation-scripts/output/gis_geometry_fix_log_2026-07-20__23_14_32.txt` (for the 78 ward codes)
- Produce (in DB): Query results to parse for report

**Interfaces:**
- Consumes: `sapnhap_geojson_objects` (new), `sapnhap_geojson_objects_old` (old) — both with 3,355 rows
- Produces: Area comparison table for 78 wards (ranked by % diff DESC)

- [ ] **Step 1: Run ward area comparison query and save to temp file**

```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp <<'EOF' > /tmp/ward_area_comparison.txt
SELECT 
    new.ma,
    new.ten,
    new.vn_ds_province_code,
    ROUND(ST_Area(ST_GeomFromText(old.geom_wkt, 4326)::geography) / 1000000, 6) AS old_area_km2,
    ROUND(ST_Area(ST_GeomFromText(new.geom_wkt, 4326)::geography) / 1000000, 6) AS new_area_km2,
    ROUND(
        (ST_Area(ST_GeomFromText(new.geom_wkt, 4326)::geography) - 
         ST_Area(ST_GeomFromText(old.geom_wkt, 4326)::geography)) / 1000000, 6
    ) AS area_diff_km2,
    ROUND(
        ABS((ST_Area(ST_GeomFromText(new.geom_wkt, 4326)::geography) - 
             ST_Area(ST_GeomFromText(old.geom_wkt, 4326)::geography)) 
        / NULLIF(ST_Area(ST_GeomFromText(old.geom_wkt, 4326)::geography), 0)) * 100, 6
    ) AS area_diff_pct
FROM sapnhap_geojson_objects new
JOIN sapnhap_geojson_objects_old old ON new.ma = old.ma
WHERE new.ma IN (
    '19066','31673','05542','06577','06607','03356','03472','03549','03394','03760',
    '15661','16177','19351','22504','21925','21943','23602','23586','21997','21835',
    '23611','23764','23767','23728','21892','24502','24529','25459','25585','25588',
    '25498','25510','26461','25843','25777','25807','00832','01096','00820','02788',
    '03583','03434','21040','23908','06565','06541','03460','03358','03352','19333',
    '16186','20656','20965','20242','20257','20669','23332','22870','22624','22888',
    '22741','22759','21985','24846','25567','28087','28075','04402','31249','30028',
    '32071','02842','31261','11983','12452','01075','00958','30154'
)
ORDER BY area_diff_pct DESC;
EOF
```

- [ ] **Step 2: Verify query returned 78 rows**

```bash
wc -l /tmp/ward_area_comparison.txt
```

Expected: ~82 lines (78 data rows + header + separator). Confirm with:

```bash
grep -c '^ ' /tmp/ward_area_comparison.txt
```

Should be at least 78 (data rows prefixed with whitespace in psql output).

- [ ] **Step 3: Inspect top 5 results for anomalies**

```bash
head -20 /tmp/ward_area_comparison.txt
```

Look for any `area_diff_pct` values > 0.01% (ALARM threshold). Document findings.

---

### Task 6: Run Tier 1 — Area Comparison (34 Province Guard)

**Files:**
- Produce (in DB): Query results for all 34 provinces

**Interfaces:**
- Consumes: Same two tables from Task 5
- Produces: Area comparison table for 34 provinces

- [ ] **Step 1: Run province area comparison query**

```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp <<'EOF' > /tmp/province_area_comparison.txt
SELECT 
    new.ma,
    new.ten,
    ROUND(ST_Area(ST_GeomFromText(old.geom_wkt, 4326)::geography) / 1000000, 4) AS old_area_km2,
    ROUND(ST_Area(ST_GeomFromText(new.geom_wkt, 4326)::geography) / 1000000, 4) AS new_area_km2,
    ROUND(
        (ST_Area(ST_GeomFromText(new.geom_wkt, 4326)::geography) - 
         ST_Area(ST_GeomFromText(old.geom_wkt, 4326)::geography)) / 1000000, 4
    ) AS area_diff_km2,
    ROUND(
        ABS((ST_Area(ST_GeomFromText(new.geom_wkt, 4326)::geography) - 
             ST_Area(ST_GeomFromText(old.geom_wkt, 4326)::geography)) 
        / NULLIF(ST_Area(ST_GeomFromText(old.geom_wkt, 4326)::geography), 0)) * 100, 6
    ) AS area_diff_pct
FROM sapnhap_geojson_objects new
JOIN sapnhap_geojson_objects_old old ON new.ma = old.ma
WHERE new.magoc IS NULL
ORDER BY area_diff_pct DESC;
EOF
```

`magoc IS NULL` identifies province-level records (they have no parent reference).

- [ ] **Step 2: Verify 34 rows returned**

```bash
grep -c '^ ' /tmp/province_area_comparison.txt
```

Expected: 34 data rows.

- [ ] **Step 3: Check for any nonzero area changes**

```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "
SELECT new.ma, new.ten,
  ROUND(ABS((ST_Area(ST_GeomFromText(new.geom_wkt, 4326)::geography) - ST_Area(ST_GeomFromText(old.geom_wkt, 4326)::geography)) / NULLIF(ST_Area(ST_GeomFromText(old.geom_wkt, 4326)::geography), 0)) * 100, 6) AS area_diff_pct
FROM sapnhap_geojson_objects new
JOIN sapnhap_geojson_objects_old old ON new.ma = old.ma
WHERE new.magoc IS NULL AND ABS((ST_Area(ST_GeomFromText(new.geom_wkt, 4326)::geography) - ST_Area(ST_GeomFromText(old.geom_wkt, 4326)::geography)) / NULLIF(ST_Area(ST_GeomFromText(old.geom_wkt, 4326)::geography), 0)) * 100 > 0.000001
ORDER BY area_diff_pct DESC;
"
```

Any row returned = unexpected province change. Document in report if any appear.

---

### Task 7: Run Tier 2 — Topology Change Detection

**Files:**
- Produce: Topology comparison for the 78 fixed wards

**Interfaces:**
- Consumes: Same two tables
- Produces: `ST_Equals`, `ST_NPoints` delta, `ST_NumGeometries` delta, validity flags per ward

- [ ] **Step 1: Run topology comparison query**

```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp <<'EOF' > /tmp/topology_comparison.txt
SELECT 
    new.ma,
    new.ten,
    new.vn_ds_province_code,
    ST_Equals(
        ST_GeomFromText(new.geom_wkt, 4326), 
        ST_GeomFromText(old.geom_wkt, 4326)
    ) AS is_geometrically_equal,
    ST_NPoints(ST_GeomFromText(new.geom_wkt, 4326)) - 
    ST_NPoints(ST_GeomFromText(old.geom_wkt, 4326)) AS point_count_delta,
    ST_NPoints(ST_GeomFromText(old.geom_wkt, 4326)) AS old_point_count,
    ST_NPoints(ST_GeomFromText(new.geom_wkt, 4326)) AS new_point_count,
    ST_NumGeometries(ST_GeomFromText(new.geom_wkt, 4326)) - 
    ST_NumGeometries(ST_GeomFromText(old.geom_wkt, 4326)) AS subgeom_count_delta,
    NOT ST_IsValid(ST_GeomFromText(old.geom_wkt, 4326)) AS old_was_invalid,
    NOT ST_IsValid(ST_GeomFromText(new.geom_wkt, 4326)) AS new_is_invalid
FROM sapnhap_geojson_objects new
JOIN sapnhap_geojson_objects_old old ON new.ma = old.ma
WHERE new.ma IN (
    '19066','31673','05542','06577','06607','03356','03472','03549','03394','03760',
    '15661','16177','19351','22504','21925','21943','23602','23586','21997','21835',
    '23611','23764','23767','23728','21892','24502','24529','25459','25585','25588',
    '25498','25510','26461','25843','25777','25807','00832','01096','00820','02788',
    '03583','03434','21040','23908','06565','06541','03460','03358','03352','19333',
    '16186','20656','20965','20242','20257','20669','23332','22870','22624','22888',
    '22741','22759','21985','24846','25567','28087','28075','04402','31249','30028',
    '32071','02842','31261','11983','12452','01075','00958','30154'
)
ORDER BY abs(point_count_delta) DESC, abs(subgeom_count_delta) DESC;
EOF
```

- [ ] **Step 2: Verify 78 rows returned**

```bash
grep -c '^ ' /tmp/topology_comparison.txt
```

Expected: 78 data rows.

- [ ] **Step 3: Quick sanity checks**

```bash
# All 78 old records should be invalid
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "
SELECT COUNT(*) AS old_invalid_count FROM sapnhap_geojson_objects_old
WHERE NOT ST_IsValid(ST_GeomFromText(geom_wkt, 4326))
  AND ma IN ('19066','31673','05542','06577','06607','03356','03472','03549','03394','03760','15661','16177','19351','22504','21925','21943','23602','23586','21997','21835','23611','23764','23767','23728','21892','24502','24529','25459','25585','25588','25498','25510','26461','25843','25777','25807','00832','01096','00820','02788','03583','03434','21040','23908','06565','06541','03460','03358','03352','19333','16186','20656','20965','20242','20257','20669','23332','22870','22624','22888','22741','22759','21985','24846','25567','28087','28075','04402','31249','30028','32071','02842','31261','11983','12452','01075','00958','30154');
"
```

Expected: `old_invalid_count = 78`.

```bash
# All 78 new records should now be valid
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "
SELECT COUNT(*) AS new_invalid_count FROM sapnhap_geojson_objects
WHERE NOT ST_IsValid(ST_GeomFromText(geom_wkt, 4326))
  AND ma IN ('19066','31673','05542','06577','06607','03356','03472','03549','03394','03760','15661','16177','19351','22504','21925','21943','23602','23586','21997','21835','23611','23764','23767','23728','21892','24502','24529','25459','25585','25588','25498','25510','26461','25843','25777','25807','00832','01096','00820','02788','03583','03434','21040','23908','06565','06541','03460','03358','03352','19333','16186','20656','20965','20242','20257','20669','23332','22870','22624','22888','22741','22759','21985','24846','25567','28087','28075','04402','31249','30028','32071','02842','31261','11983','12452','01075','00958','30154');
"
```

Expected: `new_invalid_count = 0`.

---

### Task 8: Run Tier 3 — Data Integrity Checks

**Files:**
- Produce: Integrity check results

**Interfaces:**
- Consumes: All populated tables
- Produces: Counts for remaining invalid geometries, FK orphans

- [ ] **Step 1: Global invalid geometry check**

```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "
SELECT COUNT(*) AS remaining_invalid FROM sapnhap_geojson_objects
WHERE NOT ST_IsValid(ST_GeomFromText(geom_wkt, 4326));
"
```

Expected: `remaining_invalid = 0`.

- [ ] **Step 2: Foreign key integrity check**

```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "
SELECT COUNT(*) AS province_orphans FROM sapnhap_geojson_objects
WHERE vn_ds_province_code IS NOT NULL
  AND vn_ds_province_code NOT IN (SELECT code FROM provinces_tmp);
"
```

Expected: `province_orphans = 0`.

```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "
SELECT COUNT(*) AS ward_orphans FROM sapnhap_geojson_objects
WHERE vn_ds_ward_code IS NOT NULL
  AND vn_ds_ward_code NOT IN (SELECT code FROM wards_tmp);
"
```

Expected: `ward_orphans = 0`.

- [ ] **Step 3: Check that no records have NULL geometry**

```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "
SELECT COUNT(*) AS null_geom FROM sapnhap_geojson_objects WHERE geom_wkt IS NULL;
"
```

Expected: `null_geom = 0`.

- [ ] **Step 4: Record all integrity results to a temp file**

```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp <<'EOF' > /tmp/integrity_checks.txt
SELECT 'Total records (new)' AS check_name, COUNT(*)::text AS result FROM sapnhap_geojson_objects
UNION ALL
SELECT 'Total records (old)', COUNT(*)::text FROM sapnhap_geojson_objects_old
UNION ALL
SELECT 'Remaining invalid (new)', COUNT(*)::text FROM sapnhap_geojson_objects WHERE NOT ST_IsValid(ST_GeomFromText(geom_wkt, 4326))
UNION ALL
SELECT 'Province FK orphans', COUNT(*)::text FROM sapnhap_geojson_objects WHERE vn_ds_province_code IS NOT NULL AND vn_ds_province_code NOT IN (SELECT code FROM provinces_tmp)
UNION ALL
SELECT 'Ward FK orphans', COUNT(*)::text FROM sapnhap_geojson_objects WHERE vn_ds_ward_code IS NOT NULL AND vn_ds_ward_code NOT IN (SELECT code FROM wards_tmp)
UNION ALL
SELECT 'NULL geometry', COUNT(*)::text FROM sapnhap_geojson_objects WHERE geom_wkt IS NULL;
EOF
```

---

### Task 9: Compile Comparison Report

**Files:**
- Create: `development/corrupted_gis_ward_data/gis_comparison_report_2026-07-24.md`
- Read: `/tmp/ward_area_comparison.txt`, `/tmp/province_area_comparison.txt`, `/tmp/topology_comparison.txt`, `/tmp/integrity_checks.txt`

**Interfaces:**
- Consumes: All query results from Tasks 5-8
- Produces: Complete Markdown report with summary dashboard, ranked tables, topology anomalies, province guard, integrity results, final verdict

- [ ] **Step 1: Parse ward area results and classify each row**

Read `/tmp/ward_area_comparison.txt`. For each ward row, extract `area_diff_pct` and classify:
- < 0.001% → 🟢 OK
- 0.001% to 0.01% → 🟡 WARN
- > 0.01% → 🔴 ALARM
- Also check absolute: if old_area < 1.0 km² and abs(area_diff_km2) > 0.0001 → 🟡 WARN (absolute threshold for tiny wards)

Count totals: OK count, WARN count, ALARM count.

- [ ] **Step 2: Write report header and summary dashboard**

```markdown
# GIS Geometry Fix — Before/After Comparison Report

**Generated**: 2026-07-24
**Source**: `postgresql_ImportData_gis_2026-07-12__19_50_50.sql` (old) vs `postgresql_ImportData_gis_2026-07-20__23_14_35.sql` (new)
**Fix**: `ST_MakeValid` + `ST_CollectionExtract(..., 3)` on 78 self-intersecting ward geometries

## Summary Dashboard

| Metric | Value |
|--------|-------|
| Total wards compared | 78 |
| Total provinces compared | 34 |
| 🟢 OK (< 0.001%) | {ok_count} |
| 🟡 WARN (0.001%-0.01%) | {warn_count} |
| 🔴 ALARM (> 0.01%) | {alarm_count} |
| Max area change | {max_ward_name} ({max_change_pct}%) |
| Provinces with unexpected change | {unexpected_province_count} |
```

- [ ] **Step 3: Write ranked ward comparison table**

For each ward (sorted by `area_diff_pct` DESC), write a Markdown table row:

```markdown
## Ranked Ward Comparison (sorted by % area change)

| Rank | ma | Name | Prov | Old Area (km²) | New Area (km²) | Diff (km²) | Diff % | Status |
|------|----|------|------|----------------|----------------|------------|--------|--------|
| 1 | {ma} | {ten} | {prov} | {old_area} | {new_area} | {diff} | {pct}% | {status_emoji} |
...
| 78 | {ma} | {ten} | {prov} | {old_area} | {new_area} | {diff} | {pct}% | {status_emoji} |
```

If any WARN or ALARM rows exist, add a **Notes** subsection below the table explaining each one.

- [ ] **Step 4: Write topology anomalies section**

From `/tmp/topology_comparison.txt`, identify:
- Wards where `ST_Equals = false` (expected for all 78 — note this)
- Wards where `abs(point_count_delta) > 10` (significant vertex changes — flag)
- Wards where `subgeom_count_delta != 0` (polygon splitting/merging — flag)
- Wards where `new_is_invalid = true` (fix failed — 🔴 ALARM)

```markdown
## Topology Changes

**Expected**: All 78 wards show `ST_Equals = false` (geometry was modified by the fix).
**Expected**: All 78 old records show `old_was_invalid = true`.
**Expected**: All 78 new records show `new_is_invalid = false`.

### Point Count Changes

| ma | Name | Old Points | New Points | Delta |
|----|------|------------|------------|-------|

### Sub-Geometry Count Changes

| ma | Name | Delta |
|----|------|-------|

### Fix Failures (still invalid after fix)

(None expected — if any exist, list here with 🔴 ALARM)
```

- [ ] **Step 5: Write province guard section**

From `/tmp/province_area_comparison.txt`, list all 34 provinces. If any show `area_diff_pct > 0%`, flag them.

```markdown
## Province Guard Check

| Code | Name | Old Area (km²) | New Area (km²) | Diff % | Status |
|------|------|----------------|----------------|--------|--------|
```

- [ ] **Step 6: Write data integrity section**

From `/tmp/integrity_checks.txt`:

```markdown
## Data Integrity

| Check | Result |
|-------|--------|
| Remaining invalid geometries (new) | 0 |
| Province FK orphans | 0 |
| Ward FK orphans | 0 |
| NULL geometries | 0 |
```

- [ ] **Step 7: Write final verdict**

Based on thresholds:
- If alarm_count = 0 → "✅ **SAFE TO USE** — No data corruption detected. All 78 ward fixes are microscopic (sub-0.01% area change)."
- If alarm_count > 0 → "🔴 **BLOCKED** — {alarm_count} ward(s) show area changes > 0.01%. Investigation required before using new data."

- [ ] **Step 8: Commit report**

```bash
git add development/corrupted_gis_ward_data/gis_comparison_report_2026-07-24.md
git commit -m "report: GIS geometry fix before/after comparison (78 wards, 34 provinces)"
```

---

### Task 10: Cleanup Temporary Files

**Files:**
- Remove: `/tmp/ward_area_comparison.txt`, `/tmp/province_area_comparison.txt`, `/tmp/topology_comparison.txt`, `/tmp/integrity_checks.txt`

- [ ] **Step 1: Remove temp files**

```bash
rm -f /tmp/ward_area_comparison.txt /tmp/province_area_comparison.txt /tmp/topology_comparison.txt /tmp/integrity_checks.txt
```

- [ ] **Step 2: Stop Docker container (optional — leave running for future use)**

```bash
# Optional — only if container is not needed
docker compose -f dataset-generation-scripts/docker/docker-compose.yaml down
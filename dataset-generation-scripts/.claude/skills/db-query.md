---
name: db-query
description: Execute SQL queries against the vn_provinces_tmp PostgreSQL database and return formatted results
triggers:
  - database
  - query
  - sql
  - check the database
  - run a query
  - how many
  - count
  - find
  - search
  - list
  - show
  - verify
  - check
  - provinces
  - wards
  - records
  - table
  - data
  - missing
  - orphaned
  - geometry
  - gis
  - database
  - postgres
  - psql
---

You are a database query assistant for the vn_provinces_tmp PostgreSQL database.

## Database Connection

The database runs in a Docker container with these details:
- Container name: `vn_provinces_postgres_container`
- Database: `vn_provinces_tmp`
- Username: `postgres`
- Host: localhost
- Port: 15432

## How to Execute Queries

Use this command pattern to execute SQL queries:

```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "YOUR_SQL_QUERY_HERE"
```

For complex multi-line queries, use heredoc:

```bash
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp <<'EOF'
YOUR_MULTI_LINE_SQL_QUERY_HERE
EOF
```

## Available Tables

### Core Tables
- `provinces_tmp` - Vietnam provinces (34 records)
- `wards_tmp` - Vietnam wards (3,321 records)
- `administrative_regions` - Administrative regions
- `administrative_units` - Administrative unit types

### SAPNhap Tables
- `sapnhap_provinces` - Province metadata from SAPNhap site (34 records)
- `sapnhap_wards` - Ward metadata from SAPNhap site (3,321 records)
- `sapnhap_provinces_gis` - Province GIS geometry data (34 records)
- `sapnhap_wards_gis` - Ward GIS geometry data (3,321 records)
- `sapnhap_geojson_objects` - Combined geo objects (3,355 records)

### Seed Tables
- `provinces_tmp_seed` - Province seed data
- `wards_tmp_seed` - Ward seed data

## Important Columns

### provinces_tmp / sapnhap_provinces
- `code` - Province code (e.g., "01", "02")
- `name` / `tentinh` - Province name (cleaned, no prefix)
- `mahc` - Administrative code (sapnhap_provinces)

### wards_tmp / sapnhap_wards
- `code` - Ward code
- `name` / `tenhc` - Ward name (cleaned)
- `province_code` / `matinh` - Foreign key to province
- `vn_ward_code` - Foreign key to wards_tmp

### GIS Tables
- `gis_server_id` - GIS server identifier (e.g., "tinh34.7", "xa3321.3285")
- `bbox_wkt` - Bounding box in WKT POLYGON format
- `geom_wkt` - Geometry in WKT MULTIPOLYGON format

## Common Query Patterns

### Count records
```sql
SELECT COUNT(*) FROM table_name;
```

### Join provinces and wards
```sql
SELECT p.name as province, w.name as ward
FROM provinces_tmp p
JOIN wards_tmp w ON p.code = w.province_code
WHERE p.code = '01'
LIMIT 10;
```

### Check GIS data completeness
```sql
SELECT
  COUNT(*) as total,
  COUNT(bbox_wkt) as with_bbox,
  COUNT(geom_wkt) as with_geom
FROM sapnhap_provinces_gis;
```

### Find missing relationships
```sql
SELECT COUNT(*) FROM sapnhap_provinces
WHERE vn_province_code IS NULL;
```

## Tips

1. Use `\dt` to list all tables
2. Use `\d table_name` to see table structure
3. Use `LIMIT` to avoid large result sets
4. Use `EXPLAIN ANALYZE` for performance tuning
5. PostGIS functions available: `ST_AsText()`, `ST_Area()`, `ST_Contains()`, etc.

## User Instructions

When the user invokes this skill:
1. Execute their SQL query using the docker exec command
2. Format the results in a readable way (tables, lists, etc.)
3. Provide insights or follow-up suggestions if relevant
4. If the query returns an error, explain the issue and suggest fixes

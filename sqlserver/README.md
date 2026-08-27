# Microsoft SQL Server Dataset — Vietnamese Provinces Database

**Generated at: Thu, 27 Aug 2026 00:49:01 +0700**

Import script for the Vietnamese Provinces Database on Microsoft SQL Server.

## Files

- `mssql_ImportData_vn_units.sql` — INSERT statements for regions, units, provinces, and wards (359.24 KB)

## Overview

The Vietnamese Provinces Database for Microsoft SQL Server. The import script populates:

| Table | Rows | Description |
|-------|------|-------------|
| `administrative_regions` | 8 | Regions of Vietnam |
| `administrative_units` | 8 | Administrative unit types (city, province, ward, ...) |
| `provinces` | 34 | Provinces and municipalities |
| `wards` | 3,321 | Wards, communes, and town townships |

GIS boundary geometry (in `gis/`) populates `gis_provinces` and `gis_wards`.

## Data Structure

### provinces

| Column | Description |
|--------|-------------|
| `code` | Province code (PK, e.g. `01`) |
| `name` / `name_en` | Province name |
| `full_name` / `full_name_en` | Full name with unit prefix |
| `code_name` | Code name (e.g. `ha_noi`) |
| `administrative_unit_id` | FK to `administrative_units.id` |
| `postal_code_prefix` | Comma-separated 2-digit postal prefixes |

### wards

| Column | Description |
|--------|-------------|
| `code` | Ward code (PK, e.g. `00004`) |
| `name` / `name_en` | Ward name |
| `full_name` / `full_name_en` | Full name with unit prefix |
| `code_name` | Code name |
| `province_code` | FK to `provinces.code` |
| `administrative_unit_id` | FK to `administrative_units.id` |
| `postal_code` | 5-digit national postal code |

`administrative_regions` and `administrative_units` hold the region and unit-type lookup rows (8 each).

## Sample Document

A province row:

```sql
INSERT INTO provinces(code,name,name_en,full_name,full_name_en,code_name,administrative_unit_id,postal_code_prefix) VALUES('01',N'Hà Nội',N'Ha Noi',N'Thành phố Hà Nội',N'Ha Noi City','ha_noi',1,'10, 11, 12, 13, 14');
```

## Quick Start

1. Create the tables by running `mssql_CreateTables_vn_units.sql`.
2. Import the data with `sqlcmd -S <server> -d <db> -U <user> -P <pass> -i mssql_ImportData_vn_units.sql`.
3. Import the GIS add-on (optional): run each chunk in `gis/` in the order listed in the `.manifest` file.

## Sample Queries

```sql
SELECT (SELECT COUNT(*) FROM provinces) AS provinces, (SELECT COUNT(*) FROM wards) AS wards;

SELECT w.code, w.name FROM wards w WHERE w.province_code = '01' ORDER BY w.name;

-- GIS: province containing a point
SELECT p.code, p.name
FROM provinces p
JOIN gis_provinces g ON p.code = g.province_code
WHERE g.geom.STContains(geometry::STGeomFromText('POINT(105.8542 21.0285)', 4326)) = 1;
```

## GIS / GeoJSON

The `gis/` subfolder contains chunked GIS import scripts (`mssql_ImportData_gis-part-NN.sql`) plus a `.manifest` file listing the chunks in order. Import every chunk in order after the base import.

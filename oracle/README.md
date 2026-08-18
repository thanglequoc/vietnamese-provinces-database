# Oracle Dataset — Vietnamese Provinces Database

**Generated at: Wed, 19 Aug 2026 00:50:14 +0700**

Import script for the Vietnamese Provinces Database on Oracle.

## Files

- `oracle_ImportData_vn_units.sql` — INSERT ALL statements for regions, units, provinces, and wards (723.06 KB)

## Overview

The Vietnamese Provinces Database for Oracle. The import script populates:

| Table | Rows | Description |
|-------|------|-------------|
| `administrative_regions` | 8 | Regions of Vietnam |
| `administrative_units` | 8 | Administrative unit types (city, province, ward, ...) |
| `provinces` | 34 | Provinces and municipalities |
| `wards` | 3,321 | Wards, communes, and town townships |

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

A province row inside the multi-row `INSERT ALL` batch:

```sql
	INTO provinces(code,name,name_en,full_name,full_name_en,code_name,administrative_unit_id,postal_code_prefix) VALUES('01','Hà Nội','Ha Noi','Thành phố Hà Nội','Ha Noi City','ha_noi',1,'10, 11, 12, 13, 14')
```

## Quick Start

1. Create the tables by running `oracle_CreateTables_vn_units.sql`.
2. Import the data with `sqlplus <user>/<password>@<db> @oracle_ImportData_vn_units.sql`.

## Sample Queries

```sql
SELECT (SELECT COUNT(*) FROM provinces) AS provinces, (SELECT COUNT(*) FROM wards) AS wards FROM dual;

SELECT w.code, w.name FROM wards w WHERE w.province_code = '01' ORDER BY w.name;

-- Province by postal code prefix
SELECT p.name FROM provinces p WHERE p.postal_code_prefix LIKE '%11%';
```

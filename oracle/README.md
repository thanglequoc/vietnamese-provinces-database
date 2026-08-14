# Oracle Dataset — Vietnamese Provinces Database

**Generated at: Fri, 14 Aug 2026 08:42:28 +0700**

Import script for the Vietnamese Provinces Database on Oracle.

## Files

- `oracle_ImportData_vn_units.sql` — INSERT ALL statements for regions, units, provinces, and wards (723.06 KB)

## Data Structure

The import script populates: `administrative_regions` (8), `administrative_units` (8), `provinces` (34), and `wards` (3,321). Each province and ward carries postal code fields (`postal_code_prefix` / `postal_code`).

## Sample Queries

```sql
SELECT COUNT(*) FROM provinces;

SELECT w.code, w.name FROM wards w WHERE w.province_code = '01' ORDER BY w.name;
```

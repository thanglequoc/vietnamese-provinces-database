# PostgreSQL Dataset — Vietnamese Provinces Database

**Generated at: Fri, 14 Aug 2026 08:42:28 +0700**

Import script for the Vietnamese Provinces Database on PostgreSQL/PostGIS.

## Files

- `postgres_ImportData_vn_units.sql` — INSERT statements for regions, units, provinces, and wards (342.79 KB)

## Data Structure

The import script populates: `administrative_regions` (8), `administrative_units` (8), `provinces` (34), and `wards` (3,321). Each province and ward carries postal code fields (`postal_code_prefix` / `postal_code`).

GIS geometry (in `gis/`) populates `gis_provinces` and `gis_wards` with `bbox` and `geom` spatial columns.

## Sample Queries

```sql
SELECT COUNT(*) FROM provinces;

SELECT w.code, w.name FROM wards w WHERE w.province_code = '01' ORDER BY w.name;

-- GIS: province containing a point
SELECT p.code, p.name
FROM provinces p
JOIN gis_provinces g ON p.code = g.province_code
WHERE ST_Within(ST_SetSRID(ST_MakePoint(105.8542, 21.0285), 4326), g.geom);
```

## GIS / GeoJSON

The `gis/` subfolder contains chunked GIS import scripts (`PostgreSQL/PostGIS` part files `*-part-NN.sql`) plus a `.manifest` file listing the chunks in order. Import each chunk in order after importing the base script.

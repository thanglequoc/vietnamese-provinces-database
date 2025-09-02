## Example

```sql
DECLARE @pt geometry = geometry::STGeomFromText('POINT(106.74769171085045 10.86755559707941)', 4326);
SELECT ward_code, gis_server_id, area_km2
FROM gis_wards
WHERE geom.MakeValid().STContains(@pt) = 1;
```

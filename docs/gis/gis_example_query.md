# GIS Example Queries
## Common PostGIS Queries

This dataset is compatible with PostgreSQL + PostGIS and supports common spatial operations such as point-in-polygon lookup, area calculation, distance measurement, and GeoJSON export.

### Frequently Used Spatial Functions

| Function | Purpose | Common Use Cases |
|-----------|-----------|------------------|
| `ST_Contains()` | Checks whether a geometry contains another geometry | Find which province/ward contains a GPS coordinate |
| `ST_Within()` | Checks whether a geometry is inside another geometry | Reverse lookup of administrative boundaries |
| `ST_Intersects()` | Checks whether two geometries overlap | Spatial filtering and GIS analysis |
| `ST_Distance()` | Calculates distance between geometries | Nearest administrative unit search |
| `ST_Area()` | Calculates polygon area | Area statistics and reporting |
| `ST_Centroid()` | Returns the center point of a polygon | Display province/ward markers on maps |
| `ST_AsGeoJSON()` | Converts geometry into GeoJSON | Web map integration (Leaflet, OpenLayers, Mapbox) |
| `ST_X()` / `ST_Y()` | Extract longitude and latitude from a point | Coordinate export |
| `ST_Transform()` | Converts geometry between coordinate systems | Accurate area and distance calculations |
| `ST_Buffer()` | Creates a buffer zone around a geometry | Radius-based spatial search |

---

### Find Province/Ward by GPS Coordinate

Given a longitude and latitude, determine which province contains the point.

```sql
SELECT *
FROM gis_wards
WHERE bbox && ST_SetSRID(
        ST_Point(106.65735539347402, 10.767883941562623),
        4326
      )
AND ST_Contains(geom, ST_SetSRID(ST_Point(106.65735539347402, 10.767883941562623), 4326));
```

### Export Boundary as GeoJSON

Convert geometry into GeoJSON for use with web mapping libraries.

```sql
SELECT
    ward_code,
    ST_AsGeoJSON(geom) AS geojson
FROM gis_wards;
```

---

### Get Province Center Point

Retrieve the centroid of each province.

```sql
SELECT
    province_code,
    ST_Y(ST_Centroid(geom)) AS latitude,
    ST_X(ST_Centroid(geom)) AS longitude
FROM gis_provinces;
```

---

### Calculate Province Area

Calculate province area in square kilometers.

```sql
SELECT
    name,
    ROUND(
        ST_Area(
            ST_Transform(geom, 3857)
        ) / 1000000,
        2
    ) AS area_km2
FROM province;
```

---

### Find Provinces Intersecting a Geometry

Find provinces that intersect a given geometry.

```sql
SELECT
    name
FROM province
WHERE ST_Intersects(
    geom,
    :geometry
);
```

---

### Calculate Distance to Province Center

Measure the distance between a GPS coordinate and province centroids.

```sql
SELECT
    name,
    ROUND(
        ST_Distance(
            ST_Centroid(geom)::geography,
            ST_SetSRID(
                ST_Point(106.7009, 10.7769),
                4326
            )::geography
        ) / 1000,
        2
    ) AS distance_km
FROM province;
```

---

### Find Wards Within a Radius

Find wards within 5 km of a location.

```sql
SELECT
    ward_code
FROM gis_wards
WHERE ST_Intersects(
    geom,
    ST_Buffer(
        ST_SetSRID(
            ST_Point(106.7009, 10.7769),
            4326
        )::geography,
        5000
    )::geometry
);
```

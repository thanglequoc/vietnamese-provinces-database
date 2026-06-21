# GIS Dataset Add-on — Vietnamese Provinces Database
**Latest GIS dataset version: **v4.0.0** (June 20, 2026)**

|Platform|Download Link|File Size|
|--------|-------------|---------|
|PostgreSQL/PostGIS|[Download raw GIS Dataset for PostgreSQL][gis_dataset_postgresql_bucket_url]|~152.07 MB|
|MySQL|[Download raw GIS Dataset for MySQL][gis_dataset_mysql_bucket_url]|~150.44 MB|
|Microsoft SQL Server|[Download raw GIS Dataset for SQL Server][gis_dataset_sqlserver_bucket_url]|~152.13 MB|

## Table of Contents

1. [Introduction](#introduction)
2. [Dataset Coverage](#dataset-coverage)
3. [Installation](#installation)
4. [Database Schema](#database-schema)
5. [Query Optimization Tips](#query-optimization-tips)
6. [Version Compatibility](#version-compatibility)
7. [FAQ & Troubleshooting](#faq--troubleshooting)
8. [Additional Resources](#additional-resources)
9. [Contributing](#contributing)

---


## Introduction

The **GIS Dataset** is an optional add-on for the Vietnamese Provinces Database project that provides high-precision administrative boundary geometries for Vietnam's administrative units. This dataset contains detailed geographic boundaries for all 34 provinces and 3,321 communes/wards, enabling sophisticated geospatial queries and visualization.

GIS boundary data was derived from the [Vietnam Administrative Units Reference Map](https://sapnhap.bando.com.vn), published by the Vietnam Natural Resources, Environment and Cartography Publishing House under the Ministry of Agriculture and Environment.

### What is the GIS Dataset?

The GIS dataset consists of:
- **Geographic boundary geometries** stored as spatial data structures (Multipolygons)
- **Bounding boxes** for each administrative unit (useful for pre-filtering queries)
- **Synchronized administrative codes** linking back to the main Vietnamese Provinces Database

### Why Use It?
The GIS dataset enables:

* Map Visualization – Render province and commune boundaries in web, mobile, and desktop mapping applications.
* Administrative Boundary Lookup – Identify the administrative unit associated with a geographic location.
* Point-in-Polygon Queries – Determine which province or commune contains a given latitude/longitude coordinate.
* Spatial Analysis – Perform proximity, intersection, area, and other geospatial operations directly in the database.

---

## Dataset Coverage

### Geographic Extent

The dataset provides complete coverage of Vietnamese administrative divisions:
| Level | Count | Coverage |
|-------|-------|----------|
| **Provinces** | 34 | 100% (including Hà Nội, HCM City, and Đà Nẵng) |
| **Communes/Wards** | 3,321 | 100% (all administrative communes, wards, and town wards) |

### Coordinate Reference System

- **Standard:** WGS 84 (World Geodetic System 1984)
- **SRID:** 4326
- **Format:** Longitude, latitude (OGC standard)
- **Datum:** Global GPS standard

### Geometry Types

Each administrative unit is represented by two geometry objects:

1. **Bounding Box (`bbox`):**
   - Type: `Polygon`
   - Use case: Rapid pre-filtering for spatial queries
   - Performance: Faster computation than full geometry

2. **Full Boundary (`geom`):**
   - Type: `Multipolygon`
   - Use case: Precise boundary visualization and point-in-polygon operations
   - Coverage: Entire administrative area, including non-contiguous regions if applicable

### Core Tables

The dataset consists of two main tables:

1. **`gis_provinces`** — 34 province-level records with boundary geometries
2. **`gis_wards`** — 3,321 ward/commune-level records with boundary geometries

#### Relationship to Main Database

Administrative units in the GIS dataset are synchronized with the main Vietnamese Provinces Database using administrative codes:

- **Provinces:** `gis_provinces.province_code` field links to `provinces.code`
- **Wards:** `gis_wards.ward_code` field links to `wards.code`

This allows seamless joining of GIS boundaries with administrative unit metadata, names, and hierarchical relationships.  
[![image.png](https://i.postimg.cc/zBJQwyY1/image.png)](https://postimg.cc/nsPTpcLd)

---

## Installation

The GIS dataset requires bootstrap SQL scripts to create tables, and import scripts to load boundary data. Scripts are available for PostgreSQL, MySQL, and SQL Server.

### PostgreSQL + PostGIS

#### Prerequisites

- **PostgreSQL:** Version 12 or later
- **PostGIS:** Version 3.0 or later (required for spatial data types)
- **Disk space:** ~100 MB for all geometry data
- **Existing database:** The main Vietnamese Provinces Database must be installed first

#### Step 1: Enable PostGIS Extension

Connect to your PostgreSQL database and enable the PostGIS extension:

```sql
CREATE EXTENSION IF NOT EXISTS postgis;
-- Verify installation
SELECT postgis_version();
```

Expected output: PostGIS version information (e.g., `3.1.4`).

#### Step 2: Create GIS Tables (Bootstrap)

Execute the bootstrap script to create all GIS tables, indexes, and constraints:

```bash
psql -U <username> -d <database_name> -f postgresql/gis/postgresql_CreateGISTables.sql
```

#### Step 3: Import GIS Data

Unzip the Postg GIS data archive at [postgresql/gis](../../postgresql/gis/)
Or download the raw GIS dataset file directly from [vn-province bucket][gis_dataset_postgresql_bucket_url]

Execute the data import script to populate boundaries:

```bash
psql -U <username> -d <database_name> -f postgresql/gis/postgresql_ImportData_gis_2026-06-20__12_32_01.sql
```

### MySQL / MariaDB

#### Prerequisites

- **MySQL:** Version 8.0 or later
- **MariaDB:** Version 10.4 or later (with spatial support)
- **Disk space:** ~120 MB for all geometry data
- **Existing database:** The main Vietnamese Provinces Database must be installed first

#### Step 1: Create GIS Tables (Bootstrap)

Execute the bootstrap script to create all GIS tables, indexes, and constraints:

```bash
mysql -u <username> -p <database_name> < mysql/gis/mysql_CreateGISTables.sql
```

#### Step 2: Import GIS Data

Unzip the MySQL GIS data archive at [mysql/gis](../../mysql/gis/)
Or download the raw GIS dataset file directly from [vn-province bucket][gis_dataset_mysql_bucket_url]

Execute the data import script:

```bash
mysql -u <username> -p <database_name> < mysql/gis/mysql_ImportData_gis_2026-06-20__12_32_01.sql
```

#### MySQL Specific Notes

- MySQL uses **SPATIAL INDEX** (also displayed as `SPATIAL KEY`) for geometry columns instead of PostgreSQL's GiST indexes.
- Geometry data is stored using native spatial types such as `MULTIPOLYGON` with SRID 4326.
- Spatial functions follow the OGC `ST_*` naming convention (for example, `ST_Contains`, `ST_Intersects`, and `ST_Distance`).
- See [Coordinate Handling Across Databases](#coordinate-handling-across-databases) for details on coordinate ordering and SRID handling.

---

### Microsoft SQL Server

#### Prerequisites

- **SQL Server:** Version 2019 or later
- **Disk space:** ~110 MB for all geometry data
- **Existing database:** The main Vietnamese Provinces Database must be installed first

#### Step 1: Create GIS Tables (Bootstrap)

Execute the bootstrap script to create all GIS tables, indexes, and constraints:

```cmd
sqlcmd -S <server_name> -d <database_name> -U <username> -P <password> -i sqlserver/gis/mssql_CreateGISTables.sql
```

#### Step 2: Import GIS Data

Unzip the SQL Server GIS data archive at [sqlserver/gis](../../sqlserver/gis/)
Or download the raw GIS dataset file directly from [vn-province bucket][gis_dataset_sqlserver_bucket_url]

Execute the data import script:

```cmd
sqlcmd -S <server_name> -d <database_name> -U <username> -P <password> -i sqlserver/gis/mssql_ImportData_gis_2026-06-20__12_32_02.sql
```

#### SQL Server Specific Notes

- Geometry data is stored using SQL Server's native `geometry` type.  
- Geometries are typically created using functions such as `geometry::STGeomFromText()` and `geometry::Point()`.  
- Spatial operations use SQL Server's instance-method syntax (for example, `geom.STContains()` and `geom.STDistance()`).  
- Spatial indexes are implemented using SQL Server's native Spatial Index feature.  

---

## Database Schema

### gis_provinces

Stores GIS boundary data for all province-level administrative units.

| Column | Type | Nullable | Description |
|----------|----------|----------|----------|
| `id` | `INTEGER` | No | Surrogate primary key for the GIS record. |
| `province_code` | `VARCHAR(20)` | No | Administrative province code used to join with the `provinces` table. |
| `gis_server_id` | `VARCHAR(50)` | Yes | Original GIS source identifier for the province boundary. |
| `area_km2` | `NUMERIC(12,5)` | Yes | Province area in square kilometers. |
| `bbox` | `geometry(Polygon, 4326)` | Yes | Bounding box geometry used for spatial pre-filtering and index optimization. |
| `geom` | `geometry(MultiPolygon, 4326)` | Yes | Full province boundary geometry used for visualization and spatial analysis. |

**Primary Key:** `id`  
**Foreign Key:** `province_code` → `provinces.code`

---

### gis_wards

Stores GIS boundary data for all commune-level administrative units.

| Column | Type | Nullable | Description |
|----------|----------|----------|----------|
| `id` | `INTEGER` | No | Surrogate primary key for the GIS record. |
| `ward_code` | `VARCHAR(20)` | No | Administrative ward/commune code used to join with the `wards` table. |
| `gis_server_id` | `VARCHAR(50)` | Yes | Original GIS source identifier for the administrative unit boundary. |
| `area_km2` | `NUMERIC(12,5)` | Yes | Administrative unit area in square kilometers. |
| `bbox` | `geometry(Polygon, 4326)` | Yes | Bounding box geometry used for spatial pre-filtering and index optimization. |
| `geom` | `geometry(MultiPolygon, 4326)` | Yes | Full administrative boundary geometry used for visualization and spatial analysis. |

**Primary Key:** `id`  
**Foreign Key:** `ward_code` → `wards.code`

---

## Query Optimization Tips
#### Use Bounding Box Pre-filtering for Large Spatial Queries

The dataset stores both a full boundary geometry (`geom`) and a simplified bounding box (`bbox`).
For large-area searches, filtering with `bbox` first can reduce the number of geometries that require expensive spatial calculations.

```sql
-- Example: Find which ward that Phu Tho Horse Racing Ground is located in (coordinates: 106.65735539347402, 10.767883941562623)
SELECT *
FROM gis_wards
WHERE bbox && ST_SetSRID(
        ST_Point(106.65735539347402, 10.767883941562623),
        4326
      )
AND ST_Contains(geom, ST_SetSRID(ST_Point(106.65735539347402, 10.767883941562623), 4326));
```

---

## Version Compatibility

The GIS dataset is tested and supported on the following platform versions:

| Component | Version | Status | Notes |
|-----------|---------|--------|-------|
| **PostgreSQL** | 12+ | ✓ Fully Supported | Recommended: 14+ for better performance |
| **PostGIS** | 3.0+ | ✓ Required | Essential for spatial data types and functions |
| **MySQL** | 8.0+ | ✓ Fully Supported | MariaDB 10.4+ also supported |
| **MariaDB** | 10.4+ | ✓ Fully Supported | Equivalent functionality to MySQL 8.0 |
| **SQL Server** | 2019+ | ✓ Fully Supported | Includes built-in spatial type support |
| **WGS 84 (SRID 4326)** | ISO 19115 | ✓ Standard | All platforms support this projection |

### Minimum Requirements

- **PostgreSQL:** 12 (released 2019)
- **PostGIS:** 3.0 (released 2019)
- **MySQL:** 8.0 (released 2018)
- **SQL Server:** 2019 (released 2019)

### Recommended Versions for Optimal Performance

- **PostgreSQL:** 15+ (latest stable, better parallelization)
- **PostGIS:** 3.3+ (performance improvements)
- **MySQL:** 8.0.30+ (latest 8.0 release)
- **SQL Server:** 2022 (latest version)

---

## FAQ & Troubleshooting

### Q: Do I need to install the GIS dataset?

**A:** Only if you need geographic boundary data (mapping, spatial queries, geocoding). The main Vietnamese Provinces Database works independently for administrative unit lookups, hierarchies, and metadata queries.

### Q: What if I get "PostGIS extension not found"?

**A:** PostGIS is not installed. For PostgreSQL:

```sql
CREATE EXTENSION postgis;
-- or
sudo apt-get install postgresql-14-postgis  -- Ubuntu/Debian
brew install postgis                        -- macOS
```

### Q: How do I export boundaries to GeoJSON?

**PostgreSQL / PostGIS:**

```sql
SELECT 
    id,
    province_code,
    ST_AsGeoJSON(geom) AS geojson
FROM gis_provinces;
```

Export the result and parse the `geojson` column to standard GeoJSON format.
**There will be a dedicated GeoJSON dataset in the future releases**.

### Q: Can I use the GIS data with mapping libraries (Mapbox, Leaflet)?

**A:** Yes. The GIS dataset can be used with popular mapping libraries such as Leaflet, Mapbox GL JS, OpenLayers, and ArcGIS.

The recommended approach is to export geometries as **GeoJSON** from the database and expose them through an API endpoint:

```javascript
fetch('/api/provinces/01/boundary')
    .then(response => response.json())
    .then(geojson => {
        L.geoJSON(geojson).addTo(map);
    });
```

Typical workflow:

1. Query the boundary geometry from the GIS dataset.
2. Convert the geometry to GeoJSON (for example, using `ST_AsGeoJSON()` in PostGIS).
3. Return the GeoJSON from your backend API.
4. Render the GeoJSON in your mapping library.

> GeoJSON is the recommended interchange format for web mapping applications.

### Q: How often are boundaries updated?

**A:** Boundaries are updated when the Vietnamese government issues administrative decrees (typically 2-4 per year). Subscribe to the project repository for updates.

### Q: Is there any GUI recommendation for visualizing the GIS data?

**A**: Yes. The following tools work well with the GIS dataset:

|Tool	Type	Recommendation
|------|--------|---------------|
|DBeaver|	Database Client|	Recommended for most users. Supports PostgreSQL, MySQL, and SQL Server, and can preview geometry data directly on an OpenStreetMap-based map viewer.
|QGIS|	Desktop GIS|	Best for advanced GIS analysis, spatial editing, and cartographic visualization.
|geojson.io|	Web Viewer|	Lightweight online tool for visualizing and debugging GeoJSON exported from the database.

Example: Visualize Geometry in DBeaver
[![image.png](https://i.postimg.cc/dVmQJSFg/image.png)](https://postimg.cc/k2GPcsBy)

Example: Export GeoJSON for geojson.io

SELECT
    province_code,
    ST_AsGeoJSON(geom) AS geojson
FROM gis_provinces;

Copy the GeoJSON output and paste it into https://geojson.io/ to visualize the administrative boundaries interactively.

For most users of this project, [DBeaver](https://dbeaver.io/) is the recommended option because it is free, cross-platform, and supports PostgreSQL, MySQL, and SQL Server.

---

## Additional Resources

- [PostGIS Documentation](https://postgis.net/documentation/)
- [MySQL Spatial Data Types](https://dev.mysql.com/doc/refman/8.0/en/spatial-types.html)
- [MySQL Spatial Analysis Functions](https://dev.mysql.com/doc/refman/8.4/en/spatial-analysis-functions.html)
- [SQL Server Spatial Data](https://docs.microsoft.com/en-us/sql/relational-databases/spatial/spatial-data-sql-server)
- [GIS Concepts & Standards](https://en.wikipedia.org/wiki/Geographic_information_system)
- [WGS 84 Reference System](https://epsg.io/4326)

---

## Contributing

If you find issues with the GIS dataset or have suggestions for improvements, please open an issue in the project repository.

---

**Last Updated:** June 20, 2026

[gis_dataset_postgresql_bucket_url]: https://vn-provinces-ds.thanglequoc.xyz/v4.0.0/GISDataSet/postgresql_ImportData_gis_2026-06-20__12_32_01.sql
[gis_dataset_mysql_bucket_url]: https://vn-provinces-ds.thanglequoc.xyz/v4.0.0/GISDataSet/mysql_ImportData_gis_2026-06-20__12_32_01.sql
[gis_dataset_sqlserver_bucket_url]: https://vn-provinces-ds.thanglequoc.xyz/v4.0.0/GISDataSet/mssql_ImportData_gis_2026-06-20__12_32_02.sql

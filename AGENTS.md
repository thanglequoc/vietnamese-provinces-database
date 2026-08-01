# AI Agent Instructions — Vietnamese Provinces Database

A comprehensive database of Vietnamese administrative units (34 provinces, districts, wards) maintained through automated Go scripts that generate SQL, JSON, and GIS data for multiple database engines.

---

## Quick Start for AI Agents

### Building & Running

```bash
cd dataset-generation-scripts

# Run the generation scripts
go run main.go

# Start Postgres/PostGIS for integration tests
docker compose -f docker/docker-compose.yaml up -d

# Run tests
go test -v ./...

# Check output in: ./output/
```

**Testing note**: Some tests connect to the temporary Postgres/PostGIS database on `localhost:15432`. If Docker is not running, `go test -v ./...` will fail in integration-style packages.

## Database Query Skill

**CRITICAL: Proactively query the database whenever the task involves data verification, counts, searching, or GIS data. Do NOT wait for explicit command invocation — the triggers below signal when to auto-query.**

### Auto-Trigger Rules

Execute database queries automatically when user asks about any of these:

| Category | Trigger Keywords |
|----------|-----------------|
| Count/total | "how many", "count", "total", "number of", "how much" |
| Search/find | "find", "search", "show", "list", "get", "lookup" |
| Verification | "check", "verify", "validate", "missing", "orphaned" |
| Data topics | "database", "table", "data", "records", "provinces", "wards", "schema" |
| GIS | "geometry", "bbox", "geom", "gis", "spatial", "geojson", "coordinates" |
| Direct requests | "query from", "read from database", "get from [table]", "run a query" |

**Examples:**
- "How many wards are in Hà Nội?" → Run query immediately
- "Check if there are any missing GIS data" → Run verification query
- "Show me provinces without codes" → Query and display results
- "List all tables" → Run `\dt`
- "Get data from sapnhap_geojson_objects" → Query the table

### Connection Details

The temporary database runs in Docker:

```bash
# Container: vn_provinces_postgres_container
# Database: vn_provinces_tmp
# Username: postgres
# Host:     localhost
# Port:     15432

# Single query
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp -c "YOUR_SQL_QUERY_HERE"

# Multi-line query (heredoc)
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp <<'EOF'
YOUR_MULTI_LINE_SQL_QUERY_HERE
EOF
```

### Table Reference

| Table | Records | Description |
|-------|---------|-------------|
| `provinces_tmp` | 34 | Vietnam provinces (code, name, name_en, full_name, administrative_unit_id) |
| `wards_tmp` | 3,321 | Vietnam wards (code, name, name_en, province_code FK, administrative_unit_id) |
| `administrative_regions` | 8 | Region lookup (id, name, name_en) |
| `administrative_units` | 8 | Unit type lookup (id, full_name, short_name) |
| `sapnhap_geojson_objects` | 3,355 | Combined geo objects with PostGIS geometry (ma, ten, magoc, malk, truocsapnhap, dientichkm2, bbox_wkt, geom_wkt, vn_ds_province_code FK, vn_ds_ward_code FK) |

### Key Columns

**`provinces_tmp`**: `code` (PK, e.g. "01", "02"), `name`, `name_en`, `full_name`, `code_name`, `administrative_unit_id`

**`wards_tmp`**: `code` (PK), `name`, `name_en`, `full_name`, `province_code` (FK → `provinces_tmp.code`), `administrative_unit_id`

**`sapnhap_geojson_objects`**: `ma` (PK), `ten` (name), `magoc` (parent FK, self-ref), `truocsapnhap` (pre-merge name), `dientichkm2` (area in km²), `bbox_wkt` (WKT POLYGON), `geom_wkt` (WKT MULTIPOLYGON), `vn_ds_province_code` (FK → `provinces_tmp.code`), `vn_ds_ward_code` (FK → `wards_tmp.code`)

### Common Query Patterns

**Count records:**
```sql
SELECT COUNT(*) FROM provinces_tmp;
SELECT COUNT(*) FROM wards_tmp;
```

**Data completeness check:**
```sql
SELECT
  (SELECT COUNT(*) FROM provinces_tmp) as provinces,
  (SELECT COUNT(*) FROM wards_tmp) as wards,
  (SELECT COUNT(*) FROM sapnhap_geojson_objects) as geo_objects;
```

**Join provinces and wards:**
```sql
SELECT p.name as province, w.name as ward
FROM provinces_tmp p
JOIN wards_tmp w ON p.code = w.province_code
WHERE p.code = '01'
LIMIT 10;
```

**GIS geometry completeness:**
```sql
SELECT
  COUNT(*) as total,
  COUNT(bbox_wkt) as with_bbox,
  COUNT(geom_wkt) as with_geom
FROM sapnhap_geojson_objects;
```

**Find orphaned records:**
```sql
-- Geo objects without matching province
SELECT COUNT(*) FROM sapnhap_geojson_objects
WHERE vn_ds_province_code IS NULL;

-- Geo objects without matching ward
SELECT COUNT(*) FROM sapnhap_geojson_objects
WHERE vn_ds_ward_code IS NULL;
```

**List wards in a specific province:**
```sql
SELECT code, name, name_en FROM wards_tmp
WHERE province_code = '01'
ORDER BY name;
```

**PostGIS spatial queries:**
```sql
-- Get area of a province geometry
SELECT ten, ST_Area(geom::geography) / 1000000 as area_km2
FROM sapnhap_geojson_objects
WHERE vn_ds_province_code = '01';

-- Find objects within a bounding box (example: Hà Nội area)
SELECT ma, ten, ST_AsText(bbox) FROM sapnhap_geojson_objects
WHERE ST_Within(geom, ST_MakeEnvelope(105.5, 20.5, 106.0, 21.5, 4326));
```

### Agent Instructions

When context triggers a database query:
1. Execute the SQL query using the `docker exec` command pattern above
2. Format results in a readable way (tables, lists, summaries)
3. Provide insights or follow-up suggestions if relevant
4. If the query returns an error, explain the issue and suggest a fix
5. Use `LIMIT` for large result sets to avoid overwhelming output
6. Explore schema with: `\dt` (list tables), `\d table_name` (describe table)

---

## ElasticSearch Query Skill

**CRITICAL: Proactively query Elasticsearch whenever the task involves searching provinces/wards data, cluster management, bulk imports, or GIS spatial queries. Do NOT wait for explicit command invocation — the triggers below signal when to auto-query.**

### Auto-Trigger Rules

Execute Elasticsearch queries automatically when user asks about any of these:

| Category | Trigger Keywords |
|----------|-----------------|
| Search/find | "search", "find", "full-text", "autocomplete", "lookup in ES", "query ES" |
| Count/total | "how many in elastic", "es count", "count documents" |
| Management | "create index", "delete index", "put mapping", "bulk import", "cluster health", "index exists" |
| Data topics | "provinces index", "provinces-gis", "elasticsearch", "elastic search", "ndjson" |
| GIS queries | "geo shape", "geo point", "spatial query", "find province by point", "find by coordinates" |
| Direct requests | "query elasticsearch", "run ES", "curl ES", "search ES" |

**Examples:**
- "Search for provinces matching 'ha noi' in Elasticsearch" → Run search query immediately
- "How many documents are in the provinces index?" → Run `_count` query
- "Check ES cluster health" → Run `_cluster/health`
- "Bulk import the provinces NDJSON" → Run `_bulk` with the file
- "Find which province contains coordinates 105.85, 21.03" → Run geo_shape query on provinces-gis
- "Show me the mapping for the provinces index" → Run `_mapping`

### Connection & Tunnel Management

Elasticsearch runs in Docker on the OVHCloud VPS and is accessible via SSH tunnel:

```bash
# Check if Elasticsearch is reachable through the tunnel
curl -s -o /dev/null -w "%{http_code}" localhost:9200

# Start the SSH tunnel if unreachable (returns non-200)
ssh -f -N -L 9200:localhost:9200 -i ~/.ssh/id_ed25519 thanglequoc@machine.thanglequoc.xyz
```

| Detail | Value |
|--------|-------|
| Base URL | `http://localhost:9200` |
| Authentication | None (open on localhost via tunnel) |
| SSH Host | `machine.thanglequoc.xyz` |
| SSH Port | `22` |
| SSH User | `thanglequoc` |
| SSH Key | `~/.ssh/id_ed25519` |
| JSON Formatter | `jq` (fallback: `python3 -m json.tool`) |

### Schema Reference

**Indices:**

| Index | Documents | Description |
|-------|-----------|-------------|
| `provinces` | 34 | Province documents with embedded wards, administrative units, and search keywords (no GIS geometry) |
| `provinces-gis` | 34 | Same structure as provinces + GIS fields: Center (geo_point), BoundingBox, Geometry (geo_shape) |

**Key Fields (both indices):**

| Field | Type | Description |
|-------|------|-------------|
| `Code` | keyword | Province code (e.g., "01", "02") |
| `Name` | text + keyword | Vietnamese name |
| `NameEn` | text + keyword | English name |
| `FullName` | text | Full Vietnamese name with administrative prefix |
| `FullNameEn` | text | Full English name with administrative prefix |
| `CodeName` | keyword | Snake-case code name (e.g., "ha_noi", "ho_chi_minh") |
| `AdministrativeUnit` | object | Embedded unit: Id, FullName, FullNameEn, ShortName, ShortNameEn, CodeName, CodeNameEn |
| `SearchKeywords` | keyword[] | Pre-computed autocomplete array (code, tone-stripped Vietnamese name, English name, codeName) |
| `Wards` | nested | Array of ward documents (same fields as province: Code, Name, NameEn, FullName, FullNameEn, CodeName, AdministrativeUnit, SearchKeywords) |
| `Meta` | object | DatasetVersion, AdministrativeRevision, GeneratedAt |
| `GIS` | object | **(provinces-gis only)** Center (geo_point: lat/lon), BoundingBox (MinLongitude, MinLatitude, MaxLongitude, MaxLatitude), Geometry (geo_shape, GeoJSON MultiPolygon) |

**Ward fields** (inside `Wards[]`): `Code`, `Name`, `NameEn`, `FullName`, `FullNameEn`, `CodeName`, `AdministrativeUnit`, `SearchKeywords`; plus `GIS` in provinces-gis.

**Mapping files**: `elasticsearch/mappings/provinces.json` and `elasticsearch/mappings/provinces-gis.json`

### Common Query Patterns

**Dataset queries:**

Count documents:
```bash
curl -s "localhost:9200/provinces/_count" | jq .
curl -s "localhost:9200/provinces-gis/_count" | jq .
```

Search by keyword (autocomplete):
```bash
curl -s "localhost:9200/provinces/_search" \
  -H 'Content-Type: application/json' \
  -d '{"query": {"terms": {"SearchKeywords": ["ha noi"]}}, "_source": ["Code", "Name", "NameEn"]}' | jq .
```

Province dropdown (all provinces sorted by code):
```bash
curl -s "localhost:9200/provinces/_search" \
  -H 'Content-Type: application/json' \
  -d '{"size": 34, "sort": [{"Code": "asc"}], "_source": ["Code", "Name", "NameEn"]}' | jq .
```

Nested ward search (find which province contains a specific ward):
```bash
curl -s "localhost:9200/provinces/_search" \
  -H 'Content-Type: application/json' \
  -d '{"query": {"nested": {"path": "Wards", "query": {"match": {"Wards.FullName": "Ba Đình"}}, "inner_hits": {}}}, "_source": ["Code", "Name"]}' | jq .
```

Full-text search on province names:
```bash
curl -s "localhost:9200/provinces/_search" \
  -H 'Content-Type: application/json' \
  -d '{"query": {"multi_match": {"query": "hanoi", "fields": ["Name", "NameEn", "FullName", "FullNameEn"]}}, "_source": ["Code", "Name", "NameEn"]}' | jq .
```

Get a single province by code:
```bash
curl -s "localhost:9200/provinces/_doc/01" | jq .
```

**GIS spatial queries (provinces-gis index):**

Find province containing a point:
```bash
curl -s "localhost:9200/provinces-gis/_search" \
  -H 'Content-Type: application/json' \
  -d '{"query": {"geo_shape": {"GIS.Geometry": {"shape": {"type": "point", "coordinates": [105.8542, 21.0285]}, "relation": "intersects"}}}, "_source": ["Code", "Name"]}' | jq .
```

Find provinces within a bounding box:
```bash
curl -s "localhost:9200/provinces-gis/_search" \
  -H 'Content-Type: application/json' \
  -d '{"query": {"geo_bounding_box": {"GIS.Center": {"top_left": {"lat": 21.5, "lon": 105.5}, "bottom_right": {"lat": 20.5, "lon": 106.5}}}}, "_source": ["Code", "Name"]}' | jq .
```

**Management:**

Cluster health:
```bash
curl -s "localhost:9200/_cluster/health" | jq .
```

List all indices:
```bash
curl -s "localhost:9200/_cat/indices?v"
```

Check if an index exists:
```bash
curl -s -o /dev/null -w "%{http_code}" "localhost:9200/provinces"
```

Create index from mapping file:
```bash
curl -X PUT "localhost:9200/provinces" \
  -H 'Content-Type: application/json' \
  -d @elasticsearch/mappings/provinces.json
```

Bulk import NDJSON data:
```bash
curl -X POST "localhost:9200/_bulk" \
  -H 'Content-Type: application/x-ndjson' \
  --data-binary @elasticsearch/provinces.ndjson
```

Delete an index:
```bash
curl -X DELETE "localhost:9200/provinces"
```

Get index mapping:
```bash
curl -s "localhost:9200/provinces/_mapping" | jq .
```

Refresh an index (make recent writes visible):
```bash
curl -X POST "localhost:9200/provinces/_refresh"
```

### Agent Instructions

When context triggers an Elasticsearch query:
1. Check tunnel reachability first: `curl -s -o /dev/null -w "%{http_code}" localhost:9200`
2. If return code is not 200, start the tunnel: `ssh -f -N -L 9200:localhost:9200 -i ~/.ssh/id_ed25519 thanglequoc@machine.thanglequoc.xyz`
3. Execute the curl command with `-s` (silent mode) and pipe through `jq` for readable JSON output
4. Format results in tables, lists, or key-value summaries — not raw JSON dumps
5. Use `_count` before `_search` when checking document totals to avoid returning full documents
6. For large result sets, add `"size": 5` (or appropriate small number) to avoid overwhelming output
7. Use `_source` filtering in search requests to return only needed fields
8. Fall back to `python3 -m json.tool` if `jq` is not installed on the system
9. If Elasticsearch returns an error (non-2xx status code), explain the issue clearly and suggest a fix
10. If the SSH tunnel fails to connect, report the error and suggest: verifying the bastion is online, checking SSH key permissions (`chmod 600 ~/.ssh/id_ed25519`), or running `ssh -v thanglequoc@machine.thanglequoc.xyz` to debug

---

## MongoDB Query Skill

**CRITICAL: Proactively query MongoDB whenever the task involves searching provinces/wards data, collection management, data imports, or GIS spatial queries. Do NOT wait for explicit command invocation — the triggers below signal when to auto-query.**

### Auto-Trigger Rules

Execute MongoDB queries automatically when user asks about any of these:

| Category | Trigger Keywords |
|----------|-----------------|
| Search/find | "search", "find", "lookup in mongo", "query mongo", "findOne", "aggregate" |
| Count/total | "how many in mongo", "mongo count", "count documents", "count collections" |
| Management | "import mongo", "mongoimport", "create index", "drop collection", "show collections", "create collection" |
| Data topics | "mongo provinces", "provinces-gis", "wards-gis", "mongodb", "mongo document" |
| GIS queries | "geoIntersects", "near", "nearSphere", "geoWithin", "2dsphere", "point in mongo", "coordinates in mongo" |
| Direct requests | "query mongodb", "run mongo query", "check mongo", "get from mongo" |

**Examples:**
- "How many documents are in the provinces-gis collection?" → Run `db.provinces_gis.countDocuments()` immediately
- "Find which province contains coordinates 105.85, 21.03" → Run `$geoIntersects` query on provinces-gis
- "Show me all collections in MongoDB" → Run `show collections`
- "Import the provinces NDJSON" → Run `mongoimport` command
- "Count wards in province 01" → Run `db.wards_gis.countDocuments({ProvinceCode: "01"})`

### Connection & Tunnel Management

MongoDB runs in Docker on the OVHCloud VPS and is accessible via SSH tunnel:

```bash
# Check if MongoDB is reachable through the tunnel
echo 'db.version()' | mongosh "mongodb://root:Q35iSs8h5Y47VMcxZ5UC@localhost:27017/?authSource=admin&directConnection=true" --quiet 2>/dev/null || echo "unreachable"

# Start the SSH tunnel if unreachable
ssh -f -N -L 27017:localhost:27017 -i ~/.ssh/id_ed25519 thanglequoc@machine.thanglequoc.xyz
```

| Detail | Value |
|--------|-------|
| Connection String | `mongodb://root:Q35iSs8h5Y47VMcxZ5UC@localhost:27017/?authSource=admin&directConnection=true` |
| Container | `mongodb` (on VPS, managed by `docker compose -f /home/thanglequoc/docker/database/vietnamese-provinces-docker-compose.yaml`) |
| Authentication | Root user (username: `root`, password: `Q35iSs8h5Y47VMcxZ5UC`) |
| SSH Host | `machine.thanglequoc.xyz` |
| SSH Port | `22` |
| SSH User | `thanglequoc` |
| SSH Key | `~/.ssh/id_ed25519` |
| Default Database | `vn_provinces` |
| JSON Formatter | `EJSON.stringify()` or `tojson()` for single documents; pipe through `jq` where applicable |

### Collection Reference

| Collection | Documents | Description |
|------------|-----------|-------------|
| `provinces` | 34 | Standard province documents with embedded wards, administrative units, and search keywords (no GIS geometry) |
| `provinces-gis` | 34 | Province documents with province-level GIS: Center (GeoJSON Point), BoundingBox, Geometry (GeoJSON MultiPolygon), Properties — no embedded wards |
| `wards-gis` | 3,321 | Standalone ward documents with ward-level GIS: Center (GeoJSON Point), BoundingBox, Geometry (GeoJSON Polygon), Properties — includes `ProvinceCode` for cross-collection joins |

**Standard index names** (using dots for nested fields, hyphen for collection name separator):
- `provinces-gis`: `Code_1` (unique), `GIS.Geometry_2dsphere`, `GIS.Center_2dsphere`, `SearchKeywords_1`
- `wards-gis`: `Code_1` (unique), `ProvinceCode_1`, `GIS.Geometry_2dsphere`, `GIS.Center_2dsphere`, `SearchKeywords_1`

### Data Import Workflow

MongoDB datasets are generated by `go run main.go` and output to `dataset-generation-scripts/output/mongodb/`. Import them to the VPS MongoDB instance:

```bash
# 1. Ensure SSH tunnel is active
ssh -f -N -L 27017:localhost:27017 -i ~/.ssh/id_ed25519 thanglequoc@machine.thanglequoc.xyz

# 2. Import province GIS data (single file)
mongoimport --uri="mongodb://root:Q35iSs8h5Y47VMcxZ5UC@localhost:27017/?authSource=admin" \
  --db vn_provinces --collection provinces-gis \
  --file dataset-generation-scripts/output/mongodb/mongo_data_vn_province_gis.json \
  --jsonArray --drop

# 3. Import ward GIS data (may be chunked into parts)
# Single file:
mongoimport --uri="mongodb://root:Q35iSs8h5Y47VMcxZ5UC@localhost:27017/?authSource=admin" \
  --db vn_provinces --collection wards-gis \
  --file dataset-generation-scripts/output/mongodb/mongo_data_vn_ward_gis.json \
  --jsonArray --drop

# Chunked files (if ward data exceeds 50MB per file):
mongoimport --uri="mongodb://root:Q35iSs8h5Y47VMcxZ5UC@localhost:27017/?authSource=admin" \
  --db vn_provinces --collection wards-gis \
  --file dataset-generation-scripts/output/mongodb/mongo_data_vn_ward_gis_part_01.json \
  --jsonArray
# Repeat for part_02, part_03, etc. (omit --drop after the first import)

# 4. Create indexes (or run the create_indexes.js script)
mongosh "mongodb://root:Q35iSs8h5Y47VMcxZ5UC@localhost:27017/?authSource=admin" \
  --file dataset-generation-scripts/output/mongodb/create_indexes.js
```

### Common Query Patterns

All queries use `mongosh` through the SSH tunnel. Pipe through `--eval` for one-liners:

**Via tunnel (primary method — local mongosh required):**
```bash
# Count documents
mongosh "mongodb://root:Q35iSs8h5Y47VMcxZ5UC@localhost:27017/?authSource=admin" \
  --quiet --eval "use('vn_provinces'); db.getCollection('provinces-gis').countDocuments()"

# Find one province by code
mongosh "mongodb://root:Q35iSs8h5Y47VMcxZ5UC@localhost:27017/?authSource=admin" \
  --quiet --eval "use('vn_provinces'); EJSON.stringify(db.getCollection('provinces-gis').findOne({Code: '01'}))"
```

**Via docker exec on VPS (fallback — no SSH tunnel needed):**
```bash
ssh -i ~/.ssh/id_ed25519 thanglequoc@machine.thanglequoc.xyz \
  "docker exec mongodb mongosh -u root -p 'Q35iSs8h5Y47VMcxZ5UC' --quiet --eval \"use('vn_provinces'); db.getCollection('provinces-gis').countDocuments()\""
```

**Count documents:**
```bash
# Via tunnel
mongosh "mongodb://root:Q35iSs8h5Y47VMcxZ5UC@localhost:27017/?authSource=admin" \
  --quiet --eval "use('vn_provinces'); db.getCollection('provinces-gis').countDocuments()"
mongosh "mongodb://root:Q35iSs8h5Y47VMcxZ5UC@localhost:27017/?authSource=admin" \
  --quiet --eval "use('vn_provinces'); db.getCollection('wards-gis').countDocuments()"
```

**Search by keyword (autocomplete):**
```bash
mongosh "mongodb://root:Q35iSs8h5Y47VMcxZ5UC@localhost:27017/?authSource=admin" \
  --quiet --eval "use('vn_provinces'); EJSON.stringify(db.getCollection('provinces-gis').find({SearchKeywords: 'ha noi'}, {Code: 1, Name: 1, NameEn: 1}).toArray())"
```

**All provinces sorted by Code:**
```bash
mongosh "mongodb://root:Q35iSs8h5Y47VMcxZ5UC@localhost:27017/?authSource=admin" \
  --quiet --eval "use('vn_provinces'); EJSON.stringify(db.getCollection('provinces-gis').find({}, {Code: 1, Name: 1, NameEn: 1}).sort({Code: 1}).toArray())"
```

**GIS spatial query — find province containing a point:**
```bash
mongosh "mongodb://root:Q35iSs8h5Y47VMcxZ5UC@localhost:27017/?authSource=admin" \
  --quiet --eval "use('vn_provinces'); EJSON.stringify(db.getCollection('provinces-gis').find({ 'GIS.Geometry': { \$geoIntersects: { \$geometry: { type: 'Point', coordinates: [105.8542, 21.0285] } } } }, {Code: 1, Name: 1}).toArray())"
```

**GIS spatial query — find wards within a province:**
```bash
mongosh "mongodb://root:Q35iSs8h5Y47VMcxZ5UC@localhost:27017/?authSource=admin" \
  --quiet --eval "use('vn_provinces'); JSON.stringify(db.getCollection('wards-gis').find({ProvinceCode: '01'}, {Code: 1, Name: 1}).limit(10).toArray())"
```

**Cross-collection join — province with its wards:**
```bash
mongosh "mongodb://root:Q35iSs8h5Y47VMcxZ5UC@localhost:27017/?authSource=admin" \
  --quiet --eval "use('vn_provinces'); EJSON.stringify(db.getCollection('provinces-gis').aggregate([{ \$match: { Code: '01' } }, { \$lookup: { from: 'wards-gis', localField: 'Code', foreignField: 'ProvinceCode', as: 'Wards' } }, { \$project: { Code: 1, Name: 1, WardsCount: { \$size: '\$Wards' } } }]).toArray())"
```

**Management:**

List all collections:
```bash
mongosh "mongodb://root:Q35iSs8h5Y47VMcxZ5UC@localhost:27017/?authSource=admin" \
  --quiet --eval "use('vn_provinces'); db.getCollectionNames()"
```

List indexes for a collection:
```bash
mongosh "mongodb://root:Q35iSs8h5Y47VMcxZ5UC@localhost:27017/?authSource=admin" \
  --quiet --eval "use('vn_provinces'); db.getCollection('provinces-gis').getIndexes()"
```

Drop a collection:
```bash
mongosh "mongodb://root:Q35iSs8h5Y47VMcxZ5UC@localhost:27017/?authSource=admin" \
  --quiet --eval "use('vn_provinces'); db.getCollection('wards-gis').drop()"
```

Server version and stats:
```bash
mongosh "mongodb://root:Q35iSs8h5Y47VMcxZ5UC@localhost:27017/?authSource=admin" \
  --quiet --eval "db.version()"
mongosh "mongodb://root:Q35iSs8h5Y47VMcxZ5UC@localhost:27017/?authSource=admin" \
  --quiet --eval "db.serverStatus().connections"
```

### Agent Instructions

When context triggers a MongoDB query:
1. Check tunnel reachability first: `echo 'db.version()' | mongosh "mongodb://root:Q35iSs8h5Y47VMcxZ5UC@localhost:27017/?authSource=admin&directConnection=true" --quiet 2>/dev/null || echo "unreachable"`
2. If unreachable, start the tunnel: `ssh -f -N -L 27017:localhost:27017 -i ~/.ssh/id_ed25519 thanglequoc@machine.thanglequoc.xyz`
3. Prefer tunnel + local `mongosh` for rich output; fall back to `ssh ... docker exec mongodb mongosh` if local `mongosh` is unavailable
4. Use `--quiet` flag to suppress mongosh startup noise
5. Use `.countDocuments()` to get document counts — avoid fetching full documents
6. Use projection (`{Code: 1, Name: 1}`) to return only needed fields
7. Use `.limit(n)` for large result sets to avoid overwhelming output
8. Use `EJSON.stringify()` for readable JSON output when returning documents
9. If MongoDB returns an error, explain the issue clearly and suggest a fix
10. If the SSH tunnel fails to connect, report the error and suggest: verifying the bastion is online, checking SSH key permissions (`chmod 600 ~/.ssh/id_ed25519`), or running `ssh -v thanglequoc@machine.thanglequoc.xyz` to debug

---

## Project Structure

```
vietnamese-provinces-database/
├── AGENTS.md                          # This file
├── README.md / README_vi.md           # Dataset usage documentation
├── dataset-generation-scripts/        # Core Go automation service
│   ├── CLAUDE.md                      # Detailed subsystem context ← START HERE for code work
│   ├── main.go                        # Entry point
│   ├── go.mod / go.sum               # Go dependencies (Bun, postgres driver)
│   ├── .env                          # Database credentials
│   ├── docker/
│   │   └── docker-compose.yaml       # Local Postgres/PostGIS container
│   ├── internal/
│   │   ├── common/                   # Shared helpers
│   │   │   └── viet/                 # Vietnamese text handling (tone marks, normalization)
│   │   ├── dvhcvn_data_downloader/   # Direct DVHCVN administrative data ingestion (SOAP parser)
│   │   ├── sapnhap_bando/           # Geographic data service (HTTP API to sapnhap.bando.com.vn)
│   │   │   ├── fetcher/             # Loads JSON metadata & GeoJSON files
│   │   │   ├── service/             # Business logic (GIS fetch, backfill, metadata)
│   │   │   ├── repository/          # Database operations (sapnhap, gis, geojson_objects)
│   │   │   ├── model/               # Domain models (sapnhap, geojson)
│   │   │   ├── dto/                 # Data transfer objects (geojson_file, gis_server, sapnhap_api)
│   │   │   └── util/                # Name normalization utilities
│   │   ├── dumper/                  # Reads admin data, persists to DB
│   │   │   ├── config/              # Constant mappings
│   │   │   ├── helper/              # Dumper helper functions
│   │   │   ├── model/               # Dumper domain models
│   │   │   ├── repository/          # Seed data repository
│   │   │   └── service/             # DVHCVN SOAP seed dumper, corrector, manual seed dumper
│   │   ├── dataset_writer/          # Generates SQL/JSON/NoSQL output
│   │   │   └── dataset_file_writer/ # Per-format file writers (postgres/mysql, mssql, oracle, json, mongodb, redis, geojson)
│   │   │       ├── dto/             # Output DTOs (geojson, json, mongo)
│   │   │       └── helper/          # DTO mappers
│   │   ├── vn_provinces_tmp/        # Core VN provinces data layer (provinces_tmp, wards_tmp tables)
│   │   │   ├── model/               # Bun ORM models (Province, Ward, AdministrativeUnit, AdministrativeRegion)
│   │   │   └── repository/          # Repository queries
│   │   ├── gis/                     # GIS models and shared GIS logic
│   │   └── database/                # Postgres connection pool + bootstrap/SQL script execution
│   ├── resources/
│   │   ├── db_table_init.sql         # Core table schema (provinces_tmp, wards_tmp, etc.)
│   │   ├── db_region_administrative_unit.sql  # Region & administrative unit seed data
│   │   ├── fresh_cleanup.sql         # DB cleanup script (run before each generation)
│   │   ├── gis/
│   │   │   ├── geojson_11Mar2026/    # ← GeoJSON geometry (from deprecated API)
│   │   │   ├── sapnhapbando_geojson/ # Auxiliary GIS GeoJSON resources (3,355 files)
│   │   │   ├── sapnhap_bando_tables.sql          # GIS table schema
│   │   │   ├── sapnhapbando_init_geo_json_objects_tbl.sql  # Geo objects table init
│   │   │   └── sapnhapbando_geo_objects.sql       # Geo objects seed data
│   │   ├── manual_seeds/             # Manual fallback seed data (provinces_seed.sql + wards/)
│   │   └── rules/                    # Vietnamese text convention rules
│   │       └── vn_tone_mark_convention.md
│   ├── bando_co_dvch.sql             # Raw SAPNhap administrative unit SQL dump (497KB)
│   ├── sapnhap-bando-crawler/        # Historical/auxiliary crawler tooling (Python scripts)
│   ├── memory/
│   │   ├── MEMORY.md                # Memory index
│   │   └── feedback_*.md            # User preferences & learnings
│   ├── output/                       # Generated artifacts (gitignored, staging area)
│   └── tmp/                          # Temporary working directory
├── development/                       # Feature documentation & planning artifacts
│   ├── adapt_the_removal_of_sapnhap_api.md  # Context: API → file-based migration
│   ├── cleanup_old_reference/         # Completed cleanup plans (e.g., remove_bando_gisserver_references.md)
│   └── include_geojson_export/        # GeoJSON export feature planning
├── docs/
│   └── gis/                          # User-facing GIS documentation (gis_readme.md, gis_readme_vi.md, gis_example_query.md)
├── json/, mysql/, postgresql/, oracle/, sqlserver/, mongodb/, redis/
│   └── Generated dataset exports in various formats
└── .github/workflows/                # CI/CD pipelines (test-go.yml runs Go tests with PostGIS)
```

---

## Key Conventions

### Code Style
- **Language**: Go 1.24.0
- **ORM**: Bun with PostgreSQL dialect
- **Naming**: Database=`snake_case`, Go structs=`PascalCase`, JSON=`snake_case`
- **Error Handling**: Always wrap with context using `fmt.Errorf`

### Development Workflow

1. **Feature Planning**: Save detailed plans in `development/` folder (mandatory for AI-assisted work)
   - Example filename: `backfill-province-codes-from-tmp-tables.md`
   - Include: objectives, affected components, step-by-step logic, edge cases, assumptions

2. **Database Operations**: Use `docker exec` or `/db-query` skill for:
   - Data counts, totals, statistics
   - Data verification or integrity checks
   - Finding/searching specific records
   - GIS or geometry queries
   - Schema/table information

3. **Testing**: Use standard Go test patterns with Testify assertions

4. **Git**: Project uses Git LFS for large GIS files
   ```bash
   git lfs install
   git lfs pull
   ```

### Important Relationships

Current generation flow:
- **Administrative source of truth**: `main.go` defaults to `USE_DIRECT_DVHCVN_SOURCE = true`, so the dumper normally ingests administrative-unit data from the direct DVHCVN source before writing exports.
- **Fallback/manual path**: Manual seed data lives under `dataset-generation-scripts/resources/manual_seeds/` and is used only when the direct source path is disabled.
- **GIS enrichment**: GIS metadata and geometries are joined after the main administrative dump succeeds.

Geographic data migration context (March 2026):
- **Before**: GIS metadata fetched from SAPNhap API (`/pcotinh`, `/ptracuu`)
- **After (March 2026)**: GIS metadata and geometry load from local files:
  - `./resources/gis/geojson_11Mar2026/*.geojson`
- **After (July 2026)**: `bando_gisserver/` local JSON files removed (dead code cleanup). Live GIS uses HTTP API (`pread_json`, `p.co_dvhc_id`) + An Giang manual patch.
- **Key IDs**: `mahc` (province)→`sapnhap_province_matinh` (GIS), `maxa` (ward)→`sapnhap_ward_maxa` (GIS)
- **Status**: All 3,355 records verified with 100% GIS ID match rate

---

## Decision Reference

### When to Query vs When to Generate

| Task | Approach | Tools |
|------|----------|-------|
| Check province/ward counts | Query | `docker exec psql` or `/db-query` |
| Verify data integrity (duplicates, orphans) | Query | SQL verification scripts |
| Generate new SQL dumps | Execute script | `go run main.go` |
| Migrate new government decree | Query → Plan → Execute | Read decree, find affected records, generate patch |

### Government Decree Workflow

Vietnamese government issues administrative change decrees (e.g., `30/2026/QH16`, `19/2025/QĐ-TTg`). 

**When new decree arrives:**
1. Check `patch/` directory for similar historical decrees
2. Read decree document for specific changes (ward merges, promotions, reclassifications)
3. Query `provinces_tmp` and `wards_tmp` to find affected records
4. Document changes in `development/` as a feature plan
5. Update dumper logic if systematic changes required
6. Generate and validate patches for all database formats

**See**: `development/adapt_the_removal_of_sapnhap_api.md` for a recent complex migration example.

---

## Multi-Database Output

The project generates compatible SQL/data for:

- **PostgreSQL** (`postgresql/`) — Primary data source, with PostGIS
- **MySQL / MariaDB** (`mysql/`)
- **Microsoft SQL Server** (`sqlserver/`)
- **Oracle** (`oracle/`)
- **MongoDB** (`mongodb/`)
- **Redis** (`redis/`)

Output locations:
- `dataset-generation-scripts/output/` is the generator's immediate output/staging area when `go run main.go` is executed.
- Top-level folders such as `postgresql/`, `mysql/`, `sqlserver/`, `oracle/`, `mongodb/`, `redis/`, and `json/` contain the repository's published/exported dataset artifacts.

Typical contents in `dataset-generation-scripts/output/` include:
- timestamped import/export artifacts such as `postgresql_mysql_generated_ImportData_vn_units_*.sql`
- JSON, MongoDB, and Redis exports in format-specific subdirectories
- GIS import scripts in `output/gis/`

---

## Testing & Validation

- Unit tests in each `internal/*/` package (run with `go test -v ./...`)
- Historical patch verification in `patch/` directory
- GIS server ID matching (100% validation before release)

### CI/CD Pipeline (`.github/workflows/test-go.yml`)

The CI pipeline runs on pull requests to `main`/`master` and on manual dispatch:

1. **Service setup**: Spins up `postgis/postgis:15-3.3` as a service container on port `15432`
2. **Schema initialization**: Runs SQL scripts in order:
   - `resources/db_table_init.sql` — core tables
   - `resources/db_region_administrative_unit.sql` — region/unit seed data
   - `resources/gis/sapnhap_bando_tables.sql` — GIS tables
   - `resources/gis/sapnhapbando_init_geo_json_objects_tbl.sql` — geo objects table
3. **Environment variables** (required for tests):
   - `POSTGRES_DB_HOST=localhost`, `POSTGRES_DB_PORT=15432`
   - `POSTGRES_DB_USERNAME=postgres`, `POSTGRES_DB_PSWD=root`
   - `POSTGRES_TMP_DB_NAME=vn_provinces_tmp`
4. **Test execution**: `go test -v ./...` from `dataset-generation-scripts/`

> **Note**: `first-workflow.yml` is a trivial hello-world stub, not functional CI.

---

## Subsystem Deep Dive

For detailed code-level context on Go services, database models, and implementation details, see:

👉 **[dataset-generation-scripts/CLAUDE.md](dataset-generation-scripts/CLAUDE.md)** — Comprehensive guide to:
- Automatic database query triggers
- Docker PostgreSQL setup
- Database schema and relationships
- Project structure and component descriptions
- Code conventions and error handling

---

## Memory & Feedback

This project maintains persistent learnings:
- [Automatic Skill Invocation](dataset-generation-scripts/memory/feedback_auto_db_query.md) — Skills should auto-trigger based on context, not require manual `/skill-name` invocation

---

## Quick Links

| Resource | Purpose |
|----------|---------|
| [README.md](README.md) | User guide — dataset installation & usage |
| [CLAUDE.md](dataset-generation-scripts/CLAUDE.md) | Code agent guide — detailed subsystem context |
| [dataset-generation-scripts/README.md](dataset-generation-scripts/README.md) | Maintainer guide — how to run generation scripts |
| [docs/gis/](docs/gis/) | User-facing GIS documentation (readme, example queries) |
| [development/](development/) | Feature documentation & planning artifacts |
| [patch/](patch/) | Historical decree patches & changelog |
| [resources/gis/](dataset-generation-scripts/resources/gis/) | GeoJSON & metadata source files |
| [resources/rules/](dataset-generation-scripts/resources/rules/) | Vietnamese text convention rules |

---

## When to Escalate

| Scenario | Action |
|----------|--------|
| New government decree needs incorporation | Create plan doc in `development/`, verify GIS ID matches before release |
| Database schema needs changes | Query current state first, document migration plan, test on temporary DB |
| Performance issue in data generation | Profile with `pprof`, check `memory/` for previous optimization notes |
| New database format requested | Study existing implementations (e.g., `dataset_writer/`), add new dialect support |

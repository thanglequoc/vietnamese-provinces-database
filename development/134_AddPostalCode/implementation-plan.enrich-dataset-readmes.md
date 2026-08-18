# Enrich Generated Dataset READMEs — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the terse generated dataset READMEs with rich, useful documentation (Overview / Data Structure / Sample Document / Quick Start / Sample Queries / GIS) for every dataset type.

**Architecture:** Each writer's README function passes a large static `sections []string` slice to the existing shared `writeDatasetReadme` helper. No skeleton changes; only content changes. Restore-and-extend the removed `mongodb/gis` and original Elasticsearch README bodies.

**Tech Stack:** Go 1.24, stdlib, Testify. Tests run from `dataset-generation-scripts/`.

## Global Constraints

- Module root: `dataset-generation-scripts/`; tests: `go test ./internal/dataset_writer/dataset_file_writer/... -v`.
- `writeDatasetReadme(outputFolderPath, title, intro string, files []DatasetReadmeFile, sections []string) error` is unchanged.
- Every README must contain these six `## ` section headers: `Overview`, `Data Structure`, `Sample Document`, `Quick Start`, `Sample Queries`, `GIS / GeoJSON` (Oracle has no GIS section — it uses the first five; JSON uses `GIS / GeoJSON` describing `geojson/`).
- Counts are literals: 8 regions, 8 units, 34 provinces, 3,321 wards.
- No database/data access in README generation.

---

### Task 1: PostgreSQL/MySQL — rich README content

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/postgres_mysql_dataset_file_writer.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/postgres_mysql_dataset_file_writer_test.go`

**Interfaces:**
- Consumes: `writeDatasetReadme`, `DatasetReadmeFile` (unchanged).
- Produces: `postgresMySQLReadmeSections(engine, createTablesFile, pointLiteral string) []string`.

- [ ] **Step 1: Update the failing test**

In `TestPostgresMySQLDatasetFileWriter_WriteToFile_README` (postgres_mysql_dataset_file_writer_test.go), replace the assertions with richer section checks:

```go
func TestPostgresMySQLDatasetFileWriter_WriteToFile_README(t *testing.T) {
	tmpDir := t.TempDir()
	writer := &PostgresMySQLDatasetFileWriter{
		OutputFilePath: filepath.Join(tmpDir, "postgres_ImportData_vn_units.sql"),
	}
	provinces := []vn_provinces_tmp_model.Province{{Code: "01", Name: "Hà Nội", AdministrativeUnitId: 1}}

	err := writer.WriteToFile(nil, nil, provinces, nil)
	assert.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "README.md"))
	assert.NoError(t, err)
	s := string(content)
	assert.Contains(t, s, "**Generated at:")
	assert.Contains(t, s, "postgres_ImportData_vn_units.sql")
	assert.Contains(t, s, "## Overview")
	assert.Contains(t, s, "## Data Structure")
	assert.Contains(t, s, "## Sample Document")
	assert.Contains(t, s, "## Quick Start")
	assert.Contains(t, s, "## Sample Queries")
	assert.Contains(t, s, "## GIS / GeoJSON")
	assert.Contains(t, s, "postgres_CreateTables_vn_units.sql")
	assert.Contains(t, s, "administrative_regions")
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestPostgresMySQLDatasetFileWriter_WriteToFile_README -v`
Expected: FAIL — current README lacks `## Overview`, `## Sample Document`, `## Quick Start`, and `postgres_CreateTables_vn_units.sql`.

- [ ] **Step 3: Implement the rich content**

Replace the whole `postgresMySQLReadmeSections` function in `postgres_mysql_dataset_file_writer.go` with:

```go
func postgresMySQLReadmeSections(engine, createTablesFile, pointLiteral string) []string {
	return []string{
		"## Overview",
		"",
		"The Vietnamese Provinces Database for " + engine + ". The import script populates:",
		"",
		"| Table | Rows | Description |",
		"|-------|------|-------------|",
		"| `administrative_regions` | 8 | Regions of Vietnam |",
		"| `administrative_units` | 8 | Administrative unit types (city, province, ward, ...) |",
		"| `provinces` | 34 | Provinces and municipalities |",
		"| `wards` | 3,321 | Wards, communes, and town townships |",
		"",
		"GIS boundary geometry (in `gis/`) populates `gis_provinces` and `gis_wards`.",
		"",
		"## Data Structure",
		"",
		"### administrative_regions",
		"",
		"| Column | Description |",
		"|--------|-------------|",
		"| `id` | Region id (1-8) |",
		"| `name` | Region name (Vietnamese) |",
		"| `name_en` | Region name (English) |",
		"| `code_name` | URL-safe code name |",
		"| `code_name_en` | English code name |",
		"",
		"### administrative_units",
		"",
		"| Column | Description |",
		"|--------|-------------|",
		"| `id` | Unit type id (1-8) |",
		"| `full_name` / `short_name` | Vietnamese unit type names |",
		"| `full_name_en` / `short_name_en` | English unit type names |",
		"| `code_name` / `code_name_en` | Code names |",
		"",
		"### provinces",
		"",
		"| Column | Description |",
		"|--------|-------------|",
		"| `code` | Province code (PK, e.g. `01`) |",
		"| `name` / `name_en` | Province name |",
		"| `full_name` / `full_name_en` | Full name with unit prefix |",
		"| `code_name` | Code name (e.g. `ha_noi`) |",
		"| `administrative_unit_id` | FK to `administrative_units.id` |",
		"| `postal_code_prefix` | Comma-separated 2-digit postal prefixes |",
		"",
		"### wards",
		"",
		"| Column | Description |",
		"|--------|-------------|",
		"| `code` | Ward code (PK, e.g. `00004`) |",
		"| `name` / `name_en` | Ward name |",
		"| `full_name` / `full_name_en` | Full name with unit prefix |",
		"| `code_name` | Code name |",
		"| `province_code` | FK to `provinces.code` |",
		"| `administrative_unit_id` | FK to `administrative_units.id` |",
		"| `postal_code` | 5-digit national postal code |",
		"",
		"## Sample Document",
		"",
		"A province row:",
		"",
		"```sql",
		"INSERT INTO provinces(code,name,name_en,full_name,full_name_en,code_name,administrative_unit_id,postal_code_prefix) VALUES('01','Hà Nội','Ha Noi','Thành phố Hà Nội','Ha Noi City','ha_noi',1,'10, 11, 12, 13, 14');",
		"```",
		"",
		"## Quick Start",
		"",
		"1. Create the tables by running `" + createTablesFile + "`.",
		"2. Import the data with `psql -U <user> -d <db> -f postgres_ImportData_vn_units.sql` (PostgreSQL) or `mysql -u <user> -p <db> < mysql_ImportData_vn_units.sql` (MySQL).",
		"3. Import the GIS add-on (optional): run each chunk in `gis/` in the order listed in the `.manifest` file.",
		"",
		"## Sample Queries",
		"",
		"```sql",
		"-- Total provinces and wards",
		"SELECT (SELECT COUNT(*) FROM provinces) AS provinces, (SELECT COUNT(*) FROM wards) AS wards;",
		"",
		"-- Wards of Hà Nội (code 01), sorted by name",
		"SELECT w.code, w.name FROM wards w WHERE w.province_code = '01' ORDER BY w.name;",
		"",
		"-- Province by postal code prefix",
		"SELECT p.name FROM provinces p WHERE p.postal_code_prefix LIKE '%11%';",
		"",
		"-- GIS: province containing a point",
		"SELECT p.code, p.name",
		"FROM provinces p",
		"JOIN gis_provinces g ON p.code = g.province_code",
		"WHERE ST_Within(" + pointLiteral + ", g.geom);",
		"```",
		"",
		"## GIS / GeoJSON",
		"",
		"The `gis/` subfolder contains chunked GIS import scripts (`" + engine + "` `*-part-NN.sql`) plus a `.manifest` file listing the chunks in order. Import every chunk in order after the base import to populate `gis_provinces` and `gis_wards` with `bbox` and `geom`.",
	}
}
```

Update `writePostgresMySQLReadme` to pass the new arguments:

```go
func writePostgresMySQLReadme(outputFolderPath, outputFilePath string) error {
	isPostgres := strings.Contains(filepath.Base(outputFilePath), "postgres")
	if isPostgres {
		return writeDatasetReadme(outputFolderPath,
			"PostgreSQL Dataset — Vietnamese Provinces Database",
			"Import script for the Vietnamese Provinces Database on PostgreSQL/PostGIS.",
			[]DatasetReadmeFile{
				{Name: "postgres_ImportData_vn_units.sql", Description: "INSERT statements for regions, units, provinces, and wards"},
			},
			postgresMySQLReadmeSections("PostgreSQL", "postgres_CreateTables_vn_units.sql", "ST_SetSRID(ST_MakePoint(105.8542, 21.0285), 4326)"))
	}
	return writeDatasetReadme(outputFolderPath,
		"MySQL Dataset — Vietnamese Provinces Database",
		"Import script for the Vietnamese Provinces Database on MySQL/MariaDB.",
		[]DatasetReadmeFile{
			{Name: "mysql_ImportData_vn_units.sql", Description: "INSERT statements for regions, units, provinces, and wards"},
		},
		postgresMySQLReadmeSections("MySQL", "mysql_CreateTables_vn_units.sql", "ST_GeomFromText('POINT(105.8542 21.0285)', 4326)"))
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestPostgresMySQLDatasetFileWriter_WriteToFile_README -v`
Expected: PASS.

- [ ] **Step 5: Run the full package test suite**

Run: `go test ./internal/dataset_writer/dataset_file_writer/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/postgres_mysql_dataset_file_writer.go internal/dataset_writer/dataset_file_writer/postgres_mysql_dataset_file_writer_test.go
git commit -m "feat: enrich postgres/mysql dataset READMEs"
```

---

### Task 2: MSSQL — rich README content

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/mssql_dataset_file_writer.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/mssql_dataset_file_writer_test.go`

- [ ] **Step 1: Update the failing test**

In `TestMssqlDatasetFileWriter_WriteToFile_README` (mssql_dataset_file_writer_test.go), replace the assertions:

```go
	content, err := os.ReadFile(filepath.Join(tmpDir, "README.md"))
	assert.NoError(t, err)
	s := string(content)
	assert.Contains(t, s, "**Generated at:")
	assert.Contains(t, s, "mssql_ImportData_vn_units.sql")
	assert.Contains(t, s, "## Overview")
	assert.Contains(t, s, "## Data Structure")
	assert.Contains(t, s, "## Sample Document")
	assert.Contains(t, s, "## Quick Start")
	assert.Contains(t, s, "## Sample Queries")
	assert.Contains(t, s, "## GIS / GeoJSON")
	assert.Contains(t, s, "mssql_CreateTables_vn_units.sql")
	assert.Contains(t, s, "STContains")
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestMssqlDatasetFileWriter_WriteToFile_README -v`
Expected: FAIL — missing the new sections and markers.

- [ ] **Step 3: Implement the rich content**

Replace `writeMssqlReadme` in `mssql_dataset_file_writer.go` with:

```go
func writeMssqlReadme(outputFolderPath string) error {
	return writeDatasetReadme(outputFolderPath,
		"Microsoft SQL Server Dataset — Vietnamese Provinces Database",
		"Import script for the Vietnamese Provinces Database on Microsoft SQL Server.",
		[]DatasetReadmeFile{
			{Name: "mssql_ImportData_vn_units.sql", Description: "INSERT statements for regions, units, provinces, and wards"},
		},
		[]string{
			"## Overview",
			"",
			"The Vietnamese Provinces Database for Microsoft SQL Server. The import script populates:",
			"",
			"| Table | Rows | Description |",
			"|-------|------|-------------|",
			"| `administrative_regions` | 8 | Regions of Vietnam |",
			"| `administrative_units` | 8 | Administrative unit types (city, province, ward, ...) |",
			"| `provinces` | 34 | Provinces and municipalities |",
			"| `wards` | 3,321 | Wards, communes, and town townships |",
			"",
			"GIS boundary geometry (in `gis/`) populates `gis_provinces` and `gis_wards`.",
			"",
			"## Data Structure",
			"",
			"### provinces",
			"",
			"| Column | Description |",
			"|--------|-------------|",
			"| `code` | Province code (PK, e.g. `01`) |",
			"| `name` / `name_en` | Province name |",
			"| `full_name` / `full_name_en` | Full name with unit prefix |",
			"| `code_name` | Code name (e.g. `ha_noi`) |",
			"| `administrative_unit_id` | FK to `administrative_units.id` |",
			"| `postal_code_prefix` | Comma-separated 2-digit postal prefixes |",
			"",
			"### wards",
			"",
			"| Column | Description |",
			"|--------|-------------|",
			"| `code` | Ward code (PK, e.g. `00004`) |",
			"| `name` / `name_en` | Ward name |",
			"| `full_name` / `full_name_en` | Full name with unit prefix |",
			"| `code_name` | Code name |",
			"| `province_code` | FK to `provinces.code` |",
			"| `administrative_unit_id` | FK to `administrative_units.id` |",
			"| `postal_code` | 5-digit national postal code |",
			"",
			"`administrative_regions` and `administrative_units` hold the region and unit-type lookup rows (8 each).",
			"",
			"## Sample Document",
			"",
			"A province row:",
			"",
			"```sql",
			"INSERT INTO provinces(code,name,name_en,full_name,full_name_en,code_name,administrative_unit_id,postal_code_prefix) VALUES('01',N'Hà Nội',N'Ha Noi',N'Thành phố Hà Nội',N'Ha Noi City','ha_noi',1,'10, 11, 12, 13, 14');",
			"```",
			"",
			"## Quick Start",
			"",
			"1. Create the tables by running `mssql_CreateTables_vn_units.sql`.",
			"2. Import the data with `sqlcmd -S <server> -d <db> -U <user> -P <pass> -i mssql_ImportData_vn_units.sql`.",
			"3. Import the GIS add-on (optional): run each chunk in `gis/` in the order listed in the `.manifest` file.",
			"",
			"## Sample Queries",
			"",
			"```sql",
			"SELECT (SELECT COUNT(*) FROM provinces) AS provinces, (SELECT COUNT(*) FROM wards) AS wards;",
			"",
			"SELECT w.code, w.name FROM wards w WHERE w.province_code = '01' ORDER BY w.name;",
			"",
			"-- GIS: province containing a point",
			"SELECT p.code, p.name",
			"FROM provinces p",
			"JOIN gis_provinces g ON p.code = g.province_code",
			"WHERE g.geom.STContains(geometry::STGeomFromText('POINT(105.8542 21.0285)', 4326)) = 1;",
			"```",
			"",
			"## GIS / GeoJSON",
			"",
			"The `gis/` subfolder contains chunked GIS import scripts (`mssql_ImportData_gis-part-NN.sql`) plus a `.manifest` file listing the chunks in order. Import every chunk in order after the base import.",
		})
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestMssqlDatasetFileWriter_WriteToFile_README -v`
Expected: PASS.

- [ ] **Step 5: Run the full package test suite**

Run: `go test ./internal/dataset_writer/dataset_file_writer/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/mssql_dataset_file_writer.go internal/dataset_writer/dataset_file_writer/mssql_dataset_file_writer_test.go
git commit -m "feat: enrich mssql dataset README"
```

---

### Task 3: Oracle — rich README content

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/oracle_dataset_file_writer.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/oracle_dataset_file_writer_test.go`

- [ ] **Step 1: Update the failing test**

In `TestOracleDatasetFileWriter_WriteToFile_README` (oracle_dataset_file_writer_test.go), replace the assertions:

```go
	content, err := os.ReadFile(filepath.Join(tmpDir, "README.md"))
	assert.NoError(t, err)
	s := string(content)
	assert.Contains(t, s, "**Generated at:")
	assert.Contains(t, s, "oracle_ImportData_vn_units.sql")
	assert.Contains(t, s, "## Overview")
	assert.Contains(t, s, "## Data Structure")
	assert.Contains(t, s, "## Sample Document")
	assert.Contains(t, s, "## Quick Start")
	assert.Contains(t, s, "## Sample Queries")
	assert.Contains(t, s, "oracle_CreateTables_vn_units.sql")
	assert.Contains(t, s, "INSERT ALL")
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestOracleDatasetFileWriter_WriteToFile_README -v`
Expected: FAIL — missing new sections/markers.

- [ ] **Step 3: Implement the rich content**

Replace `writeOracleReadme` in `oracle_dataset_file_writer.go` with:

```go
func writeOracleReadme(outputFolderPath string) error {
	return writeDatasetReadme(outputFolderPath,
		"Oracle Dataset — Vietnamese Provinces Database",
		"Import script for the Vietnamese Provinces Database on Oracle.",
		[]DatasetReadmeFile{
			{Name: "oracle_ImportData_vn_units.sql", Description: "INSERT ALL statements for regions, units, provinces, and wards"},
		},
		[]string{
			"## Overview",
			"",
			"The Vietnamese Provinces Database for Oracle. The import script populates:",
			"",
			"| Table | Rows | Description |",
			"|-------|------|-------------|",
			"| `administrative_regions` | 8 | Regions of Vietnam |",
			"| `administrative_units` | 8 | Administrative unit types (city, province, ward, ...) |",
			"| `provinces` | 34 | Provinces and municipalities |",
			"| `wards` | 3,321 | Wards, communes, and town townships |",
			"",
			"## Data Structure",
			"",
			"### provinces",
			"",
			"| Column | Description |",
			"|--------|-------------|",
			"| `code` | Province code (PK, e.g. `01`) |",
			"| `name` / `name_en` | Province name |",
			"| `full_name` / `full_name_en` | Full name with unit prefix |",
			"| `code_name` | Code name (e.g. `ha_noi`) |",
			"| `administrative_unit_id` | FK to `administrative_units.id` |",
			"| `postal_code_prefix` | Comma-separated 2-digit postal prefixes |",
			"",
			"### wards",
			"",
			"| Column | Description |",
			"|--------|-------------|",
			"| `code` | Ward code (PK, e.g. `00004`) |",
			"| `name` / `name_en` | Ward name |",
			"| `full_name` / `full_name_en` | Full name with unit prefix |",
			"| `code_name` | Code name |",
			"| `province_code` | FK to `provinces.code` |",
			"| `administrative_unit_id` | FK to `administrative_units.id` |",
			"| `postal_code` | 5-digit national postal code |",
			"",
			"`administrative_regions` and `administrative_units` hold the region and unit-type lookup rows (8 each).",
			"",
			"## Sample Document",
			"",
			"A province row inside the multi-row `INSERT ALL` batch:",
			"",
			"```sql",
			"\tINTO provinces(code,name,name_en,full_name,full_name_en,code_name,administrative_unit_id,postal_code_prefix) VALUES('01','Hà Nội','Ha Noi','Thành phố Hà Nội','Ha Noi City','ha_noi',1,'10, 11, 12, 13, 14')",
			"```",
			"",
			"## Quick Start",
			"",
			"1. Create the tables by running `oracle_CreateTables_vn_units.sql`.",
			"2. Import the data with `sqlplus <user>/<password>@<db> @oracle_ImportData_vn_units.sql`.",
			"",
			"## Sample Queries",
			"",
			"```sql",
			"SELECT (SELECT COUNT(*) FROM provinces) AS provinces, (SELECT COUNT(*) FROM wards) AS wards FROM dual;",
			"",
			"SELECT w.code, w.name FROM wards w WHERE w.province_code = '01' ORDER BY w.name;",
			"",
			"-- Province by postal code prefix",
			"SELECT p.name FROM provinces p WHERE p.postal_code_prefix LIKE '%11%';",
			"```",
		})
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestOracleDatasetFileWriter_WriteToFile_README -v`
Expected: PASS.

- [ ] **Step 5: Run the full package test suite**

Run: `go test ./internal/dataset_writer/dataset_file_writer/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/oracle_dataset_file_writer.go internal/dataset_writer/dataset_file_writer/oracle_dataset_file_writer_test.go
git commit -m "feat: enrich oracle dataset README"
```

---

### Task 4: MongoDB — rich README content (restore gis content + base collections)

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/mongodb_file_writer.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/mongodb_file_writer_test.go`

- [ ] **Step 1: Update the failing test**

In `TestMongoDBDatasetFileWriter_WriteToFile_README` (mongodb_file_writer_test.go), replace the assertions:

```go
	content, err := os.ReadFile(filepath.Join(tmpDir, "README.md"))
	assert.NoError(t, err)
	s := string(content)
	assert.Contains(t, s, "**Generated at:")
	assert.Contains(t, s, "mongo_data_vn_unit.json")
	assert.Contains(t, s, "## Overview")
	assert.Contains(t, s, "## Data Structure")
	assert.Contains(t, s, "## Sample Document")
	assert.Contains(t, s, "## Quick Start")
	assert.Contains(t, s, "## Sample Queries")
	assert.Contains(t, s, "## GIS / GeoJSON")
	assert.Contains(t, s, "$geoIntersects")
	assert.Contains(t, s, "mongoimport")
	assert.Contains(t, s, "create_indexes.js")
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestMongoDBDatasetFileWriter_WriteToFile_README -v`
Expected: FAIL — missing new sections/markers.

- [ ] **Step 3: Implement the rich content**

Replace `writeMongoReadme` in `mongodb_file_writer.go` with:

```go
func writeMongoReadme(outputFolderPath string) error {
	return writeDatasetReadme(outputFolderPath,
		"MongoDB Dataset — Vietnamese Provinces Database",
		"MongoDB documents for Vietnamese provinces and wards, with an optional GIS add-on.",
		[]DatasetReadmeFile{
			{Name: "administrative_units.json", Description: "Array of 8 administrative unit types"},
			{Name: "administrative_regions.json", Description: "Array of 8 regions"},
			{Name: "mongo_data_vn_unit.json", Description: "Array of 34 province documents, each embedding its Wards array"},
		},
		[]string{
			"## Overview",
			"",
			"| Collection | Documents | Description |",
			"|------------|-----------|-------------|",
			"| `provinces` | 34 | Province documents with embedded wards (`mongo_data_vn_unit.json`) |",
			"| `provinces-gis` | 34 | GIS add-on: province documents with geometry |",
			"| `wards-gis` | 3,321 | GIS add-on: standalone ward documents with geometry |",
			"",
			"`administrative_units.json` and `administrative_regions.json` provide the lookup data (8 each).",
			"",
			"## Data Structure",
			"",
			"A province document in the `provinces` collection:",
			"",
			"- **`Code`** — province code (`01`)",
			"- **`Name` / `NameEn`** — province name",
			"- **`FullName` / `FullNameEn`** — full name with unit prefix",
			"- **`CodeName`** — code name (`ha_noi`)",
			"- **`AdministrativeUnit`** — embedded unit object (Id, FullName, ShortName, ...)",
			"- **`SearchKeywords`** — pre-computed autocomplete keywords (code, tone-stripped name, English name, codeName)",
			"- **`Wards`** — embedded array of ward documents (same field shape, plus `PostalCode` and `ProvinceCode`)",
			"- **`Meta`** — dataset version metadata",
			"",
			"The GIS collections add a **`GIS`** object: `Center` (GeoJSON Point), `BoundingBox`, `Geometry` (GeoJSON MultiPolygon/Polygon), and `Properties`.",
			"",
			"## Sample Document",
			"",
			"```json",
			"{",
			"  \"Code\": \"01\",",
			"  \"Name\": \"Hà Nội\",",
			"  \"NameEn\": \"Ha Noi\",",
			"  \"FullName\": \"Thành phố Hà Nội\",",
			"  \"FullNameEn\": \"Ha Noi City\",",
			"  \"CodeName\": \"ha_noi\",",
			"  \"AdministrativeUnit\": { \"Id\": 1, \"FullName\": \"Thành phố trực thuộc trung ương\", \"ShortName\": \"Thành phố\" },",
			"  \"SearchKeywords\": [\"01\", \"ha noi\", \"ha_noi\"],",
			"  \"Wards\": [",
			"    { \"Code\": \"00004\", \"Name\": \"Ba Đình\", \"FullName\": \"Phường Ba Đình\", \"PostalCode\": \"11120\" }",
			"  ]",
			"}",
			"```",
			"",
			"## Quick Start",
			"",
			"1. Import the data:",
			"",
			"```bash",
			"mongoimport --db vn_provinces --collection provinces --file mongo_data_vn_unit.json --jsonArray",
			"mongoimport --db vn_provinces --collection administrative_units --file administrative_units.json --jsonArray",
			"mongoimport --db vn_provinces --collection administrative_regions --file administrative_regions.json --jsonArray",
			"```",
			"",
			"2. GIS add-on (optional): import each file in `gis/` (ward files may be chunked — follow the `.manifest`), then run `mongosh vn_provinces create_indexes.js`.",
			"",
			"## Sample Queries",
			"",
			"```javascript",
			"// Count provinces",
			"db.getCollection('provinces').countDocuments();",
			"",
			"// Wards of Hà Nội",
			"db.getCollection('provinces').findOne({Code: '01'}, {Name: 1, Wards: 1});",
			"",
			"// GIS: province containing a point",
			"db.getCollection('provinces-gis').findOne({",
			"  \"GIS.Geometry\": { $geoIntersects: { $geometry: { type: 'Point', coordinates: [105.8542, 21.0285] } } }",
			"});",
			"",
			"// GIS: wards in a province",
			"db.getCollection('wards-gis').find({ProvinceCode: '01'}).limit(10);",
			"```",
			"",
			"## GIS / GeoJSON",
			"",
			"The `gis/` subfolder contains the `provinces-gis` and `wards-gis` collections (`mongo_data_vn_province_gis.json`, `mongo_data_vn_ward_gis[_part_NN].json`), the `create_indexes.js` index script, and a `.manifest`. Import them, run `create_indexes.js`, then query with `$geoIntersects` and `$near`.",
		})
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestMongoDBDatasetFileWriter_WriteToFile_README -v`
Expected: PASS.

- [ ] **Step 5: Run the full package test suite**

Run: `go test ./internal/dataset_writer/dataset_file_writer/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/mongodb_file_writer.go internal/dataset_writer/dataset_file_writer/mongodb_file_writer_test.go
git commit -m "feat: enrich mongodb dataset README"
```

---

### Task 5: Redis — rich README content

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/redis_file_writer.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/redis_file_writer_test.go`

- [ ] **Step 1: Update the failing test**

In `TestRedisDatasetFileWriter_WriteToFile_README` (redis_file_writer_test.go), replace the assertions:

```go
	content, err := os.ReadFile(filepath.Join(tmpDir, "README.md"))
	assert.NoError(t, err)
	s := string(content)
	assert.Contains(t, s, "**Generated at:")
	assert.Contains(t, s, "redis_vn_provinces_dataset.redis")
	assert.Contains(t, s, "## Overview")
	assert.Contains(t, s, "## Data Structure")
	assert.Contains(t, s, "## Sample Document")
	assert.Contains(t, s, "## Quick Start")
	assert.Contains(t, s, "## Sample Queries")
	assert.Contains(t, s, "--pipe")
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestRedisDatasetFileWriter_WriteToFile_README -v`
Expected: FAIL — missing new sections and `--pipe`.

- [ ] **Step 3: Implement the rich content**

Replace `writeRedisReadme` in `redis_file_writer.go` with:

```go
func writeRedisReadme(outputFolderPath string) error {
	return writeDatasetReadme(outputFolderPath,
		"Redis Dataset — Vietnamese Provinces Database",
		"Redis commands (HSET/SADD) loading all Vietnamese provinces, wards, regions, and administrative units.",
		[]DatasetReadmeFile{
			{Name: "redis_vn_provinces_dataset.redis", Description: "Redis HSET/SADD commands"},
		},
		[]string{
			"## Overview",
			"",
			"The dataset stores every administrative unit as Redis hashes and sets:",
			"",
			"| Key pattern | Type | Count |",
			"|-------------|------|-------|",
			"| `province:<code>` | hash | 34 |",
			"| `ward:<code>` | hash | 3,321 |",
			"| `administrativeUnit:<id>` | hash | 8 |",
			"| `region:<id>` | hash | 8 |",
			"| `province:<code>:wards` | set | 34 |",
			"| `province:<code>:wards:vn` / `:en` | hash | 34 each |",
			"",
			"## Data Structure",
			"",
			"`province:<code>` fields: `code`, `name`, `nameEn`, `fullName`, `fullNameEn`, `codeName`, `postalCodePrefix`, `administrativeUnitId`.",
			"",
			"`ward:<code>` fields: `code`, `name`, `nameEn`, `fullName`, `fullNameEn`, `codeName`, `postalCode`, `administrativeUnitId`, `districtCode`.",
			"",
			"`province:<code>:wards:vn` / `:en` map ward codes to Vietnamese/English full names.",
			"",
			"## Sample Document",
			"",
			"```bash",
			"HSET province:01 code \"01\" name \"Hà Nội\" nameEn \"Ha Noi\" fullName \"Thành phố Hà Nội\" fullNameEn \"Ha Noi City\" codeName \"ha_noi\" postalCodePrefix \"10, 11, 12, 13, 14\" administrativeUnitId 1",
			"",
			"SADD province:01:wards \"00004\"",
			"HSET province:01:wards:vn \"00004\" \"Phường Ba Đình\"",
			"```",
			"",
			"## Quick Start",
			"",
			"```bash",
			"redis-cli --pipe < redis_vn_provinces_dataset.redis",
			"```",
			"",
			"## Sample Queries",
			"",
			"```bash",
			"redis-cli HGETALL province:01",
			"redis-cli SMEMBERS province:01:wards",
			"redis-cli HGET ward:00004 fullName",
			"redis-cli HGET province:01:wards:vn 00004",
			"```",
		})
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestRedisDatasetFileWriter_WriteToFile_README -v`
Expected: PASS.

- [ ] **Step 5: Run the full package test suite**

Run: `go test ./internal/dataset_writer/dataset_file_writer/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/redis_file_writer.go internal/dataset_writer/dataset_file_writer/redis_file_writer_test.go
git commit -m "feat: enrich redis dataset README"
```

---

### Task 6: Elasticsearch — restore rich README content

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer_test.go`

- [ ] **Step 1: Update the failing test**

In `TestWriteToFile_NonGIS` (elasticsearch_file_writer_test.go), after the existing `## Sample Queries` assertion, append:

```go
	if !bytes.Contains(readme, []byte("## Overview")) {
		t.Fatal("README.md missing Overview section")
	}
	if !bytes.Contains(readme, []byte("## Data Structure")) {
		t.Fatal("README.md missing Data Structure section")
	}
	if !bytes.Contains(readme, []byte("## Sample Document")) {
		t.Fatal("README.md missing Sample Document section")
	}
	if !bytes.Contains(readme, []byte("## Quick Start")) {
		t.Fatal("README.md missing Quick Start section")
	}
	if !bytes.Contains(readme, []byte("_bulk")) {
		t.Fatal("README.md missing _bulk import reference")
	}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestWriteToFile_NonGIS -v`
Expected: FAIL — current README lacks `## Sample Document`, `## Quick Start`, and `_bulk`.

- [ ] **Step 3: Implement the rich content**

Replace `writeESReadme` in `elasticsearch_file_writer.go` with:

```go
// writeESReadme writes the README.md for the Elasticsearch dataset.
func writeESReadme(path string) error {
	outputFolderPath := filepath.Dir(path)

	sections := []string{
		"## Overview",
		"",
		"This dataset provides Vietnamese provinces and wards in Elasticsearch document format with two indices:",
		"",
		"| Index | Documents | Description |",
		"|-------|-----------|-------------|",
		"| `provinces` | 34 | Provincial metadata with embedded wards, search keywords, and administrative unit data (no GIS geometry) |",
		"| `provinces-gis` | 34 | Same structure plus GIS geometry for both provinces and wards (bounding boxes + GeoJSON polygons) |",
		"",
		"## Data Structure",
		"",
		"Each province is a single denormalized document with:",
		"",
		"- **Core fields**: `Code`, `Name`, `NameEn`, `FullName`, `FullNameEn`, `CodeName`",
		"- **`AdministrativeUnit`**: embedded administrative unit object (Id, FullName, ShortName, CodeName, ...)",
		"- **`SearchKeywords`**: pre-computed autocomplete keywords (code, tone-stripped name, English name, codeName)",
		"- **`Wards`**: nested array of ward documents with the same field shape (plus `PostalCode`)",
		"- **`Meta`**: `DatasetVersion`, `AdministrativeRevision`, `GeneratedAt`",
		"- **`GIS`**: (provinces-gis only) `Center` (geo_point), `BoundingBox`, `Geometry` (geo_shape), `Properties`",
		"",
		"## Sample Document",
		"",
		"```json",
		"{",
		"  \"Code\": \"01\",",
		"  \"Name\": \"Hà Nội\",",
		"  \"NameEn\": \"Ha Noi\",",
		"  \"FullName\": \"Thành phố Hà Nội\",",
		"  \"FullNameEn\": \"Ha Noi City\",",
		"  \"CodeName\": \"ha_noi\",",
		"  \"AdministrativeUnit\": { \"Id\": 1, \"FullName\": \"Thành phố trực thuộc trung ương\", \"ShortName\": \"Thành phố\" },",
		"  \"SearchKeywords\": [\"01\", \"ha noi\", \"ha_noi\"],",
		"  \"Wards\": [",
		"    { \"Code\": \"00004\", \"Name\": \"Ba Đình\", \"FullName\": \"Phường Ba Đình\", \"PostalCode\": \"11120\" }",
		"  ]",
		"}",
		"```",
		"",
		"## Quick Start",
		"",
		"1. Create the indices with the mappings in `mappings/`.",
		"",
		"```bash",
		`curl -X PUT "localhost:9200/provinces" -H 'Content-Type: application/json' -d @mappings/provinces.json`,
		`curl -X PUT "localhost:9200/provinces-gis" -H 'Content-Type: application/json' -d @mappings/provinces-gis.json`,
		"```",
		"",
		"2. Bulk import `provinces.ndjson`, and the `provinces-gis-part-*.ndjson` chunks in order (per `provinces-gis.ndjson.manifest`):",
		"",
		"```bash",
		`curl -X POST "localhost:9200/_bulk" -H 'Content-Type: application/x-ndjson' --data-binary @provinces.ndjson`,
		`curl -X POST "localhost:9200/_bulk" -H 'Content-Type: application/x-ndjson' --data-binary @provinces-gis-part-01.ndjson`,
		"```",
		"",
		"3. Verify: 34 documents in each index.",
		"",
		"## Sample Queries",
		"",
		"```json",
		"// Count documents",
		"POST /provinces/_count",
		"",
		"// Autocomplete search",
		"POST /provinces/_search",
		"{ \"query\": { \"terms\": { \"SearchKeywords\": [\"ha noi\"] } }, \"_source\": [\"Code\", \"Name\", \"NameEn\"] }",
		"",
		"// Search a ward and return the matched nested document only",
		"POST /provinces/_search",
		"{ \"_source\": false, \"query\": { \"nested\": { \"path\": \"Wards\", \"query\": { \"match\": { \"Wards.CodeName\": \"ba_dinh\" } }, \"inner_hits\": {} } } }",
		"",
		"// GIS: find province covering a point",
		"POST /provinces-gis/_search",
		"{ \"query\": { \"geo_shape\": { \"GIS.Geometry\": { \"shape\": { \"type\": \"point\", \"coordinates\": [105.8542, 21.0285] }, \"relation\": \"intersects\" } } }, \"_source\": [\"Code\", \"Name\"] }",
		"```",
		"",
		"## Notes",
		"",
		"- Field names use **PascalCase** (consistent with MongoDB/JSON exports).",
		"- The `Meta` field is named without an underscore prefix — Elasticsearch reserves `_`-prefixed field names.",
		"- NDJSON files use the Elasticsearch Bulk API format.",
	}

	return writeDatasetReadme(outputFolderPath,
		"Elasticsearch Dataset — Vietnamese Provinces Database",
		"Provinces and wards as Elasticsearch documents in two indices: `provinces` (no geometry) and `provinces-gis` (with GIS geometry).",
		[]DatasetReadmeFile{
			{Name: "provinces.ndjson", Description: "Bulk API NDJSON for the provinces index"},
			{Name: "provinces-gis-part-01.ndjson", Description: "Bulk API NDJSON for provinces-gis (part 1 of 5)"},
			{Name: "provinces-gis.ndjson.manifest", Description: "Ordered chunk list for provinces-gis"},
			{Name: "mappings/provinces.json", Description: "Index mapping for provinces"},
			{Name: "mappings/provinces-gis.json", Description: "Index mapping for provinces-gis"},
		},
		sections)
}
```

> Note: `## Sample Queries` uses JSON query bodies (as in the original README) rather than `curl` one-liners; the `curl` quick-start covers execution. The `## Sample Queries` header still appears, satisfying the existing test assertion.

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run "TestWriteToFile_NonGIS|TestWriteElasticsearchGISDataToFile_GIS" -v`
Expected: PASS.

- [ ] **Step 5: Run the full package test suite**

Run: `go test ./internal/dataset_writer/dataset_file_writer/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer.go internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer_test.go
git commit -m "feat: restore rich elasticsearch dataset README"
```

---

### Task 7: JSON — rich README content

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/json_file_writer.go`
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/json_file_writer_test.go`

- [ ] **Step 1: Update the failing test**

In `TestJSONDatasetFileWriter_WriteToFile_README` (json_file_writer_test.go), extend the assertions:

```go
	assert.Contains(t, contentStr, "## Data Structure")
	assert.Contains(t, contentStr, "## Sample Queries")
	assert.Contains(t, contentStr, "## Overview")
	assert.Contains(t, contentStr, "## Sample Document")
	assert.Contains(t, contentStr, "## Quick Start")
	assert.Contains(t, contentStr, "require(")
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestJSONDatasetFileWriter_WriteToFile_README -v`
Expected: FAIL — missing new sections and `require(`.

- [ ] **Step 3: Implement the rich content**

Replace `writeJSONDatasetReadme` in `json_file_writer.go` with:

```go
// writeJSONDatasetReadme writes a README describing the JSON dataset files and the
// optional geojson artifacts, with a bold generation timestamp at the top.
func writeJSONDatasetReadme(outputFolderPath string) error {
	return writeDatasetReadme(outputFolderPath,
		"JSON Dataset — Vietnamese Provinces Database",
		"Administrative unit JSON data for Vietnam: provinces with embedded wards, in full and simplified forms.",
		[]DatasetReadmeFile{
			{Name: "full_json_generated_data_vn_units.json", Description: "Full dataset (provinces + wards + administrative info)"},
			{Name: "simplified_json_generated_data_vn_units.json", Description: "Simplified dataset (pretty-printed)"},
			{Name: "simplified_json_generated_data_vn_units_minified.json", Description: "Simplified dataset (minified)"},
			{Name: "vn_only_simplified_json_generated_data_vn_units.json", Description: "Vietnamese-only simplified (pretty-printed)"},
			{Name: "vn_only_simplified_json_generated_data_vn_units_minified.json", Description: "Vietnamese-only simplified (minified)"},
		},
		[]string{
			"## Overview",
			"",
			"| File | Contents |",
			"|------|----------|",
			"| `full_json_generated_data_vn_units.json` | 34 province objects with full administrative info and embedded wards |",
			"| `simplified_json_generated_data_vn_units.json` | Same structure, fewer fields (pretty-printed) |",
			"| `simplified_json_generated_data_vn_units_minified.json` | Simplified, minified (no whitespace) |",
			"| `vn_only_simplified_json_generated_data_vn_units.json` | Vietnamese-only fields (pretty-printed) |",
			"| `vn_only_simplified_json_generated_data_vn_units_minified.json` | Vietnamese-only fields (minified) |",
			"",
			"## Data Structure",
			"",
			"Each entry is a province object:",
			"",
			"- **`code`** — province code (`01`)",
			"- **`name` / `nameEn`** — province name",
			"- **`fullName` / `fullNameEn`** — full name with unit prefix",
			"- **`codeName`** — code name (`ha_noi`)",
			"- **`administrativeUnitId` / `administrativeUnitShortName` / `administrativeUnitFullName`** — unit type",
			"- **`postalCodePrefix`** — comma-separated 2-digit postal prefixes",
			"- **`wards`** — array of ward objects (`code`, `name`, `nameEn`, `fullName`, `fullNameEn`, `codeName`, `provinceCode`, `postalCode`, unit fields)",
			"",
			"## Sample Document",
			"",
			"```json",
			"[",
			"  {",
			"    \"code\": \"01\",",
			"    \"name\": \"Hà Nội\",",
			"    \"nameEn\": \"Ha Noi\",",
			"    \"fullName\": \"Thành phố Hà Nội\",",
			"    \"fullNameEn\": \"Ha Noi City\",",
			"    \"codeName\": \"ha_noi\",",
			"    \"postalCodePrefix\": \"10, 11, 12, 13, 14\",",
			"    \"wards\": [",
			"      { \"code\": \"00004\", \"name\": \"Ba Đình\", \"nameEn\": \"Ba Dinh\", \"postalCode\": \"11120\" }",
			"    ]",
			"  }",
			"]",
			"```",
			"",
			"## Quick Start",
			"",
			"```js",
			"const dataset = require('./full_json_generated_data_vn_units.json');",
			"```",
			"",
			"```python",
			"import json",
			"with open('full_json_generated_data_vn_units.json', encoding='utf-8') as f:",
			"    dataset = json.load(f)",
			"```",
			"",
			"## Sample Queries",
			"",
			"```js",
			"// Find Hà Nội",
			"dataset.find(p => p.code === '01');",
			"",
			"// Wards with postal code 11024",
			"dataset.flatMap(p => p.wards).filter(w => w.postalCode === '11024');",
			"",
			"// All province names",
			"dataset.map(p => p.name);",
			"```",
			"",
			"## GIS / GeoJSON",
			"",
			"The `geojson/` subfolder contains per-province and per-ward GeoJSON boundary exports, and `vn_provinces_wards_geojson.zip` is the combined archive of those files. These artifacts are present when the GIS generation step runs.",
		})
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/dataset_writer/dataset_file_writer/... -run TestJSONDatasetFileWriter_WriteToFile_README -v`
Expected: PASS.

- [ ] **Step 5: Run the full package test suite**

Run: `go test ./internal/dataset_writer/dataset_file_writer/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/dataset_writer/dataset_file_writer/json_file_writer.go internal/dataset_writer/dataset_file_writer/json_file_writer_test.go
git commit -m "feat: enrich json dataset README"
```

---

### Task 8: Final verification and publish

**Files:** none (regeneration + publish + commit of generated data)

- [ ] **Step 1: Run the full test suite**

Run from `dataset-generation-scripts/`: `go test ./... 2>&1 | grep -E 'FAIL|ok '`
Expected: all `ok`, no FAIL.

- [ ] **Step 2: Run `go vet`**

Run: `go vet ./internal/dataset_writer/...`
Expected: no output.

- [ ] **Step 3: Regenerate and publish**

Start Docker Postgres if needed (`docker compose -f docker/docker-compose.yaml up -d`), then:

```bash
go run main.go
bash copy-datasets-to-repo.sh
rm -f json/geojson/README.md mongodb/gis/README.md   # stale gis-subfolder READMEs (already absent in output)
```

- [ ] **Step 4: Spot-check READMEs**

Run: `head -30 postgresql/README.md mongodb/README.md elasticsearch/README.md`
Expected: all six section headers present in each; `**Generated at:**` bold timestamp.

- [ ] **Step 5: Commit the regenerated data**

```bash
git add json/ postgresql/ mysql/ sqlserver/ oracle/ mongodb/ redis/ elasticsearch/
git commit -m "data: regenerate and publish datasets with enriched READMEs"
```

(Leave `docs/release_notes_v4.1.0.md` and `docs/release_notes_v4.2.0.md` untracked — unrelated.)

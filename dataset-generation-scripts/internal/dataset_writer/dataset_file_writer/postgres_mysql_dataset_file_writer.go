package dataset_writer

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	sapnhapmodels "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"

	"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
)

type PostgresMySQLDatasetFileWriter struct {
	OutputFilePath string
}

// region insert statement
const insertAdministrativeRegionTemplate string = "INSERT INTO administrative_regions(id,name,name_en,code_name,code_name_en) VALUES(%d,'%s','%s','%s','%s');"

// administrative_unit insert_statement
const insertAdministrativeUnitTemplate string = "INSERT INTO administrative_units(id,full_name,full_name_en,short_name,short_name_en,code_name,code_name_en) VALUES(%d,'%s','%s','%s','%s','%s','%s');"

// province insert statement
const insertProvinceTemplate string = "INSERT INTO provinces(code,name,name_en,full_name,full_name_en,code_name,administrative_unit_id,postal_code_prefix) VALUES"
const insertProvinceValueTemplate string = "('%s','%s','%s','%s','%s','%s',%d,%s)"

// ward insert statement
const insertWardTemplate string = "INSERT INTO wards(code,name,name_en,full_name,full_name_en,code_name,province_code,administrative_unit_id,postal_code) VALUES"
const insertDistrictWardValueTemplate string = "('%s','%s','%s','%s','%s','%s','%s',%d,%s)"

// GIS section
const insertProvinceGISTemplate string = "INSERT INTO gis_provinces(province_code, gis_server_id, area_km2, bbox, geom) VALUES ('%s','%s',%f,ST_GeomFromText('%s', 4326),ST_GeomFromText('%s', 4326));"
const insertWardGISTemplate string = "INSERT INTO gis_wards(ward_code, gis_server_id, area_km2, bbox, geom) VALUES"
const insertWardGISValueTemplate string = "('%s','%s',%f,ST_GeomFromText('%s', 4326),ST_GeomFromText('%s', 4326))"

const batchInsertItemSize int = 50

func (w *PostgresMySQLDatasetFileWriter) WriteToFile(
	regions []model.AdministrativeRegion,
	administrativeUnits []model.AdministrativeUnit,
	provinces []model.Province,
	wards []model.Ward) error {

	outputFilePath := w.OutputFilePath
	if strings.Contains(outputFilePath, "%s") {
		outputFilePath = fmt.Sprintf(outputFilePath, getFileTimeSuffix())
	}
	os.MkdirAll(filepath.Dir(outputFilePath), os.ModePerm)
	file, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Unable to write to file", err)
		panic(err)
	}

	dataWriter := bufio.NewWriter(file)
	dataWriter.WriteString("/* === Vietnamese Provinces Database Dataset for PostgreSQL/MySQL === */\n")
	dataWriter.WriteString(fmt.Sprintf("/* Created at:  %s */\n", time.Now().Format(time.RFC1123Z)))
	dataWriter.WriteString("/* Reference: https://github.com/thanglequoc/vietnamese-provinces-database */\n")
	dataWriter.WriteString("/* =============================================== */\n\n")

	dataWriter.WriteString("-- DATA for administrative_regions --\n")
	for _, r := range regions {
		insertLine := fmt.Sprintf(insertAdministrativeRegionTemplate,
			r.Id, r.Name, r.NameEn, r.CodeName, r.CodeNameEn)
		dataWriter.WriteString(insertLine + "\n")
	}
	dataWriter.WriteString("-- ----------------------------------\n\n")

	dataWriter.WriteString("-- DATA for administrative_units --\n")

	// Write for administrativeUnits
	for _, u := range administrativeUnits {
		insertLine := fmt.Sprintf(insertAdministrativeUnitTemplate,
			u.Id, u.FullName, u.FullNameEn, u.ShortName, u.ShortNameEn, u.CodeName, u.CodeNameEn)
		dataWriter.WriteString(insertLine + "\n")
	}
	dataWriter.WriteString("-- ----------------------------------\n\n")

	// variable to generate batch insert statement
	counter := 0
	isAppending := false

	dataWriter.WriteString("-- DATA for provinces --\n")
	for i, p := range provinces {
		if !isAppending {
			dataWriter.WriteString(insertProvinceTemplate + "\n")
		}
		dataWriter.WriteString(
			fmt.Sprintf(insertProvinceValueTemplate, p.Code, escapeSingleQuote(p.Name), escapeSingleQuote(p.NameEn), escapeSingleQuote(p.FullName),
				escapeSingleQuote(p.FullNameEn), p.CodeName, p.AdministrativeUnitId, nullableSQLString(p.PostalCodePrefix)))
		counter++

		// the batch insert statement batch reach limit, break and create a new batch insert statement
		if counter == batchInsertItemSize || i == len(provinces)-1 {
			isAppending = false
			dataWriter.WriteString(";\n\n")
			counter = 0 // reset counter
		} else {
			dataWriter.WriteString(",\n")
			isAppending = true
		}
	}
	dataWriter.WriteString("-- ----------------------------------\n\n")

	dataWriter.WriteString("-- DATA for wards --\n")
	counter = 0
	isAppending = false

	for i, w := range wards {
		if !isAppending {
			dataWriter.WriteString(insertWardTemplate + "\n")
		}
		dataWriter.WriteString(
			fmt.Sprintf(insertDistrictWardValueTemplate, w.Code, escapeSingleQuote(w.Name), escapeSingleQuote(w.NameEn), escapeSingleQuote(w.FullName),
				escapeSingleQuote(w.FullNameEn), w.CodeName, w.ProvinceCode, w.AdministrativeUnitId, nullableSQLString(w.PostalCode)))
		counter++

		// the batch insert statement batch reach limit, break and create a new batch insert statement
		if counter == batchInsertItemSize || i == len(wards)-1 {
			isAppending = false
			dataWriter.WriteString(";\n\n")
			counter = 0 // reset counter
		} else {
			dataWriter.WriteString(",\n")
			isAppending = true
		}
	}
	dataWriter.WriteString("-- ----------------------------------\n")
	dataWriter.WriteString("-- END OF SCRIPT FILE --\n")

	dataWriter.Flush()
	file.Close()

	return writePostgresMySQLReadme(filepath.Dir(outputFilePath), outputFilePath)
}

func (w *PostgresMySQLDatasetFileWriter) WriteGISDataToFile(sapNhapProvincesGIS []*sapnhapmodels.SapNhapSiteGeoUnit, sapNhapWardsGIS []*sapnhapmodels.SapNhapSiteGeoUnit) error {
	postgresGISDir := filepath.Join("./output/postgresql", "gis")
	mysqlGISDir := filepath.Join("./output/mysql", "gis")

	_ = os.MkdirAll(postgresGISDir, os.ModePerm)
	_ = os.MkdirAll(mysqlGISDir, os.ModePerm)

	postgresGISFilePath := filepath.Join(postgresGISDir, "postgresql_ImportData_gis.sql")
	mysqlGISFilePath := filepath.Join(mysqlGISDir, "mysql_ImportData_gis.sql")

	createdAt := time.Now().Format(time.RFC1123Z)
	postgresHeader := chunkHeaderInfo{
		Banner:     "Add-on GIS Dataset for PostgreSQL of Vietnamese Provinces Database",
		CreatedAt:  createdAt,
		Repository: "https://github.com/thanglequoc/vietnamese-provinces-database",
	}
	mysqlHeader := chunkHeaderInfo{
		Banner:     "Add-on GIS Dataset for MySQL of Vietnamese Provinces Database",
		CreatedAt:  createdAt,
		Repository: "https://github.com/thanglequoc/vietnamese-provinces-database",
	}

	var postgresBlocks [][]byte
	var mysqlBlocks [][]byte

	postgresBlocks = append(postgresBlocks, []byte("-- DATA for gis_provinces --\n"))
	mysqlBlocks = append(mysqlBlocks, []byte("-- DATA for gis_provinces --\n"))

	for _, p := range sapNhapProvincesGIS {
		vnProvinceCode := p.VNDSProvinceCode

		// Postgres - PostGIS use OGC (Open Geospatial Consortium) standard (lng - lat)
		postgresInsertLine := fmt.Sprintf(insertProvinceGISTemplate+"\n",
			vnProvinceCode, p.MaLK, p.DienTichKM2, p.BBoxWKT, p.GeomWKT)
		postgresBlocks = append(postgresBlocks, []byte(postgresInsertLine))

		// MySQL uses official EPSG standard (lat - lng)
		mysqlInsertLine := fmt.Sprintf(insertProvinceGISTemplate+"\n",
			vnProvinceCode, p.MaLK, p.DienTichKM2, p.BBoxWKTLatLng, p.GeomWKTLatLng)
		mysqlBlocks = append(mysqlBlocks, []byte(mysqlInsertLine))
	}

	postgresBlocks = append(postgresBlocks, []byte("-- ----------------------------------\n\n-- DATA for gis_wards --\n"))
	mysqlBlocks = append(mysqlBlocks, []byte("-- ----------------------------------\n\n-- DATA for gis_wards --\n"))

	counter := 0
	postgresBatch := strings.Builder{}
	mysqlBatch := strings.Builder{}
	for i, w := range sapNhapWardsGIS {
		if counter == 0 {
			postgresBatch.WriteString(insertWardGISTemplate + "\n")
			mysqlBatch.WriteString(insertWardGISTemplate + "\n")
		}

		vnWardCode := w.VNDSWardCode
		postgresInsertLine := fmt.Sprintf(insertWardGISValueTemplate+"\n",
			vnWardCode, w.MaLK, w.DienTichKM2, w.BBoxWKT, w.GeomWKT)
		postgresBatch.WriteString(postgresInsertLine)

		mysqlInsertLine := fmt.Sprintf(insertWardGISValueTemplate+"\n",
			vnWardCode, w.MaLK, w.DienTichKM2, w.BBoxWKTLatLng, w.GeomWKTLatLng)
		mysqlBatch.WriteString(mysqlInsertLine)

		counter++
		if counter == batchInsertItemSize || i == len(sapNhapWardsGIS)-1 {
			postgresBatch.WriteString(";\n\n")
			mysqlBatch.WriteString(";\n\n")
			postgresBlocks = append(postgresBlocks, []byte(postgresBatch.String()))
			mysqlBlocks = append(mysqlBlocks, []byte(mysqlBatch.String()))
			counter = 0
			postgresBatch.Reset()
			mysqlBatch.Reset()
		} else {
			postgresBatch.WriteString(",\n")
			mysqlBatch.WriteString(",\n")
		}
	}

	postgresBlocks = append(postgresBlocks, []byte("-- ----------------------------------\n\n-- END OF SCRIPT FILE --\n"))
	mysqlBlocks = append(mysqlBlocks, []byte("-- ----------------------------------\n\n-- END OF SCRIPT FILE --\n"))

	if err := writeChunkedSQLFile(postgresGISFilePath, postgresBlocks, postgresHeader); err != nil {
		return fmt.Errorf("write postgres GIS chunks: %w", err)
	}
	if err := writeChunkedSQLFile(mysqlGISFilePath, mysqlBlocks, mysqlHeader); err != nil {
		return fmt.Errorf("write mysql GIS chunks: %w", err)
	}
	return nil
}

// writePostgresMySQLReadme writes the dataset README for the PostgreSQL or
// MySQL export. isPostgres is inferred from the output file path.
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

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

type MssqlDatasetFileWriter struct {
	OutputFilePath string
}

// region insert statement
const insertAdministrativeRegionTemplateMsSql string = "INSERT INTO administrative_regions(id,name,name_en,code_name,code_name_en) VALUES(%d,N'%s',N'%s',N'%s',N'%s');"

// administrative_unit insert_statement
const insertAdministrativeUnitMsSqlTemplate string = "INSERT INTO administrative_units(id,full_name,full_name_en,short_name,short_name_en,code_name,code_name_en) VALUES(%d,N'%s',N'%s',N'%s',N'%s',N'%s',N'%s');"

// province insert statement
const insertProvinceValueMsSqlTemplate string = "('%s',N'%s',N'%s',N'%s',N'%s','%s',%d,%s)"
const insertProvinceWardValueMsSqlTemplate string = "('%s',N'%s',N'%s',N'%s',N'%s','%s','%s',%d,%s)"

// GIS section
const insertMssqlGISProvinceTemplate string = "INSERT INTO gis_provinces(province_code, gis_server_id, area_km2, bbox, geom) VALUES ('%s','%s',%f,geometry::STGeomFromText('%s', 4326),geometry::STGeomFromText('%s', 4326));"
const insertMssqlGISWardTemplate string = "INSERT INTO gis_wards(ward_code, gis_server_id, area_km2, bbox, geom) VALUES"
const insertMssqlGISWardValueTemplate string = "('%s','%s',%f,geometry::STGeomFromText('%s', 4326),geometry::STGeomFromText('%s', 4326))"

func (w *MssqlDatasetFileWriter) WriteToFile(
	regions []model.AdministrativeRegion,
	administrativeUnits []model.AdministrativeUnit,
	provinces []model.Province,
	wards []model.Ward) error {

	outputFilePath := w.OutputFilePath
	if strings.Contains(outputFilePath, "%s") {
		outputFilePath = fmt.Sprintf(outputFilePath, getFileTimeSuffix())
	}
	os.MkdirAll(filepath.Dir(outputFilePath), os.ModePerm)

	fileMsSql, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Unable to write to file", err)
		panic(err)
	}

	dataWriterMsSql := bufio.NewWriter(fileMsSql)
	dataWriterMsSql.WriteString("/* === Vietnamese Provinces Database Dataset for Microsoft SQL Server === */\n")
	dataWriterMsSql.WriteString(fmt.Sprintf("/* Created at:  %s */\n", time.Now().Format(time.RFC1123Z)))
	dataWriterMsSql.WriteString("/* Reference: https://github.com/thanglequoc/vietnamese-provinces-database */\n")
	dataWriterMsSql.WriteString("/* =============================================== */\n\n")

	dataWriterMsSql.WriteString("-- DATA for administrative_regions --\n")
	for _, r := range regions {
		insertLineMsSql := fmt.Sprintf(insertAdministrativeRegionTemplateMsSql,
			r.Id, r.Name, r.NameEn, r.CodeName, r.CodeNameEn)
		dataWriterMsSql.WriteString(insertLineMsSql + "\n")
	}
	dataWriterMsSql.WriteString("-- ----------------------------------\n\n")

	dataWriterMsSql.WriteString("-- DATA for administrative_units --\n")

	for _, u := range administrativeUnits {
		insertLineMsSql := fmt.Sprintf(insertAdministrativeUnitMsSqlTemplate,
			u.Id, u.FullName, u.FullNameEn, u.ShortName, u.ShortNameEn, u.CodeName, u.CodeNameEn)
		dataWriterMsSql.WriteString(insertLineMsSql + "\n")
	}
	dataWriterMsSql.WriteString("-- ----------------------------------\n\n")

	// variable to generate batch insert statement
	counter := 0
	isAppending := false

	dataWriterMsSql.WriteString("-- DATA for provinces --\n")
	for i, p := range provinces {
		if !isAppending {
			dataWriterMsSql.WriteString(insertProvinceTemplate + "\n")
		}
		dataWriterMsSql.WriteString(
			fmt.Sprintf(insertProvinceValueMsSqlTemplate, p.Code, escapeSingleQuote(p.Name), escapeSingleQuote(p.NameEn), escapeSingleQuote(p.FullName),
				escapeSingleQuote(p.FullNameEn), p.CodeName, p.AdministrativeUnitId, nullableNString(p.PostalCodePrefix)))
		counter++

		// the batch insert statement batch reach limit, break and create a new batch insert statement
		if counter == batchInsertItemSize || i == len(provinces)-1 {
			isAppending = false
			dataWriterMsSql.WriteString(";\n\n")
			counter = 0 // reset counter
		} else {
			dataWriterMsSql.WriteString(",\n")
			isAppending = true
		}
	}
	dataWriterMsSql.WriteString("-- ----------------------------------\n\n")

	dataWriterMsSql.WriteString("-- DATA for wards --\n")
	counter = 0
	isAppending = false
	for i, w := range wards {
		if !isAppending {
			dataWriterMsSql.WriteString(insertWardTemplate + "\n")
		}
		dataWriterMsSql.WriteString(
			fmt.Sprintf(insertProvinceWardValueMsSqlTemplate, w.Code, escapeSingleQuote(w.Name), escapeSingleQuote(w.NameEn), escapeSingleQuote(w.FullName),
				escapeSingleQuote(w.FullNameEn), w.CodeName, w.ProvinceCode, w.AdministrativeUnitId, nullableNString(w.PostalCode)))
		counter++

		// the batch insert statement batch reach limit, break and create a new batch insert statement
		if counter == batchInsertItemSize || i == len(wards)-1 {
			isAppending = false
			dataWriterMsSql.WriteString(";\n\n")
			counter = 0 // reset counter
		} else {
			dataWriterMsSql.WriteString(",\n")
			isAppending = true
		}
	}
	dataWriterMsSql.WriteString("-- ----------------------------------\n")
	dataWriterMsSql.WriteString("-- END OF SCRIPT FILE --\n")
	dataWriterMsSql.Flush()
	fileMsSql.Close()

	return writeMssqlReadme(filepath.Dir(outputFilePath))
}

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

func (w *MssqlDatasetFileWriter) WriteGISDataToFile(sapNhapProvincesGIS []*sapnhapmodels.SapNhapSiteGeoUnit, sapNhapWardsGIS []*sapnhapmodels.SapNhapSiteGeoUnit) error {
	gisOutputFolderPath := "./output/sqlserver/gis"
	if err := os.MkdirAll(gisOutputFolderPath, os.ModePerm); err != nil {
		return fmt.Errorf("create output folder %s: %w", gisOutputFolderPath, err)
	}

	mssqlGISFilePath := gisOutputFolderPath + "/mssql_ImportData_gis.sql"

	header := chunkHeaderInfo{
		Banner:     "Add-on GIS Dataset for Microsoft SQL Server of Vietnamese Provinces Database",
		CreatedAt:  time.Now().Format(time.RFC1123Z),
		Repository: "https://github.com/thanglequoc/vietnamese-provinces-database",
	}

	var blocks [][]byte
	blocks = append(blocks, []byte("-- DATA for gis_provinces --\n"))
	for _, p := range sapNhapProvincesGIS {
		vnProvinceCode := p.VNDSProvinceCode
		mssqlInsertLine := fmt.Sprintf(insertMssqlGISProvinceTemplate+"\n",
			vnProvinceCode, p.MaLK, p.DienTichKM2, p.BBoxWKT, p.GeomWKT)
		blocks = append(blocks, []byte(mssqlInsertLine))
	}
	blocks = append(blocks, []byte("-- ----------------------------------\n\n-- DATA for gis_wards --\n"))

	counter := 0
	batch := strings.Builder{}
	for i, w := range sapNhapWardsGIS {
		if counter == 0 {
			batch.WriteString(insertMssqlGISWardTemplate + "\n")
		}

		vnWardCode := w.VNDSWardCode
		mssqlInsertLine := fmt.Sprintf(insertMssqlGISWardValueTemplate,
			vnWardCode, w.MaLK, w.DienTichKM2, w.BBoxWKT, w.GeomWKT)
		batch.WriteString(mssqlInsertLine)

		counter++
		if counter == batchInsertItemSize || i == len(sapNhapWardsGIS)-1 {
			batch.WriteString(";\n\nGO\n\n")
			blocks = append(blocks, []byte(batch.String()))
			counter = 0
			batch.Reset()
		} else {
			batch.WriteString(",\n")
		}
	}

	blocks = append(blocks, []byte("-- ----------------------------------\n\n-- END OF SCRIPT FILE --\n"))

	if err := writeChunkedSQLFile(mssqlGISFilePath, blocks, header); err != nil {
		return fmt.Errorf("write mssql GIS chunks: %w", err)
	}
	return nil
}

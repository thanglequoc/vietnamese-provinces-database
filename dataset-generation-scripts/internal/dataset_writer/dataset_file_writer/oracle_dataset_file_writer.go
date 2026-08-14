package dataset_writer

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
)

const insertProvinceOracleTemplate string = "\tINTO provinces(code,name,name_en,full_name,full_name_en,code_name,administrative_unit_id,postal_code_prefix) VALUES('%s','%s','%s','%s','%s','%s',%d,%s)"

type OracleDatasetFileWriter struct {
	OutputFilePath string
}

func (w *OracleDatasetFileWriter) WriteToFile(
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
	dataWriter.WriteString("/* === Vietnamese Provinces Database Dataset for Oracle === */\n")
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
			dataWriter.WriteString("INSERT ALL\n")
		}
		dataWriter.WriteString(
			fmt.Sprintf(insertProvinceOracleTemplate, p.Code, escapeSingleQuote(p.Name), escapeSingleQuote(p.NameEn), escapeSingleQuote(p.FullName),
				escapeSingleQuote(p.FullNameEn), p.CodeName, p.AdministrativeUnitId, nullableSQLString(p.PostalCodePrefix)))
		counter++

		// the batch insert statement batch reach limit, break and create a new batch insert statement
		// In oracle, the last insert batch statement require a dummy select after multiple INSERT ALL INTO statements
		if counter == batchInsertItemSize || i == len(provinces)-1 {
			isAppending = false
			dataWriter.WriteString("\n\tSELECT 1 FROM DUAL;\n\n")
			counter = 0 // reset counter
		} else {
			dataWriter.WriteString("\n")
			isAppending = true
		}
	}
	dataWriter.WriteString("-- ----------------------------------\n\n")

	// ward insert statement
	const insertWardOracleTemplate string = "\tINTO wards(code,name,name_en,full_name,full_name_en,code_name,province_code,administrative_unit_id,postal_code) VALUES('%s','%s','%s','%s','%s','%s','%s',%d,%s)"

	dataWriter.WriteString("-- DATA for wards --\n")
	counter = 0
	isAppending = false
	for i, d := range wards {
		if !isAppending {
			dataWriter.WriteString("INSERT ALL\n")
		}
		dataWriter.WriteString(
			fmt.Sprintf(insertWardOracleTemplate, d.Code, escapeSingleQuote(d.Name), escapeSingleQuote(d.NameEn), escapeSingleQuote(d.FullName),
				escapeSingleQuote(d.FullNameEn), d.CodeName, d.ProvinceCode, d.AdministrativeUnitId, nullableSQLString(d.PostalCode)))
		counter++

		// the batch insert statement batch reach limit, break and create a new batch insert statement
		// In oracle, the last insert batch statement require a dummy select after multiple INSERT ALL INTO statements
		if counter == batchInsertItemSize || i == len(wards)-1 {
			isAppending = false
			dataWriter.WriteString("\n\tSELECT 1 FROM DUAL;\n\n")
			counter = 0 // reset counter
		} else {
			dataWriter.WriteString("\n")
			isAppending = true
		}
	}
	dataWriter.WriteString("-- ----------------------------------\n\n")

	dataWriter.WriteString("-- END OF SCRIPT FILE --\n")
	dataWriter.Flush()
	file.Close()

	return writeOracleReadme(filepath.Dir(outputFilePath))
}

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

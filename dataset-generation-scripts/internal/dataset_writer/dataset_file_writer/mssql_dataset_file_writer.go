package dataset_writer

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"
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
const insertProvinceValueMsSqlTemplate string = "('%s',N'%s',N'%s',N'%s',N'%s','%s',%d)"

const insertProvinceWardValueMsSqlTemplate string = "('%s',N'%s',N'%s',N'%s',N'%s','%s','%s',%d)"

func (w *MssqlDatasetFileWriter) WriteToFile(
	regions []model.AdministrativeRegion,
	administrativeUnits []model.AdministrativeUnit,
	provinces []model.Province,
	wards []model.Ward) error {

	fileTimeSuffix := getFileTimeSuffix()
	outputFilePath := fmt.Sprintf(w.OutputFilePath, fileTimeSuffix)

	fileMsSql, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Unable to write to file", err)
		panic(err)
	}

	dataWriterMsSql := bufio.NewWriter(fileMsSql)
	dataWriterMsSql.WriteString("/* === Vietnamese Provinces Database Dataset for Microsoft SQL Server === */\n")
	dataWriterMsSql.WriteString(fmt.Sprintf("/* Created at:  %s */\n", time.Now().Format(time.RFC1123Z)))
	dataWriterMsSql.WriteString("/* Reference: https://github.com/ThangLeQuoc/vietnamese-provinces-database */\n")
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
				escapeSingleQuote(p.FullNameEn), p.CodeName, p.AdministrativeUnitId))
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
				escapeSingleQuote(w.FullNameEn), w.CodeName, w.ProvinceCode, w.AdministrativeUnitId))
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

	return nil
}

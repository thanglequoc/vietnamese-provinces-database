package dataset_writer

import (
	"bufio"
	"fmt"
	"log"
	"os"
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
const insertProvinceTemplate string = "INSERT INTO provinces(code,name,name_en,full_name,full_name_en,code_name,administrative_unit_id) VALUES"
const insertProvinceValueTemplate string = "('%s','%s','%s','%s','%s','%s',%d)"

// ward insert statement
const insertWardTemplate string = "INSERT INTO wards(code,name,name_en,full_name,full_name_en,code_name,province_code,administrative_unit_id) VALUES"
const insertDistrictWardValueTemplate string = "('%s','%s','%s','%s','%s','%s','%s',%d)"

// GIS section
const insertProvinceGISTemplate string = "INSERT INTO gis_provinces(province_code, gis_server_id, area_km2, bbox, geom) VALUES ('%s','%s',%f,ST_GeomFromText('%s', 4326),ST_GeomFromText('%s', 4326));"
const insertWardGISTemplate string = "INSERT INTO gis_wards(ward_code, gis_server_id, area_km2, bbox, geom) VALUES ('%s','%s',%f,ST_GeomFromText('%s', 4326),ST_GeomFromText('%s', 4326));"

const batchInsertItemSize int = 50

func (w *PostgresMySQLDatasetFileWriter) WriteToFile(
	regions []model.AdministrativeRegion,
	administrativeUnits []model.AdministrativeUnit,
	provinces []model.Province,
	wards []model.Ward) error {

	fileTimeSuffix := getFileTimeSuffix()
	outputFilePath := fmt.Sprintf(w.OutputFilePath, fileTimeSuffix)
	file, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Unable to write to file", err)
		panic(err)
	}

	dataWriter := bufio.NewWriter(file)
	dataWriter.WriteString("/* === Vietnamese Provinces Database Dataset for PostgreSQL/MySQL === */\n")
	dataWriter.WriteString(fmt.Sprintf("/* Created at:  %s */\n", time.Now().Format(time.RFC1123Z)))
	dataWriter.WriteString("/* Reference: https://github.com/ThangLeQuoc/vietnamese-provinces-database */\n")
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
				escapeSingleQuote(p.FullNameEn), p.CodeName, p.AdministrativeUnitId))
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
				escapeSingleQuote(w.FullNameEn), w.CodeName, w.ProvinceCode, w.AdministrativeUnitId))
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
	return nil
}

func (w *PostgresMySQLDatasetFileWriter) WriteGISDataToFile(sapNhapProvincesGIS []sapnhapmodels.SapNhapProvinceGIS, sapNhapWardsGIS []sapnhapmodels.SapNhapWardGIS) error {
	fileTimeSuffix := getFileTimeSuffix()

	postgresMySQLGISOutputFolderPath := "./output/gis"
	err := os.MkdirAll(postgresMySQLGISOutputFolderPath, os.ModePerm)

	postgresGISFilePath := fmt.Sprintf(postgresMySQLGISOutputFolderPath+"/postgresql_ImportData_gis_%s.sql", fileTimeSuffix)
	provinceGISFile, err := os.OpenFile(postgresGISFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Unable to write to file", err)
		panic(err)
	}
	defer provinceGISFile.Close()

	mysqlGISFilePath := fmt.Sprintf(postgresMySQLGISOutputFolderPath+"/mysql_ImportData_gis_%s.sql", fileTimeSuffix)
	mysqlGISFile, err := os.OpenFile(mysqlGISFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Unable to write to file", err)
		panic(err)
	}
	defer mysqlGISFile.Close()

	postgresScriptDataWriter := bufio.NewWriter(provinceGISFile)
	postgresScriptDataWriter.WriteString("/* === Add-on GIS Dataset for PostgreSQL of Vietnamese Provinces Database  === */\n")
	postgresScriptDataWriter.WriteString(fmt.Sprintf("/* Created at:  %s */\n", time.Now().Format(time.RFC1123Z)))
	postgresScriptDataWriter.WriteString("/* Reference: https://github.com/ThangLeQuoc/vietnamese-provinces-database */\n")
	postgresScriptDataWriter.WriteString("/* =============================================== */\n\n")

	mysqlScriptDataWriter := bufio.NewWriter(mysqlGISFile)
	mysqlScriptDataWriter.WriteString("/* === Add-on GIS Dataset for MySQL of Vietnamese Provinces Database  === */\n")
	mysqlScriptDataWriter.WriteString(fmt.Sprintf("/* Created at:  %s */\n", time.Now().Format(time.RFC1123Z)))
	mysqlScriptDataWriter.WriteString("/* Reference: https://github.com/ThangLeQuoc/vietnamese-provinces-database */\n")
	mysqlScriptDataWriter.WriteString("/* =============================================== */\n\n")

	postgresScriptDataWriter.WriteString("-- DATA for gis_provinces --\n")
	for _, p := range sapNhapProvincesGIS {
		areaKm2, err := parseEuropeanFloat(p.SapNhapSiteProvince.DienTichKm2)
		vnProvinceCode := p.SapNhapSiteProvince.VNProvinceCode
		if err != nil {
			log.Panicf("Unable to parse area km2 for province %s, value: %s", vnProvinceCode, p.SapNhapSiteProvince.DienTichKm2)
		}

		// Postgres - Postgis use OGC (Open Geospatial Consortium) standard (lng - lat)
		postgresInsertLine := fmt.Sprintf(insertProvinceGISTemplate + "\n",
			vnProvinceCode, p.GISServerID, areaKm2, p.BBoxWKT, p.GeomWKT)
		postgresScriptDataWriter.WriteString(postgresInsertLine)

		// MySQL use official EPSG standard (lat - lng)
		mysqlInsertLine := fmt.Sprintf(insertProvinceGISTemplate + "\n",
			vnProvinceCode, p.GISServerID, areaKm2, p.BBoxWKTLatLng, p.GeomWKTLatLng)
		mysqlScriptDataWriter.WriteString(mysqlInsertLine)
	}

	postgresScriptDataWriter.WriteString("-- ----------------------------------\n\n")
	mysqlScriptDataWriter.WriteString("-- ----------------------------------\n\n")

	postgresScriptDataWriter.WriteString("-- DATA for gis_wards --\n")
	mysqlScriptDataWriter.WriteString("-- DATA for gis_wards --\n")

	for _, w := range sapNhapWardsGIS {
		vnWardCode := w.SapNhapSiteWard.VNWardCode
		postgresInsertLine := fmt.Sprintf(insertWardGISTemplate + "\n",
			vnWardCode, w.GISServerID, w.SapNhapSiteWard.DienTichKm2, w.BBoxWKT, w.GeomWKT)
		
		postgresScriptDataWriter.WriteString(postgresInsertLine)

		mysqlInsertLine := fmt.Sprintf(insertWardGISTemplate + "\n",
			vnWardCode, w.GISServerID, w.SapNhapSiteWard.DienTichKm2, w.BBoxWKTLatLng, w.GeomWKTLatLng)
		mysqlScriptDataWriter.WriteString(mysqlInsertLine)
	}
	postgresScriptDataWriter.WriteString("-- ----------------------------------\n\n")
	postgresScriptDataWriter.WriteString("-- END OF SCRIPT FILE --\n")

	mysqlScriptDataWriter.WriteString("-- ----------------------------------\n\n")
	mysqlScriptDataWriter.WriteString("-- END OF SCRIPT FILE --\n")
	
	postgresScriptDataWriter.Flush()
	mysqlScriptDataWriter.Flush()

	return nil
}

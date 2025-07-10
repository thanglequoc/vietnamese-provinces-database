package dataset_writer

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	vn_common "github.com/thanglequoc-vn-provinces/v2/common"
	"github.com/thanglequoc-vn-provinces/v2/dataset_writer/dataset_file_writer/helper"
	gis "github.com/thanglequoc-vn-provinces/v2/gis"
)

const insertProvinceOracleTemplate string = "\tINTO provinces(code,name,name_en,full_name,full_name_en,code_name,administrative_unit_id,administrative_region_id) VALUES('%s','%s','%s','%s','%s','%s',%d,%d)"

const insertOracleGISTemplate string = "INSERT INTO vn_gis(code,unit_level,bbox, temp_wkt) VALUES "
const insertOracleGISValueTemplate string = "('%s','%s', sde.st_geomfromtext('POLYGON((%s, %s, %s, %s, %s))', 4326) ,%s);\n"

type OracleDatasetFileWriter struct {
	OutputFilePath string
}

func (w *OracleDatasetFileWriter) WriteToFile(
	regions []vn_common.AdministrativeRegion,
	administrativeUnits []vn_common.AdministrativeUnit,
	provinces []vn_common.Province,
	districts []vn_common.District,
	wards []vn_common.Ward,
	gisData []gis.ProvinceGIS) error {

	fileTimeSuffix := getFileTimeSuffix()
	outputFilePath := fmt.Sprintf(w.OutputFilePath, fileTimeSuffix)

	file, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Unable to write to file", err)
		panic(err)
	}

	dataWriter := bufio.NewWriter(file)
	dataWriter.WriteString("/* === Vietnamese Provinces Database Dataset for Oracle === */\n")
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
				escapeSingleQuote(p.FullNameEn), p.CodeName, p.AdministrativeUnitId, p.AdministrativeRegionId))
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

	const insertDistrictOracleTemplate string = "\tINTO districts(code,name,name_en,full_name,full_name_en,code_name,province_code,administrative_unit_id) VALUES('%s','%s','%s','%s','%s','%s','%s',%d)"
	dataWriter.WriteString("-- DATA for districts --\n")
	counter = 0
	isAppending = false
	for i, d := range districts {
		if !isAppending {
			dataWriter.WriteString("INSERT ALL\n")
		}
		dataWriter.WriteString(
			fmt.Sprintf(insertDistrictOracleTemplate, d.Code, escapeSingleQuote(d.Name), escapeSingleQuote(d.NameEn), escapeSingleQuote(d.FullName),
				escapeSingleQuote(d.FullNameEn), d.CodeName, d.ProvinceCode, d.AdministrativeUnitId))
		counter++

		// the batch insert statement batch reach limit, break and create a new batch insert statement
		// In oracle, the last insert batch statement require a dummy select after multiple INSERT ALL INTO statements
		if counter == batchInsertItemSize || i == len(districts)-1 {
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
	const insertWardOracleTemplate string = "\tINTO wards(code,name,name_en,full_name,full_name_en,code_name,district_code,administrative_unit_id) VALUES('%s','%s','%s','%s','%s','%s','%s',%d)"

	dataWriter.WriteString("-- DATA for wards --\n")
	counter = 0
	isAppending = false
	for i, d := range wards {
		if !isAppending {
			dataWriter.WriteString("INSERT ALL\n")
		}
		dataWriter.WriteString(
			fmt.Sprintf(insertWardOracleTemplate, d.Code, escapeSingleQuote(d.Name), escapeSingleQuote(d.NameEn), escapeSingleQuote(d.FullName),
				escapeSingleQuote(d.FullNameEn), d.CodeName, d.DistrictCode, d.AdministrativeUnitId))
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

	writeOracleGISDataToFile(gisData)
	return nil
}

func writeOracleGISDataToFile(gisData []gis.ProvinceGIS) {
	fileTimeSuffix := getFileTimeSuffix()
	outputFilePath := fmt.Sprintf("./output/oracle_generated_ImportGISData_vn_units_%s.sql", fileTimeSuffix)
	file, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Unable to write to file", err)
		panic(err)
	}
	defer file.Close()

	dataWriter := bufio.NewWriter(file)
	dataWriter.WriteString("/* === Vietnamese Provinces Database - GIS Dataset for Oracle === */\n")
	dataWriter.WriteString(fmt.Sprintf("/* Created at:  %s */\n", time.Now().Format(time.RFC1123Z)))
	dataWriter.WriteString("/* Reference: https://github.com/ThangLeQuoc/vietnamese-provinces-database */\n")
	dataWriter.WriteString("/* =============================================== */\n\n")
	
	// ring order: bottom left -> going around -> bottom left again
	for _, g := range gisData {
		dataWriter.WriteString(insertOracleGISTemplate)
		dataWriter.WriteString(fmt.Sprintf(insertOracleGISValueTemplate, 
			g.LevelId, 
			"province",
			gis.ToCoordinateStr(g.BBox.BottomLeftLng, g.BBox.BottomLeftLat),
			gis.ToCoordinateStr(g.BBox.TopLeftLng, g.BBox.TopLeftLat),
			gis.ToCoordinateStr(g.BBox.TopRightLng, g.BBox.TopRightLat),
			gis.ToCoordinateStr(g.BBox.BottomRightLng, g.BBox.BottomRightLat),
			gis.ToCoordinateStr(g.BBox.BottomLeftLng, g.BBox.BottomLeftLat), // repeat the bottom left again to close the polygon ring,
			toOracleMultiPolygon(g.Coordinates[0][0]),
			),
		)

		// district in province
		for _, gd := range g.Districts {
			dataWriter.WriteString(insertOracleGISTemplate)
			dataWriter.WriteString(fmt.Sprintf(insertOracleGISValueTemplate, 
				gd.LevelId, 
				"district",
				gis.ToCoordinateStr(gd.BBox.BottomLeftLng, gd.BBox.BottomLeftLat),
				gis.ToCoordinateStr(gd.BBox.TopLeftLng, gd.BBox.TopLeftLat),
				gis.ToCoordinateStr(gd.BBox.TopRightLng, gd.BBox.TopRightLat),
				gis.ToCoordinateStr(gd.BBox.BottomRightLng, gd.BBox.BottomRightLat),
				gis.ToCoordinateStr(gd.BBox.BottomLeftLng, gd.BBox.BottomLeftLat), // repeat the bottom left again to close the polygon ring,
				toOracleMultiPolygon(gd.Coordinates[0][0]),
				),
			)
		}
	}

	dataWriter.Flush()
}

func toOracleMultiPolygon(ringCoordinates gis.GisLinearRingCoordinate) string {
	var template = "MULTIPOLYGON(((%s)))"

	var sb strings.Builder
	for _, p := range ringCoordinates.GisPoints {
		sb.WriteString(gis.ToCoordinateStr(p.Longitude, p.Latitude))
		sb.WriteString(",")
	}
	// repeat the first point to close the ring
	firstPoint := ringCoordinates.GisPoints[0]
	sb.WriteString(gis.ToCoordinateStr(firstPoint.Longitude, firstPoint.Latitude))
	return formatToClobAppend(fmt.Sprintf(template, sb.String()))
}

// Insert into table (clob_column) values ( to_clob( 'chunk 1' ) || to_clob( 'chunk 2' ) );
func formatToClobAppend(gisStr string) string {
	chunks := helper.Chunks(gisStr, 3000)
	for i, v := range chunks {
		chunks[i] = "to_clob('" + v + "')"
	}
	return strings.Join(chunks, " || ")
}

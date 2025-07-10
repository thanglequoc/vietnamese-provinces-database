package dataset_writer

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	vn_common "github.com/thanglequoc-vn-provinces/v2/common"
	gis "github.com/thanglequoc-vn-provinces/v2/gis"
)

type MssqlDatasetFileWriter struct {
	OutputFilePath string
}

// region insert statement
const insertAdministrativeRegionTemplateMsSql string = "INSERT INTO administrative_regions(id,name,name_en,code_name,code_name_en) VALUES(%d,N'%s',N'%s',N'%s',N'%s');"

// administrative_unit insert_statement
const insertAdministrativeUnitMsSqlTemplate string = "INSERT INTO administrative_units(id,full_name,full_name_en,short_name,short_name_en,code_name,code_name_en) VALUES(%d,N'%s',N'%s',N'%s',N'%s',N'%s',N'%s');"

// province insert statement
const insertProvinceValueMsSqlTemplate string = "('%s',N'%s',N'%s',N'%s',N'%s','%s',%d,%d)"

const insertDistrictWardValueMsSqlTemplate string = "('%s',N'%s',N'%s',N'%s',N'%s','%s','%s',%d)"


const insertMsSQLGISTemplate string = "INSERT INTO vn_gis(code,level,bbox, gis_geom) VALUES "
const insertMsSQLGISValueTemplate string = "('%s','%s',geography::STGeomFromText('POLYGON((%s, %s, %s, %s, %s))', 4326), geography::STGeomFromText('MULTIPOLYGON(((%s)))', 4326));\n"

func (w *MssqlDatasetFileWriter) WriteToFile(
	regions []vn_common.AdministrativeRegion,
	administrativeUnits []vn_common.AdministrativeUnit,
	provinces []vn_common.Province,
	districts []vn_common.District,
	wards []vn_common.Ward,
	gisData []gis.ProvinceGIS) error {

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
				escapeSingleQuote(p.FullNameEn), p.CodeName, p.AdministrativeUnitId, p.AdministrativeRegionId))
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

	dataWriterMsSql.WriteString("-- DATA for districts --\n")
	counter = 0
	isAppending = false
	for i, d := range districts {
		if !isAppending {
			dataWriterMsSql.WriteString(insertDistrictTemplate + "\n")
		}
		dataWriterMsSql.WriteString(
			fmt.Sprintf(insertDistrictWardValueMsSqlTemplate, d.Code, escapeSingleQuote(d.Name), escapeSingleQuote(d.NameEn), escapeSingleQuote(d.FullName),
				escapeSingleQuote(d.FullNameEn), d.CodeName, d.ProvinceCode, d.AdministrativeUnitId))
		counter++

		// the batch insert statement batch reach limit, break and create a new batch insert statement
		if counter == batchInsertItemSize || i == len(districts)-1 {
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
			fmt.Sprintf(insertDistrictWardValueMsSqlTemplate, w.Code, escapeSingleQuote(w.Name), escapeSingleQuote(w.NameEn), escapeSingleQuote(w.FullName),
				escapeSingleQuote(w.FullNameEn), w.CodeName, w.DistrictCode, w.AdministrativeUnitId))
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

	writeMsSQLGISDataToFile(gisData)

	return nil
}

func writeMsSQLGISDataToFile(gisData []gis.ProvinceGIS) {
	fileTimeSuffix := getFileTimeSuffix()
	outputFilePath := fmt.Sprintf("./output/mssql_generated_ImportGISData_vn_units_%s.sql", fileTimeSuffix)
	file, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Unable to write to file", err)
		panic(err)
	}
	defer file.Close()

	dataWriter := bufio.NewWriter(file)
	dataWriter.WriteString("/* === Vietnamese Provinces Database - GIS Dataset for Microsoft SQL Server === */\n")
	dataWriter.WriteString(fmt.Sprintf("/* Created at:  %s */\n", time.Now().Format(time.RFC1123Z)))
	dataWriter.WriteString("/* Reference: https://github.com/ThangLeQuoc/vietnamese-provinces-database */\n")
	dataWriter.WriteString("/* =============================================== */\n\n")
	
	// ring order: bottom left -> going around -> bottom left again
	for _, g := range gisData {
		dataWriter.WriteString(insertMsSQLGISTemplate)
		dataWriter.WriteString(fmt.Sprintf(insertMsSQLGISValueTemplate, 
			g.LevelId, 
			"province",
			gis.ToCoordinateStr(g.BBox.BottomLeftLng, g.BBox.BottomLeftLat),
			gis.ToCoordinateStr(g.BBox.TopLeftLng, g.BBox.TopLeftLat),
			gis.ToCoordinateStr(g.BBox.TopRightLng, g.BBox.TopRightLat),
			gis.ToCoordinateStr(g.BBox.BottomRightLng, g.BBox.BottomRightLat),
			gis.ToCoordinateStr(g.BBox.BottomLeftLng, g.BBox.BottomLeftLat), // repeat the bottom left again to close the polygon ring,
			toMsSQLMultiPolygon(g.Coordinates[0][0]),
			),
		)

		// district in province
		for _, gd := range g.Districts {
			dataWriter.WriteString(insertMySQLGISTemplate)
			dataWriter.WriteString(fmt.Sprintf(insertMsSQLGISValueTemplate, 
				gd.LevelId, 
				"district",
				gis.ToCoordinateStr(gd.BBox.BottomLeftLng, gd.BBox.BottomLeftLat),
				gis.ToCoordinateStr(gd.BBox.TopLeftLng, gd.BBox.TopLeftLat),
				gis.ToCoordinateStr(gd.BBox.TopRightLng, gd.BBox.TopRightLat),
				gis.ToCoordinateStr(gd.BBox.BottomRightLng, gd.BBox.BottomRightLat),
				gis.ToCoordinateStr(gd.BBox.BottomLeftLng, gd.BBox.BottomLeftLat), // repeat the bottom left again to close the polygon ring,
				toMsSQLMultiPolygon(gd.Coordinates[0][0]),
				),
			)
		}
	}

	dataWriter.Flush()
}

func toMsSQLMultiPolygon(ringCoordinates gis.GisLinearRingCoordinate) string {
	var sb strings.Builder
	for _, p := range ringCoordinates.GisPoints {
		sb.WriteString(gis.ToCoordinateStr(p.Longitude, p.Latitude))
		sb.WriteString(",")
	}
	// repeat the first point to close the ring
	firstPoint := ringCoordinates.GisPoints[0]
	sb.WriteString(gis.ToCoordinateStr(firstPoint.Longitude, firstPoint.Latitude))
	return sb.String()
}

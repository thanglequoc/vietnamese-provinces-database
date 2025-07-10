package dataset_writer

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"
	"strings"

	vn_common "github.com/thanglequoc-vn-provinces/v2/common"
	gis "github.com/thanglequoc-vn-provinces/v2/gis"
)

type PostgresMySQLDatasetFileWriter struct {
	OutputFilePath string
}

// region insert statement
const insertAdministrativeRegionTemplate string = "INSERT INTO administrative_regions(id,name,name_en,code_name,code_name_en) VALUES(%d,'%s','%s','%s','%s');"

// administrative_unit insert_statement
const insertAdministrativeUnitTemplate string = "INSERT INTO administrative_units(id,full_name,full_name_en,short_name,short_name_en,code_name,code_name_en) VALUES(%d,'%s','%s','%s','%s','%s','%s');"

// province insert statement
const insertProvinceTemplate string = "INSERT INTO provinces(code,name,name_en,full_name,full_name_en,code_name,administrative_unit_id,administrative_region_id) VALUES"
const insertProvinceValueTemplate string = "('%s','%s','%s','%s','%s','%s',%d,%d)"

// district insert statement
const insertDistrictTemplate string = "INSERT INTO districts(code,name,name_en,full_name,full_name_en,code_name,province_code,administrative_unit_id) VALUES"

// ward insert statement
const insertWardTemplate string = "INSERT INTO wards(code,name,name_en,full_name,full_name_en,code_name,district_code,administrative_unit_id) VALUES"
const insertDistrictWardValueTemplate string = "('%s','%s','%s','%s','%s','%s','%s',%d)"

// PostGis insert statement
const insertPostgresGISTemplate string = "INSERT INTO vn_gis(code,level,bbox, gis_geom) VALUES "
const insertPostgresGISValueTemplate string = "('%s','%s','POLYGON((%s, %s, %s, %s, %s))', 'MULTIPOLYGON(((%s)))');\n"

// Spartial MySQL insert statement
const insertMySQLGISTemplate string = "INSERT INTO vn_gis(code,level,bbox, gis_geom) VALUES "
const insertMySQLGISValueTemplate string = "('%s','%s',ST_GeomFromText('POLYGON((%s, %s, %s, %s, %s))', 4326), ST_GeomFromText('MULTIPOLYGON(((%s)))', 4326));\n"

const batchInsertItemSize int = 50

func (w *PostgresMySQLDatasetFileWriter) WriteToFile(
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
	defer file.Close()

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
	dataWriter.WriteString("------------------------------------\n\n")

	dataWriter.WriteString("-- DATA for administrative_units --\n")

	// Write for administrativeUnits
	for _, u := range administrativeUnits {
		insertLine := fmt.Sprintf(insertAdministrativeUnitTemplate,
			u.Id, u.FullName, u.FullNameEn, u.ShortName, u.ShortNameEn, u.CodeName, u.CodeNameEn)
		dataWriter.WriteString(insertLine + "\n")
	}
	dataWriter.WriteString("------------------------------------\n\n")

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
				escapeSingleQuote(p.FullNameEn), p.CodeName, p.AdministrativeUnitId, p.AdministrativeRegionId))
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

	dataWriter.WriteString("-- DATA for districts --\n")
	counter = 0
	isAppending = false

	for i, d := range districts {
		if !isAppending {
			dataWriter.WriteString(insertDistrictTemplate + "\n")
		}
		dataWriter.WriteString(
			fmt.Sprintf(insertDistrictWardValueTemplate, d.Code, escapeSingleQuote(d.Name), escapeSingleQuote(d.NameEn), escapeSingleQuote(d.FullName),
				escapeSingleQuote(d.FullNameEn), d.CodeName, d.ProvinceCode, d.AdministrativeUnitId))
		counter++

		// the batch insert statement batch reach limit, break and create a new batch insert statement
		if counter == batchInsertItemSize || i == len(districts)-1 {
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
				escapeSingleQuote(w.FullNameEn), w.CodeName, w.DistrictCode, w.AdministrativeUnitId))
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
	
	writePostgresGISDataToFile(gisData)
	writeMySQLGISDataToFile(gisData)
	return nil
}

// TODO @thangle: Set it for postgres first and refactor later
func writePostgresGISDataToFile(gisData []gis.ProvinceGIS) {
	fileTimeSuffix := getFileTimeSuffix()
	outputFilePath := fmt.Sprintf("./output/postgresql_generated_ImportGISData_vn_units_%s.sql", fileTimeSuffix)
	file, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Unable to write to file", err)
		panic(err)
	}
	defer file.Close()

	dataWriter := bufio.NewWriter(file)
	dataWriter.WriteString("/* === Vietnamese Provinces Database - GIS Dataset for PostgreSQL === */\n")
	dataWriter.WriteString(fmt.Sprintf("/* Created at:  %s */\n", time.Now().Format(time.RFC1123Z)))
	dataWriter.WriteString("/* Reference: https://github.com/ThangLeQuoc/vietnamese-provinces-database */\n")
	dataWriter.WriteString("/* =============================================== */\n\n")
	
	// ring order: bottom left -> going around -> bottom left again
	for _, g := range gisData {
		dataWriter.WriteString(insertPostgresGISTemplate)
		dataWriter.WriteString(fmt.Sprintf(insertPostgresGISValueTemplate, 
			g.LevelId, 
			"province",
			gis.ToCoordinateStr(g.BBox.BottomLeftLng, g.BBox.BottomLeftLat),
			gis.ToCoordinateStr(g.BBox.TopLeftLng, g.BBox.TopLeftLat),
			gis.ToCoordinateStr(g.BBox.TopRightLng, g.BBox.TopRightLat),
			gis.ToCoordinateStr(g.BBox.BottomRightLng, g.BBox.BottomRightLat),
			gis.ToCoordinateStr(g.BBox.BottomLeftLng, g.BBox.BottomLeftLat), // repeat the bottom left again to close the polygon ring,
			toPostgisMultiPolygon(g.Coordinates[0][0]),
			),
		)

		// district in province
		for _, gd := range g.Districts {
			dataWriter.WriteString(insertPostgresGISTemplate)
			dataWriter.WriteString(fmt.Sprintf(insertPostgresGISValueTemplate, 
				gd.LevelId, 
				"district",
				gis.ToCoordinateStr(gd.BBox.BottomLeftLng, gd.BBox.BottomLeftLat),
				gis.ToCoordinateStr(gd.BBox.TopLeftLng, gd.BBox.TopLeftLat),
				gis.ToCoordinateStr(gd.BBox.TopRightLng, gd.BBox.TopRightLat),
				gis.ToCoordinateStr(gd.BBox.BottomRightLng, gd.BBox.BottomRightLat),
				gis.ToCoordinateStr(gd.BBox.BottomLeftLng, gd.BBox.BottomLeftLat), // repeat the bottom left again to close the polygon ring,
				toPostgisMultiPolygon(gd.Coordinates[0][0]),
				),
			)
		}
	}

	dataWriter.Flush()
}

func writeMySQLGISDataToFile(gisData []gis.ProvinceGIS) {
	fileTimeSuffix := getFileTimeSuffix()
	outputFilePath := fmt.Sprintf("./output/mysql_generated_ImportGISData_vn_units_%s.sql", fileTimeSuffix)
	file, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Unable to write to file", err)
		panic(err)
	}
	defer file.Close()

	dataWriter := bufio.NewWriter(file)
	dataWriter.WriteString("/* === Vietnamese Provinces Database - GIS Dataset for MySQL === */\n")
	dataWriter.WriteString(fmt.Sprintf("/* Created at:  %s */\n", time.Now().Format(time.RFC1123Z)))
	dataWriter.WriteString("/* Reference: https://github.com/ThangLeQuoc/vietnamese-provinces-database */\n")
	dataWriter.WriteString("/* =============================================== */\n\n")
	
	// ring order: bottom left -> going around -> bottom left again
	for _, g := range gisData {
		dataWriter.WriteString(insertMySQLGISTemplate)
		dataWriter.WriteString(fmt.Sprintf(insertMySQLGISValueTemplate, 
			g.LevelId, 
			"province",
			gis.ToCoordinateStrLatLng(g.BBox.BottomLeftLng, g.BBox.BottomLeftLat),
			gis.ToCoordinateStrLatLng(g.BBox.TopLeftLng, g.BBox.TopLeftLat),
			gis.ToCoordinateStrLatLng(g.BBox.TopRightLng, g.BBox.TopRightLat),
			gis.ToCoordinateStrLatLng(g.BBox.BottomRightLng, g.BBox.BottomRightLat),
			gis.ToCoordinateStrLatLng(g.BBox.BottomLeftLng, g.BBox.BottomLeftLat), // repeat the bottom left again to close the polygon ring,
			toMySQLMultiPolygon(g.Coordinates[0][0]),
			),
		)

		// district in province
		for _, gd := range g.Districts {
			dataWriter.WriteString(insertMySQLGISTemplate)
			dataWriter.WriteString(fmt.Sprintf(insertMySQLGISValueTemplate, 
				gd.LevelId, 
				"district",
				gis.ToCoordinateStrLatLng(gd.BBox.BottomLeftLng, gd.BBox.BottomLeftLat),
				gis.ToCoordinateStrLatLng(gd.BBox.TopLeftLng, gd.BBox.TopLeftLat),
				gis.ToCoordinateStrLatLng(gd.BBox.TopRightLng, gd.BBox.TopRightLat),
				gis.ToCoordinateStrLatLng(gd.BBox.BottomRightLng, gd.BBox.BottomRightLat),
				gis.ToCoordinateStrLatLng(gd.BBox.BottomLeftLng, gd.BBox.BottomLeftLat), // repeat the bottom left again to close the polygon ring,
				toMySQLMultiPolygon(gd.Coordinates[0][0]),
				),
			)
		}
	}

	dataWriter.Flush()
}

func toPostgisMultiPolygon(ringCoordinates gis.GisLinearRingCoordinate) string {
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

func toMySQLMultiPolygon(ringCoordinates gis.GisLinearRingCoordinate) string {
	var sb strings.Builder
	for _, p := range ringCoordinates.GisPoints {
		sb.WriteString(gis.ToCoordinateStrLatLng(p.Longitude, p.Latitude))
		sb.WriteString(",")
	}
	// repeat the first point to close the ring
	firstPoint := ringCoordinates.GisPoints[0]
	sb.WriteString(gis.ToCoordinateStrLatLng(firstPoint.Longitude, firstPoint.Latitude))
	return sb.String()
}

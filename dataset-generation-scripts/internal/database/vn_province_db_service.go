package database

import (
	"context"
	"fmt"
	"os"
)

const cleanUpScript = "./resources/fresh_cleanup.sql"
const pathToTableInitFile = "./resources/db_table_init.sql"
const pathToRegionAdministrativeInitFile = "./resources/db_region_administrative_unit.sql"

// GIS data structure scripts
const pathToSapNhapTables = "./resources/gis/sapnhap_bando_tables.sql"

/*
Bootstrap the Temporary Dataset Structure
@thangle: This function is the remnant of the past. From the previous version of the dataset-generation-script, data are get from CSV file and dump to database.
We may skip this this in the future
*/
func BootstrapTemporaryDatasetStructure() {
	err := ExecuteSQLScript(cleanUpScript)
	if err != nil {
		panic(err)
	}
	fmt.Println("DB Fresh cleanup script executed")
	
	err = ExecuteSQLScript(pathToTableInitFile)
	if err != nil {
		panic(err)
	}
	fmt.Println("Temporary Provinces tables created")

	err = ExecuteSQLScript(pathToRegionAdministrativeInitFile)
	if err != nil {
		panic(err)
	}
	fmt.Println("Data for regions & administrative unit persisted")
}

/*
Bootstrap GIS related data structure
*/
func BootstrapGISDataStructure() {
	err := ExecuteSQLScript(pathToSapNhapTables)
	if err != nil {
		panic(err)
	}
	fmt.Println("GIS related data structure created")
}

// Useful method to execute SQL script located in this project
func ExecuteSQLScript(pathToSQL string) error {
	bytesVal, err := os.ReadFile(pathToSQL)
	if err != nil {
		panic(err)
	}
	query := string(bytesVal)
	db := GetPostgresDBConnection()
	ctx := context.Background()
	_, err = db.ExecContext(ctx, query)
	ctx.Done()
	if err != nil {
		return err
	}
	return nil
}

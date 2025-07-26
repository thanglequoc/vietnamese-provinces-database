package common

import (
	"fmt"
)

const pathToTableInitFile = "./resources/db_table_init.sql"
const pathToRegionAdministrativeInitFile = "./resources/db_region_administrative_unit.sql"

/*
Bootstrap the Temporary Dataset Structure
@thangle: This function is the remnant of the past. From the previous version of the dataset-generation-script, data are get from CSV file and dump to database.
We may skip this this in the future
*/
func BootstrapTemporaryDatasetStructure() {
	err := ExecuteSQLScript(pathToTableInitFile)
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

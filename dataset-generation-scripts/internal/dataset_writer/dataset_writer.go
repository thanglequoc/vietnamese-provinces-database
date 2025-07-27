package dataset_writer

import (
	"fmt"
	"log"
	"os"

	db "github.com/thanglequoc-vn-provinces/v2/internal/database"
	dataset_file_writer "github.com/thanglequoc-vn-provinces/v2/internal/dataset_writer/dataset_file_writer"
	vn_provinces_tmp_repo "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/repository"
)

/*
Generate the Vietnamese Provinces Dataset SQL files
*/
func ReadAndGenerateSQLDatasets() {

	vn_provinces_tmp_repo := vn_provinces_tmp_repo.NewVnProvincesTmpRepository(db.GetPostgresDBConnection())


	// Clean up the output folder
	os.RemoveAll("./output")
	os.MkdirAll("./output", 0746)

	regions := vn_provinces_tmp_repo.GetAllAdministrativeRegions()
	administrativeUnits := vn_provinces_tmp_repo.GetAllAdministrativeUnits()
	provinces := vn_provinces_tmp_repo.GetAllProvinces()
	wards := vn_provinces_tmp_repo.GetAllWards()

	// Postgresql & MySQL
	postgresMySQLDatasetFileWriter := dataset_file_writer.PostgresMySQLDatasetFileWriter{
		OutputFilePath: "./output/postgresql_mysql_generated_ImportData_vn_units_%s.sql",
	}
	err := postgresMySQLDatasetFileWriter.WriteToFile(regions, administrativeUnits, provinces, wards)
	if err != nil {
		log.Fatal("Unable to generate Postgresql-MySQL Dataset", err)
	} else {
		fmt.Println("✅ Postgresql-MySQL Dataset successfully generated")
	}

	// Mssql
	mssqlDatasetFileWriter := dataset_file_writer.MssqlDatasetFileWriter{
		OutputFilePath: "./output/mssql_generated_ImportData_vn_units_%s.sql",
	}
	err = mssqlDatasetFileWriter.WriteToFile(regions, administrativeUnits, provinces, wards)
	if err != nil {
		log.Fatal("Unable to generate Mssql Dataset", err)
	} else {
		fmt.Println("✅ Mssql Dataset successfully generated")
	}

	// Oracle
	oracleDatasetFileWriter := dataset_file_writer.OracleDatasetFileWriter{
		OutputFilePath: "./output/oracle_generated_ImportData_vn_units_%s.sql",
	}
	err = oracleDatasetFileWriter.WriteToFile(regions, administrativeUnits, provinces, wards)
	if err != nil {
		log.Fatal("Unable to generate Oracle Dataset", err)
	} else {
		fmt.Println("✅ Oracle Dataset successfully generated")
	}

	// JSON
	jsonDatasetFileWriter := dataset_file_writer.JSONDatasetFileWriter{
		OutputFolderPath: "./output/json",
	}
	err = jsonDatasetFileWriter.WriteToFile(regions, administrativeUnits, provinces, wards)
	if err != nil {
		log.Fatal("Unable to generate JSON Dataset", err)
	} else {
		fmt.Println("✅ JSON Dataset successfully generated")
	}

	// MongoDB
	mongoDBDatasetFileWriter := dataset_file_writer.MongoDBDatasetFileWriter{
		OutputFolderPath: "./output/mongodb",
	}
	err = mongoDBDatasetFileWriter.WriteToFile(regions, administrativeUnits, provinces, wards)
	if err != nil {
		log.Fatal("Unable to generate MongoDB Dataset", err)
	} else {
		fmt.Println("✅ MongoDB Dataset successfully generated")
	}

	// Redis
	redisDatasetFileWriter := dataset_file_writer.RedisDatasetFileWriter{
		OutputFolderPath: "./output/redis",
	}
	err = redisDatasetFileWriter.WriteToFile(regions, administrativeUnits, provinces, wards)
	if err != nil {
		log.Fatal("Unable to generate Redis Dataset", err)
	} else {
		fmt.Println("✅ Redis Dataset successfully generated")
	}
}

package dataset_writer

import (
	"fmt"
	"log"
	"os"

	db "github.com/thanglequoc-vn-provinces/v2/internal/database"
	datasetfilewriter "github.com/thanglequoc-vn-provinces/v2/internal/dataset_writer/dataset_file_writer"
	vnprovincestmprepo "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/repository"
	sapnhapbandorepo "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/repository"
)

/*
Generate the Vietnamese Provinces Dataset SQL files
*/
func ReadAndGenerateSQLDatasets() {

	vn_provinces_tmp_repo := vnprovincestmprepo.NewVnProvincesTmpRepository(db.GetPostgresDBConnection())

	// Clean up the output folder
	os.RemoveAll("./output")
	os.MkdirAll("./output", 0746)

	regions := vn_provinces_tmp_repo.GetAllAdministrativeRegions()
	administrativeUnits := vn_provinces_tmp_repo.GetAllAdministrativeUnits()
	provinces := vn_provinces_tmp_repo.GetAllProvinces()
	wards := vn_provinces_tmp_repo.GetAllWards()

	// Postgresql & MySQL
	postgresMySQLDatasetFileWriter := datasetfilewriter.PostgresMySQLDatasetFileWriter{
		OutputFilePath: "./output/postgresql_mysql_generated_ImportData_vn_units_%s.sql",
	}
	err := postgresMySQLDatasetFileWriter.WriteToFile(regions, administrativeUnits, provinces, wards)
	if err != nil {
		log.Fatal("Unable to generate Postgresql-MySQL Dataset", err)
	} else {
		fmt.Println("✅ Postgresql-MySQL Dataset successfully generated")
	}

	// Mssql
	mssqlDatasetFileWriter := datasetfilewriter.MssqlDatasetFileWriter{
		OutputFilePath: "./output/mssql_generated_ImportData_vn_units_%s.sql",
	}
	err = mssqlDatasetFileWriter.WriteToFile(regions, administrativeUnits, provinces, wards)
	if err != nil {
		log.Fatal("Unable to generate Mssql Dataset", err)
	} else {
		fmt.Println("✅ Mssql Dataset successfully generated")
	}

	// Oracle
	oracleDatasetFileWriter := datasetfilewriter.OracleDatasetFileWriter{
		OutputFilePath: "./output/oracle_generated_ImportData_vn_units_%s.sql",
	}
	err = oracleDatasetFileWriter.WriteToFile(regions, administrativeUnits, provinces, wards)
	if err != nil {
		log.Fatal("Unable to generate Oracle Dataset", err)
	} else {
		fmt.Println("✅ Oracle Dataset successfully generated")
	}

	// JSON
	jsonDatasetFileWriter := datasetfilewriter.JSONDatasetFileWriter{
		OutputFolderPath: "./output/json",
	}
	err = jsonDatasetFileWriter.WriteToFile(regions, administrativeUnits, provinces, wards)
	if err != nil {
		log.Fatal("Unable to generate JSON Dataset", err)
	} else {
		fmt.Println("✅ JSON Dataset successfully generated")
	}

	// MongoDB
	mongoDBDatasetFileWriter := datasetfilewriter.MongoDBDatasetFileWriter{
		OutputFolderPath: "./output/mongodb",
	}
	err = mongoDBDatasetFileWriter.WriteToFile(regions, administrativeUnits, provinces, wards)
	if err != nil {
		log.Fatal("Unable to generate MongoDB Dataset", err)
	} else {
		fmt.Println("✅ MongoDB Dataset successfully generated")
	}

	// Redis
	redisDatasetFileWriter := datasetfilewriter.RedisDatasetFileWriter{
		OutputFolderPath: "./output/redis",
	}
	err = redisDatasetFileWriter.WriteToFile(regions, administrativeUnits, provinces, wards)
	if err != nil {
		log.Fatal("Unable to generate Redis Dataset", err)
	} else {
		fmt.Println("✅ Redis Dataset successfully generated")
	}
}

/*
Generate the GIS SQL files
TODO @thangle:Implement this
*/
func GenerateGISSQLDatasets() {
	sapNhapBanDoRepo := sapnhapbandorepo.NewSapNhapRepository(db.GetPostgresDBConnection())
	sapNhapProvinces, err := sapNhapBanDoRepo.GetAllSapNhapSiteProvinces()
	if err != nil {
		log.Fatal("Unable to get SapNhapSiteProvinces", err)
		return
	}
	sapNhapWards, err := sapNhapBanDoRepo.GetAllSapNhapSiteWards()
	if err != nil {
		log.Fatal("Unable to get SapNhapSiteWards", err)
		return
	}
	

	// Postgresql & MySQL
	postgresMySQLDatasetFileWriter := datasetfilewriter.PostgresMySQLDatasetFileWriter{
		OutputFilePath: "./output/postgresql_mysql_generated_ImportData_vn_units_%s.sql",
	}

	postgresMySQLDatasetFileWriter.WriteGISDataToFile(sapNhapProvinces, sapNhapWards)
}

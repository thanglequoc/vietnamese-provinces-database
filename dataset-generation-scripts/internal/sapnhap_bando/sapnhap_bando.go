package sapnhapbando

import (
	"fmt"
	"log"
	"os"
	"strconv"

	db "github.com/thanglequoc-vn-provinces/v2/internal/database"

	sapNhapR "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/repository"
	sapNhapService "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/service"
	vnRepo "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/repository"
)

const (
	PROVINCE_GIS_EXISTING_DUMP_PATH="./resources/gis/exported/sapnhap_provinces_gis_202509011857_lfs.sql"
	WARD_GIS_EXISTING_DUMP_PATH="./resources/gis/exported/sapnhap_wards_gis_202509011858_lfs.sql"
)

func DumpDataFromSapNhapBando() {
	// Initialize the SapNhapRepository
	sapNhapRepo := sapNhapR.NewSapNhapRepository(db.GetPostgresDBConnection())
	sapNhapGISRepo := sapNhapR.NewSapNhapGISRepository(db.GetPostgresDBConnection())
	vnRepo := vnRepo.NewVnProvincesTmpRepository(db.GetPostgresDBConnection())

	sapNhapService := sapNhapService.NewSapNhapService(sapNhapRepo, sapNhapGISRepo, vnRepo)

	if err := sapNhapService.BootstrapSapNhapSiteProvinces(); err != nil {
		log.Fatalf("Failed to dump SapNhapSiteProvinces: %v", err)
		panic(err)
	}

	if err := sapNhapService.BootstrapSapNhapSiteWards(); err != nil {
		log.Fatalf("Failed to dump SapNhapSiteWards: %v", err)
		panic(err)
	}

	shouldGetGISFromDumpSQLPatch, err := strconv.ParseBool(os.Getenv("DUMP_GIS_RESOURCE_FROM_EXISTING_DB_PATCH"))
	if err != nil {
		fmt.Println("Error parsing DUMP_GIS_RESOURCE_FROM_EXISTING_DB_PATCH, default to false")
		shouldGetGISFromDumpSQLPatch = false
	}

	if shouldGetGISFromDumpSQLPatch {
		log.Println("ℹ️ DUMP_GIS_RESOURCE_FROM_EXISTING_DB_PATCH is set to true, will load GIS data from existing SQL dump file")
		// Just go ahead and execute the existing SQL dump file
		if err := db.ExecuteSQLScript(PROVINCE_GIS_EXISTING_DUMP_PATH); err != nil {
			log.Fatalf("Failed to execute province GIS dump script: %v", err)
			panic(err)
		}
		if err := db.ExecuteSQLScript(WARD_GIS_EXISTING_DUMP_PATH); err != nil {
			log.Fatalf("Failed to execute ward GIS dump script: %v", err)
			panic(err)
		}
	} else {
		log.Println("ℹ️ DUMP_GIS_RESOURCE_FROM_EXISTING_DB_PATCH is set to false, will fetch GIS data from GIS server")
		if err := sapNhapService.BootstrapGISDataFromGISServer(); err != nil {
			log.Fatalf("Failed to bootstrap GIS Data: %v", err)
			panic(err)
		}
	}

	log.Println("✅ Data dump from SapNhap Bando completed successfully")
}

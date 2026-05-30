package sapnhapbando

import (
	"context"
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
	PROVINCE_GIS_EXISTING_DUMP_PATH = "./resources/gis/exported/sapnhap_provinces_gis_202509011857_lfs.sql"
	WARD_GIS_EXISTING_DUMP_PATH     = "./resources/gis/exported/sapnhap_wards_gis_202509011858_lfs.sql"
)

func DumpDataFromSapNhapBando() {
	// Initialize the SapNhapRepository
	postgresDB := db.GetPostgresDBConnection()
	sapNhapRepo := sapNhapR.NewSapNhapRepository(postgresDB)
	sapNhapGISRepo := sapNhapR.NewSapNhapGISRepository(postgresDB)
	sapNhapGeoJSONObjectRepository := sapNhapR.NewSapNhapGeoJSONObjectRepository(postgresDB)
	vnRepo := vnRepo.NewVnProvincesTmpRepository(postgresDB)

	sapNhapService := sapNhapService.NewSapNhapService(sapNhapRepo, sapNhapGISRepo, vnRepo, sapNhapGeoJSONObjectRepository, postgresDB)

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

func FetchGISDataFromSapNhapBando() {
	// Initialize repository
	postgresDB := db.GetPostgresDBConnection()
	sapNhapGeoJSONObjectRepository := sapNhapR.NewSapNhapGeoJSONObjectRepository(postgresDB)
	
	// Initialize service with required dependencies
	sapNhapRepo := sapNhapR.NewSapNhapRepository(postgresDB)
	sapNhapGISRepo := sapNhapR.NewSapNhapGISRepository(postgresDB)
	vnRepo := vnRepo.NewVnProvincesTmpRepository(postgresDB)
	
	sapNhapService := sapNhapService.NewSapNhapService(sapNhapRepo, sapNhapGISRepo, vnRepo, sapNhapGeoJSONObjectRepository, postgresDB)
	
	// Fetch GIS data from Bando server and update database
	log.Println("ℹ️ Starting to fetch GIS data from Bando server...")
	if err := sapNhapService.FetchGISDataFromSapNhapBando(sapNhapGeoJSONObjectRepository); err != nil {
		log.Fatalf("Failed to fetch GIS data from Bando: %v", err)
		panic(err)
	}
	
	log.Println("✅ Fetching GIS data from Bando completed successfully")
}

// BackfillProvinceAndWardCodesInSapNhapGeojsonObjects backfills vn_ds_province_code and vn_ds_ward_code
// fields in sapnhap_geojson_objects table by matching names against provinces_tmp and wards_tmp tables.
// This is a standalone function that can be called independently of GIS data fetching.
func BackfillProvinceAndWardCodesInSapNhapGeojsonObjects() {
	// Initialize service with required dependencies
	postgresDB := db.GetPostgresDBConnection()
	sapNhapRepo := sapNhapR.NewSapNhapRepository(postgresDB)
	sapNhapGISRepo := sapNhapR.NewSapNhapGISRepository(postgresDB)
	vnRepo := vnRepo.NewVnProvincesTmpRepository(postgresDB)
	sapNhapGeoJSONObjectRepository := sapNhapR.NewSapNhapGeoJSONObjectRepository(postgresDB)

	sapNhapService := sapNhapService.NewSapNhapService(sapNhapRepo, sapNhapGISRepo, vnRepo, sapNhapGeoJSONObjectRepository, postgresDB)

	ctx := context.Background()
	if err := sapNhapService.FillMetaDataForGeoJSONObjects(ctx); err != nil {
		log.Fatalf("Failed to fill metadata for geo objects: %v", err)
	}

	// Backfill province and ward codes
	log.Println("ℹ️ Starting to backfill province and ward codes...")
	if err := sapNhapService.BackfillProvinceAndWardCodesInSapNhapGeojsonObjects(); err != nil {
		log.Fatalf("Failed to backfill province and ward codes: %v", err)
		panic(err)
	}
	
	log.Println("✅ Backfill of province and ward codes completed successfully")
}

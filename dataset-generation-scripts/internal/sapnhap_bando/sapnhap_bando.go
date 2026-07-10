package sapnhapbando

import (
	"context"
	"log"

	db "github.com/thanglequoc-vn-provinces/v2/internal/database"

	sapNhapR "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/repository"
	sapNhapService "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/service"
	vnRepo "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/repository"
)

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

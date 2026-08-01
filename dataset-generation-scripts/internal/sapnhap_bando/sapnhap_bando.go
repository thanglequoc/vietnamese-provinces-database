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
	vnRepo := vnRepo.NewVnProvincesTmpRepository(postgresDB)
	
	sapNhapService := sapNhapService.NewSapNhapService(vnRepo, sapNhapGeoJSONObjectRepository, postgresDB)
	
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
	vnRepo := vnRepo.NewVnProvincesTmpRepository(postgresDB)
	sapNhapGeoJSONObjectRepository := sapNhapR.NewSapNhapGeoJSONObjectRepository(postgresDB)

	sapNhapService := sapNhapService.NewSapNhapService(vnRepo, sapNhapGeoJSONObjectRepository, postgresDB)

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

/*
PatchIslandProvincesGeometry merges island ward geometries (Hoàng Sa, Trường Sa)
into their parent province geometries (Da Nang, Khanh Hoa). This fixes an upstream
GIS data defect where the province-level API response excludes island territories
that are present at the ward level.

This should be called after FetchGISDataFromSapNhapBando() and before
GenerateGISSQLDatasets() in the generation flow.
*/
func PatchIslandProvincesGeometry() {
	postgresDB := db.GetPostgresDBConnection()
	sapNhapGeoJSONObjectRepository := sapNhapR.NewSapNhapGeoJSONObjectRepository(postgresDB)
	vnRepo := vnRepo.NewVnProvincesTmpRepository(postgresDB)
	sapNhapService := sapNhapService.NewSapNhapService(vnRepo, sapNhapGeoJSONObjectRepository, postgresDB)

	log.Println("ℹ️ Patching island province geometries (Hoàng Sa → Da Nang, Trường Sa → Khanh Hoa)...")
	if err := sapNhapService.PatchIslandProvincesGeometry(context.Background()); err != nil {
		log.Fatalf("Failed to patch island province geometries: %v", err)
		panic(err)
	}
	log.Println("✅ Island province geometry patching completed successfully")
}

// ValidateAndFixGeometries checks all ward geometries in sapnhap_geojson_objects
// for self-intersections and fixes them using ST_MakeValid + ST_CollectionExtract.
// An audit log of all fixed wards is written to output/gis_geometry_fix_log_<timestamp>.txt.
//
// This should be called after PatchIslandProvincesGeometry() and before
// GenerateGISSQLDatasets() in the generation flow.
func ValidateAndFixGeometries() {
	postgresDB := db.GetPostgresDBConnection()
	sapNhapGeoJSONObjectRepository := sapNhapR.NewSapNhapGeoJSONObjectRepository(postgresDB)
	vnTmpRepo := vnRepo.NewVnProvincesTmpRepository(postgresDB)
	sapNhapSvc := sapNhapService.NewSapNhapService(vnTmpRepo, sapNhapGeoJSONObjectRepository, postgresDB)

	log.Println("ℹ️ Validating and fixing GIS geometries (self-intersection repair)...")
	if err := sapNhapSvc.ValidateAndFixGeometries(context.Background()); err != nil {
		log.Fatalf("Failed to fix GIS geometries: %v", err)
		panic(err)
	}
	log.Println("✅ GIS geometry validation and fix completed successfully")
}

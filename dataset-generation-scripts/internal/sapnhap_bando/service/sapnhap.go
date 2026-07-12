package service

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/dto"
	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/fetcher"
	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/repository"
	vnRepo "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/repository"
	"github.com/uptrace/bun"
)

type SapNhapService struct {
	sapNhapGeoJSONRepo *repository.SapNhapGeoJSONObjectRepository
	vnProvinceTmpRepo  *vnRepo.VnProvincesTmpRepository
	db                 *bun.DB
}

func NewSapNhapService(vnRepo *vnRepo.VnProvincesTmpRepository,
	sapNhapGeoJSONRepo *repository.SapNhapGeoJSONObjectRepository, db *bun.DB) *SapNhapService {
	return &SapNhapService{
		sapNhapGeoJSONRepo: sapNhapGeoJSONRepo,
		vnProvinceTmpRepo:  vnRepo,
		db:                 db,
	}
}

// ProcessingError represents a record that failed to process
type ProcessingError struct {
	Ma    string
	Ten   string
	MaLK  string
	Error string
}

// ProcessingResult represents the result of processing a single geo object
type ProcessingResult struct {
	Success bool
	Error   ProcessingError
}

// BackfillProvinceAndWardCodesInSapNhapGeojsonObjects backfills vn_ds_province_code and vn_ds_ward_code
// fields in sapnhap_geojson_objects table by matching names against provinces_tmp and wards_tmp tables.
// This is a standalone function that can be called independently of GIS data fetching.
func (s *SapNhapService) BackfillProvinceAndWardCodesInSapNhapGeojsonObjects() error {
	// Create a new repository instance for geo objects using the DB from service
	geoJSONRepo := repository.NewSapNhapGeoJSONObjectRepository(s.db)

	// Create a backfill service instance
	backfillService := NewSapNhapBackfillService(s.vnProvinceTmpRepo, geoJSONRepo)

	ctx := context.Background()

	// Execute backfill using the dedicated backfill service
	err := backfillService.ExecuteBackfill(ctx)
	if err != nil {
		return fmt.Errorf("backfill failed: %w", err)
	}

	log.Println("Backfill of province and ward codes completed successfully")
	return nil
}

// FetchGISDataFromSapNhapBando fetches WKT geometry data from the Bando GIS server
// for all records in the sapnhap_geojson_objects table and updates them with the retrieved data.
// Uses parallel processing with a worker pool for improved performance.
func (s *SapNhapService) FetchGISDataFromSapNhapBando(geoJSONRepo *repository.SapNhapGeoJSONObjectRepository) error {
	ctx := context.Background()

	// Get all geo objects from the database
	geoObjects, err := geoJSONRepo.GetAllSapNhapGeoJSONObjects(ctx)
	if err != nil {
		return fmt.Errorf("failed to get sapnhap geojson objects: %w", err)
	}

	log.Printf("Found %d geo objects to process", len(geoObjects))

	// Number of concurrent workers
	numWorkers := 10
	log.Printf("Processing with %d concurrent workers", numWorkers)

	// Create channels for work distribution and result collection
	workChan := make(chan *model.SapNhapSiteGeoUnit, len(geoObjects))
	resultChan := make(chan ProcessingResult, len(geoObjects))

	// WaitGroup to wait for all workers to finish
	var wg sync.WaitGroup

	// Mutex for thread-safe error collection
	var errorMutex sync.Mutex

	// Start worker pool
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go s.worker(ctx, &wg, workChan, resultChan, geoJSONRepo)
	}

	// Start a goroutine to close resultChan when all workers finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Send work to workers
	for _, geoObject := range geoObjects {
		workChan <- geoObject
	}
	close(workChan)

	// Collect results
	successCount := 0
	processingErrors := make([]ProcessingError, 0)

	for result := range resultChan {
		if result.Success {
			successCount++
		} else {
			errorMutex.Lock()
			processingErrors = append(processingErrors, result.Error)
			errorMutex.Unlock()
			log.Printf("Error processing geo object [ma: %s, ten: %s, malk: %s]: %v",
				result.Error.Ma, result.Error.Ten, result.Error.MaLK, result.Error.Error)
		}
	}

	log.Printf("Processing complete. Success: %d, Errors: %d", successCount, len(processingErrors))

	// Print summary of failed records for manual inspection
	if len(processingErrors) > 0 {
		log.Println("\n==========================================")
		log.Println("FAILED RECORDS SUMMARY FOR MANUAL INSPECTION:")
		log.Println("==========================================")
		for i, pe := range processingErrors {
			log.Printf("%d. MA: %s, TEN: %s, MALK: %s", i+1, pe.Ma, pe.Ten, pe.MaLK)
			log.Printf("   Error: %s", pe.Error)
			log.Println("------------------------------------------")
		}
		log.Println("==========================================")
		log.Printf("Total failed records: %d out of %d", len(processingErrors), len(geoObjects))
		log.Println("==========================================")

		return fmt.Errorf("completed with %d errors out of %d total records. See above for details.", len(processingErrors), len(geoObjects))
	}

	return nil
}

// worker processes geo objects from the work channel
func (s *SapNhapService) worker(ctx context.Context, wg *sync.WaitGroup, workChan <-chan *model.SapNhapSiteGeoUnit, resultChan chan<- ProcessingResult, geoJSONRepo *repository.SapNhapGeoJSONObjectRepository) {
	defer wg.Done()

	for geoObject := range workChan {
		err := s.processGeoJSONObject(ctx, geoObject, geoJSONRepo)
		if err != nil {
			resultChan <- ProcessingResult{
				Success: false,
				Error: ProcessingError{
					Ma:    geoObject.Ma,
					Ten:   geoObject.Ten,
					MaLK:  geoObject.MaLK,
					Error: err.Error(),
				},
			}
		} else {
			resultChan <- ProcessingResult{Success: true}
		}
	}
}

/*
loadAnGiangProvinceFromLocalFile loads An Giang province geometry from a local GeoJSON file
This is a manual patch for corrupted upstream data for An Giang province (MA: ti32)
Returns WKT format bbox and geometry
*/
func loadAnGiangProvinceFromLocalFile() (wktGeometry string, err error) {
	// Path to the manual patch file
	const anGiangPatchPath = "./resources/gis/geojson_11Mar2026/32_tinh_an_giang/province.geojson"

	// Load GeoJSON file
	geojson, err := dto.LoadGeoJSONFile(anGiangPatchPath)
	if err != nil {
		return "", fmt.Errorf("failed to load An Giang patch file from %s: %w", anGiangPatchPath, err)
	}

	if len(geojson.Features) == 0 {
		return "", fmt.Errorf("no features found in An Giang patch file")
	}

	feature := geojson.Features[0]

	return feature.Geometry.ToWKTMultiPolygon(), nil
}

/*
processGeoJSONObject fetches GIS data for a single geo object and updates the database
*/
func (s *SapNhapService) processGeoJSONObject(ctx context.Context, geoObject *model.SapNhapSiteGeoUnit, geoJSONRepo *repository.SapNhapGeoJSONObjectRepository) error {
	// Check if the geo object has a MaLK value
	if geoObject.MaLK == "" {
		return fmt.Errorf("malk field is empty, cannot fetch GIS data")
	}

	log.Printf("Fetching GIS data for [ma: %s, ten: %s, malk: %s]", geoObject.Ma, geoObject.Ten, geoObject.MaLK)

	// SPECIAL CASE: An Giang province (91) has corrupted upstream data
	// Use manual patch from local file instead
	if geoObject.Ma == "91" {
		log.Printf("⚠️  DETECTED CORRUPTED DATA: An Giang province (MA: 91, MALK: %s)", geoObject.MaLK)
		log.Printf("🔧 APPLYING MANUAL PATCH: Loading An Giang GIS data from local file: ./resources/gis/geojson_11Mar2026/32_tinh_an_giang/province.geojson")

		wktGeometry, err := loadAnGiangProvinceFromLocalFile()
		if err != nil {
			return fmt.Errorf("failed to load An Giang province from local patch file: %w", err)
		}

		// Update the database with patched geometry; bbox is derived in PostGIS.
		err = geoJSONRepo.UpdateSapNhapGeoJSONObjectGeomWKT(ctx, geoObject.Ma, wktGeometry)
		if err != nil {
			return fmt.Errorf("failed to update geo object [ma: %s] with patched data: %w", geoObject.Ma, err)
		}

		log.Printf("✅ Successfully updated geo object [ma: %s, ten: %s] using MANUAL PATCH", geoObject.Ma, geoObject.Ten)
		return nil
	}

	// Normal processing for other provinces/wards
	// Fetch GIS data from the server
	gisResponse, err := fetcher.GetGISLocationCoordinates(geoObject.MaLK)
	if err != nil {
		return fmt.Errorf("failed to get GIS location coordinates for malk %s: %w", geoObject.MaLK, err)
	}

	// Validate response
	if len(gisResponse.Features) == 0 {
		return fmt.Errorf("no features found in GIS response for malk %s", geoObject.MaLK)
	}

	feature := gisResponse.Features[0]

	wktGeometry := feature.Geometry.ToWKTCoordinate()

	// Update the database; bbox is derived in PostGIS from the persisted geometry.
	err = geoJSONRepo.UpdateSapNhapGeoJSONObjectGeomWKT(ctx, geoObject.Ma, wktGeometry)
	if err != nil {
		return fmt.Errorf("failed to update geo object [ma: %s]: %w", geoObject.Ma, err)
	}

	log.Printf("Successfully updated geo object [ma: %s, ten: %s]", geoObject.Ma, geoObject.Ten)
	return nil
}

func (s *SapNhapService) FillMetaDataForGeoJSONObjects(ctx context.Context) error {
	// Get all geo objects from the database
	geoObjects, err := s.sapNhapGeoJSONRepo.GetAllSapNhapGeoJSONObjects(ctx)
	if err != nil {
		return fmt.Errorf("failed to get sapnhap geojson objects: %w", err)
	}

	for _, geoObject := range geoObjects {
		log.Printf("Processing geo object [ma: %s, ten: %s, malk: %s]", geoObject.Ma, geoObject.Ten, geoObject.MaLK)
		geoMetadata, err := fetcher.GetMetadataOfSapNhapGeoObject(ctx, geoObject.MaLK)
		if err != nil {
			log.Printf("Error fetching metadata for geo object [ma: %s, ten: %s, malk: %s]: %v", geoObject.Ma, geoObject.Ten, geoObject.MaLK, err)
			continue
		}
		err = s.sapNhapGeoJSONRepo.UpdateSapNhapGeoJSONObjectMetadata(ctx, geoObject.MaLK, geoMetadata)
		if err != nil {
			log.Printf("Error updating metadata for geo object [ma: %s, ten: %s, malk: %s]: %v", geoObject.Ma, geoObject.Ten, geoObject.MaLK, err)
		} else {
			log.Printf("Successfully updated metadata for geo object [ma: %s, ten: %s, malk: %s]", geoObject.Ma, geoObject.Ten, geoObject.MaLK)
		}
	}
	return nil
}

/*
PatchIslandProvincesGeometry merges island ward geometries into their parent
province geometries. This fixes an upstream GIS data defect where the province-level
API response from sapnhap.bando.com.vn excludes island territories (Hoàng Sa and
Trường Sa) that are present at the ward level.

Affected provinces:
  - Da Nang (ma=48) ← Hoàng Sa (ma=20333)
  - Khanh Hoa (ma=56) ← Trường Sa (ma=22736)

After the patch, each province geometry spatially contains all of its administrative
subdivisions, including the Paracel and Spratly Islands.
*/
func (s *SapNhapService) PatchIslandProvincesGeometry(ctx context.Context) error {
	// Merge Hoàng Sa into Da Nang
	err := s.mergeWardGeometryIntoProvince(ctx, "48", "20333")
	if err != nil {
		return fmt.Errorf("failed to patch Da Nang with Hoàng Sa geometry: %w", err)
	}
	log.Println("✅ Patched Da Nang (48) with Hoàng Sa (20333) island geometry")

	// Merge Trường Sa into Khanh Hoa
	err = s.mergeWardGeometryIntoProvince(ctx, "56", "22736")
	if err != nil {
		return fmt.Errorf("failed to patch Khanh Hoa with Trường Sa geometry: %w", err)
	}
	log.Println("✅ Patched Khanh Hoa (56) with Trường Sa (22736) island geometry")

	return nil
}

/*
mergeWardGeometryIntoProvince merges the geometry of a ward (identified by wardMa)
into the geometry of a province (identified by provinceMa) using PostGIS ST_Union.
Both geom_wkt and bbox_wkt are updated for the province record. The bbox is
recalculated via ST_Envelope to encompass the merged geometry.
*/
func (s *SapNhapService) mergeWardGeometryIntoProvince(ctx context.Context, provinceMa, wardMa string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE sapnhap_geojson_objects
		SET geom_wkt = ST_AsText(ST_Union(geom, (SELECT geom FROM sapnhap_geojson_objects WHERE ma = ?))),
		    bbox_wkt = ST_AsText(ST_Envelope(ST_Union(geom, (SELECT geom FROM sapnhap_geojson_objects WHERE ma = ?))))
		WHERE ma = ?`,
		wardMa, wardMa, provinceMa)
	return err
}

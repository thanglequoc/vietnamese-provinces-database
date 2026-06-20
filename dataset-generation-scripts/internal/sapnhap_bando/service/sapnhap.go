package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/dto"
	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/fetcher"
	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/repository"
	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/util"
	vnRepo "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/repository"
	"github.com/uptrace/bun"
)

const BANDO_GIS_PROVINCES_FILE_PATH = "./resources/gis/bando_gisserver/provinces.json"
const BANDO_GIS_WARDS_FILE_PATH = "./resources/gis/bando_gisserver/wards.json"
const METADATA_API_URL = "https://sapnhap.bando.com.vn/p.co_dvhc_id"

type SapNhapService struct {
	sapNhapRepo        *repository.SapNhapRepository
	sapNhapGISRepo     *repository.SapNhapGISRepository
	sapNhapGeoJSONRepo *repository.SapNhapGeoJSONObjectRepository
	vnProvinceTmpRepo  *vnRepo.VnProvincesTmpRepository
	db                 *bun.DB
}

func NewSapNhapService(repo *repository.SapNhapRepository, sapNhapGISRepo *repository.SapNhapGISRepository, vnRepo *vnRepo.VnProvincesTmpRepository,
	sapNhapGeoJSONRepo *repository.SapNhapGeoJSONObjectRepository, db *bun.DB) *SapNhapService {
	return &SapNhapService{
		sapNhapRepo:        repo,
		sapNhapGISRepo:     sapNhapGISRepo,
		sapNhapGeoJSONRepo: sapNhapGeoJSONRepo,
		vnProvinceTmpRepo:  vnRepo,
		db:                 db,
	}
}

// cleanAdministrativeUnitPrefix removes common administrative unit prefixes from a name
// in a case-insensitive way, preserving the case of the original name.
func cleanAdministrativeUnitPrefix(name string) string {
	// Prefixes are in lowercase and include a space for accurate matching.
	prefixes := []string{"tỉnh ", "thành phố ", "thủ đô "}
	lowerName := strings.ToLower(name)

	for _, prefix := range prefixes {
		if strings.HasPrefix(lowerName, prefix) {
			// Slice the original string to preserve the casing of the name part.
			return strings.TrimSpace(name[len(prefix):])
		}
	}
	return name // Return original name if no prefix is found
}

func (s *SapNhapService) BootstrapSapNhapSiteProvinces() error {
	ctx := context.Background() // Use context.Background for top-level contexts in scripts.

	// Load provinces from JSON file instead of deprecated API
	sapNhapSiteProvinces, err := fetcher.LoadProvincesFromJSONFile()
	if err != nil {
		return fmt.Errorf("failed to load provinces from JSON file: %w", err)
	}

	// Insert each province into the repository
	for _, provinceData := range sapNhapSiteProvinces {
		// Clean the province name by removing administrative unit prefixes.
		cleanedProvinceName := cleanAdministrativeUnitPrefix(provinceData.TenTinh)
		cleanedProvinceName = util.NormalizeString(cleanedProvinceName)

		// Attempt to look up the vn province by name
		vnProvince, err := s.vnProvinceTmpRepo.FindProvinceByName(ctx, cleanedProvinceName)
		if err != nil {
			return fmt.Errorf("error finding province by name '%s': %w", cleanedProvinceName, err)
		}
		if vnProvince == nil {
			return fmt.Errorf("VN Province not found for name: %s", cleanedProvinceName)
		}

		province := &model.SapNhapSiteProvince{
			ID:             provinceData.ID,
			MaHC:           provinceData.MaHC,
			TenTinh:        cleanedProvinceName,
			DienTichKm2:    provinceData.DienTichKm2,
			DanSoNguoi:     provinceData.DanSoNguoi,
			TrungTamHC:     provinceData.TrungTamHC,
			KinhDo:         provinceData.KinhDo,
			ViDo:           provinceData.ViDo,
			TruocSN:        provinceData.TruocSN,
			Con:            provinceData.Con,
			VNProvinceCode: vnProvince.Code,
		}

		if err := s.sapNhapRepo.InsertSapNhapSiteProvince(province); err != nil {
			// Correctly wrap and return the error.
			return fmt.Errorf("failed to insert province %s: %w", province.TenTinh, err)
		}
	}

	log.Default().Println("Bootstrap SapNhapSiteProvinces completed successfully")

	return nil
}

func (s *SapNhapService) BootstrapSapNhapSiteWards() error {
	ctx := context.TODO()

	// Load all wards from JSON file instead of fetching per province from deprecated API
	sapNhapSiteWards, err := fetcher.LoadWardsFromJSONFile()
	if err != nil {
		return fmt.Errorf("failed to load wards from JSON file: %w", err)
	}

	for _, wardData := range sapNhapSiteWards {
		/* Very edge case, for Tuyen Quang, Phố Bảng is actually Phó Bảng */
		if strings.EqualFold(wardData.TenHC, "Phố Bảng") && wardData.Matinh == 8 {
			wardData.TenHC = "Phó Bảng"
		}

		wardData.TenHC = util.NormalizeString(wardData.TenHC)

		// Find province first to get vn_province_code
		province, err := s.sapNhapRepo.FindSapNhapSiteProvinceByMaHC(ctx, wardData.Matinh)
		if err != nil {
			return fmt.Errorf("error finding province by MaHC '%d': %w", wardData.Matinh, err)
		}
		if province == nil {
			return fmt.Errorf("SapNhap Province not found for MaHC: %d", wardData.Matinh)
		}

		vnWard, err := s.vnProvinceTmpRepo.FindWardByName(ctx, strings.TrimSpace(wardData.TenHC), province.VNProvinceCode)
		if err != nil {
			return fmt.Errorf("error finding ward by name '%s': %w", wardData.TenHC, err)
		}
		if vnWard == nil {
			return fmt.Errorf("VN Ward not found for name: %s", wardData.TenHC)
		}

		ward := &model.SapNhapSiteWard{
			ID:           wardData.ID,
			MaTinh:       wardData.Matinh,
			Ma:           wardData.Ma,
			TenTinh:      province.TenTinh, // Use province name from DB
			Loai:         wardData.Loai,
			TenHC:        wardData.TenHC,
			Cay:          wardData.Cay,
			DienTichKm2:  wardData.DienTichKm2,
			DanSoNguoi:   wardData.DanSoNguoi,
			TrungTamHC:   wardData.TrungTamHC,
			KinhDo:       wardData.KinhDo,
			ViDo:         wardData.ViDo,
			TruocSapNhap: wardData.TruocSapNhap,
			MaXa:         wardData.MaXa,
			Khoa:         wardData.Khoa,
			VNWardCode:   vnWard.Code,
		}

		if err := s.sapNhapRepo.InsertSapNhapSiteWard(ward); err != nil {
			return fmt.Errorf("failed to insert ward %s in province %s: %w", ward.TenHC, ward.TenTinh, err)
		}
	}

	log.Default().Println("Bootstrap SapNhapSiteWards completed successfully")
	return nil
}

/*
Bootstrap GIS data coordinate by loading from local GeoJSON files
Replaces the old API-based approach that used GetGISLocationCoordinates()
*/
func (s *SapNhapService) BootstrapGISDataFromGISServer() error {
	// Load province GIS data from GeoJSON files
	provincesGIS, err := fetcher.LoadProvincesGISFromGeoJSONFiles()
	if err != nil {
		return fmt.Errorf("failed to load provinces GIS data: %w", err)
	}

	// Insert province GIS data
	for _, gisData := range provincesGIS {
		matinhInt, err := strconv.Atoi(gisData.SapNhapProvinceMaTinh)
		if err != nil {
			log.Printf("Warning: failed to parse matinh %s for %s: %v", gisData.SapNhapProvinceMaTinh, gisData.Ten, err)
			continue
		}

		sapNhapProvinceGIS := &model.SapNhapProvinceGIS{
			Stt:                   gisData.STT,
			Ten:                   gisData.Ten,
			TruocSapNhap:          gisData.TruocSN,
			GISServerID:           gisData.GISServerID,
			SapNhapProvinceMaTinh: matinhInt,
			BBoxWKT:               gisData.BBoxWKT,
			GeomWKT:               gisData.GeomWKT,
		}

		if err := s.sapNhapGISRepo.InsertSapNhapProvinceGIS(sapNhapProvinceGIS); err != nil {
			log.Printf("Warning: failed to insert province GIS data for %s: %v", gisData.Ten, err)
		} else {
			log.Printf("Inserted GIS data for province %s", gisData.Ten)
		}
	}

	// Load ward GIS data from GeoJSON files
	wardsGIS, err := fetcher.LoadWardsGISFromGeoJSONFiles()
	if err != nil {
		return fmt.Errorf("failed to load wards GIS data: %w", err)
	}

	// Insert ward GIS data
	for _, gisData := range wardsGIS {
		maxaInt, err := strconv.Atoi(gisData.SapNhapWardMaXa)
		if err != nil {
			log.Printf("Warning: failed to parse maxa %s for %s: %v", gisData.SapNhapWardMaXa, gisData.Ten, err)
			continue
		}

		sapNhapWardGIS := &model.SapNhapWardGIS{
			Stt:             gisData.STT,
			Ten:             gisData.Ten,
			TruocSapNhap:    gisData.TruocSN,
			GISServerID:     gisData.GISServerID,
			SapNhapWardMaXa: maxaInt,
			BBoxWKT:         gisData.BBoxWKT,
			GeomWKT:         gisData.GeomWKT,
		}

		if err := s.sapNhapGISRepo.InsertSapNhapWardGIS(sapNhapWardGIS); err != nil {
			log.Printf("Warning: failed to insert ward GIS data for %s: %v", gisData.Ten, err)
		} else {
			log.Printf("Inserted GIS data for ward %s", gisData.Ten)
		}
	}

	log.Printf("Completed loading GIS data: %d provinces, %d wards", len(provincesGIS), len(wardsGIS))
	return nil
}

func (s *SapNhapService) BootstrapGISDataFromGISServerV2() error {
	return nil
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

	correctedCount, err := geoJSONRepo.CorrectMismatchedBBoxWKTFromGeom(ctx)
	if err != nil {
		return fmt.Errorf("failed to correct mismatched bbox_wkt from geom after GIS import: %w", err)
	}
	log.Printf("Post-import bbox correction complete. Corrected %d rows using ST_Envelope(geom)", correctedCount)

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
func loadAnGiangProvinceFromLocalFile() (wktBBox string, wktGeometry string, err error) {
	// Path to the manual patch file
	const anGiangPatchPath = "./resources/gis/geojson_11Mar2026/32_tinh_an_giang/province.geojson"

	// Load GeoJSON file
	geojson, err := dto.LoadGeoJSONFile(anGiangPatchPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to load An Giang patch file from %s: %w", anGiangPatchPath, err)
	}

	if len(geojson.Features) == 0 {
		return "", "", fmt.Errorf("no features found in An Giang patch file")
	}

	feature := geojson.Features[0]

	// Convert to WKT format
	wktBBox = feature.ToWKBboxPolygon()
	wktGeometry = feature.Geometry.ToWKTMultiPolygon()

	return wktBBox, wktGeometry, nil
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

		wktBBox, wktGeometry, err := loadAnGiangProvinceFromLocalFile()
		if err != nil {
			return fmt.Errorf("failed to load An Giang province from local patch file: %w", err)
		}

		// Update the database with patched data
		err = geoJSONRepo.UpdateSapNhapGeoJSONObjectWKT(ctx, geoObject.Ma, wktBBox, wktGeometry)
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

	// Convert to WKT format
	wktBBox := feature.BBox.ToWKTPolygon()
	wktGeometry := feature.Geometry.ToWKTCoordinate()

	// Update the database
	err = geoJSONRepo.UpdateSapNhapGeoJSONObjectWKT(ctx, geoObject.Ma, wktBBox, wktGeometry)
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

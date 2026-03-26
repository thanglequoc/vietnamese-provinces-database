package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/thanglequoc-vn-provinces/v2/internal/common/viet"
	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/fetcher"
	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/repository"
	vnRepo "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/repository"
	"golang.org/x/text/unicode/norm"
)

const BANDO_GIS_PROVINCES_FILE_PATH = "./resources/gis/bando_gisserver/provinces.json"
const BANDO_GIS_WARDS_FILE_PATH = "./resources/gis/bando_gisserver/wards.json"

type SapNhapService struct {
	sapNhapRepo       *repository.SapNhapRepository
	sapNhapGISRepo    *repository.SapNhapGISRepository
	vnProvinceTmpRepo *vnRepo.VnProvincesTmpRepository
}

func NewSapNhapService(repo *repository.SapNhapRepository, sapNhapGISRepo *repository.SapNhapGISRepository, vnRepo *vnRepo.VnProvincesTmpRepository) *SapNhapService {
	return &SapNhapService{
		sapNhapRepo:       repo,
		sapNhapGISRepo:    sapNhapGISRepo,
		vnProvinceTmpRepo: vnRepo,
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
	sapNhapSiteProvinces := fetcher.GetAllProvincesDataFromSapNhapSite()

	// Insert each province into the repository
	for _, provinceData := range sapNhapSiteProvinces {
		// Clean the province name by removing administrative unit prefixes.
		cleanedProvinceName := cleanAdministrativeUnitPrefix(provinceData.TenTinh)
		cleanedProvinceName = normalizeString(cleanedProvinceName)

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
	sapNhapSiteProvinces, err := s.sapNhapRepo.GetAllSapNhapSiteProvinces()
	if err != nil {
		return fmt.Errorf("failed to get SapNhapSiteProvinces: %w", err)
	}

	for _, province := range sapNhapSiteProvinces {
		sapNhapWards := fetcher.GetAllWardsOfProvinceFromSapNhapSite(province.MaHC)
		if len(sapNhapWards) == 0 {
			log.Printf("No wards found for province ID %d (%s)", province.ID, province.TenTinh)
			return fmt.Errorf("no wards found for province ID %d (%s)", province.ID, province.TenTinh)
		}

		for _, wardData := range sapNhapWards {
			/* Very edge case, for Tuyen Quang, Phố Bảng is actually Phó Bảng */
			if strings.EqualFold(wardData.TenHC, "Phố Bảng") && province.VNProvinceCode == "08" {
				wardData.TenHC = "Phó Bảng"
			}

			wardData.TenHC = normalizeString(wardData.TenHC)

			vnWard, err := s.vnProvinceTmpRepo.FindWardByName(ctx, strings.TrimSpace(wardData.TenHC), province.VNProvinceCode)
			if err != nil {
				return fmt.Errorf("error finding ward by name '%s': %w", wardData.TenHC, err)
			}
			if vnWard == nil {
				return fmt.Errorf("VN Ward not found for name: %s", wardData.TenHC)
			}

			ward := &model.SapNhapSiteWard{
				ID:           wardData.ID,
				MaTinh:       province.MaHC,
				Ma:           wardData.Ma,
				TenTinh:      province.TenTinh,
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
				fmt.Errorf("failed to insert ward %s in province %s: %w", ward.TenTinh, province.TenTinh, err)
				return err
			}
		}
	}

	return nil
}

/*
Bootstrap GIS data coordinate by fetching from sapnhap bando GIS server
*/
func (s *SapNhapService) BootstrapGISDataFromGISServer() error {
	// Read from JSON file
	// Provinces Data
	bandoProvinces, err := fetcher.LoadBanDoGISProvincesFromFile(BANDO_GIS_PROVINCES_FILE_PATH)
	if err != nil {
		return err
	}
	for _, bandoProvince := range bandoProvinces {
		gisRes, err := bandoProvince.GetGISResponse()
		if err != nil {
			return err
		}

		// Normally there should just be a single element in the features
		if len(gisRes.Features) > 1 {
			return fmt.Errorf("got more than 1 feature from the GIS response of %v", bandoProvince)
		}

		id := gisRes.Features[0].ID
		matinh := fmt.Sprintf("%v", gisRes.Features[0].Properties["matinh"])

		fmt.Printf("Fetching GIS coordinate data for %s - {gisServerID: %s - matinh: %s}\n", bandoProvince.Ten, id, matinh)
		gisCoordinateResponse, err := fetcher.GetGISLocationCoordinates(id)
		if err != nil {
			log.Printf("Unable to get GIS Coordinate Response of location [Name: %s - ID %s]. Error: %v", bandoProvince.Ten, id, err)
		}

		sttNumber, err := strconv.Atoi(bandoProvince.STT)
		if err != nil {
			return err
		}
		matinhNumber, err := strconv.Atoi(matinh)
		if err != nil {
			return err
		}

		wktBBoxPolygon := gisCoordinateResponse.Features[0].BBox.ToWKTPolygon()
		wktMultiPolygon := gisCoordinateResponse.Features[0].Geometry.ToWKTCoordinate()

		sapNhapProvinceGIS := &model.SapNhapProvinceGIS{
			Stt:                   sttNumber,
			Ten:                   bandoProvince.Ten,
			TruocSapNhap:          bandoProvince.TruocSN,
			GISServerID:           id,
			SapNhapProvinceMaTinh: matinhNumber,
			BBoxWKT:               wktBBoxPolygon,
			GeomWKT:               wktMultiPolygon,
		}

		if err := s.sapNhapGISRepo.InsertSapNhapProvinceGIS(sapNhapProvinceGIS); err != nil {
			log.Fatalf("Unable to insert to sapnhap_province_gis table %v. Error: %v", sapNhapProvinceGIS, err)
			return err
		}

		fmt.Printf("Inserted GIS data for %s complete \n", bandoProvince.Ten)
		fmt.Println("-- ---------------------------------------")
	}

	// Wards data
	bandoWards, err := fetcher.LoadBanDoGISWardsFromFile(BANDO_GIS_WARDS_FILE_PATH)
	if err != nil {
		return err
	}

	for _, bandoWard := range bandoWards {
		gisRes, err := bandoWard.GetGISResponse()
		if err != nil {
			return err
		}
		// Normally there should just be a single element in the features
		if len(gisRes.Features) > 1 {
			return fmt.Errorf("got more than 1 feature from the GIS response of %s", bandoWard.Ten)
		}

		id := gisRes.Features[0].ID
		maxa := fmt.Sprintf("%v", gisRes.Features[0].Properties["maxa"])
		fmt.Printf("Fetching GIS coordinate data for %s - {gisServerID: %s - maxa: %s}\n", bandoWard.Ten, id, maxa)
		gisCoordinateResponse, err := fetcher.GetGISLocationCoordinates(id)
		if err != nil {
			log.Printf("Unable to get GIS Coordinate Response of location [Name: %s - ID %s]. Error: %v", bandoWard.Ten, id, err)
		}

		sttNumber, err := strconv.Atoi(bandoWard.STT)
		if err != nil {
			return err
		}
		maxaNumber, err := strconv.Atoi(maxa)
		if err != nil {
			return err
		}

		wktBBoxPolygon := gisCoordinateResponse.Features[0].BBox.ToWKTPolygon()
		wktMultiPolygon := gisCoordinateResponse.Features[0].Geometry.ToWKTCoordinate()

		sapNhapWardGIS := &model.SapNhapWardGIS{
			Stt:             sttNumber,
			Ten:             bandoWard.Ten,
			TruocSapNhap:    bandoWard.TruocSN,
			GISServerID:     id,
			SapNhapWardMaXa: maxaNumber,
			BBoxWKT:         wktBBoxPolygon,
			GeomWKT:         wktMultiPolygon,
		}

		if err := s.sapNhapGISRepo.InsertSapNhapWardGIS(sapNhapWardGIS); err != nil {
			log.Fatalf("Unable to insert to sapnhap_ward_gis table %v. Error: %v", sapNhapWardGIS, err)
			return err
		}

		fmt.Printf("Inserted GIS data for %s complete \n", bandoWard.Ten)
		fmt.Println("-- ---------------------------------------")
	}

	fmt.Printf("Total %d provinces are processed with GIS data successfully", len(bandoProvinces))
	fmt.Printf("Total %d wards are processed with GIS data successfully", len(bandoWards))

	return nil
}

func (s *SapNhapService) BootstrapGISDataFromGISServerV2() error {
	return nil
}

// ProcessingError represents a record that failed to process
type ProcessingError struct {
	Ma     string
	Ten    string
	MaLK   string
	Error  string
}

// ProcessingResult represents the result of processing a single geo object
type ProcessingResult struct {
	Success bool
	Error    ProcessingError
}

// FetchGISDataFromSapNhapBando fetches WKT geometry data from the Bando GIS server
// for all records in the sapnhap_geojson_objects table and updates them with the retrieved data.
// Uses parallel processing with a worker pool for improved performance.
func (s *SapNhapService) FetchGISDataFromSapNhapBando(geoJSONRepo *repository.SapNhapGeoJSONObjectRepository) error {
	// Get all geo objects from the database
	geoObjects, err := geoJSONRepo.GetAllSapNhapGeoJSONObjects()
	if err != nil {
		return fmt.Errorf("failed to get sapnhap geojson objects: %w", err)
	}

	log.Printf("Found %d geo objects to process", len(geoObjects))

	// Number of concurrent workers
	numWorkers := 10
	log.Printf("Processing with %d concurrent workers", numWorkers)

	ctx := context.Background()
	
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
processGeoJSONObject fetches GIS data for a single geo object and updates the database
*/
func (s *SapNhapService) processGeoJSONObject(ctx context.Context, geoObject *model.SapNhapSiteGeoUnit, geoJSONRepo *repository.SapNhapGeoJSONObjectRepository) error {
	// Check if the geo object has a MaLK value
	if geoObject.MaLK == "" {
		return fmt.Errorf("malk field is empty, cannot fetch GIS data")
	}

	log.Printf("Fetching GIS data for [ma: %s, ten: %s, malk: %s]", geoObject.Ma, geoObject.Ten, geoObject.MaLK)

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

// normalizeString handles all normalization steps for Vietnamese text
// 1. Replaces smart apostrophes with standard apostrophes
// 2. Normalizes to NFC form to handle decomposed Unicode characters
// This ensures consistent encoding across all data
func normalizeString(s string) string {
	// Replace smart apostrophe with standard apostrophe
	result := strings.ReplaceAll(s, "’", "'")
	result = viet.NormalizeToneMarks(result)
	
	// Normalize to NFC form to handle decomposed characters
	// This ensures "X ̃" (decomposed) becomes "Xã" (precomposed)
	result = norm.NFC.String(result)
	
	return result
}
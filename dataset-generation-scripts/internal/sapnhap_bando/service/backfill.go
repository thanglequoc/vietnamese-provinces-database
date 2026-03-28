package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/repository"
	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/util"
	vnModel "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
	vnRepo "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/repository"
	"github.com/thanglequoc-vn-provinces/v2/internal/common/viet"
)

// SapNhapBackfillService handles backfill operations for province and ward codes
type SapNhapBackfillService struct {
	vnProvinceTmpRepo *vnRepo.VnProvincesTmpRepository
	geoJSONRepo       *repository.SapNhapGeoJSONObjectRepository
}

// NewSapNhapBackfillService creates a new backfill service instance
func NewSapNhapBackfillService(vnRepo *vnRepo.VnProvincesTmpRepository, geoJSONRepo *repository.SapNhapGeoJSONObjectRepository) *SapNhapBackfillService {
	return &SapNhapBackfillService{
		vnProvinceTmpRepo: vnRepo,
		geoJSONRepo:       geoJSONRepo,
	}
}

// BackfillError represents a record that failed to match during backfill
// This type is exported for use by the main SapNhapService
type BackfillError struct {
	Ma         string
	Ten        string
	MaGoc      string
	IsProvince bool
	Error      string
}

// createProvinceLookupMap creates a normalized name to code lookup map for provinces
func (s *SapNhapBackfillService) createProvinceLookupMap(provinces []vnModel.Province) (map[string]string, error) {
	lookupMap := make(map[string]string)
	
	for _, province := range provinces {
		// Use Name (Vietnamese name with tone marks) for matching
		normalizedName := strings.ToLower(strings.TrimSpace(province.Name))
		normalizedName = viet.NormalizeToneMarks(normalizedName)
		
		// Check for duplicates (should not happen per requirements)
		if existingCode, exists := lookupMap[normalizedName]; exists {
			return nil, fmt.Errorf("duplicate normalized province name '%s' found for codes '%s' and '%s'", 
				normalizedName, existingCode, province.Code)
		}
		
		lookupMap[normalizedName] = province.Code
	}
	
	return lookupMap, nil
}

// createWardLookupMap creates a normalized (ward name, province name) tuple to code lookup map for wards
func (s *SapNhapBackfillService) createWardLookupMap(wards []vnModel.Ward, provinces []vnModel.Province) (map[string]string, error) {
	// Create a province code to normalized name map for easy lookup
	provinceNameMap := make(map[string]string)
	for _, province := range provinces {
		normalizedName := strings.ToLower(strings.TrimSpace(province.Name))
		normalizedName = viet.NormalizeToneMarks(normalizedName)
		provinceNameMap[province.Code] = normalizedName
	}
	
	lookupMap := make(map[string]string)
	
	for _, ward := range wards {
		// Use Name (Vietnamese name with tone marks) for matching
		normalizedWardName := strings.ToLower(strings.TrimSpace(ward.Name))
		normalizedWardName = viet.NormalizeToneMarks(normalizedWardName)
		
		// Get the parent province normalized name
		provinceName, exists := provinceNameMap[ward.ProvinceCode]
		if !exists {
			return nil, fmt.Errorf("ward code '%s' references non-existent province code '%s'", 
				ward.Code, ward.ProvinceCode)
		}
		
		// Create composite key: normalizedWardName + "|" + normalizedProvinceName
		compositeKey := normalizedWardName + "|" + provinceName
		
		// Check for duplicates - this should NOT happen as each ward should be unique
		if existingCode, exists := lookupMap[compositeKey]; exists {
			return nil, fmt.Errorf("duplicate normalized ward key '%s' found for codes '%s' and '%s'. " +
				"This indicates data inconsistency - multiple wards with the same normalized name in the same province.",
				compositeKey, existingCode, ward.Code)
		}
		
		lookupMap[compositeKey] = ward.Code
	}
	
	return lookupMap, nil
}

// ExecuteBackfill populates vn_ds_province_code and vn_ds_ward_code fields
// in sapnhap_geojson_objects table by matching names against provinces_tmp and wards_tmp tables
func (s *SapNhapBackfillService) ExecuteBackfill(ctx context.Context) error {
	log.Println("Starting backfill of province and ward codes...")
	
	// Load all provinces and wards into memory
	provinces := s.vnProvinceTmpRepo.GetAllProvinces()
	log.Printf("Loaded %d provinces from provinces_tmp", len(provinces))
	
	wards := s.vnProvinceTmpRepo.GetAllWards()
	log.Printf("Loaded %d wards from wards_tmp", len(wards))
	
	// Create normalized lookup maps
	provinceLookupMap, err := s.createProvinceLookupMap(provinces)
	if err != nil {
		return fmt.Errorf("failed to create province lookup map: %w", err)
	}
	
	wardLookupMap, err := s.createWardLookupMap(wards, provinces)
	if err != nil {
		return fmt.Errorf("failed to create ward lookup map: %w", err)
	}
	
	// Get all geo objects
	geoObjects, err := s.geoJSONRepo.GetAllSapNhapGeoJSONObjects()
	if err != nil {
		return fmt.Errorf("failed to get sapnhap geojson objects: %w", err)
	}
	
	log.Printf("Processing %d geo objects for backfill", len(geoObjects))
	
	// Create a map of geo objects by ma for parent lookup
	geoObjectMap := make(map[string]*model.SapNhapSiteGeoUnit)
	for _, obj := range geoObjects {
		geoObjectMap[obj.Ma] = obj
	}
	
	// Collect updates for batch processing
	updates := make([]repository.CodeUpdate, 0)
	backfillErrors := make([]BackfillError, 0)
	
	for _, geoObject := range geoObjects {
		normalizedTen := util.NormalizeForMatching(geoObject.Ten)
		
		// Handle edge case: xa254 "Xã Phố Bảng" should match "Xã Phó Bảng" (Tuyên Quang province)
		// This is a known typo in the upstream source data specific to this record.
		// 
		// ROOT CAUSE: The geo object has "Phố Bảng" (with "ố") but wards_tmp has "Phó Bảng" (with "ó").
		// The typo is: "Phố" → "Phó" (changing the tone mark from "ố" to "ó")
		// 
		// SOLUTION: Detect this specific case and override the lookup key to use the corrected name.
		// We use the corrected raw name "Xã Phó Bảng" then normalize it, which gives us "xã phó bảng".
		// This matches the lookup map key which was built from wards_tmp's corrected name.
		if geoObject.Ma == "xa254" {
			// Remove prefix to isolate the name part for checking
			nameWithoutPrefix := util.RemoveAdministrativeUnitPrefix(geoObject.Ten)
			normalizedRaw := strings.ToLower(strings.TrimSpace(nameWithoutPrefix))
			normalizedRaw = viet.NormalizeToneMarks(normalizedRaw)
			
			// Check for the typo: "phố bảng" (with ố) should be corrected to "phó bảng" (with ó)
			if normalizedRaw == "phố bảng" {
				// Use the corrected raw name and normalize it
				// "Xã Phó Bảng" → normalize → "xã phó bảng" (with prefix and corrected ó)
				correctedRawName := "Xã Phó Bảng"
				normalizedTen = util.NormalizeForMatching(correctedRawName)
			}
		}
		
		// Province matching: magoc IS NULL
		if geoObject.MaGoc == "" {
			provinceCode, exists := provinceLookupMap[normalizedTen]
			if !exists {
				backfillErrors = append(backfillErrors, BackfillError{
					Ma:         geoObject.Ma,
					Ten:        geoObject.Ten,
					MaGoc:      geoObject.MaGoc,
					IsProvince: true,
					Error:      fmt.Sprintf("no matching province found for normalized name '%s'", normalizedTen),
				})
				continue
			}
			
			updates = append(updates, repository.CodeUpdate{
				Ma:               geoObject.Ma,
				VNDSProvinceCode: provinceCode,
				VNDSWardCode:     "",
			})
		} else {
			// Ward matching: magoc IS NOT NULL
			// Get parent province name from the parent record
			parent, exists := geoObjectMap[geoObject.MaGoc]
			if !exists {
				backfillErrors = append(backfillErrors, BackfillError{
					Ma:         geoObject.Ma,
					Ten:        geoObject.Ten,
					MaGoc:      geoObject.MaGoc,
					IsProvince: false,
					Error:      fmt.Sprintf("parent record with ma '%s' not found", geoObject.MaGoc),
				})
				continue
			}
			
			normalizedParentName := util.NormalizeForMatching(parent.Ten)
			compositeKey := normalizedTen + "|" + normalizedParentName
			
			wardCode, exists := wardLookupMap[compositeKey]
			if !exists {
				backfillErrors = append(backfillErrors, BackfillError{
					Ma:         geoObject.Ma,
					Ten:        geoObject.Ten,
					MaGoc:      geoObject.MaGoc,
					IsProvince: false,
					Error:      fmt.Sprintf("no matching ward found for composite key '%s'", compositeKey),
				})
				continue
			}
			
			updates = append(updates, repository.CodeUpdate{
				Ma:               geoObject.Ma,
				VNDSProvinceCode: "",
				VNDSWardCode:     wardCode,
			})
		}
	}
	
	// Batch update records with matched codes
	if len(updates) > 0 {
		err = s.geoJSONRepo.BatchUpdateProvinceAndWardCodes(ctx, updates)
		if err != nil {
			return fmt.Errorf("failed to batch update province and ward codes: %w", err)
		}
		log.Printf("Successfully updated %d records with province/ward codes", len(updates))
	} else {
		log.Println("No records to update")
	}
	
	// Log backfill errors
	if len(backfillErrors) > 0 {
		log.Printf("\n==========================================")
		log.Printf("BACKFILL ERRORS SUMMARY:")
		log.Printf("==========================================")
		for i, be := range backfillErrors {
			recordType := "Province"
			if !be.IsProvince {
				recordType = "Ward"
			}
			log.Printf("%d. %s [MA: %s, TEN: %s, MAGOC: %s]", i+1, recordType, be.Ma, be.Ten, be.MaGoc)
			log.Printf("   Error: %s", be.Error)
			log.Println("------------------------------------------")
		}
		log.Printf("==========================================")
		log.Printf("Total backfill errors: %d out of %d records", len(backfillErrors), len(geoObjects))
		log.Printf("==========================================\n")
		
		// Return error if there were failures, but processing has completed
		return fmt.Errorf("backfill completed with %d errors out of %d total records. See above for details.", 
			len(backfillErrors), len(geoObjects))
	}
	
	log.Println("Backfill completed successfully with no errors")
	return nil
}
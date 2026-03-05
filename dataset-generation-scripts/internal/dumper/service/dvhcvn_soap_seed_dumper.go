package service

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"
	"golang.org/x/text/unicode/norm"

	data_downloader "github.com/thanglequoc-vn-provinces/v2/internal/dvhcvn_data_downloader"
	db "github.com/thanglequoc-vn-provinces/v2/internal/database"

	"github.com/thanglequoc-vn-provinces/v2/internal/common/viet"
	"github.com/thanglequoc-vn-provinces/v2/internal/dumper/config"
	"github.com/thanglequoc-vn-provinces/v2/internal/dumper/helper"
	vn_provinces_tmp_model "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
	vn_repo "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/repository"
)

type DvhcvnSoapSeedDumperService struct {
	vnProvinceTmpRepo *vn_repo.VnProvincesTmpRepository
}

func NewDvhcvnSoapSeedDumperService() *DvhcvnSoapSeedDumperService {
	return &DvhcvnSoapSeedDumperService{
		vnProvinceTmpRepo: vn_repo.NewVnProvincesTmpRepository(db.GetPostgresDBConnection()),
	}
}

func (s *DvhcvnSoapSeedDumperService) BeginDumpingDataWithDvhcvnDirectSource() {
	fmt.Print("(Optional) Please specify the data date (dd/MM/YYYY). Leave empty to go with default option: ")

	reader := bufio.NewReader(os.Stdin)
	userInput, _ := reader.ReadString('\n')
	userInput = strings.TrimSpace(userInput)
	fmt.Println("Selected date: ", userInput)

	var dataSetTime time.Time
	if len(strings.TrimSpace(userInput)) == 0 {
		fmt.Println("No input is recorded, using tomorrow as the default day...")
		dataSetTime = time.Now().Add(time.Hour * 24)
	} else {
		dataSetTime, _ = time.Parse("02/01/2006", userInput) // dd/MM/yyyy
	}

	dvhcvnUnits := data_downloader.FetchDvhcvnData(dataSetTime)

	s.insertToProvinces(dvhcvnUnits.ProvinceData)
	s.insertToWards(dvhcvnUnits.WardData)
	fmt.Println("📥 Dumper operation finished")
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

// capitalizeWords capitalizes the first letter of each word in a string
// Uses rune indexing to properly handle multi-byte UTF-8 characters
// Preserves the case of the rest of the word (important for proper names like "H'Leo")
func capitalizeWords(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			runes := []rune(word)
			if len(runes) > 0 {
				// Capitalize first rune only - preserve rest of the word as-is
				runes[0] = unicode.ToUpper(runes[0])
				words[i] = string(runes)
			}
		}
	}
	return strings.Join(words, " ")
}

func (s *DvhcvnSoapSeedDumperService) insertToWards(dvhcvnWardModels []data_downloader.DvhcvnWardModel) {
	ctx := context.Background()
	totalWard := 0

	for _, w := range dvhcvnWardModels {
		wardFullName := helper.RemoveWhiteSpaces(w.WardName)
		administrativeUnitLevel := helper.GetAdministrativeUnit_WardLevel(wardFullName)
		
		unitName := config.AdministrativeUnitNamesShortNameMap_vn[administrativeUnitLevel]
		unitName_en := config.AdministrativeUnitNamesShortNameMap_en[administrativeUnitLevel]
		
		// Normalize strings (apostrophe replacement + NFC normalization)
		wardFullNameNormalized := normalizeString(wardFullName)
		unitNameNormalized := normalizeString(unitName)
		
		// Remove unit name (case-insensitive)
		wardShortName := strings.Trim(strings.Replace(wardFullNameNormalized, unitNameNormalized, "", 1), " ")
		// Capitalize first letter of each word using rune-aware function
		wardShortName = capitalizeWords(wardShortName)
		// Normalize again after capitalization to ensure consistent encoding
		wardShortName = normalizeString(wardShortName)
		
		codeName := helper.ToCodeName(wardShortName)
		wardShortNameEn := viet.RemoveVietToneMark(wardShortName)

		// Case when ward name is a number
		isNumber, _ := regexp.MatchString("[0-9]+", wardShortName)
		var wardFullNameEn string
		if isNumber {
			wardFullNameEn = unitName_en + " " + wardShortNameEn
		} else {
			wardFullNameEn = wardShortNameEn + " " + unitName_en
		}

		wardModel := &vn_provinces_tmp_model.Ward{
			Code:                 w.WardCode,
			Name:                 wardShortName,
			NameEn:               wardShortNameEn,
			FullName:             wardFullNameNormalized,
			FullNameEn:           wardFullNameEn,
			CodeName:             codeName,
			AdministrativeUnitId: administrativeUnitLevel,
			ProvinceCode:         w.ProvinceCode,
		}

		err := s.vnProvinceTmpRepo.InsertWard(ctx, wardModel)
		totalWard++
		if err != nil {
			fmt.Println(err)
			panic("Exception happens while inserting into wards table")
		}
	}

	fmt.Printf("Inserted %d wards to tables\n", totalWard)
}

func (s *DvhcvnSoapSeedDumperService) insertToProvinces(dvhcvnProvinceModels []data_downloader.DvhcvnProvinceModel) {
	ctx := context.Background()

	for _, p := range dvhcvnProvinceModels {
		provinceFullName := helper.RemoveWhiteSpaces(p.ProvinceName)
		administrativeUnitLevel := helper.GetAdministrativeUnit_ProvinceLevel(provinceFullName)
		
		unitName := config.AdministrativeUnitNamesShortNameMap_vn[administrativeUnitLevel]
		unitName_en := config.AdministrativeUnitNamesShortNameMap_en[administrativeUnitLevel]
		
		// Normalize strings (apostrophe replacement + NFC normalization)
		provinceFullNameNormalized := normalizeString(provinceFullName)
		unitNameNormalized := normalizeString(unitName)
		
		// Remove unit name (case-insensitive)
		provinceShortName := strings.Trim(strings.Replace(provinceFullNameNormalized, unitNameNormalized, "", 1), " ")
		// Capitalize first letter of each word using rune-aware function
		provinceShortName = capitalizeWords(provinceShortName)
		// Normalize again after capitalization to ensure consistent encoding
		provinceShortName = normalizeString(provinceShortName)
		
		codeName := helper.ToCodeName(provinceShortName)
		provinceShortNameEn := viet.RemoveVietToneMark(provinceShortName)
		provinceFullNameEn := provinceShortNameEn + " " + unitName_en

		provinceModel := &vn_provinces_tmp_model.Province{
			Code:                 p.ProvinceCode,
			Name:                 provinceShortName,
			NameEn:               provinceShortNameEn,
			FullName:             provinceFullNameNormalized,
			FullNameEn:           provinceFullNameEn,
			CodeName:             codeName,
			AdministrativeUnitId: administrativeUnitLevel,
		}

		err := s.vnProvinceTmpRepo.InsertProvince(ctx, provinceModel)
		if err != nil {
			fmt.Println(err)
			panic("Exception happens while inserting into provinces table")
		}
	}

	fmt.Printf("Inserted %d provinces to tables\n", len(dvhcvnProvinceModels))
}

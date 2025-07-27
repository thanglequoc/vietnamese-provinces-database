package service

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	db "github.com/thanglequoc-vn-provinces/v2/internal/database"

	cfg "github.com/thanglequoc-vn-provinces/v2/internal/dumper/config"
	"github.com/thanglequoc-vn-provinces/v2/internal/dumper/helper"
	"github.com/thanglequoc-vn-provinces/v2/internal/dumper/model"
	"github.com/thanglequoc-vn-provinces/v2/internal/dumper/repository"

	vn_provinces_tmp_model "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
	vn_repo "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/repository"
)

type ManualSeedDumperService struct {
	seedRepo          *repository.SeedDataRepository
	vnProvinceTmpRepo *vn_repo.VnProvincesTmpRepository
}

func NewManualSeedDumperService() *ManualSeedDumperService {
	return &ManualSeedDumperService{
		seedRepo:          repository.NewSeedDataRepository(db.GetPostgresDBConnection()),
		vnProvinceTmpRepo: vn_repo.NewVnProvincesTmpRepository(db.GetPostgresDBConnection()),
	}
}

/* Bootstrap the manual seed data onto the temporary database for safe keeping and cross reference */
func (s *ManualSeedDumperService) BootstrapManualSeedDataToDatabase() error {
	db.ExecuteSQLScript("./resources/manual_seeds/provinces_seed.sql")

	wardSeedRootFolder := "./resources/manual_seeds/wards"
	err := filepath.WalkDir(wardSeedRootFolder, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".sql" {
			fmt.Printf("Executing script: %s\n", path)
			db.ExecuteSQLScript(path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking through manual decree seeds: %v\n", err)
	}
	return nil
}

func (s *ManualSeedDumperService) DumpToVNProvinceFromManualSeed() {
	fmt.Println("Dumping data from manual database degree seed...")

	// Thing to do: Manual inject data to the database - Done
	// Read from the insert data, construct the dvhcvn data
	seedProvinces := s.seedRepo.GetAllSeedProvinces()
	seedWards := s.seedRepo.GetAllSeedWards()
	s.insertToProvinces(seedProvinces)
	s.insertToWards(seedWards)
	fmt.Println("📥 Dumper operation finished")
}

func (s *ManualSeedDumperService) insertToWards(seedWardModels []model.SeedWard) {
	ctx := context.Background()
	totalWard := 0

	for _, w := range seedWardModels {
		wardFullName := helper.RemoveWhiteSpaces(w.Name)
		administrativeUnitLevel := helper.GetAdministrativeUnit_WardLevel(wardFullName)
		unitName := cfg.AdministrativeUnitNamesShortNameMap_vn[administrativeUnitLevel]
		unitName_en := cfg.AdministrativeUnitNamesShortNameMap_en[administrativeUnitLevel]
		wardShortName := strings.Trim(strings.Replace(wardFullName, unitName, "", 1), " ")
		codeName := helper.ToCodeName(wardShortName)
		wardShortNameEn := helper.NormalizeString(wardShortName)

		// Case when ward name is a number
		isNumber, _ := regexp.MatchString(`^[0-9]+$`, wardShortName)
		var wardFullNameEn string
		if isNumber {
			wardFullNameEn = unitName_en + " " + wardShortNameEn
		} else {
			wardFullNameEn = wardShortNameEn + " " + unitName_en
		}

		wardModel := &vn_provinces_tmp_model.Ward{
			Code:                 w.Code,
			Name:                 wardShortName,
			NameEn:               wardShortNameEn,
			FullName:             wardFullName,
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

func (s *ManualSeedDumperService) insertToProvinces(seedProvinceModels []model.SeedProvince) {
	ctx := context.Background()

	for _, p := range seedProvinceModels {
		provinceFullName := helper.RemoveWhiteSpaces(p.Name)
		administrativeUnitLevel := helper.GetAdministrativeUnit_ProvinceLevel(provinceFullName)
		unitName := cfg.AdministrativeUnitNamesShortNameMap_vn[administrativeUnitLevel]
		unitName_en := cfg.AdministrativeUnitNamesShortNameMap_en[administrativeUnitLevel]
		provinceShortName := strings.Trim(strings.Replace(provinceFullName, unitName, "", 1), " ")
		codeName := helper.ToCodeName(provinceShortName)
		provinceShortNameEn := helper.NormalizeString(provinceShortName)
		provinceFullNameEn := provinceShortNameEn + " " + unitName_en

		provinceModel := &vn_provinces_tmp_model.Province{
			Code:                 p.Code,
			Name:                 provinceShortName,
			NameEn:               provinceShortNameEn,
			FullName:             provinceFullName,
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

	fmt.Printf("Inserted %d provinces to tables\n", len(seedProvinceModels))
}

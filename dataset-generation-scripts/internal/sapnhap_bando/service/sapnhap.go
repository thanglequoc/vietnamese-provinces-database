package service

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

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

			wardData.TenHC = normalizeVietnameseToneSimple(wardData.TenHC)

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

func normalizeVietnameseToneSimple(text string) string {
	text = norm.NFC.String(text)
	text = strings.Replace(text, "’", "'", 1)

	if strings.Contains(text, "oài") {
		// Replace "oà" with "òa" only if it is not at the end of the string
		return text
	} else if strings.Contains(text, "oà ") || regexp.MustCompile(`oà$`).MatchString(text) {
		// Replace "oà" with "òa" at the end of the string
		return strings.Replace(text, "oà", "òa", 1)
	} else if strings.Contains(text, "oá ") || regexp.MustCompile(`oá$`).MatchString(text) {
		// Replace "oà" with "òa" at the end of the string
		return strings.Replace(text, "oá", "óa", 1)
	}
	return text
}

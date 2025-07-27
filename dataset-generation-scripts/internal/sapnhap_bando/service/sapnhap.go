package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/fetcher"
	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/repository"
	vnRepo "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/repository"
)

type SapNhapService struct {
	sapNhapRepo       *repository.SapNhapRepository
	vnProvinceTmpRepo *vnRepo.VnProvincesTmpRepository
}

func NewSapNhapService(repo *repository.SapNhapRepository, vnRepo *vnRepo.VnProvincesTmpRepository) *SapNhapService {
	return &SapNhapService{
		sapNhapRepo:       repo,
		vnProvinceTmpRepo: vnRepo,
	}
}

func (s *SapNhapService) BootstrapSapNhapSiteProvinces() error {
	ctx := context.TODO()
	sapNhapSiteProvinces := fetcher.GetAllProvincesDataFromSapNhapSite()

	// Insert each province into the repository
	for _, provinceData := range sapNhapSiteProvinces {
		tenTinh := strings.ToLower(provinceData.TenTinh)
		if strings.HasPrefix(tenTinh, "tỉnh ") {
			// Remove the "Tỉnh " prefix from the province name
			provinceData.TenTinh = strings.TrimPrefix(tenTinh, "tỉnh ")
		} else if strings.HasPrefix(tenTinh, "thành phố ") {
			// Remove the "Thành phố " prefix from the province name
			provinceData.TenTinh = strings.TrimPrefix(tenTinh, "thành phố")
		} else if strings.HasPrefix(tenTinh, "thủ đô ") {
			provinceData.TenTinh = strings.TrimPrefix(tenTinh, "thủ đô ")
		}

		// Attempt to look up the vn province by name
		vnProvince, err := s.vnProvinceTmpRepo.FindProvinceByName(ctx, strings.TrimSpace(provinceData.TenTinh))
		if err != nil {
			return err
		}
		if vnProvince == nil {
			return fmt.Errorf("VN Province not found for name: %s", provinceData.TenTinh)
		}

		province := &model.SapNhapSiteProvince{
			ID:             provinceData.ID,
			MaHC:           provinceData.MaHC,
			TenTinh:        provinceData.TenTinh,
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
			fmt.Errorf("failed to insert province %s: %w", provinceData.TenTinh, err)
			return err
		}
	}

	log.Default().Println("Bootstrap SapNhapSiteProvinces completed successfully")

	return nil
}

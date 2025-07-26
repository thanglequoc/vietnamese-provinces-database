package service

import (
	"fmt"
	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/fetcher"
	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/repository"
)

type SapNhapService struct {
	Repository *repository.SapNhapRepository
}

func NewSapNhapService(repo *repository.SapNhapRepository) *SapNhapService {
	return &SapNhapService{
		Repository: repo,
	}
}

func (s *SapNhapService) BootstrapSapNhapSiteProvinces() error {
	sapNhapSiteProvinces := fetcher.GetAllProvincesDataFromSapNhapSite()

	

	// Insert each province into the repository
	for _, provinceData := range sapNhapSiteProvinces {
		province := &model.SapNhapSiteProvince{
			MaHC:           provinceData.MaHC,
			TenTinh:        provinceData.TenTinh,
			DienTichKm2:    provinceData.DienTichKm2,
			DanSoNguoi:     provinceData.DanSoNguoi,
			TrungTamHC:     provinceData.TrungTamHC,
			KinhDo:         provinceData.KinhDo,
			ViDo:           provinceData.ViDo,
			TruocSN:        provinceData.TruocSN,
			Con:            provinceData.Con,
			VNProvinceCode: provinceData.VNProvinceCode,
		}
		if err := s.Repository.InsertSapNhapSiteProvince(province); err != nil {
			fmt.Errorf("failed to insert province %s: %w", provinceData.TenTinh, err)
			return err
		}
	}

	return nil
}

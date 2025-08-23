package sapnhapbando

import (
	"log"
	db "github.com/thanglequoc-vn-provinces/v2/internal/database"

	sapNhapService "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/service"
	sapNhapRepo "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/repository"
	vnRepo "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/repository"
)

func DumpDataFromSapNhapBando() {
	// Initialize the SapNhapRepository
	sapNhapRepo := sapNhapRepo.NewSapNhapRepository(db.GetPostgresDBConnection())
	vnRepo := vnRepo.NewVnProvincesTmpRepository(db.GetPostgresDBConnection())

	sapNhapService := sapNhapService.NewSapNhapService(sapNhapRepo, vnRepo)

	if err := sapNhapService.BootstrapSapNhapSiteProvinces(); err != nil {
		log.Fatalf("Failed to dump SapNhapSiteProvinces: %v", err)
		panic(err)
	}

	if err := sapNhapService.BootstrapSapNhapSiteWards(); err != nil {
		log.Fatalf("Failed to dump SapNhapSiteWards: %v", err)
		panic(err)
	}

	if err := sapNhapService.BootstrapGISData(); err != nil {
		log.Fatalf("Failed to bootstrap GIS Data: %v", err)
		panic(err)
	}

	log.Println("🗺️ Data dump from SapNhap Bando completed successfully")
}

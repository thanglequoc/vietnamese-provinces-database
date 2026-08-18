package postal_code

import (
	"context"
	"log"

	db "github.com/thanglequoc-vn-provinces/v2/internal/database"
	"github.com/thanglequoc-vn-provinces/v2/internal/postal_code/repository"
	"github.com/thanglequoc-vn-provinces/v2/internal/postal_code/service"
	vnRepo "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/repository"
)

const seedDir = "./resources/postal"

// ImportPostalCodes reads the postal code seed files and populates
// postal_code_prefix / postal_code on provinces_tmp and wards_tmp.
// Must run after BeginDumpingDataWithDvhcvnDirectSource() and before
// ReadAndGenerateSQLDatasets().
func ImportPostalCodes() {
	postgresDB := db.GetPostgresDBConnection()
	vnTmpRepo := vnRepo.NewVnProvincesTmpRepository(postgresDB)
	postalRepo := repository.NewPostalCodeRepository(postgresDB)
	postalService := service.NewPostalCodeService(vnTmpRepo, postalRepo, seedDir)

	ctx := context.Background()
	if err := postalService.ImportPostalCodes(ctx); err != nil {
		log.Fatalf("❌ Failed to import postal codes: %v", err)
		panic(err)
	}
	log.Println("✅ Postal codes imported successfully")
}

package dumper

import (
	"fmt"

	"github.com/thanglequoc-vn-provinces/v2/internal/dumper/service"
)

func DumpFromManualSeed() {
	manualSeedDumperSvc := service.NewManualSeedDumperService()
	manualSeedDumperSvc.BootstrapManualSeedDataToDatabase()
	manualSeedDumperSvc.DumpToVNProvinceFromManualSeed()
	fmt.Println("📥 Dumper operation finished")
}

func BeginDumpingDataWithDvhcvnDirectSource() {
	soapSeedDumperSvc := service.NewDvhcvnSoapSeedDumperService()
	soapSeedDumperSvc.BeginDumpingDataWithDvhcvnDirectSource()
}

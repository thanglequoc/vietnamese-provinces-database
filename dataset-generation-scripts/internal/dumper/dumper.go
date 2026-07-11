package dumper

import (
	"github.com/thanglequoc-vn-provinces/v2/internal/dumper/service"
)

func BeginDumpingDataWithDvhcvnDirectSource() {
	soapSeedDumperSvc := service.NewDvhcvnSoapSeedDumperService()
	soapSeedDumperSvc.BeginDumpingDataWithDvhcvnDirectSource()
}

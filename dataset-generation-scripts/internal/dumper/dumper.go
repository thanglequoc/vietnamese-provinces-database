package dumper

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
	"github.com/thanglequoc-vn-provinces/v2/internal/dumper/service"
	data_downloader "github.com/thanglequoc-vn-provinces/v2/internal/dvhcvn_data_downloader"
)

// Temporary deprecated, API upstream data is not up to date
func BeginDumpingDataWithDvhcvnDirectSource() {
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

	fmt.Println(dvhcvnUnits)

	manualSeedDumperSvc := service.NewManualSeedDumperService()
	manualSeedDumperSvc.BootstrapManualSeedDataToDatabase()
	manualSeedDumperSvc.DumpToVNProvinceFromManualSeed()
	fmt.Println("📥 Dumper operation finished")
}

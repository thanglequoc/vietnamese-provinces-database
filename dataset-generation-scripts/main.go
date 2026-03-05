package main

import (
	db "github.com/thanglequoc-vn-provinces/v2/internal/database"
	dataset_writer "github.com/thanglequoc-vn-provinces/v2/internal/dataset_writer"
	dumper "github.com/thanglequoc-vn-provinces/v2/internal/dumper"
	sapNhap "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando"
)

const INCLUDE_GIS = true
const USE_DIRECT_DVHCVN_SOURCE = true

func main() {
	// pre-run
	// Refresh temporary dataset, import existing dataset
	db.BootstrapTemporaryDatasetStructure()

	if (USE_DIRECT_DVHCVN_SOURCE) {
		dumper.BeginDumpingDataWithDvhcvnDirectSource()
	} else {
		dumper.DumpFromManualSeed()
	}
	dataset_writer.ReadAndGenerateSQLDatasets()

	if (INCLUDE_GIS) {
		db.BootstrapGISDataStructure()
		sapNhap.DumpDataFromSapNhapBando()
		dataset_writer.GenerateGISSQLDatasets()
	}
}

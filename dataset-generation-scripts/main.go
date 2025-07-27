package main

import (
	db "github.com/thanglequoc-vn-provinces/v2/internal/database"
	dataset_writer "github.com/thanglequoc-vn-provinces/v2/internal/dataset_writer"
	dumper "github.com/thanglequoc-vn-provinces/v2/internal/dumper"
	sapNhap "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando"
)

func main() {
	// pre-run
	// Refresh temporary dataset, import existing dataset
	db.BootstrapTemporaryDatasetStructure()
	dumper.DumpFromManualSeed()
	// dumper.BeginDumpingDataWithDvhcvnDirectSource()
	dataset_writer.ReadAndGenerateSQLDatasets()

	runSapNhap := true
	if (runSapNhap) {
		db.BootstrapGISDataStructure()
		sapNhap.DumpDataFromSapNhapBando()
	}
}

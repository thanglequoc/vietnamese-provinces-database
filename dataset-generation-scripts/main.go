package main

import (
	vn_common "github.com/thanglequoc-vn-provinces/v2/internal/common"
	dataset_writer "github.com/thanglequoc-vn-provinces/v2/internal/dataset_writer"
	dumper "github.com/thanglequoc-vn-provinces/v2/internal/dumper"
)

func main() {
	// pre-run
	// Refresh temporary dataset, import existing dataset
	vn_common.BootstrapTemporaryDatasetStructure()
	dumper.DumpFromManualSeed()
	// dumper.BeginDumpingDataWithDvhcvnDirectSource()
	dataset_writer.ReadAndGenerateSQLDatasets()
}

package main

import (
	vn_common "github.com/thanglequoc-vn-provinces/v2/common"
	// dataset_writer "github.com/thanglequoc-vn-provinces/v2/dataset_writer"
	// dumper "github.com/thanglequoc-vn-provinces/v2/dumper"
)

func main() {
	// pre-run
	// Refresh temporary dataset, import existing dataset
	vn_common.BootstrapTemporaryDatasetStructure()
	// dumper.BeginDumpingDataWithDvhcvnDirectSource()
	// dataset_writer.ReadAndGenerateSQLDatasets()
}

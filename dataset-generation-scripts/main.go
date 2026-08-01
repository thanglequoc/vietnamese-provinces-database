package main

import (
	"io"
	"os"
	"path/filepath"

	db "github.com/thanglequoc-vn-provinces/v2/internal/database"
	dataset_writer "github.com/thanglequoc-vn-provinces/v2/internal/dataset_writer"
	dumper "github.com/thanglequoc-vn-provinces/v2/internal/dumper"
	sapnhap "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando"
)

const INCLUDE_GIS = true

func main() {
	// Capture all console output (stdout + stderr) into output/generation-log.txt
	_ = os.MkdirAll("./output", 0746)
	logPath := filepath.Join("./output", "generation-log.txt")
	logFile, err := os.Create(logPath)
	if err == nil {
		// Tee stdout to both the original stdout and the log file
		origStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		go func() {
			_, _ = io.Copy(io.MultiWriter(origStdout, logFile), r)
		}()

		defer func() {
			w.Close()
			logFile.Close()
		}()
	}

	// pre-run
	// Refresh temporary dataset, import existing dataset
	db.BootstrapTemporaryDatasetStructure()

	dumper.BeginDumpingDataWithDvhcvnDirectSource()
	dataset_writer.ReadAndGenerateSQLDatasets()

	if INCLUDE_GIS {
		db.BootstrapGISDataStructure()
		sapnhap.BackfillProvinceAndWardCodesInSapNhapGeojsonObjects()
		sapnhap.FetchGISDataFromSapNhapBando()
		sapnhap.PatchIslandProvincesGeometry()
		sapnhap.ValidateAndFixGeometries()
		dataset_writer.GenerateGISSQLDatasets()
	}
}
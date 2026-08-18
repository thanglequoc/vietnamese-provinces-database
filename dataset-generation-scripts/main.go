package main

import (
	"io"
	"log"
	"os"
	"path/filepath"

	db "github.com/thanglequoc-vn-provinces/v2/internal/database"
	dataset_writer "github.com/thanglequoc-vn-provinces/v2/internal/dataset_writer"
	dumper "github.com/thanglequoc-vn-provinces/v2/internal/dumper"
	postal_code "github.com/thanglequoc-vn-provinces/v2/internal/postal_code"
	sapnhap "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando"
)

const INCLUDE_GIS = true

func main() {
	// Capture all console output (stdout + stderr) into output/generation-log.txt
	_ = os.MkdirAll("./output", 0746)
	logPath := filepath.Join("./output", "generation-log.txt")
	logFile, err := os.Create(logPath)
	if err == nil {
		// Tee stdout and stderr to both the original console and the log file.
		// Go's log package writes to stderr by default, so both streams must be
		// routed through the pipe to ensure the log file captures everything.
		origStdout := os.Stdout
		origStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stdout = w
		os.Stderr = w
		log.SetOutput(w)

		go func() {
			_, _ = io.Copy(io.MultiWriter(origStdout, origStderr, logFile), r)
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
	postal_code.ImportPostalCodes()
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
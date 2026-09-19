package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	datasetmetadata "github.com/thanglequoc-vn-provinces/v2/internal/dataset_metadata"
	"github.com/thanglequoc-vn-provinces/v2/internal/mock_gis_server"
	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/capture"
)

/*
mockgis serves the locally captured GIS responses so the dataset generation
pipeline can run without contacting the government server.

Usage (from dataset-generation-scripts/):

	go run ./cmd/mockgis --data ./resources/gis/gis_server_cache --addr 127.0.0.1:18080

Then, in another terminal:

	GIS_SERVER_BASE_URL=http://127.0.0.1:18080 go run main.go
*/
func main() {
	addr := flag.String("addr", "127.0.0.1:18080", "listen address")
	dataDir := flag.String("data", "./resources/gis/gis_server_cache", "capture cache directory")
	verbose := flag.Bool("verbose", false, "log every request")
	flag.Parse()

	cache, err := mock_gis_server.LoadCache(*dataDir)
	if err != nil {
		log.Fatalf("Failed to load GIS cache from %s: %v", *dataDir, err)
	}

	preadCount, metadataCount := cache.Counts()
	if preadCount == 0 || metadataCount == 0 {
		log.Fatalf(
			"Cache is incomplete (%d pread_json, %d co_dvhc_id). Run `go run ./cmd/giscapture` first.",
			preadCount, metadataCount,
		)
	}
	log.Printf("Loaded GIS cache from %s: %d pread_json, %d co_dvhc_id responses", cache.Dir(), preadCount, metadataCount)
	warnIfStale(filepath.Join(*dataDir, "manifest.json"))

	server := &http.Server{
		Addr:              *addr,
		Handler:           mock_gis_server.NewServer(cache, *verbose).Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("Mock GIS server listening on http://%s", *addr)
	log.Printf("Run the generator with: GIS_SERVER_BASE_URL=http://%s go run main.go", *addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Mock GIS server stopped: %v", err)
	}
}

// warnIfStale logs a hint when the cache was captured for a different dataset
// version than the current version.txt.
func warnIfStale(manifestPath string) {
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		return
	}

	var manifest capture.Manifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return
	}

	metadata, err := datasetmetadata.Load()
	if err != nil {
		return
	}

	if manifest.DatasetVersion != "" && manifest.DatasetVersion != metadata.DatasetVersion {
		log.Printf(
			"Warning: cache captured for dataset %s but version.txt is %s — re-run cmd/giscapture to refresh",
			manifest.DatasetVersion, metadata.DatasetVersion,
		)
	}
}

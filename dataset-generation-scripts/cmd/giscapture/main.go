package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	db "github.com/thanglequoc-vn-provinces/v2/internal/database"
	datasetmetadata "github.com/thanglequoc-vn-provinces/v2/internal/dataset_metadata"
	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/capture"
	sapNhapR "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/repository"
)

/*
giscapture dumps the raw responses of the government GIS server into a local,
gzip-compressed cache so the dataset generation pipeline can run offline against
the mock server (cmd/mockgis).

Usage (from dataset-generation-scripts/):

	# Capture both endpoints for every geo object (requires the PostGIS container)
	go run ./cmd/giscapture --force

	# Capture only the metadata endpoint from a JSON malk source
	go run ./cmd/giscapture --source json --only co_dvhc_id

After capture, run the generator against the mock:

	go run ./cmd/mockgis --data ./resources/gis/gis_server_cache
	GIS_SERVER_BASE_URL=http://127.0.0.1:18080 go run main.go
*/
func main() {
	outDir := flag.String("out", "./resources/gis/gis_server_cache", "output cache directory")
	baseURL := flag.String("base-url", "https://sapnhap.bando.com.vn", "GIS server to capture from (always explicit; never uses GIS_SERVER_BASE_URL)")
	workers := flag.Int("workers", 10, "number of concurrent capture workers")
	force := flag.Bool("force", false, "overwrite existing cache files")
	source := flag.String("source", "db", "malk source: 'db' (sapnhap_geojson_objects) or 'json'")
	jsonPath := flag.String("json", "./sapnhap-bando-crawler/donvi_tinhthanh.json", "JSON malk source used when --source=json")
	only := flag.String("only", "", "comma-separated endpoints to capture (pread_json, co_dvhc_id); empty = both")
	flag.Parse()

	malks, err := loadMalks(*source, *jsonPath)
	if err != nil {
		log.Fatalf("Failed to load malk list: %v", err)
	}
	log.Printf("Loaded %d unique malk values (source: %s)", len(malks), *source)

	metadata, err := datasetmetadata.Load()
	if err != nil {
		log.Printf("Warning: could not read version metadata: %v", err)
	}

	opts := capture.Options{
		BaseURL:        *baseURL,
		OutDir:         *outDir,
		Workers:        *workers,
		Force:          *force,
		Only:           splitCSV(*only),
		DatasetVersion: metadata.DatasetVersion,
		LatestDecree:   metadata.LatestDecree,
	}

	manifest, err := capture.Run(context.Background(), malks, opts)
	if manifest != nil {
		for endpoint, stats := range manifest.Endpoints {
			log.Printf("Endpoint %-12s captured %d/%d (missing %d)", endpoint, stats.Captured, stats.Expected, len(stats.Missing))
		}
	}
	if err != nil {
		log.Fatalf("Capture failed: %v", err)
	}

	log.Printf("Capture complete. Cache written to %s", *outDir)
}

// loadMalks resolves the malk list from the database or a JSON file.
func loadMalks(source, jsonPath string) ([]string, error) {
	switch source {
	case "json":
		return capture.LoadMalksFromJSON(jsonPath)
	case "db":
		postgresDB := db.GetPostgresDBConnection()
		repo := sapNhapR.NewSapNhapGeoJSONObjectRepository(postgresDB)
		objects, err := repo.GetAllSapNhapGeoJSONObjects(context.Background())
		if err != nil {
			return nil, fmt.Errorf("query sapnhap_geojson_objects: %w", err)
		}
		malks := make([]string, 0, len(objects))
		for _, obj := range objects {
			malks = append(malks, obj.MaLK)
		}
		return malks, nil
	default:
		return nil, fmt.Errorf("unknown source %q (expected 'db' or 'json')", source)
	}
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

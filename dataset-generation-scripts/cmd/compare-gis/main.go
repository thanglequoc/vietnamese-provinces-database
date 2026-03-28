package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/thanglequoc-vn-provinces/v2/internal/database"
	"github.com/thanglequoc-vn-provinces/v2/internal/gis_comparator/model"
	"github.com/thanglequoc-vn-provinces/v2/internal/gis_comparator/service"
)

func main() {
	// Define command-line flags
	compareFlag := flag.String("compare", "both", "What to compare: provinces, wards, or both (default: both)")
	provincesDumpFlag := flag.String("provinces-dump", "", "Path to provinces dump file")
	wardsDumpFlag := flag.String("wards-dump", "", "Path to wards dump file")
	verboseFlag := flag.Bool("v", false, "Enable verbose logging")
	_ = flag.Bool("cleanup", true, "Cleanup temporary tables after comparison (no-op - cleanup happens automatically)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Compares GIS data between current database and SQL dump files.")
		fmt.Fprintln(os.Stderr, "\nOptions:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nExamples:")
		fmt.Fprintf(os.Stderr, "  %s --compare=both\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --compare=provinces --provinces-dump=./path/to/dump.sql\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --provinces-dump=./resources/gis/exported/sapnhap_provinces_gis_202509011857_lfs.sql \\\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "     --wards-dump=./resources/gis/exported/sapnhap_wards_gis_202509011858_lfs.sql\n")
	}

	flag.Parse()

	// Set default dump file paths if not provided
	if *provincesDumpFlag == "" {
		*provincesDumpFlag = filepath.Join(".", "resources", "gis", "exported", "sapnhap_provinces_gis_202509011857_lfs.sql")
	}
	if *wardsDumpFlag == "" {
		*wardsDumpFlag = filepath.Join(".", "resources", "gis", "exported", "sapnhap_wards_gis_202509011858_lfs.sql")
	}

	if !*verboseFlag {
		log.SetFlags(0) // Remove timestamp from logs for cleaner output
	}

	log.Println("🔍 Starting GIS data comparison...")

	// Connect to database
	db := database.GetPostgresDBConnection()
	defer db.Close()

	comparatorService := service.NewGISComparatorService(db)

	// Run all operations in a single transaction to ensure temp tables are visible
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("❌ Failed to begin transaction: %v", err)
	}
	defer tx.Rollback() // Will rollback if not committed

	// Setup temporary tables and load dump data within the transaction
	log.Println("📋 Setting up temporary tables...")
	err = comparatorService.SetupTempTablesWithTx(tx, *provincesDumpFlag, *wardsDumpFlag)
	if err != nil {
		log.Fatalf("❌ Failed to setup temporary tables: %v", err)
	}

	// Perform comparisons based on flag
	var provincesSummary, wardsSummary *model.ComparisonSummary

	if *compareFlag == "provinces" || *compareFlag == "both" {
		log.Println("📍 Comparing provinces GIS data...")
		provincesSummary, err = comparatorService.CompareProvincesGISWithTx(tx)
		if err != nil {
			log.Fatalf("❌ Failed to compare provinces: %v", err)
		}
		service.LogComparisonResults(provincesSummary)
	}

	if *compareFlag == "wards" || *compareFlag == "both" {
		log.Println("🏘️  Comparing wards GIS data...")
		wardsSummary, err = comparatorService.CompareWardsGISWithTx(tx)
		if err != nil {
			log.Fatalf("❌ Failed to compare wards: %v", err)
		}
		service.LogComparisonResults(wardsSummary)
	}

	// Print summary
	service.PrintSummary(provincesSummary, wardsSummary)

	// Commit the transaction (this also removes temp tables automatically)
	err = tx.Commit()
	if err != nil {
		log.Printf("⚠️  Warning: Failed to commit transaction: %v", err)
	}

	// Exit with error code if differences found
	exitCode := 0
	if (*compareFlag == "provinces" || *compareFlag == "both") && provincesSummary != nil && !provincesSummary.IsEqual() {
		exitCode = 1
	}
	if (*compareFlag == "wards" || *compareFlag == "both") && wardsSummary != nil && !wardsSummary.IsEqual() {
		exitCode = 1
	}

	if exitCode != 0 {
		log.Println("\n⚠️  Comparison completed with differences")
		os.Exit(exitCode)
	}

	log.Println("\n✅ Comparison completed successfully - all data matches!")
}

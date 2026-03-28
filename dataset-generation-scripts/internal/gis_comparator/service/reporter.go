package service

import (
	"fmt"
	"log"

	"github.com/thanglequoc-vn-provinces/v2/internal/gis_comparator/model"
)

// LogComparisonResults prints a detailed comparison report
func LogComparisonResults(summary *model.ComparisonSummary) {
	log.Printf("==========================================")
	log.Printf("GIS DATA COMPARISON REPORT: %s", summary.TableName)
	log.Printf("==========================================")
	log.Printf("Total Records: %d", summary.TotalRecords)
	log.Printf("Matched: %d ✓", summary.MatchedRecords)
	log.Printf("Mismatched: %d ✗", summary.MismatchedRecords)
	log.Printf("Missing in Database: %d", len(summary.MissingInDB))
	log.Printf("Missing in Dump: %d", len(summary.MissingInDump))

	if len(summary.MissingInDB) > 0 {
		log.Printf("\nMissing in Database (%d):", len(summary.MissingInDB))
		for _, id := range summary.MissingInDB {
			log.Printf("  - %s", id)
		}
	}

	if len(summary.MissingInDump) > 0 {
		log.Printf("\nMissing in Dump (%d):", len(summary.MissingInDump))
		// Only show first 20 to avoid spamming
		limit := 20
		if len(summary.MissingInDump) < limit {
			limit = len(summary.MissingInDump)
		}
		for i := 0; i < limit; i++ {
			log.Printf("  - %s", summary.MissingInDump[i])
		}
		if len(summary.MissingInDump) > limit {
			log.Printf("  ... and %d more", len(summary.MissingInDump)-limit)
		}
	}

	if len(summary.Differences) > 0 {
		log.Printf("\nDetailed Differences:")
		for i, diff := range summary.Differences {
			log.Printf("\n%d. %s [%s]", i+1, diff.RecordName, diff.RecordIdentifier)
			log.Printf("   Field: %s", diff.FieldName)
			log.Printf("   Expected: %s", truncateWKT(diff.ExpectedWKT, 100))
			log.Printf("   Actual:   %s", truncateWKT(diff.ActualWKT, 100))
		}
	}

	log.Printf("==========================================\n")
}

// PrintSummary prints a brief summary
func PrintSummary(provincesSummary, wardsSummary *model.ComparisonSummary) {
	fmt.Println("\n╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║           GIS DATA COMPARISON SUMMARY                        ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")

	if provincesSummary != nil {
		fmt.Printf("\n📍 Provinces (sapnhap_provinces_gis):\n")
		fmt.Printf("   Total: %d | Matched: %d ✓ | Mismatched: %d", provincesSummary.TotalRecords, provincesSummary.MatchedRecords, provincesSummary.MismatchedRecords)
		if len(provincesSummary.MissingInDB) > 0 {
			fmt.Printf(" | Missing in DB: %d", len(provincesSummary.MissingInDB))
		}
		if len(provincesSummary.MissingInDump) > 0 {
			fmt.Printf(" | Missing in Dump: %d", len(provincesSummary.MissingInDump))
		}
		fmt.Println()

		if provincesSummary.IsEqual() {
			fmt.Println("   ✅ All provinces match!")
		} else {
			fmt.Println("   ⚠️  Differences found!")
		}
	}

	if wardsSummary != nil {
		fmt.Printf("\n🏘️  Wards (sapnhap_wards_gis):\n")
		fmt.Printf("   Total: %d | Matched: %d ✓ | Mismatched: %d", wardsSummary.TotalRecords, wardsSummary.MatchedRecords, wardsSummary.MismatchedRecords)
		if len(wardsSummary.MissingInDB) > 0 {
			fmt.Printf(" | Missing in DB: %d", len(wardsSummary.MissingInDB))
		}
		if len(wardsSummary.MissingInDump) > 0 {
			fmt.Printf(" | Missing in Dump: %d", len(wardsSummary.MissingInDump))
		}
		fmt.Println()

		if wardsSummary.IsEqual() {
			fmt.Println("   ✅ All wards match!")
		} else {
			fmt.Println("   ⚠️  Differences found!")
		}
	}

	fmt.Println("\n╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                     END OF SUMMARY                            ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
}

// truncateWKT truncates a WKT string to a specified length for display
func truncateWKT(wkt string, maxLen int) string {
	if len(wkt) <= maxLen {
		return wkt
	}
	return wkt[:maxLen] + "..."
}

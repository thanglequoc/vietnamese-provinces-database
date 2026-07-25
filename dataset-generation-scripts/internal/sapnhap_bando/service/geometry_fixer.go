package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// FixedWardRecord represents a single ward geometry that was corrected
type FixedWardRecord struct {
	Ma           string `bun:"ma"`
	Ten          string `bun:"ten"`
	ProvinceCode string `bun:"vn_ds_province_code"`
	WardCode     string `bun:"vn_ds_ward_code"`
}

// ValidateAndFixGeometries finds and fixes all invalid (self-intersecting) ward
// geometries in sapnhap_geojson_objects. It updates geom_wkt (WKT text), and
// the computed geom column follows automatically.
//
// Fix chain: ST_CollectionExtract(ST_MakeValid(ST_GeomFromText(geom_wkt, 4326)), 3)
//
// An audit log is written to output/gis_geometry_fix_log_<timestamp>.txt with
// the codes and names of every fixed ward for cross-verification.
func (s *SapNhapService) ValidateAndFixGeometries(ctx context.Context) error {
	// Step 1: Count total records for the audit log
	totalCount, err := s.db.NewSelect().
		TableExpr("sapnhap_geojson_objects").
		Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to count total records: %w", err)
	}

	// Step 2: Count invalid records before fix
	invalidCount, err := s.db.NewSelect().
		TableExpr("sapnhap_geojson_objects").
		Where("NOT ST_IsValid(ST_GeomFromText(geom_wkt, 4326))").
		Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to count invalid records: %w", err)
	}

	log.Printf("ℹ️ GIS Geometry Check: %d total, %d invalid", totalCount, invalidCount)

	if invalidCount == 0 {
		log.Println("✅ All GIS geometries are valid — no fixes needed")
		return nil
	}

	// Step 3: Execute the fix UPDATE with RETURNING to get affected rows
	fixSQL := `UPDATE sapnhap_geojson_objects
SET geom_wkt = ST_AsText(
    ST_CollectionExtract(
        ST_MakeValid(ST_GeomFromText(geom_wkt, 4326)),
        3
    )
)
WHERE NOT ST_IsValid(ST_GeomFromText(geom_wkt, 4326))
RETURNING ma, ten, vn_ds_province_code, vn_ds_ward_code`

	var fixedWards []FixedWardRecord
	if err := s.db.NewRaw(fixSQL).Scan(ctx, &fixedWards); err != nil {
		return fmt.Errorf("failed to fix geometries: %w", err)
	}

	log.Printf("✅ Fixed %d ward geometries", len(fixedWards))

	// Step 4: Run verification query
	remainingInvalid, err := s.db.NewSelect().
		TableExpr("sapnhap_geojson_objects").
		Where("NOT ST_IsValid(ST_GeomFromText(geom_wkt, 4326))").
		Count(ctx)
	if err != nil {
		return fmt.Errorf("verification query failed: %w", err)
	}

	if remainingInvalid > 0 {
		return fmt.Errorf("verification failed: %d invalid geometries remain after fix", remainingInvalid)
	}
	log.Println("✅ Verification passed: 0 invalid geometries remain")

	// Step 5: Write audit log
	if err := writeAuditLog(totalCount, invalidCount, fixedWards); err != nil {
		log.Printf("⚠️ Failed to write audit log: %v", err)
		// Don't fail the pipeline — fix was already applied successfully
	}

	return nil
}

// writeAuditLog writes a detailed fix log to output/gis_geometry_fix_log_<timestamp>.txt
func writeAuditLog(totalCount, invalidCount int, fixedWards []FixedWardRecord) error {
	// Ensure output directory exists
	outputDir := "output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate timestamped filename
	timestamp := time.Now().Format("2006-01-02__15_04_05")
	filename := filepath.Join(outputDir, fmt.Sprintf("gis_geometry_fix_log_%s.txt", timestamp))

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create audit log file: %w", err)
	}
	defer f.Close()

	// Count unique provinces
	provinceSet := make(map[string]struct{})
	for _, w := range fixedWards {
		if w.ProvinceCode != "" {
			provinceSet[w.ProvinceCode] = struct{}{}
		}
	}

	// Write header
	fmt.Fprintf(f, "GIS Geometry Fix Audit Log\n")
	fmt.Fprintf(f, "Generated: %s ICT\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(f, "============================================================\n")
	fmt.Fprintf(f, "Total records in sapnhap_geojson_objects: %d\n", totalCount)
	fmt.Fprintf(f, "Records checked (invalid): %d\n", invalidCount)
	fmt.Fprintf(f, "Records fixed: %d\n", len(fixedWards))
	fmt.Fprintf(f, "Provinces affected: %d\n", len(provinceSet))
	fmt.Fprintf(f, "\n--- Fixed Wards ---\n")
	fmt.Fprintf(f, "%-8s  %-28s  %-10s  %-10s\n", "ma", "ten", "prov_code", "ward_code")
	fmt.Fprintf(f, "%-8s  %-28s  %-10s  %-10s\n", "--------", "----------------------------", "----------", "----------")

	for _, w := range fixedWards {
		fmt.Fprintf(f, "%-8s  %-28s  %-10s  %-10s\n", w.Ma, w.Ten, w.ProvinceCode, w.WardCode)
	}

	fmt.Fprintf(f, "\n--- Verification ---\n")
	fmt.Fprintf(f, "Remaining invalid geometries: 0\n")
	fmt.Fprintf(f, "Fix command: ST_CollectionExtract(ST_MakeValid(ST_GeomFromText(geom_wkt, 4326)), 3)\n")
	fmt.Fprintf(f, "Applied to: geom_wkt column (computed geom column follows automatically)\n")
	fmt.Fprintf(f, "Safety: Idempotent — re-running produces zero changes\n")

	log.Printf("📄 Audit log written to %s", filename)
	return nil
}

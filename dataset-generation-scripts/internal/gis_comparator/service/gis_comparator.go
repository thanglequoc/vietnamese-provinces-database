package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/thanglequoc-vn-provinces/v2/internal/gis_comparator/model"
	"github.com/uptrace/bun"
)

type GISComparatorService struct {
	db *bun.DB
}

func NewGISComparatorService(db *bun.DB) *GISComparatorService {
	return &GISComparatorService{
		db: db,
	}
}

// SetupTempTablesWithTx creates temporary tables and loads data using the provided transaction
func (s *GISComparatorService) SetupTempTablesWithTx(tx bun.Tx, provincesDumpPath, wardsDumpPath string) error {
	// Step 1: Read the setup SQL script
	setupSQLPath := filepath.Join(".", "internal", "gis_comparator", "sql", "setup_temp_tables.sql")
	setupSQL, err := os.ReadFile(setupSQLPath)
	if err != nil {
		return fmt.Errorf("failed to read setup SQL file: %w", err)
	}

	// Step 2: Execute the setup SQL to create temp tables
	_, err = tx.Exec(string(setupSQL))
	if err != nil {
		return fmt.Errorf("failed to create temp tables: %w", err)
	}
	log.Printf("Created temporary tables")

	// Step 3: Load provinces dump data
	if provincesDumpPath != "" {
		err = s.loadDumpFile(tx, provincesDumpPath, "temp_sapnhap_provinces_gis_dump", "public.sapnhap_provinces_gis")
		if err != nil {
			return fmt.Errorf("failed to load provinces dump: %w", err)
		}
		log.Printf("Loaded provinces dump data from %s", provincesDumpPath)
	}

	// Step 4: Load wards dump data
	if wardsDumpPath != "" {
		err = s.loadDumpFile(tx, wardsDumpPath, "temp_sapnhap_wards_gis_dump", "public.sapnhap_wards_gis")
		if err != nil {
			return fmt.Errorf("failed to load wards dump: %w", err)
		}
		log.Printf("Loaded wards dump data from %s", wardsDumpPath)
	}

	return nil
}

// loadDumpFile reads a SQL dump file, modifies table names, and executes it
func (s *GISComparatorService) loadDumpFile(tx bun.Tx, dumpPath, tempTableName, originalTableName string) error {
	// Read the entire file and modify INSERT statements
	content, err := os.ReadFile(dumpPath)
	if err != nil {
		return fmt.Errorf("failed to read dump file: %w", err)
	}

	// Replace the original table name with temp table name in INSERT statements
	modifiedSQL := strings.ReplaceAll(string(content), "INSERT INTO "+originalTableName, "INSERT INTO "+tempTableName)

	// Execute the modified SQL
	_, err = tx.Exec(modifiedSQL)
	if err != nil {
		return fmt.Errorf("failed to execute dump SQL: %w", err)
	}

	return nil
}

// CompareProvincesGISWithTx compares current provinces_gis with dump data using the provided transaction
func (s *GISComparatorService) CompareProvincesGISWithTx(tx bun.Tx) (*model.ComparisonSummary, error) {
	summary := &model.ComparisonSummary{
		TableName: "sapnhap_provinces_gis",
	}


	// Get total counts
	var totalCount, dumpCount int
	err := tx.NewRaw("SELECT COUNT(*) FROM sapnhap_provinces_gis").Scan(context.Background(), &totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get current provinces count: %w", err)
	}

	err = tx.NewRaw("SELECT COUNT(*) FROM temp_sapnhap_provinces_gis_dump").Scan(context.Background(), &dumpCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get dump provinces count: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get dump provinces count: %w", err)
	}

	// Use the larger count as total
	if totalCount > dumpCount {
		summary.TotalRecords = totalCount
	} else {
		summary.TotalRecords = dumpCount
	}

	// Find missing in current DB
	err = tx.NewRaw(`
		SELECT gis_server_id
		FROM temp_sapnhap_provinces_gis_dump
		WHERE gis_server_id NOT IN (SELECT gis_server_id FROM sapnhap_provinces_gis)
		ORDER BY gis_server_id
	`).Scan(context.Background(), &summary.MissingInDB)
	if err != nil {
		return nil, fmt.Errorf("failed to find missing records in DB: %w", err)
	}

	// Find missing in dump
	err = tx.NewRaw(`
		SELECT gis_server_id
		FROM sapnhap_provinces_gis
		WHERE gis_server_id NOT IN (SELECT gis_server_id FROM temp_sapnhap_provinces_gis_dump)
		ORDER BY gis_server_id
	`).Scan(context.Background(), &summary.MissingInDump)
	if err != nil {
		return nil, fmt.Errorf("failed to find missing records in dump: %w", err)
	}

	// Find mismatches in bbox_wkt or geom_wkt
	type MismatchResult struct {
		GISServerID     string `bun:"gis_server_id"`
		Ten             string `bun:"ten"`
		CurrentBBoxWKT  string `bun:"current_bbox_wkt"`
		DumpBBoxWKT     string `bun:"dump_bbox_wkt"`
		CurrentGeomWKT  string `bun:"current_geom_wkt"`
		DumpGeomWKT     string `bun:"dump_geom_wkt"`
		ComparisonField string `bun:"comparison_field"`
	}

	var mismatches []MismatchResult
	err = tx.NewRaw(`
		WITH current AS (
			SELECT gis_server_id, ten, bbox_wkt, geom_wkt
			FROM sapnhap_provinces_gis
		),
		dump AS (
			SELECT gis_server_id, ten, bbox_wkt, geom_wkt
			FROM temp_sapnhap_provinces_gis_dump
		)
		SELECT
			c.gis_server_id,
			COALESCE(c.ten, d.ten) as ten,
			c.bbox_wkt as current_bbox_wkt,
			d.bbox_wkt as dump_bbox_wkt,
			c.geom_wkt as current_geom_wkt,
			d.geom_wkt as dump_geom_wkt,
			CASE
				WHEN replace(c.bbox_wkt, ' ', '') <> replace(d.bbox_wkt, ' ', '') 
				 AND replace(c.geom_wkt, ' ', '') <> replace(d.geom_wkt, ' ', '') THEN 'both'
				WHEN replace(c.bbox_wkt, ' ', '') <> replace(d.bbox_wkt, ' ', '') THEN 'bbox_wkt'
				WHEN replace(c.geom_wkt, ' ', '') <> replace(d.geom_wkt, ' ', '') THEN 'geom_wkt'
			END as comparison_field
		FROM current c
		INNER JOIN dump d ON c.gis_server_id = d.gis_server_id
		WHERE replace(c.bbox_wkt, ' ', '') <> replace(d.bbox_wkt, ' ', '') 
		   OR replace(c.geom_wkt, ' ', '') <> replace(d.geom_wkt, ' ', '')
		ORDER BY c.gis_server_id
	`).Scan(context.Background(), &mismatches)
	if err != nil {
		return nil, fmt.Errorf("failed to find mismatches: %w", err)
	}

	summary.MismatchedRecords = len(mismatches)
	summary.MatchedRecords = summary.TotalRecords - summary.MismatchedRecords - len(summary.MissingInDB) - len(summary.MissingInDump)

	// Build detailed differences
	for _, m := range mismatches {
		if m.ComparisonField == "both" || m.ComparisonField == "bbox_wkt" {
			summary.Differences = append(summary.Differences, model.WKTComparisonResult{
				RecordIdentifier: m.GISServerID,
				RecordName:       m.Ten,
				FieldName:        "bbox_wkt",
				ExpectedWKT:      m.DumpBBoxWKT,
				ActualWKT:        m.CurrentBBoxWKT,
				Difference:       "bbox_wkt values differ",
			})
		}
		if m.ComparisonField == "both" || m.ComparisonField == "geom_wkt" {
			summary.Differences = append(summary.Differences, model.WKTComparisonResult{
				RecordIdentifier: m.GISServerID,
				RecordName:       m.Ten,
				FieldName:        "geom_wkt",
				ExpectedWKT:      m.DumpGeomWKT,
				ActualWKT:        m.CurrentGeomWKT,
				Difference:       "geom_wkt values differ",
			})
		}
	}

	return summary, nil
}

// CompareWardsGISWithTx compares current wards_gis with dump data using the provided transaction
func (s *GISComparatorService) CompareWardsGISWithTx(tx bun.Tx) (*model.ComparisonSummary, error) {
	summary := &model.ComparisonSummary{
		TableName: "sapnhap_wards_gis",
	}

	// Get total counts
	var totalCount, dumpCount int
	err := tx.NewRaw("SELECT COUNT(*) FROM sapnhap_wards_gis").Scan(context.Background(), &totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get current wards count: %w", err)
	}
	err = tx.NewRaw("SELECT COUNT(*) FROM temp_sapnhap_wards_gis_dump").Scan(context.Background(), &dumpCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get dump wards count: %w", err)
	}

	// Use the larger count as total
	if totalCount > dumpCount {
		summary.TotalRecords = totalCount
	} else {
		summary.TotalRecords = dumpCount
	}

	// Find missing in current DB
	err = tx.NewRaw(`
		SELECT gis_server_id
		FROM temp_sapnhap_wards_gis_dump
		WHERE gis_server_id NOT IN (SELECT gis_server_id FROM sapnhap_wards_gis)
		ORDER BY gis_server_id
	`).Scan(context.Background(), &summary.MissingInDB)
	if err != nil {
		return nil, fmt.Errorf("failed to find missing records in DB: %w", err)
	}

	// Find missing in dump
	err = tx.NewRaw(`
		SELECT gis_server_id
		FROM sapnhap_wards_gis
		WHERE gis_server_id NOT IN (SELECT gis_server_id FROM temp_sapnhap_wards_gis_dump)
		ORDER BY gis_server_id
	`).Scan(context.Background(), &summary.MissingInDump)
	if err != nil {
		return nil, fmt.Errorf("failed to find missing records in dump: %w", err)
	}

	// Find mismatches in bbox_wkt or geom_wkt
	type MismatchResult struct {
		GISServerID     string `bun:"gis_server_id"`
		Ten             string `bun:"ten"`
		CurrentBBoxWKT  string `bun:"current_bbox_wkt"`
		DumpBBoxWKT     string `bun:"dump_bbox_wkt"`
		CurrentGeomWKT  string `bun:"current_geom_wkt"`
		DumpGeomWKT     string `bun:"dump_geom_wkt"`
		ComparisonField string `bun:"comparison_field"`
	}

	var mismatches []MismatchResult
	err = tx.NewRaw(`
		WITH current AS (
			SELECT gis_server_id, ten, bbox_wkt, geom_wkt
			FROM sapnhap_wards_gis
		),
		dump AS (
			SELECT gis_server_id, ten, bbox_wkt, geom_wkt
			FROM temp_sapnhap_wards_gis_dump
		)
		SELECT
			c.gis_server_id,
			COALESCE(c.ten, d.ten) as ten,
			c.bbox_wkt as current_bbox_wkt,
			d.bbox_wkt as dump_bbox_wkt,
			c.geom_wkt as current_geom_wkt,
			d.geom_wkt as dump_geom_wkt,
			CASE
				WHEN replace(c.bbox_wkt, ' ', '') <> replace(d.bbox_wkt, ' ', '') 
				 AND replace(c.geom_wkt, ' ', '') <> replace(d.geom_wkt, ' ', '') THEN 'both'
				WHEN replace(c.bbox_wkt, ' ', '') <> replace(d.bbox_wkt, ' ', '') THEN 'bbox_wkt'
				WHEN replace(c.geom_wkt, ' ', '') <> replace(d.geom_wkt, ' ', '') THEN 'geom_wkt'
			END as comparison_field
		FROM current c
		INNER JOIN dump d ON c.gis_server_id = d.gis_server_id
		WHERE replace(c.bbox_wkt, ' ', '') <> replace(d.bbox_wkt, ' ', '') 
		   OR replace(c.geom_wkt, ' ', '') <> replace(d.geom_wkt, ' ', '')
		ORDER BY c.gis_server_id
	`).Scan(context.Background(), &mismatches)
	if err != nil {
		return nil, fmt.Errorf("failed to find mismatches: %w", err)
	}

	summary.MismatchedRecords = len(mismatches)
	summary.MatchedRecords = summary.TotalRecords - summary.MismatchedRecords - len(summary.MissingInDB) - len(summary.MissingInDump)

	// Build detailed differences
	for _, m := range mismatches {
		if m.ComparisonField == "both" || m.ComparisonField == "bbox_wkt" {
			summary.Differences = append(summary.Differences, model.WKTComparisonResult{
				RecordIdentifier: m.GISServerID,
				RecordName:       m.Ten,
				FieldName:        "bbox_wkt",
				ExpectedWKT:      m.DumpBBoxWKT,
				ActualWKT:        m.CurrentBBoxWKT,
				Difference:       "bbox_wkt values differ",
			})
		}
		if m.ComparisonField == "both" || m.ComparisonField == "geom_wkt" {
			summary.Differences = append(summary.Differences, model.WKTComparisonResult{
				RecordIdentifier: m.GISServerID,
				RecordName:       m.Ten,
				FieldName:        "geom_wkt",
				ExpectedWKT:      m.DumpGeomWKT,
				ActualWKT:        m.CurrentGeomWKT,
				Difference:       "geom_wkt values differ",
			})
		}
	}

	return summary, nil
}

// GetDatabaseContainerID returns the Docker container ID for the database
func (s *GISComparatorService) GetDatabaseContainerID() (string, error) {
	cmd := exec.Command("docker", "ps", "--filter", "name=vn_provinces_postgres_container", "--format", "{{.ID}}")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get database container ID: %w", err)
	}
	containerID := strings.TrimSpace(string(output))
	if containerID == "" {
		return "", fmt.Errorf("database container not found")
	}
	return containerID, nil
}

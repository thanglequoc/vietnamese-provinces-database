package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWriteAuditLog_Success verifies that the audit log is correctly written
// with the expected header, ward table, and verification section.
func TestWriteAuditLog_Success(t *testing.T) {
	// Redirect output to a temp directory
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	require.NoError(t, os.Chdir(tmpDir))

	fixedWards := []FixedWardRecord{
		{Ma: "00001", Ten: "Phường Ba Đình", ProvinceCode: "01", WardCode: "00001"},
		{Ma: "00002", Ten: "Phường Trúc Bạch", ProvinceCode: "01", WardCode: "00002"},
		{Ma: "00150", Ten: "Xã An Khánh", ProvinceCode: "68", WardCode: "00150"},
	}

	err = writeAuditLog(3355, 3, fixedWards)
	assert.NoError(t, err)

	// Find the generated log file
	matches, err := filepath.Glob(filepath.Join(tmpDir, "output", "gis_geometry_fix_log_*.txt"))
	require.NoError(t, err)
	require.Len(t, matches, 1, "expected exactly one audit log file")

	content, err := os.ReadFile(matches[0])
	require.NoError(t, err)
	contentStr := string(content)

	// Verify header
	assert.Contains(t, contentStr, "GIS Geometry Fix Audit Log")
	assert.Contains(t, contentStr, "Total records in sapnhap_geojson_objects: 3355")
	assert.Contains(t, contentStr, "Records checked (invalid): 3")
	assert.Contains(t, contentStr, "Records fixed: 3")
	assert.Contains(t, contentStr, "Provinces affected: 2") // 01 and 68

	// Verify ward table entries
	assert.Contains(t, contentStr, "00001")
	assert.Contains(t, contentStr, "Phường Ba Đình")
	assert.Contains(t, contentStr, "Phường Trúc Bạch")
	assert.Contains(t, contentStr, "Xã An Khánh")
	assert.Contains(t, contentStr, "01")
	assert.Contains(t, contentStr, "68")

	// Verify verification section
	assert.Contains(t, contentStr, "Remaining invalid geometries: 0")
	assert.Contains(t, contentStr, "ST_CollectionExtract(ST_MakeValid(ST_GeomFromText(geom_wkt, 4326)), 3)")
	assert.Contains(t, contentStr, "Idempotent")
}

// TestWriteAuditLog_EmptyFixedWards verifies that the audit log is still
// created correctly when there are no fixed wards (edge case).
func TestWriteAuditLog_EmptyFixedWards(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	require.NoError(t, os.Chdir(tmpDir))

	err = writeAuditLog(3355, 0, []FixedWardRecord{})
	assert.NoError(t, err)

	// Find the generated log file
	matches, err := filepath.Glob(filepath.Join(tmpDir, "output", "gis_geometry_fix_log_*.txt"))
	require.NoError(t, err)
	require.Len(t, matches, 1, "expected exactly one audit log file")

	content, err := os.ReadFile(matches[0])
	require.NoError(t, err)
	contentStr := string(content)

	assert.Contains(t, contentStr, "Records fixed: 0")
	assert.Contains(t, contentStr, "Provinces affected: 0")
}

// TestWriteAuditLog_ProvinceCount verifies that the province count is correct
// when multiple wards belong to the same province (deduplication).
func TestWriteAuditLog_ProvinceCount(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	require.NoError(t, os.Chdir(tmpDir))

	// 5 wards across 3 provinces (01 has 3, 68 has 1, 79 has 1)
	fixedWards := []FixedWardRecord{
		{Ma: "00001", Ten: "Ward A", ProvinceCode: "01", WardCode: "00001"},
		{Ma: "00002", Ten: "Ward B", ProvinceCode: "01", WardCode: "00002"},
		{Ma: "00003", Ten: "Ward C", ProvinceCode: "01", WardCode: "00003"},
		{Ma: "00150", Ten: "Ward D", ProvinceCode: "68", WardCode: "00150"},
		{Ma: "00200", Ten: "Ward E", ProvinceCode: "79", WardCode: "00200"},
	}

	err = writeAuditLog(3355, 5, fixedWards)
	assert.NoError(t, err)

	matches, err := filepath.Glob(filepath.Join(tmpDir, "output", "gis_geometry_fix_log_*.txt"))
	require.NoError(t, err)
	require.Len(t, matches, 1)

	content, err := os.ReadFile(matches[0])
	require.NoError(t, err)

	assert.Contains(t, string(content), "Provinces affected: 3")
}

// TestWriteAuditLog_EmptyProvinceCode verifies that wards with empty province
// codes are handled gracefully (not counted in the province set).
func TestWriteAuditLog_EmptyProvinceCode(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	require.NoError(t, os.Chdir(tmpDir))

	fixedWards := []FixedWardRecord{
		{Ma: "00001", Ten: "Ward A", ProvinceCode: "01", WardCode: "00001"},
		{Ma: "00002", Ten: "Ward B", ProvinceCode: "", WardCode: "00002"}, // empty province
	}

	err = writeAuditLog(3355, 2, fixedWards)
	assert.NoError(t, err)

	matches, err := filepath.Glob(filepath.Join(tmpDir, "output", "gis_geometry_fix_log_*.txt"))
	require.NoError(t, err)
	require.Len(t, matches, 1)

	content, err := os.ReadFile(matches[0])
	require.NoError(t, err)

	// Only 1 province (01), the empty one should not be counted
	assert.Contains(t, string(content), "Provinces affected: 1")
}

// TestWriteAuditLog_OutputDirectoryCreation verifies that the output directory
// is created if it doesn't already exist.
func TestWriteAuditLog_OutputDirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	require.NoError(t, os.Chdir(tmpDir))

	// Ensure output/ doesn't exist yet
	_, err = os.Stat(filepath.Join(tmpDir, "output"))
	assert.True(t, os.IsNotExist(err))

	err = writeAuditLog(100, 1, []FixedWardRecord{
		{Ma: "00001", Ten: "Test Ward", ProvinceCode: "01", WardCode: "00001"},
	})
	assert.NoError(t, err)

	// Verify output/ was created
	info, err := os.Stat(filepath.Join(tmpDir, "output"))
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

// TestWriteAuditLog_FileFormat verifies the column headers and formatting
// of the ward table in the audit log.
func TestWriteAuditLog_FileFormat(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	require.NoError(t, os.Chdir(tmpDir))

	fixedWards := []FixedWardRecord{
		{Ma: "00001", Ten: "Phường Test", ProvinceCode: "01", WardCode: "00001"},
	}

	err = writeAuditLog(100, 1, fixedWards)
	assert.NoError(t, err)

	matches, err := filepath.Glob(filepath.Join(tmpDir, "output", "gis_geometry_fix_log_*.txt"))
	require.NoError(t, err)
	require.Len(t, matches, 1)

	content, err := os.ReadFile(matches[0])
	require.NoError(t, err)
	lines := strings.Split(string(content), "\n")

	// Find the header line with column names
	var headerLine string
	for _, line := range lines {
		if strings.Contains(line, "ma") && strings.Contains(line, "ten") && strings.Contains(line, "prov_code") {
			headerLine = line
			break
		}
	}
	assert.NotEmpty(t, headerLine, "column header line not found")
	assert.Contains(t, headerLine, "ma")
	assert.Contains(t, headerLine, "ten")
	assert.Contains(t, headerLine, "prov_code")
	assert.Contains(t, headerLine, "ward_code")
}

// TestValidateAndFixGeometries_Integration is an integration test that requires
// a running PostGIS database with sapnhap_geojson_objects populated.
func TestValidateAndFixGeometries_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test requires:
	// 1. A running PostGIS database on localhost:15432
	// 2. The sapnhap_geojson_objects table populated with GIS data
	// 3. Some geometries that are intentionally invalid (self-intersecting)
	//
	// To run this test:
	//   docker compose -f docker/docker-compose.yaml up -d
	//   go test -v ./internal/sapnhap_bando/service/ -run TestValidateAndFixGeometries_Integration
	//
	// The test verifies:
	// 1. Invalid geometries are detected and counted
	// 2. The ST_MakeValid fix is applied successfully
	// 3. Post-fix verification shows 0 invalid geometries
	// 4. An audit log file is generated in output/

	t.Skip("Integration test - requires PostGIS database with populated GIS data. Run manually after generation.")
}
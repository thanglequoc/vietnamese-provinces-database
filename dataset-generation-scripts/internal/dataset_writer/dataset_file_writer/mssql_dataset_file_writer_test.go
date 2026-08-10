package dataset_writer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	sapnhapmodels "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
)

func TestMssqlDatasetFileWriter_WriteGISDataToFile_Chunked(t *testing.T) {
	tmpDir := t.TempDir()
	originalWD, err := os.Getwd()
	assert.NoError(t, err)

	err = os.Chdir(tmpDir)
	assert.NoError(t, err)
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	writer := &MssqlDatasetFileWriter{}
	provinces := []*sapnhapmodels.SapNhapSiteGeoUnit{
		{
			VNDSProvinceCode: "01",
			MaLK:             "tinh.1",
			DienTichKM2:      3359.84,
			BBoxWKT:          "POLYGON((105 20,106 20,106 21,105 21,105 20))",
			GeomWKT:          "MULTIPOLYGON(((105 20,106 20,106 21,105 21,105 20)))",
		},
	}
	wards := []*sapnhapmodels.SapNhapSiteGeoUnit{
		{
			VNDSWardCode: "00001",
			MaLK:         "xa.1",
			DienTichKM2:  5.23,
			BBoxWKT:      "POLYGON((105.8 21,105.9 21,105.9 21.1,105.8 21.1,105.8 21))",
			GeomWKT:      "POLYGON((105.8 21,105.9 21,105.9 21.1,105.8 21.1,105.8 21))",
		},
		{
			VNDSWardCode: "00002",
			MaLK:         "xa.2",
			DienTichKM2:  3.14,
			BBoxWKT:      "POLYGON((106 21,106.1 21,106.1 21.1,106 21.1,106 21))",
			GeomWKT:      "POLYGON((106 21,106.1 21,106.1 21.1,106 21.1,106 21))",
		},
	}

	err = writer.WriteGISDataToFile(provinces, wards)
	assert.NoError(t, err)

	content := readGeneratedMssqlGISFile(t, tmpDir)

	assert.Contains(t, content, "/* === Add-on GIS Dataset for Microsoft SQL Server of Vietnamese Provinces Database === */")
	assert.Contains(t, content, "/* Part 1 of 1 */")
	assert.Contains(t, content, "/* Reference: https://github.com/thanglequoc/vietnamese-provinces-database */")
	assert.Contains(t, content, "INSERT INTO gis_provinces(province_code, gis_server_id, area_km2, bbox, geom) VALUES")
	assert.Contains(t, content, "geometry::STGeomFromText('POLYGON((105 20,106 20,106 21,105 21,105 20))', 4326)")
	assert.Contains(t, content, "('00001','xa.1',5.230000")
	assert.Contains(t, content, "('00002','xa.2',3.140000")
	assert.Contains(t, content, ",\n")
	assert.Contains(t, content, ";\n\nGO\n\n")
	assert.Contains(t, content, "-- END OF SCRIPT FILE --")
}

func readGeneratedMssqlGISFile(t *testing.T, rootDir string) string {
	t.Helper()

	manifestMatches, err := filepath.Glob(filepath.Join(rootDir, "output", "sqlserver", "gis", "mssql_ImportData_gis_*.sql.manifest"))
	assert.NoError(t, err)
	if !assert.Len(t, manifestMatches, 1, "should have created one MSSQL GIS manifest file") {
		return ""
	}

	manifestData, err := os.ReadFile(manifestMatches[0])
	assert.NoError(t, err)

	var sb strings.Builder
	for _, name := range strings.Split(strings.TrimSpace(string(manifestData)), "\n") {
		content, err := os.ReadFile(filepath.Join(rootDir, "output", "sqlserver", "gis", name))
		assert.NoError(t, err)
		sb.Write(content)
	}
	return sb.String()
}

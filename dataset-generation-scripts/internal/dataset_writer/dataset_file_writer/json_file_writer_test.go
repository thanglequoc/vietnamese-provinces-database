package dataset_writer

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sapnhapmodels "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
	vn_provinces_tmp_model "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
)

func TestJSONDatasetFileWriter_WriteToFile_FullJSON(t *testing.T) {
	tmpDir := t.TempDir()

	writer := &JSONDatasetFileWriter{
		OutputFolderPath: tmpDir,
	}

	provinces := []vn_provinces_tmp_model.Province{
		{
			Code:                 "01",
			Name:                 "Hà Nội",
			NameEn:               "Ha Noi",
			FullName:             "Thành phố Hà Nội",
			FullNameEn:           "Ha Noi City",
			CodeName:             "ha_noi",
			AdministrativeUnitId: 1,
		},
	}

	err := writer.WriteToFile(nil, nil, provinces, nil)
	assert.NoError(t, err)

	files, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)
	assert.Len(t, files, 3, "should create 3 JSON files (full, simplified, vn_only)")

	// Verify full JSON file
	fullContent, err := os.ReadFile(tmpDir + "/full_json_generated_data_vn_units_" + files[0].Name()[len("full_json_generated_data_vn_units_"):])
	assert.NoError(t, err)

	var data interface{}
	err = json.Unmarshal(fullContent, &data)
	assert.NoError(t, err, "should produce valid JSON")

	contentStr := string(fullContent)
	assert.Contains(t, contentStr, "Hà Nội")
}

func TestJSONDatasetFileWriter_WriteToFile_EmptyDataset(t *testing.T) {
	tmpDir := t.TempDir()

	writer := &JSONDatasetFileWriter{
		OutputFolderPath: tmpDir,
	}

	err := writer.WriteToFile(nil, nil, nil, nil)
	assert.NoError(t, err)

	files, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)
	assert.Len(t, files, 3, "should create 3 JSON files even with empty data")

	// Verify files are created and valid JSON
	for _, f := range files {
		content, err := os.ReadFile(tmpDir + "/" + f.Name())
		assert.NoError(t, err)

		var data interface{}
		err = json.Unmarshal(content, &data)
		assert.NoError(t, err, f.Name()+" should be valid JSON")
	}
}

func TestJSONDatasetFileWriter_WriteToFile_MultipleProvinces(t *testing.T) {
	tmpDir := t.TempDir()

	writer := &JSONDatasetFileWriter{
		OutputFolderPath: tmpDir,
	}

	provinces := []vn_provinces_tmp_model.Province{
		{
			Code:                 "01",
			Name:                 "Hà Nội",
			NameEn:               "Ha Noi",
			FullName:             "Thành phố Hà Nội",
			FullNameEn:           "Ha Noi City",
			CodeName:             "ha_noi",
			AdministrativeUnitId: 1,
		},
		{
			Code:                 "02",
			Name:                 "Hải Phòng",
			NameEn:               "Hai Phong",
			FullName:             "Thành phố Hải Phòng",
			FullNameEn:           "Hai Phong City",
			CodeName:             "hai_phong",
			AdministrativeUnitId: 1,
		},
		{
			Code:                 "10",
			Name:                 "Khánh Hòa",
			NameEn:               "Khanh Hoa",
			FullName:             "Tỉnh Khánh Hòa",
			FullNameEn:           "Khanh Hoa Province",
			CodeName:             "khanh_hoa",
			AdministrativeUnitId: 2,
		},
	}

	err := writer.WriteToFile(nil, nil, provinces, nil)
	assert.NoError(t, err)

	files, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)
	assert.Len(t, files, 3)

	// Verify full JSON contains all provinces
	fullContent, _ := os.ReadFile(tmpDir + "/" + files[0].Name())
	contentStr := string(fullContent)
	assert.Contains(t, contentStr, "Hà Nội")
	assert.Contains(t, contentStr, "Hải Phòng")
	assert.Contains(t, contentStr, "Khánh Hòa")
}

func TestJSONDatasetFileWriter_WriteGISGeoJSONToFile(t *testing.T) {
	rootDir := t.TempDir()
	tmpDir := filepath.Join(rootDir, "geojson")

	writer := &JSONDatasetFileWriter{
		OutputFolderPath: tmpDir,
	}

	provinces := []*sapnhapmodels.SapNhapSiteGeoUnit{
		{
			MaLK:             "diaphanhanhchinhcaptinh_sn.108",
			DienTichKM2:      3359.84,
			VNDSProvinceCode: "01",
			BBoxGeoJSON:      json.RawMessage(`[102.1,20.1,103.2,21.3]`),
			GeomGeoJSON:      json.RawMessage(`{"type":"MultiPolygon","coordinates":[[[[102.1,20.1],[103.2,20.1],[103.2,21.3],[102.1,21.3],[102.1,20.1]]]]}`),
			VNProvince: vn_provinces_tmp_model.Province{
				Code:     "01",
				Name:     "Hà Nội",
				CodeName: "ha_noi",
			},
		},
	}

	wards := []*sapnhapmodels.SapNhapSiteGeoUnit{
		{
			MaLK:             "diaphanhanhchinhcapxa_2025.3256",
			DienTichKM2:      2.97,
			VNDSProvinceCode: "01",
			VNDSWardCode:     "00004",
			BBoxGeoJSON:      json.RawMessage(`[105.1,20.1,105.2,20.2]`),
			GeomGeoJSON:      json.RawMessage(`{"type":"MultiPolygon","coordinates":[[[[105.1,20.1],[105.2,20.1],[105.2,20.2],[105.1,20.2],[105.1,20.1]]]]}`),
			VNProvince: vn_provinces_tmp_model.Province{
				Code:     "01",
				Name:     "Hà Nội",
				CodeName: "ha_noi",
			},
			VNWard: vn_provinces_tmp_model.Ward{
				Code:     "00004",
				Name:     "Ba Đình",
				CodeName: "ba_dinh",
			},
		},
	}

	err := writer.WriteGISGeoJSONToFile(provinces, wards)
	require.NoError(t, err)

	provinceFile := filepath.Join(tmpDir, "01_ha_noi", "01_ha_noi.geojson")
	provinceContent, err := os.ReadFile(provinceFile)
	require.NoError(t, err)

	var provinceJSON map[string]any
	require.NoError(t, json.Unmarshal(provinceContent, &provinceJSON))
	assert.Equal(t, "FeatureCollection", provinceJSON["type"])
	assert.NotNil(t, provinceJSON["bbox"])

	features := provinceJSON["features"].([]any)
	require.Len(t, features, 1)
	feature := features[0].(map[string]any)
	assert.Equal(t, "01", feature["id"])
	assert.Equal(t, provinceJSON["bbox"], feature["bbox"])
	properties := feature["properties"].(map[string]any)
	assert.Equal(t, "Hà Nội", properties["unit_name"])
	assert.Equal(t, "01", properties["unit_code"])
	assert.Equal(t, "ha_noi", properties["unit_code_name"])
	assert.Equal(t, "diaphanhanhchinhcaptinh_sn.108", properties["gis_server_id"])
	assert.Equal(t, 3359.84, properties["area_km2"])

	wardFile := filepath.Join(tmpDir, "01_ha_noi", "wards", "00004_ba_dinh.geojson")
	wardContent, err := os.ReadFile(wardFile)
	require.NoError(t, err)

	var wardJSON map[string]any
	require.NoError(t, json.Unmarshal(wardContent, &wardJSON))
	assert.Equal(t, "FeatureCollection", wardJSON["type"])
	assert.NotNil(t, wardJSON["bbox"])

	wardFeatures := wardJSON["features"].([]any)
	require.Len(t, wardFeatures, 1)
	wardFeature := wardFeatures[0].(map[string]any)
	assert.Equal(t, "00004", wardFeature["id"])
	assert.Equal(t, wardJSON["bbox"], wardFeature["bbox"])

	readmeContent, err := os.ReadFile(filepath.Join(tmpDir, "README.md"))
	require.NoError(t, err)
	assert.Contains(t, string(readmeContent), "Created at:")
	assert.Contains(t, string(readmeContent), "geojson.io")
	assert.Contains(t, string(readmeContent), "{province_code}_{province_code_name}")

	zipMatches, err := filepath.Glob(filepath.Join(rootDir, "vn_provinces_wards_geojson_*.zip"))
	require.NoError(t, err)
	require.Len(t, zipMatches, 1)

	zipReader, err := zip.OpenReader(zipMatches[0])
	require.NoError(t, err)
	defer zipReader.Close()

	names := make([]string, 0, len(zipReader.File))
	for _, f := range zipReader.File {
		names = append(names, f.Name)
	}
	assert.Contains(t, names, "geojson/README.md")
	assert.Contains(t, names, "geojson/01_ha_noi/01_ha_noi.geojson")
	assert.Contains(t, names, "geojson/01_ha_noi/wards/00004_ba_dinh.geojson")
}

package dataset_writer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	vn_provinces_tmp_model "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
)

func TestMongoDBDatasetFileWriter_WriteToFile_CompleteDataset(t *testing.T) {
	tmpDir := t.TempDir()
	
	writer := &MongoDBDatasetFileWriter{
		OutputFolderPath: tmpDir,
	}
	
	regions := []vn_provinces_tmp_model.AdministrativeRegion{
		{Id: 1, Name: "Test Region", NameEn: "Test Region", CodeName: "test", CodeNameEn: "test"},
	}
	
	administrativeUnits := []vn_provinces_tmp_model.AdministrativeUnit{
		{Id: 1, FullName: "TP", FullNameEn: "City", ShortName: "TP", ShortNameEn: "City", CodeName: "tp", CodeNameEn: "city"},
	}
	
	provinces := []vn_provinces_tmp_model.Province{
		{Code: "01", Name: "Hà Nội", NameEn: "Ha Noi", FullName: "Thành phố Hà Nội", FullNameEn: "Ha Noi", CodeName: "ha_noi", AdministrativeUnitId: 1},
	}
	
	err := writer.WriteToFile(regions, administrativeUnits, provinces, []vn_provinces_tmp_model.Ward{})
	assert.NoError(t, err)
	
	files, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)
	
	// Should have created files
	assert.GreaterOrEqual(t, len(files), 2)
	
	// Find and verify administrative_units file
	var adminUnitsFile string
	for _, f := range files {
		name := f.Name()
		if len(name) >= 23 && name[:23] == "administrative_units" {
			adminUnitsFile = filepath.Join(tmpDir, name)
			break
		}
	}
	
	if adminUnitsFile != "" {
		content, _ := os.ReadFile(adminUnitsFile)
		var unitsData []map[string]interface{}
		json.Unmarshal(content, &unitsData)
		assert.Len(t, unitsData, 1)
	}
	
	// Find and verify administrative_regions file
	var regionsFile string
	for _, f := range files {
		name := f.Name()
		if len(name) >= 24 && name[:24] == "administrative_regions" {
			regionsFile = filepath.Join(tmpDir, name)
			break
		}
	}
	
	if regionsFile != "" {
		content, _ := os.ReadFile(regionsFile)
		var regionsData []map[string]interface{}
		json.Unmarshal(content, &regionsData)
		assert.Len(t, regionsData, 1)
	}
	
	// Find and verify mongo_data_vn_unit file
	var mongoDataFile string
	for _, f := range files {
		name := f.Name()
		if len(name) >= 18 && name[:18] == "mongo_data_vn_unit" {
			mongoDataFile = filepath.Join(tmpDir, name)
			break
		}
	}
	
	if mongoDataFile != "" {
		content, err := os.ReadFile(mongoDataFile)
		assert.NoError(t, err)
		assert.Greater(t, len(content), 0, "mongo_data_vn_unit file should have content")
		
		var provincesData []map[string]interface{}
		err = json.Unmarshal(content, &provincesData)
		assert.NoError(t, err)
		assert.Greater(t, len(provincesData), 0, "should have at least one province")
	}
}

func TestMongoDBDatasetFileWriter_WriteToFile_PostalCodes(t *testing.T) {
	tmpDir := t.TempDir()

	writer := &MongoDBDatasetFileWriter{OutputFolderPath: tmpDir}

	provinces := []vn_provinces_tmp_model.Province{
		{
			Code: "01", Name: "Hà Nội", NameEn: "Ha Noi",
			FullName: "Thành phố Hà Nội", FullNameEn: "Ha Noi City",
			CodeName: "ha_noi", AdministrativeUnitId: 1, PostalCodePrefix: "10, 11, 12, 13, 14",
			Wards: []*vn_provinces_tmp_model.Ward{
				{
					Code: "00070", Name: "Hoàn Kiếm", NameEn: "Hoan Kiem",
					FullName: "Phường Hoàn Kiếm", FullNameEn: "Hoan Kiem Ward",
					CodeName: "hoan_kiem", ProvinceCode: "01", AdministrativeUnitId: 3, PostalCode: "11024",
				},
			},
		},
	}

	err := writer.WriteToFile(nil, nil, provinces, []vn_provinces_tmp_model.Ward{})
	assert.NoError(t, err)

	files, _ := os.ReadDir(tmpDir)
	var mongoContent []byte
	for _, f := range files {
		if len(f.Name()) >= 18 && f.Name()[:18] == "mongo_data_vn_unit" {
			mongoContent, _ = os.ReadFile(filepath.Join(tmpDir, f.Name()))
		}
	}
	assert.Contains(t, string(mongoContent), "11024")
	assert.Contains(t, string(mongoContent), "10, 11, 12, 13, 14")
}

func TestMongoDBDatasetFileWriter_WriteToFile_README(t *testing.T) {
	tmpDir := t.TempDir()
	writer := MongoDBDatasetFileWriter{OutputFolderPath: tmpDir}
	provinces := []vn_provinces_tmp_model.Province{{Code: "01", Name: "Hà Nội", AdministrativeUnitId: 1}}

	err := writer.WriteToFile(nil, nil, provinces, nil)
	assert.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "README.md"))
	assert.NoError(t, err)
	s := string(content)
	assert.Contains(t, s, "**Generated at:")
	assert.Contains(t, s, "mongo_data_vn_unit.json")
	assert.Contains(t, s, "## Sample Queries")
	assert.Contains(t, s, "gis/")
}

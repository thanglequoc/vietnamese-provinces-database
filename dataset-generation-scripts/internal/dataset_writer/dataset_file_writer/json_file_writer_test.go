package dataset_writer

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
		assert.NoError(t, err, f.Name() + " should be valid JSON")
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
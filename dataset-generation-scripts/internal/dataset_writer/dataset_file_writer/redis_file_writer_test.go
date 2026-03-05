package dataset_writer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	vn_provinces_tmp_model "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
)

func TestRedisDatasetFileWriter_WriteToFile_Provinces(t *testing.T) {
	tmpDir := t.TempDir()
	
	writer := &RedisDatasetFileWriter{
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
	assert.Len(t, files, 1)
	
	content, err := os.ReadFile(filepath.Join(tmpDir, files[0].Name()))
	assert.NoError(t, err)
	
	contentStr := string(content)
	assert.Contains(t, contentStr, "HSET province:01")
	assert.Contains(t, contentStr, "code \"01\"")
	assert.Contains(t, contentStr, "name \"Hà Nội\"")
	assert.Contains(t, contentStr, "nameEn \"Ha Noi\"")
}

func TestRedisDatasetFileWriter_WriteToFile_AdministrativeUnits(t *testing.T) {
	tmpDir := t.TempDir()
	
	writer := &RedisDatasetFileWriter{
		OutputFolderPath: tmpDir,
	}
	
	administrativeUnits := []vn_provinces_tmp_model.AdministrativeUnit{
		{
			Id:          1,
			FullName:    "Thành phố",
			FullNameEn:  "City",
			ShortName:   "TP",
			ShortNameEn: "City",
			CodeName:    "thanh_pho",
			CodeNameEn:  "city",
		},
	}
	
	err := writer.WriteToFile(nil, administrativeUnits, nil, nil)
	assert.NoError(t, err)
	
	files, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)
	
	content, err := os.ReadFile(filepath.Join(tmpDir, files[0].Name()))
	assert.NoError(t, err)
	
	contentStr := string(content)
	assert.Contains(t, contentStr, "HSET administrativeUnit:1")
	assert.Contains(t, contentStr, "fullName \"Thành phố\"")
	assert.Contains(t, contentStr, "fullNameEn \"City\"")
	assert.Contains(t, contentStr, "shortName \"TP\"")
}

func TestRedisDatasetFileWriter_WriteToFile_Regions(t *testing.T) {
	tmpDir := t.TempDir()
	
	writer := &RedisDatasetFileWriter{
		OutputFolderPath: tmpDir,
	}
	
	regions := []vn_provinces_tmp_model.AdministrativeRegion{
		{
			Id:        1,
			Name:      "Đông Bắc Bộ",
			NameEn:    "Northeast",
			CodeName:  "dong_bac_bo",
			CodeNameEn: "northeast",
		},
	}
	
	err := writer.WriteToFile(regions, nil, nil, nil)
	assert.NoError(t, err)
	
	files, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)
	
	content, err := os.ReadFile(filepath.Join(tmpDir, files[0].Name()))
	assert.NoError(t, err)
	
	contentStr := string(content)
	assert.Contains(t, contentStr, "HSET region:1")
	assert.Contains(t, contentStr, "name \"Đông Bắc Bộ\"")
	assert.Contains(t, contentStr, "nameEn \"Northeast\"")
	assert.Contains(t, contentStr, "codeName \"dong_bac_bo\"")
}

func TestRedisDatasetFileWriter_WriteToFile_Wards(t *testing.T) {
	tmpDir := t.TempDir()
	
	writer := &RedisDatasetFileWriter{
		OutputFolderPath: tmpDir,
	}
	
	wards := []vn_provinces_tmp_model.Ward{
		{
			Code:                 "001",
			Name:                 "Bắc Sơn",
			NameEn:               "Bac Son",
			FullName:             "Phường Bắc Sơn",
			FullNameEn:           "Bac Son Ward",
			CodeName:             "bac_son",
			ProvinceCode:         "01",
			AdministrativeUnitId: 3,
		},
	}
	
	err := writer.WriteToFile(nil, nil, nil, wards)
	assert.NoError(t, err)
	
	files, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)
	
	content, err := os.ReadFile(filepath.Join(tmpDir, files[0].Name()))
	assert.NoError(t, err)
	
	contentStr := string(content)
	assert.Contains(t, contentStr, "HSET ward:001")
	assert.Contains(t, contentStr, "name \"Bắc Sơn\"")
	assert.Contains(t, contentStr, "districtCode \"01\"")
	assert.Contains(t, contentStr, "SADD province:01:wards")
	assert.Contains(t, contentStr, "HSET province:01:wards:vn")
	assert.Contains(t, contentStr, "HSET province:01:wards:en")
}

func TestRedisDatasetFileWriter_WriteToFile_CompleteDataset(t *testing.T) {
	tmpDir := t.TempDir()
	
	writer := &RedisDatasetFileWriter{
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
	
	wards := []vn_provinces_tmp_model.Ward{
		{Code: "001", Name: "Bắc Sơn", NameEn: "Bac Son", FullName: "Phường Bắc Sơn", FullNameEn: "Bac Son", CodeName: "bac_son", ProvinceCode: "01", AdministrativeUnitId: 3},
	}
	
	err := writer.WriteToFile(regions, administrativeUnits, provinces, wards)
	assert.NoError(t, err)
	
	files, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)
	assert.Len(t, files, 1)
	
	content, err := os.ReadFile(filepath.Join(tmpDir, files[0].Name()))
	assert.NoError(t, err)
	
	contentStr := string(content)
	
	// Verify all record types are present
	assert.Contains(t, contentStr, "HSET administrativeUnit:")
	assert.Contains(t, contentStr, "HSET region:")
	assert.Contains(t, contentStr, "HSET province:")
	assert.Contains(t, contentStr, "HSET ward:")
	assert.Contains(t, contentStr, "SADD province:")
	
	// Verify Redis commands
	lines := strings.Split(contentStr, "\n")
	hsetCount := 0
	saddCount := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "HSET") {
			hsetCount++
		}
		if strings.HasPrefix(line, "SADD") {
			saddCount++
		}
	}
	assert.Greater(t, hsetCount, 0, "should have HSET commands")
	assert.Greater(t, saddCount, 0, "should have SADD commands")
}
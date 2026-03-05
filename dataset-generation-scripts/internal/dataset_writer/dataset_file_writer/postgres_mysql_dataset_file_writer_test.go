package dataset_writer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	vn_provinces_tmp_model "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
)

func TestPostgresMySQLDatasetFileWriter_WriteToFile_Regions(t *testing.T) {
	// Create temporary file for testing
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test_%s.sql")
	
	writer := &PostgresMySQLDatasetFileWriter{
		OutputFilePath: outputPath,
	}
	
	regions := []vn_provinces_tmp_model.AdministrativeRegion{
		{
			Id:        1,
			Name:      "Đông Bắc Bộ",
			NameEn:    "Northeast",
			CodeName:  "dong_bac_bo",
			CodeNameEn: "northeast",
		},
		{
			Id:        2,
			Name:      "Tây Bắc Bộ",
			NameEn:    "Northwest",
			CodeName:  "tay_bac_bo",
			CodeNameEn: "northwest",
		},
	}
	
	err := writer.WriteToFile(regions, nil, nil, nil)
	assert.NoError(t, err)
	
	// Read the file and verify content
	files, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)
	assert.Len(t, files, 1, "should have created one output file")
	
	content, err := os.ReadFile(filepath.Join(tmpDir, files[0].Name()))
	assert.NoError(t, err)
	
	contentStr := string(content)
	
	// Verify header
	assert.Contains(t, contentStr, "Vietnamese Provinces Database Dataset for PostgreSQL/MySQL")
	assert.Contains(t, contentStr, "administrative_regions")
	
	// Verify regions are included
	assert.Contains(t, contentStr, "INSERT INTO administrative_regions(id,name,name_en,code_name,code_name_en) VALUES")
	assert.Contains(t, contentStr, "(1,'Đông Bắc Bộ','Northeast','dong_bac_bo','northeast')")
	assert.Contains(t, contentStr, "(2,'Tây Bắc Bộ','Northwest','tay_bac_bo','northwest')")
}

func TestPostgresMySQLDatasetFileWriter_WriteToFile_AdministrativeUnits(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test_%s.sql")
	
	writer := &PostgresMySQLDatasetFileWriter{
		OutputFilePath: outputPath,
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
		{
			Id:          2,
			FullName:    "Tỉnh",
			FullNameEn:  "Province",
			ShortName:   "Tỉnh",
			ShortNameEn: "Province",
			CodeName:    "tinh",
			CodeNameEn:  "province",
		},
	}
	
	err := writer.WriteToFile(nil, administrativeUnits, nil, nil)
	assert.NoError(t, err)
	
	files, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)
	
	content, err := os.ReadFile(filepath.Join(tmpDir, files[0].Name()))
	assert.NoError(t, err)
	
	contentStr := string(content)
	assert.Contains(t, contentStr, "administrative_units")
	assert.Contains(t, contentStr, "(1,'Thành phố','City','TP','City','thanh_pho','city')")
	assert.Contains(t, contentStr, "(2,'Tỉnh','Province','Tỉnh','Province','tinh','province')")
}

func TestPostgresMySQLDatasetFileWriter_WriteToFile_Provinces(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test_%s.sql")
	
	writer := &PostgresMySQLDatasetFileWriter{
		OutputFilePath: outputPath,
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
	}
	
	err := writer.WriteToFile(nil, nil, provinces, nil)
	assert.NoError(t, err)
	
	files, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)
	
	content, err := os.ReadFile(filepath.Join(tmpDir, files[0].Name()))
	assert.NoError(t, err)
	
	contentStr := string(content)
	assert.Contains(t, contentStr, "INSERT INTO provinces(code,name,name_en,full_name,full_name_en,code_name,administrative_unit_id)")
	assert.Contains(t, contentStr, "('01','Hà Nội','Ha Noi','Thành phố Hà Nội','Ha Noi City','ha_noi',1)")
	assert.Contains(t, contentStr, "('02','Hải Phòng','Hai Phong','Thành phố Hải Phòng','Hai Phong City','hai_phong',1)")
}

func TestPostgresMySQLDatasetFileWriter_WriteToFile_Wards(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test_%s.sql")
	
	writer := &PostgresMySQLDatasetFileWriter{
		OutputFilePath: outputPath,
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
		{
			Code:                 "002",
			Name:                 "Tân Xã",
			NameEn:               "Tan Xa",
			FullName:             "Xã Tân Xã",
			FullNameEn:           "Tan Xa Commune",
			CodeName:             "tan_xa",
			ProvinceCode:         "01",
			AdministrativeUnitId: 4,
		},
	}
	
	err := writer.WriteToFile(nil, nil, nil, wards)
	assert.NoError(t, err)
	
	files, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)
	
	content, err := os.ReadFile(filepath.Join(tmpDir, files[0].Name()))
	assert.NoError(t, err)
	
	contentStr := string(content)
	assert.Contains(t, contentStr, "INSERT INTO wards(code,name,name_en,full_name,full_name_en,code_name,province_code,administrative_unit_id)")
	assert.Contains(t, contentStr, "('001','Bắc Sơn','Bac Son','Phường Bắc Sơn','Bac Son Ward','bac_son','01',3)")
	assert.Contains(t, contentStr, "('002','Tân Xã','Tan Xa','Xã Tân Xã','Tan Xa Commune','tan_xa','01',4)")
}

func TestPostgresMySQLDatasetFileWriter_WriteToFile_EscapeSingleQuote(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test_%s.sql")
	
	writer := &PostgresMySQLDatasetFileWriter{
		OutputFilePath: outputPath,
	}
	
	provinces := []vn_provinces_tmp_model.Province{
		{
			Code:                 "99",
			Name:                 "Ea H'MLay",
			NameEn:               "Ea H''MLay",
			FullName:             "Xã Ea H'MLay",
			FullNameEn:           "Ea H'MLay Commune",
			CodeName:             "ea_hmlay",
			AdministrativeUnitId: 4,
		},
	}
	
	err := writer.WriteToFile(nil, nil, provinces, nil)
	assert.NoError(t, err)
	
	files, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)
	
	content, err := os.ReadFile(filepath.Join(tmpDir, files[0].Name()))
	assert.NoError(t, err)
	
	contentStr := string(content)
	// Verify single quotes are properly escaped (double single quotes)
	assert.Contains(t, contentStr, "('99','Ea H''MLay','Ea H''''MLay','Xã Ea H''MLay','Ea H''MLay Commune','ea_hmlay',4)")
}

func TestPostgresMySQLDatasetFileWriter_WriteToFile_BatchInsert(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test_%s.sql")
	
	writer := &PostgresMySQLDatasetFileWriter{
		OutputFilePath: outputPath,
	}
	
	// Create 52 provinces to test batch insert (batch size is 50)
	provinces := make([]vn_provinces_tmp_model.Province, 52)
	for i := 0; i < 52; i++ {
		provinces[i] = vn_provinces_tmp_model.Province{
			Code:                 string(rune('A' + i)),
			Name:                 "Test Province",
			NameEn:               "Test Province",
			FullName:             "Tỉnh Test Province",
			FullNameEn:           "Test Province",
			CodeName:             "test_province",
			AdministrativeUnitId: 2,
		}
	}
	
	err := writer.WriteToFile(nil, nil, provinces, nil)
	assert.NoError(t, err)
	
	files, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)
	
	content, err := os.ReadFile(filepath.Join(tmpDir, files[0].Name()))
	assert.NoError(t, err)
	
	contentStr := string(content)
	
	// Count occurrences of INSERT INTO provinces - should be 2 (1 batch of 50 + 1 batch of 2)
	count := strings.Count(contentStr, "INSERT INTO provinces(")
	assert.Equal(t, 2, count, "should have 2 batch insert statements")
	
	// Verify the batch data format - check for a few specific entries
	assert.Contains(t, contentStr, "('A','Test Province','Test Province','Tỉnh Test Province'")
	assert.Contains(t, contentStr, "('s','Test Province','Test Province'")
	assert.Contains(t, contentStr, "('t','Test Province'")
}

func TestPostgresMySQLDatasetFileWriter_WriteToFile_CompleteDataset(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test_%s.sql")
	
	writer := &PostgresMySQLDatasetFileWriter{
		OutputFilePath: outputPath,
	}
	
	// Use test fixtures
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
	
	content, err := os.ReadFile(filepath.Join(tmpDir, files[0].Name()))
	assert.NoError(t, err)
	
	contentStr := string(content)
	
	// Verify all sections are present
	assert.Contains(t, contentStr, "administrative_regions")
	assert.Contains(t, contentStr, "administrative_units")
	assert.Contains(t, contentStr, "provinces")
	assert.Contains(t, contentStr, "wards")
	assert.Contains(t, contentStr, "END OF SCRIPT FILE")
}
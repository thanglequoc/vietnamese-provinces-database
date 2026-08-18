package dataset_writer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	dataset_file_writer_dto "github.com/thanglequoc-vn-provinces/v2/internal/dataset_writer/dataset_file_writer/dto"
	sapnhapbandomodel "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
	"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
)

func TestWriteMongoGISDataToFile(t *testing.T) {
	tmpDir := t.TempDir()

	bboxJSON := json.RawMessage(`[105.5, 20.5, 106.0, 21.0]`)
	geomJSON := json.RawMessage(`{"type":"MultiPolygon","coordinates":[]}`)

	geoProvinces := []*sapnhapbandomodel.SapNhapSiteGeoUnit{
		{
			Ma: "01", MaLK: "gis_001", Ten: "Hà Nội", VNDSProvinceCode: "01",
			DienTichKM2: 3359.84,
			BBoxGeoJSON: bboxJSON, GeomGeoJSON: geomJSON,
			VNProvince: model.Province{
				Code: "01", Name: "Hà Nội", NameEn: "Hanoi",
				FullName: "Thành phố Hà Nội", FullNameEn: "Hanoi City",
				CodeName: "ha_noi",
				PostalCodePrefix: "10, 11, 12, 13, 14",
				AdministrativeUnit: model.AdministrativeUnit{
					Id: 1, FullName: "Thành phố", FullNameEn: "City",
					ShortName: "TP.", ShortNameEn: "City",
					CodeName: "thanh_pho", CodeNameEn: "city",
				},
			},
		},
	}

	geoWards := []*sapnhapbandomodel.SapNhapSiteGeoUnit{
		{
			Ma: "00001", MaLK: "gis_00001", Ten: "Ba Đình", VNDSProvinceCode: "01",
			DienTichKM2: 5.23,
			BBoxGeoJSON: bboxJSON, GeomGeoJSON: geomJSON,
			VNWard: model.Ward{
				Code: "00001", Name: "Ba Đình", NameEn: "Ba Dinh",
				FullName: "Phường Ba Đình", FullNameEn: "Ba Dinh Ward",
				CodeName: "ba_dinh",
				PostalCode: "11024",
				AdministrativeUnit: model.AdministrativeUnit{
					Id: 3, FullName: "Phường", FullNameEn: "Ward",
					ShortName: "P.", ShortNameEn: "Ward",
					CodeName: "phuong", CodeNameEn: "ward",
				},
			},
		},
	}

	writer := MongoDBDatasetFileWriter{OutputFolderPath: tmpDir}
	err := writer.WriteMongoGISDataToFile(geoProvinces, geoWards)
	if err != nil {
		t.Fatalf("WriteMongoGISDataToFile failed: %v", err)
	}

	// Verify province GIS file exists
	provinceFiles, err := filepath.Glob(filepath.Join(tmpDir, "mongo_data_vn_province_gis*.json"))
	if err != nil || len(provinceFiles) == 0 {
		t.Fatalf("no province GIS file found in %s", tmpDir)
	}

	// Read and verify province GIS file
	provinceData, err := os.ReadFile(provinceFiles[0])
	if err != nil {
		t.Fatalf("failed to read province GIS file: %v", err)
	}
	var provinceDocs []dataset_file_writer_dto.MongoGISProvinceDocument
	if err := json.Unmarshal(provinceData, &provinceDocs); err != nil {
		t.Fatalf("failed to unmarshal province GIS JSON: %v", err)
	}
	if len(provinceDocs) != 1 {
		t.Fatalf("expected 1 province doc, got %d", len(provinceDocs))
	}
	if provinceDocs[0].Code != "01" {
		t.Errorf("expected province Code '01', got %q", provinceDocs[0].Code)
	}
	if provinceDocs[0].GIS == nil {
		t.Fatal("expected province GIS to be populated")
	}
	if provinceDocs[0].GIS.Center.Type != "Point" {
		t.Errorf("expected province Center.Type 'Point', got %q", provinceDocs[0].GIS.Center.Type)
	}
	if provinceDocs[0].GIS.Properties == nil {
		t.Fatal("expected province GIS Properties to be populated")
	}
	if provinceDocs[0].GIS.Properties.PostalCodePrefix != "10, 11, 12, 13, 14" {
		t.Errorf("expected PostalCodePrefix '10, 11, 12, 13, 14', got %q", provinceDocs[0].GIS.Properties.PostalCodePrefix)
	}

	// Verify ward GIS file exists
	wardFiles, err := filepath.Glob(filepath.Join(tmpDir, "mongo_data_vn_ward_gis*.json"))
	if err != nil || len(wardFiles) == 0 {
		t.Fatalf("no ward GIS file found in %s", tmpDir)
	}

	// Read and verify ward GIS file
	wardData, err := os.ReadFile(wardFiles[0])
	if err != nil {
		t.Fatalf("failed to read ward GIS file: %v", err)
	}
	var wardDocs []dataset_file_writer_dto.MongoGISWardDocument
	if err := json.Unmarshal(wardData, &wardDocs); err != nil {
		t.Fatalf("failed to unmarshal ward GIS JSON: %v", err)
	}
	if len(wardDocs) != 1 {
		t.Fatalf("expected 1 ward doc, got %d", len(wardDocs))
	}
	if wardDocs[0].Code != "00001" {
		t.Errorf("expected ward Code '00001', got %q", wardDocs[0].Code)
	}
	if wardDocs[0].ProvinceCode != "01" {
		t.Errorf("expected ProvinceCode '01', got %q", wardDocs[0].ProvinceCode)
	}
	if wardDocs[0].GIS == nil {
		t.Fatal("expected ward GIS to be populated")
	}
	if wardDocs[0].GIS.Properties == nil {
		t.Fatal("expected ward GIS Properties to be populated")
	}
	if wardDocs[0].GIS.Properties.PostalCode != "11024" {
		t.Errorf("expected PostalCode '11024', got %q", wardDocs[0].GIS.Properties.PostalCode)
	}

	// Verify create_indexes.js exists
	indexPath := filepath.Join(tmpDir, "create_indexes.js")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Fatal("create_indexes.js not found")
	}

	// GIS subfolder no longer carries its own README
	readmePath := filepath.Join(tmpDir, "README.md")
	if _, err := os.Stat(readmePath); !os.IsNotExist(err) {
		t.Fatal("README.md should NOT be written in the mongodb gis folder")
	}
}
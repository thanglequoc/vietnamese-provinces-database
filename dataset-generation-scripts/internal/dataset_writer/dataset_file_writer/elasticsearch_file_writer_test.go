package dataset_writer

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	dataset_file_writer_dto "github.com/thanglequoc-vn-provinces/v2/internal/dataset_writer/dataset_file_writer/dto"
	file_writer_helper "github.com/thanglequoc-vn-provinces/v2/internal/dataset_writer/dataset_file_writer/helper"
	sapnhapbandomodel "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
	"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
)

func TestWriteToFile_NonGIS(t *testing.T) {
	tmpDir := t.TempDir()

	provinces := []model.Province{
		{
			Code:     "01",
			Name:     "Hà Nội",
			NameEn:   "Hanoi",
			FullName: "Thành phố Hà Nội",
			FullNameEn: "Hanoi City",
			CodeName: "ha_noi",
			PostalCodePrefix: "10, 11, 12, 13, 14",
			AdministrativeUnit: model.AdministrativeUnit{
				Id: 1, FullName: "Thành phố", FullNameEn: "City",
				ShortName: "TP.", ShortNameEn: "City",
				CodeName: "thanh_pho", CodeNameEn: "city",
			},
			Wards: []*model.Ward{
				{
					Code: "00001", Name: "Ba Đình", NameEn: "Ba Dinh",
					FullName: "Phường Ba Đình", FullNameEn: "Ba Dinh Ward",
					CodeName: "ba_dinh", PostalCode: "11024", AdministrativeUnit: model.AdministrativeUnit{
						Id: 3, FullName: "Phường", FullNameEn: "Ward",
						ShortName: "P.", ShortNameEn: "Ward",
						CodeName: "phuong", CodeNameEn: "ward",
					},
				},
			},
		},
	}

	writer := ElasticsearchDatasetFileWriter{OutputFolderPath: tmpDir}
	err := writer.WriteToFile(nil, nil, provinces, nil)
	if err != nil {
		t.Fatalf("WriteToFile failed: %v", err)
	}

	// Verify provinces.ndjson
	ndjsonPath := filepath.Join(tmpDir, "provinces.ndjson")
	data, err := os.ReadFile(ndjsonPath)
	if err != nil {
		t.Fatalf("failed to read provinces.ndjson: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines in ndjson, got %d", len(lines))
	}

	// Verify index action line
	var action map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &action); err != nil {
		t.Fatalf("line 0 is not valid JSON: %v", err)
	}
	indexAction, ok := action["index"].(map[string]interface{})
	if !ok {
		t.Fatal("line 0 missing 'index' key")
	}
	if indexAction["_index"] != "provinces" {
		t.Errorf("expected _index 'provinces', got %v", indexAction["_index"])
	}
	if indexAction["_id"] != "01" {
		t.Errorf("expected _id '01', got %v", indexAction["_id"])
	}

	// Verify document line
	var doc dataset_file_writer_dto.ElasticsearchProvinceDocument
	if err := json.Unmarshal([]byte(lines[1]), &doc); err != nil {
		t.Fatalf("line 1 is not valid JSON: %v", err)
	}
	if doc.Code != "01" {
		t.Errorf("expected Code '01', got %q", doc.Code)
	}
	if doc.Name != "Hà Nội" {
		t.Errorf("expected Name 'Hà Nội', got %q", doc.Name)
	}
	if doc.Meta == nil {
		t.Fatal("expected Meta to be set")
	}
	if doc.Meta.DatasetVersion != esDatasetVer {
		t.Errorf("expected DatasetVersion %q, got %q", esDatasetVer, doc.Meta.DatasetVersion)
	}
	if len(doc.Wards) != 1 {
		t.Errorf("expected 1 ward, got %d", len(doc.Wards))
	}
	if len(doc.SearchKeywords) == 0 {
		t.Error("expected SearchKeywords to be populated")
	}
	if doc.PostalCodePrefix != "10, 11, 12, 13, 14" {
		t.Errorf("expected PostalCodePrefix '10, 11, 12, 13, 14', got %q", doc.PostalCodePrefix)
	}
	if len(doc.Wards) == 1 && doc.Wards[0].PostalCode != "11024" {
		t.Errorf("expected ward PostalCode '11024', got %q", doc.Wards[0].PostalCode)
	}

	// Verify mappings/provinces.json
	mappingPath := filepath.Join(tmpDir, "mappings", "provinces.json")
	if _, err := os.Stat(mappingPath); os.IsNotExist(err) {
		t.Fatal("mappings/provinces.json not found")
	}
	mapping, err := os.ReadFile(mappingPath)
	if err != nil {
		t.Fatalf("failed to read mappings/provinces.json: %v", err)
	}
	if !strings.Contains(string(mapping), "PostalCode") {
		t.Error("mapping missing PostalCode field")
	}
	if !strings.Contains(string(mapping), "PostalCodePrefix") {
		t.Error("mapping missing PostalCodePrefix field")
	}

	// Verify README.md
	readmePath := filepath.Join(tmpDir, "README.md")
	readme, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}
	if len(readme) == 0 {
		t.Fatal("README.md is empty")
	}
	if !bytes.Contains(readme, []byte("**Generated at:")) {
		t.Fatal("README.md missing bold Generated at header")
	}
	if !bytes.Contains(readme, []byte("## Files")) {
		t.Fatal("README.md missing Files section")
	}
	if !bytes.Contains(readme, []byte("## Sample Queries")) {
		t.Fatal("README.md missing Sample Queries section")
	}
	if !bytes.Contains(readme, []byte("## Overview")) {
		t.Fatal("README.md missing Overview section")
	}
	if !bytes.Contains(readme, []byte("## Data Structure")) {
		t.Fatal("README.md missing Data Structure section")
	}
	if !bytes.Contains(readme, []byte("## Sample Document")) {
		t.Fatal("README.md missing Sample Document section")
	}
	if !bytes.Contains(readme, []byte("## Quick Start")) {
		t.Fatal("README.md missing Quick Start section")
	}
	if !bytes.Contains(readme, []byte("_bulk")) {
		t.Fatal("README.md missing _bulk import reference")
	}
}

func TestWriteElasticsearchGISDataToFile_GIS(t *testing.T) {
	tmpDir := t.TempDir()

	bboxJSON := json.RawMessage(`[105.5, 20.5, 106.0, 21.0]`)
	geomJSON := json.RawMessage(`{"type":"MultiPolygon","coordinates":[[[[105.5,20.5],[106.0,20.5],[106.0,21.0],[105.5,21.0],[105.5,20.5]]]]}`)

	geoProvinces := []*sapnhapbandomodel.SapNhapSiteGeoUnit{
		{
			Ma: "01", MaLK: "diaphanhanhchinhcaptinh_sn.108", Ten: "Hà Nội", VNDSProvinceCode: "01",
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
			Ma: "00001", MaLK: "diaphanhanhchinhphuong_sn.456", Ten: "Ba Đình", VNDSProvinceCode: "01",
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

	writer := ElasticsearchDatasetFileWriter{OutputFolderPath: tmpDir}
	err := writer.WriteElasticsearchGISDataToFile(geoProvinces, geoWards)
	if err != nil {
		t.Fatalf("WriteElasticsearchGISDataToFile failed: %v", err)
	}

	// Verify provinces-gis.ndjson
	ndjsonPath := filepath.Join(tmpDir, "provinces-gis.ndjson")
	data, err := os.ReadFile(ndjsonPath)
	if err != nil {
		t.Fatalf("failed to read provinces-gis.ndjson: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines in gis ndjson, got %d", len(lines))
	}

	// Verify index action uses provinces-gis
	var action map[string]interface{}
	json.Unmarshal([]byte(lines[0]), &action)
	indexAction := action["index"].(map[string]interface{})
	if indexAction["_index"] != "provinces-gis" {
		t.Errorf("expected _index 'provinces-gis', got %v", indexAction["_index"])
	}

	// Verify document has GIS field
	var doc dataset_file_writer_dto.ElasticsearchProvinceDocument
	if err := json.Unmarshal([]byte(lines[1]), &doc); err != nil {
		t.Fatalf("line 1 is not valid JSON: %v", err)
	}
	if doc.GIS == nil {
		t.Fatal("expected GIS field to be populated")
	}
	if doc.GIS.Center.Lat == 0 && doc.GIS.Center.Lon == 0 {
		t.Error("expected non-zero center coordinates")
	}
	if len(doc.Wards) != 1 {
		t.Errorf("expected 1 ward, got %d", len(doc.Wards))
	}
	if doc.PostalCodePrefix != "10, 11, 12, 13, 14" {
		t.Errorf("expected province PostalCodePrefix '10, 11, 12, 13, 14', got %q", doc.PostalCodePrefix)
	}
	if doc.Wards[0].PostalCode != "11024" {
		t.Errorf("expected ward PostalCode '11024', got %q", doc.Wards[0].PostalCode)
	}
	if doc.Wards[0].GIS == nil {
		t.Error("expected ward GIS field to be populated")
	}

	// Verify province GIS Properties
	if doc.GIS.Properties == nil {
		t.Fatal("expected province GIS Properties to be populated")
	}
	if doc.GIS.Properties.Code != "01" {
		t.Errorf("expected Properties.Code '01', got %q", doc.GIS.Properties.Code)
	}
	if doc.GIS.Properties.Name != "Hà Nội" {
		t.Errorf("expected Properties.Name 'Hà Nội', got %q", doc.GIS.Properties.Name)
	}
	if doc.GIS.Properties.GisServerId != "diaphanhanhchinhcaptinh_sn.108" {
		t.Errorf("expected Properties.GisServerId 'diaphanhanhchinhcaptinh_sn.108', got %q", doc.GIS.Properties.GisServerId)
	}
	if doc.GIS.Properties.AreaKm2 != 3359.84 {
		t.Errorf("expected Properties.AreaKm2 3359.84, got %f", doc.GIS.Properties.AreaKm2)
	}
	if doc.GIS.Properties.PostalCodePrefix != "10, 11, 12, 13, 14" {
		t.Errorf("expected Properties.PostalCodePrefix '10, 11, 12, 13, 14', got %q", doc.GIS.Properties.PostalCodePrefix)
	}

	// Verify ward GIS Properties
	if doc.Wards[0].GIS.Properties == nil {
		t.Fatal("expected ward GIS Properties to be populated")
	}
	if doc.Wards[0].GIS.Properties.Code != "00001" {
		t.Errorf("expected ward Properties.Code '00001', got %q", doc.Wards[0].GIS.Properties.Code)
	}
	if doc.Wards[0].GIS.Properties.Name != "Ba Đình" {
		t.Errorf("expected ward Properties.Name 'Ba Đình', got %q", doc.Wards[0].GIS.Properties.Name)
	}
	if doc.Wards[0].GIS.Properties.GisServerId != "diaphanhanhchinhphuong_sn.456" {
		t.Errorf("expected ward Properties.GisServerId 'diaphanhanhchinhphuong_sn.456', got %q", doc.Wards[0].GIS.Properties.GisServerId)
	}
	if doc.Wards[0].GIS.Properties.AreaKm2 != 5.23 {
		t.Errorf("expected ward Properties.AreaKm2 5.23, got %f", doc.Wards[0].GIS.Properties.AreaKm2)
	}
	if doc.Wards[0].GIS.Properties.PostalCode != "11024" {
		t.Errorf("expected ward Properties.PostalCode '11024', got %q", doc.Wards[0].GIS.Properties.PostalCode)
	}

	// Verify mappings/provinces-gis.json
	mappingPath := filepath.Join(tmpDir, "mappings", "provinces-gis.json")
	if _, err := os.Stat(mappingPath); os.IsNotExist(err) {
		t.Fatal("mappings/provinces-gis.json not found")
	}
}

func TestParseBBox(t *testing.T) {
	bboxJSON := json.RawMessage(`[105.5, 20.5, 106.0, 21.0]`)
	bbox, center, err := parseBBox(bboxJSON)
	if err != nil {
		t.Fatalf("parseBBox failed: %v", err)
	}

	if bbox.MinLongitude != 105.5 {
		t.Errorf("expected MinLongitude 105.5, got %f", bbox.MinLongitude)
	}
	if bbox.MinLatitude != 20.5 {
		t.Errorf("expected MinLatitude 20.5, got %f", bbox.MinLatitude)
	}
	if bbox.MaxLongitude != 106.0 {
		t.Errorf("expected MaxLongitude 106.0, got %f", bbox.MaxLongitude)
	}
	if bbox.MaxLatitude != 21.0 {
		t.Errorf("expected MaxLatitude 21.0, got %f", bbox.MaxLatitude)
	}

	if center.Lat != 20.75 {
		t.Errorf("expected center.Lat 20.75, got %f", center.Lat)
	}
	if center.Lon != 105.75 {
		t.Errorf("expected center.Lon 105.75, got %f", center.Lon)
	}

	// Error: non-array
	_, _, err = parseBBox(json.RawMessage(`"not-array"`))
	if err == nil {
		t.Error("expected error for non-array input")
	}

	// Error: wrong length
	_, _, err = parseBBox(json.RawMessage(`[1, 2, 3]`))
	if err == nil {
		t.Error("expected error for wrong-length array")
	}
}

func TestGenerateSearchKeywords(t *testing.T) {
	keywords := file_writer_helper.GenerateSearchKeywords("01", "Hà Nội", "Hanoi", "ha_noi")
	if len(keywords) < 3 {
		t.Fatalf("expected at least 3 keywords, got %d: %v", len(keywords), keywords)
	}
	if keywords[0] != "01" {
		t.Errorf("expected first keyword '01', got %q", keywords[0])
	}
	// 2nd keyword should be tone-stripped lowercase name
	if keywords[1] != "ha noi" {
		t.Errorf("expected second keyword 'ha noi', got %q", keywords[1])
	}
	if keywords[2] != "hanoi" {
		t.Errorf("expected third keyword 'hanoi', got %q", keywords[2])
	}

	// Test deduplication: when tone-stripped name equals lowercase nameEn
	keywords = file_writer_helper.GenerateSearchKeywords("001", "ha noi", "ha noi", "ha_noi")
	seen := make(map[string]bool)
	for _, kw := range keywords {
		if seen[kw] {
			t.Errorf("duplicate keyword found: %q", kw)
		}
		seen[kw] = true
	}
}

func TestWriteChunkedNDJSON_SingleFile(t *testing.T) {
	// When total size is under maxNDJSONChunkSize, should produce a single file
	tmpDir := t.TempDir()
	ndjsonPath := filepath.Join(tmpDir, "provinces-gis.ndjson")

	docs := []dataset_file_writer_dto.ElasticsearchProvinceDocument{
		{Code: "01", Name: "Hà Nội"},
		{Code: "02", Name: "Hà Giang"},
	}

	err := writeChunkedNDJSON(ndjsonPath, esGISIndexName, docs)
	if err != nil {
		t.Fatalf("writeChunkedNDJSON failed: %v", err)
	}

	// Should produce a single file (no chunks, no manifest)
	data, err := os.ReadFile(ndjsonPath)
	if err != nil {
		t.Fatalf("failed to read ndjson: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("ndjson file is empty")
	}

	// Should NOT have chunk files or manifest
	manifestPath := ndjsonPath + ".manifest"
	if _, err := os.Stat(manifestPath); !os.IsNotExist(err) {
		t.Error("manifest file should not exist when total fits in one file")
	}

	// Verify content: 2 docs = 4 lines (action + doc per province)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 4 {
		t.Errorf("expected 4 lines (2 actions + 2 docs), got %d", len(lines))
	}
}

func TestWriteChunkedNDJSON_MultipleChunks(t *testing.T) {
	// When total size exceeds maxNDJSONChunkSize, should produce multiple chunk files + manifest
	tmpDir := t.TempDir()
	ndjsonPath := filepath.Join(tmpDir, "provinces-gis.ndjson")

	// Temporarily lower the chunk size so small test data triggers chunking
	originalMaxSize := maxNDJSONChunkSize
	maxNDJSONChunkSize = 100 // 100 bytes — forces chunking with any real data
	defer func() { maxNDJSONChunkSize = originalMaxSize }()

	// Create simple docs — with 100-byte limit, even small docs will be split
	docs := []dataset_file_writer_dto.ElasticsearchProvinceDocument{
		{Code: "01", Name: "Hà Nội"},
		{Code: "02", Name: "Hà Giang"},
		{Code: "03", Name: "Cao Bằng"},
	}

	err := writeChunkedNDJSON(ndjsonPath, esGISIndexName, docs)
	if err != nil {
		t.Fatalf("writeChunkedNDJSON failed: %v", err)
	}

	// Should NOT have a single ndjson file
	if _, err := os.Stat(ndjsonPath); !os.IsNotExist(err) {
		t.Error("single ndjson file should not exist when chunking is triggered")
	}

	// Should have at least 2 chunk files
	chunk1Path := filepath.Join(tmpDir, "provinces-gis-part-01.ndjson")
	if _, err := os.Stat(chunk1Path); os.IsNotExist(err) {
		t.Fatal("chunk file provinces-gis-part-01.ndjson not found")
	}

	chunk2Path := filepath.Join(tmpDir, "provinces-gis-part-02.ndjson")
	if _, err := os.Stat(chunk2Path); os.IsNotExist(err) {
		t.Fatal("chunk file provinces-gis-part-02.ndjson not found")
	}

	// Should have manifest file
	manifestPath := ndjsonPath + ".manifest"
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("failed to read manifest: %v", err)
	}
	manifestLines := strings.Split(strings.TrimSpace(string(manifestData)), "\n")
	if len(manifestLines) < 2 {
		t.Errorf("expected at least 2 lines in manifest, got %d", len(manifestLines))
	}

	// Collect all chunk files from the manifest
	var chunkFiles []string
	for _, name := range manifestLines {
		chunkFiles = append(chunkFiles, filepath.Join(tmpDir, name))
	}

	// Verify each chunk file has valid NDJSON (action + doc pairs)
	for _, chunkFile := range chunkFiles {
		data, err := os.ReadFile(chunkFile)
		if err != nil {
			t.Fatalf("failed to read chunk file %s: %v", chunkFile, err)
		}
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		if len(lines)%2 != 0 {
			t.Errorf("expected even number of lines (action+doc pairs), got %d in %s", len(lines), chunkFile)
		}
		// Verify first line is a valid index action
		var action map[string]interface{}
		if err := json.Unmarshal([]byte(lines[0]), &action); err != nil {
			t.Fatalf("first line of chunk is not valid JSON: %v", err)
		}
		indexAction, ok := action["index"].(map[string]interface{})
		if !ok {
			t.Fatal("first line of chunk missing 'index' key")
		}
		if indexAction["_index"] != "provinces-gis" {
			t.Errorf("expected _index 'provinces-gis', got %v", indexAction["_index"])
		}
	}

	// Verify all 3 docs are distributed across chunks
	totalDocs := 0
	for _, chunkFile := range chunkFiles {
		data, _ := os.ReadFile(chunkFile)
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		totalDocs += len(lines) / 2
	}
	if totalDocs != 3 {
		t.Errorf("expected 3 total docs across all chunks, got %d", totalDocs)
	}
}

func TestFilepathDir(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/tmp/output/file.ndjson", "/tmp/output"},
		{"output/file.ndjson", "output"},
		{"file.ndjson", "."},
		{"/file.ndjson", "/"},
	}
	for _, tt := range tests {
		got := filepathDir(tt.input)
		if got != tt.expected {
			t.Errorf("filepathDir(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFilepathBase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/tmp/output/file.ndjson", "file.ndjson"},
		{"output/file.ndjson", "file.ndjson"},
		{"file.ndjson", "file.ndjson"},
	}
	for _, tt := range tests {
		got := filepathBase(tt.input)
		if got != tt.expected {
			t.Errorf("filepathBase(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFilepathExt(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"file.ndjson", ".ndjson"},
		{"provinces-gis-part-01.ndjson", ".ndjson"},
		{"/tmp/output/file.json", ".json"},
		{"noext", ""},
		{"/path/file", ""},
	}
	for _, tt := range tests {
		got := filepathExt(tt.input)
		if got != tt.expected {
			t.Errorf("filepathExt(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestStringsJoin(t *testing.T) {
	tests := []struct {
		strs     []string
		sep      string
		expected string
	}{
		{[]string{"a", "b", "c"}, ",", "a,b,c"},
		{[]string{"a"}, ",", "a"},
		{[]string{}, ",", ""},
		{[]string{"part-01", "part-02"}, "\n", "part-01\npart-02"},
	}
	for _, tt := range tests {
		got := stringsJoin(tt.strs, tt.sep)
		if got != tt.expected {
			t.Errorf("stringsJoin(%v, %q) = %q, want %q", tt.strs, tt.sep, got, tt.expected)
		}
	}
}

func TestWriteToFile_MappingContainsPostalFields(t *testing.T) {
	tmpDir := t.TempDir()
	writer := ElasticsearchDatasetFileWriter{OutputFolderPath: tmpDir}
	err := writer.WriteToFile(nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("WriteToFile failed: %v", err)
	}
	mapping, err := os.ReadFile(filepath.Join(tmpDir, "mappings", "provinces.json"))
	if err != nil {
		t.Fatalf("read mapping: %v", err)
	}
	if !strings.Contains(string(mapping), "PostalCode") {
		t.Error("mapping missing PostalCode")
	}
	if !strings.Contains(string(mapping), "PostalCodePrefix") {
		t.Error("mapping missing PostalCodePrefix")
	}
}

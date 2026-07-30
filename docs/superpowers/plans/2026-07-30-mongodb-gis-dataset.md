# MongoDB GIS Dataset Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add MongoDB GIS dataset generation producing separate `provinces-gis` and `wards-gis` JSON files with native GeoJSON geometry and 2dsphere index scripts.

**Architecture:** New `WriteMongoGISDataToFile()` method on `MongoDBDatasetFileWriter` takes `SapNhapSiteGeoUnit` data, builds separate province and ward GIS documents with MongoDB-native GeoJSON Point centers, writes JSON files (chunked if >50MB), and generates index creation scripts. Follows the same pattern as `WriteElasticsearchGISDataToFile()`.

**Tech Stack:** Go 1.24, `encoding/json`, `bufio`, `os`, Bun ORM models, `SapNhapSiteGeoUnit` data source

## Global Constraints

- Go 1.24.0, module path `github.com/thanglequoc-vn-provinces/v2`
- PascalCase JSON field names (consistent with ES and existing MongoDB exports)
- MongoDB-native GeoJSON Point format for Center: `{ "type": "Point", "coordinates": [lon, lat] }`
- Output files kept under 50MB; chunk if exceeded
- Dataset version: `2026.07.01`, Administrative revision: `2026-04-30`
- All new files go under `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/`
- Tests use `package dataset_writer` (same package as code under test)
- Reuse `GenerateSearchKeywords()` from `helper/dto_mapper.go`
- Reuse `getFileTimeSuffix()`, `writeJSONFile()` from `dataset_file_writer.go`

---

## File Structure

| File | Responsibility |
|------|----------------|
| `dto/mongo_gis_dto.go` (NEW) | MongoDB GIS document types with JSON tags |
| `helper/mongo_gis_mapper.go` (NEW) | Conversion: SapNhapSiteGeoUnit → MongoGIS DTOs |
| `mongodb_gis_file_writer.go` (NEW) | `WriteMongoGISDataToFile()` method, chunking, index script, README |
| `mongodb_gis_file_writer_test.go` (NEW) | Unit tests for GIS writer |
| `helper/mongo_gis_mapper_test.go` (NEW) | Unit tests for mapper functions |
| `dataset_writer.go` (MODIFY) | Add MongoDB GIS writer call in `GenerateGISSQLDatasets()` |

---

### Task 1: MongoDB GIS DTO Types

**Files:**
- Create: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/dto/mongo_gis_dto.go`

**Interfaces:**
- Produces: `MongoGISProvinceDocument`, `MongoGISWardDocument`, `MongoGIS`, `MongoGeoPoint`, `MongoBoundingBox`, `MongoGISProperties`, `MongoAdministrativeUnit`, `MongoMeta` — all with JSON tags

- [ ] **Step 1: Create the DTO file**

```go
package dto

import "encoding/json"

// MongoGISProvinceDocument represents a province document in the
// provinces-gis MongoDB collection.
type MongoGISProvinceDocument struct {
	Code               string                 `json:"Code"`
	Name               string                 `json:"Name"`
	NameEn             string                 `json:"NameEn"`
	FullName           string                 `json:"FullName"`
	FullNameEn         string                 `json:"FullNameEn"`
	CodeName           string                 `json:"CodeName"`
	AdministrativeUnit MongoAdministrativeUnit `json:"AdministrativeUnit"`
	SearchKeywords     []string               `json:"SearchKeywords"`
	GIS                *MongoGIS              `json:"GIS,omitempty"`
	Meta               *MongoMeta             `json:"Meta,omitempty"`
}

// MongoGISWardDocument represents a ward document in the
// wards-gis MongoDB collection.
type MongoGISWardDocument struct {
	Code               string                 `json:"Code"`
	Name               string                 `json:"Name"`
	NameEn             string                 `json:"NameEn"`
	FullName           string                 `json:"FullName"`
	FullNameEn         string                 `json:"FullNameEn"`
	CodeName           string                 `json:"CodeName"`
	ProvinceCode       string                 `json:"ProvinceCode"`
	AdministrativeUnit MongoAdministrativeUnit `json:"AdministrativeUnit"`
	SearchKeywords     []string               `json:"SearchKeywords"`
	GIS                *MongoGIS              `json:"GIS,omitempty"`
	Meta               *MongoMeta             `json:"Meta,omitempty"`
}

// MongoGIS holds GIS data for the provinces-gis and wards-gis collections.
type MongoGIS struct {
	Center      MongoGeoPoint        `json:"Center"`
	BoundingBox MongoBoundingBox     `json:"BoundingBox"`
	Geometry    json.RawMessage      `json:"Geometry"`
	Properties  *MongoGISProperties  `json:"Properties,omitempty"`
}

// MongoGeoPoint is a MongoDB-native GeoJSON Point for 2dsphere indexing.
type MongoGeoPoint struct {
	Type        string     `json:"type"`
	Coordinates [2]float64 `json:"coordinates"` // [lon, lat]
}

// MongoBoundingBox holds the bounding box coordinates.
type MongoBoundingBox struct {
	MinLongitude float64 `json:"MinLongitude"`
	MinLatitude  float64 `json:"MinLatitude"`
	MaxLongitude float64 `json:"MaxLongitude"`
	MaxLatitude  float64 `json:"MaxLatitude"`
}

// MongoGISProperties holds administrative metadata inside the GIS object.
type MongoGISProperties struct {
	Code        string  `json:"Code"`
	Name        string  `json:"Name"`
	NameEn      string  `json:"NameEn"`
	FullName    string  `json:"FullName"`
	FullNameEn  string  `json:"FullNameEn"`
	CodeName    string  `json:"CodeName"`
	GisServerId string  `json:"GisServerId"`
	AreaKm2     float64 `json:"AreaKm2"`
}

// MongoAdministrativeUnit is the embedded administrative unit object.
type MongoAdministrativeUnit struct {
	Id          int    `json:"Id"`
	FullName    string `json:"FullName"`
	FullNameEn  string `json:"FullNameEn"`
	ShortName   string `json:"ShortName"`
	ShortNameEn string `json:"ShortNameEn"`
	CodeName    string `json:"CodeName"`
	CodeNameEn  string `json:"CodeNameEn"`
}

// MongoMeta holds dataset version metadata.
type MongoMeta struct {
	DatasetVersion         string `json:"DatasetVersion"`
	AdministrativeRevision string `json:"AdministrativeRevision"`
	GeneratedAt            string `json:"GeneratedAt"`
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd dataset-generation-scripts && go build ./internal/dataset_writer/dataset_file_writer/dto/`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/dto/mongo_gis_dto.go
git commit -m "feat: add MongoDB GIS document DTO types"
```

---

### Task 2: MongoDB GIS Mapper Functions

**Files:**
- Create: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/helper/mongo_gis_mapper.go`
- Test: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/helper/mongo_gis_mapper_test.go`

**Interfaces:**
- Consumes: `model.Province`, `model.Ward`, `model.AdministrativeUnit` from `vn_provinces_tmp/model`, `SapNhapSiteGeoUnit` from `sapnhap_bando/model`, `GenerateSearchKeywords()` from `helper/dto_mapper.go`, DTO types from `dto/mongo_gis_dto.go`
- Produces: `ConvertToMongoGISProvinceDocuments()`, `ConvertToMongoGISWardDocuments()`, `parseBBoxForMongo()`, `convertToMongoAdministrativeUnit()`

- [ ] **Step 1: Write the failing test for parseBBoxForMongo**

Create `helper/mongo_gis_mapper_test.go`:

```go
package helper

import (
	"encoding/json"
	"testing"

	dataset_file_writer_dto "github.com/thanglequoc-vn-provinces/v2/internal/dataset_writer/dataset_file_writer/dto"
)

func TestParseBBoxForMongo(t *testing.T) {
	bboxJSON := json.RawMessage(`[105.5, 20.5, 106.0, 21.0]`)
	bbox, center, err := parseBBoxForMongo(bboxJSON)
	if err != nil {
		t.Fatalf("parseBBoxForMongo failed: %v", err)
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

	// MongoDB GeoJSON Point: coordinates are [lon, lat]
	if center.Type != "Point" {
		t.Errorf("expected center.Type 'Point', got %q", center.Type)
	}
	if center.Coordinates[0] != 105.75 {
		t.Errorf("expected center.Coordinates[0] (lon) 105.75, got %f", center.Coordinates[0])
	}
	if center.Coordinates[1] != 20.75 {
		t.Errorf("expected center.Coordinates[1] (lat) 20.75, got %f", center.Coordinates[1])
	}

	// Error: non-array
	_, _, err = parseBBoxForMongo(json.RawMessage(`"not-array"`))
	if err == nil {
		t.Error("expected error for non-array input")
	}

	// Error: wrong length
	_, _, err = parseBBoxForMongo(json.RawMessage(`[1, 2, 3]`))
	if err == nil {
		t.Error("expected error for wrong-length array")
	}
}

func TestConvertToMongoAdministrativeUnit(t *testing.T) {
	au := model.AdministrativeUnit{
		Id: 1, FullName: "Thành phố", FullNameEn: "City",
		ShortName: "TP.", ShortNameEn: "City",
		CodeName: "thanh_pho", CodeNameEn: "city",
	}
	result := convertToMongoAdministrativeUnit(au)
	if result.Id != 1 {
		t.Errorf("expected Id 1, got %d", result.Id)
	}
	if result.FullName != "Thành phố" {
		t.Errorf("expected FullName 'Thành phố', got %q", result.FullName)
	}
	if result.CodeNameEn != "city" {
		t.Errorf("expected CodeNameEn 'city', got %q", result.CodeNameEn)
	}
}
```

Note: Add `"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"` to the test imports.

- [ ] **Step 2: Run test to verify it fails**

Run: `cd dataset-generation-scripts && go test ./internal/dataset_writer/dataset_file_writer/helper/ -run TestParseBBoxForMongo -v`
Expected: FAIL — `parseBBoxForMongo` not defined

- [ ] **Step 3: Write the mapper implementation**

Create `helper/mongo_gis_mapper.go`:

```go
package helper

import (
	"encoding/json"
	"fmt"

	dataset_file_writer_dto "github.com/thanglequoc-vn-provinces/v2/internal/dataset_writer/dataset_file_writer/dto"
	sapnhapbandomodel "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
	"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
)

// ConvertToMongoGISProvinceDocuments converts SapNhapSiteGeoUnit provinces
// to MongoDB GIS province documents.
func ConvertToMongoGISProvinceDocuments(
	geoProvinces []*sapnhapbandomodel.SapNhapSiteGeoUnit,
	datasetVersion, adminRevision, generatedAt string,
) []dataset_file_writer_dto.MongoGISProvinceDocument {
	var docs []dataset_file_writer_dto.MongoGISProvinceDocument
	for _, geoProvince := range geoProvinces {
		province := geoProvince.VNProvince
		doc := dataset_file_writer_dto.MongoGISProvinceDocument{
			Code:               province.Code,
			Name:               province.Name,
			NameEn:             province.NameEn,
			FullName:           province.FullName,
			FullNameEn:         province.FullNameEn,
			CodeName:           province.CodeName,
			AdministrativeUnit: convertToMongoAdministrativeUnit(province.AdministrativeUnit),
			SearchKeywords:     GenerateSearchKeywords(province.Code, province.Name, province.NameEn, province.CodeName),
			Meta: &dataset_file_writer_dto.MongoMeta{
				DatasetVersion:         datasetVersion,
				AdministrativeRevision: adminRevision,
				GeneratedAt:            generatedAt,
			},
		}

		// Add province GIS
		provinceProps := &dataset_file_writer_dto.MongoGISProperties{
			Code:        province.Code,
			Name:        province.Name,
			NameEn:      province.NameEn,
			FullName:    province.FullName,
			FullNameEn:  province.FullNameEn,
			CodeName:    province.CodeName,
			GisServerId: geoProvince.MaLK,
			AreaKm2:     geoProvince.DienTichKM2,
		}
		if gis, err := sapnhapGeoUnitToMongoGIS(*geoProvince, provinceProps); err == nil {
			doc.GIS = gis
		}

		docs = append(docs, doc)
	}
	return docs
}

// ConvertToMongoGISWardDocuments converts SapNhapSiteGeoUnit wards
// to MongoDB GIS ward documents.
func ConvertToMongoGISWardDocuments(
	geoWards []*sapnhapbandomodel.SapNhapSiteGeoUnit,
	datasetVersion, adminRevision, generatedAt string,
) []dataset_file_writer_dto.MongoGISWardDocument {
	var docs []dataset_file_writer_dto.MongoGISWardDocument
	for _, geoWard := range geoWards {
		ward := geoWard.VNWard
		doc := dataset_file_writer_dto.MongoGISWardDocument{
			Code:               ward.Code,
			Name:               ward.Name,
			NameEn:             ward.NameEn,
			FullName:           ward.FullName,
			FullNameEn:         ward.FullNameEn,
			CodeName:           ward.CodeName,
			ProvinceCode:       geoWard.VNDSProvinceCode,
			AdministrativeUnit: convertToMongoAdministrativeUnit(ward.AdministrativeUnit),
			SearchKeywords:     GenerateSearchKeywords(ward.Code, ward.Name, ward.NameEn, ward.CodeName),
			Meta: &dataset_file_writer_dto.MongoMeta{
				DatasetVersion:         datasetVersion,
				AdministrativeRevision: adminRevision,
				GeneratedAt:            generatedAt,
			},
		}

		// Add ward GIS
		wardProps := &dataset_file_writer_dto.MongoGISProperties{
			Code:        ward.Code,
			Name:        ward.Name,
			NameEn:      ward.NameEn,
			FullName:    ward.FullName,
			FullNameEn:  ward.FullNameEn,
			CodeName:    ward.CodeName,
			GisServerId: geoWard.MaLK,
			AreaKm2:     geoWard.DienTichKM2,
		}
		if gis, err := sapnhapGeoUnitToMongoGIS(*geoWard, wardProps); err == nil {
			doc.GIS = gis
		}

		docs = append(docs, doc)
	}
	return docs
}

// sapnhapGeoUnitToMongoGIS converts a SapNhapSiteGeoUnit's BBoxGeoJSON and
// GeomGeoJSON into a MongoGIS struct.
func sapnhapGeoUnitToMongoGIS(unit sapnhapbandomodel.SapNhapSiteGeoUnit, properties *dataset_file_writer_dto.MongoGISProperties) (*dataset_file_writer_dto.MongoGIS, error) {
	bbox, center, err := parseBBoxForMongo(unit.BBoxGeoJSON)
	if err != nil {
		return nil, err
	}
	return &dataset_file_writer_dto.MongoGIS{
		Properties:  properties,
		Center:      center,
		BoundingBox: bbox,
		Geometry:    unit.GeomGeoJSON,
	}, nil
}

// parseBBoxForMongo parses a BBoxGeoJSON array [xmin, ymin, xmax, ymax]
// into MongoBoundingBox and MongoGeoPoint (center as GeoJSON Point).
func parseBBoxForMongo(bboxGeoJSON json.RawMessage) (dataset_file_writer_dto.MongoBoundingBox, dataset_file_writer_dto.MongoGeoPoint, error) {
	var coords []float64
	if err := json.Unmarshal(bboxGeoJSON, &coords); err != nil {
		return dataset_file_writer_dto.MongoBoundingBox{}, dataset_file_writer_dto.MongoGeoPoint{}, fmt.Errorf("parse bbox geojson: %w", err)
	}
	if len(coords) != 4 {
		return dataset_file_writer_dto.MongoBoundingBox{}, dataset_file_writer_dto.MongoGeoPoint{}, fmt.Errorf("expected 4 bbox coordinates, got %d", len(coords))
	}
	xmin, ymin, xmax, ymax := coords[0], coords[1], coords[2], coords[3]
	bbox := dataset_file_writer_dto.MongoBoundingBox{
		MinLongitude: xmin,
		MinLatitude:  ymin,
		MaxLongitude: xmax,
		MaxLatitude:  ymax,
	}
	center := dataset_file_writer_dto.MongoGeoPoint{
		Type:        "Point",
		Coordinates: [2]float64{(xmin + xmax) / 2, (ymin + ymax) / 2}, // [lon, lat]
	}
	return bbox, center, nil
}

// convertToMongoAdministrativeUnit converts a model.AdministrativeUnit to
// the MongoDB DTO.
func convertToMongoAdministrativeUnit(au model.AdministrativeUnit) dataset_file_writer_dto.MongoAdministrativeUnit {
	return dataset_file_writer_dto.MongoAdministrativeUnit{
		Id:          au.Id,
		FullName:    au.FullName,
		FullNameEn:  au.FullNameEn,
		ShortName:   au.ShortName,
		ShortNameEn: au.ShortNameEn,
		CodeName:    au.CodeName,
		CodeNameEn:  au.CodeNameEn,
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd dataset-generation-scripts && go test ./internal/dataset_writer/dataset_file_writer/helper/ -run TestParseBBoxForMongo -v`
Expected: PASS

Run: `cd dataset-generation-scripts && go test ./internal/dataset_writer/dataset_file_writer/helper/ -run TestConvertToMongoAdministrativeUnit -v`
Expected: PASS

- [ ] **Step 5: Write tests for ConvertToMongoGISProvinceDocuments and ConvertToMongoGISWardDocuments**

Add to `helper/mongo_gis_mapper_test.go`:

```go
func TestConvertToMongoGISProvinceDocuments(t *testing.T) {
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
				AdministrativeUnit: model.AdministrativeUnit{
					Id: 1, FullName: "Thành phố", FullNameEn: "City",
					ShortName: "TP.", ShortNameEn: "City",
					CodeName: "thanh_pho", CodeNameEn: "city",
				},
			},
		},
	}

	docs := ConvertToMongoGISProvinceDocuments(geoProvinces, "2026.07.01", "2026-04-30", "2026-07-25T00:00:00Z")
	if len(docs) != 1 {
		t.Fatalf("expected 1 doc, got %d", len(docs))
	}
	doc := docs[0]
	if doc.Code != "01" {
		t.Errorf("expected Code '01', got %q", doc.Code)
	}
	if doc.GIS == nil {
		t.Fatal("expected GIS to be populated")
	}
	if doc.GIS.Center.Type != "Point" {
		t.Errorf("expected Center.Type 'Point', got %q", doc.GIS.Center.Type)
	}
	if doc.GIS.Properties == nil {
		t.Fatal("expected GIS Properties to be populated")
	}
	if doc.GIS.Properties.GisServerId != "gis_001" {
		t.Errorf("expected GisServerId 'gis_001', got %q", doc.GIS.Properties.GisServerId)
	}
	if doc.GIS.Properties.AreaKm2 != 3359.84 {
		t.Errorf("expected AreaKm2 3359.84, got %f", doc.GIS.Properties.AreaKm2)
	}
	if doc.Meta == nil {
		t.Fatal("expected Meta to be populated")
	}
	if doc.Meta.DatasetVersion != "2026.07.01" {
		t.Errorf("expected DatasetVersion '2026.07.01', got %q", doc.Meta.DatasetVersion)
	}
	if len(doc.SearchKeywords) == 0 {
		t.Error("expected SearchKeywords to be populated")
	}
}

func TestConvertToMongoGISWardDocuments(t *testing.T) {
	bboxJSON := json.RawMessage(`[105.5, 20.5, 106.0, 21.0]`)
	geomJSON := json.RawMessage(`{"type":"Polygon","coordinates":[]}`)

	geoWards := []*sapnhapbandomodel.SapNhapSiteGeoUnit{
		{
			Ma: "00001", MaLK: "gis_00001", Ten: "Ba Đình", VNDSProvinceCode: "01",
			DienTichKM2: 5.23,
			BBoxGeoJSON: bboxJSON, GeomGeoJSON: geomJSON,
			VNWard: model.Ward{
				Code: "00001", Name: "Ba Đình", NameEn: "Ba Dinh",
				FullName: "Phường Ba Đình", FullNameEn: "Ba Dinh Ward",
				CodeName: "ba_dinh",
				AdministrativeUnit: model.AdministrativeUnit{
					Id: 3, FullName: "Phường", FullNameEn: "Ward",
					ShortName: "P.", ShortNameEn: "Ward",
					CodeName: "phuong", CodeNameEn: "ward",
				},
			},
		},
	}

	docs := ConvertToMongoGISWardDocuments(geoWards, "2026.07.01", "2026-04-30", "2026-07-25T00:00:00Z")
	if len(docs) != 1 {
		t.Fatalf("expected 1 doc, got %d", len(docs))
	}
	doc := docs[0]
	if doc.Code != "00001" {
		t.Errorf("expected Code '00001', got %q", doc.Code)
	}
	if doc.ProvinceCode != "01" {
		t.Errorf("expected ProvinceCode '01', got %q", doc.ProvinceCode)
	}
	if doc.GIS == nil {
		t.Fatal("expected GIS to be populated")
	}
	if doc.GIS.Properties == nil {
		t.Fatal("expected GIS Properties to be populated")
	}
	if doc.GIS.Properties.GisServerId != "gis_00001" {
		t.Errorf("expected GisServerId 'gis_00001', got %q", doc.GIS.Properties.GisServerId)
	}
	if doc.GIS.Properties.AreaKm2 != 5.23 {
		t.Errorf("expected AreaKm2 5.23, got %f", doc.GIS.Properties.AreaKm2)
	}
}
```

Add `"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"` to the test imports.

- [ ] **Step 6: Run all mapper tests**

Run: `cd dataset-generation-scripts && go test ./internal/dataset_writer/dataset_file_writer/helper/ -run "TestConvertToMongo|TestParseBBoxForMongo" -v`
Expected: All PASS

- [ ] **Step 7: Commit**

```bash
git add dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/helper/mongo_gis_mapper.go \
       dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/helper/mongo_gis_mapper_test.go
git commit -m "feat: add MongoDB GIS mapper functions with tests"
```

---

### Task 3: MongoDB GIS File Writer

**Files:**
- Create: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/mongodb_gis_file_writer.go`
- Test: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/mongodb_gis_file_writer_test.go`

**Interfaces:**
- Consumes: `MongoDBDatasetFileWriter` struct (from `mongodb_file_writer.go`), `ConvertToMongoGISProvinceDocuments()` and `ConvertToMongoGISWardDocuments()` from `helper/mongo_gis_mapper.go`, `getFileTimeSuffix()` from `dataset_file_writer.go`, `SapNhapSiteGeoUnit` from `sapnhap_bando/model`
- Produces: `WriteMongoGISDataToFile()` method on `MongoDBDatasetFileWriter`

- [ ] **Step 1: Write the failing test for WriteMongoGISDataToFile**

Create `mongodb_gis_file_writer_test.go`:

```go
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

	// Verify create_indexes.js exists
	indexPath := filepath.Join(tmpDir, "create_indexes.js")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Fatal("create_indexes.js not found")
	}

	// Verify README.md exists
	readmePath := filepath.Join(tmpDir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		t.Fatal("README.md not found")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd dataset-generation-scripts && go test ./internal/dataset_writer/dataset_file_writer/ -run TestWriteMongoGISDataToFile -v`
Expected: FAIL — `WriteMongoGISDataToFile` not defined

- [ ] **Step 3: Write the GIS file writer implementation**

Create `mongodb_gis_file_writer.go`:

```go
package dataset_writer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	dataset_file_writer_dto "github.com/thanglequoc-vn-provinces/v2/internal/dataset_writer/dataset_file_writer/dto"
	file_writer_helper "github.com/thanglequoc-vn-provinces/v2/internal/dataset_writer/dataset_file_writer/helper"
	sapnhapbandomodel "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
)

const (
	mongoDatasetVer   = "2026.07.01"
	mongoAdminRev     = "2026-04-30"
	mongoGISChunkSize = 50 * 1024 * 1024 // 50 MB
)

// WriteMongoGISDataToFile generates the provinces-gis and wards-gis MongoDB
// JSON files, create_indexes.js, and README.md.
func (w *MongoDBDatasetFileWriter) WriteMongoGISDataToFile(
	sapNhapGeoProvinces []*sapnhapbandomodel.SapNhapSiteGeoUnit,
	sapNhapGeoWards []*sapnhapbandomodel.SapNhapSiteGeoUnit) error {

	os.MkdirAll(w.OutputFolderPath, 0746)
	fileTimeSuffix := getFileTimeSuffix()
	generatedAt := time.Now().UTC().Format(time.RFC3339)

	// Build province GIS documents
	provinceDocs := file_writer_helper.ConvertToMongoGISProvinceDocuments(
		sapNhapGeoProvinces, mongoDatasetVer, mongoAdminRev, generatedAt)

	// Build ward GIS documents
	wardDocs := file_writer_helper.ConvertToMongoGISWardDocuments(
		sapNhapGeoWards, mongoDatasetVer, mongoAdminRev, generatedAt)

	// Write province GIS file (chunked if needed)
	provincePath := fmt.Sprintf("%s/mongo_data_vn_province_gis_%s.json", w.OutputFolderPath, fileTimeSuffix)
	if err := writeChunkedMongoJSON(provincePath, provinceDocs); err != nil {
		return fmt.Errorf("write province GIS json: %w", err)
	}

	// Write ward GIS file (chunked if needed)
	wardPath := fmt.Sprintf("%s/mongo_data_vn_ward_gis_%s.json", w.OutputFolderPath, fileTimeSuffix)
	if err := writeChunkedMongoJSON(wardPath, wardDocs); err != nil {
		return fmt.Errorf("write ward GIS json: %w", err)
	}

	// Write create_indexes.js
	indexPath := fmt.Sprintf("%s/create_indexes.js", w.OutputFolderPath)
	if err := writeMongoGISIndexScript(indexPath); err != nil {
		return fmt.Errorf("write create_indexes.js: %w", err)
	}

	// Write README.md
	readmePath := fmt.Sprintf("%s/README.md", w.OutputFolderPath)
	if err := writeMongoGISReadme(readmePath); err != nil {
		return fmt.Errorf("write README: %w", err)
	}

	return nil
}

// writeChunkedMongoJSON writes a JSON array of documents, splitting into
// multiple chunk files if the total size exceeds mongoGISChunkSize (50MB).
// If the total fits in one file, writes a single file at path.
// If chunking is needed, files are written as path with numeric suffix:
// e.g. mongo_data_vn_ward_gis_2026..._part_01.json, _part_02.json, etc.
// A manifest file (path + ".manifest") lists the chunk filenames in order.
func writeChunkedMongoJSON(path string, docs interface{}) error {
	// Pre-serialize to determine total size
	data, err := json.MarshalIndent(docs, "", " ")
	if err != nil {
		return fmt.Errorf("marshal docs: %w", err)
	}

	// If total fits in one file, write directly
	if len(data) <= mongoGISChunkSize {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer file.Close()
		writer := bufio.NewWriter(file)
		writer.Write(data)
		return writer.Flush()
	}

	// Need to chunk — re-marshal each doc individually
	docSlice := reflectValueOf(docs)
	dir := filepathDir(path)
	base := filepathBase(path)
	ext := filepathExt(base)
	nameNoExt := base[:len(base)-len(ext)]

	// Marshal each doc individually to get sizes
	type serializedDoc struct {
		bytes []byte
		size  int
	}
	var serialized []serializedDoc
	totalSize := 0
	for i := 0; i < docSlice.Len(); i++ {
		docBytes, err := json.MarshalIndent(docSlice.Index(i).Interface(), "", " ")
		if err != nil {
			return fmt.Errorf("marshal doc %d: %w", i, err)
		}
		serialized = append(serialized, serializedDoc{docBytes, len(docBytes)})
		totalSize += len(docBytes)
	}

	// Split into chunks
	var chunks [][]serializedDoc
	currentChunk := []serializedDoc{}
	currentSize := 0
	for _, s := range serialized {
		if currentSize+s.size > mongoGISChunkSize && len(currentChunk) > 0 {
			chunks = append(chunks, currentChunk)
			currentChunk = []serializedDoc{}
			currentSize = 0
		}
		currentChunk = append(currentChunk, s)
		currentSize += s.size
	}
	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	log.Printf("📦 [MongoDB] GIS JSON chunked: %d docs → %d files (max %d MB each)",
		len(serialized), len(chunks), mongoGISChunkSize/1024/1024)

	// Write each chunk file as a JSON array
	var chunkNames []string
	for i, chunk := range chunks {
		chunkName := fmt.Sprintf("%s_part_%02d%s", nameNoExt, i+1, ext)
		chunkPath := fmt.Sprintf("%s/%s", dir, chunkName)
		chunkNames = append(chunkNames, chunkName)

		file, err := os.OpenFile(chunkPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("create chunk file %s: %w", chunkPath, err)
		}
		writer := bufio.NewWriter(file)
		writer.WriteByte('[')
		for j, s := range chunk {
			if j > 0 {
				writer.WriteByte(',')
			}
			writer.Write(s.bytes)
		}
		writer.WriteByte(']')
		if err := writer.Flush(); err != nil {
			file.Close()
			return fmt.Errorf("flush chunk file %s: %w", chunkPath, err)
		}
		file.Close()

		chunkSize := 0
		for _, s := range chunk {
			chunkSize += s.size
		}
		log.Printf("   %s: %.1f MB, %d docs", chunkName, float64(chunkSize)/1024/1024, len(chunk))
	}

	// Write manifest file
	manifestPath := path + ".manifest"
	manifestContent := stringsJoin(chunkNames, "\n") + "\n"
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		return fmt.Errorf("write manifest file: %w", err)
	}
	log.Printf("   Manifest: %s", filepathBase(manifestPath))

	return nil
}

// writeMongoGISIndexScript writes the create_indexes.js file.
func writeMongoGISIndexScript(path string) error {
	content := `// MongoDB GIS Index Creation Script
// Run with: mongo < create_indexes.js
// or: mongosh create_indexes.js

// provinces-gis collection indexes
db.provinces_gis.createIndex({ "Code": 1 }, { unique: true });
db.provinces_gis.createIndex({ "GIS.Geometry": "2dsphere" });
db.provinces_gis.createIndex({ "GIS.Center": "2dsphere" });
db.provinces_gis.createIndex({ "SearchKeywords": 1 });

// wards-gis collection indexes
db.wards_gis.createIndex({ "Code": 1 }, { unique: true });
db.wards_gis.createIndex({ "ProvinceCode": 1 });
db.wards_gis.createIndex({ "GIS.Geometry": "2dsphere" });
db.wards_gis.createIndex({ "GIS.Center": "2dsphere" });
db.wards_gis.createIndex({ "SearchKeywords": 1 });
`
	return os.WriteFile(path, []byte(content), 0644)
}

// writeMongoGISReadme writes the README.md for the MongoDB GIS dataset.
func writeMongoGISReadme(path string) error {
	loc, err := time.LoadLocation("Asia/Saigon")
	if err != nil {
		loc = time.FixedZone("GMT+7", 7*60*60)
	}
	createdAt := time.Now().In(loc).Format(time.RFC1123Z)

	content := `# Vietnamese Provinces Database — MongoDB GIS Dataset

Created at:  ` + createdAt + `

## Overview

This dataset provides Vietnamese provinces and wards in MongoDB document format
with two GIS collections:

| Collection | Documents | Description |
|------------|-----------|-------------|
| ` + "`provinces-gis`" + ` | 34 | Province documents with GIS geometry (bounding boxes + GeoJSON polygons) |
| ` + "`wards-gis`" + ` | 3,321 | Standalone ward documents with GIS geometry + ProvinceCode reference |

## Document Structure

### Province GIS Document

- **Core fields**: Code, Name, NameEn, FullName, FullNameEn, CodeName
- **` + "`AdministrativeUnit`" + `**: Embedded administrative unit object
- **` + "`SearchKeywords`" + `**: Pre-computed autocomplete keywords
- **` + "`GIS`" + `**: Center (GeoJSON Point), BoundingBox, Geometry (GeoJSON MultiPolygon), Properties
- **` + "`Meta`" + `**: Dataset version metadata

### Ward GIS Document

- Same structure as province, plus **` + "`ProvinceCode`" + `** for cross-collection joins

## Quick Start

### 1. Import the Data

` + "```bash" + `
# Import province GIS data
mongoimport --db vn_provinces --collection provinces_gis --file mongo_data_vn_province_gis_*.json --jsonArray

# Import ward GIS data (may be chunked)
mongoimport --db vn_provinces --collection wards_gis --file mongo_data_vn_ward_gis_*.json --jsonArray
` + "```" + `

### 2. Create Indexes

` + "```bash" + `
mongo vn_provinces create_indexes.js
` + "```" + `

### 3. Example Queries

` + "```javascript" + `
// Find province containing a point
db.provinces_gis.findOne({
  GIS.Geometry: {
    $geoIntersects: {
      $geometry: { type: "Point", coordinates: [105.8542, 21.0285] }
    }
  }
})

// Find all wards in a province
db.wards_gis.find({ ProvinceCode: "01" })

// Find ward containing a point
db.wards_gis.findOne({
  GIS.Geometry: {
    $geoIntersects: {
      $geometry: { type: "Point", coordinates: [105.8231, 21.0347] }
    }
  }
})
` + "```" + `

## File Listing

| File | Description |
|------|-------------|
| ` + "`mongo_data_vn_province_gis_*.json`" + ` | Province GIS documents (JSON array) |
| ` + "`mongo_data_vn_ward_gis_*.json`" + ` | Ward GIS documents (JSON array, may be chunked) |
| ` + "`create_indexes.js`" + ` | Index creation script for both collections |
`
	return os.WriteFile(path, []byte(content), 0644)
}
```

Note: The `writeChunkedMongoJSON` function uses `reflectValueOf` to handle both `[]MongoGISProvinceDocument` and `[]MongoGISWardDocument`. Add `"reflect"` to imports and this helper:

```go
// reflectValueOf returns the reflect.Value of a slice's elements.
func reflectValueOf(v interface{}) reflect.Value {
	return reflect.ValueOf(v)
}
```

Actually, to avoid reflection complexity, let's use a simpler approach — make `writeChunkedMongoJSON` accept `[]byte` pre-serialized data and handle chunking at the caller level. But that's more complex. The simplest approach: use `json.Marshal` on the whole slice, and if it exceeds the limit, use `reflect` to iterate. Let me revise to use a cleaner approach with an interface for chunkable documents.

Actually, the simplest and cleanest approach is to make the function generic using `any` and `reflect`. Let me finalize the implementation with proper imports.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd dataset-generation-scripts && go test ./internal/dataset_writer/dataset_file_writer/ -run TestWriteMongoGISDataToFile -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/mongodb_gis_file_writer.go \
       dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/mongodb_gis_file_writer_test.go
git commit -m "feat: add MongoDB GIS file writer with chunking support"
```

---

### Task 4: Integrate MongoDB GIS Writer into Dataset Generation Pipeline

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_writer.go`

**Interfaces:**
- Consumes: `MongoDBDatasetFileWriter.WriteMongoGISDataToFile()` from Task 3, `sapNhapGeoProvinces` and `sapNhapGeoWards` already fetched in `GenerateGISSQLDatasets()`

- [ ] **Step 1: Add MongoDB GIS writer call to GenerateGISSQLDatasets**

In `dataset_writer.go`, add after the Elasticsearch GIS block (after line 171):

```go
	// MongoDB GIS
	mongoDBGISFileWriter := datasetfilewriter.MongoDBDatasetFileWriter{
		OutputFolderPath: "./output/mongodb",
	}
	err = mongoDBGISFileWriter.WriteMongoGISDataToFile(sapNhapGeoProvinces, sapNhapGeoWards)
	if err != nil {
		log.Fatal("Unable to generate MongoDB GIS Dataset", err)
	} else {
		fmt.Println("✅ MongoDB GIS Dataset successfully generated")
	}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd dataset-generation-scripts && go build ./internal/dataset_writer/`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add dataset-generation-scripts/internal/dataset_writer/dataset_writer.go
git commit -m "feat: integrate MongoDB GIS writer into generation pipeline"
```

---

### Task 5: Run Tests and Verify

- [ ] **Step 1: Run all tests**

Run: `cd dataset-generation-scripts && go test ./internal/dataset_writer/... -v`
Expected: All tests PASS (including existing tests — no regressions)

- [ ] **Step 2: Verify no existing tests broke**

Check output for any FAIL results. If any existing tests fail, investigate and fix.

- [ ] **Step 3: Commit (if any fixes were needed)**

```bash
git add -A
git commit -m "test: verify all dataset writer tests pass"
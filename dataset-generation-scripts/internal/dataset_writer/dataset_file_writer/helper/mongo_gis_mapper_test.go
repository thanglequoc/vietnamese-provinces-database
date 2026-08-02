package helper

import (
	"encoding/json"
	"testing"

	sapnhapbandomodel "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
	"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
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
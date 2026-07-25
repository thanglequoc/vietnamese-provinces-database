# Elasticsearch GIS Properties Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `Properties` object to the `GIS` field in `provinces-gis` ES documents (province and ward level) containing Code, Name, NameEn, FullName, FullNameEn, CodeName, GisServerId, and AreaKm2.

**Architecture:** Add `ElasticsearchGISProperties` DTO, wire it through `sapnhapGeoUnitToESGIS` with a new parameter, populate from existing `SapNhapSiteGeoUnit` data (`Ma`, `DienTichKM2`, and related `VNProvince`/`VNWard`), update mapping generator, static mapping JSON, and README example.

**Tech Stack:** Go 1.24, Bun ORM, encoding/json, Elasticsearch 7.x+ mappings

## Global Constraints

- Geometry field stays pure geo_shape (type + coordinates) — Properties is a sibling under GIS, not inside Geometry
- PascalCase JSON field names (consistent with existing ES document conventions)
- Both province-level and ward-level GIS objects get Properties
- `omitempty` on the pointer so non-GIS documents don't emit null Properties
- Static mapping file and generated mapping must stay in sync
- All existing tests must continue to pass

---

### Task 1: Add DTO and update ElasticsearchGIS struct

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/dto/elasticsearch_dto.go`

**Interfaces:**
- Consumes: nothing new
- Produces: `ElasticsearchGISProperties` struct, updated `ElasticsearchGIS` with `Properties *ElasticsearchGISProperties` field

- [ ] **Step 1: Add the new struct and modify ElasticsearchGIS**

Add after the `ElasticsearchBoundingBox` struct (after line 64):

```go
// ElasticsearchGISProperties holds administrative metadata inside the GIS object
// for the provinces-gis index.
type ElasticsearchGISProperties struct {
	Code        string  `json:"Code"`
	Name        string  `json:"Name"`
	NameEn      string  `json:"NameEn"`
	FullName    string  `json:"FullName"`
	FullNameEn  string  `json:"FullNameEn"`
	CodeName    string  `json:"CodeName"`
	GisServerId string  `json:"GisServerId"`
	AreaKm2     float64 `json:"AreaKm2"`
}
```

Modify `ElasticsearchGIS` (line 46-50) to add the `Properties` field:

```go
type ElasticsearchGIS struct {
	Center      ElasticsearchGeoPoint        `json:"Center"`
	BoundingBox ElasticsearchBoundingBox      `json:"BoundingBox"`
	Geometry    json.RawMessage              `json:"Geometry"`
	Properties  *ElasticsearchGISProperties  `json:"Properties,omitempty"`
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd dataset-generation-scripts && go build ./internal/dataset_writer/dataset_file_writer/dto/`
Expected: exit code 0

- [ ] **Step 3: Commit**

```bash
git add dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/dto/elasticsearch_dto.go
git commit -m "feat: add ElasticsearchGISProperties DTO and Properties field to GIS"
```

---

### Task 2: Update sapnhapGeoUnitToESGIS to accept Properties

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer.go` — `sapnhapGeoUnitToESGIS` function (~line 384)

**Interfaces:**
- Consumes: `ElasticsearchGISProperties` from Task 1
- Produces: Updated `sapnhapGeoUnitToESGIS` with new signature `func sapnhapGeoUnitToESGIS(unit SapNhapSiteGeoUnit, properties *ElasticsearchGISProperties) (*ElasticsearchGIS, error)`

- [ ] **Step 1: Change the function signature and body**

Find the `sapnhapGeoUnitToESGIS` function (currently around line 384):

```go
func sapnhapGeoUnitToESGIS(unit sapnhapbandomodel.SapNhapSiteGeoUnit) (*dataset_file_writer_dto.ElasticsearchGIS, error) {
	bbox, center, err := parseBBox(unit.BBoxGeoJSON)
	if err != nil {
		return nil, err
	}
	return &dataset_file_writer_dto.ElasticsearchGIS{
		Center:      center,
		BoundingBox: bbox,
		Geometry:    unit.GeomGeoJSON,
	}, nil
}
```

Replace with:

```go
func sapnhapGeoUnitToESGIS(unit sapnhapbandomodel.SapNhapSiteGeoUnit, properties *dataset_file_writer_dto.ElasticsearchGISProperties) (*dataset_file_writer_dto.ElasticsearchGIS, error) {
	bbox, center, err := parseBBox(unit.BBoxGeoJSON)
	if err != nil {
		return nil, err
	}
	return &dataset_file_writer_dto.ElasticsearchGIS{
		Center:      center,
		BoundingBox: bbox,
		Geometry:    unit.GeomGeoJSON,
		Properties:  properties,
	}, nil
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd dataset-generation-scripts && go build ./internal/dataset_writer/dataset_file_writer/ 2>&1`
Expected: compilation errors at the call sites (line 122 and 140) — "not enough arguments in call to sapnhapGeoUnitToESGIS". This is expected and will be fixed in Task 3.

- [ ] **Step 3: Commit**

```bash
git add dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer.go
git commit -m "feat: add properties parameter to sapnhapGeoUnitToESGIS"
```

---

### Task 3: Build Properties at call sites in WriteElasticsearchGISDataToFile

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer.go:121-143`

**Interfaces:**
- Consumes: `sapnhapGeoUnitToESGIS` new signature from Task 2, `ElasticsearchGISProperties` from Task 1
- Produces: Complete province and ward GIS objects with Properties populated

- [ ] **Step 1: Update the province GIS call (around line 121-124)**

Replace:
```go
		// Add province GIS
		if gis, err := sapnhapGeoUnitToESGIS(*geoProvince); err == nil {
			doc.GIS = gis
		}
```

With:
```go
		// Add province GIS with Properties
		provinceProps := &dataset_file_writer_dto.ElasticsearchGISProperties{
			Code:        province.Code,
			Name:        province.Name,
			NameEn:      province.NameEn,
			FullName:    province.FullName,
			FullNameEn:  province.FullNameEn,
			CodeName:    province.CodeName,
			GisServerId: geoProvince.Ma,
			AreaKm2:     geoProvince.DienTichKM2,
		}
		if gis, err := sapnhapGeoUnitToESGIS(*geoProvince, provinceProps); err == nil {
			doc.GIS = gis
		}
```

- [ ] **Step 2: Update the ward GIS call (around line 140-142)**

Replace:
```go
			if gis, err := sapnhapGeoUnitToESGIS(*geoWard); err == nil {
				wardDoc.GIS = gis
			}
```

With:
```go
			wardProps := &dataset_file_writer_dto.ElasticsearchGISProperties{
				Code:        ward.Code,
				Name:        ward.Name,
				NameEn:      ward.NameEn,
				FullName:    ward.FullName,
				FullNameEn:  ward.FullNameEn,
				CodeName:    ward.CodeName,
				GisServerId: geoWard.Ma,
				AreaKm2:     geoWard.DienTichKM2,
			}
			if gis, err := sapnhapGeoUnitToESGIS(*geoWard, wardProps); err == nil {
				wardDoc.GIS = gis
			}
```

- [ ] **Step 3: Verify compilation**

Run: `cd dataset-generation-scripts && go build ./internal/dataset_writer/dataset_file_writer/`
Expected: exit code 0

- [ ] **Step 4: Run tests to verify nothing is broken (tests will need update, but GIS test should still pass structurally)**

Run: `cd dataset-generation-scripts && go test ./internal/dataset_writer/dataset_file_writer/ -v 2>&1`
Expected: Tests pass (GIS test verifies Center/BoundingBox exist; Properties is additive so no assertion fails yet)

- [ ] **Step 5: Commit**

```bash
git add dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer.go
git commit -m "feat: populate GIS Properties for province and ward in provinces-gis"
```

---

### Task 4: Update mapping generator (writeProvincesGISMapping)

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer.go` — `writeProvincesGISMapping` function (around line 552-567, and line 596-611)

**Interfaces:**
- Consumes: nothing new
- Produces: Updated generated `provinces-gis.json` mapping with Properties sub-fields under both `GIS` and `Wards.GIS`

- [ ] **Step 1: Add Properties to province-level GIS mapping (insert after line 565 `"Geometry": map[string]string{"type": "geo_shape"},`)**

Replace lines 552-567:
```go
				"GIS": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"Center":      map[string]string{"type": "geo_point"},
						"BoundingBox": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"MinLongitude": map[string]string{"type": "float"},
								"MinLatitude":  map[string]string{"type": "float"},
								"MaxLongitude": map[string]string{"type": "float"},
								"MaxLatitude":  map[string]string{"type": "float"},
							},
						},
						"Geometry": map[string]string{"type": "geo_shape"},
					},
				},
```

With:
```go
				"GIS": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"Center":      map[string]string{"type": "geo_point"},
						"BoundingBox": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"MinLongitude": map[string]string{"type": "float"},
								"MinLatitude":  map[string]string{"type": "float"},
								"MaxLongitude": map[string]string{"type": "float"},
								"MaxLatitude":  map[string]string{"type": "float"},
							},
						},
						"Geometry": map[string]string{"type": "geo_shape"},
						"Properties": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"Code":        map[string]string{"type": "keyword"},
								"Name":        map[string]string{"type": "keyword"},
								"NameEn":      map[string]string{"type": "keyword"},
								"FullName":    map[string]string{"type": "keyword"},
								"FullNameEn":  map[string]string{"type": "keyword"},
								"CodeName":    map[string]string{"type": "keyword"},
								"GisServerId": map[string]string{"type": "keyword"},
								"AreaKm2":     map[string]string{"type": "float"},
							},
						},
					},
				},
```

- [ ] **Step 2: Add Properties to ward-level GIS mapping (same change in Wards.GIS, around lines 596-611)**

Find the `"GIS"` entry under `"Wards"` (around line 596) and apply the same change — add `"Properties": {...}` after the `"Geometry"` entry.

The ward GIS section becomes:
```go
						"GIS": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"Center":      map[string]string{"type": "geo_point"},
								"BoundingBox": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"MinLongitude": map[string]string{"type": "float"},
										"MinLatitude":  map[string]string{"type": "float"},
										"MaxLongitude": map[string]string{"type": "float"},
										"MaxLatitude":  map[string]string{"type": "float"},
									},
								},
								"Geometry": map[string]string{"type": "geo_shape"},
								"Properties": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"Code":        map[string]string{"type": "keyword"},
										"Name":        map[string]string{"type": "keyword"},
										"NameEn":      map[string]string{"type": "keyword"},
										"FullName":    map[string]string{"type": "keyword"},
										"FullNameEn":  map[string]string{"type": "keyword"},
										"CodeName":    map[string]string{"type": "keyword"},
										"GisServerId": map[string]string{"type": "keyword"},
										"AreaKm2":     map[string]string{"type": "float"},
									},
								},
							},
						},
```

- [ ] **Step 3: Verify compilation**

Run: `cd dataset-generation-scripts && go build ./internal/dataset_writer/dataset_file_writer/`
Expected: exit code 0

- [ ] **Step 4: Run all tests**

Run: `cd dataset-generation-scripts && go test ./internal/dataset_writer/dataset_file_writer/ -v 2>&1`
Expected: all tests pass

- [ ] **Step 5: Commit**

```bash
git add dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer.go
git commit -m "feat: add GIS Properties to generated mapping for provinces-gis"
```

---

### Task 5: Update static mapping file

**Files:**
- Modify: `elasticsearch/mappings/provinces-gis.json`

**Interfaces:**
- Consumes: nothing
- Produces: Updated static mapping JSON with Properties sub-fields

- [ ] **Step 1: Add Properties at province-level GIS (line 65-67)**

After the `"Geometry"` entry (line 66), insert on line 67:
```json
          "Properties": {
            "properties": {
              "AreaKm2": {
                "type": "float"
              },
              "Code": {
                "type": "keyword"
              },
              "CodeName": {
                "type": "keyword"
              },
              "FullName": {
                "type": "keyword"
              },
              "FullNameEn": {
                "type": "keyword"
              },
              "GisServerId": {
                "type": "keyword"
              },
              "Name": {
                "type": "keyword"
              },
              "NameEn": {
                "type": "keyword"
              }
            },
            "type": "object"
          },
```

The complete province GIS block (lines 43-70) becomes:
```json
      "GIS": {
        "properties": {
          "BoundingBox": {
            "properties": {
              "MaxLatitude": {
                "type": "float"
              },
              "MaxLongitude": {
                "type": "float"
              },
              "MinLatitude": {
                "type": "float"
              },
              "MinLongitude": {
                "type": "float"
              }
            },
            "type": "object"
          },
          "Center": {
            "type": "geo_point"
          },
          "Geometry": {
            "type": "geo_shape"
          },
          "Properties": {
            "properties": {
              "AreaKm2": {
                "type": "float"
              },
              "Code": {
                "type": "keyword"
              },
              "CodeName": {
                "type": "keyword"
              },
              "FullName": {
                "type": "keyword"
              },
              "FullNameEn": {
                "type": "keyword"
              },
              "GisServerId": {
                "type": "keyword"
              },
              "Name": {
                "type": "keyword"
              },
              "NameEn": {
                "type": "keyword"
              }
            },
            "type": "object"
          }
        },
        "type": "object"
      },
```

- [ ] **Step 2: Add Properties at ward-level GIS (lines 144-171)**

The ward-level GIS block (lines 144-171) currently is:
```json
          "GIS": {
            "properties": {
              "BoundingBox": {
                "properties": {
                  "MaxLatitude": {
                    "type": "float"
                  },
                  "MaxLongitude": {
                    "type": "float"
                  },
                  "MinLatitude": {
                    "type": "float"
                  },
                  "MinLongitude": {
                    "type": "float"
                  }
                },
                "type": "object"
              },
              "Center": {
                "type": "geo_point"
              },
              "Geometry": {
                "type": "geo_shape"
              }
            },
            "type": "object"
          },
```

Add the same `"Properties"` block after `"Geometry"` entry (after line 167). The complete ward-level GIS block becomes:
```json
          "GIS": {
            "properties": {
              "BoundingBox": {
                "properties": {
                  "MaxLatitude": {
                    "type": "float"
                  },
                  "MaxLongitude": {
                    "type": "float"
                  },
                  "MinLatitude": {
                    "type": "float"
                  },
                  "MinLongitude": {
                    "type": "float"
                  }
                },
                "type": "object"
              },
              "Center": {
                "type": "geo_point"
              },
              "Geometry": {
                "type": "geo_shape"
              },
              "Properties": {
                "properties": {
                  "AreaKm2": {
                    "type": "float"
                  },
                  "Code": {
                    "type": "keyword"
                  },
                  "CodeName": {
                    "type": "keyword"
                  },
                  "FullName": {
                    "type": "keyword"
                  },
                  "FullNameEn": {
                    "type": "keyword"
                  },
                  "GisServerId": {
                    "type": "keyword"
                  },
                  "Name": {
                    "type": "keyword"
                  },
                  "NameEn": {
                    "type": "keyword"
                  }
                },
                "type": "object"
              }
            },
            "type": "object"
          },
```

- [ ] **Step 3: Validate JSON**

Run: `cd /Users/thanglequoc/projects/GitHub/vietnamese-provinces-database && python3 -m json.tool elasticsearch/mappings/provinces-gis.json > /dev/null && echo "Valid JSON"`
Expected: `Valid JSON`

- [ ] **Step 4: Commit**

```bash
git add elasticsearch/mappings/provinces-gis.json
git commit -m "feat: add GIS Properties to static provinces-gis mapping"
```

---

### Task 6: Update README example

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer.go` — `writeESReadme` function, the provinces-gis example section

**Interfaces:**
- Consumes: nothing new
- Produces: Updated README with Properties shown in the provinces-gis sample document

- [ ] **Step 1: Update the provinces-gis example in writeESReadme (around lines 728-744)**

The current provinces-gis example section (after the non-GIS example) shows a minimal GIS object. Replace it with a full document example including Properties.

Find this section in `writeESReadme`:
```
The ` + "`provinces-gis`" + ` index extends this same structure with a ` + "`GIS`" + ` object at both the province and ward level:

` + "```json" + `
{
  "Code": "01",
  "Name": "Hà Nội",
  "FullName": "Thành phố Hà Nội",
  ...
```

Replace the entire provinces-gis example block (from "The `provinces-gis` index extends..." through the closing ```) with:

```
The ` + "`provinces-gis`" + ` index extends this same structure with a ` + "`GIS`" + ` object at both the province and ward level:

` + "```json" + `
{
  "Code": "01",
  "Name": "Hà Nội",
  "FullName": "Thành phố Hà Nội",
  "CodeName": "ha_noi",
  "AdministrativeUnit": {
    "Id": 1,
    "FullName": "Thành phố trực thuộc trung ương",
    "ShortName": "Thành phố"
  },
  "SearchKeywords": ["01", "ha noi", "hanoi", "ha_noi"],
  "GIS": {
    "Center": { "Lat": 21.0285, "Lon": 105.8542 },
    "BoundingBox": {
      "MinLongitude": 105.2859,
      "MinLatitude": 20.4863,
      "MaxLongitude": 106.0617,
      "MaxLatitude": 21.3851
    },
    "Geometry": {
      "type": "MultiPolygon",
      "coordinates": [[[[105.2859, 21.3851], [106.0617, 21.3851], ...]]]
    },
    "Properties": {
      "Code": "01",
      "Name": "Hà Nội",
      "NameEn": "Hanoi",
      "FullName": "Thành phố Hà Nội",
      "FullNameEn": "Hanoi City",
      "CodeName": "ha_noi",
      "GisServerId": "diaphanhanhchinhcaptinh_sn.108",
      "AreaKm2": 3359.84
    }
  },
  "Wards": [
    {
      "Code": "00004",
      "Name": "Ba Đình",
      "FullName": "Phường Ba Đình",
      "CodeName": "ba_dinh",
      "AdministrativeUnit": { "Id": 3, "ShortName": "Phường" },
      "SearchKeywords": ["00004", "ba dinh", "ba_dinh"],
      "GIS": {
        "Center": { "Lat": 21.0347, "Lon": 105.8231 },
        "BoundingBox": {
          "MinLongitude": 105.8115, "MinLatitude": 21.0261,
          "MaxLongitude": 105.8347, "MaxLatitude": 21.0433
        },
        "Geometry": { "type": "Polygon", "coordinates": [[[105.8115, 21.0433], ...]] },
        "Properties": {
          "Code": "00004",
          "Name": "Ba Đình",
          "NameEn": "Ba Dinh",
          "FullName": "Phường Ba Đình",
          "FullNameEn": "Ba Dinh Ward",
          "CodeName": "ba_dinh",
          "GisServerId": "diaphanhanhchinhphuong_sn.456",
          "AreaKm2": 5.23
        }
      }
    }
  ],
  "Meta": {
    "DatasetVersion": "2026.07.01",
    "AdministrativeRevision": "2026-04-30",
    "GeneratedAt": "2026-07-25T03:00:43Z"
  }
}
` + "```" + `
```

- [ ] **Step 2: Verify compilation**

Run: `cd dataset-generation-scripts && go build ./internal/dataset_writer/dataset_file_writer/`
Expected: exit code 0

- [ ] **Step 3: Run tests**

Run: `cd dataset-generation-scripts && go test ./internal/dataset_writer/dataset_file_writer/ -v -run TestWrite 2>&1`
Expected: all tests pass

- [ ] **Step 4: Commit**

```bash
git add dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer.go
git commit -m "feat: add GIS Properties to provinces-gis README example"
```

---

### Task 7: Update tests to verify Properties

**Files:**
- Modify: `dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer_test.go` — `TestWriteElasticsearchGISDataToFile_GIS`

**Interfaces:**
- Consumes: `ElasticsearchGISProperties` from Task 1, updated `WriteElasticsearchGISDataToFile` from Tasks 2-3
- Produces: Updated test assertions verifying Properties on province and ward GIS

- [ ] **Step 1: Update test data to include Ma and DienTichKM2**

The test currently doesn't set `DienTichKM2` on the geo units. Add those fields and also set realistic `Ma` values.

Replace the geoProvinces declaration (around line 126):
```go
	geoProvinces := []*sapnhapbandomodel.SapNhapSiteGeoUnit{
		{
			Ma: "diaphanhanhchinhcaptinh_sn.108", Ten: "Hà Nội", VNDSProvinceCode: "01",
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
```

Replace the geoWards declaration (around line 143):
```go
	geoWards := []*sapnhapbandomodel.SapNhapSiteGeoUnit{
		{
			Ma: "diaphanhanhchinhphuong_sn.456", Ten: "Ba Đình", VNDSProvinceCode: "01",
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
```

- [ ] **Step 2: Add Properties assertions after existing GIS assertions (after line 200)**

After the line `t.Error("expected ward GIS field to be populated")` (line 199-200), add:

```go
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
```

- [ ] **Step 3: Run the GIS test**

Run: `cd dataset-generation-scripts && go test ./internal/dataset_writer/dataset_file_writer/ -v -run TestWriteElasticsearchGISDataToFile_GIS`
Expected: PASS with Properties assertions verified

- [ ] **Step 4: Run full test suite**

Run: `cd dataset-generation-scripts && go test ./internal/dataset_writer/dataset_file_writer/ -v 2>&1`
Expected: all 30 tests pass

- [ ] **Step 5: Commit**

```bash
git add dataset-generation-scripts/internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer_test.go
git commit -m "test: add Properties assertions to GIS integration test"
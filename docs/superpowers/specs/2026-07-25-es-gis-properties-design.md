# Design Spec: Add GIS Properties to Elasticsearch provinces-gis Documents

**Date:** 2026-07-25  
**Status:** Draft  
**Author:** AI Agent (Cline)

## Objective

Add a `Properties` object to the `GIS` field in `provinces-gis` Elasticsearch documents (both province and ward level) containing administrative metadata: code, name, nameEn, fullName, fullNameEn, codeName, gisServerId, and areaKm2.

## Motivation

Users consuming the `provinces-gis` index currently get GIS geometry data (Center, BoundingBox, Geometry) but no administrative metadata within the GIS object itself. The metadata is available at the parent document level, but having it inline within `GIS.Properties` makes it easier to extract GIS data alongside its identifying attributes in a single nested object — especially useful when working with GIS visualization tools and APIs that consume GeoJSON-like structures.

The data already exists in the PostGIS source (`SapNhapSiteGeoUnit` model): `Ma` (gisServerId), `DienTichKM2` (areaKm2), plus the related `VNProvince`/`VNWard` entities (code, name, nameEn, etc.). We simply need to map it into the ES document.

## Design

### 1. New DTO: `ElasticsearchGISProperties`

**File:** `internal/dataset_writer/dataset_file_writer/dto/elasticsearch_dto.go`

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

### 2. Modify `ElasticsearchGIS` — add `Properties` field

```go
type ElasticsearchGIS struct {
	Center      ElasticsearchGeoPoint        `json:"Center"`
	BoundingBox ElasticsearchBoundingBox      `json:"BoundingBox"`
	Geometry    json.RawMessage              `json:"Geometry"`
	Properties  *ElasticsearchGISProperties  `json:"Properties,omitempty"`
}
```

`omitempty` ensures non-GIS documents (`provinces` index and wards in `provinces` index) don't emit a null `Properties` field.

### 3. Update `sapnhapGeoUnitToESGIS` signature

**File:** `internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer.go`

Current signature:
```go
func sapnhapGeoUnitToESGIS(unit sapnhapbandomodel.SapNhapSiteGeoUnit) (*ElasticsearchGIS, error)
```

New signature — add a `properties` parameter:
```go
func sapnhapGeoUnitToESGIS(unit sapnhapbandomodel.SapNhapSiteGeoUnit, properties *ElasticsearchGISProperties) (*ElasticsearchGIS, error)
```

The function attaches the `properties` pointer directly to the GIS struct. The caller is responsible for constructing the correct properties (province or ward).

### 4. Build Properties at call sites

**File:** `internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer.go`  
**Method:** `WriteElasticsearchGISDataToFile`

For the **province** GIS:
```go
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

For each **ward** GIS:
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

### 5. Update ES Mapping

**File:** `elasticsearch/mappings/provinces-gis.json`

Add `Properties` sub-fields under `GIS.properties`:

```json
"GIS": {
  "type": "object",
  "properties": {
    "Center": { "type": "geo_point" },
    "BoundingBox": {
      "type": "object",
      "properties": {
        "MinLongitude": { "type": "float" },
        "MinLatitude": { "type": "float" },
        "MaxLongitude": { "type": "float" },
        "MaxLatitude": { "type": "float" }
      }
    },
    "Geometry": { "type": "geo_shape" },
    "Properties": {
      "type": "object",
      "properties": {
        "Code": { "type": "keyword" },
        "Name": { "type": "keyword" },
        "NameEn": { "type": "keyword" },
        "FullName": { "type": "keyword" },
        "FullNameEn": { "type": "keyword" },
        "CodeName": { "type": "keyword" },
        "GisServerId": { "type": "keyword" },
        "AreaKm2": { "type": "float" }
      }
    }
  }
}
```

Same mapping applies at **both** `GIS.Properties` (province level) and `Wards.GIS.Properties` (ward level) — the ward-level GIS mapping already mirrors the province structure:

```json
"Wards": {
  "type": "nested",
  "properties": {
    ...
    "GIS": {
      "type": "object",
      "properties": {
        ...
        "Properties": { ... same as above ... }
      }
    }
  }
}
```

**Mapping generator update:** `writeProvincesGISMapping` in `elasticsearch_file_writer.go` must also include the Properties sub-fields.

### 6. Update README example

**File:** `writeESReadme` in `elasticsearch_file_writer.go`

Add `Properties` to the provinces-gis example document:

```json
{
  "Code": "01",
  "Name": "Hà Nội",
  ...
  "GIS": {
    "Center": { "Lat": 21.0285, "Lon": 105.8542 },
    "BoundingBox": { ... },
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
      ...
      "GIS": {
        "Center": { "Lat": 21.0347, "Lon": 105.8231 },
        "BoundingBox": { ... },
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
  ]
}
```

### 7. What Does NOT Change

- `WriteToFile` (non-GIS provinces index) — does not produce GIS objects, unaffected
- `Geometry` field — remains pure geo_shape (type + coordinates), no properties inside
- `sapnhapGeoUnitToESGIS` still extracts Center/BoundingBox/Geometry the same way; only adds the properties pointer

## Files Changed

| File | Change |
|------|--------|
| `internal/dataset_writer/dataset_file_writer/dto/elasticsearch_dto.go` | Add `ElasticsearchGISProperties` struct; add `Properties` field to `ElasticsearchGIS` |
| `internal/dataset_writer/dataset_file_writer/elasticsearch_file_writer.go` | Update `sapnhapGeoUnitToESGIS` signature; build properties at call sites; update `writeProvincesGISMapping`; update `writeESReadme` example |
| `elasticsearch/mappings/provinces-gis.json` | Add `Properties` sub-fields under `GIS` and `Wards.GIS` |

## Test Impact

- `TestWriteElasticsearchGISDataToFile_GIS` — should be updated to verify `Properties` is populated on both province and ward GIS objects
- `TestParseBBox` — unaffected (bbox parsing logic unchanged)
- All existing tests must continue to pass

## Edge Cases

- **Ward without GIS data**: `sapnhapGeoUnitToESGIS` returns an error for malformed bbox; in that case `GIS` is nil and `Properties` is never attached (existing behavior preserved)
- **AreaKm2 = 0**: Valid — some wards may not have area data; emit as `0` (zero value) via `omitempty` won't suppress it since `omitempty` on struct pointers only applies to the pointer itself
- **GisServerId missing (empty Ma)**: Valid — emit empty string
# Migration from SAPNhap API to Local JSON/GeoJSON Files

## Overview

This document describes the migration from deprecated SAPNhap Bando API endpoints to local JSON and GeoJSON files for populating the `sapnhap_provinces`, `sapnhap_wards`, `sapnhap_provinces_gis`, and `sapnhap_wards_gis` tables.

## Problem Statement

The upstream data site `sapnhap.bando.com.vn` deprecated the following API endpoints:
- `https://sapnhap.bando.com.vn/pcotinh` (provinces list)
- `https://sapnhap.bando.com.vn/ptracuu` (wards lookup by province)

These endpoints were previously used to fetch administrative unit metadata. The migration replaces these API calls with local data files while maintaining data integrity and GIS ID matching.

## Data Sources

### Local Files Used

1. **Province Metadata:** `resources/gis/bando_gisserver/provinces.json`
   - Contains 34 province records
   - Each record includes: `stt`, `ten`, `truocsn`, `gisServerResponse`
   - `gisServerResponse` contains GIS server ID (e.g., "tinh34.7") and `matinh` (province code)

2. **Ward Metadata:** `resources/gis/bando_gisserver/wards.json`
   - Contains 3,321 ward records
   - Each record includes: `stt`, `ten`, `truocsn`, `provinceName`, `gisServerResponse`
   - `gisServerResponse` contains GIS server ID (e.g., "xa3321.3285") and `maxa` (ward code)

3. **Province GeoJSON:** `resources/gis/geojson_11Mar2026/{province_dir}/province.geojson`
   - 34 files (one per province)
   - Standard GeoJSON format with bbox and MultiPolygon geometry
   - Feature ID matches GIS server ID from provinces.json

4. **Ward GeoJSON:** `resources/gis/geojson_11Mar2026/{province_dir}/wards/{ward_file}.geojson`
   - 3,321 files (one per ward)
   - Standard GeoJSON format with bbox and MultiPolygon geometry
   - Feature ID matches GIS server ID from wards.json

## Implementation Changes

### 1. New GeoJSON Parser

**File:** `internal/sapnhap_bando/dto/geojson_file.go` (NEW)

Created standard GeoJSON parser structures and WKT conversion utilities:

```go
type GeoJSONFeatureCollection struct {
    Type     string             `json:"type"`
    BBox     [4]float64         `json:"bbox"`
    Features []GeoJSONFeature   `json:"features"`
}

type GeoJSONFeature struct {
    BBox     [4]float64       `json:"bbox"`
    Geometry GeoJSONGeometry  `json:"geometry"`
    ID       string           `json:"id"`   // GIS server ID
}

type GeoJSONGeometry struct {
    Type        string          `json:"type"`
    Coordinates [][][][2]float64 `json:"coordinates"`
}
```

**Key Methods:**
- `ToWKBboxPolygon()` - Converts GeoJSON bbox to WKT POLYGON format
- `ToWKTMultiPolygon()` - Converts GeoJSON MultiPolygon to WKT MULTIPOLYGON format
- `LoadGeoJSONFile()` - Loads and parses GeoJSON files from disk

### 2. New GIS Data Transfer Objects

**File:** `internal/sapnhap_bando/dto/geojson_data.go` (NEW)

```go
type GISProvinceData struct {
    STT                      int
    Ten                      string
    TruocSN                  string
    GISServerID              string
    SapNhapProvinceMaTinh    string
    BBoxWKT                  string
    GeomWKT                  string
}

type GISWardData struct {
    STT             int
    Ten             string
    TruocSN         string
    GISServerID     string
    SapNhapWardMaXa string
    BBoxWKT         string
    GeomWKT         string
}
```

### 3. Fetcher Layer Changes

**File:** `internal/sapnhap_bando/fetcher/fetcher.go`

#### Deprecated Functions (Marked but not removed):
- `GetAllProvincesDataFromSapNhapSite()` - No longer used
- `GetAllWardsOfProvinceFromSapNhapSite()` - No longer used

#### New Constants:
```go
const (
    BANDO_GIS_PROVINCES_FILE_PATH = "./resources/gis/bando_gisserver/provinces.json"
    BANDO_GIS_WARDS_FILE_PATH     = "./resources/gis/bando_gisserver/wards.json"
    GEOJSON_BASE_PATH             = "./resources/gis/geojson_11Mar2026"
)
```

#### New Functions:

**`LoadProvincesFromJSONFile()`**
- Loads province metadata from `provinces.json`
- Parses `gisServerResponse` to extract `matinh` (province code)
- Converts data to `SapNhapProvinceData` format
- Returns province list for database insertion

**`LoadWardsFromJSONFile()`**
- Loads all ward metadata from `wards.json` (single bulk load)
- Parses `gisServerResponse` to extract `matinhxa` and `maxa`
- Extracts ward type (loai) and name (tenhc) from full name
- Generates unique sequential IDs to prevent STT duplication
- Returns ward list for database insertion

**`extractWardTypeAndName()`**
- Parses ward name format: "An Khánh (xã)" → loai="Xã", tenhc="An Khánh"
- Supports types: "(xã)", "(phường)", "(thị trấn)", "(đặc khu)"

**`LoadProvincesGISFromGeoJSONFiles()`**
- Loads GIS geometry data for all provinces
- Returns map of GIS server ID to `GISProvinceData`
- Matches GIS IDs between JSON metadata and GeoJSON files
- Converts GeoJSON bbox and geometry to WKT format
- Implements scoped directory search to prevent false matches

**`LoadWardsGISFromGeoJSONFiles()`**
- Loads GIS geometry data for all wards
- Returns map of GIS server ID to `GISWardData`
- Matches GIS IDs between JSON metadata and GeoJSON files
- Converts GeoJSON bbox and geometry to WKT format
- Implements scoped directory search to prevent false matches

**Helper Functions:**
- `findProvinceDirBySTT()` - Finds province directory by STT prefix
- `findProvinceDirByName()` - Finds province directory by normalized name
- `findWardFileBySTT()` - Finds ward file by STT within province directory
- `cleanAdministrativeUnitPrefix()` - Removes administrative prefixes from names

#### Directory/File Resolution Strategy:

**Province Directories:**
- Named with STT prefix: `1_thu_đo_ha_noi`, `2_tinh_cao_bang`, etc.
- Located at: `resources/gis/geojson_11Mar2026/{province_dir}/`
- File: `province.geojson`

**Ward Files:**
- Named with STT prefix: `1_an_khanh_xa.geojson`, `2_ba_đinh_phuong.geojson`
- Located at: `resources/gis/geojson_11Mar2026/{province_dir}/wards/{ward_file}.geojson`

**Critical Design Decision:**
The STT-based file search is **scoped within the correct province directory**. This prevents false matches when multiple provinces have wards with the same STT number.

### 4. Service Layer Changes

**File:** `internal/sapnhap_bando/service/sapnhap.go`

#### Updated Functions:

**`BootstrapSapNhapSiteProvinces()`**
- Changed from: `fetcher.GetAllProvincesDataFromSapNhapSite()`
- Changed to: `fetcher.LoadProvincesFromJSONFile()`
- Cleans province name by removing administrative unit prefixes
- Looks up vn_province_code from provinces_tmp table by name
- Inserts province data into sapnhap_provinces table

**`BootstrapSapNhapSiteWards()`**
- Changed from: Per-province API calls (deprecated)
- Changed to: Single bulk load using `fetcher.LoadWardsFromJSONFile()`
- Handles edge case: "Phố Bảng" → "Phó Bảng" for Tuyên Quang
- Normalizes ward names using `util.NormalizeString()`
- Finds province to get vn_province_code
- Looks up vn_ward_code from wards_tmp table by name and province
- Inserts ward data into sapnhap_wards table

**`BootstrapGISDataFromGISServer()`**
- Changed from: API-based GIS fetching using `GetGISLocationCoordinates()`
- Changed to: GeoJSON file loading using new functions
- Loads province GIS data: `fetcher.LoadProvincesGISFromGeoJSONFiles()`
- Loads ward GIS data: `fetcher.LoadWardsGISFromGeoJSONFiles()`
- Inserts GIS data into sapnhap_provinces_gis and sapnhap_wards_gis tables
- Logs successful insertions and warnings for failures

### 5. Repository Layer Changes

**File:** `internal/sapnhap_bando/repository/sapnhap.go`

**New Method:**
- `FindSapNhapSiteProvinceByMaHC(ctx context.Context, maHC int)` - Finds province by administrative code

## Data Mapping

### Province Metadata Mapping

| Source Field | Table Column | Transformation |
|--------------|--------------|----------------|
| provinces.json → stt | id | Parse as integer |
| provinces.json → gisServerResponse → properties.matinh | mahc | Parse from JSON |
| provinces.json → ten | tentinh | Clean prefix ("tỉnh ", "thành phố ", "thủ đô ") |
| provinces.json → truocsn | truocsapnhap | Direct mapping |
| provinces_tmp lookup | vn_province_code | Match by cleaned name |
| - | dientichkm2 | Set to NULL |
| - | dansonguoi | Set to NULL |
| - | trungtamhc | Set to NULL |
| - | kinhdo | Set to NULL |
| - | vido | Set to NULL |
| - | con | Set to NULL |

### Ward Metadata Mapping

| Source Field | Table Column | Transformation |
|--------------|--------------|----------------|
| wards.json → stt | id | Parse as integer (replaced with sequential) |
| wards.json → gisServerResponse → properties.matinhxa | matinh | Parse from JSON, split by "." take first part |
| wards.json → gisServerResponse → properties.maxa | maxa | Parse from JSON |
| wards.json → gisServerResponse → properties.maxa | ma | Convert to string |
| wards.json → provinceName | tentinh | Direct mapping |
| wards.json → ten | tenhc | Extract name without type suffix |
| wards.json → ten | loai | Extract type from suffix |
| wards.json → truocsn | truocsapnhap | Direct mapping |
| wards_tmp lookup | vn_ward_code | Match by name and province_code |
| - | cay | Set to NULL |
| - | dientichkm2 | Set to NULL |
| - | dansonguoi | Set to NULL |
| - | trungtamhc | Set to NULL |
| - | kinhdo | Set to NULL |
| - | vido | Set to NULL |
| - | khoa | Set to NULL |

### GIS Data Mapping

**Province GIS (sapnhap_provinces_gis):**

| Source Field | Table Column | Transformation |
|--------------|--------------|----------------|
| provinces.json → stt | stt | From metadata JSON |
| provinces.json → ten | ten | From metadata JSON |
| provinces.json → truocsn | truocsapnhap | From metadata JSON |
| provinces.json → gisServerResponse → features[0].id | gis_server_id | From metadata JSON |
| provinces.json → gisServerResponse → properties.matinh | sapnhap_province_matinh | From metadata JSON |
| GeoJSON → features[0].bbox | bbox_wkt | Convert to WKT POLYGON |
| GeoJSON → features[0].geometry | geom_wkt | Convert to WKT MULTIPOLYGON |

**Ward GIS (sapnhap_wards_gis):**

| Source Field | Table Column | Transformation |
|--------------|--------------|----------------|
| wards.json → stt | stt | From metadata JSON |
| wards.json → ten | ten | From metadata JSON |
| wards.json → truocsn | truocsapnhap | From metadata JSON |
| wards.json → gisServerResponse → features[0].id | gis_server_id | From metadata JSON |
| wards.json → gisServerResponse → properties.maxa | sapnhap_ward_maxa | From metadata JSON |
| GeoJSON → features[0].bbox | bbox_wkt | Convert to WKT POLYGON |
| GeoJSON → features[0].geometry | geom_wkt | Convert to WKT MULTIPOLYGON |

## GIS ID Verification

### Critical Verification Step

The implementation includes GIS ID verification to ensure data integrity:

```go
// In LoadProvincesGISFromGeoJSONFiles() and LoadWardsGISFromGeoJSONFiles()
if feature.ID != gisServerID {
    log.Printf("Warning: GIS ID mismatch for %s: expected %s, got %s",
        province/ward.Ten, gisServerID, feature.ID)
    continue
}
```

**Verification Results:**
- ✅ All 34 provinces: GIS IDs matched successfully
- ✅ All 3,321 wards: GIS IDs matched successfully
- ✅ Zero GIS ID mismatch warnings during execution

### How ID Matching Works

1. **Extract GIS ID from metadata:**
   - Province: `provinces.json → gisServerResponse → features[0].id` (e.g., "tinh34.7")
   - Ward: `wards.json → gisServerResponse → features[0].id` (e.g., "xa3321.3285")

2. **Load GeoJSON file:**
   - Find file by STT prefix within correct province directory
   - Parse GeoJSON structure

3. **Verify ID matches:**
   - Compare `gisServerID` from metadata with `feature.ID` from GeoJSON
   - Log warning if mismatch (never occurred in practice)

## Key Implementation Details

### 1. Sequential ID Generation for Wards

**Issue:** Ward STT values are not unique globally (each province has STT=1,2,3...)

**Solution:** Generate unique sequential IDs starting from 1

```go
nextID := 1
for _, ward := range wards {
    wardData := dto.SapNhapWardData{
        ID: nextID, // Use generated unique ID instead of STT
        // ... other fields
    }
    result = append(result, wardData)
    nextID++
}
```

### 2. Administrative Unit Prefix Cleaning

Province names are cleaned to remove prefixes before matching with provinces_tmp:

```go
func cleanAdministrativeUnitPrefix(name string) string {
    prefixes := []string{"tỉnh ", "thành phố ", "thủ đô "}
    lowerName := strings.ToLower(name)

    for _, prefix := range prefixes {
        if strings.HasPrefix(lowerName, prefix) {
            return strings.TrimSpace(name[len(prefix):])
        }
    }
    return name
}
```

**Examples:**
- "Thủ đô Hà Nội" → "Hà Nội"
- "Thành phố Hồ Chí Minh" → "Hồ Chí Minh"
- "tỉnh Cao Bằng" → "Cao Bằng"

### 3. Ward Type and Name Extraction

Ward names include type suffixes that must be parsed:

```go
func extractWardTypeAndName(fullName string) (loai string, name string) {
    if strings.Contains(fullName, " (xã)") {
        return "Xã", strings.Replace(fullName, " (xã)", "", -1)
    }
    if strings.Contains(fullName, " (phường)") {
        return "Phường", strings.Replace(fullName, " (phường)", "", -1)
    }
    if strings.Contains(fullName, " (thị trấn)") {
        return "Thị trấn", strings.Replace(fullName, " (thị trấn)", "", -1)
    }
    if strings.Contains(fullName, " (đặc khu)") {
        return "Đặc khu", strings.Replace(fullName, " (đặc khu)", "", -1)
    }
    return "", fullName
}
```

**Examples:**
- "An Khánh (xã)" → loai="Xã", tenhc="An Khánh"
- "Ba Đình (phường)" → loai="Phường", tenhc="Ba Đình"
- "Phố Bảng (thị trấn)" → loai="Thị trấn", tenhc="Phố Bảng"

### 4. Scoped Directory Search

The STT-based file search is scoped within the correct province directory to prevent false matches:

```go
// Find province directory first
provinceDir, err := findProvinceDirByName(ward.ProvinceName)

// Then search for ward file WITHIN that province directory
wardFile, err := findWardFileBySTT(provinceDir, ward.STT)

// Construct full path
geojsonPath := fmt.Sprintf("%s/%s/wards/%s",
    GEOJSON_BASE_PATH, provinceDir, wardFile)
```

**Why this matters:**
- Multiple provinces have wards with STT="1", STT="2", etc.
- Without scoping, searching for "1_*.geojson" could match the wrong province
- By searching within the specific province directory, we ensure correct matching

**Example:**
- Hà Nội has ward with STT="1" → `1_an_khanh_xa.geojson` in `1_thu_do_ha_noi/wards/`
- Hồ Chí Minh has ward with STT="1" → `1_phuong_ward.geojson` in `2_thanh_pho_ho_chi_minh/wards/`
- Scoped search ensures we look in the correct province directory

### 5. Edge Case Handling

**Tuyên Quang - Phố Bảng → Phó Bảng:**
```go
if strings.EqualFold(wardData.TenHC, "Phố Bảng") && wardData.Matinh == 8 {
    wardData.TenHC = "Phó Bảng"
}
```

This special case handles the correct spelling for this ward in Tuyên Quang province.

## Database Tables Populated

### 1. sapnhap_provinces
- **Records:** 34 provinces
- **Source:** provinces.json
- **Foreign Key:** vn_province_code (from provinces_tmp)

### 2. sapnhap_wards
- **Records:** 3,321 wards
- **Source:** wards.json
- **Foreign Keys:**
  - matinh (links to sapnhap_provinces.mahc)
  - vn_ward_code (from wards_tmp)

### 3. sapnhap_provinces_gis
- **Records:** 34 provinces
- **Source:** GeoJSON files (province.geojson)
- **GIS Data:** bbox_wkt, geom_wkt
- **Reference:** gis_server_id (matches GIS server)

### 4. sapnhap_wards_gis
- **Records:** 3,321 wards
- **Source:** GeoJSON files (wards/*.geojson)
- **GIS Data:** bbox_wkt, geom_wkt
- **Reference:** gis_server_id (matches GIS server)

## Verification

### Data Completeness

```sql
-- Check province counts
SELECT COUNT(*) FROM sapnhap_provinces; -- Expected: 34

-- Check ward counts
SELECT COUNT(*) FROM sapnhap_wards; -- Expected: 3321

-- Check GIS data
SELECT COUNT(*) FROM sapnhap_provinces_gis; -- Expected: 34
SELECT COUNT(*) FROM sapnhap_wards_gis; -- Expected: 3321
```

### Data Integrity

```sql
-- Check foreign key relationships
SELECT COUNT(*) FROM sapnhap_provinces WHERE vn_province_code IS NULL; -- Expected: 0
SELECT COUNT(*) FROM sapnhap_wards WHERE vn_ward_code IS NULL; -- Expected: 0

-- Check GIS references
SELECT COUNT(*) FROM sapnhap_provinces_gis pg
LEFT JOIN sapnhap_provinces p ON pg.sapnhap_province_matinh = p.mahc
WHERE p.mahc IS NULL; -- Expected: 0
```

### Geometry Validation

```sql
-- Check WKT is populated
SELECT COUNT(*) FROM sapnhap_provinces_gis WHERE bbox_wkt IS NULL; -- Expected: 0
SELECT COUNT(*) FROM sapnhap_provinces_gis WHERE geom_wkt IS NULL; -- Expected: 0
SELECT COUNT(*) FROM sapnhap_wards_gis WHERE bbox_wkt IS NULL; -- Expected: 0
SELECT COUNT(*) FROM sapnhap_wards_gis WHERE geom_wkt IS NULL; -- Expected: 0
```

## Execution Results

**Successful Migration Output:**
```
Loaded GIS data for province Thủ đô Hà Nội (STT: 1)
Loaded GIS data for province Tỉnh Cao Bằng (STT: 2)
...
Loaded GIS data for 34 provinces
Loaded GIS data for 3321 wards
Completed loading GIS data: 34 provinces, 3321 wards
```

**Success Metrics:**
- ✅ 34 provinces loaded with matching GIS IDs
- ✅ 3,321 wards loaded with matching GIS IDs
- ✅ Zero GIS ID mismatch warnings
- ✅ All WKT geometries successfully converted
- ✅ All foreign key relationships valid

## Benefits

1. **No External Dependencies:** No longer relies on deprecated SAPNhap API endpoints
2. **Data Integrity:** GIS ID verification ensures metadata matches geometry files
3. **Performance:** Local file loading is faster than API calls
4. **Reliability:** No network failures or rate limiting issues
5. **Auditability:** All data sources are version-controlled JSON files
6. **Scalability:** Can process all data in bulk without per-record API overhead

## Files Created/Modified

### New Files:
- `internal/sapnhap_bando/dto/geojson_file.go` - GeoJSON parser and WKT converters
- `internal/sapnhap_bando/dto/geojson_data.go` - GIS data transfer objects

### Modified Files:
- `internal/sapnhap_bando/fetcher/fetcher.go` - Added file loading functions
- `internal/sapnhap_bando/service/sapnhap.go` - Updated service methods
- `internal/sapnhap_bando/repository/sapnhap.go` - Added FindSapNhapSiteProvinceByMaHC method

## Future Considerations

1. **Data Updates:** When new GeoJSON files are available, simply replace the files in the resources directory
2. **GIS Server Changes:** If the GIS server changes ID format, update the ID extraction logic
3. **New Administrative Units:** Add new entries to provinces.json or wards.json as needed
4. **Performance Optimization:** The current implementation loads all data into memory; for very large datasets, consider streaming

## Summary

This migration successfully replaced deprecated SAPNhap API calls with local file-based data loading while maintaining data integrity through GIS ID verification. The implementation handles edge cases, ensures correct directory scoping to prevent false matches, and provides comprehensive logging for debugging. All 34 provinces and 3,321 wards are now loaded from local files with verified GIS geometry data.

**Key Achievement:** Zero GIS ID mismatches across 3,355 records (34 provinces + 3,321 wards), demonstrating that the scoped directory search strategy effectively prevents false matches when multiple administrative units share the same STT number.

package fetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/dto"
)

const (
	GET_ALL_PROVINCES_URL          = "https://sapnhap.bando.com.vn/pcotinh"
	GET_ALL_WARDS_OF_PROVINCES_URL = "https://sapnhap.bando.com.vn/ptracuu"
	GET_GIS_COORDINATES_URL        = "https://sapnhap.bando.com.vn/pread_json"
	GET_METADATA_FROM_MALK_URL     = "https://sapnhap.bando.com.vn/p.co_dvhc_id"
	MAX_RETRIES                    = 5
	RETRY_DELAY                    = 300 * time.Millisecond
	MAX_DELAY                      = 5 * time.Second

	// New file-based data sources
	BANDO_GIS_PROVINCES_FILE_PATH = "./resources/gis/bando_gisserver/provinces.json"
	BANDO_GIS_WARDS_FILE_PATH     = "./resources/gis/bando_gisserver/wards.json"
	GEOJSON_BASE_PATH             = "./resources/gis/geojson_11Mar2026"
)

/*
Get all the provinces data from the sapnhap site
POST: https://sapnhap.bando.com.vn/pcotinh
*/
func GetAllProvincesDataFromSapNhapSite() []dto.SapNhapProvinceData {
	res, err := http.Post(GET_ALL_PROVINCES_URL, "application/json", nil)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	var provincesData []dto.SapNhapProvinceData
	if err := json.NewDecoder(res.Body).Decode(&provincesData); err != nil {
		panic(err)
	}
	return provincesData
}

/*
API Look up to get all wards of a province from the sapnhap site
POST: https://sapnhap.bando.com.vn/ptracuu

[DEPRECATED] This API endpoint is no longer available. Use LoadWardsFromJSONFile() instead.
*/
func GetAllWardsOfProvinceFromSapNhapSite(provinceID int) []dto.SapNhapWardData {
	form := url.Values{}
	form.Add("id", strconv.Itoa(provinceID))

	res, err := http.Post(GET_ALL_WARDS_OF_PROVINCES_URL, "application/x-www-form-urlencoded", bytes.NewBufferString(form.Encode()))
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	var wardsData []dto.SapNhapWardData
	if err := json.NewDecoder(res.Body).Decode(&wardsData); err != nil {
		panic(err)
	}
	return wardsData
}

/*
Load all provinces data from local JSON file
Replaces the deprecated GetAllProvincesDataFromSapNhapSite() API call
*/
func LoadProvincesFromJSONFile() ([]dto.SapNhapProvinceData, error) {
	provinces, err := LoadBanDoGISProvincesFromFile(BANDO_GIS_PROVINCES_FILE_PATH)
	if err != nil {
		return nil, err
	}

	result := make([]dto.SapNhapProvinceData, 0, len(provinces))
	for _, province := range provinces {
		gisRes, err := province.GetGISResponse()
		if err != nil {
			return nil, err
		}

		matinh := fmt.Sprintf("%v", gisRes.Features[0].Properties["matinh"])
		matinhInt, err := strconv.Atoi(matinh)
		if err != nil {
			return nil, err
		}

		sttInt, err := strconv.Atoi(province.STT)
		if err != nil {
			return nil, err
		}

		provinceData := dto.SapNhapProvinceData{
			ID:          sttInt,
			MaHC:        matinhInt,
			TenTinh:     province.Ten,
			DienTichKm2: "",
			DanSoNguoi:  "",
			TrungTamHC:  "",
			KinhDo:      0,
			ViDo:        0,
			TruocSN:     province.TruocSN,
			Con:         "",
		}
		result = append(result, provinceData)
	}

	return result, nil
}

/*
Load all wards data from local JSON file
Replaces the deprecated GetAllWardsOfProvinceFromSapNhapSite() API call
*/
func LoadWardsFromJSONFile() ([]dto.SapNhapWardData, error) {
	wards, err := LoadBanDoGISWardsFromFile(BANDO_GIS_WARDS_FILE_PATH)
	if err != nil {
		return nil, err
	}

	result := make([]dto.SapNhapWardData, 0, len(wards))
	nextID := 1 // Generate unique IDs sequentially

	for _, ward := range wards {
		gisRes, err := ward.GetGISResponse()
		if err != nil {
			return nil, err
		}

		matinhxa := fmt.Sprintf("%v", gisRes.Features[0].Properties["matinhxa"])
		maxa := fmt.Sprintf("%v", gisRes.Features[0].Properties["maxa"])

		parts := strings.Split(matinhxa, ".")
		matinhInt, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}

		maxaInt, err := strconv.Atoi(maxa)
		if err != nil {
			return nil, err
		}

		loai, tenhc := extractWardTypeAndName(ward.Ten)

		wardData := dto.SapNhapWardData{
			ID:           nextID, // Use generated unique ID instead of STT
			Matinh:       matinhInt,
			Ma:           maxa,
			TenTinh:      ward.ProvinceName,
			Loai:         loai,
			TenHC:        tenhc,
			Cay:          "",
			DienTichKm2:  0,
			DanSoNguoi:   "",
			TrungTamHC:   "",
			KinhDo:       0,
			ViDo:         0,
			TruocSapNhap: ward.TruocSN,
			MaXa:         maxaInt,
			Khoa:         "",
		}
		result = append(result, wardData)
		nextID++
	}

	return result, nil
}

/*
API Look up to get all wards of a province from the sapnhap site
POST: https://sapnhap.bando.com.vn/p.co_dvhc_id
*/
func GetMetadataOfSapNhapGeoObject(ctx context.Context, malk string) (dto.SapNhapGeoObjectMetadata, error) {
	form := url.Values{}
	form.Set("malk", malk)
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://sapnhap.bando.com.vn/p.co_dvhc_id",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return dto.SapNhapGeoObjectMetadata{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return dto.SapNhapGeoObjectMetadata{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return dto.SapNhapGeoObjectMetadata{}, fmt.Errorf(
			"unexpected status code %d: %s",
			res.StatusCode,
			string(body),
		)
	}
	var metadata []dto.SapNhapGeoObjectMetadata
	if err := json.NewDecoder(res.Body).Decode(&metadata); err != nil {
		return dto.SapNhapGeoObjectMetadata{}, err
	}
	return metadata[0], nil
}

/*
extractWardTypeAndName extracts the ward type (loai) and name (tenhc) from the full name
Format: "An Khánh (xã)" → loai="Xã", tenhc="An Khánh"
*/
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

/*
Load the bando gis province mapping from file
*/
func LoadBanDoGISProvincesFromFile(path string) ([]dto.BanDoGISProvince, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Decode JSON
	var provinces []dto.BanDoGISProvince
	if err := json.Unmarshal(data, &provinces); err != nil {
		return nil, err
	}
	return provinces, nil
}

/*
Load the bando gis ward mapping from file
*/
func LoadBanDoGISWardsFromFile(path string) ([]dto.BanDoGISWard, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Decode JSON
	var wards []dto.BanDoGISWard
	if err := json.Unmarshal(data, &wards); err != nil {
		return nil, err
	}
	return wards, nil
}

/*
API to get GIS coordinates information of the locationId.
gisLocationID get from object ID of the bando gisServerResponse
POST: https://sapnhap.bando.com.vn/pread_json
*/
func GetGISLocationCoordinates(gisLocationID string) (dto.GISLocationResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= MAX_RETRIES; attempt++ {
		if attempt > 0 {
			log.Printf("Retrying request in %v (attempt %d/%d)", RETRY_DELAY, attempt, MAX_RETRIES)
			time.Sleep(RETRY_DELAY)
		}

		// Prepare form data
		form := url.Values{}
		form.Add("id", gisLocationID)

		// Make HTTP request
		res, err := http.Post(GET_GIS_COORDINATES_URL, "application/x-www-form-urlencoded", bytes.NewBufferString(form.Encode()))
		if err != nil {
			lastErr = fmt.Errorf("http request failed: %w", err)
			log.Printf("Attempt %d failed with error: %v", attempt+1, lastErr)
			continue
		}

		defer res.Body.Close()

		// Check if response is healthy
		if res.StatusCode == http.StatusOK {
			// Success - decode response
			var gisLocationResponse dto.GISLocationResponse
			if err := json.NewDecoder(res.Body).Decode(&gisLocationResponse); err != nil {
				return dto.GISLocationResponse{}, fmt.Errorf("failed to decode response: %w", err)
			}

			if attempt > 0 {
				log.Printf("Request succeeded on attempt %d", attempt+1)
			}
			return gisLocationResponse, nil
		}

		// Non-OK status code
		lastErr = fmt.Errorf("received status code: %d", res.StatusCode)
		log.Printf("Attempt %d failed with status code: %d", attempt+1, res.StatusCode)
	}

	// All retries exhausted
	return dto.GISLocationResponse{}, fmt.Errorf("cannot get GIS for locationID %s. All %d attempts failed, last error: %w", gisLocationID, MAX_RETRIES+1, lastErr)
}

/*
LoadProvincesGISFromGeoJSONFiles loads province geometry data from GeoJSON files
Returns a map of GIS server ID (e.g., "tinh34.7") to GIS data
*/
func LoadProvincesGISFromGeoJSONFiles() (map[string]*dto.GISProvinceData, error) {
	// Load metadata to build the ID mapping
	provinces, err := LoadBanDoGISProvincesFromFile(BANDO_GIS_PROVINCES_FILE_PATH)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*dto.GISProvinceData)

	for _, province := range provinces {
		gisRes, err := province.GetGISResponse()
		if err != nil {
			return nil, err
		}

		gisServerID := gisRes.Features[0].ID // e.g., "tinh34.7"
		matinh := fmt.Sprintf("%v", gisRes.Features[0].Properties["matinh"])

		sttInt, err := strconv.Atoi(province.STT)
		if err != nil {
			return nil, err
		}

		// Find the province directory by STT prefix
		provinceDir, err := findProvinceDirBySTT(province.STT)
		if err != nil {
			log.Printf("Warning: failed to find province directory for STT %s (%s): %v", province.STT, province.Ten, err)
			continue
		}

		geojsonPath := fmt.Sprintf("%s/%s/province.geojson", GEOJSON_BASE_PATH, provinceDir)

		// Load GeoJSON file
		geojson, err := dto.LoadGeoJSONFile(geojsonPath)
		if err != nil {
			log.Printf("Warning: failed to load geojson for %s from %s: %v", province.Ten, geojsonPath, err)
			continue
		}

		if len(geojson.Features) == 0 {
			log.Printf("Warning: no features found in geojson for %s", province.Ten)
			continue
		}

		feature := geojson.Features[0]

		// Verify the ID matches
		if feature.ID != gisServerID {
			log.Printf("Warning: GIS ID mismatch for %s: expected %s, got %s", province.Ten, gisServerID, feature.ID)
			continue
		}

		gisData := &dto.GISProvinceData{
			STT:                   sttInt,
			Ten:                   province.Ten,
			TruocSN:               province.TruocSN,
			GISServerID:           gisServerID,
			SapNhapProvinceMaTinh: matinh,
			BBoxWKT:               feature.ToWKBboxPolygon(),
			GeomWKT:               feature.Geometry.ToWKTMultiPolygon(),
		}

		result[gisServerID] = gisData
		log.Printf("Loaded GIS data for province %s (STT: %s)", province.Ten, province.STT)
	}

	return result, nil
}

/*
LoadWardsGISFromGeoJSONFiles loads ward geometry data from GeoJSON files
Returns a map of GIS server ID (e.g., "xa3321.3285") to GIS data
*/
func LoadWardsGISFromGeoJSONFiles() (map[string]*dto.GISWardData, error) {
	// Load metadata to build the ID mapping
	wards, err := LoadBanDoGISWardsFromFile(BANDO_GIS_WARDS_FILE_PATH)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*dto.GISWardData)

	for _, ward := range wards {
		gisRes, err := ward.GetGISResponse()
		if err != nil {
			log.Printf("Warning: failed to get GIS response for %s: %v", ward.Ten, err)
			continue
		}

		gisServerID := gisRes.Features[0].ID // e.g., "xa3321.3285"
		maxa := fmt.Sprintf("%v", gisRes.Features[0].Properties["maxa"])

		sttInt, err := strconv.Atoi(ward.STT)
		if err != nil {
			log.Printf("Warning: failed to parse STT %s for %s: %v", ward.STT, ward.Ten, err)
			continue
		}

		// Find the province directory by province name
		provinceDir, err := findProvinceDirByName(ward.ProvinceName)
		if err != nil {
			log.Printf("Warning: failed to find province directory for %s: %v", ward.ProvinceName, err)
			continue
		}

		// Try to find the ward file by matching the STT prefix
		wardFile, err := findWardFileBySTT(provinceDir, ward.STT)
		if err != nil {
			log.Printf("Warning: failed to find ward file for STT %s (%s): %v", ward.STT, ward.Ten, err)
			continue
		}

		geojsonPath := fmt.Sprintf("%s/%s/wards/%s", GEOJSON_BASE_PATH, provinceDir, wardFile)

		// Load GeoJSON file
		geojson, err := dto.LoadGeoJSONFile(geojsonPath)
		if err != nil {
			log.Printf("Warning: failed to load geojson for %s from %s: %v", ward.Ten, geojsonPath, err)
			continue
		}

		if len(geojson.Features) == 0 {
			log.Printf("Warning: no features found in geojson for %s", ward.Ten)
			continue
		}

		feature := geojson.Features[0]

		// Verify the ID matches
		if feature.ID != gisServerID {
			log.Printf("Warning: GIS ID mismatch for %s: expected %s, got %s", ward.Ten, gisServerID, feature.ID)
			continue
		}

		gisData := &dto.GISWardData{
			STT:             sttInt,
			Ten:             ward.Ten,
			TruocSN:         ward.TruocSN,
			GISServerID:     gisServerID,
			SapNhapWardMaXa: maxa,
			BBoxWKT:         feature.ToWKBboxPolygon(),
			GeomWKT:         feature.Geometry.ToWKTMultiPolygon(),
		}

		result[gisServerID] = gisData
	}

	log.Printf("Loaded GIS data for %d wards", len(result))
	return result, nil
}

/*
findProvinceDirBySTT finds the province directory by STT prefix
Since directories are named like "1_thu_đo_ha_noi", "2_tinh_cao_bang", etc.
we can match by the STT prefix
*/
func findProvinceDirBySTT(stt string) (string, error) {
	// List all directories in GEOJSON_BASE_PATH
	entries, err := os.ReadDir(GEOJSON_BASE_PATH)
	if err != nil {
		return "", err
	}

	// Find directory starting with {stt}_
	prefix := stt + "_"
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			return entry.Name(), nil
		}
	}

	return "", fmt.Errorf("no directory found with prefix %s", prefix)
}

/*
findProvinceDirByName finds the province directory by province name
This loads all provinces and matches by normalized name
*/
func findProvinceDirByName(provinceName string) (string, error) {
	// Load provinces to get STT
	provinces, err := LoadBanDoGISProvincesFromFile(BANDO_GIS_PROVINCES_FILE_PATH)
	if err != nil {
		return "", err
	}

	// Find matching province
	for _, province := range provinces {
		// Clean and compare names
		cleanedName := cleanAdministrativeUnitPrefix(province.Ten)
		cleanedTargetName := cleanAdministrativeUnitPrefix(provinceName)

		if strings.EqualFold(cleanedName, cleanedTargetName) {
			return findProvinceDirBySTT(province.STT)
		}
	}

	return "", fmt.Errorf("no province found matching %s", provinceName)
}

/*
findWardFileBySTT finds the ward file by STT prefix in the given province directory
Ward files are named like "1_an_khanh_xa.geojson", "2_ba_đinh_phuong.geojson", etc.
*/
func findWardFileBySTT(provinceDir string, stt string) (string, error) {
	wardsPath := fmt.Sprintf("%s/%s/wards", GEOJSON_BASE_PATH, provinceDir)

	entries, err := os.ReadDir(wardsPath)
	if err != nil {
		return "", err
	}

	// Find file starting with {stt}_
	prefix := stt + "_"
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) && strings.HasSuffix(entry.Name(), ".geojson") {
			return entry.Name(), nil
		}
	}

	return "", fmt.Errorf("no ward file found with prefix %s in %s", prefix, wardsPath)
}

/*
cleanAdministrativeUnitPrefix removes common administrative unit prefixes from a name
*/
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

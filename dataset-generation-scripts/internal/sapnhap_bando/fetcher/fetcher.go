package fetcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/dto"
)

const (
	GET_ALL_PROVINCES_URL          = "https://sapnhap.bando.com.vn/pcotinh"
	GET_ALL_WARDS_OF_PROVINCES_URL = "https://sapnhap.bando.com.vn/ptracuu"
	GET_GIS_COORDINATES_URL        = "https://sapnhap.bando.com.vn/pread_json"
	MAX_RETRIES                    = 5
	RETRY_DELAY                    = 300 * time.Millisecond
	MAX_DELAY                      = 5 * time.Second
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

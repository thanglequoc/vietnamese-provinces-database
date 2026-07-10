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
	"strings"
	"time"

	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/dto"
)

const (
	GET_GIS_COORDINATES_URL    = "https://sapnhap.bando.com.vn/pread_json"
	GET_METADATA_FROM_MALK_URL = "https://sapnhap.bando.com.vn/p.co_dvhc_id"
	MAX_RETRIES                = 5
	RETRY_DELAY                = 300 * time.Millisecond
	MAX_DELAY                  = 5 * time.Second
)

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

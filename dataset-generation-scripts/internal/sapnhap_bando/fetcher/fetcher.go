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
	"strings"
	"sync"
	"time"

	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/dto"
)

const (
	// DefaultGISServerBaseURL is the production GIS server. Override it with the
	// GIS_SERVER_BASE_URL environment variable to point at the local mock server
	// (see cmd/mockgis) for offline dataset generation.
	DefaultGISServerBaseURL = "https://sapnhap.bando.com.vn"

	PREAD_JSON_PATH = "/pread_json"
	CO_DVHC_ID_PATH = "/p.co_dvhc_id"
	MAX_RETRIES     = 5
	RETRY_DELAY     = 300 * time.Millisecond
	MAX_DELAY       = 5 * time.Second
)

// baseURLLogOnce ensures the active GIS server base URL is logged only once.
var baseURLLogOnce sync.Once

// GISServerBaseURL resolves the active GIS server base URL from the
// GIS_SERVER_BASE_URL environment variable, falling back to the production
// server. Trailing slashes are trimmed so path joins stay well-formed.
func GISServerBaseURL() string {
	if v := strings.TrimSpace(os.Getenv("GIS_SERVER_BASE_URL")); v != "" {
		return strings.TrimRight(v, "/")
	}
	return DefaultGISServerBaseURL
}

// LogActiveGISServer logs the active GIS server base URL once per process and
// flags whether it is the production government server or a custom/mock server.
// Safe to call multiple times; only the first call logs.
func LogActiveGISServer() {
	base := GISServerBaseURL()
	baseURLLogOnce.Do(func() {
		source := "government server"
		if base != DefaultGISServerBaseURL {
			source = "custom/mock server"
		}
		log.Printf("🌐 GIS server base URL: %s (%s)", base, source)
	})
}

// gisEndpoint builds an absolute URL for a GIS server path and logs the active
// base URL once per process to make mock vs. real runs obvious in the logs.
func gisEndpoint(path string) string {
	LogActiveGISServer()
	return GISServerBaseURL() + path
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
		gisEndpoint(CO_DVHC_ID_PATH),
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
		res, err := http.Post(gisEndpoint(PREAD_JSON_PATH), "application/x-www-form-urlencoded", bytes.NewBufferString(form.Encode()))
		if err != nil {
			lastErr = fmt.Errorf("http request failed: %w", err)
			log.Printf("Attempt %d failed with error: %v", attempt+1, lastErr)
			continue
		}

		// Check if response is healthy
		if res.StatusCode == http.StatusOK {
			// Success - decode response
			var gisLocationResponse dto.GISLocationResponse
			decodeErr := json.NewDecoder(res.Body).Decode(&gisLocationResponse)
			res.Body.Close()
			if decodeErr != nil {
				return dto.GISLocationResponse{}, fmt.Errorf("failed to decode response: %w", decodeErr)
			}

			if attempt > 0 {
				log.Printf("Request succeeded on attempt %d", attempt+1)
			}
			return gisLocationResponse, nil
		}

		// Non-OK status code
		res.Body.Close()
		lastErr = fmt.Errorf("received status code: %d", res.StatusCode)
		log.Printf("Attempt %d failed with status code: %d", attempt+1, res.StatusCode)
	}

	// All retries exhausted
	return dto.GISLocationResponse{}, fmt.Errorf("cannot get GIS for locationID %s. All %d attempts failed, last error: %w", gisLocationID, MAX_RETRIES+1, lastErr)
}

package capture

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/fetcher"
)

const (
	// EndpointPreadJSON is the geometry endpoint (POST form field: id).
	EndpointPreadJSON = "pread_json"
	// EndpointCoDvhcID is the metadata endpoint (POST form field: malk).
	EndpointCoDvhcID = "co_dvhc_id"

	defaultWorkers = 10
	requestTimeout = 30 * time.Second
	maxRetries     = 5
	retryDelay     = 300 * time.Millisecond
)

// Options configures a capture run.
type Options struct {
	// BaseURL is the GIS server to capture from. Defaults to the production
	// server. This is intentionally explicit (never read from
	// GIS_SERVER_BASE_URL) so a capture can never accidentally record the mock.
	BaseURL string
	// OutDir is the cache root (contains per-endpoint subdirectories).
	OutDir string
	// Workers is the number of concurrent capture workers.
	Workers int
	// Force overwrites existing cache files instead of skipping them.
	Force bool
	// Only restricts the run to specific endpoints. Empty means both.
	Only []string
	// DatasetVersion / LatestDecree are recorded in the manifest for staleness checks.
	DatasetVersion string
	LatestDecree   string
}

// EndpointManifest describes the capture result of a single endpoint.
type EndpointManifest struct {
	Expected int      `json:"expected"`
	Captured int      `json:"captured"`
	Missing  []string `json:"missing"`
}

// Manifest records the provenance of a captured cache.
type Manifest struct {
	SourceBaseURL  string                      `json:"source_base_url"`
	CapturedAt     time.Time                   `json:"captured_at"`
	DatasetVersion string                      `json:"dataset_version"`
	LatestDecree   string                      `json:"latest_decree"`
	Endpoints      map[string]EndpointManifest `json:"endpoints"`
}

// endpointPath maps an endpoint name to its GIS server path.
func endpointPath(endpoint string) string {
	switch endpoint {
	case EndpointPreadJSON:
		return fetcher.PREAD_JSON_PATH
	case EndpointCoDvhcID:
		return fetcher.CO_DVHC_ID_PATH
	default:
		return ""
	}
}

// endpointFormField maps an endpoint name to its POST form field.
func endpointFormField(endpoint string) string {
	switch endpoint {
	case EndpointPreadJSON:
		return "id"
	case EndpointCoDvhcID:
		return "malk"
	default:
		return ""
	}
}

// normalizeOnly validates and defaults the endpoint selection.
func normalizeOnly(only []string) ([]string, error) {
	if len(only) == 0 {
		return []string{EndpointPreadJSON, EndpointCoDvhcID}, nil
	}
	seen := make(map[string]bool)
	var result []string
	for _, ep := range only {
		ep = strings.TrimSpace(ep)
		if endpointPath(ep) == "" {
			return nil, fmt.Errorf("unknown endpoint %q (expected %q or %q)", ep, EndpointPreadJSON, EndpointCoDvhcID)
		}
		if !seen[ep] {
			seen[ep] = true
			result = append(result, ep)
		}
	}
	return result, nil
}

// dedupeMalks returns the unique, non-empty malk values preserving order.
func dedupeMalks(malks []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(malks))
	for _, m := range malks {
		m = strings.TrimSpace(m)
		if m == "" || seen[m] {
			continue
		}
		seen[m] = true
		result = append(result, m)
	}
	return result
}

// LoadMalksFromJSON reads the malk values from a donvi_tinhthanh-style JSON file.
func LoadMalksFromJSON(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read malk source %s: %w", path, err)
	}

	var entries []struct {
		Malk string `json:"malk"`
	}
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parse malk source %s: %w", path, err)
	}

	malks := make([]string, 0, len(entries))
	for _, entry := range entries {
		malks = append(malks, entry.Malk)
	}
	return dedupeMalks(malks), nil
}

// Run captures every (malk, endpoint) pair into the cache directory and writes a
// manifest. It is resumable: existing files are skipped unless opts.Force is set.
// The returned manifest is always non-nil when the run starts; the error is
// non-nil if any pair could not be captured.
func Run(ctx context.Context, malks []string, opts Options) (*Manifest, error) {
	if strings.TrimSpace(opts.OutDir) == "" {
		return nil, errors.New("capture out dir is required")
	}
	if strings.TrimSpace(opts.BaseURL) == "" {
		opts.BaseURL = fetcher.DefaultGISServerBaseURL
	}
	opts.BaseURL = strings.TrimRight(opts.BaseURL, "/")

	workers := opts.Workers
	if workers <= 0 {
		workers = defaultWorkers
	}

	endpoints, err := normalizeOnly(opts.Only)
	if err != nil {
		return nil, err
	}

	uniqueMalks := dedupeMalks(malks)
	if len(uniqueMalks) == 0 {
		return nil, errors.New("no malk values to capture")
	}

	for _, ep := range endpoints {
		if err := os.MkdirAll(filepath.Join(opts.OutDir, ep), 0o755); err != nil {
			return nil, fmt.Errorf("create cache dir for %s: %w", ep, err)
		}
	}

	type job struct {
		malk     string
		endpoint string
	}
	type result struct {
		malk     string
		endpoint string
		err      error
	}

	jobs := make(chan job)
	results := make(chan result)
	client := &http.Client{Timeout: requestTimeout}

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				err := captureOne(ctx, client, opts, j.endpoint, j.malk)
				results <- result{malk: j.malk, endpoint: j.endpoint, err: err}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, m := range uniqueMalks {
			for _, ep := range endpoints {
				jobs <- job{malk: m, endpoint: ep}
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	captured := make(map[string]int)
	missing := make(map[string][]string)
	processed := 0
	total := len(uniqueMalks) * len(endpoints)

	for r := range results {
		processed++
		if r.err != nil {
			missing[r.endpoint] = append(missing[r.endpoint], r.malk)
			log.Printf("[%d/%d] capture failed [%s %s]: %v", processed, total, r.endpoint, r.malk, r.err)
			continue
		}
		captured[r.endpoint]++
		if processed%100 == 0 || processed == total {
			log.Printf("[%d/%d] captured", processed, total)
		}
	}

	manifest := &Manifest{
		SourceBaseURL:  opts.BaseURL,
		CapturedAt:     time.Now().UTC(),
		DatasetVersion: opts.DatasetVersion,
		LatestDecree:   opts.LatestDecree,
		Endpoints:      make(map[string]EndpointManifest, len(endpoints)),
	}
	for _, ep := range endpoints {
		miss := missing[ep]
		if miss == nil {
			miss = []string{}
		}
		manifest.Endpoints[ep] = EndpointManifest{
			Expected: len(uniqueMalks),
			Captured: captured[ep],
			Missing:  miss,
		}
	}

	if err := WriteManifest(opts.OutDir, manifest); err != nil {
		return manifest, fmt.Errorf("write manifest: %w", err)
	}

	missingTotal := 0
	for _, ep := range endpoints {
		missingTotal += len(manifest.Endpoints[ep].Missing)
	}
	if missingTotal > 0 {
		return manifest, fmt.Errorf("capture incomplete: %d of %d pairs missing (see manifest.json)", missingTotal, total)
	}
	return manifest, nil
}

// captureOne fetches and persists a single (endpoint, malk) pair.
func captureOne(ctx context.Context, client *http.Client, opts Options, endpoint, malk string) error {
	target := filepath.Join(opts.OutDir, endpoint, malk+".json.gz")
	if !opts.Force {
		if _, err := os.Stat(target); err == nil {
			return nil
		}
	}

	body, err := fetchRaw(ctx, client, opts.BaseURL+endpointPath(endpoint), endpointFormField(endpoint), malk)
	if err != nil {
		return err
	}
	if err := validateResponse(endpoint, body); err != nil {
		return fmt.Errorf("invalid response: %w", err)
	}
	return writeGzip(target, body)
}

// fetchRaw POSTs a form value and returns the raw response body, retrying on
// transient failures with the same backoff pattern as the live fetcher.
func fetchRaw(ctx context.Context, client *http.Client, endpointURL, field, value string) ([]byte, error) {
	form := url.Values{}
	form.Set(field, value)

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if attempt > 0 {
			time.Sleep(retryDelay)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointURL, strings.NewReader(form.Encode()))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		req.Header.Set("Accept", "application/json, text/plain, */*")

		res, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		body, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if res.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("unexpected status %d: %s", res.StatusCode, truncate(string(body), 200))
			continue
		}
		return body, nil
	}
	return nil, fmt.Errorf("all %d attempts failed: %w", maxRetries+1, lastErr)
}

// validateResponse ensures the captured body is usable by the generation
// pipeline. GetMetadataOfSapNhapGeoObject indexes metadata[0] without a length
// check, so an empty metadata array must be rejected here.
func validateResponse(endpoint string, body []byte) error {
	switch endpoint {
	case EndpointPreadJSON:
		var collection struct {
			Features []json.RawMessage `json:"features"`
		}
		if err := json.Unmarshal(body, &collection); err != nil {
			return fmt.Errorf("parse pread_json body: %w", err)
		}
		if len(collection.Features) == 0 {
			return errors.New("pread_json response has no features")
		}
	case EndpointCoDvhcID:
		var metadata []json.RawMessage
		if err := json.Unmarshal(body, &metadata); err != nil {
			return fmt.Errorf("parse co_dvhc_id body: %w", err)
		}
		if len(metadata) == 0 {
			return errors.New("co_dvhc_id response is an empty array")
		}
	default:
		return fmt.Errorf("unknown endpoint %q", endpoint)
	}
	return nil
}

// writeGzip writes data as a deterministic gzip file (zero ModTime) via a
// temporary file so partial writes never corrupt the cache.
func writeGzip(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	gz, err := gzip.NewWriterLevel(f, gzip.BestCompression)
	if err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	gz.ModTime = time.Time{}

	if _, err := gz.Write(data); err != nil {
		gz.Close()
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := gz.Close(); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}

// WriteManifest persists the manifest as indented JSON.
func WriteManifest(dir string, manifest *Manifest) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "manifest.json"), append(data, '\n'), 0o644)
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

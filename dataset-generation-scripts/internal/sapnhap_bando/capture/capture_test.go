package capture

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeOnly(t *testing.T) {
	endpoints, err := normalizeOnly(nil)
	require.NoError(t, err)
	assert.Equal(t, []string{EndpointPreadJSON, EndpointCoDvhcID}, endpoints)

	endpoints, err = normalizeOnly([]string{EndpointCoDvhcID, EndpointCoDvhcID})
	require.NoError(t, err)
	assert.Equal(t, []string{EndpointCoDvhcID}, endpoints)

	_, err = normalizeOnly([]string{"bogus"})
	assert.Error(t, err)
}

func TestDedupeMalks(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, dedupeMalks([]string{"a", "", "b", "a", " "}))
}

func TestLoadMalksFromJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "donvi.json")
	require.NoError(t, os.WriteFile(path, []byte(`[
		{"ma":"01","malk":"a"},
		{"ma":"02","malk":"b"},
		{"ma":"03","malk":"a"}
	]`), 0o644))

	malks, err := LoadMalksFromJSON(path)
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, malks)
}

func TestValidateResponse(t *testing.T) {
	require.NoError(t, validateResponse(EndpointPreadJSON, []byte(`{"features":[{}]}`)))
	assert.Error(t, validateResponse(EndpointPreadJSON, []byte(`{"features":[]}`)))
	assert.Error(t, validateResponse(EndpointPreadJSON, []byte(`not json`)))

	require.NoError(t, validateResponse(EndpointCoDvhcID, []byte(`[{}]`)))
	assert.Error(t, validateResponse(EndpointCoDvhcID, []byte(`[]`)))
	assert.Error(t, validateResponse(EndpointCoDvhcID, []byte(`{}`)))
}

func TestWriteGzipRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "a.json.gz")
	content := []byte(`{"hello":"world"}`)
	require.NoError(t, writeGzip(path, content))

	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	gz, err := gzip.NewReader(f)
	require.NoError(t, err)
	defer gz.Close()

	got, err := io.ReadAll(gz)
	require.NoError(t, err)
	assert.Equal(t, content, got)
}

func TestRunCapturesAllEndpoints(t *testing.T) {
	var hits int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&hits, 1)
		switch r.URL.Path {
		case "/pread_json":
			fmt.Fprintf(w, `{"features":[{"id":%q}]}`, r.FormValue("id"))
		case "/p.co_dvhc_id":
			fmt.Fprintf(w, `[{"dientichkm2":"1","malk":%q}]`, r.FormValue("malk"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	dir := t.TempDir()
	manifest, err := Run(context.Background(), []string{"m1", "m2"}, Options{
		BaseURL: server.URL,
		OutDir:  dir,
		Workers: 2,
	})
	require.NoError(t, err)
	assert.Equal(t, 2, manifest.Endpoints[EndpointPreadJSON].Captured)
	assert.Equal(t, 2, manifest.Endpoints[EndpointCoDvhcID].Captured)
	assert.EqualValues(t, 4, atomic.LoadInt64(&hits))

	for _, endpoint := range []string{EndpointPreadJSON, EndpointCoDvhcID} {
		for _, malk := range []string{"m1", "m2"} {
			_, err := os.Stat(filepath.Join(dir, endpoint, malk+".json.gz"))
			assert.NoError(t, err)
		}
	}

	raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	require.NoError(t, err)
	var onDisk Manifest
	require.NoError(t, json.Unmarshal(raw, &onDisk))
	assert.Equal(t, server.URL, onDisk.SourceBaseURL)
}

func TestRunIsResumable(t *testing.T) {
	var hits int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&hits, 1)
		switch r.URL.Path {
		case "/pread_json":
			io.WriteString(w, `{"features":[{}]}`)
		case "/p.co_dvhc_id":
			io.WriteString(w, `[{"dientichkm2":"1"}]`)
		}
	}))
	defer server.Close()

	dir := t.TempDir()
	require.NoError(t, writeGzip(filepath.Join(dir, EndpointPreadJSON, "m1.json.gz"), []byte(`{"features":[{}]}`)))

	_, err := Run(context.Background(), []string{"m1"}, Options{
		BaseURL: server.URL,
		OutDir:  dir,
		Workers: 1,
	})
	require.NoError(t, err)
	assert.EqualValues(t, 1, atomic.LoadInt64(&hits), "only the missing metadata endpoint should be requested")
}

func TestRunReportsMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer server.Close()

	manifest, err := Run(context.Background(), []string{"m1"}, Options{
		BaseURL: server.URL,
		OutDir:  t.TempDir(),
		Workers: 1,
	})
	require.Error(t, err)
	require.NotNil(t, manifest)
	assert.Len(t, manifest.Endpoints[EndpointPreadJSON].Missing, 1)
	assert.Len(t, manifest.Endpoints[EndpointCoDvhcID].Missing, 1)
}

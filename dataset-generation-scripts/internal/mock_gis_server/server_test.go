package mock_gis_server

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, path string, content []byte, gzipped bool) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	if !gzipped {
		require.NoError(t, os.WriteFile(path, content, 0o644))
		return
	}

	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	require.NoError(t, err)
	_, err = gz.Write(content)
	require.NoError(t, err)
	require.NoError(t, gz.Close())
	require.NoError(t, os.WriteFile(path, buf.Bytes(), 0o644))
}

func newTestCache(t *testing.T) *Cache {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "pread_json", "m1.json.gz"), []byte(`{"features":[1]}`), true)
	writeFile(t, filepath.Join(dir, "pread_json", "m2.json"), []byte(`{"features":[2]}`), false)
	writeFile(t, filepath.Join(dir, "co_dvhc_id", "m1.json.gz"), []byte(`[{"dientichkm2":"1"}]`), true)
	writeFile(t, filepath.Join(dir, "co_dvhc_id", "ignored.txt"), []byte("x"), false)

	cache, err := LoadCache(dir)
	require.NoError(t, err)
	return cache
}

func TestLoadCache(t *testing.T) {
	cache := newTestCache(t)

	pread, metadata := cache.Counts()
	assert.Equal(t, 2, pread)
	assert.Equal(t, 1, metadata)

	_, ok := cache.PreadJSON("m1")
	assert.True(t, ok)
	_, ok = cache.PreadJSON("ignored")
	assert.False(t, ok)
}

func TestHandlerServesBothEndpoints(t *testing.T) {
	cache := newTestCache(t)
	server := httptest.NewServer(NewServer(cache, false).Handler())
	defer server.Close()

	t.Run("pread_json gzip", func(t *testing.T) {
		body, status := postForm(t, server.URL+"/pread_json", "id", "m1")
		assert.Equal(t, http.StatusOK, status)
		assert.JSONEq(t, `{"features":[1]}`, body)
	})

	t.Run("pread_json plain", func(t *testing.T) {
		body, status := postForm(t, server.URL+"/pread_json", "id", "m2")
		assert.Equal(t, http.StatusOK, status)
		assert.JSONEq(t, `{"features":[2]}`, body)
	})

	t.Run("co_dvhc_id", func(t *testing.T) {
		body, status := postForm(t, server.URL+"/p.co_dvhc_id", "malk", "m1")
		assert.Equal(t, http.StatusOK, status)
		assert.JSONEq(t, `[{"dientichkm2":"1"}]`, body)
	})
}

func TestHandlerErrors(t *testing.T) {
	cache := newTestCache(t)
	server := httptest.NewServer(NewServer(cache, false).Handler())
	defer server.Close()

	t.Run("unknown key returns 404", func(t *testing.T) {
		_, status := postForm(t, server.URL+"/pread_json", "id", "missing")
		assert.Equal(t, http.StatusNotFound, status)
	})

	t.Run("missing form field returns 400", func(t *testing.T) {
		_, status := postForm(t, server.URL+"/pread_json", "wrong", "m1")
		assert.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("GET returns 405", func(t *testing.T) {
		res, err := http.Get(server.URL + "/pread_json")
		require.NoError(t, err)
		defer res.Body.Close()
		assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
	})
}

func postForm(t *testing.T, endpoint, field, value string) (string, int) {
	t.Helper()
	res, err := http.PostForm(endpoint, url.Values{field: {value}})
	require.NoError(t, err)
	defer res.Body.Close()

	var buf bytes.Buffer
	_, err = buf.ReadFrom(res.Body)
	require.NoError(t, err)
	return buf.String(), res.StatusCode
}

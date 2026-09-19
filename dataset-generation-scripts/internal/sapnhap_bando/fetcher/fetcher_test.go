package fetcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGISServerBaseURL(t *testing.T) {
	t.Run("defaults to production server", func(t *testing.T) {
		t.Setenv("GIS_SERVER_BASE_URL", "")
		assert.Equal(t, DefaultGISServerBaseURL, GISServerBaseURL())
	})

	t.Run("honours override and trims trailing slash", func(t *testing.T) {
		t.Setenv("GIS_SERVER_BASE_URL", "http://127.0.0.1:18080/")
		assert.Equal(t, "http://127.0.0.1:18080", GISServerBaseURL())
	})

	t.Run("trims surrounding whitespace", func(t *testing.T) {
		t.Setenv("GIS_SERVER_BASE_URL", "  http://localhost:9999  ")
		assert.Equal(t, "http://localhost:9999", GISServerBaseURL())
	})
}

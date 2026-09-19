package mock_gis_server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Cache is an in-memory index of the captured GIS responses. Each entry maps a
// malk to the file that holds the raw response body.
type Cache struct {
	dir       string
	preadJSON map[string]string
	coDvhcID  map[string]string
}

// LoadCache scans the cache directory for the pread_json and co_dvhc_id
// endpoints. Both `.json.gz` and plain `.json` files are supported.
func LoadCache(dir string) (*Cache, error) {
	preadJSON, err := loadEndpointDir(filepath.Join(dir, "pread_json"))
	if err != nil {
		return nil, err
	}
	coDvhcID, err := loadEndpointDir(filepath.Join(dir, "co_dvhc_id"))
	if err != nil {
		return nil, err
	}
	return &Cache{
		dir:       dir,
		preadJSON: preadJSON,
		coDvhcID:  coDvhcID,
	}, nil
}

func loadEndpointDir(dir string) (map[string]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read cache dir %s: %w", dir, err)
	}

	result := make(map[string]string, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		key, ok := keyFromFilename(entry.Name())
		if !ok {
			continue
		}
		result[key] = filepath.Join(dir, entry.Name())
	}
	return result, nil
}

// keyFromFilename strips a supported extension, returning false for other files.
func keyFromFilename(name string) (string, bool) {
	switch {
	case strings.HasSuffix(name, ".json.gz"):
		return strings.TrimSuffix(name, ".json.gz"), true
	case strings.HasSuffix(name, ".json"):
		return strings.TrimSuffix(name, ".json"), true
	default:
		return "", false
	}
}

// PreadJSON returns the cache file path for a geometry response.
func (c *Cache) PreadJSON(malk string) (string, bool) {
	path, ok := c.preadJSON[malk]
	return path, ok
}

// CoDvhcID returns the cache file path for a metadata response.
func (c *Cache) CoDvhcID(malk string) (string, bool) {
	path, ok := c.coDvhcID[malk]
	return path, ok
}

// Counts returns the number of indexed geometry and metadata responses.
func (c *Cache) Counts() (preadJSON, coDvhcID int) {
	return len(c.preadJSON), len(c.coDvhcID)
}

// Dir returns the cache root directory.
func (c *Cache) Dir() string {
	return c.dir
}

package dataset_metadata

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseValidContent(t *testing.T) {
	metadata, err := Parse("dataset_version=v5.1.0\nlatest_decree=30/2026/QH16\n")
	require.NoError(t, err)

	assert.Equal(t, "v5.1.0", metadata.DatasetVersion)
	assert.Equal(t, "30/2026/QH16", metadata.LatestDecree)
	assert.Equal(t, time.UTC, metadata.GeneratedAt.Location())
	assert.WithinDuration(t, time.Now().UTC(), metadata.GeneratedAt, time.Minute)
}

func TestParseIgnoresCommentsAndBlankLines(t *testing.T) {
	content := "# comment line\n\n  dataset_version = v5.1.0  \n\n# another comment\nlatest_decree=30/2026/QH16\n"
	metadata, err := Parse(content)
	require.NoError(t, err)

	assert.Equal(t, "v5.1.0", metadata.DatasetVersion)
	assert.Equal(t, "30/2026/QH16", metadata.LatestDecree)
}

func TestParseIgnoresUnknownKeys(t *testing.T) {
	metadata, err := Parse("dataset_version=v5.1.0\nfuture_key=whatever\n")
	require.NoError(t, err)

	assert.Equal(t, "v5.1.0", metadata.DatasetVersion)
	assert.Empty(t, metadata.LatestDecree)
}

func TestParseAllowsEmptyLatestDecree(t *testing.T) {
	metadata, err := Parse("dataset_version=v5.1.0\n")
	require.NoError(t, err)

	assert.Equal(t, "v5.1.0", metadata.DatasetVersion)
	assert.Empty(t, metadata.LatestDecree)
}

func TestParseMissingDatasetVersion(t *testing.T) {
	_, err := Parse("latest_decree=30/2026/QH16\n")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "dataset_version")
}

func TestParseMalformedLine(t *testing.T) {
	_, err := Parse("dataset_version v5.1.0\n")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "key=value")
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "version.txt")
	require.NoError(t, os.WriteFile(path, []byte("dataset_version=v5.1.0\nlatest_decree=30/2026/QH16\n"), 0644))

	metadata, err := LoadFromFile(path)
	require.NoError(t, err)

	assert.Equal(t, "v5.1.0", metadata.DatasetVersion)
	assert.Equal(t, "30/2026/QH16", metadata.LatestDecree)
}

func TestLoadFromFileMissing(t *testing.T) {
	_, err := LoadFromFile(filepath.Join(t.TempDir(), "does-not-exist.txt"))
	require.Error(t, err)
}

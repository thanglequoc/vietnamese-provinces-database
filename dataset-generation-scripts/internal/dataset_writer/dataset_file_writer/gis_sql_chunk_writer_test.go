package dataset_writer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteChunkedSQLFile_SinglePart(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "mysql_ImportData_gis_2026-08-10__21_55_01.sql")

	header := chunkHeaderInfo{
		Banner:     "Add-on GIS Dataset for MySQL of Vietnamese Provinces Database",
		CreatedAt:  "Mon, 10 Aug 2026 21:55:01 +0700",
		Repository: "https://github.com/thanglequoc/vietnamese-provinces-database",
	}
	blocks := [][]byte{
		[]byte("-- DATA for gis_provinces --\n"),
		[]byte("INSERT INTO gis_provinces(province_code, gis_server_id) VALUES ('01','x');\n"),
	}

	err := writeChunkedSQLFile(path, blocks, header)
	assert.NoError(t, err)

	// Exactly one part + manifest, header rendered on the part
	partPath := filepath.Join(tmpDir, "mysql_ImportData_gis_2026-08-10__21_55_01-part-01.sql")
	content, err := os.ReadFile(partPath)
	assert.NoError(t, err)
	contentStr := string(content)
	assert.Contains(t, contentStr, "/* === Add-on GIS Dataset for MySQL of Vietnamese Provinces Database === */")
	assert.Contains(t, contentStr, "/* Part 1 of 1 */")
	assert.Contains(t, contentStr, "/* Created at:  Mon, 10 Aug 2026 21:55:01 +0700 */")
	assert.Contains(t, contentStr, "/* Reference: https://github.com/thanglequoc/vietnamese-provinces-database */")
	assert.Contains(t, contentStr, "INSERT INTO gis_provinces(province_code, gis_server_id) VALUES ('01','x');")

	// No single file at path
	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err), "single file at path should NOT exist")

	// Manifest lists the single part
	manifestData, err := os.ReadFile(path + ".manifest")
	assert.NoError(t, err)
	assert.Equal(t, "mysql_ImportData_gis_2026-08-10__21_55_01-part-01.sql\n", string(manifestData))
}

func TestWriteChunkedSQLFile_MultipleParts(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "postgresql_ImportData_gis_2026-08-10__21_55_01.sql")

	originalMax := maxSQLGISChunkSize
	maxSQLGISChunkSize = 5000
	defer func() { maxSQLGISChunkSize = originalMax }()

	header := chunkHeaderInfo{
		Banner:     "Add-on GIS Dataset for PostgreSQL of Vietnamese Provinces Database",
		CreatedAt:  "Mon, 10 Aug 2026 21:55:01 +0700",
		Repository: "https://github.com/thanglequoc/vietnamese-provinces-database",
	}

	// 10 blocks of ~1.8 KB each; with a 5000-byte limit, blocks pack 2-per-chunk
	var blocks [][]byte
	var expected string
	for i := 0; i < 10; i++ {
		block := []byte(strings.Repeat("INSERT INTO gis_wards(ward_code) VALUES ('x');\n", 40))
		blocks = append(blocks, block)
		expected += string(block)
	}

	err := writeChunkedSQLFile(path, blocks, header)
	assert.NoError(t, err)

	// No single file at path
	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err), "single file at path should NOT exist")

	manifestData, err := os.ReadFile(path + ".manifest")
	assert.NoError(t, err)
	manifestLines := strings.Split(strings.TrimSpace(string(manifestData)), "\n")
	assert.True(t, len(manifestLines) >= 2, "expected at least 2 parts, got %d", len(manifestLines))

	var body strings.Builder
	blockCount := 0
	for i, name := range manifestLines {
		partData, err := os.ReadFile(filepath.Join(tmpDir, name))
		assert.NoError(t, err)
		assert.LessOrEqual(t, len(partData), maxSQLGISChunkSize, "part %s exceeds chunk limit", name)

		partStr := string(partData)
		assert.Contains(t, partStr, fmt.Sprintf("/* Part %d of %d */", i+1, len(manifestLines)))
		assert.Contains(t, partStr, "/* Reference: https://github.com/thanglequoc/vietnamese-provinces-database */")

		// Strip the header (everything up to the first blank line) and collect the body
		idx := strings.Index(partStr, "\n\n")
		assert.NotEqual(t, -1, idx, "part %s missing header/body separator", name)
		body.WriteString(partStr[idx+2:])
		blockCount += strings.Count(partStr[idx+2:], "INSERT INTO gis_wards(ward_code)")
	}

	assert.Equal(t, expected, body.String())
	assert.Equal(t, len(blocks)*40, blockCount, "every block's INSERT lines should be present")
}

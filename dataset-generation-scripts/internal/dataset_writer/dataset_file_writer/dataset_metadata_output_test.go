package dataset_writer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	datasetmetadata "github.com/thanglequoc-vn-provinces/v2/internal/dataset_metadata"
)

var testDatasetMetadata = datasetmetadata.Metadata{
	DatasetVersion: "v5.1.0",
	LatestDecree:   "30/2026/QH16",
	GeneratedAt:    time.Date(2026, 9, 12, 10, 0, 0, 0, time.UTC),
}

func TestPostgresMySQLDatasetFileWriter_WriteMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	writer := &PostgresMySQLDatasetFileWriter{
		OutputFilePath: filepath.Join(tmpDir, "postgres_ImportData_vn_units.sql"),
		Metadata:       testDatasetMetadata,
	}

	require.NoError(t, writer.WriteToFile(nil, nil, nil, nil))

	content := readGeneratedSQLFile(t, tmpDir)
	assert.Contains(t, content,
		"INSERT INTO dataset_metadata(dataset_version,latest_decree,generated_at) VALUES('v5.1.0','30/2026/QH16','2026-09-12 10:00:00');")
}

func TestMssqlDatasetFileWriter_WriteMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	writer := &MssqlDatasetFileWriter{
		OutputFilePath: filepath.Join(tmpDir, "mssql_ImportData_vn_units.sql"),
		Metadata:       testDatasetMetadata,
	}

	require.NoError(t, writer.WriteToFile(nil, nil, nil, nil))

	content := readGeneratedSQLFile(t, tmpDir)
	assert.Contains(t, content,
		"INSERT INTO dataset_metadata(dataset_version,latest_decree,generated_at) VALUES(N'v5.1.0',N'30/2026/QH16','2026-09-12 10:00:00');")
}

func TestOracleDatasetFileWriter_WriteMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	writer := &OracleDatasetFileWriter{
		OutputFilePath: filepath.Join(tmpDir, "oracle_ImportData_vn_units.sql"),
		Metadata:       testDatasetMetadata,
	}

	require.NoError(t, writer.WriteToFile(nil, nil, nil, nil))

	content := readGeneratedSQLFile(t, tmpDir)
	assert.Contains(t, content, "INSERT INTO dataset_metadata(dataset_version,latest_decree,generated_at)")
	assert.Contains(t, content, "TO_TIMESTAMP('2026-09-12 10:00:00','YYYY-MM-DD HH24:MI:SS')")
}

func TestJSONDatasetFileWriter_WriteMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	writer := &JSONDatasetFileWriter{
		OutputFolderPath: tmpDir,
		Metadata:         testDatasetMetadata,
	}

	require.NoError(t, writer.WriteToFile(nil, nil, nil, nil))

	content, err := os.ReadFile(filepath.Join(tmpDir, "metadata.json"))
	require.NoError(t, err)
	s := string(content)
	assert.Contains(t, s, `"DatasetVersion": "v5.1.0"`)
	assert.Contains(t, s, `"LatestDecree": "30/2026/QH16"`)
	assert.Contains(t, s, `"GeneratedAt": "2026-09-12T10:00:00Z"`)
}

func TestMongoDBDatasetFileWriter_WriteMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	writer := &MongoDBDatasetFileWriter{
		OutputFolderPath: tmpDir,
		Metadata:         testDatasetMetadata,
	}

	require.NoError(t, writer.WriteToFile(nil, nil, nil, nil))

	content, err := os.ReadFile(filepath.Join(tmpDir, "mongo_data_vn_metadata.json"))
	require.NoError(t, err)
	s := string(content)
	assert.Contains(t, s, `"DatasetVersion": "v5.1.0"`)
	assert.Contains(t, s, `"LatestDecree": "30/2026/QH16"`)
	assert.Contains(t, s, `"GeneratedAt": "2026-09-12T10:00:00Z"`)
}

func TestRedisDatasetFileWriter_WriteMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	writer := &RedisDatasetFileWriter{
		OutputFolderPath: tmpDir,
		Metadata:         testDatasetMetadata,
	}

	require.NoError(t, writer.WriteToFile(nil, nil, nil, nil))

	entries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	var datasetFile string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".redis") {
			datasetFile = entry.Name()
			break
		}
	}
	require.NotEmpty(t, datasetFile)

	content, err := os.ReadFile(filepath.Join(tmpDir, datasetFile))
	require.NoError(t, err)
	assert.Contains(t, string(content),
		`HSET dataset_metadata datasetVersion "v5.1.0" latestDecree "30/2026/QH16" generatedAt "2026-09-12T10:00:00Z"`)
}

func TestElasticsearchDatasetFileWriter_WriteMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	writer := &ElasticsearchDatasetFileWriter{
		OutputFolderPath: tmpDir,
		Metadata:         testDatasetMetadata,
	}

	require.NoError(t, writer.WriteToFile(nil, nil, nil, nil))

	ndjson, err := os.ReadFile(filepath.Join(tmpDir, "dataset_metadata.ndjson"))
	require.NoError(t, err)
	s := string(ndjson)
	assert.Contains(t, s, `"_index":"dataset_metadata"`)
	assert.Contains(t, s, `"DatasetVersion":"v5.1.0"`)
	assert.Contains(t, s, `"GeneratedAt":"2026-09-12T10:00:00Z"`)

	mapping, err := os.ReadFile(filepath.Join(tmpDir, "mappings", "dataset_metadata.json"))
	require.NoError(t, err)
	assert.Contains(t, string(mapping), `"GeneratedAt"`)
	assert.Contains(t, string(mapping), `"date"`)
}

func TestWriters_EmptyMetadataProducesNoArtifacts(t *testing.T) {
	t.Run("postgres sql", func(t *testing.T) {
		tmpDir := t.TempDir()
		writer := &PostgresMySQLDatasetFileWriter{OutputFilePath: filepath.Join(tmpDir, "postgres_ImportData_vn_units.sql")}
		require.NoError(t, writer.WriteToFile(nil, nil, nil, nil))
		assert.NotContains(t, readGeneratedSQLFile(t, tmpDir), "dataset_metadata")
	})

	t.Run("json", func(t *testing.T) {
		tmpDir := t.TempDir()
		writer := &JSONDatasetFileWriter{OutputFolderPath: tmpDir}
		require.NoError(t, writer.WriteToFile(nil, nil, nil, nil))
		_, err := os.Stat(filepath.Join(tmpDir, "metadata.json"))
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("mongodb", func(t *testing.T) {
		tmpDir := t.TempDir()
		writer := &MongoDBDatasetFileWriter{OutputFolderPath: tmpDir}
		require.NoError(t, writer.WriteToFile(nil, nil, nil, nil))
		_, err := os.Stat(filepath.Join(tmpDir, "mongo_data_vn_metadata.json"))
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("redis", func(t *testing.T) {
		tmpDir := t.TempDir()
		writer := &RedisDatasetFileWriter{OutputFolderPath: tmpDir}
		require.NoError(t, writer.WriteToFile(nil, nil, nil, nil))
		entries, err := os.ReadDir(tmpDir)
		require.NoError(t, err)
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".redis") {
				content, err := os.ReadFile(filepath.Join(tmpDir, entry.Name()))
				require.NoError(t, err)
				assert.NotContains(t, string(content), "dataset_metadata")
			}
		}
	})

	t.Run("elasticsearch", func(t *testing.T) {
		tmpDir := t.TempDir()
		writer := &ElasticsearchDatasetFileWriter{OutputFolderPath: tmpDir}
		require.NoError(t, writer.WriteToFile(nil, nil, nil, nil))
		_, err := os.Stat(filepath.Join(tmpDir, "dataset_metadata.ndjson"))
		assert.True(t, os.IsNotExist(err))
	})
}

package dataset_writer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	vn_provinces_tmp_model "github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
)

func TestOracleDatasetFileWriter_WriteToFile_README(t *testing.T) {
	tmpDir := t.TempDir()
	writer := &OracleDatasetFileWriter{
		OutputFilePath: filepath.Join(tmpDir, "oracle_ImportData_vn_units.sql"),
	}
	provinces := []vn_provinces_tmp_model.Province{{Code: "01", Name: "Hà Nội", AdministrativeUnitId: 1}}

	err := writer.WriteToFile(nil, nil, provinces, nil)
	assert.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "README.md"))
	assert.NoError(t, err)
	s := string(content)
	assert.Contains(t, s, "**Generated at:")
	assert.Contains(t, s, "oracle_ImportData_vn_units.sql")
	assert.Contains(t, s, "## Data Structure")
	assert.Contains(t, s, "## Sample Queries")
}

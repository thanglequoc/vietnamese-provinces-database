package decree_check

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBumpMinor(t *testing.T) {
	cases := map[string]string{
		"v5.1.0":  "v5.2.0",
		"v1.0.9":  "v1.1.0",
		"v2.3.99": "v2.4.0",
		" v5.1.0": "v5.2.0",
	}
	for input, expected := range cases {
		actual, err := BumpMinor(input)
		require.NoError(t, err, "input %q", input)
		assert.Equal(t, expected, actual)
	}
}

func TestBumpMinorRejectsInvalidVersions(t *testing.T) {
	for _, input := range []string{"", "5.1.0", "v5.1", "v5", "vx.y.z", "v5.1.0-rc1"} {
		_, err := BumpMinor(input)
		require.Error(t, err, "input %q should be rejected", input)
	}
}

func TestUpdateVersionFilePreservesCommentsAndOrder(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "version.txt")
	original := "# header comment\n# second comment\ndataset_version=v5.1.0\nlatest_decree=36/2026/QH16\n"
	require.NoError(t, os.WriteFile(path, []byte(original), 0o644))

	require.NoError(t, UpdateVersionFile(path, "v5.2.0", "388/NQ-UBTVQH16"))

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t,
		"# header comment\n# second comment\ndataset_version=v5.2.0\nlatest_decree=388/NQ-UBTVQH16\n",
		string(content),
	)
}

func TestUpdateVersionFileErrorsOnMissingKeys(t *testing.T) {
	dir := t.TempDir()

	missingDecree := filepath.Join(dir, "missing_decree.txt")
	require.NoError(t, os.WriteFile(missingDecree, []byte("dataset_version=v5.1.0\n"), 0o644))
	require.Error(t, UpdateVersionFile(missingDecree, "v5.2.0", "388/NQ-UBTVQH16"))

	missingVersion := filepath.Join(dir, "missing_version.txt")
	require.NoError(t, os.WriteFile(missingVersion, []byte("latest_decree=36/2026/QH16\n"), 0o644))
	require.Error(t, UpdateVersionFile(missingVersion, "v5.2.0", "388/NQ-UBTVQH16"))
}

func TestUpdateVersionFileErrorsWhenFileMissing(t *testing.T) {
	err := UpdateVersionFile(filepath.Join(t.TempDir(), "nope.txt"), "v5.2.0", "388/NQ-UBTVQH16")
	require.Error(t, err)
}

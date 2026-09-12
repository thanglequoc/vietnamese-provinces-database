package dataset_metadata

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// DefaultVersionFilePath is the version source file, relative to the
// dataset-generation-scripts working directory.
const DefaultVersionFilePath = "./version.txt"

// Metadata describes the version of the generated dataset.
type Metadata struct {
	DatasetVersion string
	LatestDecree   string
	GeneratedAt    time.Time
}

// Parse reads the key=value version source content. Blank lines and lines
// starting with '#' are ignored. Unknown keys are skipped so the file can be
// extended without breaking older generators.
func Parse(content string) (Metadata, error) {
	metadata := Metadata{GeneratedAt: time.Now().UTC()}

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			return Metadata{}, fmt.Errorf("invalid version file line (expected key=value): %q", line)
		}

		switch strings.TrimSpace(key) {
		case "dataset_version":
			metadata.DatasetVersion = strings.TrimSpace(value)
		case "latest_decree":
			metadata.LatestDecree = strings.TrimSpace(value)
		}
	}
	if err := scanner.Err(); err != nil {
		return Metadata{}, fmt.Errorf("read version file content: %w", err)
	}

	if metadata.DatasetVersion == "" {
		return Metadata{}, fmt.Errorf("version file is missing required key: dataset_version")
	}

	return metadata, nil
}

// LoadFromFile reads and parses the version source at the given path.
func LoadFromFile(path string) (Metadata, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Metadata{}, fmt.Errorf("read version file %s: %w", path, err)
	}
	return Parse(string(content))
}

// Load reads and parses the default version source file.
func Load() (Metadata, error) {
	return LoadFromFile(DefaultVersionFilePath)
}

// IsEmpty reports whether no dataset version has been set.
func (m Metadata) IsEmpty() bool {
	return m.DatasetVersion == ""
}

// GeneratedAtSQL formats the generation timestamp as a SQL datetime literal (UTC).
func (m Metadata) GeneratedAtSQL() string {
	return m.GeneratedAt.UTC().Format("2006-01-02 15:04:05")
}

// GeneratedAtRFC3339 formats the generation timestamp as RFC 3339 (UTC).
func (m Metadata) GeneratedAtRFC3339() string {
	return m.GeneratedAt.UTC().Format(time.RFC3339)
}

package decree_check

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var datasetVersionPattern = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)

// BumpMinor increments the minor (middle) component of a vX.Y.Z dataset version
// and resets the patch component to zero (e.g. v5.1.0 -> v5.2.0).
func BumpMinor(version string) (string, error) {
	trimmed := strings.TrimSpace(version)
	matches := datasetVersionPattern.FindStringSubmatch(trimmed)
	if matches == nil {
		return "", fmt.Errorf("invalid dataset_version %q: expected vX.Y.Z", version)
	}

	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return "", fmt.Errorf("invalid minor component in %q: %w", version, err)
	}

	return fmt.Sprintf("v%s.%d.0", matches[1], minor+1), nil
}

// UpdateVersionFile rewrites dataset_version and latest_decree in the version
// source file while preserving comments, blank lines and key order.
func UpdateVersionFile(path, newVersion, newDecree string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read version file %s: %w", path, err)
	}

	lines := strings.Split(string(content), "\n")
	var foundVersion, foundDecree bool

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		key, _, found := strings.Cut(trimmed, "=")
		if !found {
			continue
		}

		switch strings.TrimSpace(key) {
		case "dataset_version":
			lines[i] = "dataset_version=" + newVersion
			foundVersion = true
		case "latest_decree":
			lines[i] = "latest_decree=" + newDecree
			foundDecree = true
		}
	}

	if !foundVersion {
		return fmt.Errorf("version file %s is missing dataset_version", path)
	}
	if !foundDecree {
		return fmt.Errorf("version file %s is missing latest_decree", path)
	}

	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		return fmt.Errorf("write version file %s: %w", path, err)
	}
	return nil
}

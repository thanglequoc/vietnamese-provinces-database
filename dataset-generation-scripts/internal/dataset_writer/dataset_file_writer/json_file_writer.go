package dataset_writer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
	file_writer_helper "github.com/thanglequoc-vn-provinces/v2/internal/dataset_writer/dataset_file_writer/helper"
)

type JSONDatasetFileWriter struct {
	OutputFolderPath string
}

func (w *JSONDatasetFileWriter) WriteToFile(
	regions []model.AdministrativeRegion,
	administrativeUnits []model.AdministrativeUnit,
	provinces []model.Province,
	wards []model.Ward) error {

	os.MkdirAll(w.OutputFolderPath, 0746)

	provinceData := file_writer_helper.ConvertToJsonProvinceModel(provinces)
	if err := writePrettyJSON(filepath.Join(w.OutputFolderPath, "full_json_generated_data_vn_units.json"), provinceData); err != nil {
		return err
	}

	// JSON Simplified file
	provinceSimplifiedData := file_writer_helper.ConvertToJsonProvinceSimplifiedModel(provinces)
	if err := writePrettyJSON(filepath.Join(w.OutputFolderPath, "simplified_json_generated_data_vn_units.json"), provinceSimplifiedData); err != nil {
		return err
	}
	// JSON Simplified minified file
	if err := writeMinifiedJSON(filepath.Join(w.OutputFolderPath, "simplified_json_generated_data_vn_units_minified.json"), provinceSimplifiedData); err != nil {
		return err
	}

	// VN only JSON Simplified file
	provinceVNSimplifiedData := file_writer_helper.ConvertToJsonProvinceVNSimplifiedModel(provinces)
	if err := writePrettyJSON(filepath.Join(w.OutputFolderPath, "vn_only_simplified_json_generated_data_vn_units.json"), provinceVNSimplifiedData); err != nil {
		return err
	}
	// VN only JSON Simplified minified file
	if err := writeMinifiedJSON(filepath.Join(w.OutputFolderPath, "vn_only_simplified_json_generated_data_vn_units_minified.json"), provinceVNSimplifiedData); err != nil {
		return err
	}

	return writeJSONDatasetReadme(w.OutputFolderPath)
}

// writePrettyJSON marshals payload with a single-space indent and writes it to filePath.
func writePrettyJSON(filePath string, payload any) error {
	data, err := json.MarshalIndent(payload, "", " ")
	if err != nil {
		return fmt.Errorf("marshal pretty json payload for %s: %w", filePath, err)
	}
	return os.WriteFile(filePath, data, 0644)
}

// writeMinifiedJSON marshals payload compactly (no whitespace) and writes it to filePath.
func writeMinifiedJSON(filePath string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal minified json payload for %s: %w", filePath, err)
	}
	return os.WriteFile(filePath, data, 0644)
}

// writeJSONDatasetReadme writes a README describing the JSON dataset files and the
// optional geojson artifacts, with a bold generation timestamp at the top.
func writeJSONDatasetReadme(outputFolderPath string) error {
	return writeDatasetReadme(outputFolderPath,
		"JSON Dataset — Vietnamese Provinces Database",
		"Administrative unit JSON data for Vietnam: provinces with embedded wards, in full and simplified forms.",
		[]DatasetReadmeFile{
			{Name: "full_json_generated_data_vn_units.json", Description: "Full dataset (provinces + wards + administrative info)"},
			{Name: "simplified_json_generated_data_vn_units.json", Description: "Simplified dataset (pretty-printed)"},
			{Name: "simplified_json_generated_data_vn_units_minified.json", Description: "Simplified dataset (minified)"},
			{Name: "vn_only_simplified_json_generated_data_vn_units.json", Description: "Vietnamese-only simplified (pretty-printed)"},
			{Name: "vn_only_simplified_json_generated_data_vn_units_minified.json", Description: "Vietnamese-only simplified (minified)"},
		},
		[]string{
			"## Overview",
			"",
			"| File | Contents |",
			"|------|----------|",
			"| `full_json_generated_data_vn_units.json` | 34 province objects with full administrative info and embedded wards |",
			"| `simplified_json_generated_data_vn_units.json` | Same structure, fewer fields (pretty-printed) |",
			"| `simplified_json_generated_data_vn_units_minified.json` | Simplified, minified (no whitespace) |",
			"| `vn_only_simplified_json_generated_data_vn_units.json` | Vietnamese-only fields (pretty-printed) |",
			"| `vn_only_simplified_json_generated_data_vn_units_minified.json` | Vietnamese-only fields (minified) |",
			"",
			"## Data Structure",
			"",
			"Each entry is a province object:",
			"",
			"- **`code`** — province code (`01`)",
			"- **`name` / `nameEn`** — province name",
			"- **`fullName` / `fullNameEn`** — full name with unit prefix",
			"- **`codeName`** — code name (`ha_noi`)",
			"- **`administrativeUnitId` / `administrativeUnitShortName` / `administrativeUnitFullName`** — unit type",
			"- **`postalCodePrefix`** — comma-separated 2-digit postal prefixes",
			"- **`wards`** — array of ward objects (`code`, `name`, `nameEn`, `fullName`, `fullNameEn`, `codeName`, `provinceCode`, `postalCode`, unit fields)",
			"",
			"## Sample Document",
			"",
			"```json",
			"[",
			"  {",
			"    \"code\": \"01\",",
			"    \"name\": \"Hà Nội\",",
			"    \"nameEn\": \"Ha Noi\",",
			"    \"fullName\": \"Thành phố Hà Nội\",",
			"    \"fullNameEn\": \"Ha Noi City\",",
			"    \"codeName\": \"ha_noi\",",
			"    \"postalCodePrefix\": \"10, 11, 12, 13, 14\",",
			"    \"wards\": [",
			"      { \"code\": \"00004\", \"name\": \"Ba Đình\", \"nameEn\": \"Ba Dinh\", \"postalCode\": \"11120\" }",
			"    ]",
			"  }",
			"]",
			"```",
			"",
			"## Quick Start",
			"",
			"```js",
			"const dataset = require('./full_json_generated_data_vn_units.json');",
			"```",
			"",
			"```python",
			"import json",
			"with open('full_json_generated_data_vn_units.json', encoding='utf-8') as f:",
			"    dataset = json.load(f)",
			"```",
			"",
			"## Sample Queries",
			"",
			"```js",
			"// Find Hà Nội",
			"dataset.find(p => p.code === '01');",
			"",
			"// Wards with postal code 11024",
			"dataset.flatMap(p => p.wards).filter(w => w.postalCode === '11024');",
			"",
			"// All province names",
			"dataset.map(p => p.name);",
			"```",
			"",
			"## GIS / GeoJSON",
			"",
			"The `geojson/` subfolder contains per-province and per-ward GeoJSON boundary exports, and `vn_provinces_wards_geojson.zip` is the combined archive of those files. These artifacts are present when the GIS generation step runs.",
		})
}

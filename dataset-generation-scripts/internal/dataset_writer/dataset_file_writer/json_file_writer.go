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
		"Administrative unit JSON data for Vietnam: provinces with embedded wards.",
		[]DatasetReadmeFile{
			{Name: "full_json_generated_data_vn_units.json", Description: "Full dataset (provinces + wards + administrative info)"},
			{Name: "simplified_json_generated_data_vn_units.json", Description: "Simplified dataset (pretty-printed)"},
			{Name: "simplified_json_generated_data_vn_units_minified.json", Description: "Simplified dataset (minified)"},
			{Name: "vn_only_simplified_json_generated_data_vn_units.json", Description: "Vietnamese-only simplified (pretty-printed)"},
			{Name: "vn_only_simplified_json_generated_data_vn_units_minified.json", Description: "Vietnamese-only simplified (minified)"},
		},
		[]string{
			"## Data Structure",
			"",
			"Each entry is a province object with `code`, `name`, `nameEn`, `fullName`, `fullNameEn`, `codeName`, `administrativeUnit*`, `postalCodePrefix`, and a `wards` array of ward objects.",
			"",
			"## Sample Queries",
			"",
			"```js",
			"const dataset = require('./full_json_generated_data_vn_units.json');",
			"dataset.find(p => p.code === '01');",
			"dataset.flatMap(p => p.wards).filter(w => w.postalCode === '11024');",
			"```",
			"",
			"## GIS / GeoJSON",
			"",
			"The `geojson/` subfolder contains per-province and per-ward GeoJSON boundary exports, and `vn_provinces_wards_geojson.zip` is the combined archive of those files. These artifacts are present when the GIS generation step runs.",
		})
}

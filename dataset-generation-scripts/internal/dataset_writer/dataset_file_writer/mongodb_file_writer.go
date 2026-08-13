package dataset_writer

import (
	"bufio"
	"encoding/json"
	"fmt"
	file_writer_helper "github.com/thanglequoc-vn-provinces/v2/internal/dataset_writer/dataset_file_writer/helper"
	"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
	"os"
)

type MongoDBDatasetFileWriter struct {
	OutputFolderPath string
}

func (w *MongoDBDatasetFileWriter) WriteToFile(
	regions []model.AdministrativeRegion,
	administrativeUnits []model.AdministrativeUnit,
	provinces []model.Province,
	wards []model.Ward) error {

	os.MkdirAll(w.OutputFolderPath, 0746)

	// Write file administrative units
	administrativeUnitsFilePath := fmt.Sprintf("%s/administrative_units.json", w.OutputFolderPath)
	administrativeUnitsFile, err := os.OpenFile(administrativeUnitsFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	dataWriter := bufio.NewWriter(administrativeUnitsFile)
	data, _ := json.MarshalIndent(administrativeUnits, "", " ")
	dataWriter.Write(data)
	dataWriter.Flush()
	administrativeUnitsFile.Close()

	// Write file administrative regions
	administrativeRegionsFilePath := fmt.Sprintf("%s/administrative_regions.json", w.OutputFolderPath)
	administrativeRegionsFile, err := os.OpenFile(administrativeRegionsFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	dataWriter = bufio.NewWriter(administrativeRegionsFile)
	data, _ = json.MarshalIndent(regions, "", " ")
	dataWriter.Write(data)
	dataWriter.Flush()
	administrativeRegionsFile.Close()

	// Write file to provinces (complete) data
	dataProvinceMongoPath := fmt.Sprintf("%s/mongo_data_vn_unit.json", w.OutputFolderPath)
	dataProvinceMongoFile, err := os.OpenFile(dataProvinceMongoPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	dataWriter = bufio.NewWriter(dataProvinceMongoFile)
	provinceData := file_writer_helper.ConvertToMongoProvinceModel(provinces)

	data, _ = json.MarshalIndent(provinceData, "", " ")
	dataWriter.Write(data)
	dataWriter.Flush()
	dataProvinceMongoFile.Close()
	return writeMongoReadme(w.OutputFolderPath)
}

func writeMongoReadme(outputFolderPath string) error {
	return writeDatasetReadme(outputFolderPath,
		"MongoDB Dataset — Vietnamese Provinces Database",
		"MongoDB documents for Vietnamese provinces with embedded wards.",
		[]DatasetReadmeFile{
			{Name: "administrative_units.json", Description: "Array of 8 administrative unit types"},
			{Name: "administrative_regions.json", Description: "Array of 8 regions"},
			{Name: "mongo_data_vn_unit.json", Description: "Array of 34 province documents, each embedding its Wards array"},
		},
		[]string{
			"## Data Structure",
			"",
			"- `administrative_units.json` — array of 8 administrative unit types",
			"- `administrative_regions.json` — array of 8 regions",
			"- `mongo_data_vn_unit.json` — the `provinces` collection: 34 province documents, each with an embedded `Wards` array",
			"",
			"## Sample Queries",
			"",
			"```javascript",
			"// Count provinces",
			"db.getCollection('provinces').countDocuments();",
			"",
			"// Wards of a province",
			"db.getCollection('provinces').findOne({Code: '01'}, {Name: 1, Wards: 1});",
			"```",
			"",
			"## GIS / GeoJSON",
			"",
			"The `gis/` subfolder contains the `provinces-gis` (34) and `wards-gis` (3,321) collections (`mongo_data_vn_province_gis.json`, `mongo_data_vn_ward_gis[_part_NN].json`), the `create_indexes.js` index script, and a `.manifest`. Import them, run `create_indexes.js`, then query with `$geoIntersects`.",
		})
}

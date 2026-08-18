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
		"MongoDB documents for Vietnamese provinces and wards, with an optional GIS add-on.",
		[]DatasetReadmeFile{
			{Name: "administrative_units.json", Description: "Array of 8 administrative unit types"},
			{Name: "administrative_regions.json", Description: "Array of 8 regions"},
			{Name: "mongo_data_vn_unit.json", Description: "Array of 34 province documents, each embedding its Wards array"},
		},
		[]string{
			"## Overview",
			"",
			"| Collection | Documents | Description |",
			"|------------|-----------|-------------|",
			"| `provinces` | 34 | Province documents with embedded wards (`mongo_data_vn_unit.json`) |",
			"| `provinces-gis` | 34 | GIS add-on: province documents with geometry |",
			"| `wards-gis` | 3,321 | GIS add-on: standalone ward documents with geometry |",
			"",
			"`administrative_units.json` and `administrative_regions.json` provide the lookup data (8 each).",
			"",
			"## Data Structure",
			"",
			"A province document in the `provinces` collection:",
			"",
			"- **`Code`** — province code (`01`)",
			"- **`Name` / `NameEn`** — province name",
			"- **`FullName` / `FullNameEn`** — full name with unit prefix",
			"- **`CodeName`** — code name (`ha_noi`)",
			"- **`AdministrativeUnit`** — embedded unit object (Id, FullName, ShortName, ...)",
			"- **`SearchKeywords`** — pre-computed autocomplete keywords (code, tone-stripped name, English name, codeName)",
			"- **`Wards`** — embedded array of ward documents (same field shape, plus `PostalCode` and `ProvinceCode`)",
			"",
			"The GIS collections add a **`GIS`** object: `Center` (GeoJSON Point), `BoundingBox`, `Geometry` (GeoJSON MultiPolygon/Polygon), and `Properties`.",
			"",
			"## Sample Document",
			"",
			"```json",
			"{",
			"  \"Code\": \"01\",",
			"  \"Name\": \"Hà Nội\",",
			"  \"NameEn\": \"Ha Noi\",",
			"  \"FullName\": \"Thành phố Hà Nội\",",
			"  \"FullNameEn\": \"Ha Noi City\",",
			"  \"CodeName\": \"ha_noi\",",
			"  \"AdministrativeUnit\": { \"Id\": 1, \"FullName\": \"Thành phố trực thuộc trung ương\", \"ShortName\": \"Thành phố\" },",
			"  \"SearchKeywords\": [\"01\", \"ha noi\", \"ha_noi\"],",
			"  \"Wards\": [",
			"    { \"Code\": \"00004\", \"Name\": \"Ba Đình\", \"FullName\": \"Phường Ba Đình\", \"PostalCode\": \"11120\" }",
			"  ]",
			"}",
			"```",
			"",
			"## Quick Start",
			"",
			"1. Import the data:",
			"",
			"```bash",
			"mongoimport --db vn_provinces --collection provinces --file mongo_data_vn_unit.json --jsonArray",
			"mongoimport --db vn_provinces --collection administrative_units --file administrative_units.json --jsonArray",
			"mongoimport --db vn_provinces --collection administrative_regions --file administrative_regions.json --jsonArray",
			"```",
			"",
			"2. GIS add-on (optional): import each file in `gis/` (ward files may be chunked — follow the `.manifest`), then run `mongosh vn_provinces create_indexes.js`.",
			"",
			"## Sample Queries",
			"",
			"```javascript",
			"// Count provinces",
			"db.getCollection('provinces').countDocuments();",
			"",
			"// Wards of Hà Nội",
			"db.getCollection('provinces').findOne({Code: '01'}, {Name: 1, Wards: 1});",
			"",
			"// GIS: province containing a point",
			"db.getCollection('provinces-gis').findOne({",
			"  \"GIS.Geometry\": { $geoIntersects: { $geometry: { type: 'Point', coordinates: [105.8542, 21.0285] } } }",
			"});",
			"",
			"// GIS: wards in a province",
			"db.getCollection('wards-gis').find({ProvinceCode: '01'}).limit(10);",
			"```",
			"",
			"## GIS / GeoJSON",
			"",
			"The `gis/` subfolder contains the `provinces-gis` and `wards-gis` collections (`mongo_data_vn_province_gis.json`, `mongo_data_vn_ward_gis[_part_NN].json`), the `create_indexes.js` index script, and a `.manifest`. Import them, run `create_indexes.js`, then query with `$geoIntersects` and `$near`.",
		})
}

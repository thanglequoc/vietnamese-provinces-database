package dataset_writer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	vn_common "github.com/thanglequoc-vn-provinces/v2/internal/database"

	file_writer_helper "github.com/thanglequoc-vn-provinces/v2/internal/dataset_writer/dataset_file_writer/helper"
)

type JSONDatasetFileWriter struct {
	OutputFolderPath string
}

func (w *JSONDatasetFileWriter) WriteToFile(
	regions []vn_common.AdministrativeRegion,
	administrativeUnits []vn_common.AdministrativeUnit,
	provinces []vn_common.Province,
	wards []vn_common.Ward) error {

	os.MkdirAll(w.OutputFolderPath, 0746)
	fileTimeSuffix := getFileTimeSuffix()
	outputFilePath := fmt.Sprintf("%s/full_json_generated_data_vn_units_%s.json", w.OutputFolderPath, fileTimeSuffix)
	file, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	dataWriter := bufio.NewWriter(file)

	provinceData := file_writer_helper.ConvertToJsonProvinceModel(provinces)
	data, _ := json.MarshalIndent(provinceData, "", " ")
	dataWriter.Write(data)
	dataWriter.Flush()
	file.Close()

	// JSON Simplified file
	outputFilePath = fmt.Sprintf("%s/simplified_json_generated_data_vn_units_%s.json", w.OutputFolderPath, fileTimeSuffix)
	file, err = os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	dataWriter = bufio.NewWriter(file)
	provinceSimplifiedData := file_writer_helper.ConvertToJsonProvinceSimplifiedModel(provinces)
	data, _ = json.MarshalIndent(provinceSimplifiedData, "", " ")
	dataWriter.Write(data)
	dataWriter.Flush()
	file.Close()

	// VN only JSON Simplified file
	outputFilePath = fmt.Sprintf("%s/vn_only_simplified_json_generated_data_vn_units_%s.json", w.OutputFolderPath, fileTimeSuffix)
	file, err = os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	dataWriter = bufio.NewWriter(file)
	provinceVNSimplifiedData := file_writer_helper.ConvertToJsonProvinceVNSimplifiedModel(provinces)
	data, _ = json.MarshalIndent(provinceVNSimplifiedData, "", " ")
	dataWriter.Write(data)
	dataWriter.Flush()
	file.Close()
	return err
}

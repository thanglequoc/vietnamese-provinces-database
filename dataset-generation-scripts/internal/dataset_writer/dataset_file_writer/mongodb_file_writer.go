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
	fileTimeSuffix := getFileTimeSuffix()

	// Write file administrative units
	administrativeUnitsFilePath := fmt.Sprintf("%s/administrative_units_%s.json", w.OutputFolderPath, fileTimeSuffix)
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
	administrativeRegionsFilePath := fmt.Sprintf("%s/administrative_regions_%s.json", w.OutputFolderPath, fileTimeSuffix)
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
	dataProvinceMongoPath := fmt.Sprintf("%s/mongo_data_vn_unit_%s.json", w.OutputFolderPath, fileTimeSuffix)
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
	return err
}

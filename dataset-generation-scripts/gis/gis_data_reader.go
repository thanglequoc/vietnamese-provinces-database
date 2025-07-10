/* GIS Data Reader */

package gis

import (
	"encoding/json"
	"io"
	"os"
	"log"
)

const GIS_RESOURCE_DIR string = "./resources/gis/"

func ReadGISDataFiles() []ProvinceGIS {
	files, err := os.ReadDir(GIS_RESOURCE_DIR)
	if err != nil {
		log.Fatal(err)
	}
	provinceGIS := make([]ProvinceGIS, len(files))

	for i, f := range files {
		jsonFile, err := os.Open(GIS_RESOURCE_DIR + f.Name())
		if err != nil {
			log.Fatal(err)
		}
		defer jsonFile.Close()

		byteValue, _ := io.ReadAll(jsonFile)
		json.Unmarshal(byteValue, &provinceGIS[i])

	}
	return provinceGIS
}

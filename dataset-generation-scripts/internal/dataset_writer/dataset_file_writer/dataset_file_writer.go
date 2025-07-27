package dataset_writer

import (
	"strings"
	"time"

	"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
)

type DatasetFileWriter interface {
	WriteToFile(
		regions []model.AdministrativeRegion,
		administrativeUnits []model.AdministrativeUnit,
		provinces []model.Province,
		wards []model.Ward) error
}

func getFileTimeSuffix() string {
	return strings.ReplaceAll(strings.ReplaceAll(time.Now().Format(time.DateTime), ":", "_"), " ", "__")
}

/*
Some unit name might have a single quote character, e.g: Ea H'MLay. This method return the escaped single quote
*/
func escapeSingleQuote(source string) string {
	return strings.ReplaceAll(source, "'", "''")
}

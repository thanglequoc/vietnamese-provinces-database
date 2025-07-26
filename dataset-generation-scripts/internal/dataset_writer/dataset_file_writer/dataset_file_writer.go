package dataset_writer

import (
	vn_common "github.com/thanglequoc-vn-provinces/v2/internal/common"
	"strings"
	"time"
)

type DatasetFileWriter interface {
	WriteToFile(
		regions []vn_common.AdministrativeRegion,
		administrativeUnits []vn_common.AdministrativeUnit,
		provinces []vn_common.Province,
		wards []vn_common.Ward) error
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

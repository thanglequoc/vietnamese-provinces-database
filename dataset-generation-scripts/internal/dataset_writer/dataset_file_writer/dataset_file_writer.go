package dataset_writer

import (
	"strconv"
	"strings"
	"time"

	sapnhapmodels "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
	"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
)

type DatasetFileWriter interface {
	WriteToFile(
		regions []model.AdministrativeRegion,
		administrativeUnits []model.AdministrativeUnit,
		provinces []model.Province,
		wards []model.Ward) error

	WriteGISDataToFile(
		sapNhapProvincesGIS []*sapnhapmodels.SapNhapSiteGeoUnit,
		sapNhapWardsGIS []*sapnhapmodels.SapNhapSiteGeoUnit) error
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

// nullableSQLString returns 'value' escaped, or NULL when value is empty.
func nullableSQLString(s string) string {
	if s == "" {
		return "NULL"
	}
	return "'" + escapeSingleQuote(s) + "'"
}

// nullableNString returns N'value' escaped, or NULL when value is empty.
func nullableNString(s string) string {
	if s == "" {
		return "NULL"
	}
	return "N'" + escapeSingleQuote(s) + "'"
}

func parseEuropeanFloat(s string) (float64, error) {
	// Step 1: remove dots (thousands separator)
	s = strings.ReplaceAll(s, ".", "")
	// Step 2: replace comma with dot (decimal separator)
	s = strings.ReplaceAll(s, ",", ".")
	// Step 3: parse as float64 (or float32 if you want)
	return strconv.ParseFloat(s, 64)
}

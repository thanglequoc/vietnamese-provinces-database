package dataset_writer

import (
	"archive/zip"
	"compress/flate"
	"fmt"
	"io"
	"log"
	"os"
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

// zipFile compresses a single file to <sourcePath>.zip using best compression.
// On failure, logs a warning and returns the error. The caller may discard
// the error if zip failure should be non-fatal.
func zipFile(sourcePath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		log.Printf("[WARN] Unable to open source file for zip archive %s: %v", sourcePath, err)
		return fmt.Errorf("open source file %s for zipping: %w", sourcePath, err)
	}
	defer source.Close()

	archivePath := sourcePath + ".zip"
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		log.Printf("[WARN] Unable to create zip archive %s: %v", archivePath, err)
		return fmt.Errorf("create zip archive %s: %w", archivePath, err)
	}
	defer archiveFile.Close()

	zipWriter := zip.NewWriter(archiveFile)
	defer zipWriter.Close()

	zipWriter.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, flate.BestCompression)
	})

	sourceInfo, err := source.Stat()
	if err != nil {
		log.Printf("[WARN] Unable to stat source file %s: %v", sourcePath, err)
		return fmt.Errorf("stat source file %s: %w", sourcePath, err)
	}

	header, err := zip.FileInfoHeader(sourceInfo)
	if err != nil {
		log.Printf("[WARN] Unable to create zip header for %s: %v", sourcePath, err)
		return fmt.Errorf("create zip header for %s: %w", sourcePath, err)
	}
	header.Name = sourceInfo.Name()
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		log.Printf("[WARN] Unable to create zip entry for %s: %v", sourcePath, err)
		return fmt.Errorf("create zip entry for %s: %w", sourcePath, err)
	}

	if _, err := io.Copy(writer, source); err != nil {
		log.Printf("[WARN] Unable to write content to zip archive %s: %v", archivePath, err)
		return fmt.Errorf("copy source content into zip archive %s: %w", archivePath, err)
	}

	return nil
}

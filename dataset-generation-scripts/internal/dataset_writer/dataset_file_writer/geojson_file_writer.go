package dataset_writer

import (
	"archive/zip"
	"compress/flate"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	file_writer_dto "github.com/thanglequoc-vn-provinces/v2/internal/dataset_writer/dataset_file_writer/dto"
	sapnhapmodels "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
)

func (w *JSONDatasetFileWriter) WriteGISGeoJSONToFile(
	sapNhapProvincesGIS []*sapnhapmodels.SapNhapSiteGeoUnit,
	sapNhapWardsGIS []*sapnhapmodels.SapNhapSiteGeoUnit,
) error {
	outputFolderPath := w.OutputFolderPath
	if outputFolderPath == "" {
		outputFolderPath = "./output/gis/geojson"
	}

	if err := os.MkdirAll(outputFolderPath, 0755); err != nil {
		return fmt.Errorf("create geojson output folder %s: %w", outputFolderPath, err)
	}

	sort.SliceStable(sapNhapProvincesGIS, func(i, j int) bool {
		return sapNhapProvincesGIS[i].VNDSProvinceCode < sapNhapProvincesGIS[j].VNDSProvinceCode
	})
	sort.SliceStable(sapNhapWardsGIS, func(i, j int) bool {
		if sapNhapWardsGIS[i].VNDSProvinceCode == sapNhapWardsGIS[j].VNDSProvinceCode {
			return sapNhapWardsGIS[i].VNDSWardCode < sapNhapWardsGIS[j].VNDSWardCode
		}
		return sapNhapWardsGIS[i].VNDSProvinceCode < sapNhapWardsGIS[j].VNDSProvinceCode
	})

	wardsByProvince := make(map[string][]*sapnhapmodels.SapNhapSiteGeoUnit)
	for _, ward := range sapNhapWardsGIS {
		wardsByProvince[ward.VNDSProvinceCode] = append(wardsByProvince[ward.VNDSProvinceCode], ward)
	}

	for _, province := range sapNhapProvincesGIS {
		if err := writeProvinceGeoJSON(outputFolderPath, province); err != nil {
			return err
		}

		for _, ward := range wardsByProvince[province.VNDSProvinceCode] {
			if err := writeWardGeoJSON(outputFolderPath, province, ward); err != nil {
				return err
			}
		}

		delete(wardsByProvince, province.VNDSProvinceCode)
	}

	if len(wardsByProvince) > 0 {
		remainingProvinceCodes := make([]string, 0, len(wardsByProvince))
		for provinceCode := range wardsByProvince {
			remainingProvinceCodes = append(remainingProvinceCodes, provinceCode)
		}
		sort.Strings(remainingProvinceCodes)
		return fmt.Errorf("found wards without a matching province export folder: %v", remainingProvinceCodes)
	}

	if err := archiveGeoJSONDirectory(outputFolderPath); err != nil {
		return err
	}

	return nil
}

func writeProvinceGeoJSON(outputFolderPath string, province *sapnhapmodels.SapNhapSiteGeoUnit) error {
	if province == nil {
		return fmt.Errorf("province export row is nil")
	}

	if province.VNProvince.CodeName == "" {
		return fmt.Errorf("province %s is missing vn province metadata", province.VNDSProvinceCode)
	}

	if len(province.BBoxGeoJSON) == 0 {
		return fmt.Errorf("province %s is missing bbox geojson", province.VNDSProvinceCode)
	}

	if len(province.GeomGeoJSON) == 0 {
		return fmt.Errorf("province %s is missing geometry geojson", province.VNDSProvinceCode)
	}

	fileStem := fmt.Sprintf("%s_%s", province.VNDSProvinceCode, province.VNProvince.CodeName)
	provinceDir := filepath.Join(outputFolderPath, fileStem)
	if err := os.MkdirAll(provinceDir, 0755); err != nil {
		return fmt.Errorf("create province geojson folder %s: %w", provinceDir, err)
	}

	collection := file_writer_dto.GeoJSONFeatureCollection{
		Type: "FeatureCollection",
		BBox: province.BBoxGeoJSON,
		Features: []file_writer_dto.GeoJSONFeature{
			{
				Type:     "Feature",
				ID:       province.VNDSProvinceCode,
				BBox:     province.BBoxGeoJSON,
				Geometry: province.GeomGeoJSON,
				Properties: file_writer_dto.GeoJSONFeatureProperties{
					Code:             province.VNDSProvinceCode,
					Name:             province.VNProvince.Name,
					NameEn:           province.VNProvince.NameEn,
					FullName:         province.VNProvince.FullName,
					FullNameEn:       province.VNProvince.FullNameEn,
					CodeName:         province.VNProvince.CodeName,
					PostalCodePrefix: province.VNProvince.PostalCodePrefix,
					GISServerID:      province.MaLK,
					AreaKm2:          province.DienTichKM2,
				},
			},
		},
	}

	return writeJSONFile(filepath.Join(provinceDir, fileStem+".geojson"), collection)
}

func writeWardGeoJSON(outputFolderPath string, province *sapnhapmodels.SapNhapSiteGeoUnit, ward *sapnhapmodels.SapNhapSiteGeoUnit) error {
	if province == nil {
		return fmt.Errorf("ward export parent province is nil")
	}

	if ward == nil {
		return fmt.Errorf("ward export row is nil")
	}

	if ward.VNWard.CodeName == "" {
		return fmt.Errorf("ward %s is missing vn ward metadata", ward.VNDSWardCode)
	}

	if len(ward.BBoxGeoJSON) == 0 {
		return fmt.Errorf("ward %s is missing bbox geojson", ward.VNDSWardCode)
	}

	if len(ward.GeomGeoJSON) == 0 {
		return fmt.Errorf("ward %s is missing geometry geojson", ward.VNDSWardCode)
	}

	provinceDir := filepath.Join(outputFolderPath, fmt.Sprintf("%s_%s", province.VNDSProvinceCode, province.VNProvince.CodeName))
	wardDir := filepath.Join(provinceDir, "wards")
	if err := os.MkdirAll(wardDir, 0755); err != nil {
		return fmt.Errorf("create ward geojson folder %s: %w", wardDir, err)
	}

	fileStem := fmt.Sprintf("%s_%s", ward.VNDSWardCode, ward.VNWard.CodeName)
	collection := file_writer_dto.GeoJSONFeatureCollection{
		Type: "FeatureCollection",
		BBox: ward.BBoxGeoJSON,
		Features: []file_writer_dto.GeoJSONFeature{
			{
				Type:     "Feature",
				ID:       ward.VNDSWardCode,
				BBox:     ward.BBoxGeoJSON,
				Geometry: ward.GeomGeoJSON,
				Properties: file_writer_dto.GeoJSONFeatureProperties{
					Code:        ward.VNDSWardCode,
					Name:        ward.VNWard.Name,
					NameEn:      ward.VNWard.NameEn,
					FullName:    ward.VNWard.FullName,
					FullNameEn:  ward.VNWard.FullNameEn,
					CodeName:    ward.VNWard.CodeName,
					PostalCode:  ward.VNWard.PostalCode,
					GISServerID: ward.MaLK,
					AreaKm2:     ward.DienTichKM2,
				},
			},
		},
	}

	return writeJSONFile(filepath.Join(wardDir, fileStem+".geojson"), collection)
}

func writeJSONFile(filePath string, payload any) error {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal geojson payload for %s: %w", filePath, err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("write geojson file %s: %w", filePath, err)
	}

	return nil
}

func archiveGeoJSONDirectory(outputFolderPath string) error {
	archiveBaseDir := filepath.Dir(outputFolderPath)
	archiveName := "vn_provinces_wards_geojson.zip"
	archivePath := filepath.Join(archiveBaseDir, archiveName)

	archiveFile, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("create geojson archive %s: %w", archivePath, err)
	}
	defer archiveFile.Close()

	zipWriter := zip.NewWriter(archiveFile)
	defer zipWriter.Close()

	zipWriter.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, flate.BestCompression)
	})

	return filepath.Walk(outputFolderPath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(archiveBaseDir, path)
		if err != nil {
			return fmt.Errorf("resolve relative path for %s: %w", path, err)
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("create zip header for %s: %w", path, err)
		}
		header.Name = relPath
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("create zip entry for %s: %w", path, err)
		}

		source, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open geojson file %s: %w", path, err)
		}
		if _, err := io.Copy(writer, source); err != nil {
			source.Close()
			return fmt.Errorf("copy geojson file %s into archive: %w", path, err)
		}
		if err := source.Close(); err != nil {
			return fmt.Errorf("close geojson file %s: %w", path, err)
		}

		return nil
	})
}


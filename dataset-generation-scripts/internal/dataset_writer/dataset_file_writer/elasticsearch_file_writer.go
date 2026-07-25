package dataset_writer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	dataset_file_writer_dto "github.com/thanglequoc-vn-provinces/v2/internal/dataset_writer/dataset_file_writer/dto"
	file_writer_helper "github.com/thanglequoc-vn-provinces/v2/internal/dataset_writer/dataset_file_writer/helper"
	sapnhapbandomodel "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
	"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
)

const (
	esIndexName    = "provinces"
	esGISIndexName = "provinces-gis"
	esDatasetVer   = "2026.07.01"
	esAdminRev     = "2026-04-30"

	// largeDocWarningThreshold is the size at which a single document triggers a warning.
	largeDocWarningThreshold = 8 * 1024 * 1024 // 8 MB
)

// maxNDJSONChunkSize is the maximum size of a single NDJSON chunk file.
// ES default http.max_content_length is 100 MB; we use 40 MB as a safety margin.
// This is a var (not const) so tests can override it for smaller test data.
var maxNDJSONChunkSize = 40 * 1024 * 1024 // 40 MB

// ElasticsearchDatasetFileWriter generates Elasticsearch NDJSON bulk files,
// index mappings, and a README for the provinces and provinces-gis indices.
type ElasticsearchDatasetFileWriter struct {
	OutputFolderPath string
}

// WriteToFile generates the non-GIS provinces index (NDJSON + mapping + README).
func (w *ElasticsearchDatasetFileWriter) WriteToFile(
	regions []model.AdministrativeRegion,
	administrativeUnits []model.AdministrativeUnit,
	provinces []model.Province,
	wards []model.Ward) error {

	mappingsDir := fmt.Sprintf("%s/mappings", w.OutputFolderPath)
	if err := os.MkdirAll(mappingsDir, 0746); err != nil {
		return fmt.Errorf("create output directories: %w", err)
	}

	docs := file_writer_helper.ConvertToElasticsearchProvinceModel(provinces)
	generatedAt := time.Now().UTC().Format(time.RFC3339)

	// Attach Meta to each document
	for i := range docs {
		docs[i].Meta = &dataset_file_writer_dto.ElasticsearchMeta{
			DatasetVersion:         esDatasetVer,
			AdministrativeRevision: esAdminRev,
			GeneratedAt:            generatedAt,
		}
	}

	// Write provinces.ndjson
	ndjsonPath := fmt.Sprintf("%s/provinces.ndjson", w.OutputFolderPath)
	if err := writeNDJSON(ndjsonPath, esIndexName, docs); err != nil {
		return fmt.Errorf("write provinces ndjson: %w", err)
	}

	// Write mappings/provinces.json
	mappingPath := fmt.Sprintf("%s/provinces.json", mappingsDir)
	if err := writeProvincesMapping(mappingPath); err != nil {
		return fmt.Errorf("write provinces mapping: %w", err)
	}

	// Write README.md
	readmePath := fmt.Sprintf("%s/README.md", w.OutputFolderPath)
	if err := writeESReadme(readmePath); err != nil {
		return fmt.Errorf("write README: %w", err)
	}

	return nil
}

// WriteElasticsearchGISDataToFile generates the provinces-gis index (NDJSON + mapping).
func (w *ElasticsearchDatasetFileWriter) WriteElasticsearchGISDataToFile(
	sapNhapGeoProvinces []*sapnhapbandomodel.SapNhapSiteGeoUnit,
	sapNhapGeoWards []*sapnhapbandomodel.SapNhapSiteGeoUnit) error {

	mappingsDir := fmt.Sprintf("%s/mappings", w.OutputFolderPath)
	if err := os.MkdirAll(mappingsDir, 0746); err != nil {
		return fmt.Errorf("create output directories: %w", err)
	}

	// Group wards by province code
	wardsByProvince := make(map[string][]*sapnhapbandomodel.SapNhapSiteGeoUnit)
	for _, ward := range sapNhapGeoWards {
		wardsByProvince[ward.VNDSProvinceCode] = append(wardsByProvince[ward.VNDSProvinceCode], ward)
	}

	generatedAt := time.Now().UTC().Format(time.RFC3339)

	var docs []dataset_file_writer_dto.ElasticsearchProvinceDocument
	for _, geoProvince := range sapNhapGeoProvinces {
		province := geoProvince.VNProvince

		doc := dataset_file_writer_dto.ElasticsearchProvinceDocument{
			Code:               province.Code,
			Name:               province.Name,
			NameEn:             province.NameEn,
			FullName:           province.FullName,
			FullNameEn:         province.FullNameEn,
			CodeName:           province.CodeName,
			AdministrativeUnit: convertProvinceToESAdminUnit(province.AdministrativeUnit),
			SearchKeywords:     file_writer_helper.GenerateSearchKeywords(province.Code, province.Name, province.NameEn, province.CodeName),
			Meta: &dataset_file_writer_dto.ElasticsearchMeta{
				DatasetVersion:         esDatasetVer,
				AdministrativeRevision: esAdminRev,
				GeneratedAt:            generatedAt,
			},
		}

		// Add province GIS with Properties
		provinceProps := &dataset_file_writer_dto.ElasticsearchGISProperties{
			Code:        province.Code,
			Name:        province.Name,
			NameEn:      province.NameEn,
			FullName:    province.FullName,
			FullNameEn:  province.FullNameEn,
			CodeName:    province.CodeName,
			GisServerId: geoProvince.MaLK,
			AreaKm2:     geoProvince.DienTichKM2,
		}
		if gis, err := sapnhapGeoUnitToESGIS(*geoProvince, provinceProps); err == nil {
			doc.GIS = gis
		}

		// Embed wards with their GIS
		geoWards := wardsByProvince[geoProvince.VNDSProvinceCode]
		for _, geoWard := range geoWards {
			ward := geoWard.VNWard
			wardDoc := dataset_file_writer_dto.ElasticsearchWardDocument{
				Code:               ward.Code,
				Name:               ward.Name,
				NameEn:             ward.NameEn,
				FullName:           ward.FullName,
				FullNameEn:         ward.FullNameEn,
				CodeName:           ward.CodeName,
				AdministrativeUnit: convertWardToESAdminUnit(ward.AdministrativeUnit),
				SearchKeywords:     file_writer_helper.GenerateSearchKeywords(ward.Code, ward.Name, ward.NameEn, ward.CodeName),
			}
			wardProps := &dataset_file_writer_dto.ElasticsearchGISProperties{
				Code:        ward.Code,
				Name:        ward.Name,
				NameEn:      ward.NameEn,
				FullName:    ward.FullName,
				FullNameEn:  ward.FullNameEn,
				CodeName:    ward.CodeName,
				GisServerId: geoWard.MaLK,
				AreaKm2:     geoWard.DienTichKM2,
			}
			if gis, err := sapnhapGeoUnitToESGIS(*geoWard, wardProps); err == nil {
				wardDoc.GIS = gis
			}
			doc.Wards = append(doc.Wards, wardDoc)
		}

		docs = append(docs, doc)
	}

	// Warn about large documents that may need special handling during ES import
	for _, doc := range docs {
		docBytes, _ := json.Marshal(doc)
		if len(docBytes) > largeDocWarningThreshold {
			log.Printf("⚠️  [Elasticsearch] Large province document: Code=%s, Size=%.2f MB — may require increased ES heap during import",
				doc.Code, float64(len(docBytes))/1024/1024)
		}
	}

	// Write provinces-gis.ndjson (chunked if total size exceeds maxNDJSONChunkSize)
	ndjsonPath := fmt.Sprintf("%s/provinces-gis.ndjson", w.OutputFolderPath)
	if err := writeChunkedNDJSON(ndjsonPath, esGISIndexName, docs); err != nil {
		return fmt.Errorf("write provinces-gis ndjson: %w", err)
	}

	// Write mappings/provinces-gis.json
	mappingPath := fmt.Sprintf("%s/provinces-gis.json", mappingsDir)
	if err := writeProvincesGISMapping(mappingPath); err != nil {
		return fmt.Errorf("write provinces-gis mapping: %w", err)
	}

	return nil
}

// writeNDJSON writes Elasticsearch Bulk API NDJSON (index line + document line).
func writeNDJSON(path, index string, docs []dataset_file_writer_dto.ElasticsearchProvinceDocument) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)

	for _, doc := range docs {
		// Index action line
		action := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": index,
				"_id":    doc.Code,
			},
		}
		if err := encoder.Encode(action); err != nil {
			return fmt.Errorf("encode index action for doc %s: %w", doc.Code, err)
		}
		// Document line
		if err := encoder.Encode(doc); err != nil {
			return fmt.Errorf("encode doc %s: %w", doc.Code, err)
		}
	}

	return writer.Flush()
}

// writeChunkedNDJSON writes Elasticsearch Bulk API NDJSON, splitting into
// multiple chunk files if the total size exceeds maxNDJSONChunkSize.
// This avoids the ES http.max_content_length limit (100 MB default).
//
// If the total size fits within one chunk, a single file is written at `path`
// (same as writeNDJSON). If chunking is needed, files are written as
// `path` with a numeric suffix: e.g. `provinces-gis-part-01.ndjson`,
// `provinces-gis-part-02.ndjson`, etc.
//
// A manifest file `path + ".manifest"` is also written, listing the chunk
// filenames in order, so import scripts can iterate them.
func writeChunkedNDJSON(path, index string, docs []dataset_file_writer_dto.ElasticsearchProvinceDocument) error {
	// Pre-serialize all docs to determine sizes
	type serializedDoc struct {
		actionLine []byte
		docLine    []byte
		totalSize  int
	}

	var serialized []serializedDoc
	totalSize := 0
	for _, doc := range docs {
		action := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": index,
				"_id":    doc.Code,
			},
		}
		actionBytes, err := json.Marshal(action)
		if err != nil {
			return fmt.Errorf("marshal index action for doc %s: %w", doc.Code, err)
		}
		docBytes, err := json.Marshal(doc)
		if err != nil {
			return fmt.Errorf("marshal doc %s: %w", doc.Code, err)
		}
		// Each line includes a trailing newline
		size := len(actionBytes) + 1 + len(docBytes) + 1
		serialized = append(serialized, serializedDoc{actionBytes, docBytes, size})
		totalSize += size
	}

	// If total fits in one file, use the simple path
	if totalSize <= maxNDJSONChunkSize {
		return writeNDJSON(path, index, docs)
	}

	// Split into chunks
	dir := filepathDir(path)
	base := filepathBase(path)
	ext := filepathExt(base)
	nameNoExt := base[:len(base)-len(ext)]

	var chunks [][]serializedDoc
	currentChunk := []serializedDoc{}
	currentSize := 0

	for _, s := range serialized {
		if currentSize+s.totalSize > maxNDJSONChunkSize && len(currentChunk) > 0 {
			chunks = append(chunks, currentChunk)
			currentChunk = []serializedDoc{}
			currentSize = 0
		}
		currentChunk = append(currentChunk, s)
		currentSize += s.totalSize
	}
	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	log.Printf("📦 [Elasticsearch] NDJSON chunked: %d docs → %d files (max %d MB each)",
		len(docs), len(chunks), maxNDJSONChunkSize/1024/1024)

	// Write each chunk file
	var chunkNames []string
	for i, chunk := range chunks {
		chunkName := fmt.Sprintf("%s-part-%02d%s", nameNoExt, i+1, ext)
		chunkPath := fmt.Sprintf("%s/%s", dir, chunkName)
		chunkNames = append(chunkNames, chunkName)

		file, err := os.OpenFile(chunkPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("create chunk file %s: %w", chunkPath, err)
		}

		writer := bufio.NewWriter(file)
		for _, s := range chunk {
			if _, err := writer.Write(s.actionLine); err != nil {
				file.Close()
				return fmt.Errorf("write action line: %w", err)
			}
			if err := writer.WriteByte('\n'); err != nil {
				file.Close()
				return fmt.Errorf("write newline: %w", err)
			}
			if _, err := writer.Write(s.docLine); err != nil {
				file.Close()
				return fmt.Errorf("write doc line: %w", err)
			}
			if err := writer.WriteByte('\n'); err != nil {
				file.Close()
				return fmt.Errorf("write newline: %w", err)
			}
		}
		if err := writer.Flush(); err != nil {
			file.Close()
			return fmt.Errorf("flush chunk file %s: %w", chunkPath, err)
		}
		file.Close()

		chunkSize := 0
		for _, s := range chunk {
			chunkSize += s.totalSize
		}
		log.Printf("   %s: %.1f MB, %d docs", chunkName, float64(chunkSize)/1024/1024, len(chunk))
	}

	// Write manifest file
	manifestPath := path + ".manifest"
	manifestContent := stringsJoin(chunkNames, "\n") + "\n"
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		return fmt.Errorf("write manifest file: %w", err)
	}
	log.Printf("   Manifest: %s", filepathBase(manifestPath))

	return nil
}

// filepathDir returns the directory portion of a path (like filepath.Dir).
func filepathDir(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			if i == 0 {
				return "/"
			}
			return path[:i]
		}
	}
	return "."
}

// filepathBase returns the last element of a path (like filepath.Base).
func filepathBase(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}

// filepathExt returns the file extension including the dot (like filepath.Ext).
func filepathExt(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return ""
		}
		if path[i] == '.' {
			return path[i:]
		}
	}
	return ""
}

// stringsJoin joins strings with a separator (like strings.Join).
func stringsJoin(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for _, s := range strs[1:] {
		result += sep + s
	}
	return result
}

// sapnhapGeoUnitToESGIS converts a SapNhapSiteGeoUnit's BBoxGeoJSON and
// GeomGeoJSON into an ElasticsearchGIS struct. Returns an error if the bbox
// cannot be parsed.
func sapnhapGeoUnitToESGIS(unit sapnhapbandomodel.SapNhapSiteGeoUnit, properties *dataset_file_writer_dto.ElasticsearchGISProperties) (*dataset_file_writer_dto.ElasticsearchGIS, error) {
	bbox, center, err := parseBBox(unit.BBoxGeoJSON)
	if err != nil {
		return nil, err
	}
	return &dataset_file_writer_dto.ElasticsearchGIS{
		Properties:  properties,
		Center:      center,
		BoundingBox: bbox,
		Geometry:    unit.GeomGeoJSON,
	}, nil
}

// parseBBox parses a BBoxGeoJSON array [xmin, ymin, xmax, ymax] into
// ElasticsearchBoundingBox and ElasticsearchGeoPoint (center).
func parseBBox(bboxGeoJSON json.RawMessage) (dataset_file_writer_dto.ElasticsearchBoundingBox, dataset_file_writer_dto.ElasticsearchGeoPoint, error) {
	var coords []float64
	if err := json.Unmarshal(bboxGeoJSON, &coords); err != nil {
		return dataset_file_writer_dto.ElasticsearchBoundingBox{}, dataset_file_writer_dto.ElasticsearchGeoPoint{}, fmt.Errorf("parse bbox geojson: %w", err)
	}
	if len(coords) != 4 {
		return dataset_file_writer_dto.ElasticsearchBoundingBox{}, dataset_file_writer_dto.ElasticsearchGeoPoint{}, fmt.Errorf("expected 4 bbox coordinates, got %d", len(coords))
	}
	xmin, ymin, xmax, ymax := coords[0], coords[1], coords[2], coords[3]
	bbox := dataset_file_writer_dto.ElasticsearchBoundingBox{
		MinLongitude: xmin,
		MinLatitude:  ymin,
		MaxLongitude: xmax,
		MaxLatitude:  ymax,
	}
	center := dataset_file_writer_dto.ElasticsearchGeoPoint{
		Lat: (ymin + ymax) / 2,
		Lon: (xmin + xmax) / 2,
	}
	return bbox, center, nil
}

// convertProvinceToESAdminUnit converts a model.AdministrativeUnit to the ES DTO.
func convertProvinceToESAdminUnit(au model.AdministrativeUnit) dataset_file_writer_dto.ElasticsearchAdministrativeUnit {
	return dataset_file_writer_dto.ElasticsearchAdministrativeUnit{
		Id:          au.Id,
		FullName:    au.FullName,
		FullNameEn:  au.FullNameEn,
		ShortName:   au.ShortName,
		ShortNameEn: au.ShortNameEn,
		CodeName:    au.CodeName,
		CodeNameEn:  au.CodeNameEn,
	}
}

// convertWardToESAdminUnit converts a model.AdministrativeUnit to the ES DTO.
func convertWardToESAdminUnit(au model.AdministrativeUnit) dataset_file_writer_dto.ElasticsearchAdministrativeUnit {
	return dataset_file_writer_dto.ElasticsearchAdministrativeUnit{
		Id:          au.Id,
		FullName:    au.FullName,
		FullNameEn:  au.FullNameEn,
		ShortName:   au.ShortName,
		ShortNameEn: au.ShortNameEn,
		CodeName:    au.CodeName,
		CodeNameEn:  au.CodeNameEn,
	}
}

// writeProvincesMapping writes the static mapping for the provinces index.
func writeProvincesMapping(path string) error {
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"dynamic": "strict",
			"properties": map[string]interface{}{
				"Code":           map[string]string{"type": "keyword"},
				"CodeName":       map[string]string{"type": "keyword"},
				"SearchKeywords": map[string]string{"type": "keyword"},
				"Name": map[string]interface{}{
					"type":   "text",
					"fields": map[string]interface{}{"keyword": map[string]string{"type": "keyword"}},
				},
				"NameEn": map[string]interface{}{
					"type":   "text",
					"fields": map[string]interface{}{"keyword": map[string]string{"type": "keyword"}},
				},
				"FullName":   map[string]string{"type": "text"},
				"FullNameEn": map[string]string{"type": "text"},
				"AdministrativeUnit": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"Id":          map[string]string{"type": "integer"},
						"FullName":    map[string]string{"type": "keyword"},
						"FullNameEn":  map[string]string{"type": "keyword"},
						"ShortName":   map[string]string{"type": "keyword"},
						"ShortNameEn": map[string]string{"type": "keyword"},
						"CodeName":    map[string]string{"type": "keyword"},
						"CodeNameEn":  map[string]string{"type": "keyword"},
					},
				},
				"Wards": map[string]interface{}{
					"type": "nested",
					"properties": map[string]interface{}{
						"Code":           map[string]string{"type": "keyword"},
						"CodeName":       map[string]string{"type": "keyword"},
						"SearchKeywords": map[string]string{"type": "keyword"},
						"Name": map[string]interface{}{
							"type":   "text",
							"fields": map[string]interface{}{"keyword": map[string]string{"type": "keyword"}},
						},
						"NameEn": map[string]interface{}{
							"type":   "text",
							"fields": map[string]interface{}{"keyword": map[string]string{"type": "keyword"}},
						},
						"FullName":   map[string]string{"type": "text"},
						"FullNameEn": map[string]string{"type": "text"},
						"AdministrativeUnit": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"Id":          map[string]string{"type": "integer"},
								"FullName":    map[string]string{"type": "keyword"},
								"FullNameEn":  map[string]string{"type": "keyword"},
								"ShortName":   map[string]string{"type": "keyword"},
								"ShortNameEn": map[string]string{"type": "keyword"},
								"CodeName":    map[string]string{"type": "keyword"},
								"CodeNameEn":  map[string]string{"type": "keyword"},
							},
						},
					},
				},
				"Meta": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"DatasetVersion":         map[string]string{"type": "keyword"},
						"AdministrativeRevision": map[string]string{"type": "keyword"},
						"GeneratedAt":            map[string]string{"type": "date"},
					},
				},
			},
		},
	}
	return writeJSONFile(path, mapping)
}

// writeProvincesGISMapping writes the static mapping for the provinces-gis index.
func writeProvincesGISMapping(path string) error {
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"dynamic": "strict",
			"properties": map[string]interface{}{
				"Code":           map[string]string{"type": "keyword"},
				"CodeName":       map[string]string{"type": "keyword"},
				"SearchKeywords": map[string]string{"type": "keyword"},
				"Name": map[string]interface{}{
					"type":   "text",
					"fields": map[string]interface{}{"keyword": map[string]string{"type": "keyword"}},
				},
				"NameEn": map[string]interface{}{
					"type":   "text",
					"fields": map[string]interface{}{"keyword": map[string]string{"type": "keyword"}},
				},
				"FullName":   map[string]string{"type": "text"},
				"FullNameEn": map[string]string{"type": "text"},
				"AdministrativeUnit": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"Id":          map[string]string{"type": "integer"},
						"FullName":    map[string]string{"type": "keyword"},
						"FullNameEn":  map[string]string{"type": "keyword"},
						"ShortName":   map[string]string{"type": "keyword"},
						"ShortNameEn": map[string]string{"type": "keyword"},
						"CodeName":    map[string]string{"type": "keyword"},
						"CodeNameEn":  map[string]string{"type": "keyword"},
					},
				},
				"GIS": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"Center": map[string]string{"type": "geo_point"},
						"BoundingBox": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"MinLongitude": map[string]string{"type": "float"},
								"MinLatitude":  map[string]string{"type": "float"},
								"MaxLongitude": map[string]string{"type": "float"},
								"MaxLatitude":  map[string]string{"type": "float"},
							},
						},
						"Geometry": map[string]string{"type": "geo_shape"},
						"Properties": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"Code":        map[string]string{"type": "keyword"},
								"Name":        map[string]string{"type": "keyword"},
								"NameEn":      map[string]string{"type": "keyword"},
								"FullName":    map[string]string{"type": "keyword"},
								"FullNameEn":  map[string]string{"type": "keyword"},
								"CodeName":    map[string]string{"type": "keyword"},
								"GisServerId": map[string]string{"type": "keyword"},
								"AreaKm2":     map[string]string{"type": "float"},
							},
						},
					},
				},
				"Wards": map[string]interface{}{
					"type": "nested",
					"properties": map[string]interface{}{
						"Code":           map[string]string{"type": "keyword"},
						"CodeName":       map[string]string{"type": "keyword"},
						"SearchKeywords": map[string]string{"type": "keyword"},
						"Name": map[string]interface{}{
							"type":   "text",
							"fields": map[string]interface{}{"keyword": map[string]string{"type": "keyword"}},
						},
						"NameEn": map[string]interface{}{
							"type":   "text",
							"fields": map[string]interface{}{"keyword": map[string]string{"type": "keyword"}},
						},
						"FullName":   map[string]string{"type": "text"},
						"FullNameEn": map[string]string{"type": "text"},
						"AdministrativeUnit": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"Id":          map[string]string{"type": "integer"},
								"FullName":    map[string]string{"type": "keyword"},
								"FullNameEn":  map[string]string{"type": "keyword"},
								"ShortName":   map[string]string{"type": "keyword"},
								"ShortNameEn": map[string]string{"type": "keyword"},
								"CodeName":    map[string]string{"type": "keyword"},
								"CodeNameEn":  map[string]string{"type": "keyword"},
							},
						},
						"GIS": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"Center": map[string]string{"type": "geo_point"},
								"BoundingBox": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"MinLongitude": map[string]string{"type": "float"},
										"MinLatitude":  map[string]string{"type": "float"},
										"MaxLongitude": map[string]string{"type": "float"},
										"MaxLatitude":  map[string]string{"type": "float"},
									},
								},
								"Geometry": map[string]string{"type": "geo_shape"},
								"Properties": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"Code":        map[string]string{"type": "keyword"},
										"Name":        map[string]string{"type": "keyword"},
										"NameEn":      map[string]string{"type": "keyword"},
										"FullName":    map[string]string{"type": "keyword"},
										"FullNameEn":  map[string]string{"type": "keyword"},
										"CodeName":    map[string]string{"type": "keyword"},
										"GisServerId": map[string]string{"type": "keyword"},
										"AreaKm2":     map[string]string{"type": "float"},
									},
								},
							},
						},
					},
				},
				"Meta": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"DatasetVersion":         map[string]string{"type": "keyword"},
						"AdministrativeRevision": map[string]string{"type": "keyword"},
						"GeneratedAt":            map[string]string{"type": "date"},
					},
				},
			},
		},
	}
	return writeJSONFile(path, mapping)
}

// writeESReadme writes the README.md for the Elasticsearch dataset.
func writeESReadme(path string) error {
	// Get current time in GMT+7
	loc, err := time.LoadLocation("Asia/Saigon")
	if err != nil {
		loc = time.FixedZone("GMT+7", 7*60*60)
	}
	createdAt := time.Now().In(loc).Format(time.RFC1123Z)

	content := `# Vietnamese Provinces Database — Elasticsearch Dataset

Created at:  ` + createdAt + `

## Overview

This dataset provides Vietnamese provinces and wards in Elasticsearch document format
with two indices:

| Index | Description |
|-------|-------------|
| ` + "`provinces`" + ` | Provincial metadata with embedded wards, search keywords, and administrative unit data (no GIS geometry) |
| ` + "`provinces-gis`" + ` | Same structure plus GIS geometry for both provinces and wards (bounding boxes + GeoJSON polygons) |

## Document Structure

Each province is a single denormalized document with:

- **Core fields**: Code, Name, NameEn, FullName, FullNameEn, CodeName
- **` + "`AdministrativeUnit`" + `**: Embedded administrative unit object (Id, FullName, ShortName, etc.)
- **` + "`SearchKeywords`" + `**: Pre-computed autocomplete keywords (code, tone-stripped name, English name, codeName)
- **` + "`Wards`" + `**: Array of nested ward documents with the same structure
- **` + "`GIS`" + `**: (provinces-gis only) Center (geo_point), BoundingBox, Geometry (geo_shape)
- **` + "`Meta`" + `**: Dataset version metadata (DatasetVersion, AdministrativeRevision, GeneratedAt)

## Example Preview Document

Below is a sample province document (Hà Nội) with two of its wards:

` + "```json" + `
{
  "Code": "01",
  "Name": "Hà Nội",
  "NameEn": "Hanoi",
  "FullName": "Thành phố Hà Nội",
  "FullNameEn": "Hanoi City",
  "CodeName": "ha_noi",
  "AdministrativeUnit": {
    "Id": 1,
    "FullName": "Thành phố trực thuộc trung ương",
    "FullNameEn": "Municipality",
    "ShortName": "Thành phố",
    "ShortNameEn": "City",
    "CodeName": "thanh_pho_truc_thuoc_trung_uong",
    "CodeNameEn": "municipality"
  },
  "SearchKeywords": ["01", "ha noi", "hanoi", "ha_noi"],
  "Wards": [
    {
      "Code": "00004",
      "Name": "Ba Đình",
      "NameEn": "Ba Dinh",
      "FullName": "Phường Ba Đình",
      "FullNameEn": "Ba Dinh Ward",
      "CodeName": "ba_dinh",
      "AdministrativeUnit": {
        "Id": 3,
        "FullName": "Phường",
        "FullNameEn": "Ward",
        "ShortName": "Phường",
        "ShortNameEn": "Ward",
        "CodeName": "phuong",
        "CodeNameEn": "ward"
      },
      "SearchKeywords": ["00004", "ba dinh", "ba_dinh"]
    },
    {
      "Code": "00070",
      "Name": "Hoàn Kiếm",
      "NameEn": "Hoan Kiem",
      "FullName": "Phường Hoàn Kiếm",
      "FullNameEn": "Hoan Kiem Ward",
      "CodeName": "hoan_kiem",
      "AdministrativeUnit": {
        "Id": 3,
        "FullName": "Phường",
        "FullNameEn": "Ward",
        "ShortName": "Phường",
        "ShortNameEn": "Ward",
        "CodeName": "phuong",
        "CodeNameEn": "ward"
      },
      "SearchKeywords": ["00070", "hoan kiem", "hoan_kiem"]
    }
  ],
  "Meta": {
    "DatasetVersion": "2026.07.01",
    "AdministrativeRevision": "2026-04-30",
    "GeneratedAt": "2026-07-25T03:00:43Z"
  }
}
` + "```" + `

The ` + "`provinces-gis`" + ` index extends this same structure with a ` + "`GIS`" + ` object at both the province and ward level:

` + "```json" + `
{
  "Code": "01",
  "Name": "Hà Nội",
  "FullName": "Thành phố Hà Nội",
  "CodeName": "ha_noi",
  "AdministrativeUnit": {
    "Id": 1,
    "FullName": "Thành phố trực thuộc trung ương",
    "ShortName": "Thành phố"
  },
  "SearchKeywords": ["01", "ha noi", "hanoi", "ha_noi"],
  "GIS": {
    "Center": { "Lat": 21.0285, "Lon": 105.8542 },
    "BoundingBox": {
      "MinLongitude": 105.2859,
      "MinLatitude": 20.4863,
      "MaxLongitude": 106.0617,
      "MaxLatitude": 21.3851
    },
    "Geometry": {
      "type": "MultiPolygon",
      "coordinates": [[[[105.2859, 21.3851], [106.0617, 21.3851], ...]]]
    },
    "Properties": {
      "Code": "01",
      "Name": "Hà Nội",
      "NameEn": "Hanoi",
      "FullName": "Thành phố Hà Nội",
      "FullNameEn": "Hanoi City",
      "CodeName": "ha_noi",
      "GisServerId": "diaphanhanhchinhcaptinh_sn.108",
      "AreaKm2": 3359.84
    }
  },
  "Wards": [
    {
      "Code": "00004",
      "Name": "Ba Đình",
      "FullName": "Phường Ba Đình",
      "CodeName": "ba_dinh",
      "AdministrativeUnit": { "Id": 3, "ShortName": "Phường" },
      "SearchKeywords": ["00004", "ba dinh", "ba_dinh"],
      "GIS": {
        "Center": { "Lat": 21.0347, "Lon": 105.8231 },
        "BoundingBox": {
          "MinLongitude": 105.8115, "MinLatitude": 21.0261,
          "MaxLongitude": 105.8347, "MaxLatitude": 21.0433
        },
        "Geometry": { "type": "Polygon", "coordinates": [[[105.8115, 21.0433], ...]] },
        "Properties": {
          "Code": "00004",
          "Name": "Ba Đình",
          "NameEn": "Ba Dinh",
          "FullName": "Phường Ba Đình",
          "FullNameEn": "Ba Dinh Ward",
          "CodeName": "ba_dinh",
          "GisServerId": "diaphanhanhchinhphuong_sn.456",
          "AreaKm2": 5.23
        }
      }
    }
  ],
  "Meta": {
    "DatasetVersion": "2026.07.01",
    "AdministrativeRevision": "2026-04-30",
    "GeneratedAt": "2026-07-25T03:00:43Z"
  }
}
` + "```" + `

> **Note**: The ` + "`Geometry`" + ` field contains full GeoJSON polygons/multipolygons. The example above uses ` + "`...`" + ` to abbreviate the coordinate arrays for readability. Actual geometries for provinces are MultiPolygon with thousands of coordinate pairs.

## Quick Start

### 1. Create the Indices with Mappings

` + "```bash" + `
# Create the provinces index
curl -X PUT "localhost:9200/provinces" \
  -H 'Content-Type: application/json' \
  -d @mappings/provinces.json

# Create the provinces-gis index (with GIS geometry support)
curl -X PUT "localhost:9200/provinces-gis" \
  -H 'Content-Type: application/json' \
  -d @mappings/provinces-gis.json
` + "```" + `

### 2. Bulk Import the Data

` + "```bash" + `
# Import province data (non-GIS)
curl -X POST "localhost:9200/_bulk" \
  -H 'Content-Type: application/x-ndjson' \
  --data-binary @provinces.ndjson

# Import province data (with GIS)
curl -X POST "localhost:9200/_bulk" \
  -H 'Content-Type: application/x-ndjson' \
  --data-binary @provinces-gis.ndjson
` + "```" + `

### 3. Verify Import

` + "```bash" + `
curl "localhost:9200/provinces/_count"
curl "localhost:9200/provinces-gis/_count"
` + "```" + `

Expected: 34 documents in each index (one per province).

## Example Queries

### Province dropdown (sorted by code)
` + "```json" + `
POST /provinces/_search
{
  "size": 34,
  "sort": [{"Code": "asc"}],
  "_source": ["Code", "Name", "NameEn"]
}
` + "```" + `

### Search wards within a province
` + "```json" + `
POST /provinces/_search
{
  "query": {
    "nested": {
      "path": "Wards",
      "query": {
        "bool": {
          "must": [
            {"match": {"Wards.FullName": "Ba Đình"}}
          ]
        }
      },
      "inner_hits": {}
    }
  },
  "_source": ["Code", "Name"]
}
` + "```" + `

### Return only the matched ward (no parent province)

Use ` + "`_source: false`" + ` on the parent document so only the nested ward hit is returned:

` + "```json" + `
POST /provinces/_search
{
  "_source": false,
  "query": {
    "nested": {
      "path": "Wards",
      "query": { "match": { "Wards.CodeName": "truong_sa" } },
      "inner_hits": { "name": "ward", "_source": true }
    }
  }
}
` + "```" + `

The ward document (with GIS data if present) is available at ` + "`.hits.hits[0].inner_hits.ward.hits.hits[0]._source`" + `.

### Autocomplete search
` + "```json" + `
POST /provinces/_search
{
  "query": {
    "terms": {"SearchKeywords": ["ha noi"]}
  },
  "_source": ["Code", "Name", "NameEn"]
}
` + "```" + `

### GIS: Find province covering a point
` + "```json" + `
POST /provinces-gis/_search
{
  "query": {
    "geo_shape": {
      "GIS.Geometry": {
        "shape": {
          "type": "point",
          "coordinates": [105.8542, 21.0285]
        },
        "relation": "intersects"
      }
    }
  },
  "_source": ["Code", "Name"]
}
` + "```" + `

## File Listing

| File | Description |
|------|-------------|
| ` + "`provinces.ndjson`" + ` | Bulk API NDJSON for the provinces index |
| ` + "`provinces-gis.ndjson`" + ` | Bulk API NDJSON for the provinces-gis index |
| ` + "`mappings/provinces.json`" + ` | Index mapping for provinces |
| ` + "`mappings/provinces-gis.json`" + ` | Index mapping for provinces-gis |

## Notes

- Field names use **PascalCase** (consistent with MongoDB/JSON exports)
- The ` + "`Meta`" + ` field is named without underscore prefix — Elasticsearch reserves ` + "`_`" + `-prefixed field names
- The dataset version and administrative revision are set at generation time
- NDJSON files use the Elasticsearch Bulk API format (each document = index action line + document line)
`
	return os.WriteFile(path, []byte(content), 0644)
}

package dataset_writer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	file_writer_helper "github.com/thanglequoc-vn-provinces/v2/internal/dataset_writer/dataset_file_writer/helper"
	sapnhapbandomodel "github.com/thanglequoc-vn-provinces/v2/internal/sapnhap_bando/model"
)

const (
	mongoDatasetVer   = "2026.07.01"
	mongoAdminRev     = "2026-04-30"
	mongoGISChunkSize = 50 * 1024 * 1024 // 50 MB
)

// WriteMongoGISDataToFile generates the provinces-gis and wards-gis MongoDB
// JSON files, create_indexes.js, and README.md.
func (w *MongoDBDatasetFileWriter) WriteMongoGISDataToFile(
	sapNhapGeoProvinces []*sapnhapbandomodel.SapNhapSiteGeoUnit,
	sapNhapGeoWards []*sapnhapbandomodel.SapNhapSiteGeoUnit) error {

	os.MkdirAll(w.OutputFolderPath, 0746)
	fileTimeSuffix := getFileTimeSuffix()
	generatedAt := time.Now().UTC().Format(time.RFC3339)

	// Build province GIS documents
	provinceDocs := file_writer_helper.ConvertToMongoGISProvinceDocuments(
		sapNhapGeoProvinces, mongoDatasetVer, mongoAdminRev, generatedAt)

	// Build ward GIS documents
	wardDocs := file_writer_helper.ConvertToMongoGISWardDocuments(
		sapNhapGeoWards, mongoDatasetVer, mongoAdminRev, generatedAt)

	// Write province GIS file (chunked if needed)
	provincePath := fmt.Sprintf("%s/mongo_data_vn_province_gis_%s.json", w.OutputFolderPath, fileTimeSuffix)
	if err := writeChunkedMongoJSON(provincePath, provinceDocs); err != nil {
		return fmt.Errorf("write province GIS json: %w", err)
	}

	// Write ward GIS file (chunked if needed)
	wardPath := fmt.Sprintf("%s/mongo_data_vn_ward_gis_%s.json", w.OutputFolderPath, fileTimeSuffix)
	if err := writeChunkedMongoJSON(wardPath, wardDocs); err != nil {
		return fmt.Errorf("write ward GIS json: %w", err)
	}

	// Write create_indexes.js
	indexPath := fmt.Sprintf("%s/create_indexes.js", w.OutputFolderPath)
	if err := writeMongoGISIndexScript(indexPath); err != nil {
		return fmt.Errorf("write create_indexes.js: %w", err)
	}

	// Write README.md
	readmePath := fmt.Sprintf("%s/README.md", w.OutputFolderPath)
	if err := writeMongoGISReadme(readmePath); err != nil {
		return fmt.Errorf("write README: %w", err)
	}

	return nil
}

// writeChunkedMongoJSON writes a JSON array of documents, splitting into
// multiple chunk files if the total size exceeds mongoGISChunkSize (50MB).
// If the total fits in one file, writes a single file at path.
// If chunking is needed, files are written as path with numeric suffix:
// e.g. mongo_data_vn_ward_gis_2026..._part_01.json, _part_02.json, etc.
// A manifest file (path + ".manifest") lists the chunk filenames in order.
func writeChunkedMongoJSON(path string, docs interface{}) error {
	// Pre-serialize to determine total size
	data, err := json.MarshalIndent(docs, "", " ")
	if err != nil {
		return fmt.Errorf("marshal docs: %w", err)
	}

	// If total fits in one file, write directly
	if len(data) <= mongoGISChunkSize {
		return os.WriteFile(path, data, 0644)
	}

	// Need to chunk — re-marshal each doc individually
	docSlice := reflect.ValueOf(docs)
	if docSlice.Kind() != reflect.Slice {
		return fmt.Errorf("writeChunkedMongoJSON: expected slice, got %s", docSlice.Kind())
	}

	dir := filepathDir(path)
	base := filepathBase(path)
	ext := filepathExt(base)
	nameNoExt := base[:len(base)-len(ext)]

	// Marshal each doc individually to get sizes
	type serializedDoc struct {
		bytes []byte
		size  int
	}
	var serialized []serializedDoc
	for i := 0; i < docSlice.Len(); i++ {
		docBytes, err := json.MarshalIndent(docSlice.Index(i).Interface(), "", " ")
		if err != nil {
			return fmt.Errorf("marshal doc %d: %w", i, err)
		}
		serialized = append(serialized, serializedDoc{docBytes, len(docBytes)})
	}

	// Split into chunks
	var chunks [][]serializedDoc
	currentChunk := []serializedDoc{}
	currentSize := 0
	for _, s := range serialized {
		if currentSize+s.size > mongoGISChunkSize && len(currentChunk) > 0 {
			chunks = append(chunks, currentChunk)
			currentChunk = []serializedDoc{}
			currentSize = 0
		}
		currentChunk = append(currentChunk, s)
		currentSize += s.size
	}
	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	log.Printf("📦 [MongoDB] GIS JSON chunked: %d docs → %d files (max %d MB each)",
		len(serialized), len(chunks), mongoGISChunkSize/1024/1024)

	// Write each chunk file as a JSON array
	var chunkNames []string
	for i, chunk := range chunks {
		chunkName := fmt.Sprintf("%s_part_%02d%s", nameNoExt, i+1, ext)
		chunkPath := fmt.Sprintf("%s/%s", dir, chunkName)
		chunkNames = append(chunkNames, chunkName)

		file, err := os.OpenFile(chunkPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("create chunk file %s: %w", chunkPath, err)
		}
		writer := bufio.NewWriter(file)
		writer.WriteByte('[')
		for j, s := range chunk {
			if j > 0 {
				writer.WriteByte(',')
			}
			writer.Write(s.bytes)
		}
		writer.WriteByte(']')
		if err := writer.Flush(); err != nil {
			file.Close()
			return fmt.Errorf("flush chunk file %s: %w", chunkPath, err)
		}
		file.Close()

		chunkSize := 0
		for _, s := range chunk {
			chunkSize += s.size
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

// writeMongoGISIndexScript writes the create_indexes.js file.
func writeMongoGISIndexScript(path string) error {
	content := `// MongoDB GIS Index Creation Script
// Run with: mongosh vn_provinces create_indexes.js

// provinces-gis collection indexes
db.getCollection('provinces-gis').createIndex({ "Code": 1 }, { unique: true });
db.getCollection('provinces-gis').createIndex({ "GIS.Geometry": "2dsphere" });
db.getCollection('provinces-gis').createIndex({ "GIS.Center": "2dsphere" });
db.getCollection('provinces-gis').createIndex({ "SearchKeywords": 1 });

// wards-gis collection indexes
db.getCollection('wards-gis').createIndex({ "Code": 1 }, { unique: true });
db.getCollection('wards-gis').createIndex({ "ProvinceCode": 1 });
db.getCollection('wards-gis').createIndex({ "GIS.Geometry": "2dsphere" });
db.getCollection('wards-gis').createIndex({ "GIS.Center": "2dsphere" });
db.getCollection('wards-gis').createIndex({ "SearchKeywords": 1 });
`
	return os.WriteFile(path, []byte(content), 0644)
}

// writeMongoGISReadme writes the README.md for the MongoDB GIS dataset.
func writeMongoGISReadme(path string) error {
	loc, err := time.LoadLocation("Asia/Saigon")
	if err != nil {
		loc = time.FixedZone("GMT+7", 7*60*60)
	}
	createdAt := time.Now().In(loc).Format(time.RFC1123Z)

	content := `# Vietnamese Provinces Database — MongoDB GIS Dataset

Created at:  ` + createdAt + `

## Overview

This dataset provides Vietnamese provinces and wards in MongoDB document format
with two GIS collections:

| Collection | Documents | Description |
|------------|-----------|-------------|
| ` + "`provinces-gis`" + ` | 34 | Province documents with GIS geometry (bounding boxes + GeoJSON polygons) |
| ` + "`wards-gis`" + ` | 3,321 | Standalone ward documents with GIS geometry + ProvinceCode reference |

## Document Structure

### Province GIS Document

- **Core fields**: Code, Name, NameEn, FullName, FullNameEn, CodeName
- **` + "`AdministrativeUnit`" + `**: Embedded administrative unit object
- **` + "`SearchKeywords`" + `**: Pre-computed autocomplete keywords
- **` + "`GIS`" + `**: Center (GeoJSON Point), BoundingBox, Geometry (GeoJSON MultiPolygon), Properties
- **` + "`Meta`" + `**: Dataset version metadata

### Ward GIS Document

- Same structure as province, plus **` + "`ProvinceCode`" + `** for cross-collection joins

## Quick Start

### 1. Import the Data

` + "```bash" + `
# Import province GIS data
mongoimport --db vn_provinces --collection provinces_gis --file mongo_data_vn_province_gis_*.json --jsonArray

# Import ward GIS data (may be chunked — import each part sequentially)
mongoimport --db vn_provinces --collection wards_gis --file mongo_data_vn_ward_gis_*.json --jsonArray
` + "```" + `

> **Note**: If the ward GIS file was chunked, you'll see multiple files like
> ` + "`mongo_data_vn_ward_gis_*_part_01.json`" + `, ` + "`part_02.json`" + `, etc.
> Import each part individually, or use a script that reads the manifest file
> (` + "`*.manifest`" + `) for automated sequential import.

### 2. Create Indexes

` + "```bash" + `
mongosh vn_provinces create_indexes.js
` + "```" + `

### 3. Example Queries

` + "```javascript" + `
// Find province containing a point
db.getCollection('provinces-gis').findOne({
  "GIS.Geometry": {
    $geoIntersects: {
      $geometry: { type: "Point", coordinates: [105.8542, 21.0285] }
    }
  }
})

// Find all wards in a province
db.getCollection('wards-gis').find({ ProvinceCode: "01" })

// Find ward containing a point
db.getCollection('wards-gis').findOne({
  "GIS.Geometry": {
    $geoIntersects: {
      $geometry: { type: "Point", coordinates: [105.8231, 21.0347] }
    }
  }
})

// Find provinces near a point (within 50km)
db.getCollection('provinces-gis').find({
  "GIS.Center": {
    $near: {
      $geometry: { type: "Point", coordinates: [105.8542, 21.0285] },
      $maxDistance: 50000
    }
  }
})

// Join wards with provinces using $lookup
db.getCollection('wards-gis').aggregate([
  { $match: { ProvinceCode: "01" } },
  { $lookup: {
      from: "provinces-gis",
      localField: "ProvinceCode",
      foreignField: "Code",
      as: "Province"
  }}
])
` + "```" + `

## File Listing

| File | Description |
|------|-------------|
| ` + "`mongo_data_vn_province_gis_*.json`" + ` | Province GIS documents (JSON array) |
| ` + "`mongo_data_vn_ward_gis_*.json`" + ` | Ward GIS documents (JSON array, may be chunked) |
| ` + "`create_indexes.js`" + ` | Index creation script for both collections |
`
	return os.WriteFile(path, []byte(content), 0644)
}
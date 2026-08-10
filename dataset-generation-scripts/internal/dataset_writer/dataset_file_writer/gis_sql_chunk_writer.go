package dataset_writer

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

// maxSQLGISChunkSize is the maximum size of a single GIS SQL chunk file.
// Matches the Elasticsearch maxNDJSONChunkSize (40 MB) and stays safely under
// GitHub's 50 MB file warning. var (not const) so tests can override it.
var maxSQLGISChunkSize = 40 * 1024 * 1024 // 40 MB

// chunkHeaderInfo carries the fixed text rendered into every chunk's leading
// SQL comment. The part/total numbers are interpolated per chunk at write time.
type chunkHeaderInfo struct {
	Banner     string
	CreatedAt  string
	Repository string
}

// writeChunkedSQLFile writes complete-SQL blocks as chunk files, each under
// maxSQLGISChunkSize. It always emits chunk files — never a single file — and
// writes a manifest at path + ".manifest" listing chunk filenames in order.
// Chunks are named <base>-part-NN<ext> (e.g. postgresql_ImportData_gis_<ts>-part-01.sql),
// matching the Elasticsearch naming convention.
//
// Every chunk starts with a self-describing SQL header comment containing the
// banner, "Part X of N", created-at timestamp, and repository link.
func writeChunkedSQLFile(path string, blocks [][]byte, header chunkHeaderInfo) error {
	if len(blocks) == 0 {
		return nil
	}

	// Split into chunks greedily at block boundaries (blocks are atomic).
	var chunks [][][]byte
	currentChunk := [][]byte{}
	currentSize := 0
	for _, b := range blocks {
		if currentSize+len(b) > maxSQLGISChunkSize && len(currentChunk) > 0 {
			chunks = append(chunks, currentChunk)
			currentChunk = [][]byte{}
			currentSize = 0
		}
		currentChunk = append(currentChunk, b)
		currentSize += len(b)
	}
	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	dir := filepathDir(path)
	base := filepathBase(path)
	ext := filepathExt(base)
	nameNoExt := base[:len(base)-len(ext)]

	log.Printf("📦 [GIS SQL] chunked: %d blocks → %d files (max %d MB each)",
		len(blocks), len(chunks), maxSQLGISChunkSize/1024/1024)

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
		headerLines := []string{
			fmt.Sprintf("/* === %s === */\n", header.Banner),
			fmt.Sprintf("/* Part %d of %d */\n", i+1, len(chunks)),
			fmt.Sprintf("/* Created at:  %s */\n", header.CreatedAt),
			fmt.Sprintf("/* Reference: %s */\n", header.Repository),
			"/* =============================================== */\n\n",
		}
		for _, line := range headerLines {
			if _, err := writer.WriteString(line); err != nil {
				file.Close()
				return fmt.Errorf("write header line: %w", err)
			}
		}
		for _, b := range chunk {
			if _, err := writer.Write(b); err != nil {
				file.Close()
				return fmt.Errorf("write block: %w", err)
			}
		}
		if err := writer.Flush(); err != nil {
			file.Close()
			return fmt.Errorf("flush chunk file %s: %w", chunkPath, err)
		}
		file.Close()

		chunkSize := 0
		for _, b := range chunk {
			chunkSize += len(b)
		}
		log.Printf("   %s: %.1f MB, %d blocks", chunkName, float64(chunkSize)/1024/1024, len(chunk))
	}

	manifestPath := path + ".manifest"
	manifestContent := stringsJoin(chunkNames, "\n") + "\n"
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		return fmt.Errorf("write manifest file: %w", err)
	}
	log.Printf("   Manifest: %s", filepathBase(manifestPath))

	return nil
}

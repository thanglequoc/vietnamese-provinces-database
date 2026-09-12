package dto

// DatasetMetadataDocument is the single metadata document emitted by the JSON,
// MongoDB, and Elasticsearch writers. Field names use PascalCase to match the
// other exports of those formats.
type DatasetMetadataDocument struct {
	DatasetVersion string
	LatestDecree   string
	GeneratedAt    string
}

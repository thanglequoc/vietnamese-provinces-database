package model

// WKTComparisonResult represents differences in WKT fields
type WKTComparisonResult struct {
	RecordIdentifier string // e.g., "tinh34.7" or "xa3321.3285"
	RecordName       string // Vietnamese name
	FieldName        string // "bbox_wkt" or "geom_wkt"
	ExpectedWKT      string // From dump temp table
	ActualWKT        string // From current table
	Difference       string // Human-readable diff description
}

// ComparisonSummary aggregates comparison results
type ComparisonSummary struct {
	TableName         string
	TotalRecords      int
	MatchedRecords    int
	MismatchedRecords int
	MissingInDB       []string // GIS server IDs in dump but not in DB
	MissingInDump     []string // GIS server IDs in DB but not in dump
	Differences       []WKTComparisonResult
}

// IsEqual returns true if no differences found
func (s *ComparisonSummary) IsEqual() bool {
	return s.MismatchedRecords == 0 &&
		len(s.MissingInDB) == 0 &&
		len(s.MissingInDump) == 0
}

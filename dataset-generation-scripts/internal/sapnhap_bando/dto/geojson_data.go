package dto

// GISProvinceData represents processed GIS data for a province
type GISProvinceData struct {
	STT                      int
	Ten                      string
	TruocSN                  string
	GISServerID              string
	SapNhapProvinceMaTinh    string
	BBoxWKT                  string
	GeomWKT                  string
}

// GISWardData represents processed GIS data for a ward
type GISWardData struct {
	STT             int
	Ten             string
	TruocSN         string
	GISServerID     string
	SapNhapWardMaXa string
	BBoxWKT         string
	GeomWKT         string
}

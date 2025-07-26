package dto

type JsonProvinceModel struct {
	Type       string
	Code       string
	Name       string
	NameEn     string
	FullName   string
	FullNameEn string
	CodeName   string

	// Administrative Unit props
	AdministrativeUnitId          int
	AdministrativeUnitShortName   string
	AdministrativeUnitFullName    string
	AdministrativeUnitShortNameEn string
	AdministrativeUnitFullNameEn  string

	Wards []JsonWardModel
}

type JsonWardModel struct {
	Type         string
	Code         string
	Name         string
	NameEn       string
	FullName     string
	FullNameEn   string
	CodeName     string
	ProvinceCode string

	// Administrative Unit props
	AdministrativeUnitId          int
	AdministrativeUnitShortName   string
	AdministrativeUnitFullName    string
	AdministrativeUnitShortNameEn string
	AdministrativeUnitFullNameEn  string
}

// JSON Simplified version
type JsonProvinceSimplifiedModel struct {
	Code       string
	Name       string
	NameEn     string
	FullName   string
	FullNameEn string
	CodeName   string
	Wards      []JsonWardSimplifiedModel
}

type JsonWardSimplifiedModel struct {
	Code         string
	Name         string
	NameEn       string
	FullName     string
	FullNameEn   string
	CodeName     string
	ProvinceCode string
}

// VN only Simplified version
type JsonProvinceVNSimplifiedModel struct {
	Code     string
	FullName string
	Wards    []JsonWardVNSimplifiedModel
}

type JsonWardVNSimplifiedModel struct {
	Code         string
	FullName     string
	ProvinceCode string
}

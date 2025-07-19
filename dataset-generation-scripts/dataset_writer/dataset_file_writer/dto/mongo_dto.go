package dto

type MongoProvinceModel struct {
	Type string
	Code string
	Name string
	NameEn string
	FullName string
	FullNameEn string
	CodeName string
	AdministrativeUnitId int

	Wards []MongoWardModel
}

type MongoWardModel struct {
	Type string
	Code string
	Name string
	NameEn string
	FullName string
	FullNameEn string
	CodeName string
	ProvinceCode string
	AdministrativeUnitId int
}

package common

import (
	"github.com/uptrace/bun"
)

type AdministrativeUnit struct {
	bun.BaseModel `bun:"table:administrative_units,alias:au"`
	Id int `bun:"id,pk"`
	FullName string `bun:"full_name"`
	FullNameEn string `bun:"full_name_en"`
	ShortName string `bun:"short_name"`
	ShortNameEn string `bun:"short_name_en"`
	CodeName string `bun:"code_name"`
	CodeNameEn string `bun:"code_name_en"`
}

type AdministrativeRegion struct {
	bun.BaseModel `bun:"table:administrative_regions,alias:ar"`
	Id int `bun:"id,pk"`
	Name string `bun:"name"`
	NameEn string `bun:"name_en"`
	CodeName string `bun:"code_name"`
	CodeNameEn string `bun:"code_name_en"`
}

type Province struct {
	bun.BaseModel `bun:"table:provinces_tmp,alias:p"`
	Code string `bun:"code,pk"`
	Name string `bun:"name,notnull"`
	NameEn string `bun:"name_en,notnull"`
	FullName string `bun:"full_name,notnull"`
	FullNameEn string `bun:"full_name_en"`
	CodeName string `bun:"code_name"`
	AdministrativeUnitId int `bun:"administrative_unit_id"`

	// Province has many Wards
	Ward []*Ward `bun:"rel:has-many,join:code=province_code"`
	AdministrativeUnit AdministrativeUnit `bun:"rel:belongs-to,join:administrative_unit_id=id"`
}

type Ward struct {
	bun.BaseModel `bun:"table:wards_tmp,alias:w"`
	Code string `bun:"code,pk"`
	Name string `bun:"name,notnull"`
	NameEn string `bun:"name_en,notnull"`
	FullName string `bun:"full_name,notnull"`
	FullNameEn string `bun:"full_name_en"`
	CodeName string `bun:"code_name"`
	ProvinceCode string `bun:"province_code"`
	AdministrativeUnitId int `bun:"administrative_unit_id"`
	AdministrativeUnit AdministrativeUnit `bun:"rel:belongs-to,join:administrative_unit_id=id"`
}

type SeedProvince struct {
	bun.BaseModel `bun:"table:provinces_tmp_seed,alias:ps"`
	Code string `bun:"code,pk"`
	Name string `bun:"name"`

	// Province has many wards
	SeedWard []*SeedWard `bun:"rel:has-many,join:code=province_code"`
}

type SeedWard struct {
	bun.BaseModel `bun:"table:wards_tmp_seed,alias:ws"`
	Code string `bun:"code,pk"`
	Name string `bun:"name"`
	ProvinceCode string `bun:"province_code"`
}

package model

import "github.com/uptrace/bun"


type SeedProvince struct {
	bun.BaseModel `bun:"table:provinces_tmp_seed,alias:ps"`
	Code string `bun:"code,pk"`
	Name string `bun:"name"`

	// Province has many wards
	SeedWards []*SeedWard `bun:"rel:has-many,join:code=province_code"`
}

type SeedWard struct {
	bun.BaseModel `bun:"table:wards_tmp_seed,alias:ws"`
	Code string `bun:"code,pk"`
	Name string `bun:"name"`
	ProvinceCode string `bun:"province_code"`
}


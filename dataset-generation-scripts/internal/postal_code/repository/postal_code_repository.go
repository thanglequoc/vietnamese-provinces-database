package repository

import (
	"context"

	"github.com/thanglequoc-vn-provinces/v2/internal/vn_provinces_tmp/model"
	"github.com/uptrace/bun"
)

type PostalCodeRepository struct {
	db bun.IDB
}

func NewPostalCodeRepository(db bun.IDB) *PostalCodeRepository {
	return &PostalCodeRepository{db: db}
}

func (r *PostalCodeRepository) UpdateProvincePostalCodePrefix(ctx context.Context, code, prefix string) error {
	_, err := r.db.NewUpdate().
		Model((*model.Province)(nil)).
		Set("postal_code_prefix = ?", prefix).
		Where("code = ?", code).
		Exec(ctx)
	return err
}

func (r *PostalCodeRepository) UpdateWardPostalCode(ctx context.Context, wardCode, postalCode string) error {
	_, err := r.db.NewUpdate().
		Model((*model.Ward)(nil)).
		Set("postal_code = ?", postalCode).
		Where("code = ?", wardCode).
		Exec(ctx)
	return err
}

func (r *PostalCodeRepository) CountProvincesMissingPostalPrefix(ctx context.Context) (int, error) {
	return r.db.NewSelect().
		Model((*model.Province)(nil)).
		Where("postal_code_prefix IS NULL OR postal_code_prefix = ''").
		Count(ctx)
}

func (r *PostalCodeRepository) CountWardsMissingPostalCode(ctx context.Context) (int, error) {
	return r.db.NewSelect().
		Model((*model.Ward)(nil)).
		Where("postal_code IS NULL OR postal_code = ''").
		Count(ctx)
}

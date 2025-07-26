package repository

import 
(
	"log"
	"context"
	"github.com/uptrace/bun"
	"github.com/thanglequoc-vn-provinces/v2/internal/dumper/model"
)

type SeedDataRepository struct {
	db *bun.DB
}

func NewSeedDataRepository(db *bun.DB) *SeedDataRepository {
	return &SeedDataRepository{
		db: db,
	}
}

// Get all decree provinces
func (r *SeedDataRepository) GetAllSeedProvinces() []model.SeedProvince {
	var result []model.SeedProvince
	ctx := context.Background()
	err := r.db.NewSelect().Model(&result).Scan(ctx)
	if err != nil {
		log.Fatal("Unable to query seed provinces", err)
	}
	return result
}

func (r *SeedDataRepository) GetAllSeedWards() []model.SeedWard {
	var result []model.SeedWard
	ctx := context.Background()
	err := r.db.NewSelect().Model(&result).Scan(ctx)
	if err != nil {
		log.Fatal("Unable to query seed wards", err)
	}
	return result
}

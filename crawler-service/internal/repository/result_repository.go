package repository

import (
	"context"

	"github.com/namnv2496/crawler/internal/domain"
	"gorm.io/gorm"
)

type IResultRepository interface {
	CreateResult(ctx context.Context, url *domain.Result) error
}

type ResultRepository struct {
	db *gorm.DB
}

func NewResultRepository(
	dbSource IRepository,
) *ResultRepository {
	return &ResultRepository{
		db: dbSource.GetDB(),
	}
}

func (r *ResultRepository) CreateResult(ctx context.Context, result *domain.Result) error {
	err := r.db.WithContext(ctx).Model(&domain.Result{}).Create(result).Error
	if err != nil {
		return err
	}
	return nil
}

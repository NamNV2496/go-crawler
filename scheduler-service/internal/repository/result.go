package repository

import (
	"context"
	"time"

	"github.com/namnv2496/scheduler/internal/domain"
)

type IResultRepository interface {
	IRepository[domain.Result]
	CreateResult(ctx context.Context, url *domain.Result) error
}

type ResultRepository struct {
	baseRepository[domain.Result]
}

func NewResultRepository(
	dbSource IDatabase,
) *ResultRepository {
	return &ResultRepository{
		baseRepository: newBaseRepository[domain.Result](dbSource.GetDB(), 5*time.Second),
	}
}

func (_self *ResultRepository) CreateResult(ctx context.Context, result *domain.Result) error {
	return _self.InsertOnce(ctx, result)
}

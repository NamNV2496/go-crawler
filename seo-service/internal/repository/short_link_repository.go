package repository

import (
	"context"

	"github.com/namnv2496/seo/internal/domain"
	"gorm.io/gorm"
)

type IShortLinkRepo interface {
	IRepository[domain.ShortLink]
	GetShortLinks(ctx context.Context, offset, limit int, request map[string]string) ([]*domain.ShortLink, error)
}

type ShortLinkRepo struct {
	baseRepository[domain.ShortLink]
}

func NewShortLinkRepo(
	database IDatabase,
) *ShortLinkRepo {
	database.GetDB().AutoMigrate(&domain.ShortLink{})
	return &ShortLinkRepo{
		baseRepository: newBaseRepository[domain.ShortLink](database.GetDB()),
	}
}

var _ IShortLinkRepo = &ShortLinkRepo{}

func (_self *ShortLinkRepo) GetShortLinks(ctx context.Context, offset, limit int, request map[string]string) ([]*domain.ShortLink, error) {
	opts := make([]QueryOptionFunc, 0)
	// dynamic filter
	for field, value := range request {
		opts = append(opts, func(db *gorm.DB) *gorm.DB {
			return db.Where(field+" = ?", value)
		})
	}
	if offset > 0 && limit > 0 {
		opts = append(opts, WithOffset(offset))
		opts = append(opts, WithLimit(limit))
	} else {
		opts = append(opts, WithLimit(20))
		opts = append(opts, WithOffset(0))
	}
	urls, err := _self.Finds(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return urls, nil
}

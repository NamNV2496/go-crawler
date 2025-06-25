package repository

import (
	"context"

	"github.com/namnv2496/seo/internal/domain"
	"gorm.io/gorm"
)

type IUrlRepo interface {
	IRepository[domain.Url]
	CreateUrl(ctx context.Context, tx *gorm.DB, url domain.Url) (int64, error)
	GetUrl(ctx context.Context, url string) (*domain.Url, error)
	GetUrls(ctx context.Context, offset, limit int) ([]*domain.Url, error)
	UpdateUrl(ctx context.Context, tx *gorm.DB, url domain.Url) error
	DeleteUrl(ctx context.Context, tx *gorm.DB, url string) error
}

type UrlRepo struct {
	baseRepository[domain.Url]
}

func NewUrlRepo(
	database IDatabase,
) *UrlRepo {
	database.GetDB().AutoMigrate(&domain.Url{})
	return &UrlRepo{
		baseRepository: newBaseRepository[domain.Url](database.GetDB()),
	}
}

var _ IUrlRepo = &UrlRepo{}

func (_self *UrlRepo) CreateUrl(ctx context.Context, tx *gorm.DB, url domain.Url) (int64, error) {
	err := _self.InsertOnce(ctx, url)
	return url.Id, err
}

func (_self *UrlRepo) GetUrl(ctx context.Context, url string) (*domain.Url, error) {
	var opts []QueryOptionFunc
	opts = append(opts, WithCondition("url =?", url))
	opts = append(opts, WithCondition("is_active = true"))
	return _self.Find(ctx, opts...)
}

func (_self *UrlRepo) GetUrls(ctx context.Context, offset, limit int) ([]*domain.Url, error) {
	var opts []QueryOptionFunc
	opts = append(opts, WithCondition("is_active = true"))
	opts = append(opts, WithOffset(offset))
	opts = append(opts, WithLimit(limit))

	urlData, err := _self.Finds(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return urlData, nil
}

func (_self *UrlRepo) UpdateUrl(ctx context.Context, tx *gorm.DB, url domain.Url) error {
	return _self.UpdateOnce(ctx, url)
}

func (_self *UrlRepo) DeleteUrl(ctx context.Context, tx *gorm.DB, url string) error {
	var opts []QueryOptionFunc
	opts = append(opts, func(db *gorm.DB) *gorm.DB {
		return db.Where("url = ?", url)
	})
	return _self.DeleteOnce(ctx, domain.Url{}, opts...)
}

package repository

import (
	"context"

	"github.com/namnv2496/crawler/internal/domain"
	"gorm.io/gorm"
)

type IUrlRepository interface {
	CreateUrl(ctx context.Context, url *domain.Url) (int64, error)
	GetUrls(ctx context.Context, limit, offset int32) ([]*domain.Url, error)
	UpdateUrl(ctx context.Context, url *domain.Url) error
	GetUrlByID(ctx context.Context, id int64) (*domain.Url, error)
	GetUrlByDomainAndQueue(ctx context.Context, urlDomain, queue string, limit, offset int) ([]*domain.Url, error)
	CountUrlByDomainAndQueue(ctx context.Context, urlDomain, queue string) (int64, error)
}

type UrlRepository struct {
	db *gorm.DB
}

func NewUrlRepository(
	dbSource IRepository,
) *UrlRepository {
	return &UrlRepository{
		db: dbSource.GetDB(),
	}
}

func (r *UrlRepository) CreateUrl(ctx context.Context, url *domain.Url) (int64, error) {
	result := r.db.WithContext(ctx).Create(url)
	if result.Error != nil {
		return 0, result.Error
	}
	return url.Id, nil
}

func (r *UrlRepository) GetUrls(ctx context.Context, limit, offset int32) ([]*domain.Url, error) {
	var urls []*domain.Url
	result := r.db.WithContext(ctx).Limit(int(limit)).Offset(int(offset)).Find(&urls)
	if result.Error != nil {
		return nil, result.Error
	}
	return urls, nil
}

func (r *UrlRepository) UpdateUrl(ctx context.Context, url *domain.Url) error {
	result := r.db.WithContext(ctx).Save(url)
	return result.Error
}

func (r *UrlRepository) GetUrlByID(ctx context.Context, id int64) (*domain.Url, error) {
	var url domain.Url
	result := r.db.WithContext(ctx).First(&url, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &url, nil
}

func (r *UrlRepository) GetUrlByDomainAndQueue(ctx context.Context, urlDomain, queue string, limit, offset int) ([]*domain.Url, error) {
	var urls []*domain.Url
	result := r.db.WithContext(ctx).
		Where("domain = ? AND queue = ? AND is_active=true", urlDomain, queue).
		Offset(offset).
		Limit(limit).
		Find(&urls)
	if result.Error != nil {
		return nil, result.Error
	}
	return urls, nil
}

func (r *UrlRepository) CountUrlByDomainAndQueue(ctx context.Context, urlDomain, queue string) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&domain.Url{}).
		Where("domain =? AND queue =? AND is_active=true", urlDomain, queue).
		Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}
	return count, nil
}

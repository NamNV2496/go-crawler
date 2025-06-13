package repository

import (
	"context"

	"github.com/namnv2496/crawler/internal/domain"
)

type IUrlRepository interface {
	IRepository[domain.Url]
	CreateUrl(ctx context.Context, url *domain.Url) (int64, error)
	GetUrls(ctx context.Context, limit, offset int32) ([]*domain.Url, error)
	UpdateUrl(ctx context.Context, url *domain.Url) error
	GetUrlByID(ctx context.Context, id int64) (*domain.Url, error)
	GetUrlByDomainsAndQueues(ctx context.Context, urlDomain, queue []string, limit, offset int) ([]*domain.Url, error)
	CountUrlByDomainsAndQueues(ctx context.Context, domains, queues []string) (int64, error)
}

type UrlRepository struct {
	baseRepository[domain.Url]
}

func NewUrlRepository(
	dbSource IDatabase,
) *UrlRepository {
	return &UrlRepository{
		baseRepository: newBaseRepository[domain.Url](dbSource.GetDB()),
	}
}

func (_self *UrlRepository) CreateUrl(ctx context.Context, url *domain.Url) (int64, error) {
	err := _self.InsertOnce(ctx, url)
	return url.Id, err
}

func (_self *UrlRepository) GetUrls(ctx context.Context, limit, offset int32) ([]*domain.Url, error) {
	var opts []QueryOptionFunc
	opts = append(opts, WithLimit((int(limit))))
	opts = append(opts, WithOffset((int(offset))))

	return _self.Finds(ctx, opts...)
}

func (_self *UrlRepository) UpdateUrl(ctx context.Context, url *domain.Url) error {
	var opts []QueryOptionFunc
	opts = append(opts, WithCondition("id = ?", url.Id))
	return _self.UpdateOnce(ctx, url, opts...)
}

func (_self *UrlRepository) GetUrlByID(ctx context.Context, id int64) (*domain.Url, error) {
	var opts []QueryOptionFunc
	opts = append(opts, WithCondition("id = ?", id))
	opts = append(opts, WithLimit(1))
	return _self.Find(ctx, opts...)
}

func (_self *UrlRepository) GetUrlByDomainsAndQueues(ctx context.Context, urlDomain, queue []string, limit, offset int) ([]*domain.Url, error) {
	var opts []QueryOptionFunc
	opts = append(opts, WithCondition("domain IN ? AND queue IN ? AND is_active=true", urlDomain, queue))
	opts = append(opts, WithLimit(int(limit)))
	opts = append(opts, WithOffset(int(offset)))

	return _self.Finds(ctx, opts...)
}

func (_self *UrlRepository) CountUrlByDomainsAndQueues(ctx context.Context, domains, queues []string) (int64, error) {
	var opts []QueryOptionFunc
	opts = append(opts, WithCondition("domain IN ? AND queue IN ? AND is_active=true", domains, queues))

	return _self.CountOnce(ctx, opts...)
}

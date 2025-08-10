package repository

import (
	"context"

	"github.com/namnv2496/scheduler/internal/domain"
)

type ICrawlerEventRepository interface {
	IRepository[domain.CrawlerEvent]
	CreateCrawlerEvent(ctx context.Context, crawlerEvent *domain.CrawlerEvent) (int64, error)
	GetCrawlerEvents(ctx context.Context, limit, offset int32) ([]*domain.CrawlerEvent, error)
	UpdateCrawlerEvent(ctx context.Context, url *domain.CrawlerEvent) error
	GetCrawlerEventByID(ctx context.Context, id int64) (*domain.CrawlerEvent, error)
	GetCrawlerEventByDomainAndQueue(ctx context.Context, urlDomain, queue string, limit, offset int) ([]*domain.CrawlerEvent, error)
	CountCrawlerEventByDomainsAndQueues(ctx context.Context, domains, queues []string) (int64, error)
	GetCrawlerEventByStatusAndSchedulerAt(ctx context.Context, status domain.StatusEnum, schedulerAt int64) ([]*domain.CrawlerEvent, error)
}

type CrawlerEventRepository struct {
	baseRepository[domain.CrawlerEvent]
}

func NewCrawlerEventRepository(
	dbSource IDatabase,
) ICrawlerEventRepository {
	return &CrawlerEventRepository{
		baseRepository: newBaseRepository[domain.CrawlerEvent](dbSource.GetDB()),
	}
}

func (_self *CrawlerEventRepository) CreateCrawlerEvent(ctx context.Context, url *domain.CrawlerEvent) (int64, error) {
	err := _self.InsertOnce(ctx, url)
	return url.Id, err
}

func (_self *CrawlerEventRepository) GetCrawlerEvents(ctx context.Context, limit, offset int32) ([]*domain.CrawlerEvent, error) {
	var opts []QueryOptionFunc
	opts = append(opts, WithLimit((int(limit))))
	opts = append(opts, WithOffset((int(offset))))

	return _self.Finds(ctx, opts...)
}

func (_self *CrawlerEventRepository) UpdateCrawlerEvent(ctx context.Context, url *domain.CrawlerEvent) error {
	var opts []QueryOptionFunc
	opts = append(opts, WithCondition("id = ?", url.Id))
	return _self.UpdateOnce(ctx, url, opts...)
}

func (_self *CrawlerEventRepository) GetCrawlerEventByID(ctx context.Context, id int64) (*domain.CrawlerEvent, error) {
	var opts []QueryOptionFunc
	opts = append(opts, WithCondition("id = ?", id))
	opts = append(opts, WithLimit(1))
	return _self.Find(ctx, opts...)
}

func (_self *CrawlerEventRepository) GetCrawlerEventByDomainAndQueue(ctx context.Context, urlDomain, queue string, limit, offset int) ([]*domain.CrawlerEvent, error) {
	var opts []QueryOptionFunc
	opts = append(opts, WithCondition("domain = ? AND queue = ? AND is_active=true", urlDomain, queue))
	opts = append(opts, WithLimit(int(limit)))
	opts = append(opts, WithOffset(int(offset)))

	return _self.Finds(ctx, opts...)
}

func (_self *CrawlerEventRepository) CountCrawlerEventByDomainsAndQueues(ctx context.Context, domains, queues []string) (int64, error) {
	var opts []QueryOptionFunc
	opts = append(opts, WithCondition("domain IN ? AND queue IN ? AND is_active=true", domains, queues))

	return _self.CountOnce(ctx, opts...)
}

func (_self *CrawlerEventRepository) GetCrawlerEventByStatusAndSchedulerAt(ctx context.Context, status domain.StatusEnum, schedulerAt int64) ([]*domain.CrawlerEvent, error) {
	var opts []QueryOptionFunc
	opts = append(opts, WithCondition("status = ? AND scheduler_at <= ? AND is_active=true AND repeat_times != 0", status, schedulerAt))
	return _self.Finds(ctx, opts...)
}

package repository

import (
	"context"

	"github.com/namnv2496/scheduler/internal/domain"
)

type ISchedulerEventRepository interface {
	IRepository[domain.SchedulerEvent]
	CreateSchedulerEvent(ctx context.Context, SchedulerEvent *domain.SchedulerEvent) (int64, error)
	GetSchedulerEvents(ctx context.Context, limit, offset int32) ([]*domain.SchedulerEvent, error)
	UpdateSchedulerEvent(ctx context.Context, url *domain.SchedulerEvent) error
	GetSchedulerEventByID(ctx context.Context, id int64) (*domain.SchedulerEvent, error)
	GetSchedulerEventByDomainAndQueue(ctx context.Context, urlDomain, queue string, limit, offset int) ([]*domain.SchedulerEvent, error)
	CountSchedulerEventByDomainsAndQueues(ctx context.Context, domains, queues []string) (int64, error)
	GetSchedulerEventByStatusAndSchedulerAt(ctx context.Context, status domain.StatusEnum, schedulerAt int64) ([]*domain.SchedulerEvent, error)
}

type SchedulerEventRepository struct {
	baseRepository[domain.SchedulerEvent]
}

func NewSchedulerEventRepository(
	dbSource IDatabase,
) ISchedulerEventRepository {
	return &SchedulerEventRepository{
		baseRepository: newBaseRepository[domain.SchedulerEvent](dbSource.GetDB()),
	}
}

func (_self *SchedulerEventRepository) CreateSchedulerEvent(ctx context.Context, url *domain.SchedulerEvent) (int64, error) {
	err := _self.InsertOnce(ctx, url)
	return url.Id, err
}

func (_self *SchedulerEventRepository) GetSchedulerEvents(ctx context.Context, limit, offset int32) ([]*domain.SchedulerEvent, error) {
	var opts []QueryOptionFunc
	opts = append(opts, WithLimit((int(limit))))
	opts = append(opts, WithOffset((int(offset))))

	return _self.Finds(ctx, opts...)
}

func (_self *SchedulerEventRepository) UpdateSchedulerEvent(ctx context.Context, url *domain.SchedulerEvent) error {
	var opts []QueryOptionFunc
	opts = append(opts, WithCondition("id = ?", url.Id))
	return _self.UpdateOnce(ctx, url, opts...)
}

func (_self *SchedulerEventRepository) GetSchedulerEventByID(ctx context.Context, id int64) (*domain.SchedulerEvent, error) {
	var opts []QueryOptionFunc
	opts = append(opts, WithCondition("id = ?", id))
	opts = append(opts, WithLimit(1))
	return _self.Find(ctx, opts...)
}

func (_self *SchedulerEventRepository) GetSchedulerEventByDomainAndQueue(ctx context.Context, urlDomain, queue string, limit, offset int) ([]*domain.SchedulerEvent, error) {
	var opts []QueryOptionFunc
	opts = append(opts, WithCondition("domain = ? AND queue = ? AND is_active=true", urlDomain, queue))
	opts = append(opts, WithLimit(int(limit)))
	opts = append(opts, WithOffset(int(offset)))

	return _self.Finds(ctx, opts...)
}

func (_self *SchedulerEventRepository) CountSchedulerEventByDomainsAndQueues(ctx context.Context, domains, queues []string) (int64, error) {
	var opts []QueryOptionFunc
	opts = append(opts, WithCondition("domain IN ? AND queue IN ? AND is_active=true", domains, queues))

	return _self.CountOnce(ctx, opts...)
}

func (_self *SchedulerEventRepository) GetSchedulerEventByStatusAndSchedulerAt(ctx context.Context, status domain.StatusEnum, schedulerAt int64) ([]*domain.SchedulerEvent, error) {
	var opts []QueryOptionFunc
	opts = append(opts, WithCondition("status = ? AND scheduler_at <= ? AND is_active=true AND repeat_times != 0", status, schedulerAt))
	return _self.Finds(ctx, opts...)
}

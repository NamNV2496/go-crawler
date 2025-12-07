package repository

import (
	"context"

	"github.com/namnv2496/scheduler/internal/configs"
	"github.com/namnv2496/scheduler/internal/domain"
	"gorm.io/gorm"
)

type ISchedulerEventRepository interface {
	IRepository[domain.SchedulerEvent]
	CreateSchedulerEvent(ctx context.Context, event *domain.SchedulerEvent) (int64, error)
	GetSchedulerEvents(ctx context.Context, limit, offset int32) ([]*domain.SchedulerEvent, error)
	UpdateSchedulerEvent(ctx context.Context, event *domain.SchedulerEvent) error
	UpdateSchedulerEvents(ctx context.Context, events []*domain.SchedulerEvent) error
	GetSchedulerEventByID(ctx context.Context, id int64) (*domain.SchedulerEvent, error)
	GetSchedulerEventByDomainAndQueue(ctx context.Context, urlDomain, queue string, limit, offset int) ([]*domain.SchedulerEvent, error)
	CountSchedulerEventByDomainsAndQueues(ctx context.Context, domains, queues []string) (int64, error)
	GetSchedulerEventByStatusAndSchedulerAt(ctx context.Context, status domain.StatusEnum, schedulerAt int64) ([]*domain.SchedulerEvent, error)
}

type SchedulerEventRepository struct {
	baseRepository[domain.SchedulerEvent]
	isolationLevel int
}

func NewSchedulerEventRepository(
	conf *configs.Config,
	dbSource IDatabase,
) ISchedulerEventRepository {
	return &SchedulerEventRepository{
		baseRepository: newBaseRepository[domain.SchedulerEvent](dbSource.GetDB(), conf.DatabaseConfig.Timeout),
		isolationLevel: conf.DatabaseConfig.IsolationLevel,
	}
}

func (_self *SchedulerEventRepository) CreateSchedulerEvent(ctx context.Context, event *domain.SchedulerEvent) (int64, error) {
	err := _self.InsertOnce(ctx, event)
	return event.Id, err
}

func (_self *SchedulerEventRepository) GetSchedulerEvents(ctx context.Context, limit, offset int32) ([]*domain.SchedulerEvent, error) {
	var opts []QueryOptionFunc
	opts = append(opts, WithLimit((int(limit))))
	opts = append(opts, WithOffset((int(offset))))

	return _self.Finds(ctx, opts...)
}

func (_self *SchedulerEventRepository) UpdateSchedulerEvent(ctx context.Context, event *domain.SchedulerEvent) error {
	var opts []QueryOptionFunc
	opts = append(opts, WithCondition("id = ?", event.Id))
	return _self.UpdateOnce(ctx, event, opts...)
}

// example
func (_self *SchedulerEventRepository) UpdateSchedulerEvents(ctx context.Context, events []*domain.SchedulerEvent) error {
	funcs := []FunctionExec{
		func(ctx context.Context, tx *gorm.DB) (isPass bool, err error) {
			for _, event := range events {
				var opts []QueryOptionFunc
				opts = append(opts, WithCondition("id = ?", event.Id))
				if err := _self.UpdateOnce(ctx, event, opts...); err != nil {
					return false, err
				}
			}
			return true, nil
		},
		func(ctx context.Context, tx *gorm.DB) (isPass bool, err error) {
			// another function for another tables
			return true, nil
		},
		func(ctx context.Context, tx *gorm.DB) (isPass bool, err error) {
			// another function for another tables
			return true, nil
		},
	}
	return _self.RunWithTransaction(
		ctx,
		"UpdateSchedulerEvents",
		funcs...)
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
	opts = append(opts, WithIsolationLevel(_self.isolationLevel))
	opts = append(opts, WithCondition("status = ? AND scheduler_at <= ? AND is_active=true AND repeat_times != 0", status, schedulerAt))
	opts = append(opts, WithForUpdate())
	return _self.Finds(ctx, opts...)
}

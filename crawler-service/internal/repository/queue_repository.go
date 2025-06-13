package repository

import (
	"context"

	"github.com/namnv2496/crawler/internal/domain"
)

type IQueueRepository interface {
	IRepository[domain.Queue]
	CreateQueue(ctx context.Context, queue *domain.Queue) (int64, error)
	GetQueuesByDomain(ctx context.Context, domains []string, limit, offset int32) ([]*domain.Queue, error)
	GetQueuesByDomainsAndQueue(ctx context.Context, domains []string, queue string, limit, offset int32) ([]*domain.Queue, error)
	UpdateQueue(ctx context.Context, queue *domain.Queue) error
	CountQueueByDomainsAndQueue(ctx context.Context, domains []string, queue string) (int64, error)
}

type QueueRepository struct {
	baseRepository[domain.Queue]
}

func NewQueueRepository(
	dbSource IDatabase,
) *QueueRepository {
	return &QueueRepository{
		baseRepository: newBaseRepository[domain.Queue](dbSource.GetDB()),
	}
}

func (_self *QueueRepository) CreateQueue(ctx context.Context, queue *domain.Queue) (int64, error) {
	err := _self.InsertOnce(ctx, queue)
	if err != nil {
		return 0, err
	}
	return queue.Id, nil
}

func (_self *QueueRepository) GetQueuesByDomain(ctx context.Context, domains []string, limit, offset int32) ([]*domain.Queue, error) {
	var opts []QueryOptionFunc
	opts = append(opts, WithCondition("is_active = true"))
	opts = append(opts, WithOffset(int(offset)))
	opts = append(opts, WithLimit(int(limit)))

	queues, err := _self.Finds(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return queues, nil
}

func (_self *QueueRepository) GetQueuesByDomainsAndQueue(ctx context.Context, domains []string, queue string, limit, offset int32) ([]*domain.Queue, error) {
	var opts []QueryOptionFunc
	opts = append(opts, WithOffset(int(offset)))
	opts = append(opts, WithLimit(int(limit)))
	opts = append(opts, WithCondition("is_active = true"))
	opts = append(opts, WithCondition("domain IN ? AND queue = ?", domains, queue))

	queues, err := _self.Finds(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return queues, nil
}

func (_self *QueueRepository) UpdateQueue(ctx context.Context, queue *domain.Queue) error {
	err := _self.UpdateOnce(ctx, queue)
	if err != nil {
		return err
	}
	return nil
}

func (_self *QueueRepository) CountQueueByDomainsAndQueue(ctx context.Context, domains []string, queue string) (int64, error) {
	var opts []QueryOptionFunc
	opts = append(opts, WithCondition("is_active = true"))
	opts = append(opts, WithCondition("domain IN ? AND queue = ?", domains, queue))

	return _self.CountOnce(ctx, opts...)
}

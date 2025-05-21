package repository

import (
	"context"

	"github.com/namnv2496/crawler/internal/domain"
	"gorm.io/gorm"
)

type IQueueRepository interface {
	CreateQueue(ctx context.Context, queue *domain.Queue) (int64, error)
	GetQueues(ctx context.Context, limit, offset int32) ([]*domain.Queue, error)
	UpdateQueue(ctx context.Context, queue *domain.Queue) error
	CountQueue(ctx context.Context) (int64, error)
}

type QueueRepository struct {
	db *gorm.DB
}

func NewQueueRepository(
	dbSource IRepository,
) *QueueRepository {
	return &QueueRepository{
		db: dbSource.GetDB(),
	}
}

func (r *QueueRepository) CreateQueue(ctx context.Context, queue *domain.Queue) (int64, error) {
	err := r.db.Create(queue).Error
	if err != nil {
		return 0, err
	}
	return queue.Id, nil
}
func (r *QueueRepository) GetQueues(ctx context.Context, limit, offset int32) ([]*domain.Queue, error) {
	var queues []*domain.Queue
	err := r.db.Limit(int(limit)).Offset(int(offset)).Find(&queues).Error
	if err != nil {
		return nil, err
	}
	return queues, nil
}

func (r *QueueRepository) UpdateQueue(ctx context.Context, queue *domain.Queue) error {
	err := r.db.Save(queue).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *QueueRepository) CountQueue(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.Model(&domain.Queue{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

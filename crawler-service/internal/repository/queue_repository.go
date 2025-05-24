package repository

import (
	"context"

	"github.com/namnv2496/crawler/internal/domain"
	"gorm.io/gorm"
)

type IQueueRepository interface {
	CreateQueue(ctx context.Context, queue *domain.Queue) (int64, error)
	GetQueuesByDomain(ctx context.Context, domains []string, limit, offset int32) ([]*domain.Queue, error)
	GetQueuesByDomainsAndQueue(ctx context.Context, domains []string, queue string, limit, offset int32) ([]*domain.Queue, error)
	UpdateQueue(ctx context.Context, queue *domain.Queue) error
	CountQueueByDomainsAndQueue(ctx context.Context, domains []string, queue string) (int64, error)
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

func (r *QueueRepository) GetQueuesByDomain(ctx context.Context, domains []string, limit, offset int32) ([]*domain.Queue, error) {
	var queues []*domain.Queue
	tx := r.db.Limit(int(limit)).
		Offset(int(offset)).Where("is_active =?", true)
	if len(domains) > 0 {
		tx = tx.Where("domain IN?", domains)
	}
	err := tx.Find(&queues).Error
	if err != nil {
		return nil, err
	}
	return queues, nil
}

func (r *QueueRepository) GetQueuesByDomainsAndQueue(ctx context.Context, domains []string, queue string, limit, offset int32) ([]*domain.Queue, error) {
	var queues []*domain.Queue

	tx := r.db.Limit(int(limit)).
		Offset(int(offset)).Where("is_active = ?", true)
	if len(domains) > 0 {
		tx = tx.Where("domain IN ? AND queue = ?", domains, queue)
	}

	err := tx.Find(&queues).Error
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

func (r *QueueRepository) CountQueueByDomainsAndQueue(ctx context.Context, domains []string, queue string) (int64, error) {
	var count int64
	err := r.db.Model(&domain.Queue{}).Where("domain in ? AND queue = ?", domains, queue).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

package service

import (
	"context"

	"github.com/namnv2496/crawler/internal/domain"
	"github.com/namnv2496/crawler/internal/repository"
)

type IQueueService interface {
	CreateQueue(ctx context.Context, queue *domain.Queue) (int64, error)
	GetQueues(ctx context.Context, limit, offset int32) ([]*domain.Queue, error)
	UpdateQueue(ctx context.Context, queue *domain.Queue) error
}

type QueueService struct {
	queueRepo repository.IQueueRepository
}

func NewQueueService(
	queueRepo repository.IQueueRepository,
) *QueueService {
	return &QueueService{
		queueRepo: queueRepo,
	}
}

var _ IQueueService = &QueueService{}

func (_self *QueueService) CreateQueue(ctx context.Context, queue *domain.Queue) (int64, error) {
	return _self.queueRepo.CreateQueue(ctx, queue)
}
func (_self *QueueService) GetQueues(ctx context.Context, limit, offset int32) ([]*domain.Queue, error) {
	return _self.queueRepo.GetQueuesByDomain(ctx, nil, limit, offset)
}
func (_self *QueueService) UpdateQueue(ctx context.Context, queue *domain.Queue) error {
	return _self.queueRepo.UpdateQueue(ctx, queue)
}

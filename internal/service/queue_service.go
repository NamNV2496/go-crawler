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

func (s *QueueService) CreateQueue(ctx context.Context, queue *domain.Queue) (int64, error) {
	return s.queueRepo.CreateQueue(ctx, queue)
}
func (s *QueueService) GetQueues(ctx context.Context, limit, offset int32) ([]*domain.Queue, error) {
	return s.queueRepo.GetQueuesByDomain(ctx, nil, limit, offset)
}
func (s *QueueService) UpdateQueue(ctx context.Context, queue *domain.Queue) error {
	return s.queueRepo.UpdateQueue(ctx, queue)
}

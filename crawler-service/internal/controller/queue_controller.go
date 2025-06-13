package controller

import (
	"context"

	domain "github.com/namnv2496/crawler/internal/domain"
	"github.com/namnv2496/crawler/internal/service"
	crawlerv1 "github.com/namnv2496/crawler/pkg/generated/pkg/proto"
)

type QueueController struct {
	crawlerv1.UnimplementedQueueServiceServer
	queueService service.IQueueService
}

func NewQueueController(
	queueService service.IQueueService,
) crawlerv1.QueueServiceServer {
	return &QueueController{
		queueService: queueService,
	}
}

func (_self *QueueController) CreateQueue(ctx context.Context, req *crawlerv1.CreateQueueRequest) (*crawlerv1.CreateQueueResponse, error) {
	id, err := _self.queueService.CreateQueue(ctx, &domain.Queue{
		Queue:    req.Queue.Queue,
		Domain:   req.Queue.Domain,
		Cron:     req.Queue.Cron,
		Quantity: req.Queue.Quantity,
		IsActive: true,
	})
	if err != nil {
		return nil, err
	}
	return &crawlerv1.CreateQueueResponse{
		Id:     id,
		Status: "success",
	}, nil
}
func (_self *QueueController) GetQueues(ctx context.Context, req *crawlerv1.GetQueuesRequest) (*crawlerv1.GetQueuesResponse, error) {
	queues, err := _self.queueService.GetQueues(ctx, int32(req.Limit), int32(req.Offset))
	if err != nil {
		return nil, err
	}
	queuesRes := make([]*crawlerv1.Queue, 0)
	for _, queue := range queues {
		queuesRes = append(queuesRes, &crawlerv1.Queue{
			Id:       queue.Id,
			Queue:    queue.Queue,
			Domain:   queue.Domain,
			Cron:     queue.Cron,
			Quantity: queue.Quantity,
			IsActive: queue.IsActive,
		})
	}
	return &crawlerv1.GetQueuesResponse{
		Queues: queuesRes,
	}, nil
}
func (_self *QueueController) UpdateQueue(ctx context.Context, req *crawlerv1.UpdateQueueRequest) (*crawlerv1.UpdateQueueResponse, error) {
	err := _self.queueService.UpdateQueue(ctx, &domain.Queue{
		Queue:    req.Queue.Queue,
		Domain:   req.Queue.Domain,
		Cron:     req.Queue.Cron,
		Quantity: req.Queue.Quantity,
		IsActive: true,
	})
	if err != nil {
		return nil, err
	}
	return &crawlerv1.UpdateQueueResponse{
		Status: "success",
	}, nil
}

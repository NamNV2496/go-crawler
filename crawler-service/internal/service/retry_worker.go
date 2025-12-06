package service

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/namnv2496/crawler/internal/entity"
	"github.com/namnv2496/crawler/internal/service/mq"
)

type IRetryWorker interface {
	Start(ctx context.Context)
}

type retryWorker struct {
	asynqConsumer mq.IAsynqConsumer
	crawlService  ICrawlerService
}

func NewRetryWorker(
	asynqConsumer mq.IAsynqConsumer,
	crawlService ICrawlerService,
) IRetryWorker {
	return &retryWorker{
		asynqConsumer: asynqConsumer,
		crawlService:  crawlService,
	}
}

func (_self *retryWorker) Start(ctx context.Context) {
	_self.asynqConsumer.RegisterHandler(mq.RetryEvent, _self.RetryEventHandler)

	// start server
	_self.asynqConsumer.Run()
}

func (_self *retryWorker) RetryEventHandler(ctx context.Context, task *asynq.Task) error {
	var event entity.CrawlerEvent
	if err := json.Unmarshal(task.Payload(), &event); err != nil {
		return err
	}
	return _self.crawlService.Crawl(ctx, event)
}

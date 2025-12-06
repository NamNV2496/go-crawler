package mq

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
	"github.com/namnv2496/crawler/internal/configs"
	"github.com/namnv2496/crawler/internal/pkg/logging"
)

const (
	RetryEvent = "retry_crawl"
)

type IAsynqProducer interface {
	EnqueueRetryEvent(ctx context.Context, event any, processAt time.Time) error
}

type asynqProducer struct {
	client *asynq.Client
}

func NewAsynqProducer(conf *configs.Config) IAsynqProducer {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     conf.Redis.Addr,
		DB:       10,
		Password: conf.Redis.Password,
	})
	return &asynqProducer{
		client: client,
	}
}

var _ IAsynqProducer = &asynqProducer{}

func (_self *asynqProducer) EnqueueRetryEvent(ctx context.Context, event any, processAt time.Time) error {
	deferFunc := logging.AppendPrefix("EnqueueEvent")
	defer deferFunc()
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	task := asynq.NewTask(RetryEvent, payload)
	taskInfor, err := _self.client.EnqueueContext(ctx, task, asynq.ProcessAt(processAt), asynq.MaxRetry(1))
	if err != nil {
		return err
	}
	logging.Info(ctx, "%s", taskInfor.ID)
	return nil
}

package mq

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/namnv2496/crawler/internal/configs"
)

type AsynqHandlerFunc func(ctx context.Context, task *asynq.Task) error

type IAsynqConsumer interface {
	RegisterHandler(taskName string, handlerFunc AsynqHandlerFunc)
	Run()
}

type asynqConsumer struct {
	server  *asynq.Server
	handler *asynq.ServeMux
}

func NewAsynqConsumer(
	conf *configs.Config,
) IAsynqConsumer {
	server := asynq.NewServer(asynq.RedisClientOpt{
		Addr:     conf.Redis.Addr,
		DB:       10,
		Password: conf.Redis.Password,
	}, asynq.Config{
		Concurrency: 1,
	})
	return &asynqConsumer{
		server:  server,
		handler: asynq.NewServeMux(),
	}
}

var _ IAsynqConsumer = &asynqConsumer{}

func (_self *asynqConsumer) Run() {
	_self.server.Run(_self.handler)
}

func (_self *asynqConsumer) RegisterHandler(taskName string, handlerFunc AsynqHandlerFunc) {
	_self.handler.HandleFunc(taskName, handlerFunc)
}

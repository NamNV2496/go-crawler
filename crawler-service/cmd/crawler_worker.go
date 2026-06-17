package cmd

import (
	"context"
	"encoding/json"
	"time"

	"github.com/namnv2496/crawler/internal/configs"
	"github.com/namnv2496/crawler/internal/entity"
	"github.com/namnv2496/crawler/internal/pkg/logging"
	"github.com/namnv2496/crawler/internal/repository"
	"github.com/namnv2496/crawler/internal/repository/schedulerservice"
	"github.com/namnv2496/crawler/internal/service"
	"github.com/namnv2496/crawler/internal/service/mq"
	"github.com/namnv2496/crawler/internal/service/server"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var CrawlerWorkerCmd = &cobra.Command{
	Use:   "crawler-worker",
	Short: "A simple web crawler worker",
	Run: func(cmd *cobra.Command, args []string) {
		InvokeCrawlerWorker(
			startCrawlerWorker,
			// startTest,
		)
	},
}

func InvokeCrawlerWorker(invokers ...any) *fx.App {
	config := configs.LoadConfig()
	app := fx.New(
		fx.StartTimeout(time.Second*10),
		fx.StopTimeout(time.Second*10),
		fx.Provide(
			fx.Annotate(mq.NewKafkaConsumer, fx.As(new(mq.IConsumer))),
			// fx.Annotate(mq.NewKafkaProducer, fx.As(new(mq.IProducer))), // for public result event after
			fx.Annotate(service.NewCrawlerService, fx.As(new(service.ICrawlerService))),
			fx.Annotate(service.NewTeleService, fx.As(new(service.ITeleService))),
			fx.Annotate(repository.NewDatabase, fx.As(new(repository.IDatabase))),
			fx.Annotate(repository.NewResultRepository, fx.As(new(repository.IResultRepository))),
			fx.Annotate(server.NewWorkerPool, fx.As(new(server.IWorkerPool))),

			fx.Annotate(mq.NewAsynqProducer, fx.As(new(mq.IAsynqProducer))),
			fx.Annotate(schedulerservice.NewSchedulerService, fx.As(new(schedulerservice.ISchedulerService))),
			fx.Annotate(server.NewDistributedWorkerPool, fx.As(new(server.IDistributedWorkerPool))),
			fx.Annotate(server.NewServerExecutor, fx.As(new(server.IServerExecutor))),
			fx.Annotate(server.NewServerManager, fx.As(new(server.IServerManager))),
		),
		fx.Supply(
			config,
		),
		fx.Invoke(invokers...),
	)
	return app
}

func startCrawlerWorker(
	lc fx.Lifecycle,
	config *configs.Config,
	consumer mq.IConsumer,
	crawlerService service.ICrawlerService,
	workerPool server.IWorkerPool,
) {
	// Start worker pool once at startup
	ctx := context.Background()
	if err := workerPool.Start(ctx); err != nil {
		logging.Error(ctx, "Failed to start worker pool: %v", err)
		return
	}

	// Register graceful shutdown
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			workerPool.Stop()
			return nil
		},
	})

	startConsumer(consumer, crawlerService)
	select {}
}

func startConsumer(
	consumer mq.IConsumer,
	crawlerService service.ICrawlerService,
) {
	for _, consumer := range consumer.GetConsumer() {
		go func(consumer *kafka.Reader) {
			ctx := logging.InjectTraceId(context.Background())
			logging.ResetPrefix(ctx, "startConsumer")
			rateLimiter := time.Tick(time.Second / 10) // 10 requests per second
			defer consumer.Close()
			for {
				select {
				case <-rateLimiter:
					var err error
					var m kafka.Message
					m, err = consumer.ReadMessage(ctx)
					defer func() {
						if err := consumer.CommitMessages(ctx, m); err != nil {
							logging.Error(ctx, "Failed to commit message: %v", err)
						}
					}()
					if err != nil {
						return
					}
					var url entity.CrawlerEvent
					if err := json.Unmarshal(m.Value, &url); err != nil {
						return
					}
					if err := crawlerService.Crawl(ctx, url); err != nil {
						logging.Error(ctx, "%s", err.Error())
						return
					}
					logging.Debug(ctx, "message at topic:%v partition:%v offset:%v\t%s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))

				case <-ctx.Done():
					return
				}
			}
		}(consumer)
	}
}

// func startTest(
// 	crawlerService service.ICrawlerService,
// ) {
// 	ctx := logging.InjectTraceId(context.Background())
// 	logging.ResetPrefix(ctx, "startTest")
// 	url := entity.CrawlerEvent{
// 		Id:       1,
// 		Url:      "https://cellphones.com.vn/robots.txt",
// 		Method:   "ROBOTS",
// 		Queue:    "priority",
// 		Domain:   "phone_cellphones",
// 		IsActive: true,
// 	}
// 	if err := crawlerService.Crawl(ctx, url); err != nil {
// 		logging.Error(ctx, err.Error())
// 		return
// 	}
// }

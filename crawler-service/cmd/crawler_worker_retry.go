package cmd

import (
	"context"
	"time"

	"github.com/namnv2496/crawler/internal/configs"
	"github.com/namnv2496/crawler/internal/repository"
	"github.com/namnv2496/crawler/internal/service"
	"github.com/namnv2496/crawler/internal/service/mq"
	"github.com/namnv2496/crawler/internal/service/server"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var CrawlerWorkerRetryCmd = &cobra.Command{
	Use:   "crawler-worker-retry",
	Short: "A simple web crawler worker",
	Run: func(cmd *cobra.Command, args []string) {
		InvokeCrawlerWorkerRetry(
			startCrawlerWorkerRetry,
		)
	},
}

func InvokeCrawlerWorkerRetry(invokers ...any) *fx.App {
	config := configs.LoadConfig()
	app := fx.New(
		fx.StartTimeout(time.Second*10),
		fx.StopTimeout(time.Second*10),
		fx.Provide(
			fx.Annotate(service.NewCrawlerService, fx.As(new(service.ICrawlerService))),
			fx.Annotate(service.NewTeleService, fx.As(new(service.ITeleService))),
			fx.Annotate(repository.NewDatabase, fx.As(new(repository.IDatabase))),
			fx.Annotate(repository.NewResultRepository, fx.As(new(repository.IResultRepository))),
			fx.Annotate(server.NewWorkerPool, fx.As(new(server.IWorkerPool))),
			fx.Annotate(mq.NewAsynqProducer, fx.As(new(mq.IAsynqProducer))),

			fx.Annotate(mq.NewAsynqConsumer, fx.As(new(mq.IAsynqConsumer))),
			fx.Annotate(service.NewRetryWorker, fx.As(new(service.IRetryWorker))),
		),
		fx.Supply(
			config,
		),
		fx.Invoke(invokers...),
	)
	return app
}

func startCrawlerWorkerRetry(
	lc fx.Lifecycle,
	retryWorker service.IRetryWorker,
	workerPool server.IWorkerPool,
) error {
	// Start worker pool once at startup
	ctx := context.Background()
	if err := workerPool.Start(ctx); err != nil {
		panic("failed to start worker pool: " + err.Error())
	}

	// Register graceful shutdown
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			workerPool.Stop()
			return nil
		},
	})

	return retryWorker.Start(context.Background())
}

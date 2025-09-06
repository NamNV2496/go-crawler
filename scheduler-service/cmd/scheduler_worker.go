package cmd

import (
	"time"

	"github.com/namnv2496/scheduler/internal/configs"
	"github.com/namnv2496/scheduler/internal/repository"
	"github.com/namnv2496/scheduler/internal/repository/distributedlock"
	"github.com/namnv2496/scheduler/internal/service"
	"github.com/namnv2496/scheduler/internal/service/mq"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var schedulerWorkerCmd = &cobra.Command{
	Use:   "scheduler_worker",
	Short: "Start the scheduler worker",
	Run: func(cmd *cobra.Command, args []string) {
		InvokeSchedulerWorker(
			startCronjob,
		)
	},
}

func InvokeSchedulerWorker(invokers ...any) *fx.App {
	config := configs.LoadConfig()
	app := fx.New(
		fx.StartTimeout(time.Second*10),
		fx.StopTimeout(time.Second*10),
		fx.Provide(
			fx.Annotate(repository.NewDatabase, fx.As(new(repository.IDatabase))),
			fx.Annotate(repository.NewCrawlerEventRepository, fx.As(new(repository.ICrawlerEventRepository))),
			// MQ
			fx.Annotate(mq.NewKafkaProducer, fx.As(new(mq.IProducer))),
			fx.Annotate(service.NewUrlCronJob, fx.As(new(service.ICrawlerCronJob))),
			// rate limit
			fx.Annotate(distributedlock.NewDistributedLock, fx.As(new(distributedlock.IDistributedLock))),
		),
		fx.Supply(
			config,
		),
		fx.Invoke(invokers...),
	)
	return app
}

func startCronjob(
	urlCronJob service.ICrawlerCronJob,
) error {
	// start cron job
	if err := urlCronJob.Start(); err != nil {
		panic("failed to start publisher")
	}
	select {}
}

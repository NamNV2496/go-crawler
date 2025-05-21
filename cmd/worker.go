package cmd

import (
	"time"

	"github.com/namnv2496/crawler/internal/configs"
	"github.com/namnv2496/crawler/internal/repository"
	"github.com/namnv2496/crawler/internal/service"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "A simple web crawler worker",
	Run: func(cmd *cobra.Command, args []string) {
		startWorker()
	},
}

func startWorker() *fx.App {
	config := configs.LoadConfig()
	app := fx.New(
		fx.StartTimeout(time.Second*10),
		fx.StopTimeout(time.Second*10),
		fx.Provide(
			fx.Annotate(repository.NewDatabase, fx.As(new(repository.IRepository))),
			fx.Annotate(service.NewQueueService, fx.As(new(service.IQueueService))),
			fx.Annotate(repository.NewQueueRepository, fx.As(new(repository.IQueueRepository))),
			fx.Annotate(repository.NewUrlRepository, fx.As(new(repository.IUrlRepository))),
			fx.Annotate(service.NewUrlWorker, fx.As(new(service.IUrlWorker))),
		),
		fx.Supply(
			config,
		),
	)
	return app
}

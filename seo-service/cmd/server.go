package cmd

import (
	"time"

	"github.com/namnv2496/seo/configs"
	"github.com/namnv2496/seo/internal/controller"
	"github.com/namnv2496/seo/internal/repository"
	"github.com/namnv2496/seo/internal/service"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the server",
	Run: func(cmd *cobra.Command, args []string) {
		InvokeServer(startServer)
	},
}

func InvokeServer(invokers ...any) *fx.App {
	config := configs.LoadConfig()
	app := fx.New(
		fx.StartTimeout(time.Second*10),
		fx.StopTimeout(time.Second*10),
		fx.Provide(
			// repository
			fx.Annotate(repository.NewDatabase, fx.As(new(repository.IDatabase))),
			// fx.Annotate(repository.NewUrlRepository, fx.As(new(repository.IUrlRepository))),
			fx.Annotate(repository.NewUrlRepo, fx.As(new(repository.IUrlRepo))),
			fx.Annotate(repository.NewUrlMetadataRepo, fx.As(new(repository.IUrlMetadataRepo))),
			fx.Annotate(repository.NewShortLinkRepo, fx.As(new(repository.IShortLinkRepo))),
			// service
			fx.Annotate(service.NewUrlService, fx.As(new(service.IUrlService))),
			// controller
			fx.Annotate(controller.NewUrlController, fx.As(new(controller.IController))),
		),
		fx.Supply(
			config,
		),
		fx.Invoke(invokers...),
	)
	return app
}

func startServer(
	lc fx.Lifecycle,
	urlController controller.IController,
) {
	controller.Start(urlController)
}

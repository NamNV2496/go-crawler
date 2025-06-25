package cmd

import (
	"log"
	"time"

	"github.com/namnv2496/seo/configs"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var DynamicKeywordCmd = &cobra.Command{
	Use:   "dynamic_keyword_worker",
	Short: "A simple web dynamic keyword worker",
	Run: func(cmd *cobra.Command, args []string) {
		InvokeDynamicKeyword(startDynamicKeyword)
	},
}

func InvokeDynamicKeyword(invokers ...any) *fx.App {
	config := configs.LoadConfig()
	app := fx.New(
		fx.StartTimeout(time.Second*10),
		fx.StopTimeout(time.Second*10),
		fx.Provide(),
		fx.Supply(
			config,
		),
		fx.Invoke(invokers...),
	)
	return app
}

func startDynamicKeyword(
	lc fx.Lifecycle,
	config *configs.Config,
) {
	log.Println("Start consumer to update keyword by cron job. It will develop soon")
	select {}
}

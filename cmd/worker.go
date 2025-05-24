package cmd

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/namnv2496/crawler/internal/configs"
	"github.com/namnv2496/crawler/internal/entity"
	"github.com/namnv2496/crawler/internal/repository"
	"github.com/namnv2496/crawler/internal/service"
	"github.com/namnv2496/crawler/internal/service/mq"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

const (
	MaxWorker = 10
)

var CrawlerWorkerCmd = &cobra.Command{
	Use:   "crawler-worker",
	Short: "A simple web crawler worker",
	Run: func(cmd *cobra.Command, args []string) {
		InvokeCrawlerWorker(startCrawlerWorker)
	},
}

func InvokeCrawlerWorker(invokers ...any) *fx.App {
	config := configs.LoadConfig()
	app := fx.New(
		fx.StartTimeout(time.Second*10),
		fx.StopTimeout(time.Second*10),
		fx.Provide(
			fx.Annotate(mq.NewKafkaConsumer, fx.As(new(mq.IConsumer))),
			fx.Annotate(service.NewCrawlerService, fx.As(new(service.ICrawlerService))),
			fx.Annotate(service.NewTeleService, fx.As(new(service.ITeleService))),
			fx.Annotate(repository.NewDatabase, fx.As(new(repository.IRepository))),
			fx.Annotate(repository.NewResultRepository, fx.As(new(repository.IResultRepository))),
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
) {
	log.Println("Start consumer")
	startConsumer(consumer, crawlerService)
	select {}
}
func startConsumer(
	consumer mq.IConsumer,
	crawlerService service.ICrawlerService,
) {
	for _, consumer := range consumer.GetConsumer() {
		go func(consumer *kafka.Reader) {
			ctx := context.Background()
			rateLimiter := time.Tick(time.Second / 10) // 10 requests per second
			defer consumer.Close()
			for {
				select {
				case <-rateLimiter:
					var err error
					var m kafka.Message
					m, err = consumer.ReadMessage(ctx)
					defer consumer.CommitMessages(ctx, m)
					if err != nil {
						return
					}
					var url entity.Url
					if err := json.Unmarshal(m.Value, &url); err != nil {
						return
					}
					if err := crawlerService.Crawl(ctx, url); err != nil {
						log.Println(err)
						return
					}
					log.Printf("message at topic:%v partition:%v offset:%v\t%s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))

				case <-ctx.Done():
					return
				}
			}
		}(consumer)
	}
}

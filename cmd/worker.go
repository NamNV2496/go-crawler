package cmd

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/namnv2496/crawler/internal/configs"
	"github.com/namnv2496/crawler/internal/entity"
	"github.com/namnv2496/crawler/internal/service"
	"github.com/namnv2496/crawler/internal/service/mq"
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
			fx.Annotate(service.NewCrawler, fx.As(new(service.ICrawlerService))),
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
	log.Println("Start worker")
	for range MaxWorker {
		startConsumer(consumer, crawlerService)
	}
}
func startConsumer(
	consumer mq.IConsumer,
	crawlerService service.ICrawlerService,
) {
	// crawlerService.Crawl(ctx,
	// 	`curl --location 'https://m.cafef.vn/du-lieu/Ajax/ajaxgoldprice.ashx?index=11' --header 'Accept: */*' --header 'Accept-Language: en-US,en;q=0.9,vi;q=0.8' --header 'Connection: keep-alive' --header 'Referer: https://m.cafef.vn/du-lieu/gia-vang-hom-nay/trong-nuoc.chn' --header 'Sec-Fetch-Dest: empty' --header 'Sec-Fetch-Mode: cors' --header 'Sec-Fetch-Site: same-origin' --header 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 Edg/135.0.0.0' --header 'sec-ch-ua: "Microsoft Edge";v="135", "Not-A.Brand";v="8", "Chromium";v="135"' --header 'sec-ch-ua-mobile: ?0' --header 'sec-ch-ua-platform: "macOS"' --header 'Cookie: _ga=GA1.2.1174992577.1733489327; _ga_860L8F5EZP=GS1.1.1740282133.10.0.1740282328.0.0.0; ASP.NET_SessionId=wnors2tpgmcb0lwvqwebtsf5; favorite_stocks_state=1'`,
	// 	"CURL")
	for {
		for _, consumer := range consumer.GetConsumer() {
			go func() {
				ctx := context.Background()
				defer consumer.Close()
				m, err := consumer.ReadMessage(ctx)
				if err != nil {
					return
				}
				var url entity.Url
				if err := json.Unmarshal(m.Value, &url); err != nil {
					return
				}
				crawlerService.Crawl(ctx, url.Url, url.Method)
				log.Printf("message at topic:%v partition:%v offset:%v	%s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
			}()
		}
	}
}

package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/validator"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/namnv2496/crawler/internal/configs"
	"github.com/namnv2496/crawler/internal/controller"
	"github.com/namnv2496/crawler/internal/repository"
	"github.com/namnv2496/crawler/internal/service"
	"github.com/namnv2496/crawler/internal/service/mq"
	crawlerv1 "github.com/namnv2496/crawler/pkg/generated/pkg/proto"
	"github.com/namnv2496/crawler/pkg/utils"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the server",
	Run: func(cmd *cobra.Command, args []string) {
		InvokeServer(
			startCronjob,
			startServer,
		)
	},
}

func InvokeServer(invokers ...any) *fx.App {
	config := configs.LoadConfig()
	app := fx.New(
		fx.StartTimeout(time.Second*10),
		fx.StopTimeout(time.Second*10),
		fx.Provide(
			fx.Annotate(repository.NewDatabase, fx.As(new(repository.IDatabase))),
			// url
			fx.Annotate(repository.NewUrlRepository, fx.As(new(repository.IUrlRepository))),
			fx.Annotate(service.NewUrlService, fx.As(new(service.IUrlService))),
			fx.Annotate(controller.NewUrlController, fx.As(new(crawlerv1.UrlServiceServer))),
			// queue
			fx.Annotate(controller.NewQueueController, fx.As(new(crawlerv1.QueueServiceServer))),
			fx.Annotate(service.NewQueueService, fx.As(new(service.IQueueService))),
			fx.Annotate(repository.NewQueueRepository, fx.As(new(repository.IQueueRepository))),
			// MQ
			fx.Annotate(mq.NewKafkaProducer, fx.As(new(mq.IProducer))),
			fx.Annotate(service.NewUrlCronJob, fx.As(new(service.IUrlCronJob))),
			// rate limit
			fx.Annotate(startRateLimit, fx.As(new(utils.IRateLimit))),
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
	config *configs.Config,
	urlController crawlerv1.UrlServiceServer,
	queueController crawlerv1.QueueServiceServer,
	urlCronJob service.IUrlCronJob,
) error {
	// start grpc
	listener, err := net.Listen("tcp", config.AppConfig.GRPCPort)
	if err != nil {
		return err
	}
	defer listener.Close()
	var opts = []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			validator.UnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			validator.StreamServerInterceptor(),
		),
	}
	server := grpc.NewServer(opts...)
	reflection.Register(server)
	crawlerv1.RegisterUrlServiceServer(server, urlController)
	crawlerv1.RegisterQueueServiceServer(server, queueController)
	fmt.Printf("gRPC server is running on %s\n", config.AppConfig.GRPCPort)
	// start http
	conn, err := grpc.DialContext(context.Background(), config.AppConfig.GRPCPort, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to dial gRPC server: %v", err)
	}
	defer conn.Close()
	mux := runtime.NewServeMux()
	if err := crawlerv1.RegisterUrlServiceHandler(context.Background(), mux, conn); err != nil {
		return fmt.Errorf("failed to register handler: %v", err)
	}
	if err := crawlerv1.RegisterQueueServiceHandler(context.Background(), mux, conn); err != nil {
		return fmt.Errorf("failed to register handler: %v", err)
	}
	go func() {
		fmt.Printf("HTTP server is running on %s\n", config.AppConfig.HTTPPort)
		if err := http.ListenAndServe(config.AppConfig.HTTPPort, mux); err != nil {
			log.Fatalf("failed to start HTTP server: %v", err)
		}
	}()
	return server.Serve(listener)
}

func startCronjob(
	urlCronJob service.IUrlCronJob,
) error {
	// start cron job
	if err := urlCronJob.Start(); err != nil {
		panic("failed to start publisher")
	}
	return nil
}

func startRateLimit(
	conf *configs.Config,
) *utils.RateLimit {
	rateLimit := utils.NewRateLimitWithOption(conf, &utils.RatelimitOpt{
		BlockRetention: time.Minute * 2,
		CalculateBlockDuration: func(count int) time.Duration {
			return time.Minute * 2 * time.Duration(count)
		},
	})

	fmt.Printf("Rate limit is started")
	return rateLimit
}

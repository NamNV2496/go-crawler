package cmd

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/validator"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/namnv2496/crawler/internal/configs"
	"github.com/namnv2496/crawler/internal/controller"
	"github.com/namnv2496/crawler/internal/repository"
	"github.com/namnv2496/crawler/internal/service"
	"github.com/namnv2496/crawler/internal/service/mq"
	crawlerv1 "github.com/namnv2496/crawler/pkg/generated/pkg/proto"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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
			fx.Annotate(repository.NewDatabase, fx.As(new(repository.IRepository))),
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
			fx.Annotate(service.NewUrlWorker, fx.As(new(service.IUrlWorker))),
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
	urlWorker service.IUrlWorker,
) error {
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
	conn, err := grpc.DialContext(context.Background(), config.AppConfig.GRPCPort, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to dial gRPC server: %v", err)
	}
	defer conn.Close()

	mux := runtime.NewServeMux()
	if err := crawlerv1.RegisterUrlServiceHandlerFromEndpoint(context.Background(), mux, config.AppConfig.GRPCPort, []grpc.DialOption{grpc.WithInsecure()}); err != nil {
		return fmt.Errorf("failed to register handler: %v", err)
	}
	if err := crawlerv1.RegisterQueueServiceHandlerFromEndpoint(context.Background(), mux, config.AppConfig.GRPCPort, []grpc.DialOption{grpc.WithInsecure()}); err != nil {
		return fmt.Errorf("failed to register handler: %v", err)
	}
	fmt.Printf("http server is running on %s\n", config.AppConfig.HTTPPort)
	//
	go func() {
		if err := urlWorker.Start(); err != nil {
			fmt.Println(err)
		}
	}()

	return server.Serve(listener)
}

package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/validator"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/namnv2496/scheduler/internal/configs"
	"github.com/namnv2496/scheduler/internal/controller"
	"github.com/namnv2496/scheduler/internal/repository"
	"github.com/namnv2496/scheduler/internal/service"
	internalvalidator "github.com/namnv2496/scheduler/internal/validator"
	crawlerv1 "github.com/namnv2496/scheduler/pkg/generated/pkg/proto"
	"github.com/namnv2496/scheduler/pkg/utils"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

var serverCmd = &cobra.Command{
	Use:   "scheduler",
	Short: "Start the scheduler service",
	Run: func(cmd *cobra.Command, args []string) {
		InvokeServer(
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
			// crawler event
			fx.Annotate(repository.NewSchedulerEventRepository, fx.As(new(repository.ISchedulerEventRepository))),
			fx.Annotate(service.NewSchedulerEventService, fx.As(new(service.ISchedulerEventService))),
			fx.Annotate(controller.NewSchedulerEventController, fx.As(new(crawlerv1.SchedulerEventServiceServer))),

			fx.Annotate(startRateLimit, fx.As(new(utils.IRateLimit))),
			fx.Annotate(internalvalidator.NewValidate, fx.As(new(internalvalidator.IValidate))),
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
	urlController crawlerv1.SchedulerEventServiceServer,
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
	crawlerv1.RegisterSchedulerEventServiceServer(server, urlController)
	fmt.Printf("gRPC server is running on %s\n", config.AppConfig.GRPCPort)
	// start http
	conn, err := grpc.NewClient(config.AppConfig.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to dial gRPC server: %v", err)
	}
	defer conn.Close()
	mux := runtime.NewServeMux()
	if err := crawlerv1.RegisterSchedulerEventServiceHandler(context.Background(), mux, conn); err != nil {
		return fmt.Errorf("failed to register handler: %v", err)
	}
	go func() {
		fmt.Printf("HTTP server is running on %s\n", config.AppConfig.HTTPPort)
		if err := http.ListenAndServe(config.AppConfig.HTTPPort, mux); err != nil {
			panic("failed to start HTTP server")
		}
	}()
	return server.Serve(listener)
}

func startRateLimit(
	conf *configs.Config,
) *utils.RateLimit {
	rateLimit := utils.NewRateLimitWithOption(conf, &utils.RatelimitOpt{
		BlockRetention: time.Second * 10,
		CalculateBlockDuration: func(count int) time.Duration {
			return time.Second * 10 * time.Duration(count)
		},
	})

	fmt.Printf("Rate limit is started")
	return rateLimit
}

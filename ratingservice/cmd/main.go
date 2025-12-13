package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/ochamekan/ms/gen"
	"github.com/ochamekan/ms/pkg/consul"
	"github.com/ochamekan/ms/pkg/discovery"
	"github.com/ochamekan/ms/pkg/logging"
	"github.com/ochamekan/ms/ratingservice/internal/controller/rating"
	grpchandler "github.com/ochamekan/ms/ratingservice/internal/handler/grpc"
	"github.com/ochamekan/ms/ratingservice/internal/repository/cache"
	"github.com/ochamekan/ms/ratingservice/internal/repository/postgres"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	serviceName = "rating"
	port        = 8082
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	logger = logger.With(zap.String(logging.FieldService, serviceName))
	logger.Info("Starting rating service")

	err := godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading .env file", zap.Error(err))
	}

	registry, err := consul.NewRegistry("discovery:8500")
	if err != nil {
		logger.Fatal("Failed to create consul registry", zap.Error(err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	instanceID := discovery.GenerateInstanceID(serviceName)

	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("rating:%d", port)); err != nil {
		logger.Fatal("Failed to register instance", zap.Error(err))
	}

	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				logger.Error("Failed to report healthy state", zap.Error(err))
			}
			time.Sleep(1 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceID, serviceName)

	repo, closer, err := postgres.New()
	if err != nil {
		logger.Fatal("Failed to initialize postgresql database", zap.Error(err))
	}
	defer closer()

	cache, err := cache.New(serviceName)
	if err != nil {
		logger.Fatal("Failed to initialize redis database", zap.Error(err))
	}

	ctrl := rating.New(repo, cache, logger)

	h := grpchandler.New(ctrl, logger)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	srv := grpc.NewServer()
	reflection.Register(srv)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	var wg sync.WaitGroup

	wg.Go(func() {
		s := <-sigChan
		logger.Info("Received signal, attempting graceful shutdown", zap.Stringer("signal", s))
		cancel()
		srv.GracefulStop()
		logger.Info("Gracefully stopped the gRPC server for rating service")
	})

	gen.RegisterRatingServiceServer(srv, h)
	if err := srv.Serve(lis); err != nil {
		logger.Fatal("Failed to serve", zap.Error(err))
	}

	wg.Wait()
}

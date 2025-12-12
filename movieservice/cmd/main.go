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

	"github.com/grpc-ecosystem/go-grpc-middleware/ratelimit"
	"github.com/ochamekan/ms/gen"
	"github.com/ochamekan/ms/movieservice/internal/controller/movie"
	metadatagateway "github.com/ochamekan/ms/movieservice/internal/gateway/metadata/grpc"
	ratinggateway "github.com/ochamekan/ms/movieservice/internal/gateway/rating/grpc"
	grpchandler "github.com/ochamekan/ms/movieservice/internal/handler/grpc"
	"github.com/ochamekan/ms/pkg/consul"
	"github.com/ochamekan/ms/pkg/discovery"
	"github.com/ochamekan/ms/pkg/logging"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	serviceName = "movie"
	port        = 8083
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	logger = logger.With(zap.String(logging.FieldService, serviceName))
	logger.Info("Starting movie service")

	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		logger.Fatal("Failed to create consul registry", zap.Error(err))
	}

	instanceID := discovery.GenerateInstanceID(serviceName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("localhost:%d", port)); err != nil {
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

	metadataGateway := metadatagateway.New(registry, logger)
	ratingGateway := ratinggateway.New(registry)
	ctrl := movie.New(ratingGateway, metadataGateway)

	h := grpchandler.New(ctrl, logger)

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	const limit = 100 // max 100 request per second
	const burst = 100 // max parallel requests
	lim := newLimiter(limit, burst)

	srv := grpc.NewServer(grpc.UnaryInterceptor(ratelimit.UnaryServerInterceptor(lim)))
	reflection.Register(srv)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	var wg sync.WaitGroup

	wg.Go(func() {
		s := <-sigChan
		logger.Info("Received signal, attempting graceful shutdown", zap.Stringer("signal", s))
		cancel()
		srv.GracefulStop()
		logger.Info("Gracefully stopped the gRPC server for movie service")
	})

	gen.RegisterMovieServiceServer(srv, h)
	if err := srv.Serve(lis); err != nil {
		logger.Fatal("Failed to serve", zap.Error(err))
	}

	wg.Wait()
}

// limiter implements *ratelimit.Limiter
type limiter struct {
	l *rate.Limiter
}

func newLimiter(limit int, burst int) *limiter {
	return &limiter{rate.NewLimiter(rate.Limit(limit), burst)}
}

func (l *limiter) Limit() bool {
	return l.l.Allow()
}

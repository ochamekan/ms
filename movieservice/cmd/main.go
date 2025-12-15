package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/ratelimit"
	"github.com/ochamekan/ms/gen"
	"github.com/ochamekan/ms/movieservice/internal/controller/movie"
	metadatagateway "github.com/ochamekan/ms/movieservice/internal/gateway/metadata/grpc"
	ratinggateway "github.com/ochamekan/ms/movieservice/internal/gateway/rating/grpc"
	grpchandler "github.com/ochamekan/ms/movieservice/internal/handler/grpc"
	"github.com/ochamekan/ms/pkg/consul"
	"github.com/ochamekan/ms/pkg/discovery"
	"github.com/ochamekan/ms/pkg/logging"
	"github.com/ochamekan/ms/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	serviceName = "movie"
	port        = 8083
	metricsPort = 9100
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	logger = logger.With(zap.String(logging.FieldService, serviceName))
	logger.Info("Starting movie service")

	/////////////////////////////
	// PROMETHEUS
	srvMetrics := grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
	)

	reg := prometheus.NewRegistry()
	reg.MustRegister(srvMetrics)

	metrics := metrics.New(reg)

	go func() {
		http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
		http.ListenAndServe(fmt.Sprintf(":%d", metricsPort), nil)
	}()
	/////////////////////////////

	registry, err := consul.NewRegistry("discovery:8500")
	if err != nil {
		logger.Fatal("Failed to create consul registry", zap.Error(err))
	}

	instanceID := discovery.GenerateInstanceID(serviceName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("movie:%d", port)); err != nil {
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

	h := grpchandler.New(ctrl, logger, metrics)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	const limit = 100 // max 100 request per second
	const burst = 100 // max parallel requests
	lim := newLimiter(limit, burst)

	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(
		srvMetrics.UnaryServerInterceptor(),
		ratelimit.UnaryServerInterceptor(lim),
	))

	reflection.Register(srv)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	srvMetrics.InitializeMetrics(srv)

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
	return !l.l.Allow()
}

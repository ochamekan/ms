package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/ochamekan/ms/gen"
	"github.com/ochamekan/ms/metadataservice/internal/controller/metadata"
	grpchandler "github.com/ochamekan/ms/metadataservice/internal/handler/grpc"
	"github.com/ochamekan/ms/metadataservice/internal/repository/cache"
	"github.com/ochamekan/ms/metadataservice/internal/repository/postgres"
	"github.com/ochamekan/ms/pkg/consul"
	"github.com/ochamekan/ms/pkg/discovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	serviceName = "metadata"
	port        = 8081
)

func main() {
	log.Println("Starting movie metadata service...")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		panic(err)
	}

	instanceID := discovery.GenerateInstanceID(serviceName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("localhost:%d", port)); err != nil {
		panic(err)
	}

	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				log.Println("Failed to report healthy state: " + err.Error())
			}
			time.Sleep(1 * time.Second)
		}

	}()
	defer registry.Deregister(ctx, instanceID, serviceName)

	repo, closer, err := postgres.New()
	if err != nil {
		log.Fatalf("Failed to initialize postgresql database: %v", err)
	}
	defer closer()

	cache, err := cache.New(serviceName)
	if err != nil {
		log.Fatalf("Failed to initialize redis database: %v", err)
	}

	ctrl := metadata.New(repo, cache)
	h := grpchandler.New(ctrl)

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	srv := grpc.NewServer()
	reflection.Register(srv)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	var wg sync.WaitGroup

	wg.Go(func() {
		s := <-sigChan
		log.Printf("Received signal %v, attempting graceful shutdown", s)
		cancel()
		srv.GracefulStop()
		log.Println("Gracefully stopped the gRPC server")
	})

	gen.RegisterMetadataServiceServer(srv, h)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}

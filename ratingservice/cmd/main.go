package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/joho/godotenv"
	"github.com/ochamekan/ms/gen"
	"github.com/ochamekan/ms/pkg/consul"
	"github.com/ochamekan/ms/pkg/discovery"
	"github.com/ochamekan/ms/ratingservice/internal/controller/rating"
	grpchandler "github.com/ochamekan/ms/ratingservice/internal/handler/grpc"
	"github.com/ochamekan/ms/ratingservice/internal/repository/cache"
	"github.com/ochamekan/ms/ratingservice/internal/repository/postgres"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	serviceName = "rating"
	port        = 8082
)

func main() {
	log.Println("Starting rating service...")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)

	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("metadata:%d", port)); err != nil {
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

	ctrl := rating.New(repo, cache)

	h := grpchandler.New(ctrl)

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	srv := grpc.NewServer()
	reflection.Register(srv)

	gen.RegisterRatingServiceServer(srv, h)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}

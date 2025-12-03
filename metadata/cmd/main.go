package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/joho/godotenv"
	"github.com/ochamekan/ms/gen"
	"github.com/ochamekan/ms/metadata/internal/controller/metadata"
	grpchandler "github.com/ochamekan/ms/metadata/internal/handler/grpc"
	"github.com/ochamekan/ms/metadata/internal/repository/cache"
	"github.com/ochamekan/ms/metadata/internal/repository/postgres"
	"github.com/ochamekan/ms/pkg/consul"
	"github.com/ochamekan/ms/pkg/discovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const serviceName = "metadata"

func main() {
	var port int
	flag.IntVar(&port, "port", 8081, "API handler port")
	flag.Parse()

	log.Printf("Starting the movie metadata service on port %d", port)

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
		log.Fatalf("failed to initialize postgres database: %v", err)
	}
	defer closer()

	cache, err := cache.New("metadata")
	if err != nil {
		log.Fatalf("failed to initialize redis database: %v", err)
	}

	ctrl := metadata.New(repo, cache)
	h := grpchandler.New(ctrl)

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	srv := grpc.NewServer()
	reflection.Register(srv)

	gen.RegisterMetadataServiceServer(srv, h)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}

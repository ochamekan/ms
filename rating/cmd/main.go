package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/ochamekan/ms/gen"
	"github.com/ochamekan/ms/pkg/consul"
	"github.com/ochamekan/ms/pkg/discovery"
	"github.com/ochamekan/ms/rating/internal/controller/rating"
	grpchandler "github.com/ochamekan/ms/rating/internal/handler/grpc"
	"github.com/ochamekan/ms/rating/internal/repository/cache"
	"github.com/ochamekan/ms/rating/internal/repository/postgres"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
)

const serviceName = "rating"

func main() {
	log.Println("Starting the rating service...")
	f, err := os.Open("default.yaml")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var cfg config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		panic(err)
	}

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := cfg.API.Port

	registry, err := consul.NewRegistry(cfg.ServiceDiscovery.Consul.Address)
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
		log.Fatalf("fialed to initialize postgres database: %v", err)
	}
	defer closer()

	// ingester, err := kafka.NewIngester("localhost", "rating", "ratings")
	// if err != nil {
	// 	log.Fatalf("failed to initialize ingester: %v", err)
	// }

	cache, err := cache.New("rating")
	if err != nil {
		log.Fatalf("failed to initialize redis database: %v", err)
	}

	ctrl := rating.New(repo, cache)

	// if err := ctrl.StartIngestion(ctx); err != nil {
	// 	log.Fatalf("failed to start ingestion: %v", err)
	// }

	h := grpchandler.New(ctrl)

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	srv := grpc.NewServer()
	reflection.Register(srv)

	gen.RegisterRatingServiceServer(srv, h)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}

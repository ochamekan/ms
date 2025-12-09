package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/ochamekan/ms/gen"
	"github.com/ochamekan/ms/movieservice/internal/controller/movie"
	metadatagateway "github.com/ochamekan/ms/movieservice/internal/gateway/metadata/grpc"
	ratinggateway "github.com/ochamekan/ms/movieservice/internal/gateway/rating/grpc"
	grpchandler "github.com/ochamekan/ms/movieservice/internal/handler/grpc"
	"github.com/ochamekan/ms/pkg/consul"
	"github.com/ochamekan/ms/pkg/discovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	serviceName = "movie"
	port        = 8083
)

func main() {
	log.Println("Starting movie service...")

	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		panic(err)
	}

	instanceID := discovery.GenerateInstanceID(serviceName)

	ctx := context.Background()
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

	metadataGateway := metadatagateway.New(registry)
	ratingGateway := ratinggateway.New(registry)
	ctrl := movie.New(ratingGateway, metadataGateway)

	h := grpchandler.New(ctrl)

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	srv := grpc.NewServer()
	reflection.Register(srv)

	gen.RegisterMovieServiceServer(srv, h)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}

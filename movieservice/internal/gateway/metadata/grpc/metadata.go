package grpc

import (
	"context"
	"fmt"

	"github.com/ochamekan/ms/gen"
	"github.com/ochamekan/ms/internal/grpcutil"
	"github.com/ochamekan/ms/metadataservice/pkg/model"
	"github.com/ochamekan/ms/pkg/discovery"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Gateway struct {
	registry discovery.Registry
}

func New(registry discovery.Registry) *Gateway {
	return &Gateway{registry}
}

func (g *Gateway) GetMetadata(ctx context.Context, id int) (*model.Metadata, error) {
	conn, err := grpcutil.ServiceConnection(ctx, "metadata", g.registry)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := gen.NewMetadataServiceClient(conn)
	var resp *gen.GetMetadataResponse

	// If error is retriable, try 5 times and if no success return err
	const maxRetries = 5
	for range 5 {
		resp, err = client.GetMetadata(ctx, &gen.GetMetadataRequest{Id: int32(id)})
		if err != nil {
			if shouldRetry(err) {
				continue
			}
			return nil, err
		}
		return model.MetadataFromProto(resp.Metadata), nil
	}

	return nil, fmt.Errorf("GetMetadata failed after %d retries: %s", maxRetries, err.Error())
}

func (g *Gateway) PutMetadata(ctx context.Context, title, description, director string, year int) error {
	conn, err := grpcutil.ServiceConnection(ctx, "metadata", g.registry)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := gen.NewMetadataServiceClient(conn)
	_, err = client.PutMetadata(ctx, &gen.PutMetadataRequest{Title: title, Description: description, Year: int32(year), Director: director})
	if err != nil {
		return err
	}

	return nil
}

func shouldRetry(err error) bool {
	e, ok := status.FromError(err)
	if !ok {
		return false
	}
	return e.Code() == codes.DeadlineExceeded || e.Code() == codes.ResourceExhausted || e.Code() == codes.Unavailable
}

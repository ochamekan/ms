package grpc

import (
	"context"

	"github.com/ochamekan/ms/gen"
	"github.com/ochamekan/ms/internal/grpcutil"
	"github.com/ochamekan/ms/metadata/pkg/model"
	"github.com/ochamekan/ms/pkg/discovery"
)

type Gateway struct {
	registry discovery.Registry
}

func New(registry discovery.Registry) *Gateway {
	return &Gateway{registry}
}

func (g *Gateway) GetMetadata(ctx context.Context, id string) (*model.Metadata, error) {
	conn, err := grpcutil.ServiceConnection(ctx, "metadata", g.registry)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := gen.NewMetadataServiceClient(conn)
	resp, err := client.GetMetadata(ctx, &gen.GetMetadataRequest{MovieId: id})
	if err != nil {
		return nil, err
	}

	return model.MetadataFromProto(resp.Metadata), nil
}

// func (g *Gateway) PutMetadata(ctx context.Context, metadata *model.Metadata) error

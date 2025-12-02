package grpc

import (
	"context"

	"github.com/ochamekan/ms/gen"
	"github.com/ochamekan/ms/internal/grpcutil"
	"github.com/ochamekan/ms/pkg/discovery"
	"github.com/ochamekan/ms/rating/pkg/model"
)

type Gateway struct {
	registry discovery.Registry
}

func New(registry discovery.Registry) *Gateway {
	return &Gateway{registry}
}

func (g *Gateway) GetAggregatedRating(ctx context.Context, recordID model.RecordID, recordType model.RecordType) (float64, error) {
	conn, err := grpcutil.ServiceConnection(ctx, "rating", g.registry)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	client := gen.NewRatingServiceClient(conn)

	v, err := client.GetAggregatedRating(ctx, &gen.GetAggregatedRatingRequest{RecordId: string(recordID), RecordType: string(recordType)})
	if err != nil {
		return 0, err
	}

	return v.RatingValue, nil
}

func (g *Gateway) PutRating(ctx context.Context, recordID model.RecordID, recordType model.RecordType, rating *model.Rating) error {
	conn, err := grpcutil.ServiceConnection(ctx, "rating", g.registry)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := gen.NewRatingServiceClient(conn)

	_, err = client.PutRating(ctx, &gen.PutRatingRequest{UserId: string(rating.UserID), RecordId: string(rating.RecordID), RecordType: string(rating.RecordType), RatingValue: int32(rating.Value)})

	return err
}

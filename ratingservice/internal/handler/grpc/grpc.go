package grpc

import (
	"context"
	"errors"

	"github.com/ochamekan/ms/gen"
	"github.com/ochamekan/ms/ratingservice/internal/controller/rating"
	"github.com/ochamekan/ms/ratingservice/pkg/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	gen.UnimplementedRatingServiceServer
	ctrl *rating.Controller
}

func New(ctrl *rating.Controller) *Handler {
	return &Handler{ctrl: ctrl}
}

func (h *Handler) GetAggregatedRating(ctx context.Context, req *gen.GetAggregatedRatingRequest) (*gen.GetAggregatedRatingResponse, error) {
	if req == nil || req.MovieId <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "nil req or incorrect movie id")
	}

	v, err := h.ctrl.GetAggregatedRating(ctx, model.MovieID(req.MovieId))
	if err != nil && errors.Is(err, rating.ErrNotFound) {
		return nil, status.Error(codes.NotFound, err.Error())
	} else if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &gen.GetAggregatedRatingResponse{Rating: v}, nil
}

func (h *Handler) PutRating(ctx context.Context, req *gen.PutRatingRequest) (*gen.PutRatingResponse, error) {
	if req == nil || req.MovieId <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "nil req or incorrect movie id")
	}

	if err := h.ctrl.PutRating(ctx, model.MovieID(req.MovieId), model.RatingValue(req.Rating)); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &gen.PutRatingResponse{}, nil
}

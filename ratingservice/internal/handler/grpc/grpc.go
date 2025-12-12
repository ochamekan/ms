package grpc

import (
	"context"
	"errors"

	"github.com/ochamekan/ms/gen"
	"github.com/ochamekan/ms/pkg/logging"
	"github.com/ochamekan/ms/ratingservice/internal/controller/rating"
	"github.com/ochamekan/ms/ratingservice/pkg/model"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	gen.UnimplementedRatingServiceServer
	ctrl   *rating.Controller
	logger *zap.Logger
}

func New(ctrl *rating.Controller, logger *zap.Logger) *Handler {
	return &Handler{ctrl: ctrl, logger: logger.With(zap.String(logging.FieldComponent, "rating handler"))}
}

func (h *Handler) GetAggregatedRating(ctx context.Context, req *gen.GetAggregatedRatingRequest) (*gen.GetAggregatedRatingResponse, error) {
	logger := h.logger.With(zap.String(logging.FieldEndpoint, "GetAggregatedRating"))
	if req == nil || req.MovieId < 0 {
		logger.Warn("nil request or incorrect movie id")
		return nil, status.Errorf(codes.InvalidArgument, "nil req or incorrect movie id")
	}

	logger.Info("Getting all ratings")
	v, err := h.ctrl.GetAggregatedRating(ctx, model.MovieID(req.MovieId))
	if err != nil && errors.Is(err, rating.ErrNotFound) {
		logger.Error("Failed to get ratings", zap.Error(err))
		return nil, status.Error(codes.NotFound, err.Error())
	} else if err != nil {
		logger.Error("Faield to get ratings", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	logger.Info("Ratings successfully retrieved")
	return &gen.GetAggregatedRatingResponse{Rating: v}, nil
}

func (h *Handler) PutRating(ctx context.Context, req *gen.PutRatingRequest) (*gen.PutRatingResponse, error) {
	logger := h.logger.With(zap.String(logging.FieldEndpoint, "PutRating"))
	if req == nil || req.MovieId <= 0 {
		logger.Warn("nil request or incorrect movie id")
		return nil, status.Errorf(codes.InvalidArgument, "nil req or incorrect movie id")
	}

	logger.Info("Adding rating")
	if err := h.ctrl.PutRating(ctx, model.MovieID(req.MovieId), model.RatingValue(req.Rating)); err != nil {
		logger.Error("Failed to add rating", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	logger.Info("Rating successfully added")
	return &gen.PutRatingResponse{}, nil
}

package grpc

import (
	"context"
	"errors"

	"github.com/ochamekan/ms/gen"
	"github.com/ochamekan/ms/metadataservice/pkg/model"
	"github.com/ochamekan/ms/movieservice/internal/controller/movie"
	"github.com/ochamekan/ms/pkg/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	gen.UnimplementedMovieServiceServer
	ctrl   *movie.Controller
	logger *zap.Logger
}

func New(ctrl *movie.Controller, logger *zap.Logger) *Handler {
	return &Handler{ctrl: ctrl, logger: logger.With(zap.String(logging.FieldComponent, "movie handler"))}
}

func (h *Handler) GetMovieDetails(ctx context.Context, req *gen.GetMovieDetailsRequest) (*gen.GetMovieDetailsResponse, error) {
	logger := h.logger.With(zap.String(logging.FieldEndpoint, "GetMovieDetails"))
	if req == nil || req.MovieId <= 0 {
		logger.Warn("nil request or incorrect movie id")
		return nil, status.Errorf(codes.InvalidArgument, "nil req or incorrect movie id")
	}

	logger.Info("Getting movie details")
	m, err := h.ctrl.Get(ctx, int(req.MovieId))
	if err != nil && errors.Is(err, movie.ErrNotFound) {
		logger.Error("Failed to get movie details", zap.Error(err))
		return nil, status.Error(codes.NotFound, err.Error())
	} else if err != nil {
		logger.Error("Failed to get movie details", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	logger.Info("Successfully retrieved movie details")
	return &gen.GetMovieDetailsResponse{
		MovieDetails: &gen.MovieDetails{
			Metadata: model.MetadataToProto(&m.Metadata),
			Rating:   m.Rating,
		},
	}, nil
}

package grpc

import (
	"context"
	"errors"
	"time"

	"github.com/ochamekan/ms/gen"
	"github.com/ochamekan/ms/metadataservice/pkg/model"
	"github.com/ochamekan/ms/movieservice/internal/controller/movie"
	"github.com/ochamekan/ms/pkg/logging"
	"github.com/ochamekan/ms/pkg/metrics"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	gen.UnimplementedMovieServiceServer
	ctrl    *movie.Controller
	logger  *zap.Logger
	metrics *metrics.Metrics
}

func New(ctrl *movie.Controller, logger *zap.Logger, metrics *metrics.Metrics) *Handler {
	return &Handler{ctrl: ctrl, logger: logger.With(zap.String(logging.FieldComponent, "movie handler")), metrics: metrics}
}

func (h *Handler) GetMovieDetails(ctx context.Context, req *gen.GetMovieDetailsRequest) (*gen.GetMovieDetailsResponse, error) {
	start := time.Now()
	defer func() {
		h.metrics.ObserveMovieGetDuration(time.Since(start).Seconds())
	}()

	logger := h.logger.With(zap.String(logging.FieldEndpoint, "GetMovieDetails"))
	if req == nil || req.MovieId <= 0 {
		h.metrics.IncMovieGetTotalCount(metrics.WarningOutcome)
		logger.Warn("nil request or incorrect movie id")
		return nil, status.Errorf(codes.InvalidArgument, "nil req or incorrect movie id")
	}

	logger.Info("Getting movie details")
	m, err := h.ctrl.Get(ctx, int(req.MovieId))
	if err != nil && errors.Is(err, movie.ErrNotFound) {
		h.metrics.IncMovieGetTotalCount(metrics.ErrorOutcome)
		logger.Error("Failed to get movie details", zap.Error(err))
		return nil, status.Error(codes.NotFound, err.Error())
	} else if err != nil {
		h.metrics.IncMovieGetTotalCount(metrics.ErrorOutcome)
		logger.Error("Failed to get movie details", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	// TODO: make as one
	h.metrics.IncMovieGetTotalCount(metrics.SuccessOutcome)
	h.metrics.IncMoviePopularityCount(m.Metadata.Title)
	logger.Info("Successfully retrieved movie details")
	return &gen.GetMovieDetailsResponse{
		MovieDetails: &gen.MovieDetails{
			Metadata: model.MetadataToProto(&m.Metadata),
			Rating:   m.Rating,
		},
	}, nil
}

package grpc

import (
	"context"
	"errors"

	"github.com/ochamekan/ms/gen"
	"github.com/ochamekan/ms/metadataservice/internal/controller/metadata"
	"github.com/ochamekan/ms/metadataservice/pkg/model"
	"github.com/ochamekan/ms/pkg/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	gen.UnimplementedMetadataServiceServer
	ctrl   *metadata.Controller
	logger *zap.Logger
}

func New(ctrl *metadata.Controller, logger *zap.Logger) *Handler {
	return &Handler{ctrl: ctrl, logger: logger.With(zap.String(logging.FieldComponent, "metadata handler"))}
}

func (h *Handler) GetMetadata(ctx context.Context, req *gen.GetMetadataRequest) (*gen.GetMetadataResponse, error) {
	logger := h.logger.With(zap.String(logging.FieldEndpoint, "GetMetadata"))
	if req == nil || req.Id <= 0 {
		logger.Warn("Nil request or incorrect movie id")
		return nil, status.Errorf(codes.InvalidArgument, "nil req or incorrect movie id")
	}

	logger.Info("Getting metadata by id")
	m, err := h.ctrl.GetMetadata(ctx, int(req.Id))
	if err != nil && errors.Is(err, metadata.ErrNotFound) {
		logger.Error("Failed to get metadata", zap.Error(err))
		return nil, status.Error(codes.NotFound, err.Error())
	} else if err != nil {
		logger.Error("Failed to get metadata", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	logger.Info("Successfully retrieved metadata")
	return &gen.GetMetadataResponse{Metadata: model.MetadataToProto(m)}, nil
}

func (h *Handler) PutMetadata(ctx context.Context, req *gen.PutMetadataRequest) (*gen.PutMetadataResponse, error) {
	logger := h.logger.With(zap.String(logging.FieldEndpoint, "PutMetadata"))
	if req == nil || req.Title == "" || req.Description == "" || req.Year <= 0 || req.Director == "" {
		logger.Warn("Nil request or bad request")
		return nil, status.Errorf(codes.InvalidArgument, "nil req or bad request")
	}

	logger.Info("Putting metadata")
	err := h.ctrl.PutMovieData(ctx, &model.Metadata{Title: req.Title, Description: req.Description, Year: int(req.Year), Director: req.Director})
	if err != nil {
		logger.Error("Failed to put metadata", zap.Error(err))
		return nil, err
	}

	logger.Info("Metadata successfully added")
	return &gen.PutMetadataResponse{}, nil
}

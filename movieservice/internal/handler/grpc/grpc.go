package grpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/ochamekan/ms/gen"
	"github.com/ochamekan/ms/metadataservice/pkg/model"
	"github.com/ochamekan/ms/movieservice/internal/controller/movie"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	gen.UnimplementedMovieServiceServer
	ctrl *movie.Controller
}

func New(ctrl *movie.Controller) *Handler {
	return &Handler{ctrl: ctrl}
}

func (h *Handler) GetMovieDetails(ctx context.Context, req *gen.GetMovieDetailsRequest) (*gen.GetMovieDetailsResponse, error) {
	if req == nil || req.MovieId <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "nil req or incorrect movie id")
	}
	m, err := h.ctrl.Get(ctx, int(req.MovieId))
	if err != nil && errors.Is(err, movie.ErrNotFound) {
		return nil, status.Error(codes.NotFound, err.Error())
	} else if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	fmt.Println("handler rating: ", m.Rating)

	return &gen.GetMovieDetailsResponse{
		MovieDetails: &gen.MovieDetails{
			Metadata: model.MetadataToProto(&m.Metadata),
			Rating:   m.Rating,
		},
	}, nil
}

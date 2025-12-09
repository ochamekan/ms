package memory

import (
	"context"

	"github.com/ochamekan/ms/ratingservice/internal/repository"
	"github.com/ochamekan/ms/ratingservice/pkg/model"
)

type Repository struct {
	data map[model.MovieID][]model.Rating
}

func New() *Repository {
	return &Repository{map[model.MovieID][]model.Rating{}}
}

func (r *Repository) Get(_ context.Context, movieID model.MovieID) ([]model.Rating, error) {
	if ratings, ok := r.data[movieID]; !ok || len(ratings) == 0 {
		return nil, repository.ErrNotFound
	}
	return r.data[movieID], nil
}

func (r *Repository) Put(_ context.Context, movieID model.MovieID, rating *model.Rating) error {
	r.data[movieID] = append(r.data[movieID], *rating)
	return nil
}

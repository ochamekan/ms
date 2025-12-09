package postgres

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ochamekan/ms/ratingservice/pkg/model"
)

type Repository struct {
	db *pgxpool.Pool
}

func New() (*Repository, func(), error) {
	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, nil, err
	}

	closer := func() {
		dbpool.Close()
	}

	return &Repository{dbpool}, closer, nil
}

func (r *Repository) Get(ctx context.Context, movieID model.MovieID) ([]model.Rating, error) {
	rows, err := r.db.Query(ctx, "SELECT * FROM ratings WHERE movie_id = $1", movieID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ratings []model.Rating

	for rows.Next() {
		var r model.Rating
		if err := rows.Scan(&r.ID, &r.MovieID, &r.Rating); err != nil {
			return nil, err
		}
		ratings = append(ratings, r)
	}

	return ratings, nil
}

func (r *Repository) Put(ctx context.Context, movieID model.MovieID, rating model.RatingValue) error {
	_, err := r.db.Exec(ctx, "INSERT INTO ratings (rating, movie_id) VALUES ($1, $2)", rating, movieID)
	return err
}

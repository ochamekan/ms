package postgres

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ochamekan/ms/rating/pkg/model"
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

func (r *Repository) Get(ctx context.Context, recordID model.RecordID, recordType model.RecordType) ([]model.Rating, error) {
	rows, err := r.db.Query(ctx, "SELECT * FROM ratings WHERE record_id = $1 AND record_type = $2", recordID, recordType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ratings []model.Rating

	for rows.Next() {
		var r model.Rating
		if err := rows.Scan(&r.RecordID, &r.RecordType, &r.UserID, &r.Value); err != nil {
			return nil, err
		}
		ratings = append(ratings, r)
	}

	return ratings, nil
}

func (r *Repository) Put(ctx context.Context, recordID model.RecordID, recordType model.RecordType, rating *model.Rating) error {
	_, err := r.db.Exec(ctx, "INSERT INTO ratings (user_id, record_id, record_type, rating_value) VALUES ($1, $2, $3, $4)", rating.UserID, recordID, recordType, rating.Value)
	return err
}

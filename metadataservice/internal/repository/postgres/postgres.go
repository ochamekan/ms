package postgres

import (
	"context"
	"database/sql"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ochamekan/ms/metadataservice/internal/repository"
	"github.com/ochamekan/ms/metadataservice/pkg/model"
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

func (r *Repository) Get(ctx context.Context, id int) (*model.Metadata, error) {
	var m model.Metadata

	row := r.db.QueryRow(ctx, "SELECT * FROM movies WHERE id = $1", id)
	err := row.Scan(&m.ID, &m.Title, &m.Year, &m.Description, &m.Director)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &m, nil
}

func (r *Repository) Put(ctx context.Context, metadata *model.Metadata) error {
	_, err := r.db.Exec(ctx, "INSERT INTO movies (title, year, description, director) VALUES ($1, $2, $3, $4)", metadata.Title, metadata.Year, metadata.Description, metadata.Director)
	return err

}

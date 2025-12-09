-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS movies (
  id serial PRIMARY KEY ,
  title varchar(255) NOT NULL,
  year integer NOT NULL,
  description text NOT NULL,
  director varchar(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS ratings (
  id serial PRIMARY KEY,
  movie_id integer NOT NULL,
  rating integer NOT NULL,
  FOREIGN KEY(movie_id) REFERENCES movies(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE ratings;
DROP TABLE movies;
-- +goose StatementEnd

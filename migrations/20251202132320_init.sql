-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS ratings (
  user_id varchar(255),
  record_id varchar(255),
  record_type varchar(255),
  rating_value integer,
  PRIMARY KEY(record_id, record_type, user_id)
);

CREATE TABLE IF NOT EXISTS movies (
  id serial PRIMARY KEY ,
  title varchar(255) NOT NULL,
  description text NOT NULL,
  director varchar(255) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE ratings;
DROP TABLE movies;
-- +goose StatementEnd

# ms (movie service)

A simple microservices-based system for managing movie metadata and user ratings, built as a learning project using **gRPC** and other technologies in Go.

## Run with docker compose

```shell
git clone https://github.com/ochamekan/ms

cd ms

cp .env.example .env

docker compose up -d
```

## Overview

**Metadata service**: Stores and retrieves movie metadata (title, description, year, director).

**Rating service**: Allows users to submit ratings for movies and retrieves the aggregated average rating.

**Movie service**: Acts as a composite service that combines metadata and aggregated rating to return complete movie details.

## Usage example

```
grpcurl -plaintext -d '{"movie_id": 15, "rating": 4}' localhost:8082 RatingService/PutRating

grpcurl -plaintext -d '{"movie_id": 15, "rating": 5}' localhost:8082 RatingService/PutRating

grpcurl -plaintext -d '{"movie_id": 15}' localhost:8083 MovieService/GetMovieDetails
```

## Service Discovery

**Consul** is used for service discovery, UI is accessible on `localhost:8500`.

## Cache

**Redis** caches aggregated ratings to avoid repeated calculations and stores movie metadata.

## Database

**PostgreSQL** provides persistent storage, using the **pgx** driver and **Goose** for migrations.

## Metrics & Logs

**zap** handles structured logging, with logs stored in **Loki** via **Alloy**.

**Prometheus** collects basic metrics.

Everything is visualized in **Grafana** on `localhost:3000` with simple predefined dashboard.
 
 




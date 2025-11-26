package main

import (
	"log"
	"net/http"

	"github.com/ochamekan/ms/movie/internal/controller/movie"
	metadatagateway "github.com/ochamekan/ms/movie/internal/gateway/metadata/http"
	ratinggateway "github.com/ochamekan/ms/movie/internal/gateway/rating/http"
	httphandler "github.com/ochamekan/ms/movie/internal/handler/http"
)

func main() {
	log.Println("Starting the movie service")
	metadataGateway := metadatagateway.New("localhost:8081")
	ratingGateway := ratinggateway.New("localhost:8082")
	ctrl := movie.New(ratingGateway, metadataGateway)

	h := httphandler.New(ctrl)
	http.HandleFunc("/movie", h.GetMovieDetails)
	if err := http.ListenAndServe("8083", nil); err != nil {
		panic(err)
	}
}

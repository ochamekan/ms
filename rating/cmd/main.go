package main

import (
	"log"
	"net/http"

	"github.com/ochamekan/ms/rating/internal/controller/rating"
	httphandler "github.com/ochamekan/ms/rating/internal/handler/http"
	"github.com/ochamekan/ms/rating/internal/repository/memory"
)

func main() {
	log.Println("Starting the rating service")
	repo := memory.New()
	ctrl := rating.New(repo)
	h := httphandler.New(ctrl)

	http.HandleFunc("/rating", h.Handle)
	if err := http.ListenAndServe(":8082", nil); err != nil {
		panic(err)
	}
}

package main

import (
	"log"
	"net/http"

	"github.com/ochamekan/ms/metadata/internal/controller/metadata"
	httphandler "github.com/ochamekan/ms/metadata/internal/handler/http"
	"github.com/ochamekan/ms/metadata/internal/repository/memory"
)

func main() {
	log.Println("Starting the movie metadata service")
	repo := memory.New()
	ctrl := metadata.New(repo)
	h := httphandler.New(ctrl)

	http.HandleFunc("/metadata", h.GetMetadata)
	if err := http.ListenAndServe(":8081", nil); err != nil {
		panic(err)
	}
}

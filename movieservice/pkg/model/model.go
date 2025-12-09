package model

import "github.com/ochamekan/ms/metadataservice/pkg/model"

type MovieDetails struct {
	Rating   *float64       `json:"rating,omitempty"`
	Metadata model.Metadata `json:"metadata"`
}

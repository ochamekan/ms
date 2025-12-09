package model

type (
	MovieID     int
	RatingValue int
)

type Rating struct {
	ID      int         `json:"id"`
	MovieID MovieID     `json:"movie_id"`
	Rating  RatingValue `json:"rating"`
}

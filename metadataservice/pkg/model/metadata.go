package model

type Metadata struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Year        int    `json:"year"`
	Description string `json:"description"`
	Director    string `json:"director"`
}

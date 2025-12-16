package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	// TODO: make first 2 as one metric
	MovieGetDetailsTotal    *prometheus.CounterVec
	MovieFilmPopularity     *prometheus.CounterVec
	MovieGetDetailsDuration prometheus.Histogram
}

type RequestOutcome string

const (
	SuccessOutcome RequestOutcome = "success"
	ErrorOutcome   RequestOutcome = "error"
	WarningOutcome RequestOutcome = "warning"
)

func New(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		MovieGetDetailsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "movie_get_details_requests_total",
			Help: "Number of requests",
		}, []string{"outcome"}),
		MovieFilmPopularity: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "movie_popularity_count",
			Help: "Movie popularity",
		}, []string{"film_name"}),
		MovieGetDetailsDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "movie_get_details_duration",
			Help: "Duration of the request",
		}),
	}
	reg.MustRegister(m.MovieGetDetailsTotal, m.MovieFilmPopularity, m.MovieGetDetailsDuration)

	return m
}

func (m *Metrics) IncMovieGetTotalCount(outcome RequestOutcome) {
	m.MovieGetDetailsTotal.WithLabelValues(string(outcome)).Inc()
}

func (m *Metrics) IncMoviePopularityCount(filmName string) {
	m.MovieFilmPopularity.WithLabelValues(filmName).Inc()
}

func (m *Metrics) ObserveMovieGetDuration(durationSecs float64) {
	m.MovieGetDetailsDuration.Observe(durationSecs)
}

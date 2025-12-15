package metrics

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	TotalCalls prometheus.Counter
}

func New(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		TotalCalls: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "grpc_request_count",
			Help: "Number of request handler by this handler",
		}),
	}

	reg.MustRegister(m.TotalCalls)

	return m
}

func (m *Metrics) ObserveTotalCalls(method string) {
	m.TotalCalls.Inc()
}

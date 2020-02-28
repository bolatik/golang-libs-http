package http

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	handlerDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_handlers_duration_seconds",
		Help: "Handlers request duration in seconds",
	}, []string{"path"})
)

func init() {
	prometheus.MustRegister(handlerDuration)
}

func Metrics(path string, fn Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		fn(w, r)
		handlerDuration.WithLabelValues(path).Observe(time.Since(now).Seconds())
	}
}

func NewMetricsServer(healthCheckFn Endpoint) *http.Server {
	s := &http.Server{}

	r := mux.NewRouter()

	r.Handle("/metrics", promhttp.Handler())
	r.HandleFunc("/health", Json(healthCheckFn)).Methods("GET")

	s.Handler = r
	s.Addr = ":10101"

	go func() {
		s.ListenAndServe()
	}()

	return s
}

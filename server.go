package http

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"

	"context"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server interface {
	ListenAndServe(shutDownFn func(), healthCheckFn Endpoint) error

	Handle(method string, path string, fn Endpoint)
}

type server struct {
	r   *mux.Router
	srv *http.Server
	log Logger
}

func NewServer(port string, log Logger) Server {
	s := &server{}
	r := mux.NewRouter()

	srv := &http.Server{}
	srv.Handler = r
	srv.Addr = ":" + port

	s.r = r
	s.srv = srv
	s.log = log
	s.r.NotFoundHandler = http.HandlerFunc(Json(func(w http.ResponseWriter, r *http.Request) Response {
		return &Error{
			Message: "Route not found: " + r.URL.Path,
			Status:  404,
			System:  "HTTP",
			Series:  404,
		}
	}))

	return s
}

func (s *server) Handle(method string, path string, fn Endpoint) {
	s.r.HandleFunc(path, Metrics(path, Json(
		Logging(s.log, fn),
	))).Methods(method)
}

func (s *server) ListenAndServe(shutDownFn func(), healthCheckFn Endpoint) error {
	idleConnsClosed := make(chan struct{})
	metricsServer := NewMetricsServer(healthCheckFn)

	s.r.Handle("/_/metrics", promhttp.Handler())
	s.r.HandleFunc("/_/health", Json(healthCheckFn)).Methods("GET")

	go func() {
		sigint := make(chan os.Signal, 1)

		// interrupt signal sent from terminal
		signal.Notify(sigint, os.Interrupt)
		// sigterm signal sent from kubernetes
		signal.Notify(sigint, syscall.SIGTERM)

		<-sigint

		s.log.Debug("Shutting down the server")
		// We received an interrupt signal, shut down.
		if err := s.srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			s.log.Warn("HTTP server Shutdown: ", err.Error())
		}
		s.log.Debug("Shutting down the metrics server")
		if err := metricsServer.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			s.log.Warn("Metrics HTTP server Shutdown: ", err.Error())
		}

		shutDownFn()
		close(idleConnsClosed)
	}()

	if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		s.log.Warn("HTTP server ListenAndServe: ", err.Error())
		return err
	}

	<-idleConnsClosed

	return nil
}

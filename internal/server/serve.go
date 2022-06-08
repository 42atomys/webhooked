package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"

	"atomys.codes/webhooked/internal/config"
	v1alpha1 "atomys.codes/webhooked/internal/server/v1alpha1"
)

// APIVersion is the interface for all supported API versions
// that can be served by the webhooked server
type APIVersion interface {
	Version() string
	WebhookHandler() http.HandlerFunc
}

type Server struct {
	*http.Server
}

var (
	// apiVersions is a list of supported API versions by the server
	apiVersions = []APIVersion{
		v1alpha1.NewServer(),
	}
)

// NewServer create a new server instance with the given port
func NewServer(port int) (*Server, error) {
	if !validPort(port) {
		return nil, fmt.Errorf("invalid port")
	}

	return &Server{
		Server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: nil,
		},
	}, nil
}

// Serve the proxy server on the given port for all supported API versions
func (s *Server) Serve() error {
	router := newRouter()
	router.Use(loggingMiddleware)

	if config.Current().Observability.MetricsEnabled {
		router.Use(prometheusMiddleware)
		router.Handle("/metrics", promhttp.Handler()).Name("metrics")
	}

	s.Handler = router
	log.Info().Msgf("Listening on %s", s.Addr)
	return s.ListenAndServe()
}

// newRouter returns a new router with all the routes
// for all supported API versions
func newRouter() *mux.Router {
	var api = mux.NewRouter()
	for _, version := range apiVersions {
		api.Methods("POST").PathPrefix("/" + version.Version()).Handler(version.WebhookHandler()).Name(version.Version())
	}

	api.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	return api
}

// validPort returns true if the port is valid
// following the RFC https://datatracker.ietf.org/doc/html/rfc6056#section-2.1
func validPort(port int) bool {
	return port > 0 && port < 65535
}

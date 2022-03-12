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

type APIVersion interface {
	Version() string
	WebhookHandler() http.HandlerFunc
}

var (
	// apiVersions is a list of supported API versions by the server
	apiVersions = []APIVersion{
		v1alpha1.NewServer(),
	}
)

// Serve the proxy server on the given port for all supported API versions
func Serve(port int) error {
	if !validPort(port) {
		return fmt.Errorf("invalid port")
	}

	log.Info().Msgf("Listening on port %d", port)
	router := newRouter()
	router.Use(loggingMiddleware)

	if config.Current().Observability.MetricsEnabled {
		router.Use(prometheusMiddleware)
		router.Handle("/metrics", promhttp.Handler()).Name("metrics")
	}

	return http.ListenAndServe(fmt.Sprintf(":%d", port), router)
}

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

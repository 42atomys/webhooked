package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	v1alpha1 "42stellar.org/webhooks/internal/server/v1alpha1"
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
	return http.ListenAndServe(fmt.Sprintf(":%d", port), newRouter())
}

func newRouter() *mux.Router {
	var api = mux.NewRouter()
	for _, version := range apiVersions {
		api.Methods("POST").PathPrefix("/" + version.Version()).Handler(version.WebhookHandler())
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

package server

import (
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"

	"42stellar.org/webhooks/internal/config"
)

type Server struct {
	config         *config.Configuration
	webhookService func(spec *config.WebhookSpec) error
}

func NewServer() *Server {
	return &Server{
		config:         config.Current(),
		webhookService: webhookService,
	}
}

func (s *Server) Version() string {
	return "v1alpha1"
}

func (s *Server) WebhookHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.config.APIVersion != s.Version() {
			log.Error().Str("apiVersion", s.Version()).Msg("Configuration don't match with the API version")
			w.WriteHeader(http.StatusBadRequest)
		}

		spec, err := s.config.GetSpecByEndpoint(strings.ReplaceAll(r.URL.Path, "/"+s.Version(), ""))
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if err := s.webhookService(spec); err != nil {
			log.Error().Err(err).Msg("Error while processing webhook")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func webhookService(spec *config.WebhookSpec) error {
	if spec == nil {
		return config.ErrSpecNotFound
	}

	if spec.HasSecurity() {
		// TODO Do security Layer
		log.Warn().Msg("Security layer not implemented yet")
	}

	// TODO Do the webhook storage
	return nil
}

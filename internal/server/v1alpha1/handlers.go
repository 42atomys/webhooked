package server

import (
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"42stellar.org/webhooks/internal/config"
)

type Server struct {
	config         *config.Configuration
	webhookService func(s *Server, spec *config.WebhookSpec, data []byte) error
	logger         zerolog.Logger
}

func NewServer() *Server {
	var s = &Server{
		config:         config.Current(),
		webhookService: webhookService,
	}

	s.logger = log.With().Str("apiVersion", s.Version()).Logger()
	return s
}

func (s *Server) Version() string {
	return "v1alpha1"
}

func (s *Server) WebhookHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.config.APIVersion != s.Version() {
			s.logger.Error().Msg("Configuration don't match with the API version")
			w.WriteHeader(http.StatusBadRequest)
		}

		spec, err := s.config.GetSpecByEndpoint(strings.ReplaceAll(r.URL.Path, "/"+s.Version(), ""))
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		data, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := s.webhookService(s, spec, data); err != nil {
			s.logger.Error().Err(err).Msg("Error while processing webhook")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func webhookService(s *Server, spec *config.WebhookSpec, data []byte) error {
	if spec == nil {
		return config.ErrSpecNotFound
	}
	defer s.logger.Debug().Str("entry", spec.Name).Msg("Webhook processed")

	if spec.HasSecurity() {
		// TODO Do security Layer
		s.logger.Warn().Msg("Security layer not implemented yet")
	}

	// TODO Do the webhook storage
	s.logger.Warn().Msg("Storage not implemented yet")

	for _, storage := range spec.Storages {
		if err := storage.Client.Push(string(data)); err != nil {
			return err
		}
	}

	return nil
}

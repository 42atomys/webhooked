package server

import (
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"42stellar.org/webhooks/internal/config"
	"42stellar.org/webhooks/pkg/factory"
)

type Server struct {
	config         *config.Configuration
	webhookService func(s *Server, spec *config.WebhookSpec, r *http.Request) error
	logger         zerolog.Logger
}

func NewServer() *Server {
	var s = &Server{
		config:         config.Current(),
		webhookService: webhookService,
	}

	s.logger = log.With().Str("apiVersion", s.Version()).Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})
	return s
}

func (s *Server) Version() string {
	return "v1alpha1"
}

func (s *Server) WebhookHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.config.APIVersion != s.Version() {
			s.logger.Error().Msgf("Configuration %s don't match with the API version %s", s.config.APIVersion, s.Version())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Debug().Msgf("endpoint: %+v", strings.ReplaceAll(r.URL.Path, "/"+s.Version(), ""))
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
    
		if err := s.webhookService(s, spec, r); err != nil {
			switch err {
			case factory.ErrSecurityFailed:
				w.WriteHeader(http.StatusForbidden)
				return
			default:
				s.logger.Error().Err(err).Msg("Error while processing webhook")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
}

func webhookService(s *Server, spec *config.WebhookSpec, r *http.Request) error {
	if spec == nil {
		return config.ErrSpecNotFound
	}
	defer s.logger.Debug().Str("entry", spec.Name).Msg("Webhook processed")

	if spec.HasSecurity() {
		if err := s.runSecurity(spec, r); err != nil {
			return err
		}
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

func (s *Server) runSecurity(spec *config.WebhookSpec, r *http.Request) error {
	if spec == nil {
		return config.ErrSpecNotFound
	}

	ok, err := factory.Run(spec.SecurityFactories, func(factory *factory.Factory, lastOutput string, defaultFunc factory.RunnerFunc) (string, error) {
		switch factory.Name {
		case "getHeader":
			return factory.Fn(factory.Config, "", r.Header)
		case "compareWithStaticValue":
			return factory.Fn(factory.Config, lastOutput)
		}
		return defaultFunc(factory, lastOutput)
	})

	if err != nil {
		log.Error().Err(err).Msg("Error while processing security factory")
		return err
	}

	log.Debug().Msgf("security factory passed: %t", ok)
	if !ok {
		return factory.ErrSecurityFailed
	}
	return nil
}

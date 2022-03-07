package server

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"42stellar.org/webhooks/internal/config"
)

type Server struct {
	config         *config.Configuration
	webhookService func(s *Server, spec *config.WebhookSpec, r *http.Request) error
	logger         zerolog.Logger
}

var errSecurityFailed = errors.New("security failed")

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

		if err := s.webhookService(s, spec, r); err != nil {
			switch err {
			case errSecurityFailed:
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

func webhookService(s *Server, spec *config.WebhookSpec, r *http.Request) (err error) {
	if spec == nil {
		return config.ErrSpecNotFound
	}
	defer s.logger.Debug().Str("entry", spec.Name).Msg("Webhook processed")

	if r.Body == nil {
		return errors.New("request don't have body")
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	if spec.HasSecurity() {
		if err := s.runSecurity(spec, r, data); err != nil {
			return err
		}
	}

	for _, storage := range spec.Storage {
		if err := storage.Client.Push(string(data)); err != nil {
			return err
		}
	}

	return err
}

func (s *Server) runSecurity(spec *config.WebhookSpec, r *http.Request, body []byte) error {
	if spec == nil {
		return config.ErrSpecNotFound
	}

	pipeline := spec.SecurityPipeline
	if pipeline == nil {
		return errors.New("no pipeline to run. security is not configured")
	}

	pipeline.Inputs["request"] = r
	pipeline.Inputs["payload"] = string(body)

	pipeline.WantResult(true).Run()

	log.Debug().Msgf("security factory passed: %t", pipeline.CheckResult())
	if !pipeline.CheckResult() {
		return errSecurityFailed
	}
	return nil
}

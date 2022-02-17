package server

import "net/http"

type Server struct{}

func (s Server) Version() string {
	return "v1alpha1"
}

func (s Server) WebhookHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO valpha1 handler
	}
}

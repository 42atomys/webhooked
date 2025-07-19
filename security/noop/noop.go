package noop

import (
	"github.com/valyala/fasthttp"
)

type NoopSecuritySpec struct{}

func (s *NoopSecuritySpec) EnsureConfigurationCompleteness() error {
	return nil
}

func (s *NoopSecuritySpec) Initialize() error {
	return nil
}

func (s *NoopSecuritySpec) IsSecure(ctx *fasthttp.RequestCtx) (bool, error) {
	return true, nil
}

package noop

import (
	"context"

	"github.com/42atomys/webhooked/internal/fasthttpz"
)

type NoopSecuritySpec struct{}

func (s *NoopSecuritySpec) EnsureConfigurationCompleteness() error {
	return nil
}

func (s *NoopSecuritySpec) Initialize() error {
	return nil
}

func (s *NoopSecuritySpec) IsSecure(ctx context.Context, rctx *fasthttpz.RequestCtx) (bool, error) {
	return true, nil
}

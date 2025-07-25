package security

import (
	"context"

	"github.com/42atomys/webhooked/internal/fasthttpz"
)

type Security struct {
	Type  string `json:"type"`
	Specs Specs  `json:"specs"`
}

type Specs interface {
	EnsureConfigurationCompleteness() error
	Initialize() error
	IsSecure(ctx context.Context, rctx *fasthttpz.RequestCtx) (bool, error)
}

func (s *Security) IsSecure(ctx context.Context, rctx *fasthttpz.RequestCtx) (bool, error) {
	return s.Specs.IsSecure(ctx, rctx)
}

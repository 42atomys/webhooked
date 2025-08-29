package security

import (
	"context"
	"errors"

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

func (s *Security) EnsureConfigurationCompleteness() error {
	if s.Type == "" {
		return errors.New("security type is not defined")
	}

	if s.Specs == nil {
		return errors.New("security specs are not defined")
	}

	return s.Specs.EnsureConfigurationCompleteness()
}

func (s *Security) IsSecure(ctx context.Context, rctx *fasthttpz.RequestCtx) (bool, error) {
	return s.Specs.IsSecure(ctx, rctx)
}

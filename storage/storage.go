package storage

import (
	"context"

	"github.com/42atomys/webhooked/format"
	"github.com/rs/zerolog/log"
)

type Storage struct {
	Type       string             `json:"type"`
	Formatting *format.Formatting `json:"formatting"`
	Specs      Specs              `json:"specs"`
}

type Specs interface {
	EnsureConfigurationCompleteness() error
	Initialize() error
	Store(ctx context.Context, value []byte) error
}

func (s *Storage) Store(ctx context.Context, value []byte) error {
	log.Debug().Msgf("Storing data in %s storage", s.Type)
	return s.Specs.Store(ctx, value)
}

func (s *Storage) TemplateContext() map[string]any {
	return map[string]any{
		"StorageType": s.Type,
	}
}

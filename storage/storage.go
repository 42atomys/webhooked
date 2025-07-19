package storage

import (
	"context"

	"github.com/42atomys/webhooked/format"
)

type Storage struct {
	Type       string            `json:"type"`
	Formatting format.Formatting `json:"formatting"`
	Specs      Specs             `json:"specs"`
}

type Specs interface {
	EnsureConfigurationCompleteness() error
	Initialize() error
	Store(ctx context.Context, value []byte) error
}

func (s *Storage) Store(ctx context.Context, value []byte) error {
	return s.Specs.Store(ctx, value)
}

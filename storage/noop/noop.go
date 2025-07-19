package noop

import (
	"context"
)

type NoopStorageSpec struct{}

func (s *NoopStorageSpec) EnsureConfigurationCompleteness() error {
	return nil
}

func (s *NoopStorageSpec) Initialize() error {
	return nil
}

func (s *NoopStorageSpec) Store(ctx context.Context, value []byte) error {
	return nil
}

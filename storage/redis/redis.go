package redis

import (
	"errors"
	"fmt"

	"context"

	"github.com/42atomys/webhooked/internal/valuable"
	"github.com/go-redis/redis/v8"
)

type RedisStorageSpec struct {
	Host     valuable.Valuable `json:"host"`
	Port     valuable.Valuable `json:"port"`
	Username valuable.Valuable `json:"username"`
	Password valuable.Valuable `json:"password"`
	Database int               `json:"database"`
	Key      string            `json:"key"`

	client *redis.Client
}

func (s *RedisStorageSpec) EnsureConfigurationCompleteness() error {
	if s.Host.First() == "" {
		return errors.New("host is required")
	}

	if s.Port.First() == "" {
		return errors.New("port is required")
	}

	return nil
}

func (s *RedisStorageSpec) Initialize() error {
	s.client = redis.NewClient(
		&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", s.Host, s.Port),
			Username: s.Username.First(),
			Password: s.Password.First(),
			DB:       s.Database,
		},
	)

	// Ping Redis for testing config
	if err := s.client.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("error pinging Redis: %w", err)
	}

	return nil
}

func (s *RedisStorageSpec) Store(ctx context.Context, value []byte) error {
	if err := s.client.RPush(ctx, s.Key, value).Err(); err != nil {
		return fmt.Errorf("error storing value in Redis: %w", err)
	}
	return nil
}

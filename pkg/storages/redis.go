package storages

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mitchellh/mapstructure"
)

type RedisStorage struct {
	client *redis.Client
	config *redisConfig
	ctx    context.Context
}

type redisConfig struct {
	Host       string
	Port       string
	Database   int
	Username   string
	Password   string
	Key        string
	Expiration time.Duration
}

func NewRedisStorage(config map[string]interface{}) (*RedisStorage, error) {

	newClient := RedisStorage{
		config: &redisConfig{},
		ctx:    context.Background(),
	}

	if err := mapstructure.Decode(config, &newClient.config); err != nil {
		return nil, err
	}

	newClient.client = redis.NewClient(
		&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", newClient.config.Host, newClient.config.Port),
			Username: newClient.config.Username,
			Password: newClient.config.Password,
			DB:       newClient.config.Database,
		},
	)

	// Ping Redis for testing config
	if err := newClient.client.Ping(newClient.ctx).Err(); err != nil {
		return nil, err
	}

	return &newClient, nil
}

func (c RedisStorage) Name() string {
	return "redis"
}

func (c RedisStorage) Push(value interface{}) error {
	if err := c.client.RPush(c.ctx, c.config.Key, value, c.config.Expiration).Err(); err != nil {
		return err
	}

	return nil
}

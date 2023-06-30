package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"

	"atomys.codes/webhooked/internal/valuable"
)

type storage struct {
	client *redis.Client
	config *config
}

type config struct {
	Host     valuable.Valuable `mapstructure:"host" json:"host"`
	Port     valuable.Valuable `mapstructure:"port" json:"port"`
	Username valuable.Valuable `mapstructure:"username" json:"username"`
	Password valuable.Valuable `mapstructure:"password" json:"password"`
	Database int               `mapstructure:"database" json:"database"`
	Key      string            `mapstructure:"key" json:"key"`
}

// NewStorage is the function for create new Redis storage client
// Run is made from external caller at begins programs
// @param config contains config define in the webhooks yaml file
// @return RedisStorage the struct contains client connected and config
// @return an error if the the client is not initialized successfully
func NewStorage(configRaw map[string]interface{}) (*storage, error) {

	newClient := storage{
		config: &config{},
	}

	if err := valuable.Decode(configRaw, &newClient.config); err != nil {
		return nil, err
	}

	newClient.client = redis.NewClient(
		&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", newClient.config.Host, newClient.config.Port),
			Username: newClient.config.Username.First(),
			Password: newClient.config.Password.First(),
			DB:       newClient.config.Database,
		},
	)

	// Ping Redis for testing config
	if err := newClient.client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &newClient, nil
}

// Name is the function for identified if the storage config is define in the webhooks
// @return name of the storage
func (c storage) Name() string {
	return "redis"
}

// Push is the function for push data in the storage
// A run is made from external caller
// @param value that will be pushed
// @return an error if the push failed
func (c storage) Push(ctx context.Context, value []byte) error {
	if err := c.client.RPush(ctx, c.config.Key, value).Err(); err != nil {
		return err
	}

	return nil
}

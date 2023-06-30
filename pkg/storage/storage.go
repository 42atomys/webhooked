package storage

import (
	"context"
	"fmt"

	"atomys.codes/webhooked/pkg/storage/postgres"
	"atomys.codes/webhooked/pkg/storage/rabbitmq"
	"atomys.codes/webhooked/pkg/storage/redis"
)

// Pusher is the interface for storage pusher
// The name must be unique and must be the same as the storage type, the Push
// function will be called with the receiving data
type Pusher interface {
	// Get the name of the storage
	// Will be unique across all storages
	Name() string
	// Method call when insert new data in the storage
	Push(ctx context.Context, value []byte) error
}

// Load will fetch and return the built-in storage based on the given
// storageType params and initialize it with given storageSpecs given
func Load(storageType string, storageSpecs map[string]interface{}) (pusher Pusher, err error) {
	switch storageType {
	case "redis":
		pusher, err = redis.NewStorage(storageSpecs)
	case "postgres":
		pusher, err = postgres.NewStorage(storageSpecs)
	case "rabbitmq":
		pusher, err = rabbitmq.NewStorage(storageSpecs)
	default:
		err = fmt.Errorf("storage %s is undefined", storageType)
	}
	return
}

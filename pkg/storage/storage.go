package storage

import (
	"fmt"

	"42stellar.org/webhooks/pkg/storage/postgres"
	"42stellar.org/webhooks/pkg/storage/rabbitmq"
	"42stellar.org/webhooks/pkg/storage/redis"
)

type Pusher interface {
	// Get the name of the storage
	// Will be unique across all storages
	Name() string
	// Method call when insert new data in the storage
	Push(value interface{}) error
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

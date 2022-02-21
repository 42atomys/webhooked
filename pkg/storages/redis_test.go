package storages

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedisName(t *testing.T) {
	newRedis := RedisStorage{}
	assert.Equal(t, "redis", newRedis.Name())
}

func TestRedisNewRedisStorage(t *testing.T) {
	storageSpec := map[string]interface{}{
		"host":     "localhost",
		"port":     "6379",
		"database": 0,
		"key":      "LOL",
	}

	_, err := NewRedisStorage(storageSpec)
	assert.Nil(t, err)
}

func TestRedisPush(t *testing.T) {
	newClient, err := NewRedisStorage(map[string]interface{}{
		"host":     "localhost",
		"port":     "6379",
		"database": 0,
		"key":      "LOL",
	})
	assert.Nil(t, err)

	err = newClient.Push("Hello")
	assert.Nil(t, err)
}

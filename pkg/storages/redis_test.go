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
	_, err := NewRedisStorage(map[string]interface{}{
		"hostt":    "127.0.0.1",
		"database": 0,
		"key":      "LOL",
	})
	assert.NotNil(t, err)

	_, err = NewRedisStorage(map[string]interface{}{})
	assert.NotNil(t, err)

	_, err = NewRedisStorage(map[string]interface{}{
		"host":     "127.0.0.1",
		"port":     "6379",
		"database": 0,
		"key":      "LOL",
	})
	assert.Nil(t, err)
}

func TestRedisPush(t *testing.T) {
	newClient, err := NewRedisStorage(map[string]interface{}{
		"host":     "127.0.0.1",
		"port":     "6379",
		"database": 0,
		"key":      "LOL",
	})
	assert.Nil(t, err)

	err = newClient.Push("Hello")
	assert.Nil(t, err)
}

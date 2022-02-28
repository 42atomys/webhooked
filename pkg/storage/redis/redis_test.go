package redis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedisName(t *testing.T) {
	newRedis := storage{}
	assert.Equal(t, "redis", newRedis.Name())
}

func TestRedisNewRedisStorage(t *testing.T) {
	_, err := NewStorage(map[string]interface{}{
		"host": []int{1},
	})
	assert.Error(t, err)

	_, err = NewStorage(map[string]interface{}{})
	assert.Error(t, err)

	_, err = NewStorage(map[string]interface{}{
		"host":     "127.0.0.1",
		"port":     "6379",
		"database": 0,
		"channel":  "testChannel",
	})
	assert.NoError(t, err)
}

func TestRedisPush(t *testing.T) {
	newClient, err := NewStorage(map[string]interface{}{
		"host":     "127.0.0.1",
		"port":     "6379",
		"database": 0,
		"channel":  "testChannel",
	})
	assert.NoError(t, err)

	err = newClient.Push(func() {})
	assert.Error(t, err)

	err = newClient.Push("Hello")
	assert.NoError(t, err)
}

package redis

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RedisSetupTestSuite struct {
	suite.Suite
}

func (suite *RedisSetupTestSuite) TestRedisName() {
	newRedis := storage{}
	assert.Equal(suite.T(), "redis", newRedis.Name())
}

func (suite *RedisSetupTestSuite) TestRedisNewStorage() {
	_, err := NewStorage(map[string]interface{}{
		"host": []int{1},
	})
	assert.Error(suite.T(), err)

	_, err = NewStorage(map[string]interface{}{})
	assert.Error(suite.T(), err)

	_, err = NewStorage(map[string]interface{}{
		"host":     "127.0.0.1",
		"port":     "6379",
		"database": 0,
		"key":      "testKey",
	})
	assert.NoError(suite.T(), err)
}

func (suite *RedisSetupTestSuite) TestRedisPush() {
	newClient, err := NewStorage(map[string]interface{}{
		"host":     "127.0.0.1",
		"port":     "6379",
		"database": 0,
		"key":      "testKey",
	})
	assert.NoError(suite.T(), err)

	err = newClient.Push(func() {})
	assert.Error(suite.T(), err)

	err = newClient.Push("Hello")
	assert.NoError(suite.T(), err)
}

func TestRunRedisPush(t *testing.T) {
	if testing.Short() {
		t.Skip("redis testing is skiped in short version of test")
		return
	}

	suite.Run(t, new(RedisSetupTestSuite))
}

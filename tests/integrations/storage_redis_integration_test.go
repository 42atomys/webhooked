package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	StorageTypeRedis storageType = "redis"
)

type RedisIntegrationTestSuite struct {
	IntegrationTestSuite
}

func (suite *RedisIntegrationTestSuite) SetupSuite() {
	suite.IntegrationTestSuite.SetupSuite()

	// Initialize Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	ctx := context.Background()
	err := client.Ping(ctx).Err()
	require.NoError(suite.T(), err, "Failed to connect to Redis")

	// Clean up test keys
	keys, err := client.Keys(ctx, "integration:*").Result()
	if err == nil && len(keys) > 0 {
		err = client.Del(ctx, keys...).Err()
		require.NoError(suite.T(), err, "Failed to clean up test keys")
	}

	suite.storages[StorageTypeRedis] = client
}

func (suite *RedisIntegrationTestSuite) TearDownSuite() {
	if client, ok := suite.storages[StorageTypeRedis].(*redis.Client); ok {
		client.Close()
	}
	suite.IntegrationTestSuite.TearDownSuite()
}

func (suite *RedisIntegrationTestSuite) TestRedisStorageScenarios() {
	tests := []testInput{
		{
			name:     "redis-basic-storage",
			endpoint: "/integration/redis-basic",
			headers: map[string]string{
				"X-Token": "integration-test",
			},
			payload: map[string]any{
				"event_type": "user.updated",
				"user_id":    67890,
				"timestamp":  "2023-06-28T19:30:00Z",
			},
			expectedResponse: expectedResponse{
				statusCode: 204,
			},
			expectedStorage: expectedStorage{
				storageType: StorageTypeRedis,
				key:         "integration:redis-basic",
				data:        `{"event_type":"user.updated","user_id":67890,"timestamp":"2023-06-28T19:30:00Z"}`,
				isJson:      true,
			},
		},
		{
			name:     "redis-formatted-storage",
			endpoint: "/integration/redis-formatted",
			headers: map[string]string{
				"X-Token":      "integration-test",
				"X-Delivery":   "xyz789",
				"Content-Type": "application/json",
			},
			payload: map[string]any{
				"action": "login",
				"ip":     "192.168.1.1",
				"type":   "mobile",
			},
			expectedResponse: expectedResponse{
				statusCode: 200,
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body: `{"status":"stored","delivery_id":"xyz789"}`,
			},
			expectedStorage: expectedStorage{
				storageType: StorageTypeRedis,
				key:         "integration:redis-formatted",
				data:        `redis|192.168.1.1|mobile`,
				isJson:      false,
			},
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			suite.runRedisTest(test)
		})
	}
}

func (suite *RedisIntegrationTestSuite) runRedisTest(test testInput) {
	suite.doRequest(test)

	// Additional Redis-specific verification
	if test.expectedStorage.storageType == StorageTypeRedis {
		client := suite.storages[StorageTypeRedis].(*redis.Client)

		time.Sleep(100 * time.Millisecond) // Allow time for async storage

		data, err := client.LPop(suite.ctx, test.expectedStorage.key).Result()
		if err != redis.Nil {
			suite.NoError(err, "Failed to get data from Redis")
		}
		if test.expectedStorage.isJson {
			suite.JSONEq(test.expectedStorage.data, data, "Data mismatch in Redis storage")
		} else {
			suite.Equal(test.expectedStorage.data, data, "Data mismatch in Redis storage")
		}
	}
}

func TestRedisIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration tests in short mode")
	}
	suite.Run(t, new(RedisIntegrationTestSuite))
}

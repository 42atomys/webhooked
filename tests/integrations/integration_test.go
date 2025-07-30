//go:build integrations

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type testInput struct {
	name             string
	endpoint         string
	headers          map[string]string
	payload          any
	expectedStorage  expectedStorage
	expectedResponse expectedResponse
}

type expectedResponse struct {
	statusCode int
	body       string
	headers    map[string]string
}

type expectedStorage struct {
	storageType storageType
	key         string
	data        string
	isJson      bool
}

type storageType string

const (
	BaseURL = "http://localhost:8081/webhooks/v1alpha2"
)

type IntegrationTestSuite struct {
	suite.Suite
	ctx      context.Context
	storages map[storageType]any
}

func (suite *IntegrationTestSuite) SetupSuite() {
	// Initialize logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	log.Logger = log.Logger.Level(zerolog.InfoLevel)

	suite.ctx = context.Background()

	redisHost, defined := os.LookupEnv("REDIS_HOST")
	if !defined {
		redisHost = "redis"
	}
	redisPort, defined := os.LookupEnv("REDIS_PORT")
	if !defined {
		redisPort = "6379"
	}
	redisPassword, defined := os.LookupEnv("REDIS_PASSWORD")
	if !defined {
		redisPassword = ""
	}

	// Initialize storage configuration
	redisclient := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		DB:       0, // use default DB
		Password: redisPassword,
	})
	suite.NoError(redisclient.Ping(suite.ctx).Err(), "Failed to create Redis client")
	suite.NoError(redisclient.FlushDB(suite.ctx).Err(), "Failed to flush Redis database")

	suite.storages = map[storageType]any{
		StorageTypeRedis: redisclient,
	}

	log.Info().Msg("Starting integration test suite...")
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	log.Info().Msg("Integration test suite completed")
}

func (suite *IntegrationTestSuite) TestIntegrationScenarios() {
	tests := []testInput{
		{
			name:     "empty-payload",
			endpoint: "/integration/empty-payload",
			headers: map[string]string{
				"X-Token": "integration-test",
			},
			payload: map[string]any{},
			expectedResponse: expectedResponse{
				statusCode: 204,
			},
			expectedStorage: expectedStorage{
				storageType: StorageTypeRedis,
				key:         "empty-payload:events",
				data:        `{}`,
			},
		},
		{
			name:     "basic-usage",
			endpoint: "/integration/basic-usage",
			headers: map[string]string{
				"X-Token": "integration-test",
			},
			payload: map[string]string{
				"key": "value",
			},
			expectedResponse: expectedResponse{
				statusCode: 204,
			},
			expectedStorage: expectedStorage{
				storageType: StorageTypeRedis,
				key:         "integration:basic-usage",
				data:        `{"key":"value"}`,
			},
		},
		{
			name:     "basic-formatted-usage",
			endpoint: "/integration/basic-formatted-usage",
			headers: map[string]string{
				"X-Token":      "integration-test",
				"Content-Type": "application/json",
			},
			payload: map[string]string{
				"key": "value",
			},
			expectedResponse: expectedResponse{
				statusCode: 200,
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body: `{"status": "OK", "data": {"key":"value"}}`,
			},
			expectedStorage: expectedStorage{
				storageType: StorageTypeRedis,
				key:         "integration:basic-formatted-usage",
				data:        `redis:application/json|{"key":"value"}`,
			},
		},
		{
			name:     "advanced-formatted-usage",
			endpoint: "/integration/advanced-formatted-usage",
			headers: map[string]string{
				"X-Token":      "integration-test",
				"Content-Type": "application/json",
			},
			payload: map[string]any{
				"id":   12345,
				"name": "John Doe",
				"childrens": []map[string]any{
					{
						"name": "Jane",
						"age":  5,
					},
					{
						"name": "Bob",
						"age":  8,
					},
				},
				"pets": []string{},
				"favoriteColors": map[string]any{
					"primary":   nil,
					"secondary": "blue",
				},
				"lastLogin": "2023-06-28T18:30:00Z",
				"notes":     "I miss Gab so much",
			},
			expectedStorage: expectedStorage{
				storageType: StorageTypeRedis,
				key:         "integration:advanced-formatted-usage",
				data:        "redis:12345|John Doe|hasNotes:true|hasPets:false|hasChildrens:true|childrensCount:2",
			},
			expectedResponse: expectedResponse{
				statusCode: 200,
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body: `{
  "user": {
    "id": 12345,
    "name": "John Doe"
  },
  "hasNotes": true,
  "hasChildrens": true,
  "hasPets": false,
  "favoriteColor": "blue",
  "childrenNames": ["Jane","Bob"]
}`,
			},
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			suite.runTest(test)
		})
	}
}

func (suite *IntegrationTestSuite) doRequest(test testInput) {
	// Prepare request
	jsonValue, err := json.Marshal(test.payload)
	suite.NoError(err, "Failed to marshal payload")

	req, err := http.NewRequestWithContext(suite.ctx, "POST", BaseURL+test.endpoint, bytes.NewBuffer(jsonValue))
	suite.NoError(err, "Failed to create request")

	req.Header.Set("Content-Type", "application/json")
	for key, value := range test.headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(suite.T(), err, "Failed to send request")

	// Check response status code
	suite.Equal(test.expectedResponse.statusCode, resp.StatusCode, "Unexpected status code")

	// Check headers
	for key, expectedValue := range test.expectedResponse.headers {
		suite.Equal(expectedValue, resp.Header.Get(key), "Header mismatch for %s", key)
	}

	// Check response body
	if test.expectedResponse.body != "" {
		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(resp.Body)
		suite.NoError(err, "Failed to read response body")

		body := buf.String()
		suite.Equal(test.expectedResponse.body, strings.Trim(body, "\n"), "Response body mismatch")
	}

	_ = resp.Body.Close()
	time.Sleep(100 * time.Millisecond) // Allow some time for async processing
}

func (suite *IntegrationTestSuite) runTest(test testInput) {
	suite.doRequest(test)

	// Check storage
	if test.expectedStorage.storageType != "" {
		storage, exists := suite.storages[test.expectedStorage.storageType]
		suite.True(exists, "Storage type %s not found", test.expectedStorage.storageType)

		switch test.expectedStorage.storageType {
		case StorageTypeRedis:
			redisClient := storage.(*redis.Client)
			data, err := redisClient.LPop(suite.ctx, test.expectedStorage.key).Result()
			if err != redis.Nil {
				suite.NoError(err, "Failed to get data from Redis")
			}

			if test.expectedStorage.isJson {
				suite.JSONEq(test.expectedStorage.data, data, "Data mismatch in Redis storage")
			} else {
				suite.Equal(test.expectedStorage.data, data, "Data mismatch in Redis storage")
			}
		default:
		}
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

//go:build integrations

package integration_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ErrorScenariosIntegrationTestSuite struct {
	IntegrationTestSuite
}

func (suite *ErrorScenariosIntegrationTestSuite) TestErrorScenarios() {
	largePayload := suite.generateLargePayload(1024 * 100) // 100KB payload
	largePayloadJSON, err := json.Marshal(largePayload)

	require.NoError(suite.T(), err, "Failed to marshal large payload")

	tests := []testInput{
		{
			name:     "invalid-endpoint",
			endpoint: "/integration/non-existent",
			headers: map[string]string{
				"X-Token": "integration-test",
			},
			payload: map[string]any{
				"test": "data",
			},
			expectedResponse: expectedResponse{
				statusCode: 404,
			},
		},
		{
			// Must not return 400 error to don't lose data
			name:     "invalid-json-payload",
			endpoint: "/integration/invalid-json-payload",
			headers: map[string]string{
				"X-Token":      "integration-test",
				"Content-Type": "application/json",
			},
			payload: "invalid json{",
			expectedResponse: expectedResponse{
				statusCode: 204,
			},
		},
		{
			name:     "missing-required-header",
			endpoint: "/integration/basic-usage",
			headers:  map[string]string{}, // No X-Token header
			payload: map[string]any{
				"test": "should fail",
			},
			expectedResponse: expectedResponse{
				statusCode: 401,
			},
		},
		{
			name:     "large-payload-handling",
			endpoint: "/integration/large-payload",
			headers: map[string]string{
				"X-Token": "integration-test",
			},
			payload: largePayload,
			expectedResponse: expectedResponse{
				statusCode: 204,
			},
			expectedStorage: expectedStorage{
				storageType: StorageTypeRedis,
				key:         "integration:large-payload-events",
				data:        string(largePayloadJSON),
			},
		},
		{
			name:     "webhook-spec-not-found",
			endpoint: "/integration/missing-webhook",
			headers: map[string]string{
				"X-Token": "integration-test",
			},
			payload: map[string]any{
				"test": "should return 404",
			},
			expectedResponse: expectedResponse{
				statusCode: 404,
			},
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			suite.runErrorTest(test)
		})
	}
}

func (suite *ErrorScenariosIntegrationTestSuite) runErrorTest(test testInput) {
	if test.name == "concurrent-requests-same-webhook" {
		suite.runConcurrencyTest(test)
	} else {
		suite.runTest(test)
	}
}

func (suite *ErrorScenariosIntegrationTestSuite) runConcurrencyTest(test testInput) {
	// Send multiple concurrent requests
	concurrency := 10
	results := make(chan int, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			testCopy := test
			testCopy.headers["X-Request-ID"] = fmt.Sprintf("concurrent-%d", id)
			testCopy.payload = map[string]any{
				"request_id": id,
				"data":       "concurrent test",
			}

			suite.runTest(testCopy)
			results <- 1
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < concurrency; i++ {
		<-results
	}
}

func (suite *ErrorScenariosIntegrationTestSuite) generateLargePayload(size int) map[string]any {
	largeString := make([]byte, size)
	for i := range largeString {
		largeString[i] = 'A' + byte(i%26)
	}

	return map[string]any{
		"large_data": string(largeString),
		"metadata": map[string]any{
			"size":      size,
			"timestamp": "2023-06-28T18:30:00Z",
		},
	}
}

func TestErrorScenariosIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(ErrorScenariosIntegrationTestSuite))
}

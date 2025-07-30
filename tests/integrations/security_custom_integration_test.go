//go:build integration

package integration_test

func (suite *SecurityIntegrationTestSuite) TestSecurityCustomScenarios() {
	tests := []testInput{
		{
			name:     "custom-security-valid",
			endpoint: "/integration/custom-security",
			headers: map[string]string{
				"Authorization": "Bearer valid-token-123",
				"X-API-Key":     "secret-api-key",
			},
			payload: map[string]any{
				"data": "protected content",
				"type": "secure_event",
			},
			expectedResponse: expectedResponse{
				statusCode: 204,
			},
			expectedStorage: expectedStorage{
				storageType: StorageTypeRedis,
				key:         "custom-security:events",
				data:        `{"data":"protected content","type":"secure_event"}`,
			},
		},
		{
			name:     "custom-security-invalid",
			endpoint: "/integration/custom-security",
			headers: map[string]string{
				"Authorization": "Bearer invalid-token",
				"X-API-Key":     "wrong-key",
			},
			payload: map[string]any{
				"data": "should not be stored",
			},
			expectedResponse: expectedResponse{
				statusCode: 401,
				body:       "",
			},
			expectedStorage: expectedStorage{
				storageType: StorageTypeRedis,
				key:         "custom-security:events",
				data:        "",
			},
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			suite.runTest(test)
		})
	}
}

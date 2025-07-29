package integration_test

func (suite *SecurityIntegrationTestSuite) TestSecurityNoopScenarios() {
	tests := []testInput{
		{
			name:     "no-security-webhook",
			endpoint: "/integration/no-security",
			headers:  map[string]string{},
			payload: map[string]any{
				"message": "this webhook has no security",
			},
			expectedResponse: expectedResponse{
				statusCode: 204,
			},
			expectedStorage: expectedStorage{
				storageType: StorageTypeRedis,
				key:         "no-security:events",
				data:        `{"message":"this webhook has no security"}`,
			},
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			suite.runTest(test)
		})
	}
}

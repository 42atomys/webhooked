//go:build integration

package integration_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func (suite *SecurityIntegrationTestSuite) TestSecurityGithubScenarios() {
	tests := []testInput{
		{
			name:     "github-webhook-valid-signature",
			endpoint: "/integration/github-webhook",
			headers: map[string]string{
				"X-Hub-Signature-256": generateGitHubSignature(`{"action":"push","ref":"refs/heads/main"}`, "github-secret"),
				"X-GitHub-Event":      "push",
			},
			payload: map[string]any{
				"action": "push",
				"ref":    "refs/heads/main",
			},
			expectedResponse: expectedResponse{
				statusCode: 204,
			},
			expectedStorage: expectedStorage{
				storageType: StorageTypeRedis,
				key:         "github:events",
				data:        `{"action":"push","ref":"refs/heads/main"}`,
			},
		},
		{
			name:     "github-webhook-invalid-signature",
			endpoint: "/integration/github-webhook",
			headers: map[string]string{
				"X-Hub-Signature-256": "sha256=invalid_signature",
				"X-GitHub-Event":      "push",
			},
			payload: map[string]any{
				"action": "push",
				"ref":    "refs/heads/main",
			},
			expectedResponse: expectedResponse{
				statusCode: 401,
			},
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			suite.runTest(test)
		})
	}
}

func generateGitHubSignature(payload, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	return "sha256=" + hex.EncodeToString(h.Sum(nil))
}

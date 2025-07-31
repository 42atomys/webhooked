//go:build unit

package config

import (
	"os"
	"testing"

	"github.com/42atomys/webhooked/security"
	securityNoop "github.com/42atomys/webhooked/security/noop"
	"github.com/42atomys/webhooked/storage"
	storageNoop "github.com/42atomys/webhooked/storage/noop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TestSuiteConfig struct {
	suite.Suite

	validConfig      *Config
	invalidAPIConfig *Config
	invalidKindConfig *Config
	validConfigFile  string
	invalidConfigFile string
	tempConfigPath   string
}

func (suite *TestSuiteConfig) BeforeTest(suiteName, testName string) {
	suite.validConfig = &Config{
		APIVersion: APIVersionV1Alpha2,
		Kind:       KindConfiguration,
		Metadata: Metadata{
			Name: "test-config",
		},
		Specs: []*Spec{
			{
				MetricsEnabled: true,
				Webhooks: []*Webhook{
					{
						Name:          "test-webhook",
						EntrypointURL: "/test",
						Security: security.Security{
							Type:  "noop",
							Specs: &securityNoop.NoopSecuritySpec{},
						},
						Storage: []*storage.Storage{
							{
								Type:  "noop",
								Specs: &storageNoop.NoopStorageSpec{},
							},
						},
					},
				},
			},
		},
	}

	suite.invalidAPIConfig = &Config{
		APIVersion: "v1beta1", // Invalid API version
		Kind:       KindConfiguration,
		Specs:      []*Spec{},
	}

	suite.invalidKindConfig = &Config{
		APIVersion: APIVersionV1Alpha2,
		Kind:       "InvalidKind", // Invalid kind
		Specs:      []*Spec{},
	}

	// Create temporary config files for testing
	suite.tempConfigPath = "/tmp/webhooked_test_config.yaml"
	suite.validConfigFile = `
apiVersion: v1alpha2
kind: Configuration
metadata:
  name: test-config
specs:
  - metricsEnabled: true
    webhooks:
      - name: test-webhook
        entrypointUrl: /test
        security:
          type: noop
        storage:
          - type: noop
`

	suite.invalidConfigFile = `
invalid_yaml: [
  missing_bracket
`
}

func (suite *TestSuiteConfig) AfterTest(suiteName, testName string) {
	// Clean up temporary files
	os.Remove(suite.tempConfigPath)
}

func (suite *TestSuiteConfig) TestConfigValidate_Success() {
	assert := assert.New(suite.T())

	err := suite.validConfig.Validate()

	assert.NoError(err)
}

func (suite *TestSuiteConfig) TestConfigValidate_InvalidAPIVersion() {
	assert := assert.New(suite.T())

	err := suite.invalidAPIConfig.Validate()

	assert.Error(err)
	assert.Contains(err.Error(), "unsupported API version")
}

func (suite *TestSuiteConfig) TestConfigValidate_InvalidKind() {
	assert := assert.New(suite.T())

	err := suite.invalidKindConfig.Validate()

	assert.Error(err)
	assert.Contains(err.Error(), "invalid kind, expected 'Configuration'")
}

func (suite *TestSuiteConfig) TestConfigValidate_EmptyEntrypointURL() {
	assert := assert.New(suite.T())

	config := &Config{
		APIVersion: APIVersionV1Alpha2,
		Kind:       KindConfiguration,
		Specs: []*Spec{
			{
				Webhooks: []*Webhook{
					{
						Name:          "test-webhook",
						EntrypointURL: "", // Empty entrypoint URL
						Security: security.Security{
							Type:  "noop",
							Specs: &securityNoop.NoopSecuritySpec{},
						},
					},
				},
			},
		},
	}

	err := config.Validate()

	assert.Error(err)
	assert.Contains(err.Error(), "webhook entrypoint URL cannot be empty")
}

func (suite *TestSuiteConfig) TestConfigValidate_WebhookValidationError() {
	assert := assert.New(suite.T())

	config := &Config{
		APIVersion: APIVersionV1Alpha2,
		Kind:       KindConfiguration,
		Specs: []*Spec{
			{
				Webhooks: []*Webhook{
					{
						Name:          "test-webhook",
						EntrypointURL: "/test",
						Response: Response{
							StatusCode: 999, // Invalid status code
						},
						Security: security.Security{
							Type:  "noop",
							Specs: &securityNoop.NoopSecuritySpec{},
						},
					},
				},
			},
		},
	}

	err := config.Validate()

	assert.Error(err)
	assert.Contains(err.Error(), "error validating webhook test-webhook")
}

func (suite *TestSuiteConfig) TestFetchWebhookByPath_Success() {
	assert := assert.New(suite.T())

	// Path format: /webhooks/v1alpha2/test
	path := []byte("/webhooks/v1alpha2/test")

	webhook, err := suite.validConfig.FetchWebhookByPath(path)

	assert.NoError(err)
	assert.NotNil(webhook)
	assert.Equal("test-webhook", webhook.Name)
	assert.Equal("/test", webhook.EntrypointURL)
}

func (suite *TestSuiteConfig) TestFetchWebhookByPath_PathTooShort() {
	assert := assert.New(suite.T())

	// Path too short
	path := []byte("/webhooks")

	webhook, err := suite.validConfig.FetchWebhookByPath(path)

	assert.Error(err)
	assert.ErrorIs(err, ErrSpecNotFound)
	assert.Nil(webhook)
}

func (suite *TestSuiteConfig) TestFetchWebhookByPath_WebhookNotFound() {
	assert := assert.New(suite.T())

	// Non-existent webhook path
	path := []byte("/webhooks/v1alpha2/nonexistent")

	webhook, err := suite.validConfig.FetchWebhookByPath(path)

	assert.Error(err)
	assert.ErrorIs(err, ErrSpecNotFound)
	assert.Nil(webhook)
}

func (suite *TestSuiteConfig) TestFetchWebhookByPath_MultipleSpecs() {
	assert := assert.New(suite.T())

	// Create config with multiple specs and webhooks
	config := &Config{
		APIVersion: APIVersionV1Alpha2,
		Kind:       KindConfiguration,
		Specs: []*Spec{
			{
				Webhooks: []*Webhook{
					{
						Name:          "webhook1",
						EntrypointURL: "/webhook1",
					},
				},
			},
			{
				Webhooks: []*Webhook{
					{
						Name:          "webhook2",
						EntrypointURL: "/webhook2",
					},
				},
			},
		},
	}

	// Test finding webhook from first spec
	path1 := []byte("/webhooks/v1alpha2/webhook1")
	webhook1, err1 := config.FetchWebhookByPath(path1)

	assert.NoError(err1)
	assert.NotNil(webhook1)
	assert.Equal("webhook1", webhook1.Name)

	// Test finding webhook from second spec
	path2 := []byte("/webhooks/v1alpha2/webhook2")
	webhook2, err2 := config.FetchWebhookByPath(path2)

	assert.NoError(err2)
	assert.NotNil(webhook2)
	assert.Equal("webhook2", webhook2.Name)
}

func (suite *TestSuiteConfig) TestWebhooksEndpointPrefix() {
	assert := assert.New(suite.T())

	prefix := WebhooksEndpointPrefix()

	assert.Equal([]byte("/webhooks"), prefix)
}

func (suite *TestSuiteConfig) TestWebhookTemplateContext() {
	assert := assert.New(suite.T())

	webhook := &Webhook{
		Name:          "test-webhook",
		EntrypointURL: "/test",
	}

	context := webhook.TemplateContext()

	assert.NotNil(context)
	assert.Contains(context, "SpecName")
	assert.Contains(context, "SpecEntrypointURL")
	assert.Equal("test-webhook", context["SpecName"])
	assert.Equal("/test", context["SpecEntrypointURL"])
}

func (suite *TestSuiteConfig) TestLoad_ValidConfig() {
	assert := assert.New(suite.T())

	// Write valid config to temporary file
	err := os.WriteFile(suite.tempConfigPath, []byte(suite.validConfigFile), 0644)
	require.NoError(suite.T(), err)

	config, err := Load(suite.tempConfigPath)

	assert.NoError(err)
	assert.NotNil(config)
	assert.Equal(APIVersionV1Alpha2, config.APIVersion)
	assert.Equal(KindConfiguration, config.Kind)
	assert.Equal("test-config", config.Metadata.Name)
	assert.Len(config.Specs, 1)
	assert.Len(config.Specs[0].Webhooks, 1)
	assert.Equal("test-webhook", config.Specs[0].Webhooks[0].Name)
}

func (suite *TestSuiteConfig) TestLoad_InvalidYAML() {
	assert := assert.New(suite.T())

	// Write invalid YAML to temporary file
	err := os.WriteFile(suite.tempConfigPath, []byte(suite.invalidConfigFile), 0644)
	require.NoError(suite.T(), err)

	_, err = Load(suite.tempConfigPath)

	// Should return error for invalid YAML
	assert.Error(err)
}

func (suite *TestSuiteConfig) TestLoad_NonexistentFile() {
	assert := assert.New(suite.T())

	_, err := Load("/nonexistent/path/to/config.yaml")

	assert.Error(err)
}

func (suite *TestSuiteConfig) TestLoad_WithEnvironmentVariables() {
	assert := assert.New(suite.T())

	// Set environment variable
	originalDebug := os.Getenv("WH_DEBUG")
	defer os.Setenv("WH_DEBUG", originalDebug)
	
	os.Setenv("WH_DEBUG", "true")

	// Write minimal config to temporary file
	minimalConfig := `
apiVersion: v1alpha2
kind: Configuration
specs: []
`
	err := os.WriteFile(suite.tempConfigPath, []byte(minimalConfig), 0644)
	require.NoError(suite.T(), err)

	config, err := Load(suite.tempConfigPath)

	assert.NoError(err)
	assert.NotNil(config)
}

func (suite *TestSuiteConfig) TestConstants() {
	assert := assert.New(suite.T())

	// Test constants are correctly defined
	assert.Equal(APIVersion("v1alpha2"), APIVersionV1Alpha2)
	assert.Equal(Kind("Configuration"), KindConfiguration)

	// Test error variables
	assert.NotNil(ErrSpecNotFound)
	assert.NotNil(ErrInvalidStatusCode)
	assert.Equal("spec not found", ErrSpecNotFound.Error())
	assert.Equal("invalid status code", ErrInvalidStatusCode.Error())

	// Test template constants
	assert.Equal([]byte(`{{ .Payload }}`), defaultPayloadTemplate)
	assert.Equal([]byte(``), defaultResponseTemplate)
	assert.Equal([]byte("/webhooks"), webhooksPrefix)
}

func TestRunConfigSuite(t *testing.T) {
	suite.Run(t, new(TestSuiteConfig))
}

// Benchmarks

func BenchmarkConfigValidate(b *testing.B) {
	config := &Config{
		APIVersion: APIVersionV1Alpha2,
		Kind:       KindConfiguration,
		Specs: []*Spec{
			{
				Webhooks: []*Webhook{
					{
						Name:          "benchmark-webhook",
						EntrypointURL: "/benchmark",
						Security: security.Security{
							Type:  "noop",
							Specs: &securityNoop.NoopSecuritySpec{},
						},
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.Validate() // nolint:errcheck
	}
}

func BenchmarkFetchWebhookByPath(b *testing.B) {
	config := &Config{
		APIVersion: APIVersionV1Alpha2,
		Kind:       KindConfiguration,
		Specs: []*Spec{
			{
				Webhooks: []*Webhook{
					{
						Name:          "benchmark-webhook",
						EntrypointURL: "/benchmark",
					},
				},
			},
		},
	}

	path := []byte("/webhooks/v1alpha2/benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.FetchWebhookByPath(path) // nolint:errcheck
	}
}

func BenchmarkWebhookTemplateContext(b *testing.B) {
	webhook := &Webhook{
		Name:          "benchmark-webhook",
		EntrypointURL: "/benchmark",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		webhook.TemplateContext()
	}
}

func BenchmarkFetchWebhookByPath_MultipleWebhooks(b *testing.B) {
	// Create config with many webhooks to test search performance
	webhooks := make([]*Webhook, 100)
	for i := 0; i < 100; i++ {
		webhooks[i] = &Webhook{
			Name:          "webhook-" + string(rune(i)),
			EntrypointURL: "/webhook-" + string(rune(i)),
		}
	}

	config := &Config{
		APIVersion: APIVersionV1Alpha2,
		Kind:       KindConfiguration,
		Specs: []*Spec{
			{Webhooks: webhooks},
		},
	}

	// Search for last webhook (worst case)
	path := []byte("/webhooks/v1alpha2/webhook-c") // webhook-99

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.FetchWebhookByPath(path) // nolint:errcheck
	}
}
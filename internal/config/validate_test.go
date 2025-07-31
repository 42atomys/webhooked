//go:build unit

package config

import (
	"context"
	"testing"

	"github.com/42atomys/webhooked/format"
	"github.com/42atomys/webhooked/internal/fasthttpz"
	"github.com/42atomys/webhooked/security"
	securityNoop "github.com/42atomys/webhooked/security/noop"
	"github.com/42atomys/webhooked/storage"
	storageNoop "github.com/42atomys/webhooked/storage/noop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TestSuiteConfigValidate struct {
	suite.Suite

	validWebhook    *Webhook
	minimalWebhook  *Webhook
	invalidWebhook  *Webhook
	testFormatting  *format.Formatting
}

func (suite *TestSuiteConfigValidate) BeforeTest(suiteName, testName string) {
	var err error
	suite.testFormatting, err = format.New(format.Specs{TemplateString: "test template"})
	require.NoError(suite.T(), err)

	suite.validWebhook = &Webhook{
		Name:          "test-webhook",
		EntrypointURL: "/test",
		Response: Response{
			ContentType: "application/json",
			StatusCode:  200,
			Formatting:  suite.testFormatting,
		},
		Security: security.Security{
			Type:  "noop",
			Specs: &securityNoop.NoopSecuritySpec{},
		},
		Storage: []*storage.Storage{
			{
				Type:       "noop",
				Formatting: suite.testFormatting,
				Specs:      &storageNoop.NoopStorageSpec{},
			},
		},
	}

	suite.minimalWebhook = &Webhook{
		Name:          "minimal-webhook",
		EntrypointURL: "/minimal",
		Response:      Response{}, // Empty response to test defaults
		Security:      security.Security{}, // Empty security to test defaults
		Storage:       []*storage.Storage{},
	}

	suite.invalidWebhook = &Webhook{
		Name:          "invalid-webhook",
		EntrypointURL: "/invalid",
		Response: Response{
			StatusCode: 999, // Invalid status code
		},
		Security: security.Security{
			Type:  "noop",
			Specs: &securityNoop.NoopSecuritySpec{},
		},
		Storage: []*storage.Storage{},
	}
}

func (suite *TestSuiteConfigValidate) TestValidateAndSetDefaults_Success() {
	assert := assert.New(suite.T())

	err := validateAndSetDefaults(suite.validWebhook)

	assert.NoError(err)
	assert.Equal("application/json", suite.validWebhook.Response.ContentType)
	assert.Equal(200, suite.validWebhook.Response.StatusCode)
	assert.NotNil(suite.validWebhook.Response.Formatting)
}

func (suite *TestSuiteConfigValidate) TestValidateAndSetDefaults_MinimalWebhook() {
	assert := assert.New(suite.T())

	err := validateAndSetDefaults(suite.minimalWebhook)

	assert.NoError(err)
	// Check that defaults were set
	assert.Equal("application/json", suite.minimalWebhook.Response.ContentType)
	assert.Equal(200, suite.minimalWebhook.Response.StatusCode)
	assert.NotNil(suite.minimalWebhook.Response.Formatting)
	assert.Equal("noop", suite.minimalWebhook.Security.Type)
	assert.NotNil(suite.minimalWebhook.Security.Specs)
}

func (suite *TestSuiteConfigValidate) TestValidateAndSetDefaults_InvalidStatusCode() {
	assert := assert.New(suite.T())

	err := validateAndSetDefaults(suite.invalidWebhook)

	assert.Error(err)
	assert.Contains(err.Error(), "error validating webhook invalid-webhook")
	assert.Contains(err.Error(), "invalid status code")
}

func (suite *TestSuiteConfigValidate) TestEnsureResponseCompleteness_DefaultValues() {
	assert := assert.New(suite.T())

	webhook := &Webhook{
		Name:     "test",
		Response: Response{}, // Empty response
	}

	err := ensureResponseCompleteness(webhook)

	assert.NoError(err)
	assert.Equal("application/json", webhook.Response.ContentType)
	assert.Equal(200, webhook.Response.StatusCode)
	assert.NotNil(webhook.Response.Formatting)
	// Default response template is empty, so HasTemplate returns false even after WithTemplate("")
	assert.False(webhook.Response.Formatting.HasTemplate())
}

func (suite *TestSuiteConfigValidate) TestEnsureResponseCompleteness_ValidStatusCodes() {
	assert := assert.New(suite.T())

	testCases := []struct {
		name       string
		statusCode int
		shouldPass bool
	}{
		{"Valid 200", 200, true},
		{"Valid 201", 201, true},
		{"Valid 400", 400, true},
		{"Valid 500", 500, true},
		{"Valid 100", 100, true},
		{"Valid 599", 599, true},
		{"Invalid 99", 99, false},
		{"Invalid 600", 600, false},
		{"Invalid 0", 0, true}, // 0 gets set to default 200
	}

	for _, tc := range testCases {
		webhook := &Webhook{
			Name: "test",
			Response: Response{
				StatusCode: tc.statusCode,
			},
		}

		err := ensureResponseCompleteness(webhook)

		if tc.shouldPass {
			assert.NoError(err, "Test case: %s", tc.name)
			if tc.statusCode == 0 {
				assert.Equal(200, webhook.Response.StatusCode, "Default should be 200")
			} else {
				assert.Equal(tc.statusCode, webhook.Response.StatusCode)
			}
		} else {
			assert.Error(err, "Test case: %s should fail", tc.name)
			assert.ErrorIs(err, ErrInvalidStatusCode)
		}
	}
}

func (suite *TestSuiteConfigValidate) TestEnsureResponseCompleteness_FormattingSetup() {
	assert := assert.New(suite.T())

	tests := []struct {
		name               string
		initialFormatting  *format.Formatting
		expectedHasTemplate bool
	}{
		{
			name:               "nil formatting gets initialized",
			initialFormatting:  nil,
			expectedHasTemplate: false, // defaultResponseTemplate is empty
		},
		{
			name:               "formatting without template gets template",
			initialFormatting:  &format.Formatting{},
			expectedHasTemplate: false, // defaultResponseTemplate is empty
		},
		{
			name:               "formatting with template remains unchanged",
			initialFormatting:  suite.testFormatting,
			expectedHasTemplate: true,
		},
	}

	for _, test := range tests {
		webhook := &Webhook{
			Name: "test",
			Response: Response{
				Formatting: test.initialFormatting,
			},
		}

		err := ensureResponseCompleteness(webhook)

		assert.NoError(err, "Test case: %s", test.name)
		assert.NotNil(webhook.Response.Formatting, "Formatting should not be nil for: %s", test.name)
		assert.Equal(test.expectedHasTemplate, webhook.Response.Formatting.HasTemplate(), "Test case: %s", test.name)
	}
}

func (suite *TestSuiteConfigValidate) TestEnsureSecurityCompleteness_DefaultNoop() {
	assert := assert.New(suite.T())

	webhook := &Webhook{
		Name:     "test",
		Security: security.Security{}, // Empty security
	}

	err := ensureSecurityCompleteness(webhook)

	assert.NoError(err)
	assert.Equal("noop", webhook.Security.Type)
	assert.NotNil(webhook.Security.Specs)
	assert.IsType(&securityNoop.NoopSecuritySpec{}, webhook.Security.Specs)
}

func (suite *TestSuiteConfigValidate) TestEnsureSecurityCompleteness_ExistingSecurity() {
	assert := assert.New(suite.T())

	webhook := &Webhook{
		Name: "test",
		Security: security.Security{
			Type:  "noop",
			Specs: &securityNoop.NoopSecuritySpec{},
		},
	}

	err := ensureSecurityCompleteness(webhook)

	assert.NoError(err)
	assert.Equal("noop", webhook.Security.Type)
	assert.NotNil(webhook.Security.Specs)
}

func (suite *TestSuiteConfigValidate) TestEnsureSecurityCompleteness_SecurityError() {
	assert := assert.New(suite.T())

	// Create a mock security spec that will fail validation
	mockSecurity := &mockFailingSecuritySpec{}
	webhook := &Webhook{
		Name: "test",
		Security: security.Security{
			Type:  "failing",
			Specs: mockSecurity,
		},
	}

	err := ensureSecurityCompleteness(webhook)

	assert.Error(err)
	assert.Contains(err.Error(), "error validating security failing")
}

func (suite *TestSuiteConfigValidate) TestEnsureStorageCompleteness_Success() {
	assert := assert.New(suite.T())

	webhook := &Webhook{
		Name: "test",
		Storage: []*storage.Storage{
			{
				Type:       "noop",
				Formatting: &format.Formatting{}, // Formatting without template
				Specs:      &storageNoop.NoopStorageSpec{},
			},
		},
	}

	err := ensureStorageCompleteness(webhook)

	assert.NoError(err)
	// HasTemplateCompiled will still be false because WithTemplate doesn't compile
	// The template will be compiled later when Format() is called or during actual usage
	assert.False(webhook.Storage[0].Formatting.HasTemplateCompiled())
	// But it should have a template string set now (defaultPayloadTemplate = "{{ .Payload }}")
	assert.True(webhook.Storage[0].Formatting.HasTemplate())
}

func (suite *TestSuiteConfigValidate) TestEnsureStorageCompleteness_MultipleStorages() {
	assert := assert.New(suite.T())

	webhook := &Webhook{
		Name: "test",
		Storage: []*storage.Storage{
			{
				Type:       "noop",
				Formatting: &format.Formatting{},
				Specs:      &storageNoop.NoopStorageSpec{},
			},
			{
				Type:       "noop",
				Formatting: &format.Formatting{},
				Specs:      &storageNoop.NoopStorageSpec{},
			},
		},
	}

	err := ensureStorageCompleteness(webhook)

	assert.NoError(err)
	for i, stor := range webhook.Storage {
		// HasTemplateCompiled will be false because WithTemplate doesn't compile
		assert.False(stor.Formatting.HasTemplateCompiled(), "Storage %d template not compiled yet", i)
		// But template string should be set
		assert.True(stor.Formatting.HasTemplate(), "Storage %d should have template string", i)
	}
}

func (suite *TestSuiteConfigValidate) TestEnsureStorageCompleteness_StorageError() {
	assert := assert.New(suite.T())

	// Create a mock storage spec that will fail validation
	mockStorage := &mockFailingStorageSpec{}
	webhook := &Webhook{
		Name: "test",
		Storage: []*storage.Storage{
			{
				Type:       "failing",
				Formatting: &format.Formatting{},
				Specs:      mockStorage,
			},
		},
	}

	err := ensureStorageCompleteness(webhook)

	assert.Error(err)
	assert.Contains(err.Error(), "error validating storage failing")
}

func TestRunConfigValidateSuite(t *testing.T) {
	suite.Run(t, new(TestSuiteConfigValidate))
}

// Mock implementations for testing error scenarios

type mockFailingSecuritySpec struct{}

func (m *mockFailingSecuritySpec) EnsureConfigurationCompleteness() error {
	return assert.AnError
}

func (m *mockFailingSecuritySpec) Initialize() error {
	return nil
}

func (m *mockFailingSecuritySpec) IsSecure(ctx context.Context, rctx *fasthttpz.RequestCtx) (bool, error) {
	return false, nil
}

type mockFailingStorageSpec struct{}

func (m *mockFailingStorageSpec) EnsureConfigurationCompleteness() error {
	return assert.AnError
}

func (m *mockFailingStorageSpec) Initialize() error {
	return nil
}

func (m *mockFailingStorageSpec) Store(ctx context.Context, data []byte) error {
	return nil
}

// Benchmarks

func BenchmarkValidateAndSetDefaults(b *testing.B) {
	webhook := &Webhook{
		Name:          "benchmark-webhook",
		EntrypointURL: "/benchmark",
		Response:      Response{},
		Security:      security.Security{},
		Storage:       []*storage.Storage{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset webhook state for each iteration
		webhook.Response = Response{}
		webhook.Security = security.Security{}
		validateAndSetDefaults(webhook) // nolint:errcheck
	}
}

func BenchmarkEnsureResponseCompleteness(b *testing.B) {
	webhook := &Webhook{
		Name:     "benchmark",
		Response: Response{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		webhook.Response = Response{} // Reset for each iteration
		ensureResponseCompleteness(webhook) // nolint:errcheck
	}
}

func BenchmarkEnsureSecurityCompleteness(b *testing.B) {
	webhook := &Webhook{
		Name:     "benchmark",
		Security: security.Security{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		webhook.Security = security.Security{} // Reset for each iteration
		ensureSecurityCompleteness(webhook) // nolint:errcheck
	}
}

func BenchmarkEnsureStorageCompleteness(b *testing.B) {
	webhook := &Webhook{
		Name: "benchmark",
		Storage: []*storage.Storage{
			{
				Type:       "noop",
				Formatting: &format.Formatting{},
				Specs:      &storageNoop.NoopStorageSpec{},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		webhook.Storage[0].Formatting = &format.Formatting{} // Reset for each iteration
		ensureStorageCompleteness(webhook) // nolint:errcheck
	}
}
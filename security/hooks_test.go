//go:build unit

package security

import (
	"reflect"
	"testing"

	"github.com/42atomys/webhooked/security/custom"
	"github.com/42atomys/webhooked/security/github"
	"github.com/42atomys/webhooked/security/noop"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestSuiteSecurityHooks struct {
	suite.Suite

	validNoopData   map[string]any
	validGitHubData map[string]any
	validCustomData map[string]any
	invalidTypeData map[string]any
	invalidSpecData map[string]any
	wrongTypeData   any
}

func (suite *TestSuiteSecurityHooks) BeforeTest(suiteName, testName string) {
	suite.validNoopData = map[string]any{
		"type":  "noop",
		"specs": map[string]any{},
	}

	suite.validGitHubData = map[string]any{
		"type": "github",
		"specs": map[string]any{
			"secretToken": "test-secret",
		},
	}

	suite.validCustomData = map[string]any{
		"type": "custom",
		"specs": map[string]any{
			"headerName":  "X-Custom-Secret",
			"secretToken": "test-token",
		},
	}

	suite.invalidTypeData = map[string]any{
		"type":  123, // Non-string type
		"specs": map[string]any{},
	}

	suite.invalidSpecData = map[string]any{
		"type":  "unknown-security-type",
		"specs": map[string]any{},
	}

	suite.wrongTypeData = "not-a-map"
}

func (suite *TestSuiteSecurityHooks) TestDecodeHook_ValidNoopSecurity() {
	assert := assert.New(suite.T())

	fromType := reflect.TypeOf(suite.validNoopData)
	toType := reflect.TypeOf(Security{})

	result, err := DecodeHook(fromType, toType, suite.validNoopData)

	assert.NoError(err)
	assert.IsType(Security{}, result)

	security := result.(Security)
	assert.Equal("noop", security.Type)
	assert.IsType(&noop.NoopSecuritySpec{}, security.Specs)
}

func (suite *TestSuiteSecurityHooks) TestDecodeHook_ValidGitHubSecurity() {
	assert := assert.New(suite.T())

	fromType := reflect.TypeOf(suite.validGitHubData)
	toType := reflect.TypeOf(Security{})

	result, err := DecodeHook(fromType, toType, suite.validGitHubData)

	assert.NoError(err)
	assert.IsType(Security{}, result)

	security := result.(Security)
	assert.Equal("github", security.Type)
	assert.IsType(&github.GitHubSecuritySpec{}, security.Specs)
}

func (suite *TestSuiteSecurityHooks) TestDecodeHook_ValidCustomSecurity() {
	assert := assert.New(suite.T())

	fromType := reflect.TypeOf(suite.validCustomData)
	toType := reflect.TypeOf(Security{})

	result, err := DecodeHook(fromType, toType, suite.validCustomData)

	assert.NoError(err)
	assert.IsType(Security{}, result)

	security := result.(Security)
	assert.Equal("custom", security.Type)
	assert.IsType(&custom.CustomSecuritySpec{}, security.Specs)
}

func (suite *TestSuiteSecurityHooks) TestDecodeHook_WrongFromType() {
	assert := assert.New(suite.T())

	// Test with string instead of map
	fromType := reflect.TypeOf("string")
	toType := reflect.TypeOf(Security{})

	result, err := DecodeHook(fromType, toType, "test-data")

	assert.NoError(err)
	assert.Equal("test-data", result) // Should return data unchanged
}

func (suite *TestSuiteSecurityHooks) TestDecodeHook_WrongToType() {
	assert := assert.New(suite.T())

	fromType := reflect.TypeOf(suite.validNoopData)
	toType := reflect.TypeOf("string") // Wrong target type

	result, err := DecodeHook(fromType, toType, suite.validNoopData)

	assert.NoError(err)
	assert.Equal(suite.validNoopData, result) // Should return data unchanged
}

func (suite *TestSuiteSecurityHooks) TestDecodeHook_InvalidDataType() {
	assert := assert.New(suite.T())

	// Use map type so we pass the first check, but pass wrong data type
	fromType := reflect.TypeOf(map[string]any{})
	toType := reflect.TypeOf(Security{})

	result, err := DecodeHook(fromType, toType, suite.wrongTypeData)

	assert.Error(err)
	assert.Contains(err.Error(), "expected map[string]any for Security")
	assert.Equal(suite.wrongTypeData, result)
}

func (suite *TestSuiteSecurityHooks) TestDecodeHook_InvalidSecurityType() {
	assert := assert.New(suite.T())

	fromType := reflect.TypeOf(suite.invalidTypeData)
	toType := reflect.TypeOf(Security{})

	result, err := DecodeHook(fromType, toType, suite.invalidTypeData)

	assert.Error(err)
	assert.Contains(err.Error(), "security type must be a string")
	assert.Equal(suite.invalidTypeData, result)
}

func (suite *TestSuiteSecurityHooks) TestDecodeHook_UnknownSecurityType() {
	assert := assert.New(suite.T())

	fromType := reflect.TypeOf(suite.invalidSpecData)
	toType := reflect.TypeOf(Security{})

	result, err := DecodeHook(fromType, toType, suite.invalidSpecData)

	assert.Error(err)
	assert.Contains(err.Error(), "error creating spec")
	assert.Contains(err.Error(), "unknown security type: unknown-security-type")
	assert.Nil(result)
}

func (suite *TestSuiteSecurityHooks) TestDecodeHook_MissingTypeField() {
	assert := assert.New(suite.T())

	dataWithoutType := map[string]any{
		"specs": map[string]any{},
	}

	fromType := reflect.TypeOf(dataWithoutType)
	toType := reflect.TypeOf(Security{})

	result, err := DecodeHook(fromType, toType, dataWithoutType)

	assert.Error(err)
	assert.Contains(err.Error(), "security type must be a string")
	assert.Equal(dataWithoutType, result)
}

func (suite *TestSuiteSecurityHooks) TestCreateSpec_ValidTypes() {
	assert := assert.New(suite.T())

	testCases := []struct {
		securityType string
		expectedType any
	}{
		{"noop", &noop.NoopSecuritySpec{}},
		{"github", &github.GitHubSecuritySpec{}},
		{"custom", &custom.CustomSecuritySpec{}},
	}

	for _, tc := range testCases {
		spec, err := createSpec(tc.securityType)

		assert.NoError(err, "Security type: %s", tc.securityType)
		assert.IsType(tc.expectedType, spec, "Security type: %s", tc.securityType)
	}
}

func (suite *TestSuiteSecurityHooks) TestCreateSpec_UnknownType() {
	assert := assert.New(suite.T())

	spec, err := createSpec("unknown-type")

	assert.Error(err)
	assert.Contains(err.Error(), "unknown security type: unknown-type")
	assert.Nil(spec)
}

func (suite *TestSuiteSecurityHooks) TestDecodeHook_ComplexGitHubSpecs() {
	assert := assert.New(suite.T())

	complexGitHubData := map[string]any{
		"type": "github",
		"specs": map[string]any{
			"secretToken": "github-webhook-secret",
			"eventTypes":  []string{"push", "pull_request"},
		},
	}

	fromType := reflect.TypeOf(complexGitHubData)
	toType := reflect.TypeOf(Security{})

	result, err := DecodeHook(fromType, toType, complexGitHubData)

	assert.NoError(err)
	assert.IsType(Security{}, result)

	security := result.(Security)
	assert.Equal("github", security.Type)
	assert.IsType(&github.GitHubSecuritySpec{}, security.Specs)
}

func (suite *TestSuiteSecurityHooks) TestDecodeHook_ComplexCustomSpecs() {
	assert := assert.New(suite.T())

	complexCustomData := map[string]any{
		"type": "custom",
		"specs": map[string]any{
			"headerName":   "X-Custom-Token",
			"secretToken":  "custom-secret-123",
			"allowedHosts": []string{"example.com", "api.example.com"},
		},
	}

	fromType := reflect.TypeOf(complexCustomData)
	toType := reflect.TypeOf(Security{})

	result, err := DecodeHook(fromType, toType, complexCustomData)

	assert.NoError(err)
	assert.IsType(Security{}, result)

	security := result.(Security)
	assert.Equal("custom", security.Type)
	assert.IsType(&custom.CustomSecuritySpec{}, security.Specs)
}

func TestRunSecurityHooksSuite(t *testing.T) {
	suite.Run(t, new(TestSuiteSecurityHooks))
}

// Benchmarks

func BenchmarkDecodeHook_NoopSecurity(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	data := map[string]any{
		"type":  "noop",
		"specs": map[string]any{},
	}
	fromType := reflect.TypeOf(data)
	toType := reflect.TypeOf(Security{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeHook(fromType, toType, data) // nolint:errcheck
	}
}

func BenchmarkDecodeHook_GitHubSecurity(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	data := map[string]any{
		"type": "github",
		"specs": map[string]any{
			"secretToken": "test-secret",
		},
	}
	fromType := reflect.TypeOf(data)
	toType := reflect.TypeOf(Security{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeHook(fromType, toType, data) // nolint:errcheck
	}
}

func BenchmarkDecodeHook_CustomSecurity(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	data := map[string]any{
		"type": "custom",
		"specs": map[string]any{
			"headerName":  "X-Custom-Secret",
			"secretToken": "test-token",
		},
	}
	fromType := reflect.TypeOf(data)
	toType := reflect.TypeOf(Security{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeHook(fromType, toType, data) // nolint:errcheck
	}
}

func BenchmarkCreateSpec_AllTypes(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	types := []string{"noop", "github", "custom"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		securityType := types[i%len(types)]
		createSpec(securityType) // nolint:errcheck
	}
}

func BenchmarkDecodeHook_WrongType(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	fromType := reflect.TypeOf("string")
	toType := reflect.TypeOf(Security{})
	data := "test-data"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeHook(fromType, toType, data) // nolint:errcheck
	}
}

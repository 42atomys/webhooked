//go:build unit

package github

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/42atomys/webhooked/internal/fasthttpz"
	"github.com/42atomys/webhooked/internal/valuable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

type TestSuiteGitHubSecurity struct {
	suite.Suite

	validSecret      *valuable.Valuable
	emptySecret      *valuable.Valuable
	testSecret       string
	testPayload      []byte
	validSignature   string
	invalidSignature string
	ctx              context.Context
	requestCtx       *fasthttpz.RequestCtx
}

func (suite *TestSuiteGitHubSecurity) BeforeTest(suiteName, testName string) {
	var err error

	suite.testSecret = "my-secret-key"
	suite.testPayload = []byte(`{"action":"opened","number":1}`)

	// Valid secret
	suite.validSecret, err = valuable.Serialize(suite.testSecret)
	require.NoError(suite.T(), err)

	// Empty secret
	suite.emptySecret, err = valuable.Serialize("")
	require.NoError(suite.T(), err)

	// Generate valid signature
	h := hmac.New(sha256.New, []byte(suite.testSecret))
	h.Write(suite.testPayload)
	suite.validSignature = "sha256=" + hex.EncodeToString(h.Sum(nil))

	// Invalid signature
	suite.invalidSignature = "sha256=invalid_signature_hash"

	// Setup context and request context
	suite.ctx = context.Background()
	suite.requestCtx = &fasthttpz.RequestCtx{
		RequestCtx: &fasthttp.RequestCtx{},
	}
	suite.requestCtx.Request.SetBody(suite.testPayload)
}

func (suite *TestSuiteGitHubSecurity) TestEnsureConfigurationCompleteness_Always() {
	assert := assert.New(suite.T())

	spec := &GitHubSecuritySpec{
		Secret: suite.validSecret,
	}

	err := spec.EnsureConfigurationCompleteness()

	// GitHub security spec always returns nil for configuration completeness
	assert.NoError(err)
}

func (suite *TestSuiteGitHubSecurity) TestEnsureConfigurationCompleteness_NilSecret() {
	assert := assert.New(suite.T())

	spec := &GitHubSecuritySpec{
		Secret: nil,
	}

	err := spec.EnsureConfigurationCompleteness()

	// GitHub security spec always returns nil for configuration completeness
	assert.NoError(err)
}

func (suite *TestSuiteGitHubSecurity) TestInitialize_Always() {
	assert := assert.New(suite.T())

	spec := &GitHubSecuritySpec{
		Secret: suite.validSecret,
	}

	err := spec.Initialize()

	// GitHub security spec always returns nil for initialization
	assert.NoError(err)
}

func (suite *TestSuiteGitHubSecurity) TestInitialize_NilSecret() {
	assert := assert.New(suite.T())

	spec := &GitHubSecuritySpec{
		Secret: nil,
	}

	err := spec.Initialize()

	// GitHub security spec always returns nil for initialization
	assert.NoError(err)
}

func (suite *TestSuiteGitHubSecurity) TestIsSecure_ValidSignature() {
	assert := assert.New(suite.T())

	spec := &GitHubSecuritySpec{
		Secret: suite.validSecret,
	}

	// Set the valid signature header
	suite.requestCtx.Request.Header.Set(headerName, suite.validSignature)

	result, err := spec.IsSecure(suite.ctx, suite.requestCtx)

	assert.NoError(err)
	assert.True(result)
}

func (suite *TestSuiteGitHubSecurity) TestIsSecure_InvalidSignature() {
	assert := assert.New(suite.T())

	spec := &GitHubSecuritySpec{
		Secret: suite.validSecret,
	}

	// Set the invalid signature header
	suite.requestCtx.Request.Header.Set(headerName, suite.invalidSignature)

	result, err := spec.IsSecure(suite.ctx, suite.requestCtx)

	assert.NoError(err)
	assert.False(result)
}

func (suite *TestSuiteGitHubSecurity) TestIsSecure_MissingHeader() {
	assert := assert.New(suite.T())

	spec := &GitHubSecuritySpec{
		Secret: suite.validSecret,
	}

	// Don't set any signature header
	result, err := spec.IsSecure(suite.ctx, suite.requestCtx)

	assert.NoError(err)
	assert.False(result)
}

func (suite *TestSuiteGitHubSecurity) TestIsSecure_EmptyHeader() {
	assert := assert.New(suite.T())

	spec := &GitHubSecuritySpec{
		Secret: suite.validSecret,
	}

	// Set empty signature header
	suite.requestCtx.Request.Header.Set(headerName, "")

	result, err := spec.IsSecure(suite.ctx, suite.requestCtx)

	assert.NoError(err)
	assert.False(result)
}

func (suite *TestSuiteGitHubSecurity) TestIsSecure_NilSecret() {
	assert := assert.New(suite.T())

	spec := &GitHubSecuritySpec{
		Secret: nil,
	}

	// Set valid signature header
	suite.requestCtx.Request.Header.Set(headerName, suite.validSignature)

	result, err := spec.IsSecure(suite.ctx, suite.requestCtx)

	assert.Error(err)
	assert.Contains(err.Error(), "secret is required")
	assert.False(result)
}

func (suite *TestSuiteGitHubSecurity) TestIsSecure_EmptySecret() {
	assert := assert.New(suite.T())

	spec := &GitHubSecuritySpec{
		Secret: suite.emptySecret,
	}

	// Set valid signature header
	suite.requestCtx.Request.Header.Set(headerName, suite.validSignature)

	result, err := spec.IsSecure(suite.ctx, suite.requestCtx)

	assert.Error(err)
	assert.Contains(err.Error(), "secret is required")
	assert.False(result)
}

func (suite *TestSuiteGitHubSecurity) TestIsSecure_DifferentPayload() {
	assert := assert.New(suite.T())

	spec := &GitHubSecuritySpec{
		Secret: suite.validSecret,
	}

	// Change the payload but keep the same signature
	suite.requestCtx.Request.SetBody([]byte(`{"action":"closed","number":2}`))
	suite.requestCtx.Request.Header.Set(headerName, suite.validSignature)

	result, err := spec.IsSecure(suite.ctx, suite.requestCtx)

	assert.NoError(err)
	assert.False(result) // Should fail because payload changed
}

func (suite *TestSuiteGitHubSecurity) TestIsSecure_DifferentSecret() {
	assert := assert.New(suite.T())

	// Create spec with different secret
	differentSecret, err := valuable.Serialize("different-secret")
	require.NoError(suite.T(), err)

	spec := &GitHubSecuritySpec{
		Secret: differentSecret,
	}

	// Use signature generated with original secret
	suite.requestCtx.Request.Header.Set(headerName, suite.validSignature)

	result, err := spec.IsSecure(suite.ctx, suite.requestCtx)

	assert.NoError(err)
	assert.False(result) // Should fail because secret is different
}

func (suite *TestSuiteGitHubSecurity) TestIsSecure_EmptyPayload() {
	assert := assert.New(suite.T())

	spec := &GitHubSecuritySpec{
		Secret: suite.validSecret,
	}

	// Generate signature for empty payload
	emptyPayload := []byte("")
	h := hmac.New(sha256.New, []byte(suite.testSecret))
	h.Write(emptyPayload)
	emptySignature := "sha256=" + hex.EncodeToString(h.Sum(nil))

	suite.requestCtx.Request.SetBody(emptyPayload)
	suite.requestCtx.Request.Header.Set(headerName, emptySignature)

	result, err := spec.IsSecure(suite.ctx, suite.requestCtx)

	assert.NoError(err)
	assert.True(result)
}

func (suite *TestSuiteGitHubSecurity) TestIsSecure_LargePayload() {
	assert := assert.New(suite.T())

	spec := &GitHubSecuritySpec{
		Secret: suite.validSecret,
	}

	// Create large payload
	largePayload := make([]byte, 10000)
	for i := range largePayload {
		largePayload[i] = byte(i % 256)
	}

	// Generate signature for large payload
	h := hmac.New(sha256.New, []byte(suite.testSecret))
	h.Write(largePayload)
	largeSignature := "sha256=" + hex.EncodeToString(h.Sum(nil))

	suite.requestCtx.Request.SetBody(largePayload)
	suite.requestCtx.Request.Header.Set(headerName, largeSignature)

	result, err := spec.IsSecure(suite.ctx, suite.requestCtx)

	assert.NoError(err)
	assert.True(result)
}

func (suite *TestSuiteGitHubSecurity) TestIsSecure_MalformedSignature() {
	assert := assert.New(suite.T())

	spec := &GitHubSecuritySpec{
		Secret: suite.validSecret,
	}

	// Set malformed signature (missing "sha256=" prefix)
	malformedSignature := hex.EncodeToString([]byte("invalid"))
	suite.requestCtx.Request.Header.Set(headerName, malformedSignature)

	result, err := spec.IsSecure(suite.ctx, suite.requestCtx)

	assert.NoError(err)
	assert.False(result)
}

func (suite *TestSuiteGitHubSecurity) TestIsSecure_CaseInsensitiveHeader() {
	assert := assert.New(suite.T())

	spec := &GitHubSecuritySpec{
		Secret: suite.validSecret,
	}

	// Set header with different case (fasthttp handles case sensitivity)
	suite.requestCtx.Request.Header.Set("x-hub-signature-256", suite.validSignature)

	result, err := spec.IsSecure(suite.ctx, suite.requestCtx)

	assert.NoError(err)
	assert.True(result) // Should work because fasthttp normalizes headers
}

func (suite *TestSuiteGitHubSecurity) TestFullWorkflow_ValidConfiguration() {
	assert := assert.New(suite.T())

	spec := &GitHubSecuritySpec{
		Secret: suite.validSecret,
	}

	// Test complete workflow
	err := spec.EnsureConfigurationCompleteness()
	assert.NoError(err)

	err = spec.Initialize()
	assert.NoError(err)

	suite.requestCtx.Request.Header.Set(headerName, suite.validSignature)
	result, err := spec.IsSecure(suite.ctx, suite.requestCtx)
	assert.NoError(err)
	assert.True(result)
}

// Note: Nil receiver tests removed as GitHub security methods
// return nil/false gracefully rather than panicking

func (suite *TestSuiteGitHubSecurity) TestHeaderConstant() {
	assert := assert.New(suite.T())

	// Test that the header constant is correct
	assert.Equal("X-Hub-Signature-256", headerName)
}

func TestRunGitHubSecuritySuite(t *testing.T) {
	suite.Run(t, new(TestSuiteGitHubSecurity))
}

// Benchmarks

func BenchmarkEnsureConfigurationCompleteness(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	secret, _ := valuable.Serialize("test-secret")
	spec := &GitHubSecuritySpec{
		Secret: secret,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spec.EnsureConfigurationCompleteness() // nolint:errcheck
	}
}

func BenchmarkInitialize(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	secret, _ := valuable.Serialize("test-secret")
	spec := &GitHubSecuritySpec{
		Secret: secret,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spec.Initialize() // nolint:errcheck
	}
}

func BenchmarkIsSecure_ValidSignature(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	secret, _ := valuable.Serialize("test-secret")
	spec := &GitHubSecuritySpec{
		Secret: secret,
	}

	payload := []byte(`{"action":"opened","number":1}`)
	h := hmac.New(sha256.New, []byte("test-secret"))
	h.Write(payload)
	signature := "sha256=" + hex.EncodeToString(h.Sum(nil))

	ctx := context.Background()
	requestCtx := &fasthttpz.RequestCtx{
		RequestCtx: &fasthttp.RequestCtx{},
	}
	requestCtx.Request.SetBody(payload)
	requestCtx.Request.Header.Set(headerName, signature)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spec.IsSecure(ctx, requestCtx) // nolint:errcheck
	}
}

func BenchmarkIsSecure_InvalidSignature(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	secret, _ := valuable.Serialize("test-secret")
	spec := &GitHubSecuritySpec{
		Secret: secret,
	}

	payload := []byte(`{"action":"opened","number":1}`)
	invalidSignature := "sha256=invalid_signature_hash"

	ctx := context.Background()
	requestCtx := &fasthttpz.RequestCtx{
		RequestCtx: &fasthttp.RequestCtx{},
	}
	requestCtx.Request.SetBody(payload)
	requestCtx.Request.Header.Set(headerName, invalidSignature)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spec.IsSecure(ctx, requestCtx) // nolint:errcheck
	}
}

func BenchmarkHMACGeneration(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	secret := []byte("test-secret")
	payload := []byte(`{"action":"opened","number":1}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h := hmac.New(sha256.New, secret)
		h.Write(payload)
		hex.EncodeToString(h.Sum(nil))
	}
}

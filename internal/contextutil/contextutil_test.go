//go:build unit

package contextutil

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestSuiteContextUtil struct {
	suite.Suite

	baseCtx context.Context
}

func (suite *TestSuiteContextUtil) BeforeTest(suiteName, testName string) {
	suite.baseCtx = context.Background()
}

// Test WebhookSpec context operations

func (suite *TestSuiteContextUtil) TestWithWebhookSpec_ValidValue() {
	assert := assert.New(suite.T())

	spec := map[string]string{"name": "test-webhook"}
	ctx := WithWebhookSpec(suite.baseCtx, spec)

	assert.NotNil(ctx)
	assert.NotEqual(suite.baseCtx, ctx)
}

func (suite *TestSuiteContextUtil) TestWebhookSpecFromContext_ValidType() {
	assert := assert.New(suite.T())

	spec := map[string]string{"name": "test-webhook"}
	ctx := WithWebhookSpec(suite.baseCtx, spec)

	retrieved, ok := WebhookSpecFromContext[map[string]string](ctx)

	assert.True(ok)
	assert.Equal(spec, retrieved)
}

func (suite *TestSuiteContextUtil) TestWebhookSpecFromContext_InvalidType() {
	assert := assert.New(suite.T())

	spec := map[string]string{"name": "test-webhook"}
	ctx := WithWebhookSpec(suite.baseCtx, spec)

	// Try to retrieve as wrong type
	retrieved, ok := WebhookSpecFromContext[map[string]int](ctx)

	assert.False(ok)
	assert.Empty(retrieved)
}

func (suite *TestSuiteContextUtil) TestWebhookSpecFromContext_NoValue() {
	assert := assert.New(suite.T())

	// Context without webhook spec
	retrieved, ok := WebhookSpecFromContext[map[string]string](suite.baseCtx)

	assert.False(ok)
	assert.Empty(retrieved)
}

func (suite *TestSuiteContextUtil) TestWebhookSpecFromContext_NilValue() {
	assert := assert.New(suite.T())

	ctx := WithWebhookSpec(suite.baseCtx, nil)
	retrieved, ok := WebhookSpecFromContext[map[string]string](ctx)

	assert.False(ok)
	assert.Empty(retrieved)
}

func (suite *TestSuiteContextUtil) TestWebhookSpecFromContext_DifferentTypes() {
	assert := assert.New(suite.T())

	// Test with string
	ctx1 := WithWebhookSpec(suite.baseCtx, "string-spec")
	stringSpec, ok := WebhookSpecFromContext[string](ctx1)
	assert.True(ok)
	assert.Equal("string-spec", stringSpec)

	// Test with int
	ctx2 := WithWebhookSpec(suite.baseCtx, 42)
	intSpec, ok := WebhookSpecFromContext[int](ctx2)
	assert.True(ok)
	assert.Equal(42, intSpec)

	// Test with struct
	type testStruct struct {
		Name string
		ID   int
	}
	testSpec := testStruct{Name: "test", ID: 123}
	ctx3 := WithWebhookSpec(suite.baseCtx, testSpec)
	structSpec, ok := WebhookSpecFromContext[testStruct](ctx3)
	assert.True(ok)
	assert.Equal(testSpec, structSpec)
}

// Test RequestCtx context operations

func (suite *TestSuiteContextUtil) TestWithRequestCtx_ValidValue() {
	assert := assert.New(suite.T())

	reqCtx := map[string]any{"method": "POST", "path": "/webhook"}
	ctx := WithRequestCtx(suite.baseCtx, reqCtx)

	assert.NotNil(ctx)
	assert.NotEqual(suite.baseCtx, ctx)
}

func (suite *TestSuiteContextUtil) TestRequestCtxFromContext_ValidType() {
	assert := assert.New(suite.T())

	reqCtx := map[string]any{"method": "POST", "path": "/webhook"}
	ctx := WithRequestCtx(suite.baseCtx, reqCtx)

	retrieved, ok := RequestCtxFromContext[map[string]any](ctx)

	assert.True(ok)
	assert.Equal(reqCtx, retrieved)
}

func (suite *TestSuiteContextUtil) TestRequestCtxFromContext_InvalidType() {
	assert := assert.New(suite.T())

	reqCtx := map[string]any{"method": "POST"}
	ctx := WithRequestCtx(suite.baseCtx, reqCtx)

	// Try to retrieve as wrong type
	retrieved, ok := RequestCtxFromContext[string](ctx)

	assert.False(ok)
	assert.Empty(retrieved)
}

func (suite *TestSuiteContextUtil) TestRequestCtxFromContext_NoValue() {
	assert := assert.New(suite.T())

	retrieved, ok := RequestCtxFromContext[map[string]any](suite.baseCtx)

	assert.False(ok)
	assert.Empty(retrieved)
}

func (suite *TestSuiteContextUtil) TestRequestCtxFromContext_NilValue() {
	assert := assert.New(suite.T())

	ctx := WithRequestCtx(suite.baseCtx, nil)
	retrieved, ok := RequestCtxFromContext[map[string]any](ctx)

	assert.False(ok)
	assert.Empty(retrieved)
}

// Test Store context operations

func (suite *TestSuiteContextUtil) TestWithStore_ValidValue() {
	assert := assert.New(suite.T())

	store := map[string]string{"type": "redis", "addr": "localhost:6379"}
	ctx := WithStore(suite.baseCtx, store)

	assert.NotNil(ctx)
	assert.NotEqual(suite.baseCtx, ctx)
}

func (suite *TestSuiteContextUtil) TestStoreFromContext_ValidType() {
	assert := assert.New(suite.T())

	store := map[string]string{"type": "redis", "addr": "localhost:6379"}
	ctx := WithStore(suite.baseCtx, store)

	retrieved, ok := StoreFromContext[map[string]string](ctx)

	assert.True(ok)
	assert.Equal(store, retrieved)
}

func (suite *TestSuiteContextUtil) TestStoreFromContext_InvalidType() {
	assert := assert.New(suite.T())

	store := map[string]string{"type": "redis"}
	ctx := WithStore(suite.baseCtx, store)

	// Try to retrieve as wrong type
	retrieved, ok := StoreFromContext[[]string](ctx)

	assert.False(ok)
	assert.Empty(retrieved)
}

func (suite *TestSuiteContextUtil) TestStoreFromContext_NoValue() {
	assert := assert.New(suite.T())

	retrieved, ok := StoreFromContext[map[string]string](suite.baseCtx)

	assert.False(ok)
	assert.Empty(retrieved)
}

func (suite *TestSuiteContextUtil) TestStoreFromContext_NilValue() {
	assert := assert.New(suite.T())

	ctx := WithStore(suite.baseCtx, nil)
	retrieved, ok := StoreFromContext[map[string]string](ctx)

	assert.False(ok)
	assert.Empty(retrieved)
}

// Test multiple context values together

func (suite *TestSuiteContextUtil) TestMultipleContextValues() {
	assert := assert.New(suite.T())

	webhookSpec := "test-webhook"
	requestCtx := "test-request"
	store := "test-store"

	// Add all values to context
	ctx := WithWebhookSpec(suite.baseCtx, webhookSpec)
	ctx = WithRequestCtx(ctx, requestCtx)
	ctx = WithStore(ctx, store)

	// Retrieve all values
	retrievedWebhookSpec, ok1 := WebhookSpecFromContext[string](ctx)
	retrievedRequestCtx, ok2 := RequestCtxFromContext[string](ctx)
	retrievedStore, ok3 := StoreFromContext[string](ctx)

	assert.True(ok1)
	assert.True(ok2)
	assert.True(ok3)
	assert.Equal(webhookSpec, retrievedWebhookSpec)
	assert.Equal(requestCtx, retrievedRequestCtx)
	assert.Equal(store, retrievedStore)
}

func (suite *TestSuiteContextUtil) TestOverwriteContextValues() {
	assert := assert.New(suite.T())

	// Set initial value
	initialSpec := "initial-webhook"
	ctx := WithWebhookSpec(suite.baseCtx, initialSpec)

	// Overwrite with new value
	newSpec := "new-webhook"
	ctx = WithWebhookSpec(ctx, newSpec)

	// Should retrieve the new value
	retrieved, ok := WebhookSpecFromContext[string](ctx)

	assert.True(ok)
	assert.Equal(newSpec, retrieved)
	assert.NotEqual(initialSpec, retrieved)
}

// Test context key constants

func (suite *TestSuiteContextUtil) TestContextKeys_Uniqueness() {
	assert := assert.New(suite.T())

	// Ensure all context keys are unique
	assert.NotEqual(webhookSpecCtxKey, requestCtxKey)
	assert.NotEqual(webhookSpecCtxKey, storeCtxKey)
	assert.NotEqual(requestCtxKey, storeCtxKey)
}

func (suite *TestSuiteContextUtil) TestContextKeys_Type() {
	assert := assert.New(suite.T())

	// Ensure context keys are the correct type
	assert.IsType(ContextKey(0), webhookSpecCtxKey)
	assert.IsType(ContextKey(0), requestCtxKey)
	assert.IsType(ContextKey(0), storeCtxKey)
}

func (suite *TestSuiteContextUtil) TestContextKeys_Values() {
	assert := assert.New(suite.T())

	// Test the actual values (based on iota)
	assert.Equal(ContextKey(0), webhookSpecCtxKey)
	assert.Equal(ContextKey(1), requestCtxKey)
	assert.Equal(ContextKey(2), storeCtxKey)
}

// Test edge cases

func (suite *TestSuiteContextUtil) TestNilContext() {
	assert := assert.New(suite.T())

	// Test with nil context (should panic)
	assert.Panics(func() {
		WithWebhookSpec(nil, "test")
	})

	assert.Panics(func() {
		WithRequestCtx(nil, "test")
	})

	assert.Panics(func() {
		WithStore(nil, "test")
	})
}

func (suite *TestSuiteContextUtil) TestEmptyValueRetrieval() {
	assert := assert.New(suite.T())

	// Test retrieving empty string
	ctx := WithWebhookSpec(suite.baseCtx, "")
	retrieved, ok := WebhookSpecFromContext[string](ctx)

	assert.True(ok)
	assert.Equal("", retrieved)
}

func (suite *TestSuiteContextUtil) TestZeroValueRetrieval() {
	assert := assert.New(suite.T())

	// Test retrieving zero int
	ctx := WithWebhookSpec(suite.baseCtx, 0)
	retrieved, ok := WebhookSpecFromContext[int](ctx)

	assert.True(ok)
	assert.Equal(0, retrieved)

	// Test retrieving false bool
	ctx2 := WithWebhookSpec(suite.baseCtx, false)
	retrieved2, ok2 := WebhookSpecFromContext[bool](ctx2)

	assert.True(ok2)
	assert.Equal(false, retrieved2)
}

func (suite *TestSuiteContextUtil) TestComplexTypeRetrieval() {
	assert := assert.New(suite.T())

	type complexStruct struct {
		Name     string
		Values   []int
		Metadata map[string]any
		Nested   struct {
			ID   int
			Tags []string
		}
	}

	complex := complexStruct{
		Name:   "test",
		Values: []int{1, 2, 3},
		Metadata: map[string]any{
			"version": "1.0",
			"active":  true,
		},
		Nested: struct {
			ID   int
			Tags []string
		}{
			ID:   42,
			Tags: []string{"tag1", "tag2"},
		},
	}

	ctx := WithWebhookSpec(suite.baseCtx, complex)
	retrieved, ok := WebhookSpecFromContext[complexStruct](ctx)

	assert.True(ok)
	assert.Equal(complex, retrieved)
	assert.Equal("test", retrieved.Name)
	assert.Equal([]int{1, 2, 3}, retrieved.Values)
	assert.Equal(42, retrieved.Nested.ID)
}

func TestRunContextUtilSuite(t *testing.T) {
	suite.Run(t, new(TestSuiteContextUtil))
}

// Benchmarks

func BenchmarkWithWebhookSpec(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	ctx := context.Background()
	spec := map[string]string{"name": "benchmark-webhook"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WithWebhookSpec(ctx, spec)
	}
}

func BenchmarkWebhookSpecFromContext(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	ctx := context.Background()
	spec := map[string]string{"name": "benchmark-webhook"}
	ctx = WithWebhookSpec(ctx, spec)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WebhookSpecFromContext[map[string]string](ctx) // nolint:errcheck
	}
}

func BenchmarkWithRequestCtx(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	ctx := context.Background()
	reqCtx := map[string]any{"method": "POST", "path": "/webhook"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WithRequestCtx(ctx, reqCtx)
	}
}

func BenchmarkRequestCtxFromContext(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	ctx := context.Background()
	reqCtx := map[string]any{"method": "POST", "path": "/webhook"}
	ctx = WithRequestCtx(ctx, reqCtx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RequestCtxFromContext[map[string]any](ctx) // nolint:errcheck
	}
}

func BenchmarkWithStore(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	ctx := context.Background()
	store := map[string]string{"type": "redis", "addr": "localhost:6379"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WithStore(ctx, store)
	}
}

func BenchmarkStoreFromContext(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	ctx := context.Background()
	store := map[string]string{"type": "redis", "addr": "localhost:6379"}
	ctx = WithStore(ctx, store)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		StoreFromContext[map[string]string](ctx) // nolint:errcheck
	}
}

func BenchmarkMultipleContextOperations(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	ctx := context.Background()
	webhookSpec := "benchmark-webhook"
	requestCtx := "benchmark-request"
	store := "benchmark-store"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := WithWebhookSpec(ctx, webhookSpec)
		ctx = WithRequestCtx(ctx, requestCtx)
		ctx = WithStore(ctx, store)

		WebhookSpecFromContext[string](ctx) // nolint:errcheck
		RequestCtxFromContext[string](ctx)  // nolint:errcheck
		StoreFromContext[string](ctx)       // nolint:errcheck
	}
}

func BenchmarkTypeAssertion_Success(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	ctx := context.Background()
	spec := map[string]string{"name": "benchmark"}
	ctx = WithWebhookSpec(ctx, spec)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WebhookSpecFromContext[map[string]string](ctx) // nolint:errcheck
	}
}

func BenchmarkTypeAssertion_Failure(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	ctx := context.Background()
	spec := map[string]string{"name": "benchmark"}
	ctx = WithWebhookSpec(ctx, spec)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WebhookSpecFromContext[[]string](ctx) // nolint:errcheck // Wrong type
	}
}

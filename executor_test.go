//go:build unit

package webhooked

import (
	"context"
	"errors"
	"testing"

	"github.com/42atomys/webhooked/format"
	"github.com/42atomys/webhooked/internal/config"
	"github.com/42atomys/webhooked/internal/fasthttpz"
	"github.com/42atomys/webhooked/security"
	securityNoop "github.com/42atomys/webhooked/security/noop"
	"github.com/42atomys/webhooked/storage"
	storageNoop "github.com/42atomys/webhooked/storage/noop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestNewExecutor(t *testing.T) {
	executor := NewExecutor(&config.Config{})
	assert.NotNil(t, executor)
	assert.IsType(t, &DefaultExecutor{}, executor)
}

func TestDefaultExecutor_IncomingRequest_SpecNotFound(t *testing.T) {
	executor := NewExecutor(&config.Config{})

	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}
	ctx.Request.SetRequestURI("/nonexistent/path")

	// Execute
	err := executor.IncomingRequest(context.Background(), ctx)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, fasthttp.StatusNotFound, ctx.Response.StatusCode())
}

func TestDefaultExecutor_IncomingRequest_Success(t *testing.T) {
	executor := NewExecutor(setupTestConfig(t))

	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}
	ctx.Request.SetRequestURI("/webhooks/v1alpha2/test")
	ctx.Request.Header.SetMethod("POST")
	ctx.Request.SetBody([]byte(`{"test": "data"}`))

	// Execute
	err := executor.IncomingRequest(context.Background(), ctx)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fasthttp.StatusNoContent, ctx.Response.StatusCode())
}

func TestDefaultExecutor_IncomingRequest_SecurityFailure(t *testing.T) {
	executor := NewExecutor(setupTestConfigWithFailingSecurity(t))

	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}
	ctx.Request.SetRequestURI("/webhooks/v1alpha2/secure-test")
	ctx.Request.Header.SetMethod("POST")
	ctx.Request.SetBody([]byte(`{"test": "data"}`))

	// Execute
	err := executor.IncomingRequest(context.Background(), ctx)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, fasthttp.StatusUnauthorized, ctx.Response.StatusCode())
}

func TestDefaultExecutor_IncomingRequest_SecurityError(t *testing.T) {
	executor := NewExecutor(setupTestConfigWithFailingSecurity(t))

	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}
	ctx.Request.SetRequestURI("/webhooks/v1alpha2/secure-test-error")
	ctx.Request.Header.SetMethod("POST")
	ctx.Request.SetBody([]byte(`{"test": "data"}`))

	// Execute
	err := executor.IncomingRequest(context.Background(), ctx)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, fasthttp.StatusInternalServerError, ctx.Response.StatusCode())
}
func TestDefaultExecutor_pipelineOrder(t *testing.T) {
	executor := &DefaultExecutor{}
	pipeline := executor.pipelineOrder()

	assert.Len(t, pipeline, 3)
	// We can't directly test function equality, but we can test the count
}

func TestDefaultExecutor_pipelineSecure_Success(t *testing.T) {
	executor := &DefaultExecutor{}

	webhook := &config.Webhook{
		Security: security.Security{
			Type:  "noop",
			Specs: &securityNoop.NoopSecuritySpec{},
		},
	}

	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}

	resultCtx, err := executor.pipelineSecure(context.Background(), ctx, webhook)

	assert.NoError(t, err)
	assert.NotNil(t, resultCtx)
}

func TestDefaultExecutor_pipelineResponse_NoTemplate(t *testing.T) {
	executor := &DefaultExecutor{}

	webhook := &config.Webhook{
		Response: config.Response{},
	}

	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}

	resultCtx, err := executor.pipelineResponse(context.Background(), ctx, webhook)

	assert.NoError(t, err)
	assert.NotNil(t, resultCtx)
	assert.Equal(t, fasthttp.StatusNoContent, ctx.Response.StatusCode())
}

func TestDefaultExecutor_pipelineStore_Success(t *testing.T) {
	executor := &DefaultExecutor{}

	// Create a webhook with noop storage
	webhook := &config.Webhook{
		Storage: []*storage.Storage{
			{
				Type:       "noop",
				Formatting: &format.Formatting{},
				Specs:      &storageNoop.NoopStorageSpec{},
			},
		},
	}

	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}
	ctx.Request.SetBody([]byte(`{"test": "data"}`))

	resultCtx, err := executor.pipelineStore(context.Background(), ctx, webhook)

	assert.NoError(t, err)
	assert.NotNil(t, resultCtx)
}

// Helper functions for test setup

func setupTestConfig(t testing.TB) *config.Config {
	config := &config.Config{
		APIVersion: config.APIVersionV1Alpha2,
		Kind:       config.KindConfiguration,
		Specs: []*config.Spec{
			{
				Webhooks: []*config.Webhook{
					{
						Name:          "success-test",
						EntrypointURL: "/test",
						Security: security.Security{
							Type:  "noop",
							Specs: &securityNoop.NoopSecuritySpec{},
						},
					},
				},
			},
		},
	}

	require.NoError(t, config.Validate())
	return config
}

func setupTestConfigWithFailingSecurity(t *testing.T) *config.Config {
	config := &config.Config{
		APIVersion: config.APIVersionV1Alpha2,
		Kind:       config.KindConfiguration,
		Specs: []*config.Spec{
			{
				Webhooks: []*config.Webhook{
					{
						Name:          "secure-test",
						EntrypointURL: "/secure-test",
						Security: security.Security{
							Type:  "failling",
							Specs: &mockFailingSecurity{},
						},
					},
					{
						Name:          "secure-test-error",
						EntrypointURL: "/secure-test-error",
						Security: security.Security{
							Type:  "error",
							Specs: &mockErrorSecurity{},
						},
					},
				},
			},
		},
	}

	require.NoError(t, config.Validate())
	return config
}

// Mock security implementation that always fails
type mockFailingSecurity struct{}

func (m *mockFailingSecurity) IsSecure(ctx context.Context, rctx *fasthttpz.RequestCtx) (bool, error) {
	return false, nil
}

func (m *mockFailingSecurity) EnsureConfigurationCompleteness() error {
	return nil
}

func (m *mockFailingSecurity) Initialize() error {
	return nil
}

// Mock security implementation that returns an error
type mockErrorSecurity struct{}

func (m *mockErrorSecurity) IsSecure(ctx context.Context, rctx *fasthttpz.RequestCtx) (bool, error) {
	return false, errors.New("security check failed")
}

func (m *mockErrorSecurity) EnsureConfigurationCompleteness() error {
	return nil
}

func (m *mockErrorSecurity) Initialize() error {
	return nil
}

func TestDefaultExecutor_pipelineSecure_Error(t *testing.T) {
	executor := &DefaultExecutor{}

	webhook := &config.Webhook{
		Security: security.Security{
			Type:  "mock-error",
			Specs: &mockErrorSecurity{},
		},
	}

	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}

	resultCtx, err := executor.pipelineSecure(context.Background(), ctx, webhook)

	assert.Error(t, err)
	assert.NotNil(t, resultCtx)
	assert.Equal(t, fasthttp.StatusInternalServerError, ctx.Response.StatusCode())
}

func TestDefaultExecutor_pipelineSecure_Unauthorized(t *testing.T) {
	executor := &DefaultExecutor{}

	webhook := &config.Webhook{
		Security: security.Security{
			Type:  "mock-fail",
			Specs: &mockFailingSecurity{},
		},
	}

	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}

	resultCtx, err := executor.pipelineSecure(context.Background(), ctx, webhook)

	assert.Error(t, err)
	assert.NotNil(t, resultCtx)
	assert.Equal(t, fasthttp.StatusUnauthorized, ctx.Response.StatusCode())
}

// Benchmarks

func BenchmarkDefaultExecutor_IncomingRequest(b *testing.B) {
	executor := NewExecutor(setupTestConfig(b))

	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}
	ctx.Request.SetRequestURI("/webhooks/v1alpha2/test")
	ctx.Request.Header.SetMethod("POST")
	ctx.Request.SetBody([]byte(`{"test": "data"}`))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset response for each iteration
		ctx.Response.Reset()
		executor.IncomingRequest(context.Background(), ctx) // nolint:errcheck
	}
}

func BenchmarkDefaultExecutor_pipelineSecure(b *testing.B) {
	executor := &DefaultExecutor{}
	webhook := &config.Webhook{
		Security: security.Security{
			Type:  "noop",
			Specs: &securityNoop.NoopSecuritySpec{},
		},
	}

	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor.pipelineSecure(context.Background(), ctx, webhook) // nolint:errcheck
	}
}

func BenchmarkDefaultExecutor_pipelineStore_Single(b *testing.B) {
	executor := NewExecutor(&config.Config{})
	webhook := &config.Webhook{
		Storage: []*storage.Storage{
			{
				Type:       "noop",
				Formatting: &format.Formatting{},
				Specs:      &storageNoop.NoopStorageSpec{},
			},
		},
	}

	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}
	ctx.Request.SetBody([]byte(`{"test": "data"}`))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor.pipelineStore(context.Background(), ctx, webhook) // nolint:errcheck
	}
}

func BenchmarkDefaultExecutor_pipelineStore_Multiple(b *testing.B) {
	executor := NewExecutor(&config.Config{})

	// Create multiple storage backends
	storages := make([]*storage.Storage, 5)
	for i := 0; i < 5; i++ {
		storages[i] = &storage.Storage{
			Type:       "noop",
			Formatting: &format.Formatting{},
			Specs:      &storageNoop.NoopStorageSpec{},
		}
	}

	webhook := &config.Webhook{
		Storage: storages,
	}

	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}
	ctx.Request.SetBody([]byte(`{"test": "data"}`))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor.pipelineStore(context.Background(), ctx, webhook) // nolint:errcheck
	}
}

func BenchmarkDefaultExecutor_pipelineResponse_NoTemplate(b *testing.B) {
	executor := &DefaultExecutor{}
	webhook := &config.Webhook{
		Response: config.Response{},
	}

	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.Response.Reset()
		executor.pipelineResponse(context.Background(), ctx, webhook) // nolint:errcheck
	}
}

func BenchmarkDefaultExecutor_pipelineStore_Concurrent(b *testing.B) {
	executor := NewExecutor(&config.Config{})

	// Create multiple storage backends
	storages := make([]*storage.Storage, 10)
	for i := 0; i < 10; i++ {
		storages[i] = &storage.Storage{
			Type:       "noop",
			Formatting: &format.Formatting{},
			Specs:      &storageNoop.NoopStorageSpec{},
		}
	}

	webhook := &config.Webhook{
		Storage: storages,
	}

	b.RunParallel(func(pb *testing.PB) {
		ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}
		ctx.Request.SetBody([]byte(`{"test": "data"}`))

		for pb.Next() {
			executor.pipelineStore(context.Background(), ctx, webhook) // nolint:errcheck
		}
	})
}

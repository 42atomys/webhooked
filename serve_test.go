//go:build unit

package webhooked

import (
	"context"
	"testing"
	"time"

	"github.com/42atomys/webhooked/internal/config"
	"github.com/42atomys/webhooked/internal/fasthttpz"
	"github.com/42atomys/webhooked/security"
	"github.com/42atomys/webhooked/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestNewServer(t *testing.T) {
	server, err := NewServer(&config.Config{
		APIVersion: config.APIVersionV1Alpha2,
		Kind:       config.KindConfiguration,
		Specs:      []*config.Spec{},
	}, 8080)

	require.NoError(t, err)
	assert.NotNil(t, server)
	assert.Equal(t, 8080, server.port)
	assert.NotNil(t, server.server)
	assert.NotNil(t, server.executor)
}

func TestNewServer_InvalidConfig(t *testing.T) {
	_, err := NewServer(&config.Config{}, 8080)

	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid configuration")
}

func TestServer_HealthCheck(t *testing.T) {
	server, err := NewServer(&config.Config{
		APIVersion: config.APIVersionV1Alpha2,
		Kind:       config.KindConfiguration,
		Specs:      []*config.Spec{},
	}, 8080)
	require.NoError(t, err)

	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}
	ctx.Request.SetRequestURI("/health")

	server.handleHealthCheck(ctx)

	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.Contains(t, string(ctx.Response.Body()), "healthy")
	assert.Contains(t, string(ctx.Response.Body()), Version)
	assert.Equal(t, "application/json", string(ctx.Response.Header.ContentType()))
}

func TestServer_ReadinessCheck_NoConfig(t *testing.T) {
	server, err := NewServer(&config.Config{
		APIVersion: config.APIVersionV1Alpha2,
		Kind:       config.KindConfiguration,
		Specs:      []*config.Spec{},
	}, 8080)
	require.NoError(t, err)

	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}
	ctx.Request.SetRequestURI("/ready")

	server.handleReadinessCheck(ctx)

	assert.Equal(t, fasthttp.StatusServiceUnavailable, ctx.Response.StatusCode())
	assert.Contains(t, string(ctx.Response.Body()), "not ready")
	assert.Contains(t, string(ctx.Response.Body()), "reason")
	assert.Equal(t, "application/json", string(ctx.Response.Header.ContentType()))
}

func TestServer_ReadinessCheck_WithConfig(t *testing.T) {
	server, err := NewServer(setupMinimalConfig(), 8080)
	require.NoError(t, err)

	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}
	ctx.Request.SetRequestURI("/ready")

	server.handleReadinessCheck(ctx)

	// Since we can't easily mock the global config, expect ServiceUnavailable
	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.Contains(t, string(ctx.Response.Body()), "ready")
	assert.Contains(t, string(ctx.Response.Body()), "version")
	assert.Equal(t, "application/json", string(ctx.Response.Header.ContentType()))
}

func TestServer_RequestHandler_HealthEndpoints(t *testing.T) {
	server, err := NewServer(setupMinimalConfig(), 8080)
	require.NoError(t, err)

	handler := server.requestHandlerFunc()

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "health check",
			path:           "/health",
			expectedStatus: fasthttp.StatusOK,
		},
		{
			name:           "readiness check",
			path:           "/ready",
			expectedStatus: fasthttp.StatusOK, // No config loaded
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}
			ctx.Request.SetRequestURI(tt.path)

			handler(ctx.RequestCtx)

			assert.Equal(t, tt.expectedStatus, ctx.Response.StatusCode())
		})
	}
}

func TestServer_RequestHandler_WebhookPath(t *testing.T) {
	server, err := NewServer(setupMinimalConfig(), 8080)
	require.NoError(t, err)

	handler := server.requestHandlerFunc()

	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}
	ctx.Request.SetRequestURI("/webhooks/v1alpha2/test")
	ctx.Request.Header.SetMethod("POST")
	ctx.Request.SetBody([]byte(`{"test": "data"}`))

	handler(ctx.RequestCtx)

	// Should return 404 since we don't have a matching webhook configured
	assert.Equal(t, fasthttp.StatusNotFound, ctx.Response.StatusCode())
}

func TestServer_RequestHandler_WebhookPath_RateLimitExceeded(t *testing.T) {
	config := &config.Config{
		APIVersion: config.APIVersionV1Alpha2,
		Kind:       config.KindConfiguration,
		Specs: []*config.Spec{
			{
				Webhooks: []*config.Webhook{
					{
						Name:          "test",
						EntrypointURL: "/test",
						Security:      security.Security{},
						Storage:       []*storage.Storage{},
						Response:      config.Response{},
					},
				},
				Throttling: &config.Throttling{
					Enabled:     true,
					MaxRequests: 1,
					Window:      10,
				},
			},
		},
	}

	server, err := NewServer(config, 8080)
	require.NoError(t, err)

	handler := server.requestHandlerFunc()

	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}
	ctx.Request.SetRequestURI("/webhooks/v1alpha2/test")
	ctx.Request.Header.SetMethod("POST")
	ctx.Request.SetBody([]byte(`{"test": "data"}`))

	// First request should succeed
	handler(ctx.RequestCtx)
	assert.Equal(t, fasthttp.StatusNoContent, ctx.Response.StatusCode())

	// Second request should hit rate limit
	handler(ctx.RequestCtx)
	assert.Equal(t, fasthttp.StatusTooManyRequests, ctx.Response.StatusCode())
}

func TestServer_Shutdown(t *testing.T) {
	require.NotPanics(t, func() {
		server := &Server{}
		server.Shutdown(context.Background())
	})

	server, err := NewServer(setupMinimalConfig(), 8080)
	require.NoError(t, err)

	// Test shutdown without starting
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = server.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestServer_Shutdown_WithTimeout(t *testing.T) {
	server, err := NewServer(setupMinimalConfig(), 8080)
	require.NoError(t, err)

	// Create a context that expires immediately
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for context to expire
	time.Sleep(1 * time.Millisecond)

	err = server.Shutdown(ctx)
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestVersionInfo(t *testing.T) {
	info := VersionInfo()

	assert.Contains(t, info, "version")
	assert.Contains(t, info, "commit")
	assert.Contains(t, info, "buildDate")
	assert.Contains(t, info, "goVersion")
	assert.Contains(t, info, "goOS")
	assert.Contains(t, info, "goArch")

	assert.NotEmpty(t, info["version"])
	assert.NotEmpty(t, info["goVersion"])
}

func TestBuildInfo(t *testing.T) {
	info := BuildInfo()

	assert.Contains(t, info, "webhooked")
	assert.Contains(t, info, Version)
	assert.Contains(t, info, "commit:")
	assert.Contains(t, info, "built:")
	assert.Contains(t, info, "go:")
}

func TestGetRaeLimitStats_Disabled(t *testing.T) {
	server, err := NewServer(&config.Config{
		APIVersion: config.APIVersionV1Alpha2,
		Kind:       config.KindConfiguration,
		Specs:      []*config.Spec{},
	}, 8080)
	require.NoError(t, err)

	assert.Equal(t, map[string]any{"enabled": false}, server.GetRateLimitStats())
}

func TestGetRaeLimitStats_Enabled(t *testing.T) {
	server, err := NewServer(&config.Config{
		APIVersion: config.APIVersionV1Alpha2,
		Kind:       config.KindConfiguration,
		Specs: []*config.Spec{
			{
				Throttling: &config.Throttling{
					Enabled: true,
				},
			},
		},
	}, 8080)
	require.NoError(t, err)

	stats := server.GetRateLimitStats()
	assert.Equal(t, map[string]any{
		"enabled":        true,
		"active_clients": 0,
		"burst_limit":    0,
		"burst_window":   0,
		"max_requests":   0,
		"total_requests": 0,
		"window_seconds": 0,
	}, stats)
}

// Helper function to setup minimal configuration for testing
func setupMinimalConfig() *config.Config {
	return &config.Config{
		APIVersion: config.APIVersionV1Alpha2,
		Kind:       config.KindConfiguration,
		Specs: []*config.Spec{
			{
				Webhooks: []*config.Webhook{},
			},
		},
	}
}

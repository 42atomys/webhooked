//go:build unit

package webhooked

import (
	"errors"
	"testing"

	"github.com/42atomys/webhooked/internal/fasthttpz"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestErrHTTPNotFound(t *testing.T) {
	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}
	testErr := errors.New("test error")

	returnedErr := ErrHTTPNotFound(ctx, testErr)

	assert.Equal(t, fasthttp.StatusNotFound, ctx.Response.StatusCode())
	assert.Equal(t, notFound, ctx.Response.Body())
	assert.ErrorIs(t, returnedErr, testErr)
}

func TestErrHTTPUnauthorized(t *testing.T) {
	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}
	testErr := errors.New("unauthorized error")

	returnedErr := ErrHTTPUnauthorized(ctx, testErr)

	assert.Equal(t, fasthttp.StatusUnauthorized, ctx.Response.StatusCode())
	assert.Equal(t, unauthorized, ctx.Response.Body())
	assert.ErrorIs(t, returnedErr, testErr)
}

func TestErrHTTPInternalServerError(t *testing.T) {
	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}
	testErr := errors.New("internal server error")

	returnedErr := ErrHTTPInternalServerError(ctx, testErr)

	assert.Equal(t, fasthttp.StatusInternalServerError, ctx.Response.StatusCode())
	assert.Equal(t, internalServerError, ctx.Response.Body())
	assert.ErrorIs(t, returnedErr, testErr)
}

func TestErrHTTPBadRequest(t *testing.T) {
	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}
	testErr := errors.New("bad request error")

	returnedErr := ErrHTTPBadRequest(ctx, testErr)

	assert.Equal(t, fasthttp.StatusBadRequest, ctx.Response.StatusCode())
	assert.Equal(t, badRequest, ctx.Response.Body())
	assert.ErrorIs(t, returnedErr, testErr)
}

func TestErrHTTPNotFound_WithNilError(t *testing.T) {
	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}

	returnedErr := ErrHTTPNotFound(ctx, nil)

	assert.Equal(t, fasthttp.StatusNotFound, ctx.Response.StatusCode())
	assert.Equal(t, notFound, ctx.Response.Body())
	assert.Nil(t, returnedErr)
}

func TestErrHTTPUnauthorized_WithNilError(t *testing.T) {
	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}

	returnedErr := ErrHTTPUnauthorized(ctx, nil)

	assert.Equal(t, fasthttp.StatusUnauthorized, ctx.Response.StatusCode())
	assert.Equal(t, unauthorized, ctx.Response.Body())
	assert.Nil(t, returnedErr)
}

func TestErrorConstants(t *testing.T) {
	// Test that our error message constants are reasonable
	assert.Equal(t, []byte("Not Found"), notFound)
	assert.Equal(t, []byte("Internal Server Error"), internalServerError)
	assert.Equal(t, []byte("Unauthorized"), unauthorized)
	assert.Equal(t, []byte("Bad Request"), badRequest)
}

func TestMultipleErrorCalls(t *testing.T) {
	// Test that multiple error calls on the same context work correctly
	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}

	// First error
	err1 := ErrHTTPBadRequest(ctx, errors.New("first error"))
	assert.Error(t, err1)
	assert.Equal(t, "bad request: first error", err1.Error())
	assert.Equal(t, fasthttp.StatusBadRequest, ctx.Response.StatusCode())

	// Second error (should overwrite)
	err2 := ErrHTTPInternalServerError(ctx, errors.New("second error"))
	assert.Error(t, err2)
	assert.Equal(t, "internal server error: second error", err2.Error())
	assert.Equal(t, fasthttp.StatusInternalServerError, ctx.Response.StatusCode())
	assert.Equal(t, internalServerError, ctx.Response.Body())
}

// Benchmarks

func BenchmarkErrHTTPNotFound(b *testing.B) {
	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}
	testErr := errors.New("test error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.Response.Reset()
		ErrHTTPNotFound(ctx, testErr) // nolint:errcheck
	}
}

func BenchmarkErrHTTPUnauthorized(b *testing.B) {
	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}
	testErr := errors.New("unauthorized error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.Response.Reset()
		ErrHTTPUnauthorized(ctx, testErr) // nolint:errcheck
	}
}

func BenchmarkErrHTTPInternalServerError(b *testing.B) {
	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}
	testErr := errors.New("internal server error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.Response.Reset()
		ErrHTTPInternalServerError(ctx, testErr) // nolint:errcheck
	}
}

func BenchmarkErrHTTPBadRequest(b *testing.B) {
	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}
	testErr := errors.New("bad request error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.Response.Reset()
		ErrHTTPBadRequest(ctx, testErr) // nolint:errcheck
	}
}

func BenchmarkAllErrorFunctions(b *testing.B) {
	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}
	testErr := errors.New("test error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.Response.Reset()
		switch i % 4 {
		case 0:
			ErrHTTPNotFound(ctx, testErr) // nolint:errcheck
		case 1:
			ErrHTTPUnauthorized(ctx, testErr) // nolint:errcheck
		case 2:
			ErrHTTPInternalServerError(ctx, testErr) // nolint:errcheck
		case 3:
			ErrHTTPBadRequest(ctx, testErr) // nolint:errcheck
		}
	}
}

func BenchmarkErrorWithNil(b *testing.B) {
	ctx := &fasthttpz.RequestCtx{RequestCtx: &fasthttp.RequestCtx{}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.Response.Reset()
		ErrHTTPNotFound(ctx, nil) // nolint:errcheck
	}
}

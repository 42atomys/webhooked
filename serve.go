package webhooked

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/42atomys/webhooked/internal/config"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/pprofhandler"
	"github.com/valyala/fasthttp/reuseport"
)

func Serve(port int) {
	executor := NewExecutor()

	// Use fasthttp.Server with optimized settings for high concurrency
	server := &fasthttp.Server{
		Handler:            requestHandlerFunc(executor),
		ReadBufferSize:     8192,            // Increase buffer size to handle larger requests
		WriteBufferSize:    8192,            // Increase buffer size to handle larger responses
		MaxConnsPerIP:      0,               // No limit on connections per IP
		MaxRequestsPerConn: 0,               // No limit on requests per connection
		Concurrency:        100000,          // Allow high concurrency
		IdleTimeout:        5 * time.Second, // Timeout to close idle connections
	}

	listener, err := reuseport.Listen("tcp4", fmt.Sprint(":", port))
	if err != nil {
		panic(err)
	}

	if err := server.Serve(listener); err != nil {
		panic(err)
	}
}

func requestHandlerFunc(executor Executor) fasthttp.RequestHandler {
	return func(rctx *fasthttp.RequestCtx) {
		log.Debug().Msgf("Incoming request: %s", rctx.Path())

		start := rctx.Time()
		path := rctx.Path()

		if bytes.HasPrefix(path, config.WebhooksEndpointPrefix()) {
			// Use context with background for goroutine safety since there are additional goroutines in IncomingRequest
			ctx := context.Background()

			err := executor.IncomingRequest(ctx, rctx)

			if err != nil && rctx.Response.StatusCode() == fasthttp.StatusOK {
				rctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
				rctx.SetBody(internalServerError)
				_ = ErrHTTPInternalServerError(rctx, fmt.Errorf("error processing incoming request: %w", err))
			}
			log.Debug().Msgf("Request processed in %v", time.Since(start))
			return
		}

		pprofhandler.PprofHandler(rctx)
	}
}

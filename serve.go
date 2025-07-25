package webhooked

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/42atomys/webhooked/internal/config"
	"github.com/42atomys/webhooked/internal/fasthttpz"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/pprofhandler"
	"github.com/valyala/fasthttp/reuseport"
)

// Server represents the webhooked HTTP server
type Server struct {
	config      *config.Config
	port        int
	server      *fasthttp.Server
	listener    net.Listener
	executor    Executor
	rateLimiter *RateLimiter
}

// NewServer creates a new Server instance
func NewServer(config *config.Config, port int) (*Server, error) {
	executor := NewExecutor(config)

	// Use fasthttp.Server with optimized settings for high concurrency
	server := &fasthttp.Server{
		ReadBufferSize:     8192,             // Increase buffer size to handle larger requests
		WriteBufferSize:    8192,             // Increase buffer size to handle larger responses
		MaxConnsPerIP:      0,                // No limit on connections per IP
		MaxRequestsPerConn: 0,                // No limit on requests per connection
		Concurrency:        100000,           // Allow high concurrency
		IdleTimeout:        5 * time.Second,  // Timeout to close idle connections
		ReadTimeout:        30 * time.Second, // Request read timeout
		WriteTimeout:       30 * time.Second, // Response write timeout
	}

	s := &Server{
		config:   config,
		port:     port,
		server:   server,
		executor: executor,
	}

	// Initialize rate limiter when configuration is available
	// TODO: Make ratelimiter works correctly
	s.initializeRateLimiter()

	// Set the handler
	server.Handler = s.requestHandlerFunc()

	return s, nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	listener, err := reuseport.Listen("tcp4", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to create listener on port %d: %w", s.port, err)
	}

	s.listener = listener

	log.Info().Int("port", s.port).Msg("server listening")

	if err := s.server.Serve(listener); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	log.Info().Msg("shutting down HTTP server...")

	// Shutdown the server with context timeout
	done := make(chan error, 1)
	go func() {
		done <- s.server.Shutdown()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// requestHandlerFunc returns the HTTP request handler for the server
func (s *Server) requestHandlerFunc() fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		rctx := &fasthttpz.RequestCtx{ctx}
		log.Debug().Msgf("Incoming request: %s", rctx.Path())

		start := rctx.Time()
		path := rctx.Path()

		// Health check endpoints
		if bytes.Equal(path, []byte("/health")) {
			s.handleHealthCheck(rctx)
			return
		}

		if bytes.Equal(path, []byte("/ready")) {
			s.handleReadinessCheck(rctx)
			return
		}

		if bytes.HasPrefix(path, config.WebhooksEndpointPrefix()) {
			// Check rate limiting
			clientIP := string(rctx.RemoteIP())
			if s.rateLimiter != nil && !s.rateLimiter.Allow(clientIP) {
				rctx.Response.SetStatusCode(fasthttp.StatusTooManyRequests)
				rctx.SetContentType("application/json")
				rctx.SetBody([]byte(`{"error":"rate limit exceeded","client_ip":"` + clientIP + `"}`))
				log.Warn().Str("client_ip", clientIP).Msg("rate limit exceeded")
				return
			}

			// Create context with timeout for webhook processing
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			err := s.executor.IncomingRequest(ctx, rctx)

			if err != nil && rctx.Response.StatusCode() == fasthttp.StatusOK {
				// Check if the error was due to context timeout
				if errors.Is(err, context.DeadlineExceeded) {
					rctx.Response.SetStatusCode(fasthttp.StatusRequestTimeout)
					rctx.SetBody([]byte("Request Timeout"))
					_ = ErrHTTPInternalServerError(rctx, fmt.Errorf("request timeout: %w", err))
				} else {
					rctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
					rctx.SetBody(internalServerError)
					_ = ErrHTTPInternalServerError(rctx, fmt.Errorf("error processing incoming request: %w", err))
				}
			}
			log.Debug().Msgf("Request processed in %v", time.Since(start))
			return
		}

		pprofhandler.PprofHandler(rctx.RequestCtx)
	}
}

// handleHealthCheck handles the /health endpoint
func (s *Server) handleHealthCheck(rctx *fasthttpz.RequestCtx) {
	rctx.SetStatusCode(fasthttp.StatusOK)
	rctx.SetContentType("application/json")
	rctx.SetBody([]byte(`{"status":"healthy","version":"` + Version + `"}`))
}

// handleReadinessCheck handles the /ready endpoint
func (s *Server) handleReadinessCheck(rctx *fasthttpz.RequestCtx) {
	// Check if configuration is loaded
	if s.config == nil || len(s.config.Specs) == 0 {
		rctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
		rctx.SetContentType("application/json")
		rctx.SetBody([]byte(`{"status":"not ready","reason":"no configuration loaded"}`))
		return
	}

	rctx.SetStatusCode(fasthttp.StatusOK)
	rctx.SetContentType("application/json")
	rctx.SetBody([]byte(`{"status":"ready","version":"` + Version + `"}`))
}

// initializeRateLimiter initializes the rate limiter based on configuration
func (s *Server) initializeRateLimiter() {
	if s.config == nil || len(s.config.Specs) == 0 {
		return
	}

	// Use throttling configuration from the first spec
	// In a more advanced implementation, you might want to support per-webhook throttling
	for _, spec := range s.config.Specs {
		if spec.Throttling != nil && spec.Throttling.Enabled {
			s.rateLimiter = NewRateLimiter(spec.Throttling)
			s.rateLimiter.StartCleanupRoutine()

			log.Info().
				Int("max_requests", spec.Throttling.MaxRequests).
				Int("window_seconds", spec.Throttling.Window).
				Int("burst", spec.Throttling.Burst).
				Int("burst_window", spec.Throttling.BurstWindow).
				Msg("rate limiter initialized")
			break
		}
	}
}

// GetRateLimitStats returns current rate limiting statistics
func (s *Server) GetRateLimitStats() map[string]any {
	if s.rateLimiter == nil {
		return map[string]any{
			"enabled": false,
		}
	}
	return s.rateLimiter.GetStats()
}

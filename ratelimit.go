package webhooked

import (
	"sync"
	"time"

	"github.com/42atomys/webhooked/internal/config"
	"github.com/rs/zerolog/log"
)

// RateLimiter implements a rate limiting mechanism based on sliding window
type RateLimiter struct {
	mu       sync.RWMutex
	windows  map[string]*Window
	throttle *config.Throttling
}

// Window represents a time window for rate limiting
type Window struct {
	requests []time.Time
	mu       sync.RWMutex
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter(throttle *config.Throttling) *RateLimiter {
	return &RateLimiter{
		windows:  make(map[string]*Window),
		throttle: throttle,
	}
}

// Allow checks if a request should be allowed based on rate limiting rules
func (rl *RateLimiter) Allow(clientIP string) bool {
	if rl.throttle == nil || !rl.throttle.Enabled {
		return true
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	window, exists := rl.windows[clientIP]
	if !exists {
		window = &Window{
			requests: make([]time.Time, 0),
		}
		rl.windows[clientIP] = window
	}

	return rl.checkWindow(window, clientIP)
}

// checkWindow checks if the request fits within the rate limiting window
func (rl *RateLimiter) checkWindow(window *Window, clientIP string) bool {
	window.mu.Lock()
	defer window.mu.Unlock()

	now := time.Now()
	windowDuration := time.Duration(rl.throttle.Window) * time.Second

	// Clean old requests outside the window
	cutoff := now.Add(-windowDuration)
	window.requests = rl.filterRequests(window.requests, cutoff)

	// Check if we're within the rate limit
	if len(window.requests) >= rl.throttle.MaxRequests {
		log.Debug().
			Str("client_ip", clientIP).
			Int("current_requests", len(window.requests)).
			Int("max_requests", rl.throttle.MaxRequests).
			Msg("rate limit exceeded")
		return false
	}

	// Check burst limits if configured
	if rl.throttle.Burst > 0 && rl.throttle.BurstWindow > 0 {
		burstDuration := time.Duration(rl.throttle.BurstWindow) * time.Second
		burstCutoff := now.Add(-burstDuration)
		recentRequests := rl.filterRequests(window.requests, burstCutoff)

		if len(recentRequests) >= rl.throttle.Burst {
			log.Debug().
				Str("client_ip", clientIP).
				Int("current_burst", len(recentRequests)).
				Int("max_burst", rl.throttle.Burst).
				Msg("burst limit exceeded")
			return false
		}
	}

	// Add current request to window
	window.requests = append(window.requests, now)

	log.Debug().
		Str("client_ip", clientIP).
		Int("requests_in_window", len(window.requests)).
		Msg("request allowed")

	return true
}

// filterRequests removes requests older than the cutoff time
func (rl *RateLimiter) filterRequests(requests []time.Time, cutoff time.Time) []time.Time {
	filtered := requests[:0] // Reuse slice capacity
	for _, req := range requests {
		if req.After(cutoff) {
			filtered = append(filtered, req)
		}
	}
	return filtered
}

// Cleanup removes old windows for clients that haven't made requests recently
func (rl *RateLimiter) Cleanup() {
	if rl.throttle == nil || !rl.throttle.Enabled {
		return
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cleanupDuration := time.Duration(rl.throttle.Window*2) * time.Second
	cutoff := now.Add(-cleanupDuration)

	for clientIP, window := range rl.windows {
		window.mu.RLock()
		lastRequest := time.Time{}
		if len(window.requests) > 0 {
			lastRequest = window.requests[len(window.requests)-1]
		}
		window.mu.RUnlock()

		if lastRequest.Before(cutoff) {
			delete(rl.windows, clientIP)
			log.Debug().
				Str("client_ip", clientIP).
				Msg("cleaned up rate limit window")
		}
	}
}

// StartCleanupRoutine starts a background goroutine to clean up old windows
func (rl *RateLimiter) StartCleanupRoutine() {
	if rl.throttle == nil || !rl.throttle.Enabled {
		return
	}

	cleanupInterval := time.Duration(rl.throttle.Window) * time.Second
	if cleanupInterval < time.Minute {
		cleanupInterval = time.Minute
	}

	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()

		for range ticker.C {
			rl.Cleanup()
		}
	}()

	log.Info().
		Dur("cleanup_interval", cleanupInterval).
		Msg("rate limiter cleanup routine started")
}

// GetStats returns current rate limiting statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	if rl.throttle == nil || !rl.throttle.Enabled {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	rl.mu.RLock()
	defer rl.mu.RUnlock()

	activeClients := len(rl.windows)
	totalRequests := 0

	for _, window := range rl.windows {
		window.mu.RLock()
		totalRequests += len(window.requests)
		window.mu.RUnlock()
	}

	return map[string]interface{}{
		"enabled":        true,
		"active_clients": activeClients,
		"total_requests": totalRequests,
		"max_requests":   rl.throttle.MaxRequests,
		"window_seconds": rl.throttle.Window,
		"burst_limit":    rl.throttle.Burst,
		"burst_window":   rl.throttle.BurstWindow,
	}
}

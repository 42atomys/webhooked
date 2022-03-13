package server

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"

	"atomys.codes/webhooked/internal/config"
)

//statusRecorder to record the status code from the ResponseWriter
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

var (
	// versionAndEndpointRegexp is a regexp to extract the version and endpoint from the given path
	versionAndEndpointRegexp = regexp.MustCompile(`(?m)/(?P<version>v[0-9a-z]+)(?P<endpoint>/.+)`)
	// responseTimeHistogram is a histogram of response times
	// used to export the response time to Prometheus
	responseTimeHistogram *prometheus.HistogramVec = promauto.
				NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "webhooked",
			Name:      "http_server_request_duration_seconds",
			Help:      "Histogram of response time for handler in seconds",
		}, []string{"method", "status_code", "version", "spec", "secure"})
)

// WriteHeader sets the status code for the response
func (rec *statusRecorder) WriteHeader(statusCode int) {
	rec.statusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

// prometheusMiddleware is a middleware that records the response time and
// exports it to Prometheus metrics for the given request
// Example:
// webhooked_http_server_request_duration_seconds_count{method="POST",secure="false",spec="exampleHook",status_code="200",version="v1alpha1"} 1
func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := statusRecorder{w, 200}

		next.ServeHTTP(&rec, r)

		pp := getVersionAndEndpoint(r.URL.Path)
		spec, err := config.Current().GetSpecByEndpoint(pp["endpoint"])
		if err != nil {
			return
		}

		duration := time.Since(start)
		statusCode := strconv.Itoa(rec.statusCode)
		responseTimeHistogram.WithLabelValues(r.Method, statusCode, pp["version"], spec.Name, fmt.Sprintf("%t", spec.HasSecurity())).Observe(duration.Seconds())
	})
}

// loggingMiddleware is a middleware that logs the request and response
// Example:
// INF Webhook is processed duration="586µs" secure=false spec=exampleHook statusCode=200 version=v1alpha1
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := statusRecorder{w, 200}

		next.ServeHTTP(&rec, r)

		var logEvent = log.Info().
			Str("duration", time.Since(start).String()).
			Int("statusCode", rec.statusCode)

		pp := getVersionAndEndpoint(r.URL.Path)
		spec, _ := config.Current().GetSpecByEndpoint(pp["endpoint"])
		if spec != nil {
			logEvent.Str("version", pp["version"]).Str("spec", spec.Name).Bool("secure", spec.HasSecurity()).Msgf("Webhook is processed")
		}
	})
}

// getVersionAndEndpoint returns the version and endpoint from the given path
// Example: /v0/webhooks/example
// Returns: {"version": "v0", "endpoint": "/webhooks/example"}
func getVersionAndEndpoint(path string) map[string]string {
	match := versionAndEndpointRegexp.FindStringSubmatch(path)
	result := make(map[string]string)
	for i, name := range versionAndEndpointRegexp.SubexpNames() {
		if i != 0 && i <= len(match) && name != "" {
			result[name] = match[i]
		}
	}

	return result
}

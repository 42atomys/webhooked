package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

//statusRecorder to record the status code from the ResponseWriter
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

var (
	responseTimeHistogram *prometheus.HistogramVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "webhooked",
		Name:      "http_server_request_duration_seconds",
		Help:      "Histogram of response time for handler in seconds",
	}, []string{"route", "method", "status_code"})
)

func (rec *statusRecorder) WriteHeader(statusCode int) {
	rec.statusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

func prometheusMiddleware(next http.Handler) http.Handler {
	prometheus.MustRegister(responseTimeHistogram)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := statusRecorder{w, 200}

		next.ServeHTTP(&rec, r)

		duration := time.Since(start)
		statusCode := strconv.Itoa(rec.statusCode)
		route := getRoutePattern(r)
		responseTimeHistogram.WithLabelValues(route, r.Method, statusCode).Observe(duration.Seconds())
	})
}

// getRoutePattern returns the route pattern from the chi context there are 3 conditions
// a) static routes "/example" => "/example"
// b) dynamic routes "/example/:id" => "/example/{id}"
// c) if nothing matches the output is undefined
func getRoutePattern(r *http.Request) string {
	if currentRoute := mux.CurrentRoute(r); currentRoute != nil {
		if pattern := currentRoute.GetName(); pattern != "" {
			return pattern
		}
	}
	return "undefined"
}

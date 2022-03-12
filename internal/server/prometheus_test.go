package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/suite"
)

type testSuitePrometheusMiddleware struct {
	suite.Suite
	httpHandler http.Handler
}

func (suite *testSuitePrometheusMiddleware) BeforeTest(suiteName, testName string) {
	suite.httpHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})
}

func TestPrometheusMiddleware(t *testing.T) {
	suite.Run(t, new(testSuitePrometheusMiddleware))
}

func (suite *testSuitePrometheusMiddleware) TestRun() {
	handler := prometheusMiddleware(suite.httpHandler)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	suite.Equal(http.StatusAccepted, w.Code)
	suite.Equal(1, testutil.CollectAndCount(responseTimeHistogram))
}

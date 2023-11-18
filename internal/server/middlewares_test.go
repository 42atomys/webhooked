package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/suite"

	"atomys.codes/webhooked/internal/config"
)

func init() {
	if err := config.Load("../../tests/webhooks.tests.yaml"); err != nil {
		panic(err)
	}
}

type testSuiteMiddlewares struct {
	suite.Suite
	httpHandler http.Handler
}

func (suite *testSuiteMiddlewares) BeforeTest(suiteName, testName string) {
	suite.httpHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})
}

func TestLoggingMiddleware(t *testing.T) {
	suite.Run(t, new(testSuiteMiddlewares))
}

func (suite *testSuiteMiddlewares) TestLogging() {
	handler := loggingMiddleware(suite.httpHandler)

	req := httptest.NewRequest(http.MethodGet, "/v0/webhooks/example", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	suite.Equal(http.StatusAccepted, w.Code)
}

func (suite *testSuiteMiddlewares) TestPrometheus() {
	handler := prometheusMiddleware(suite.httpHandler)

	req := httptest.NewRequest(http.MethodGet, "/v0/webhooks/example", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	suite.Equal(http.StatusAccepted, w.Code)
	suite.Equal(1, testutil.CollectAndCount(responseTimeHistogram))
}

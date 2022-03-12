package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"

	"atomys.codes/webhooked/internal/config"
)

func init() {
	viper.SetConfigName("webhooks.tests")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../../tests")
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	if err := config.Load(); err != nil {
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

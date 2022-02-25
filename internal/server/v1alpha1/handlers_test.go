package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"42stellar.org/webhooks/internal/config"
	"42stellar.org/webhooks/pkg/factory"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	var s = NewServer()
	assert.NotNil(t, s)
	assert.Equal(t, "v1alpha1", s.Version())
	assert.Equal(t, config.Current(), s.config)
}

func TestServer_Version(t *testing.T) {
	var s = &Server{}
	assert.Equal(t, "v1alpha1", s.Version())
}

func TestServer_WebhookHandler(t *testing.T) {
	assert.Equal(t,
		http.StatusBadRequest,
		testServer_WebhookHandler_Helper(t, &Server{config: &config.Configuration{APIVersion: "invalidVersion"}}).Code,
	)

	assert.Equal(t,
		http.StatusNotFound,
		testServer_WebhookHandler_Helper(t, &Server{config: &config.Configuration{APIVersion: "v1alpha1"}}).Code,
	)

	var expectedError = errors.New("err during processing webhook")
	assert.Equal(t,
		http.StatusInternalServerError,
		testServer_WebhookHandler_Helper(t, &Server{
			config: &config.Configuration{
				APIVersion: "v1alpha1",
				Specs: []*config.WebhookSpec{
					{
						Name:          "test",
						EntrypointURL: "/test",
					}},
			},
			webhookService: func(s *Server, spec *config.WebhookSpec, r *http.Request) error { return expectedError },
		}).Code,
	)

	assert.Equal(t,
		http.StatusOK,
		testServer_WebhookHandler_Helper(t, &Server{
			config: &config.Configuration{
				APIVersion: "v1alpha1",
				Specs: []*config.WebhookSpec{
					{
						Name:          "test",
						EntrypointURL: "/test",
					}},
			},
			webhookService: func(s *Server, spec *config.WebhookSpec, r *http.Request) error { return nil },
		}).Code,
	)

	assert.Equal(t,
		http.StatusForbidden,
		testServer_WebhookHandler_Helper(t, &Server{
			config: &config.Configuration{
				APIVersion: "v1alpha1",
				Specs: []*config.WebhookSpec{
					{
						Name:          "test",
						EntrypointURL: "/test",
					}},
			},
			webhookService: func(s *Server, spec *config.WebhookSpec, r *http.Request) error { return factory.ErrSecurityFailed },
		}).Code,
	)

	assert.Equal(t,
		http.StatusBadRequest,
		testServer_WebhookHandler_Helper(t, &Server{
			config: &config.Configuration{
				APIVersion: "v0test",
				Specs: []*config.WebhookSpec{
					{
						Name:          "test",
						EntrypointURL: "/test",
					}},
			},
			webhookService: func(s *Server, spec *config.WebhookSpec, r *http.Request) error { return nil },
		}).Code,
	)
}

func testServer_WebhookHandler_Helper(t *testing.T, server *Server) *httptest.ResponseRecorder {
	server.logger = log.With().Str("apiVersion", server.Version()).Logger()

	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("POST", "/v1alpha1/test", strings.NewReader("Hello"))
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	server.WebhookHandler().ServeHTTP(rr, req)

	return rr
}

func Test_webhookService(t *testing.T) {
	assert := assert.New(t)

	req, err := http.NewRequest("POST", "/v1alpha1/test", strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Token", "test")

	var tests = []struct {
		name    string
		input   *config.WebhookSpec
		wantErr bool
	}{
		{"no spec", nil, true},
		{"no security", &config.WebhookSpec{
			Security: nil,
		}, false},
		{"empty security", &config.WebhookSpec{
			SecurityFactories: make([]*factory.Factory, 0),
		}, false},
		{"one invalid security", &config.WebhookSpec{
			SecurityFactories: []*factory.Factory{
				{
					Name: "test",
					Fn: func(configRaw map[string]interface{}, lastOuput string, inputs ...interface{}) (string, error) {
						return "", nil
					},
				},
			},
		}, true},
		{"valid security", &config.WebhookSpec{
			SecurityFactories: []*factory.Factory{
				{
					Name: "getHeader",
					Fn: func(configRaw map[string]interface{}, lastOuput string, inputs ...interface{}) (string, error) {
						return inputs[0].(http.Header).Get("X-Token"), nil
					},
				},
				{
					Name: "compareWithStaticValue",
					Fn: func(configRaw map[string]interface{}, lastOuput string, inputs ...interface{}) (string, error) {
						return "t", nil
					},
				},
			},
		}, false},
	}

	for _, test := range tests {
		got := webhookService(&Server{}, test.input, req)
		if test.wantErr {
			assert.Error(got, "input: %s", test.name)
		} else {
			assert.NoError(got, "input: %s", test.name)
		}
	}
}

func TestServer_runSecurity(t *testing.T) {
	assert := assert.New(t)
	var s = &Server{}

	req, err := http.NewRequest("POST", "/v1alpha1/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Token", "test")

	var tests = []struct {
		name    string
		input   *config.WebhookSpec
		wantErr bool
	}{
		{"no spec", nil, true},
		{"no security", &config.WebhookSpec{
			Security: nil,
		}, false},
		{"empty security", &config.WebhookSpec{
			SecurityFactories: make([]*factory.Factory, 0),
		}, false},
		{"one invalid security", &config.WebhookSpec{
			SecurityFactories: []*factory.Factory{
				{
					Name: "test",
					Fn: func(configRaw map[string]interface{}, lastOuput string, inputs ...interface{}) (string, error) {
						return "", nil
					},
				},
			},
		}, true},
		{"valid security", &config.WebhookSpec{
			SecurityFactories: []*factory.Factory{
				{
					Name: "getHeader",
					Fn: func(configRaw map[string]interface{}, lastOuput string, inputs ...interface{}) (string, error) {
						return inputs[0].(http.Header).Get("X-Token"), nil
					},
				},
				{
					Name: "compareWithStaticValue",
					Fn: func(configRaw map[string]interface{}, lastOuput string, inputs ...interface{}) (string, error) {
						return "t", nil
					},
				},
			},
		}, false},
		{"invalid security forced", &config.WebhookSpec{
			SecurityFactories: []*factory.Factory{
				{
					Name: "getHeader",
					Fn: func(configRaw map[string]interface{}, lastOuput string, inputs ...interface{}) (string, error) {
						return inputs[0].(http.Header).Get("X-Token"), nil
					},
				},
				{
					Name: "compareWithStaticValue",
					Fn: func(configRaw map[string]interface{}, lastOuput string, inputs ...interface{}) (string, error) {
						return "f", nil
					},
				},
			},
		}, true},
	}

	for _, test := range tests {
		got := s.runSecurity(test.input, req)
		if test.wantErr {
			assert.Error(got, "input: %s", test.name)
		} else {
			assert.NoError(got, "input: %s", test.name)
		}
	}
}

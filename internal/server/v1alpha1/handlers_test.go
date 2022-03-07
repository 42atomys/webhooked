package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"

	"42stellar.org/webhooks/internal/config"
	"42stellar.org/webhooks/internal/valuable"
	"42stellar.org/webhooks/pkg/factory"
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
			webhookService: func(s *Server, spec *config.WebhookSpec, r *http.Request) error { return errSecurityFailed },
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
	req, err := http.NewRequest("POST", "/v1alpha1/test", strings.NewReader("{}"))
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

	headerFactory, _ := factory.GetFactoryByName("header")
	compareFactory, _ := factory.GetFactoryByName("compare")
	validPipeline := factory.NewPipeline().AddFactory(headerFactory).AddFactory(compareFactory)

	req := httptest.NewRequest("POST", "/v1alpha1/test", strings.NewReader("{}"))
	req.Header.Set("X-Token", "test")
	validPipeline.Inputs["request"] = req
	validPipeline.Inputs["headerName"] = &factory.InputConfig{Name: "headerName", Valuable: valuable.Valuable{Values: []string{"X-Token"}}}
	validPipeline.Inputs["first"] = &factory.InputConfig{Name: "headerName", Valuable: valuable.Valuable{Values: []string{"{{ .Outputs.header.value }}"}}}
	validPipeline.Inputs["second"] = &factory.InputConfig{Name: "headerName", Valuable: valuable.Valuable{Values: []string{"test"}}}

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
			SecurityPipeline: factory.NewPipeline(),
		}, false},

		{"valid security", &config.WebhookSpec{
			SecurityPipeline: validPipeline,
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

	headerFactory, _ := factory.GetFactoryByName("header")
	compareFactory, _ := factory.GetFactoryByName("compare")
	validPipeline := factory.NewPipeline().AddFactory(headerFactory).AddFactory(compareFactory)

	req := httptest.NewRequest("POST", "/v1alpha1/test", strings.NewReader("{}"))
	req.Header.Set("X-Token", "test")
	validPipeline.Inputs["request"] = req
	validPipeline.Inputs["headerName"] = &factory.InputConfig{Name: "headerName", Valuable: valuable.Valuable{Values: []string{"X-Token"}}}
	validPipeline.Inputs["first"] = &factory.InputConfig{Name: "headerName", Valuable: valuable.Valuable{Values: []string{"{{ .Outputs.header.value }}"}}}
	validPipeline.Inputs["second"] = &factory.InputConfig{Name: "headerName", Valuable: valuable.Valuable{Values: []string{"test"}}}

	var tests = []struct {
		name    string
		input   *config.WebhookSpec
		wantErr bool
	}{
		{"no spec", nil, true},
		{"no security", &config.WebhookSpec{
			Security: nil,
		}, true},
		{"empty security", &config.WebhookSpec{
			SecurityPipeline: factory.NewPipeline(),
		}, true},

		{"valid security", &config.WebhookSpec{
			SecurityPipeline: validPipeline,
		}, false},
	}

	for _, test := range tests {
		got := s.runSecurity(test.input, req, []byte("data"))
		if test.wantErr {
			assert.Error(got, "input: %s", test.name)
		} else {
			assert.NoError(got, "input: %s", test.name)
		}
	}
}

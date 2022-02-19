package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"42stellar.org/webhooks/internal/config"
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
}

func testServer_WebhookHandler_Helper(t *testing.T, server *Server) *httptest.ResponseRecorder {

	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("POST", "/v1alpha1/test", nil)
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

	var tests = []struct {
		input    *config.WebhookSpec
		expected error
	}{
		{nil, config.ErrSpecNotFound},
		{&config.WebhookSpec{
			Security: nil,
		}, nil},
		{&config.WebhookSpec{
			SecurityFactories: make([]*factory.Factory, 0),
		}, nil},
	}

	for _, test := range tests {
		assert.Equal(webhookService(&Server{}, test.input, nil), test.expected, "input: %d", test.input)
	}
}

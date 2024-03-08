package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"

	"atomys.codes/webhooked/internal/config"
	"atomys.codes/webhooked/internal/valuable"
	"atomys.codes/webhooked/pkg/factory"
	"atomys.codes/webhooked/pkg/storage"
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
		testServerWebhookHandlerHelper(t, &Server{config: &config.Configuration{APIVersion: "invalidVersion"}}).Code,
	)

	assert.Equal(t,
		http.StatusNotFound,
		testServerWebhookHandlerHelper(t, &Server{config: &config.Configuration{APIVersion: "v1alpha1"}}).Code,
	)

	var expectedError = errors.New("err during processing webhook")
	assert.Equal(t,
		http.StatusInternalServerError,
		testServerWebhookHandlerHelper(t, &Server{
			config: &config.Configuration{
				APIVersion: "v1alpha1",
				Specs: []*config.WebhookSpec{
					{
						Name:          "test",
						EntrypointURL: "/test",
					}},
			},
			webhookService: func(s *Server, spec *config.WebhookSpec, r *http.Request) (string, error) { return "", expectedError },
		}).Code,
	)

	assert.Equal(t,
		http.StatusOK,
		testServerWebhookHandlerHelper(t, &Server{
			config: &config.Configuration{
				APIVersion: "v1alpha1",
				Specs: []*config.WebhookSpec{
					{
						Name:          "test",
						EntrypointURL: "/test",
					}},
			},
			webhookService: func(s *Server, spec *config.WebhookSpec, r *http.Request) (string, error) { return "", nil },
		}).Code,
	)

	assert.Equal(t,
		http.StatusOK,
		testServerWebhookHandlerHelper(t, &Server{
			config: &config.Configuration{
				APIVersion: "v1alpha1",
				Specs: []*config.WebhookSpec{
					{
						Name:          "test",
						EntrypointURL: "/test",
						Response: config.ResponseSpec{
							Formatting:  &config.FormattingSpec{Template: "test-payload"},
							HttpCode:    200,
							ContentType: "application/json",
						},
					}},
			},
			webhookService: func(s *Server, spec *config.WebhookSpec, r *http.Request) (string, error) { return "test-payload", nil },
		}).Code,
	)

	assert.Equal(t,
		http.StatusForbidden,
		testServerWebhookHandlerHelper(t, &Server{
			config: &config.Configuration{
				APIVersion: "v1alpha1",
				Specs: []*config.WebhookSpec{
					{
						Name:          "test",
						EntrypointURL: "/test",
					}},
			},
			webhookService: func(s *Server, spec *config.WebhookSpec, r *http.Request) (string, error) {
				return "", errSecurityFailed
			},
		}).Code,
	)

	assert.Equal(t,
		http.StatusBadRequest,
		testServerWebhookHandlerHelper(t, &Server{
			config: &config.Configuration{
				APIVersion: "v0test",
				Specs: []*config.WebhookSpec{
					{
						Name:          "test",
						EntrypointURL: "/test",
					}},
			},
			webhookService: func(s *Server, spec *config.WebhookSpec, r *http.Request) (string, error) { return "", nil },
		}).Code,
	)
}

func testServerWebhookHandlerHelper(t *testing.T, server *Server) *httptest.ResponseRecorder {
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

	req := httptest.NewRequest("POST", "/v1alpha1/test", strings.NewReader("{}"))
	req.Header.Set("X-Token", "test")

	invalidReq := httptest.NewRequest("POST", "/v1alpha1/test", nil)
	invalidReq.Body = nil

	validPipeline := factory.NewPipeline().AddFactory(headerFactory).AddFactory(compareFactory)
	validPipeline.Inputs["request"] = req
	validPipeline.Inputs["headerName"] = &factory.InputConfig{Name: "headerName", Valuable: valuable.Valuable{Values: []string{"X-Token"}}}
	validPipeline.Inputs["first"] = &factory.InputConfig{Name: "headerName", Valuable: valuable.Valuable{Values: []string{"{{ .Outputs.header.value }}"}}}
	validPipeline.Inputs["second"] = &factory.InputConfig{Name: "headerName", Valuable: valuable.Valuable{Values: []string{"test"}}}

	invalidPipeline := factory.NewPipeline().AddFactory(headerFactory).AddFactory(compareFactory)
	invalidPipeline.Inputs["request"] = req
	invalidPipeline.Inputs["headerName"] = &factory.InputConfig{Name: "headerName", Valuable: valuable.Valuable{Values: []string{"X-Token"}}}
	invalidPipeline.Inputs["first"] = &factory.InputConfig{Name: "headerName", Valuable: valuable.Valuable{Values: []string{"{{ .Outputs.header.value }}"}}}
	invalidPipeline.Inputs["second"] = &factory.InputConfig{Name: "headerName", Valuable: valuable.Valuable{Values: []string{"INVALID"}}}

	type input struct {
		spec *config.WebhookSpec
		req  *http.Request
	}

	var tests = []struct {
		name     string
		input    *input
		wantErr  bool
		matchErr error
	}{
		{"no spec", &input{nil, req}, true, config.ErrSpecNotFound},
		{"no security", &input{&config.WebhookSpec{Security: nil}, req}, false, nil},
		{"empty security", &input{&config.WebhookSpec{
			SecurityPipeline: factory.NewPipeline(),
		}, req}, false, nil},
		{"valid security", &input{&config.WebhookSpec{
			SecurityPipeline: validPipeline,
		}, req}, false, nil},
		{"invalid security", &input{&config.WebhookSpec{
			SecurityPipeline: invalidPipeline,
		}, req}, true, errSecurityFailed},
		{"valid payload with response", &input{
			&config.WebhookSpec{
				SecurityPipeline: validPipeline,
				Response: config.ResponseSpec{
					Formatting:  &config.FormattingSpec{Template: "{{.Payload}}"},
					HttpCode:    200,
					ContentType: "application/json",
				},
			},
			req,
		}, false, nil},
		{"invalid body payload", &input{&config.WebhookSpec{
			SecurityPipeline: validPipeline,
		}, invalidReq}, true, errRequestBodyMissing},
	}

	for _, test := range tests {
		log.Warn().Msgf("body %+v", test.input.req.Body)
		_, got := webhookService(&Server{}, test.input.spec, test.input.req)
		if test.wantErr {
			assert.ErrorIs(got, test.matchErr, "input: %s", test.name)
		} else {
			assert.NoError(got, "input: %s", test.name)
		}
	}
}

func TestServer_webhokServiceStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("TestServer_webhokServiceStorage testing is skiped in short version of test")
		return
	}

	pusher, err := storage.Load("redis", map[string]interface{}{
		"host":     os.Getenv("REDIS_HOST"),
		"port":     os.Getenv("REDIS_PORT"),
		"database": 0,
		"key":      "testKey",
	})
	assert.NoError(t, err)

	var tests = []struct {
		name           string
		req            *http.Request
		templateString string
		wantErr        bool
	}{
		{
			"basic",
			httptest.NewRequest("POST", "/v1alpha1/test", strings.NewReader("{}")),
			"{{ .Payload }}",
			false,
		},
		{
			"invalid template",
			httptest.NewRequest("POST", "/v1alpha1/test", strings.NewReader("{}")),
			"{{ ",
			true,
		},
	}

	for _, test := range tests {
		spec := &config.WebhookSpec{
			Security: nil,
			Storage: []*config.StorageSpec{
				{
					Type: "redis",
					Formatting: &config.FormattingSpec{
						Template: test.templateString,
					},
					Client: pusher,
				},
			},
		}

		_, got := webhookService(&Server{}, spec, test.req)
		if test.wantErr {
			assert.Error(t, got, "input: %s", test.name)
		} else {
			assert.NoError(t, got, "input: %s", test.name)
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

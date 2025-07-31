//go:build unit

package fasthttpz

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

type TestSuiteRequestCtx struct {
	suite.Suite

	requestCtx *RequestCtx
	fastCtx    *fasthttp.RequestCtx
}

func (suite *TestSuiteRequestCtx) BeforeTest(suiteName, testName string) {
	// Create a new fasthttp.RequestCtx for each test
	suite.fastCtx = &fasthttp.RequestCtx{}
	suite.requestCtx = &RequestCtx{RequestCtx: suite.fastCtx}

	// Set up some default request data
	suite.fastCtx.Request.SetRequestURI("https://example.com/webhook?param1=value1&param2=value2")
	suite.fastCtx.Request.Header.SetMethod("POST")
	suite.fastCtx.Request.Header.SetHost("example.com")
	suite.fastCtx.Request.Header.SetUserAgent("test-agent/1.0")
	suite.fastCtx.Request.SetBody([]byte(`{"test": "payload", "data": 123}`))

	// Mock remote address
	addr := &net.TCPAddr{
		IP:   net.ParseIP("192.168.1.100"),
		Port: 12345,
	}
	suite.fastCtx.SetRemoteAddr(addr)
}

func (suite *TestSuiteRequestCtx) TestTemplateContext_AllFields() {
	assert := assert.New(suite.T())

	context := suite.requestCtx.TemplateContext()

	// Verify all expected fields are present
	expectedFields := []string{
		"ConnID", "ConnTime", "Host", "IsTLS", "Method",
		"QueryArgs", "RemoteAddr", "RemoteIP", "RequestTime",
		"URI", "UserAgent", "Request", "Payload",
	}

	for _, field := range expectedFields {
		assert.Contains(context, field, "Context should contain field: %s", field)
	}
}

func (suite *TestSuiteRequestCtx) TestTemplateContext_HostField() {
	assert := assert.New(suite.T())

	context := suite.requestCtx.TemplateContext()

	assert.Equal("example.com", context["Host"])
}

func (suite *TestSuiteRequestCtx) TestTemplateContext_MethodField() {
	assert := assert.New(suite.T())

	context := suite.requestCtx.TemplateContext()

	assert.Equal("POST", context["Method"])
}

func (suite *TestSuiteRequestCtx) TestTemplateContext_UserAgentField() {
	assert := assert.New(suite.T())

	context := suite.requestCtx.TemplateContext()

	assert.Equal("test-agent/1.0", context["UserAgent"])
}

func (suite *TestSuiteRequestCtx) TestTemplateContext_PayloadField() {
	assert := assert.New(suite.T())

	context := suite.requestCtx.TemplateContext()

	expectedPayload := `{"test": "payload", "data": 123}`
	assert.Equal(expectedPayload, context["Payload"])
}

func (suite *TestSuiteRequestCtx) TestTemplateContext_IsTLSField() {
	assert := assert.New(suite.T())

	// Test HTTP (not TLS)
	suite.fastCtx.Request.SetRequestURI("http://example.com/webhook")
	context := suite.requestCtx.TemplateContext()
	assert.False(context["IsTLS"].(bool))

	// Test HTTPS (TLS)
	suite.fastCtx.Request.SetRequestURI("https://example.com/webhook")
	context = suite.requestCtx.TemplateContext()
	// Note: IsTLS() may return false in unit tests without proper TLS setup
	// but we're testing that the field is accessible and returns a boolean
	assert.IsType(false, context["IsTLS"])
}

func (suite *TestSuiteRequestCtx) TestTemplateContext_ConnIDField() {
	assert := assert.New(suite.T())

	context := suite.requestCtx.TemplateContext()

	// ConnID should be a uint64
	assert.IsType(uint64(0), context["ConnID"])
}

func (suite *TestSuiteRequestCtx) TestTemplateContext_ConnTimeField() {
	assert := assert.New(suite.T())

	context := suite.requestCtx.TemplateContext()

	// ConnTime should be a time.Time
	assert.IsType(time.Time{}, context["ConnTime"])
}

func (suite *TestSuiteRequestCtx) TestTemplateContext_RequestTimeField() {
	assert := assert.New(suite.T())

	context := suite.requestCtx.TemplateContext()

	// RequestTime should be a time.Time
	assert.IsType(time.Time{}, context["RequestTime"])
}

func (suite *TestSuiteRequestCtx) TestTemplateContext_RemoteAddrField() {
	assert := assert.New(suite.T())

	context := suite.requestCtx.TemplateContext()

	// RemoteAddr should be accessible
	assert.NotNil(context["RemoteAddr"])
}

func (suite *TestSuiteRequestCtx) TestTemplateContext_RemoteIPField() {
	assert := assert.New(suite.T())

	context := suite.requestCtx.TemplateContext()

	// RemoteIP should be accessible
	assert.NotNil(context["RemoteIP"])
}

func (suite *TestSuiteRequestCtx) TestTemplateContext_QueryArgsField() {
	assert := assert.New(suite.T())

	context := suite.requestCtx.TemplateContext()

	// QueryArgs should be accessible
	assert.NotNil(context["QueryArgs"])
}

func (suite *TestSuiteRequestCtx) TestTemplateContext_URIField() {
	assert := assert.New(suite.T())

	context := suite.requestCtx.TemplateContext()

	// URI should be accessible and not nil
	assert.NotNil(context["URI"])
}

func (suite *TestSuiteRequestCtx) TestTemplateContext_RequestField() {
	assert := assert.New(suite.T())

	context := suite.requestCtx.TemplateContext()

	// Request should be accessible and be a pointer to fasthttp.Request
	assert.NotNil(context["Request"])
	assert.IsType(&fasthttp.Request{}, context["Request"])
}

func (suite *TestSuiteRequestCtx) TestTemplateContext_DifferentMethods() {
	assert := assert.New(suite.T())

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

	for _, method := range methods {
		suite.fastCtx.Request.Header.SetMethod(method)
		context := suite.requestCtx.TemplateContext()

		assert.Equal(method, context["Method"], "Method should be correctly set for %s", method)
	}
}

func (suite *TestSuiteRequestCtx) TestTemplateContext_DifferentHosts() {
	assert := assert.New(suite.T())

	hosts := []string{"localhost", "example.com", "api.example.org", "webhook.test"}

	for _, host := range hosts {
		// Create fresh context for each host test
		fastCtx := &fasthttp.RequestCtx{}
		requestCtx := &RequestCtx{RequestCtx: fastCtx}
		
		fastCtx.Request.Header.SetHost(host)
		context := requestCtx.TemplateContext()

		assert.Equal(host, context["Host"], "Host should be correctly set for %s", host)
	}
}

func (suite *TestSuiteRequestCtx) TestTemplateContext_DifferentUserAgents() {
	assert := assert.New(suite.T())

	userAgents := []string{
		"Mozilla/5.0 (compatible; bot/1.0)",
		"curl/7.68.0",
		"PostmanRuntime/7.28.4",
		"webhook-client/2.1",
	}

	for _, ua := range userAgents {
		suite.fastCtx.Request.Header.SetUserAgent(ua)
		context := suite.requestCtx.TemplateContext()

		assert.Equal(ua, context["UserAgent"], "UserAgent should be correctly set for %s", ua)
	}
}

func (suite *TestSuiteRequestCtx) TestTemplateContext_DifferentPayloads() {
	assert := assert.New(suite.T())

	payloads := []string{
		`{"simple": "json"}`,
		`<xml><test>data</test></xml>`,
		`form=data&encoded=true`,
		`plain text payload`,
		``,
	}

	for _, payload := range payloads {
		suite.fastCtx.Request.SetBody([]byte(payload))
		context := suite.requestCtx.TemplateContext()

		assert.Equal(payload, context["Payload"], "Payload should be correctly set")
	}
}

func (suite *TestSuiteRequestCtx) TestTemplateContext_EmptyPayload() {
	assert := assert.New(suite.T())

	suite.fastCtx.Request.SetBody(nil)
	context := suite.requestCtx.TemplateContext()

	assert.Equal("", context["Payload"])
}

func (suite *TestSuiteRequestCtx) TestTemplateContext_LargePayload() {
	assert := assert.New(suite.T())

	// Create a large payload (10KB)
	largePayload := make([]byte, 10240)
	for i := range largePayload {
		largePayload[i] = byte('A' + (i % 26))
	}

	suite.fastCtx.Request.SetBody(largePayload)
	context := suite.requestCtx.TemplateContext()

	assert.Equal(string(largePayload), context["Payload"])
	assert.Len(context["Payload"].(string), 10240)
}

func (suite *TestSuiteRequestCtx) TestTemplateContext_ConsistentData() {
	assert := assert.New(suite.T())

	// Get context twice and ensure data is consistent
	context1 := suite.requestCtx.TemplateContext()
	context2 := suite.requestCtx.TemplateContext()

	// All fields should be identical
	for key, value1 := range context1 {
		value2, exists := context2[key]
		assert.True(exists, "Key %s should exist in both contexts", key)
		assert.Equal(value1, value2, "Value for key %s should be consistent", key)
	}
}

func (suite *TestSuiteRequestCtx) TestTemplateContext_NilFastHttpRequestCtx() {
	assert := assert.New(suite.T())

	// Test with nil embedded RequestCtx - this should panic
	nilRequestCtx := &RequestCtx{RequestCtx: nil}

	assert.Panics(func() {
		nilRequestCtx.TemplateContext()
	})
}

func TestRunRequestCtxSuite(t *testing.T) {
	suite.Run(t, new(TestSuiteRequestCtx))
}

// Benchmarks

func BenchmarkTemplateContext(b *testing.B) {
	fastCtx := &fasthttp.RequestCtx{}
	requestCtx := &RequestCtx{RequestCtx: fastCtx}

	// Set up request data
	fastCtx.Request.SetRequestURI("https://example.com/webhook?param=value")
	fastCtx.Request.Header.SetMethod("POST")
	fastCtx.Request.Header.SetHost("example.com")
	fastCtx.Request.Header.SetUserAgent("benchmark-agent/1.0")
	fastCtx.Request.SetBody([]byte(`{"benchmark": "data"}`))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		requestCtx.TemplateContext()
	}
}

func BenchmarkTemplateContext_LargePayload(b *testing.B) {
	fastCtx := &fasthttp.RequestCtx{}
	requestCtx := &RequestCtx{RequestCtx: fastCtx}

	// Create large payload (1MB)
	largePayload := make([]byte, 1024*1024)
	for i := range largePayload {
		largePayload[i] = byte('A' + (i % 26))
	}

	fastCtx.Request.SetRequestURI("https://example.com/webhook")
	fastCtx.Request.Header.SetMethod("POST")
	fastCtx.Request.SetBody(largePayload)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		requestCtx.TemplateContext()
	}
}

func BenchmarkTemplateContext_ManyFields(b *testing.B) {
	fastCtx := &fasthttp.RequestCtx{}
	requestCtx := &RequestCtx{RequestCtx: fastCtx}

	// Set up many headers and query parameters
	fastCtx.Request.SetRequestURI("https://example.com/webhook?a=1&b=2&c=3&d=4&e=5&f=6&g=7&h=8&i=9&j=10")
	fastCtx.Request.Header.SetMethod("POST")
	fastCtx.Request.Header.SetHost("example.com")
	fastCtx.Request.Header.Set("X-Custom-Header-1", "value1")
	fastCtx.Request.Header.Set("X-Custom-Header-2", "value2")
	fastCtx.Request.Header.Set("X-Custom-Header-3", "value3")
	fastCtx.Request.SetBody([]byte(`{"complex": {"nested": {"data": "structure"}}}`))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		requestCtx.TemplateContext()
	}
}
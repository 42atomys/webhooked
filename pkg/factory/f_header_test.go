package factory

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"42stellar.org/webhooks/internal/valuable"
)

// func Test_f_header(t *testing.T) {
// 	assert := assert.New(t)

// 	headerName := "X-Token"
// 	header := make(http.Header)
// 	header.Add(headerName, "test")

// 	request := httptest.NewRequest("POST", "/", nil)

// 	tests := []struct {
// 		name    string
// 		header  http.Header
// 		want    string
// 		wantErr bool
// 	}{
// 		{"no config will fail", nil, "", true},
// 		{"no values", nil, "", false},
// 		{"no inputs", nil, "", true},
// 		{"valid config but no header set", nil, "", false},
// 		{"valid config and header set", header, "test", false},
// 		{"invalid config name and no header set", nil, "", false},
// 		{"invalid config name but header set", header, "", false},
// 		{"invalid config map", nil, "", true},
// 		{"invalid input type", nil, "", true},
// 	}
// 	for _, test := range tests {
// 		factory := newFactory(&headerFactory{})
// 		request.Header = header
// 		factory.WithInput("request", request)
// 		factory.WithInput("headerName", &InputConfig{Valuable: valuable.Valuable{Value: &headerName}})

// 		err := factory.Run()
// 		if test.wantErr {
// 			assert.Error(err, test.name)
// 		} else {
// 			assert.NoError(err, test.name)
// 		}

// 		assert.Equal(test.want, factory.Outputs[0].Value, test.name)
// 	}
// }

type testSuiteFactoryHeader struct {
	suite.Suite
	request  *http.Request
	iFactory *headerFactory
}

func (suite *testSuiteFactoryHeader) BeforeTest(suiteName, testName string) {
	headerName := "X-Token"
	header := make(http.Header)
	header.Add(headerName, "test")

	suite.request = httptest.NewRequest("POST", "/", nil)
	suite.request.Header = header

	suite.iFactory = &headerFactory{}
}

func TestFactoryHeader(t *testing.T) {
	suite.Run(t, new(testSuiteFactoryHeader))
}

func (suite *testSuiteFactoryHeader) TestRunFactoryWithoutInputs() {
	var factory = newFactory(&headerFactory{})
	factory.Inputs = make([]*Var, 0)
	suite.Errorf(factory.Run(), "missing input headerName")

	factory.Inputs = suite.iFactory.DefinedInpus()[1:]
	suite.Errorf(factory.Run(), "missing input request")

	factory.Inputs = suite.iFactory.DefinedInpus()
	suite.Errorf(factory.Run(), "missing input request")
	suite.Equal("", factory.Outputs[0].Value)
}

func (suite *testSuiteFactoryHeader) TestRunFactory() {
	headerName := "X-Token"
	header := make(http.Header)
	header.Add(headerName, "test")
	factory := newFactory(&headerFactory{})

	factory.WithInput("request", suite.request)
	factory.WithInput("headerName", &InputConfig{Valuable: valuable.Valuable{Value: &headerName}})

	suite.NoError(factory.Run())
	suite.Equal("test", factory.Outputs[0].Value)
}

package factory

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"atomys.codes/webhooked/internal/valuable"
)

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

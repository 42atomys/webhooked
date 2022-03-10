package factory

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"atomys.codes/webhooked/internal/valuable"
)

type testSuiteFactoryDebug struct {
	suite.Suite
	iFactory    *debugFactory
	inputHelper func(name, data string) *InputConfig
}

func (suite *testSuiteFactoryDebug) BeforeTest(suiteName, testName string) {
	suite.inputHelper = func(name, data string) *InputConfig {
		return &InputConfig{
			Name:     name,
			Valuable: valuable.Valuable{Value: &data},
		}
	}
	suite.iFactory = &debugFactory{}
}

func TestFactoryDebug(t *testing.T) {
	suite.Run(t, new(testSuiteFactoryDebug))
}

func (suite *testSuiteFactoryDebug) TestRunFactoryWithoutInputs() {
	var factory = newFactory(&debugFactory{})
	factory.Inputs = make([]*Var, 0)
	suite.Errorf(factory.Run(), "missing input first")
}

func (suite *testSuiteFactoryDebug) TestRunFactory() {
	factory := newFactory(&debugFactory{})

	factory.WithInput("", suite.inputHelper("first", "yes"))
	suite.NoError(factory.Run())

}

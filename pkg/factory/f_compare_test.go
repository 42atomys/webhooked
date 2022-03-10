package factory

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"atomys.codes/webhooked/internal/valuable"
)

type testSuiteFactoryCompare struct {
	suite.Suite
	iFactory    *compareFactory
	inputHelper func(name, data string) *InputConfig
}

func (suite *testSuiteFactoryCompare) BeforeTest(suiteName, testName string) {
	suite.inputHelper = func(name, data string) *InputConfig {
		return &InputConfig{
			Name:     name,
			Valuable: valuable.Valuable{Value: &data},
		}
	}
	suite.iFactory = &compareFactory{}
}

func TestFactoryCompare(t *testing.T) {
	suite.Run(t, new(testSuiteFactoryCompare))
}

func (suite *testSuiteFactoryCompare) TestRunFactoryWithoutInputs() {
	var factory = newFactory(&compareFactory{})
	factory.Inputs = make([]*Var, 0)
	suite.Errorf(factory.Run(), "missing input first")

	factory.Inputs = suite.iFactory.DefinedInpus()[:1]
	suite.Errorf(factory.Run(), "missing input second")
}

func (suite *testSuiteFactoryCompare) TestRunFactory() {
	factory := newFactory(&compareFactory{})

	factory.WithInput("first", suite.inputHelper("first", "test")).WithInput("second", suite.inputHelper("second", "test"))
	suite.NoError(factory.Run())
	suite.Equal(true, factory.Outputs[0].Value)

	factory.WithInput("first", suite.inputHelper("first", "yes")).WithInput("second", suite.inputHelper("second", "no"))
	suite.NoError(factory.Run())
	suite.Equal(false, factory.Outputs[0].Value)

	factory.
		WithInput("first", suite.inputHelper("first", "yes")).
		WithInput("second", suite.inputHelper("second", "no")).
		WithConfig(map[string]interface{}{"inverse": true})
	suite.NoError(factory.Run())
	suite.Equal(true, factory.Outputs[0].Value)

}

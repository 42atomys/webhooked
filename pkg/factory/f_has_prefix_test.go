package factory

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"atomys.codes/webhooked/internal/valuable"
)

type testSuiteFactoryHasPrefix struct {
	suite.Suite
	iFactory    *hasPrefixFactory
	inputHelper func(name, data string) *InputConfig
}

func (suite *testSuiteFactoryHasPrefix) BeforeTest(suiteName, testName string) {
	suite.inputHelper = func(name, data string) *InputConfig {
		return &InputConfig{
			Name:     name,
			Valuable: valuable.Valuable{Value: &data},
		}
	}
	suite.iFactory = &hasPrefixFactory{}
}

func TestFactoryHasPrefix(t *testing.T) {
	suite.Run(t, new(testSuiteFactoryHasPrefix))
}

func (suite *testSuiteFactoryHasPrefix) TestRunFactoryWithoutInputs() {
	var factory = newFactory(&hasPrefixFactory{})
	factory.Inputs = make([]*Var, 0)
	suite.Errorf(factory.Run(), "missing input text")

	factory.Inputs = suite.iFactory.DefinedInpus()[:1]
	suite.Errorf(factory.Run(), "missing input prefix")
}

func (suite *testSuiteFactoryHasPrefix) TestRunFactory() {
	factory := newFactory(&hasPrefixFactory{})

	factory.WithInput("text", suite.inputHelper("text", "yes")).WithInput("prefix", suite.inputHelper("prefix", "y"))
	suite.NoError(factory.Run())
	suite.Equal(true, factory.Outputs[0].Value)

	factory.WithInput("text", suite.inputHelper("text", "yes")).WithInput("prefix", suite.inputHelper("prefix", "no"))
	suite.NoError(factory.Run())
	suite.Equal(false, factory.Outputs[0].Value)

	factory.
		WithInput("text", suite.inputHelper("text", "yes")).
		WithInput("prefix", suite.inputHelper("prefix", "no")).
		WithConfig(map[string]interface{}{"inverse": true})
	suite.NoError(factory.Run())
	suite.Equal(true, factory.Outputs[0].Value)

}

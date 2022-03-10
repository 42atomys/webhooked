package factory

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"atomys.codes/webhooked/internal/valuable"
)

type testSuiteFactoryHasSuffix struct {
	suite.Suite
	iFactory    *hasSuffixFactory
	inputHelper func(name, data string) *InputConfig
}

func (suite *testSuiteFactoryHasSuffix) BeforeTest(suiteName, testName string) {
	suite.inputHelper = func(name, data string) *InputConfig {
		return &InputConfig{
			Name:     name,
			Valuable: valuable.Valuable{Value: &data},
		}
	}
	suite.iFactory = &hasSuffixFactory{}
}

func TestFactoryHasSuffix(t *testing.T) {
	suite.Run(t, new(testSuiteFactoryHasSuffix))
}

func (suite *testSuiteFactoryHasSuffix) TestRunFactoryWithoutInputs() {
	var factory = newFactory(&hasSuffixFactory{})
	factory.Inputs = make([]*Var, 0)
	suite.Errorf(factory.Run(), "missing input text")

	factory.Inputs = suite.iFactory.DefinedInpus()[:1]
	suite.Errorf(factory.Run(), "missing input suffix")
}

func (suite *testSuiteFactoryHasSuffix) TestRunFactory() {
	factory := newFactory(&hasSuffixFactory{})

	factory.WithInput("text", suite.inputHelper("text", "yes")).WithInput("suffix", suite.inputHelper("suffix", "s"))
	suite.NoError(factory.Run())
	suite.Equal(true, factory.Outputs[0].Value)

	factory.WithInput("text", suite.inputHelper("text", "yes")).WithInput("suffix", suite.inputHelper("suffix", "no"))
	suite.NoError(factory.Run())
	suite.Equal(false, factory.Outputs[0].Value)

	factory.
		WithInput("text", suite.inputHelper("text", "yes")).
		WithInput("suffix", suite.inputHelper("suffix", "no")).
		WithConfig(map[string]interface{}{"inverse": true})
	suite.NoError(factory.Run())
	suite.Equal(true, factory.Outputs[0].Value)

}

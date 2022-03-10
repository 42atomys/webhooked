package factory

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"atomys.codes/webhooked/internal/valuable"
)

type testSuiteFactoryGenerateHMAC256 struct {
	suite.Suite
	iFactory    *generateHMAC256Factory
	inputHelper func(name, data string) *InputConfig
}

func (suite *testSuiteFactoryGenerateHMAC256) BeforeTest(suiteName, testName string) {
	suite.inputHelper = func(name, data string) *InputConfig {
		return &InputConfig{
			Name:     name,
			Valuable: valuable.Valuable{Value: &data},
		}
	}
	suite.iFactory = &generateHMAC256Factory{}
}

func TestFactoryGenerateHMAC256(t *testing.T) {
	suite.Run(t, new(testSuiteFactoryGenerateHMAC256))
}

func (suite *testSuiteFactoryGenerateHMAC256) TestRunFactoryWithoutInputs() {
	var factory = newFactory(&generateHMAC256Factory{})
	factory.Inputs = make([]*Var, 0)
	suite.Errorf(factory.Run(), "missing input secret")

	factory.Inputs = suite.iFactory.DefinedInpus()[:1]
	suite.Errorf(factory.Run(), "missing input payload")
}

func (suite *testSuiteFactoryGenerateHMAC256) TestRunFactory() {
	factory := newFactory(&generateHMAC256Factory{})

	factory.WithInput("payload", suite.inputHelper("payload", "test")).WithInput("secret", suite.inputHelper("secret", "test"))
	suite.NoError(factory.Run())
	suite.Equal("88cd2108b5347d973cf39cdf9053d7dd42704876d8c9a9bd8e2d168259d3ddf7", factory.Outputs[0].Value)
}

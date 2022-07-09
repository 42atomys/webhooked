package factory

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"atomys.codes/webhooked/internal/valuable"
)

type fakeFactory struct{}

func (*fakeFactory) Name() string         { return "fake" }
func (*fakeFactory) DefinedInpus() []*Var { return []*Var{{false, reflect.TypeOf(""), "name", ""}} }
func (*fakeFactory) DefinedOutputs() []*Var {
	return []*Var{{false, reflect.TypeOf(""), "message", ""}}
}
func (*fakeFactory) Func() RunFunc {
	return func(factory *Factory, configRaw map[string]interface{}) error {
		n, ok := factory.Input("name")
		if !ok {
			return errors.New("name is not defined")
		}
		factory.Output("message", fmt.Sprintf("hello %s", n.Value))
		return nil
	}
}

type testSuiteFactory struct {
	suite.Suite
}

func (suite *testSuiteFactory) BeforeTest(suiteName, testName string) {
}

func TestFactory(t *testing.T) {
	suite.Run(t, new(testSuiteFactory))
}

func (suite *testSuiteFactory) TestFactoryName() {
	var factory = newFactory(&fakeFactory{})
	suite.Equal("fake", factory.Name)
}

func (suite *testSuiteFactory) TestFactoryInputs() {
	var factory = newFactory(&fakeFactory{})
	suite.Len(factory.Inputs, 1)

	var i, ok = factory.Input("name")
	suite.True(ok)
	suite.Equal(false, i.Internal)
	suite.Equal("name", i.Name)
	suite.Equal(reflect.TypeOf(""), i.Type)
	suite.Equal("", i.Value)
}

func (suite *testSuiteFactory) TestFactoryOutputs() {
	var factory = newFactory(&fakeFactory{})
	suite.Len(factory.Outputs, 1)

	var i, ok = GetVar(factory.Outputs, "message")
	suite.True(ok)
	suite.Equal(false, i.Internal)
	suite.Equal("message", i.Name)
	suite.Equal(reflect.TypeOf(""), i.Type)
	suite.Equal("", i.Value)
}

func (suite *testSuiteFactory) TestAddInput() {
	var factory = newFactory(&fakeFactory{})

	factory.WithInput("name", 1)
	suite.Len(factory.Inputs, 1)

	slice, err := factory.with(factory.Inputs, "name", 1)
	suite.Error(err)
	suite.Len(slice, 1)

	slice, err = factory.with(factory.Inputs, "invalid", nil)
	suite.Error(err)
	suite.Len(slice, 1)

	slice, err = factory.with(factory.Inputs, "name", "test")
	suite.NoError(err)
	suite.Len(slice, 1)
}

func (suite *testSuitePipeline) TestAddPipelineInput() {
	var factory = newFactory(&fakeFactory{})
	factory.withPipelineInput("name", "pipeline")
	suite.Equal("pipeline", factory.Inputs[0].Value)

	factory.withPipelineInput("name", 1)
	suite.Equal("pipeline", factory.Inputs[0].Value)
}

func (suite *testSuiteFactory) TestWithID() {
	var factory = newFactory(&fakeFactory{})
	factory.WithID("id")
	suite.Equal("id", factory.ID)
	suite.Equal("id", factory.Identifier())

	factory.WithID("")
	suite.Equal("", factory.ID)
	suite.Equal(factory.Name, factory.Identifier())
}

func (suite *testSuiteFactory) TestWithConfig() {
	var factory = newFactory(&fakeFactory{})
	factory.WithConfig(map[string]interface{}{"name": "test"})
	suite.Equal("test", factory.Config["name"])

	factory = newFactory(&fakeFactory{})
	factory.WithConfig(map[string]interface{}{"id": "configID"})
	suite.Equal("configID", factory.ID)
	suite.Equal("configID", factory.Identifier())
	suite.Len(factory.Config, 0)
}

func (suite *testSuiteFactory) TestRun() {
	var factory = newFactory(&fakeFactory{})
	factory.WithInput("name", "test")
	suite.NoError(factory.Run())
	suite.Equal("hello test", factory.Outputs[0].Value)

	factory = newFactory(&fakeFactory{})
	factory.Inputs = []*Var{}
	suite.Error(factory.Run())
	suite.Equal("", factory.Outputs[0].Value)
}

func (suite *testSuiteFactory) TestProcessInputConfig() {
	var v = &Var{Name: "name", Value: &InputConfig{Valuable: valuable.Valuable{Values: []string{"{{ .Outputs.id.message }}", "static"}}}}

	var factory = newFactory(&fakeFactory{})
	ctx := context.WithValue(context.Background(), ctxPipeline, Pipeline{Outputs: map[string]map[string]interface{}{
		"id": {
			"message": "testValue",
		},
	}})
	factory.ctx = ctx

	v, ok := factory.processInputConfig(v)
	suite.True(ok)
	suite.ElementsMatch(v.Value.(*InputConfig).Get(), []string{"testValue", "static"})

	factory = newFactory(&fakeFactory{})
	factory.ctx = ctx

	factory.Inputs[0] = v
	v, ok = factory.Input("name")
	suite.True(ok)
	suite.ElementsMatch(v.Value.(*InputConfig).Get(), []string{"testValue", "static"})
}

func (suite *testSuiteFactory) TestGoTempalteValue() {
	ret := goTemplateValue("{{ .test }}", map[string]interface{}{"test": "testValue"})
	suite.Equal("testValue", ret)
}

func (suite *testSuiteFactory) TestFactoryDeepCopy() {
	var factory = newFactory(&fakeFactory{})
	factory.WithConfig(map[string]interface{}{"name": "test"})

	suite.NotSame(factory, factory.DeepCopy())
}

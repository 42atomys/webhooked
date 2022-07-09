package factory

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type testSuitePipeline struct {
	suite.Suite
	pipeline    *Pipeline
	testFactory *Factory
}

func (suite *testSuitePipeline) BeforeTest(suiteName, testName string) {
	suite.pipeline = NewPipeline()
	suite.testFactory = newFactory(&fakeFactory{})
	suite.pipeline.AddFactory(suite.testFactory)
}

func TestPipeline(t *testing.T) {
	suite.Run(t, new(testSuitePipeline))
}

func (suite *testSuitePipeline) TestPipelineInput() {
	suite.pipeline.Inputs["name"] = "test"
	suite.pipeline.Inputs["invalid"] = "test"

	suite.pipeline.Run()

	i, ok := suite.testFactory.Input("name")
	suite.True(ok)
	suite.Equal("test", i.Value)

	i, ok = suite.testFactory.Input("invalid")
	suite.False(ok)
	suite.Nil(i)
}

func (suite *testSuitePipeline) TestPipelineCreation() {
	var pipeline = NewPipeline()
	pipeline.AddFactory(suite.testFactory)

	suite.Equal(1, pipeline.FactoryCount())
	suite.True(pipeline.HasFactories())
}

func (suite *testSuitePipeline) TestRunEmptyPipeline() {
	var pipeline = NewPipeline()

	suite.Equal(0, pipeline.FactoryCount())
	suite.False(pipeline.HasFactories())

	f := pipeline.Run()
	suite.Nil(f)
}

func (suite *testSuitePipeline) TestPipelineRun() {
	var pipeline = NewPipeline()
	var wantedResult = "hello test"

	pipeline.AddFactory(suite.testFactory)
	pipeline.Inputs["name"] = "test"
	pipeline.WantResult(wantedResult)

	f := pipeline.Run()
	suite.Equal(f, suite.testFactory)

	suite.True(pipeline.CheckResult())
	suite.Equal(wantedResult, pipeline.Outputs["fake"]["message"])
}

func (suite *testSuitePipeline) TestPipelineWithInput() {
	var pipeline = NewPipeline()
	pipeline.WithInput("test", true)

	suite.True(pipeline.Inputs["test"].(bool))
}

func (suite *testSuitePipeline) TestPipelineResultWithInvalidType() {
	var pipeline = NewPipeline()

	pipeline.AddFactory(suite.testFactory)
	pipeline.Inputs["name"] = "test"
	pipeline.WantResult(true)
	pipeline.Run()

	suite.False(pipeline.CheckResult())
}

func (suite *testSuitePipeline) TestCheckResultWithoutWantedResult() {
	var pipeline = NewPipeline()

	pipeline.AddFactory(suite.testFactory)
	pipeline.Inputs["name"] = "test"
	pipeline.Run()

	suite.False(pipeline.CheckResult())
}

func (suite *testSuitePipeline) TestPipelineFailedDueToFactoryErr() {
	var pipeline = NewPipeline()
	var factory = newFactory(&fakeFactory{})
	var factory2 = newFactory(&fakeFactory{})
	factory.Inputs = make([]*Var, 0)

	pipeline.AddFactory(factory).AddFactory(factory2)
	ret := pipeline.Run()
	suite.Equal(factory, ret)
}

func (suite *testSuitePipeline) TestPipelineDeepCopy() {
	var pipeline = NewPipeline()
	var factory = newFactory(&fakeFactory{})
	var factory2 = newFactory(&fakeFactory{})
	factory.Inputs = make([]*Var, 0)

	pipeline.AddFactory(factory).AddFactory(factory2)
	pipeline.Inputs["name"] = "test"
	pipeline.WantResult("test")

	var pipeline2 = pipeline.DeepCopy()
	suite.NotSame(pipeline, pipeline2)
}

package factory

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type testSuiteRegistry struct {
	suite.Suite
}

func (suite *testSuiteRegistry) BeforeTest(suiteName, testName string) {
}

func TestRegistry(t *testing.T) {
	suite.Run(t, new(testSuiteRegistry))
}

func (suite *testSuiteRegistry) TestRegisterANewFactory() {
	var actualFactoryLenSize = len(factoryMap)
	err := Register(&fakeFactory{})

	suite.NoError(err)
	suite.Equal(actualFactoryLenSize+1, len(factoryMap))

	var factory, ok = GetFactoryByName("fake")
	suite.True(ok)
	suite.Equal("fake", factory.Name)

}

func (suite *testSuiteRegistry) TestRegisterFactoryTwice() {
	err := Register(&fakeFactory{})
	suite.Error(err)
}

func (suite *testSuiteRegistry) TestGetFactoryByHerName() {
	factory, ok := GetFactoryByName("invalid")
	suite.False(ok)
	suite.Nil(factory)
}

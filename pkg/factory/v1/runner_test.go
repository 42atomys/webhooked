package factory

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testGet42(configRaw map[string]interface{}, lastOuput string, inputs ...interface{}) (string, error) {
	return "42", nil
}

func testCompareWith42(configRaw map[string]interface{}, lastOuput string, inputs ...interface{}) (string, error) {
	if lastOuput == "42" {
		return "t", nil
	}

	return "f", nil
}

func testRunnerFn(factory *Factory, lastOutput string, defaultFn RunnerFunc) (string, error) {
	return factory.Fn(factory.Config, lastOutput)
}

func testRunnerFnWithFailure(factory *Factory, lastOutput string, defaultFn RunnerFunc) (string, error) {
	return "", errors.New("test error")
}

func TestRun(t *testing.T) {
	assert := assert.New(t)

	type args struct {
		factories []*Factory
		runnerFn  RunnerExternalFunc
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"no factories", args{[]*Factory{}, testRunnerFn}, true, false},
		{"one factory without lastOutput", args{[]*Factory{
			{Name: "get42", Fn: testGet42, Config: make(map[string]interface{})}}, testRunnerFn}, false, true,
		},
		{"one factory with only compare", args{[]*Factory{
			{Name: "compareWith42", Fn: testCompareWith42, Config: make(map[string]interface{})}}, testRunnerFn}, false, false,
		},
		{"get and compare 42", args{[]*Factory{
			{Name: "get42", Fn: testGet42, Config: make(map[string]interface{})},
			{Name: "compareWith42", Fn: testCompareWith42, Config: make(map[string]interface{})},
		}, testRunnerFn}, true, false},

		{"no custom runner without comparation will error", args{[]*Factory{
			{Name: "get42", Fn: testGet42, Config: make(map[string]interface{})},
		}, nil}, false, true},

		{"runner function will return an error", args{[]*Factory{
			{Name: "get42", Fn: testGet42, Config: make(map[string]interface{})},
		}, testRunnerFnWithFailure}, false, true},

		{"internal runner will error with not implemented function", args{[]*Factory{
			{Name: "en", Config: make(map[string]interface{})},
		}, nil}, false, true},
	}

	for _, test := range tests {
		got, err := Run(test.args.factories, test.args.runnerFn)
		if test.wantErr {
			assert.Error(err)
		} else {
			assert.NoError(err)
		}

		assert.Equal(test.want, got)
	}
}

func Test_internalRunner(t *testing.T) {
	assert := assert.New(t)

	type args struct {
		factory    *Factory
		lastOutput string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"not implemented method", args{
			&Factory{Name: "invalid"}, ""},
			"", true,
		},
		{"getHeader", args{
			&Factory{Name: "getHeader", Fn: getHeader, Config: make(map[string]interface{})},
			"",
		}, "", true},
		{"compareWithStaticValue", args{
			&Factory{Name: "compareWithStaticValue", Fn: compareWithStaticValue, Config: make(map[string]interface{})},
			"nope",
		}, "f", false},
	}
	for _, test := range tests {
		got, err := internalRunner(test.args.factory, test.args.lastOutput)
		assert.Equal(test.want, got, "%s: want %v, got %v", test.name, test.want, got)
		if test.wantErr {
			assert.Error(err)
		} else {
			assert.NoError(err)
		}
	}
}

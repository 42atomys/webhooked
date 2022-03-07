package factory

import (
	"fmt"
	"net/http"
	"reflect"
)

type headerFactory struct{ Factory }

func (*headerFactory) Name() string {
	return "header"
}

func (*headerFactory) DefinedInpus() []*Var {
	return []*Var{
		{true, reflect.TypeOf(&http.Request{}), "request", nil},
		{false, reflect.TypeOf(&InputConfig{}), "headerName", &InputConfig{}},
	}
}

func (*headerFactory) DefinedOutputs() []*Var {
	return []*Var{
		{false, reflect.TypeOf(""), "value", ""},
	}
}

func (*headerFactory) Func() RunFunc {
	return func(factory *Factory, configRaw map[string]interface{}) error {
		nameVar, ok := factory.Input("headerName")
		if !ok {
			return fmt.Errorf("missing input headerName")
		}

		requestVar, ok := factory.Input("request")
		if !ok || requestVar.Value == nil {
			return fmt.Errorf("missing input request")
		}

		factory.Output("value",
			requestVar.Value.(*http.Request).Header.Get(
				nameVar.Value.(*InputConfig).First(),
			))

		return nil
	}
}

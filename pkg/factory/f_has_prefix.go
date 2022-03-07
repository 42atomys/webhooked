package factory

import (
	"fmt"
	"reflect"
	"strings"
)

type hasPrefixFactory struct{ Factory }

func (*hasPrefixFactory) Name() string {
	return "hasPrefix"
}

func (*hasPrefixFactory) DefinedInpus() []*Var {
	return []*Var{
		{false, reflect.TypeOf(&InputConfig{}), "text", &InputConfig{}},
		{false, reflect.TypeOf(&InputConfig{}), "prefix", &InputConfig{}},
	}
}

func (*hasPrefixFactory) DefinedOutputs() []*Var {
	return []*Var{
		{false, reflect.TypeOf(false), "result", false},
	}
}

func (c *hasPrefixFactory) Func() RunFunc {
	return func(factory *Factory, configRaw map[string]interface{}) error {
		textVar, ok := factory.Input("text")
		if !ok {
			return fmt.Errorf("missing input text")
		}

		prefixVar, ok := factory.Input("prefix")
		if !ok {
			return fmt.Errorf("missing input prefix")
		}

		var result bool
		for _, text := range textVar.Value.(*InputConfig).Get() {
			for _, prefix := range prefixVar.Value.(*InputConfig).Get() {
				if strings.HasPrefix(text, prefix) {
					result = true
					break
				}
			}
		}

		inverse, _ := configRaw["inverse"].(bool)
		if inverse {
			result = !result
		}

		factory.Output("result", result)
		return nil
	}
}

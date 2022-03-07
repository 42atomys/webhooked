package factory

import (
	"fmt"
	"reflect"
	"strings"
)

type hasSuffixFactory struct{ Factory }

func (*hasSuffixFactory) Name() string {
	return "hasSuffix"
}

func (*hasSuffixFactory) DefinedInpus() []*Var {
	return []*Var{
		{false, reflect.TypeOf(&InputConfig{}), "text", &InputConfig{}},
		{false, reflect.TypeOf(&InputConfig{}), "suffix", &InputConfig{}},
	}
}

func (*hasSuffixFactory) DefinedOutputs() []*Var {
	return []*Var{
		{false, reflect.TypeOf(false), "result", false},
	}
}

func (c *hasSuffixFactory) Func() RunFunc {
	return func(factory *Factory, configRaw map[string]interface{}) error {
		textVar, ok := factory.Input("text")
		if !ok {
			return fmt.Errorf("missing input text")
		}

		suffixVar, ok := factory.Input("suffix")
		if !ok {
			return fmt.Errorf("missing input suffix")
		}

		var result bool
		for _, text := range textVar.Value.(*InputConfig).Get() {
			for _, suffix := range suffixVar.Value.(*InputConfig).Get() {
				if strings.HasSuffix(text, suffix) {
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

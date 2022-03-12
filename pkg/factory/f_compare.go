package factory

import (
	"fmt"
	"reflect"

	"github.com/rs/zerolog/log"
)

type compareFactory struct{ Factory }

func (*compareFactory) Name() string {
	return "compare"
}

func (*compareFactory) DefinedInpus() []*Var {
	return []*Var{
		{false, reflect.TypeOf(&InputConfig{}), "first", &InputConfig{}},
		{false, reflect.TypeOf(&InputConfig{}), "second", &InputConfig{}},
	}
}

func (*compareFactory) DefinedOutputs() []*Var {
	return []*Var{
		{false, reflect.TypeOf(false), "result", false},
	}
}

func (c *compareFactory) Func() RunFunc {
	return func(factory *Factory, configRaw map[string]interface{}) error {
		firstVar, ok := factory.Input("first")
		if !ok {
			return fmt.Errorf("missing input first")
		}

		secondVar, ok := factory.Input("second")
		if !ok {
			return fmt.Errorf("missing input second")
		}

		result := c.sliceMatches(
			firstVar.Value.(*InputConfig).Get(),
			secondVar.Value.(*InputConfig).Get(),
		)

		inverse, _ := configRaw["inverse"].(bool)
		if inverse {
			result = !result
		}

		log.Debug().Bool("inversed", inverse).Msgf("factory compared slice %+v and %+v = %+v",
			firstVar.Value.(*InputConfig).Get(),
			secondVar.Value.(*InputConfig).Get(),
			result,
		)
		factory.Output("result", result)
		return nil
	}
}

// sliceMatches returns true if one element match in all slices
func (*Factory) sliceMatches(slice1, slice2 []string) bool {
	// Loop two times, first to find slice1 strings not in slice2,
	// second loop to find slice2 strings not in slice1
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			for _, s2 := range slice2 {
				if s1 == s2 {
					return true
				}
			}
		}
		// Swap the slices, only if it was the first loop
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}
	return false
}

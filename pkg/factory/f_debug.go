package factory

import (
	"fmt"
	"reflect"

	"github.com/rs/zerolog/log"
)

type debugFactory struct{ Factory }

func (*debugFactory) Name() string {
	return "debug"
}

func (*debugFactory) DefinedInpus() []*Var {
	return []*Var{
		{false, reflect.TypeOf(&InputConfig{}), "", &InputConfig{}},
	}
}

func (*debugFactory) DefinedOutputs() []*Var {
	return []*Var{}
}

func (c *debugFactory) Func() RunFunc {
	return func(factory *Factory, configRaw map[string]interface{}) error {
		debugValue, ok := factory.Input("")
		if !ok {
			return fmt.Errorf("missing input")
		}

		log.Debug().Msgf("debug value: %+v", debugValue.Value)
		return nil
	}
}

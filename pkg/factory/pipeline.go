package factory

import (
	"context"
	"reflect"

	"github.com/rs/zerolog/log"
)

// NewPipeline initializes a new pipeline
func NewPipeline() *Pipeline {
	return &Pipeline{
		Outputs:   make(map[string]map[string]interface{}),
		Variables: make(map[string]interface{}),
		Config:    make(map[string]interface{}),
		Inputs:    make(map[string]interface{}),
	}
}

// AddFactory adds a new factory to the pipeline. New Factory is added to the
// end of the pipeline.
func (p *Pipeline) AddFactory(f *Factory) *Pipeline {
	p.factories = append(p.factories, f)
	return p
}

// HasFactories returns true if the pipeline has at least one factory.
func (p *Pipeline) HasFactories() bool {
	return p.FactoryCount() > 0
}

// FactoryCount returns the number of factories in the pipeline.
func (p *Pipeline) FactoryCount() int {
	return len(p.factories)
}

// WantResult sets the wanted result of the pipeline.
// the result is compared to the last result of the pipeline.
// type and value of the result must be the same as the last result
func (p *Pipeline) WantResult(result interface{}) *Pipeline {
	p.Result = result
	return p
}

// CheckResult checks if the pipeline result is the same as the wanted result.
// type and value of the result must be the same as the last result
func (p *Pipeline) CheckResult() bool {
	for _, lr := range p.LastResults {
		if reflect.TypeOf(lr) != reflect.TypeOf(p.Result) {
			log.Warn().Msgf("pipeline result is not the same type as wanted result")
			return false
		}
		if lr == p.Result {
			return true
		}
	}
	return false
}

// Run executes the pipeline.
// Factories are executed in the order they were added to the pipeline.
// The last factory is returned
//
// @return the last factory
func (p *Pipeline) Run() *Factory {
	for _, f := range p.factories {
		f.ctx = context.WithValue(f.ctx, ctxPipeline, p)
		for k, v := range p.Inputs {
			f.withPipelineInput(k, v)
		}

		log.Debug().Msgf("running factory %s", f.Name)
		for _, v := range f.Inputs {
			log.Debug().Msgf("factory %s input %s = %+v", f.Name, v.Name, v.Value)
		}
		if err := f.Run(); err != nil {
			log.Error().Msgf("factory %s failed: %s", f.Name, err.Error())
			return f
		}

		for _, v := range f.Outputs {
			log.Debug().Msgf("factory %s output %s = %+v", f.Name, v.Name, v.Value)
		}

		var key = f.Identifier()
		if p.Outputs[key] == nil {
			p.Outputs[key] = make(map[string]interface{})
		}

		if p.Result != nil {
			p.LastResults = make([]interface{}, 0)
		}

		for _, v := range f.Outputs {
			p.Outputs[key][v.Name] = v.Value

			if p.Result != nil {
				p.LastResults = append(p.LastResults, v.Value)
			}
		}
	}

	if p.HasFactories() {
		return p.factories[len(p.factories)-1]
	}

	return nil
}

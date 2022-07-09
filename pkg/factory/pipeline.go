package factory

import (
	"context"
	"reflect"

	"github.com/rs/zerolog/log"
)

// NewPipeline initializes a new pipeline
func NewPipeline() *Pipeline {
	return &Pipeline{
		Outputs: make(map[string]map[string]interface{}),
		Inputs:  make(map[string]interface{}),
	}
}

// DeepCopy creates a deep copy of the pipeline.
func (p *Pipeline) DeepCopy() *Pipeline {
	deepCopy := NewPipeline().WantResult(p.WantedResult)
	for _, f := range p.factories {
		deepCopy.AddFactory(f.DeepCopy())
	}
	for k, v := range p.Inputs {
		deepCopy.WithInput(k, v)
	}
	return deepCopy
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
	p.WantedResult = result
	return p
}

// CheckResult checks if the pipeline result is the same as the wanted result.
// type and value of the result must be the same as the last result
func (p *Pipeline) CheckResult() bool {
	for _, lr := range p.LastResults {
		if reflect.TypeOf(lr) != reflect.TypeOf(p.WantedResult) {
			log.Warn().Msgf("pipeline result is not the same type as wanted result")
			return false
		}
		if lr == p.WantedResult {
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

		if p.WantedResult != nil {
			p.LastResults = make([]interface{}, 0)
		}

		for _, v := range f.Outputs {
			p.writeOutputSafely(f.Identifier(), v.Name, v.Value)

			if p.WantedResult != nil {
				p.LastResults = append(p.LastResults, v.Value)
			}
		}
	}

	if p.HasFactories() {
		return p.factories[len(p.factories)-1]
	}

	// Clean up the pipeline
	p.Inputs = make(map[string]interface{})
	p.Outputs = make(map[string]map[string]interface{})

	return nil
}

// WithInput adds a new input to the pipeline. The input is added safely to prevent
// concurrent map writes error.
func (p *Pipeline) WithInput(name string, value interface{}) *Pipeline {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.Inputs[name] = value
	return p
}

// writeOutputSafely writes the output to the pipeline output map. If the key
// already exists, the value is overwritten. This is principally used to
// write on the map withtout create a new map or PANIC due to concurrency map writes.
func (p *Pipeline) writeOutputSafely(factoryIdentifier, factoryOutputName string, value interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Ensure the factory output map exists
	if p.Outputs[factoryIdentifier] == nil {
		p.Outputs[factoryIdentifier] = make(map[string]interface{})
	}

	p.Outputs[factoryIdentifier][factoryOutputName] = value
}

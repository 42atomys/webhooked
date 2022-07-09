package factory

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"text/template"

	"github.com/rs/zerolog/log"

	"atomys.codes/webhooked/internal/valuable"
)

const ctxPipeline contextKey = "pipeline"

// newFactory creates a new factory with the given IFactory implementation.
// and initialize it.
func newFactory(f IFactory) *Factory {
	return &Factory{
		ctx:     context.Background(),
		mu:      sync.RWMutex{},
		Name:    f.Name(),
		Fn:      f.Func(),
		Config:  make(map[string]interface{}),
		Inputs:  f.DefinedInpus(),
		Outputs: f.DefinedOutputs(),
	}
}

// DeepCopy creates a deep copy of the pipeline.
func (f *Factory) DeepCopy() *Factory {
	deepCopy := &Factory{
		ctx:     f.ctx,
		mu:      sync.RWMutex{},
		Name:    f.Name,
		Fn:      f.Fn,
		Config:  make(map[string]interface{}),
		Inputs:  make([]*Var, len(f.Inputs)),
		Outputs: make([]*Var, len(f.Outputs)),
	}

	copy(deepCopy.Inputs, f.Inputs)
	copy(deepCopy.Outputs, f.Outputs)

	for k, v := range f.Config {
		deepCopy.Config[k] = v
	}

	return deepCopy
}

// GetVar returns the variable with the given name from the given slice.
// @param list the Var slice to search in
// @param name the name of the variable to search for
// @return the variable with the given name from the given slice
// @return true if the variable was found
func GetVar(list []*Var, name string) (*Var, bool) {
	for _, v := range list {
		if v.Name == name {
			return v, true
		}
	}
	return nil, false
}

// with adds a new variable to the given slice.
// @param slice the slice to add the variable to
// @param name the name of the variable
// @param value the value of the variable
// @return the new slice with the added variable
func (f *Factory) with(slice []*Var, name string, value interface{}) ([]*Var, error) {
	v, ok := GetVar(slice, name)
	if !ok {
		log.Error().Msgf("variable %s is not registered for %s", name, f.Name)
		return slice, fmt.Errorf("variable %s is not registered for %s", name, f.Name)
	}

	if reflect.TypeOf(value) != v.Type {
		log.Error().Msgf("invalid type for %s expected %s, got %s", name, v.Type.String(), reflect.TypeOf(value).String())
		return slice, fmt.Errorf("invalid type for %s expected %s, got %s", name, v.Type.String(), reflect.TypeOf(value).String())
	}

	v.Value = value
	return slice, nil
}

// WithPipelineInput adds the given pipeline input to the factory.
// only if the pipeline input is matching the factory desired input.
// Dont thow an error if the pipeline input is not matching the factory input
//
// @param name the name of the input variable
// @param value the value of the input variable
func (f *Factory) withPipelineInput(name string, value interface{}) {
	v, ok := GetVar(f.Inputs, name)
	if !ok {
		return
	}
	if reflect.TypeOf(value) != v.Type {
		return
	}
	v.Value = value
}

// WithInput adds the given input to the factory.
// @param name the name of the input variable
// @param value the value of the input variable
// @return the factory
func (f *Factory) WithInput(name string, value interface{}) *Factory {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.Inputs, _ = f.with(f.Inputs, name, value)
	return f
}

// WithID sets the id of the factory.
// @param id the id of the factory
// @return the factory
func (f *Factory) WithID(id string) *Factory {
	f.ID = id
	return f
}

// WithConfig sets the config of the factory.
// @param config the config of the factory
// @return the factory
func (f *Factory) WithConfig(config map[string]interface{}) *Factory {
	f.mu.Lock()
	defer f.mu.Unlock()

	if id, ok := config["id"]; ok {
		f.WithID(id.(string))
		delete(config, "id")
	}

	for k, v := range config {
		f.Config[k] = v
	}
	return f
}

// Input retrieve the input variable of the given name.
// @param name the name of the input variable
// @return the input variable of the given name
// @return true if the input variable was found
func (f *Factory) Input(name string) (v *Var, ok bool) {
	v, ok = GetVar(f.Inputs, name)
	if !ok {
		return nil, false
	}

	if (reflect.TypeOf(v.Value) == reflect.TypeOf(&InputConfig{})) {
		return f.processInputConfig(v)
	}

	return v, ok
}

// Output store the output variable of the given name.
// @param name the name of the output variable
// @param value the value of the output variable
// @return the factory
func (f *Factory) Output(name string, value interface{}) *Factory {
	f.Outputs, _ = f.with(f.Outputs, name, value)
	return f
}

// Identifier will return the id of the factory or the name of the factory if
// the id is not set.
func (f *Factory) Identifier() string {
	if f.ID != "" {
		return f.ID
	}
	return f.Name
}

// Run executes the factory function
func (f *Factory) Run() error {
	if err := f.Fn(f, f.Config); err != nil {
		log.Error().Err(err).Msgf("error during factory %s run", f.Name)
		return err
	}
	return nil
}

// processInputConfig process all input config struct to apply custom
// processing on the value. This is used to process the input config
// with a go template value. Example to retrieve an output of previous
// factory with `{{ .Outputs.ID.value }}`. The template is executed
// with the pipeline object as data.
//
// @param v the input config variable
// @return the processed input config variable
func (f *Factory) processInputConfig(v *Var) (*Var, bool) {
	v2 := &Var{true, reflect.TypeOf(v.Value), v.Name, &InputConfig{}}
	input := v2.Value.(*InputConfig)

	var vub = &valuable.Valuable{}
	for _, value := range v.Value.(*InputConfig).Get() {
		if strings.Contains(value, "{{") && strings.Contains(value, "}}") {
			vub.Values = append(input.Values, goTemplateValue(value, f.ctx.Value(ctxPipeline)))
		} else {
			vub.Values = append(vub.Values, value)
		}
	}

	input.Valuable = *vub
	v2.Value = input
	return v2, true
}

// goTemplateValue executes the given template with the given data.
// @param template the template to execute
// @param data the data to use for the template
// @return the result of the template execution
func goTemplateValue(tmpl string, data interface{}) string {
	t := template.New("gotmpl")
	t, err := t.Parse(tmpl)
	if err != nil {
		panic(err)
	}

	buf := new(bytes.Buffer)
	if err := t.Execute(buf, data); err != nil {
		panic(err)
	}
	return buf.String()
}

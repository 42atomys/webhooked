package factory

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"text/template"

	"github.com/rs/zerolog/log"

	"42stellar.org/webhooks/internal/valuable"
)

type InputConfig struct {
	valuable.Valuable
	Name string `mapstructure:"name"`
}

func DecodeHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if t != reflect.TypeOf(InputConfig{}) {
		return data, nil
	}

	// log.Warn().Msgf("inputConfigDecodeHook: from : %s to %s -- %+v", f.String(), t.String(), data)

	v, err := valuable.SerializeValuable(data)
	var name = ""
	for k, v := range data.(map[interface{}]interface{}) {
		if fmt.Sprintf("%v", k) == "name" {
			name = fmt.Sprintf("%s", v)
			break
		}
	}

	return &InputConfig{
		Valuable: *v,
		Name:     name,
	}, err
}

var (
	// ErrSecurityFailed is returned when the factory failed
	ErrSecurityFailed = errors.New("factory failed")
)

// FACTORY
// FACTORY
// FACTORY
// FACTORY
// FACTORY
// FACTORY
// FACTORY
// FACTORY
// FACTORY
// FACTORY
// FACTORY
// FACTORY
// FACTORY
// FACTORY
// FACTORY
// FACTORY
// FACTORY

type RunFunc func(factory *Factory, configRaw map[string]interface{}) error

type Factory struct {
	ctx context.Context
	// Name is the name of the factory function
	Name string
	// ID is the unique ID of the factory
	ID string
	// Fn is the factory function
	Fn RunFunc
	// Config is the configuration for the factory function
	Config map[string]interface{}

	Inputs  []*Var
	Outputs []*Var
}

type Var struct {
	Internal bool
	Type     reflect.Type
	Name     string
	Value    interface{}
}

type IFactory interface {
	Name() string
	DefinedInpus() []*Var
	DefinedOutputs() []*Var
	Func() RunFunc
}

func NewFactory(f IFactory) *Factory {
	return &Factory{
		ctx:     context.Background(),
		Name:    f.Name(),
		Fn:      f.Func(),
		Config:  make(map[string]interface{}),
		Inputs:  f.DefinedInpus(),
		Outputs: f.DefinedOutputs(),
	}
}

func GetVar(list []*Var, name string) (*Var, bool) {
	for _, v := range list {
		if v.Name == name {
			return v, true
		}
	}
	return nil, false
}

func (f *Factory) with(slice []*Var, name string, value interface{}) []*Var {
	v, ok := GetVar(slice, name)
	if !ok {
		log.Fatal().Msgf("variable %s is not registered for %s", name, f.Name)
	}

	if reflect.TypeOf(value) != v.Type {
		log.Fatal().Msgf("invalid type for %s expected %s, got %s", name, v.Type.String(), reflect.TypeOf(value).String())
	}

	v.Value = value
	return slice
}

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

func (f *Factory) WithInput(name string, value interface{}) *Factory {
	f.Inputs = f.with(f.Inputs, name, value)
	return f
}

func (f *Factory) WithID(id string) *Factory {
	f.ID = id
	return f
}

func (f *Factory) WithConfig(config map[string]interface{}) *Factory {
	if id, ok := config["id"]; ok {
		f.WithID(id.(string))
		delete(config, "id")
	}

	for k, v := range config {
		f.Config[k] = v
	}
	return f
}

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

func (f *Factory) Output(name string, value interface{}) *Factory {
	f.Outputs = f.with(f.Outputs, name, value)
	return f
}

func (f *Factory) Run() {
	if err := f.Fn(f, f.Config); err != nil {
		log.Fatal().Msgf("errorduring factory run %s", err)
	}
}

func (f *Factory) processInputConfig(v *Var) (*Var, bool) {
	v2 := &Var{true, reflect.TypeOf(v.Value), v.Name, &InputConfig{}}
	input := v2.Value.(*InputConfig)

	var vub = &valuable.Valuable{}
	for _, value := range v.Value.(*InputConfig).Get() {
		if strings.Contains(value, "{{") && strings.Contains(value, "}}") {
			vub.Values = append(input.Values, goTemplateValue(value, f.ctx.Value("pipeline")))
		} else {
			vub.Values = append(vub.Values, value)
		}
	}

	input.Valuable = *vub
	v2.Value = input
	return v2, true
}

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

// FACTORY DEFINITION
// FACTORY DEFINITION
// FACTORY DEFINITION
// FACTORY DEFINITION
// FACTORY DEFINITION
// FACTORY DEFINITION
// FACTORY DEFINITION
// FACTORY DEFINITION
// FACTORY DEFINITION
// FACTORY DEFINITION
// FACTORY DEFINITION
// FACTORY DEFINITION
// FACTORY DEFINITION
// FACTORY DEFINITION
// FACTORY DEFINITION
// FACTORY DEFINITION

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
		if !ok {
			return fmt.Errorf("missing input request")
		}

		factory.Output("value",
			requestVar.Value.(*http.Request).Header.Get(
				nameVar.Value.(*InputConfig).First(),
			))

		return nil
	}
}

/////////////////////////////////////////////
/////////////////////////////////////////////

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

		factory.Output("result", c.compareSlice(
			firstVar.Value.(*InputConfig).Get(),
			secondVar.Value.(*InputConfig).Get(),
		))
		return nil
	}
}

func (*compareFactory) compareSlice(slice1, slice2 []string) bool {
	// Loop two times, first to find slice1 strings not in slice2,
	// second loop to find slice2 strings not in slice1
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			for _, s2 := range slice2 {
				if s1 == s2 {
					log.Warn().Msgf("found %s - %s in both slices", s1, s2)
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

// PIPELINE
// PIPELINE
// PIPELINE
// PIPELINE
// PIPELINE
// PIPELINE
// PIPELINE
// PIPELINE
// PIPELINE
// PIPELINE
// PIPELINE
// PIPELINE
// PIPELINE
// PIPELINE
// PIPELINE
// PIPELINE
// PIPELINE
// PIPELINE
// PIPELINE

type Pipeline struct {
	factories []*Factory

	Result      interface{}
	LastResults []interface{}

	Variables, Config, Inputs map[string]interface{}

	Outputs map[string]map[string]interface{}
}

func NewPipeline() *Pipeline {
	return &Pipeline{
		Outputs:   make(map[string]map[string]interface{}),
		Variables: make(map[string]interface{}),
		Config:    make(map[string]interface{}),
		Inputs:    make(map[string]interface{}),
	}
}

func (p *Pipeline) AddFactory(f *Factory) *Pipeline {
	p.factories = append(p.factories, f)
	return p
}

func (p *Pipeline) HasFactories() bool {
	return p.FactoryCount() > 0
}

func (p *Pipeline) FactoryCount() int {
	return len(p.factories)
}

func (p *Pipeline) WantResult(result interface{}) *Pipeline {
	p.Result = result
	return p
}

func (p *Pipeline) Check() bool {
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

func (p *Pipeline) Run() *Factory {
	for _, f := range p.factories {
		f.ctx = context.WithValue(f.ctx, "pipeline", p)
		for k, v := range p.Inputs {
			f.withPipelineInput(k, v)
		}

		log.Debug().Msgf("running factory %s", f.Name)
		for _, v := range f.Inputs {
			log.Debug().Msgf("factory %s input %s = %+v", f.Name, v.Name, v.Value)
		}
		f.Run()

		for _, v := range f.Outputs {
			log.Debug().Msgf("factory %s output %s = %+v", f.Name, v.Name, v.Value)
		}

		var key string
		if f.ID != "" {
			key = f.ID
		} else {
			key = f.Name
		}

		if p.Result != nil {
			p.LastResults = make([]interface{}, 0)
		}

		for _, v := range f.Outputs {
			if p.Outputs[key] == nil {
				p.Outputs[key] = make(map[string]interface{})
			}
			p.Outputs[key][v.Name] = v.Value

			if p.Result != nil {
				p.LastResults = append(p.LastResults, v.Value)
			}
		}
	}

	return p.factories[len(p.factories)-1]
}

// MAPPING

var (
	// FunctionMap contains the map of function names to their respective functions
	// This is used to validate the function name and to get the function by name
	factoryMap = map[string]*Factory{
		"header":  NewFactory(&headerFactory{}),
		"compare": NewFactory(&compareFactory{}),
	}
)

// GetFunctionByName returns true if the function name is contained in the map
func GetFactoryByName(name string) (*Factory, bool) {
	fn, ok := factoryMap[name]
	return fn, ok
}

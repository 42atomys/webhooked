package factory

import (
	"context"
	"reflect"
	"sync"

	"atomys.codes/webhooked/internal/valuable"
)

// contextKey is used to define context key inside the factory package
type contextKey string

// InputConfig is a struct that contains the name and the value of an input.
// It is used to store the inputs of a factory. The name is used to retrieve
// the value of the input from the factory.
//
// This is used to load the inputs of a factory from the configuration file.
type InputConfig struct {
	valuable.Valuable
	Name string `mapstructure:"name"`
}

// Pipeline is a struct that contains informations about the pipeline.
// It is used to store the inputs and outputs of all factories executed
// by the pipeline and secure the result of the pipeline.
type Pipeline struct {
	mu        sync.RWMutex
	factories []*Factory

	WantedResult interface{}
	LastResults  []interface{}

	Inputs map[string]interface{}

	Outputs map[string]map[string]interface{}
}

// RunFunc is a function that is used to run a factory.
// It is used to run a factory in a pipeline.
// @param factory the factory to run
// @param configRaw the raw configuration of the factory
type RunFunc func(factory *Factory, configRaw map[string]interface{}) error

// Factory represents a factory that can be executed by the pipeline.
type Factory struct {
	ctx context.Context
	// Name is the name of the factory function
	Name string
	// ID is the unique ID of the factory
	ID string
	// Fn is the factory function
	Fn RunFunc
	// Protect following fields
	mu sync.RWMutex
	// Config is the configuration for the factory function
	Config map[string]interface{}
	// Inputs is the inputs of the factory
	Inputs []*Var
	// Outputs is the outputs of the factory
	Outputs []*Var
}

// Var is a struct that contains the name and the value of an input or output.
// It is used to store the inputs and outputs of a factory.
type Var struct {
	// Internal is to specify if the variable is an internal provided variable
	Internal bool
	// Type is the type of the wanted variable
	Type reflect.Type
	// Name is the name of the variable
	Name string
	// Value is the value of the variable, type can be retrieved from Type field
	Value interface{}
}

// IFactory is an interface that represents a factory.
type IFactory interface {
	// Name is the name of the factory function
	// The name must be unique in the registry
	// @return the name of the factory function
	Name() string
	// DefinedInputs returns the wanted inputs of the factory used
	// by the function during the execution of the pipeline
	DefinedInpus() []*Var
	// DefinedOutputs returns the wanted outputs of the factory used
	// by the function during the execution of the pipeline
	DefinedOutputs() []*Var
	// Func is used to build the factory function
	// @return the factory function
	Func() RunFunc
}

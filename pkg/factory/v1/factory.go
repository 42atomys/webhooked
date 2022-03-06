package factory

// FactoryFunc is the function signature for a factory function
// @param configRaw is the raw configuration for the factory
// @param lastOuput is the last output from the previous factory
// @param inputs is the list of additional inputs for the factory
// @return the output of the factory
// @return an error if the factory function fails or when the comparation is not valid
type FactoryFunc func(configRaw map[string]interface{}, lastOuput string, inputs ...interface{}) (string, error)

// Factory represents a factory function that can be used to run a simple or
// complex chained factory run.
type Factory struct {
	// Name is the name of the factory function
	Name string
	// Fn is the factory function
	Fn FactoryFunc
	// Config is the configuration for the factory function
	Config map[string]interface{}
}

var (
	// FunctionMap contains the map of function names to their respective functions
	// This is used to validate the function name and to get the function by name
	FunctionMap = map[string]FactoryFunc{
		"getHeader":              getHeader,
		"compareWithStaticValue": compareWithStaticValue,
	}
)

// GetFunctionByName returns true if the function name is contained in the map
func GetFunctionByName(name string) (FactoryFunc, bool) {
	fn, ok := FunctionMap[name]
	return fn, ok
}

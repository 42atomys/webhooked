package factory

import (
	"fmt"
)

var (
	// FunctionMap contains the map of function names to their respective functions
	// This is used to validate the function name and to get the function by name
	factoryMap = map[string]*Factory{
		"debug":           newFactory(&debugFactory{}),
		"header":          newFactory(&headerFactory{}),
		"compare":         newFactory(&compareFactory{}),
		"hasPrefix":       newFactory(&hasPrefixFactory{}),
		"hasSuffix":       newFactory(&hasSuffixFactory{}),
		"generateHmac256": newFactory(&generateHMAC256Factory{}),
	}
)

// GetFunctionByName returns true if the function name is contained in the map
func GetFactoryByName(name string) (*Factory, bool) {
	fn, ok := factoryMap[name]
	return fn, ok
}

// Register a new factory in the factory map with the built-in factory name
func Register(factory IFactory) error {
	if _, ok := GetFactoryByName(factory.Name()); ok {
		return fmt.Errorf("factory %s is already exist", factory.Name())
	}
	factoryMap[factory.Name()] = newFactory(factory)
	return nil
}

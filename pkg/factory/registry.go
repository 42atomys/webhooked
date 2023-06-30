package factory

import (
	"fmt"
	"strings"
)

var (
	// FunctionMap contains the map of function names to their respective functions
	// This is used to validate the function name and to get the function by name
	factoryMap = map[string]IFactory{
		"debug":           &debugFactory{},
		"header":          &headerFactory{},
		"compare":         &compareFactory{},
		"hasPrefix":       &hasPrefixFactory{},
		"hasSuffix":       &hasSuffixFactory{},
		"generateHmac256": &generateHMAC256Factory{},
	}
)

// GetFactoryByName returns true if the function name is contained in the map
func GetFactoryByName(name string) (*Factory, bool) {
	for k, v := range factoryMap {
		if strings.EqualFold(k, name) {
			return newFactory(v), true
		}
	}

	return nil, false
}

// Register a new factory in the factory map with the built-in factory name
func Register(factory IFactory) error {
	if _, ok := GetFactoryByName(factory.Name()); ok {
		return fmt.Errorf("factory %s is already exist", factory.Name())
	}
	factoryMap[factory.Name()] = factory
	return nil
}

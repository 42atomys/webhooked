package factory

import "github.com/mitchellh/mapstructure"

// compareWithStaticValueConfig is the configuration for the compareWithStaticValue factory
// @field value is the value to compare with
// @field values is the list of values to compare with
type compareWithStaticValueConfig struct {
	// Value is the value to compare with
	Value string `mapstructure:"value"`
	// Values is the list of values to compare with
	Values []string `mapstructure:"values"`
}

// compareWithStaticValue will compare the last output with the given value in
// the configuration
// @param configRaw is the raw configuration for the factory
// @param lastOuput is the last output from the previous factory
// @param inputs is the list of additional inputs for the factory
// @return result of comparation between last output and value
// @return an error if cannot decode configuration
//
// factory developer usage :
// additional inputs: NONE
//
// factory example:
// - compareWithStaticValue:
//     value: 'test'
//     values: ['foo', 'bar']
//
func compareWithStaticValue(configRaw map[string]interface{}, lastOuput string, inputs ...interface{}) (string, error) {
	config := &compareWithStaticValueConfig{}
	if err := mapstructure.Decode(configRaw, &config); err != nil {
		return "", err
	}

	config.Values = append(config.Values, config.Value)

	for _, v := range config.Values {
		if v == lastOuput {
			return "t", nil
		}
	}

	return "f", nil
}

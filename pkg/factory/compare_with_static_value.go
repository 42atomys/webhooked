package factory

import (
	"errors"

	"42stellar.org/webhooks/internal/valuable"
)

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
//     valueFrom:
//       envRef: 'PRIVATE_TOKEN'
//
func compareWithStaticValue(configRaw map[string]interface{}, lastOuput string, inputs ...interface{}) (string, error) {
	if len(inputs) > 0 {
		return "", errors.New("compareWithStaticValue factory does not accept additional inputs")
	}

	config := valuable.Valuable{}
	if err := valuable.Decode(configRaw, &config); err != nil {
		return "", err
	}

	if (len(config.Get()) == 0 && lastOuput == "") || config.Contains(lastOuput) {
		return "t", nil
	}

	return "f", nil
}

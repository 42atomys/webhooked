package factory

import (
	"errors"
	"net/http"

	"github.com/mitchellh/mapstructure"
)

// getHeaderConfig is the configuration for the getHeader factory
// @field name is the name of the header to get
type getHeaderConfig struct {
	// Name is the name of the header to get
	Name string `mapstructure:"name"`
}

// getHeader returns the value of the header with the given name
// @param configRaw is the raw configuration for the factory
// @param lastOuput is the last output from the previous factory
// @param inputs is the list of additional inputs for the factory
// @return the header value
// @return an error if cannot decode configuration
//
// factory developer usage :
// additional inputs: [0] needs to be an *http.Request not nil
//
// factory example:
// - getHeader:
//     name: "X-Request-Id"
//
func getHeader(configRaw map[string]interface{}, lastOuput string, inputs ...interface{}) (string, error) {
	if len(inputs) != 1 {
		return "", errors.New("getHeader factory requires 1 input")
	}

	header, ok := inputs[0].(http.Header)
	if !ok {
		return "", errors.New("getHeader factory requires input to be an http.Header object")
	}

	config := &getHeaderConfig{}
	if err := mapstructure.Decode(configRaw, &config); err != nil {
		return "", err
	}

	return header.Get(config.Name), nil
}

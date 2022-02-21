package factory

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
)

var (
	// ErrSecurityFailed is returned when the factory failed
	ErrSecurityFailed = errors.New("factory failed")
)

// RunnerFunc is the function signature for the runner to compare and
// run the factory layer
// TODO docs
type RunnerExternalFunc func(factory *Factory, lastOutput string, defaultFn RunnerFunc) (string, error)
type RunnerFunc func(factory *Factory, lastOutput string) (string, error)

// Run the factory chain for the given factories
// TODO: Docs
func Run(factories []*Factory, runnerFn RunnerExternalFunc) (bool, error) {
	var lastOutput string
	var err error

	if (factories == nil) || (len(factories) == 0) {
		return true, nil
	}

	for _, factory := range factories {
		if runnerFn != nil {
			lastOutput, err = runnerFn(factory, lastOutput, internalRunner)
		} else {
			lastOutput, err = internalRunner(factory, lastOutput)
		}

		if err != nil {
			log.Error().Err(err).Msg("Error while processing security layer")
			return false, err
		}
		log.Debug().Str("factory", factory.Name).Str("output", lastOutput).Msg("Security layer output")
	}

	if (lastOutput != "t") && (lastOutput != "f") {
		return false, fmt.Errorf("security layer didn't return a valid value. Got %s want a compare function at the end", lastOutput)
	}

	return lastOutput == "t", nil
}

// internalRunner is the runner function used when no external runner is provided
// Basic operation can be runned here but complex operations like get information
// from an appliation context should be done in the external runner
func internalRunner(factory *Factory, lastOutput string) (string, error) {
	switch factory.Name {
	case "getHeader":
		return "", errors.New("getHeader cannot be implemented internally. Please use the external runner")
	case "compareWithStaticValue":
		return compareWithStaticValue(factory.Config, lastOutput)
	}
	return "", fmt.Errorf("factory %s not implemented", factory.Name)
}

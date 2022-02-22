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
// run by an external caller to override the internal runner
// @param factory is the factory to run
// @param lastOutput is the last output from the previous factory
// @param defaultRnner is the default runner to use if you want to use the
//                     internal runner inside your runner
// @return the output of the factory
// @return an error if the factory failed
type RunnerExternalFunc func(factory *Factory, lastOutput string, defaultFn RunnerFunc) (string, error)

// RunnerFunc is the function signature for the runner to compare and
// run the factory inside the Factory Builder.
// @param factory is the factory to run
// @param lastOutput is the last output from the previous factory
// @return the output of the factory
// @return an error if the factory failed
type RunnerFunc func(factory *Factory, lastOutput string) (string, error)

// Run the factory chain for the given factories in order
// A run is made by calling the runner function for each factory
// The lat output of the previous factory is passed to the next factory
// At the end of the chain the last factory needs to be a comparator (start with compareWith)
// @param factories is the list of factories to run
// @param runnerFn is the runner function to use to run the factories if necessary.
//                 If nil, the internal runner will be used
// @return true if the chain is valid
// @return an error if the chain fail
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

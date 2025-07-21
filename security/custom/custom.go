// In package security/custom
package custom

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"text/template"

	"github.com/42atomys/webhooked/format"
	"github.com/42atomys/webhooked/internal/valuable"
	"github.com/go-sprout/sprout"
	"github.com/go-sprout/sprout/group/all"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
)

// CustomSecuritySpec is a security specification that allows defining custom
// conditions.
// This condition is evaluated using Go templates, allowing dynamic and highly
// customizable security rules.
type CustomSecuritySpec struct {
	Condition *valuable.Valuable `json:"condition"`

	template    *template.Template
	builderPool sync.Pool
}

// EnsureConfigurationCompleteness ensures that the CustomSecuritySpec is properly
// configured.
// The only requirement is that a condition is provided.
//
// Returns:
//   - error: An error if the configuration is incomplete.
func (s *CustomSecuritySpec) EnsureConfigurationCompleteness() error {
	if s.Condition == nil || s.Condition.First() == "" {
		return errors.New("condition is required")
	}
	return nil
}

// Initialize initializes the CustomSecuritySpec.
// This method initializes the template and the builder pool.
//
// Returns:
//   - error: An error if the initialization process fails.
func (s *CustomSecuritySpec) Initialize() error {
	var err error

	sproutHandler := sprout.New(sprout.WithGroups(all.RegistryGroup()))

	s.template, err = template.New("condition").Funcs(sproutHandler.Build()).Parse(s.Condition.First())
	if err != nil {
		return err
	}

	s.builderPool = sync.Pool{
		New: func() any {
			return new(strings.Builder)
		},
	}

	return nil
}

// IsSecure evaluates the custom security condition specified in the CustomSecuritySpec.
// The condition is parsed and executed using Go templates, with context
// information about the request.
// If the condition evaluates to "true", the request is considered secure.
// Otherwise, it is rejected.
//
// Parameters:
//   - ctx: The fasthttp.RequestCtx containing all the request details.
//
// Returns:
//   - bool: True if the condition is met, otherwise false.
//   - error: An error if the validation process fails (e.g., template parsing or execution errors).
func (s *CustomSecuritySpec) IsSecure(ctx *fasthttp.RequestCtx) (bool, error) {
	// Acquire a builder from the pool
	sb := s.builderPool.Get().(*strings.Builder)
	sb.Reset()

	// Execute the template with the request context
	if err := s.template.Execute(sb, format.GenerateRequestContext(ctx)); err != nil {
		return false, err
	}

	result, err := strconv.ParseBool(strings.Trim(sb.String(), "\n"))
	if err != nil {
		return false, fmt.Errorf("failed to parse custom security condition result as boolean: %w", err)
	}

	s.builderPool.Put(sb)
	log.Debug().Str("condition", s.Condition.First()).Bool("result", result).Msgf("custom security condition evaluated")
	return result, nil
}

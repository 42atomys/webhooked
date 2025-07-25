// In package security/custom
package custom

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/42atomys/webhooked/format"
	"github.com/42atomys/webhooked/internal/fasthttpz"
	"github.com/42atomys/webhooked/internal/valuable"
	"github.com/rs/zerolog/log"
)

// CustomSecuritySpec is a security specification that allows defining custom
// conditions.
// This condition is evaluated using Go templates, allowing dynamic and highly
// customizable security rules.
type CustomSecuritySpec struct {
	Condition *valuable.Valuable `json:"condition"`

	formatter *format.Formatting
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

	s.formatter, err = format.New(format.Specs{
		TemplateString: s.Condition.First(),
	})
	if err != nil {
		return err
	}

	if !s.formatter.HasTemplate() {
		return errors.New("condition template is required")
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
//   - ctx: The fasthttpz.RequestCtx containing all the request details.
//
// Returns:
//   - bool: True if the condition is met, otherwise false.
//   - error: An error if the validation process fails (e.g., template parsing or execution errors).
func (s *CustomSecuritySpec) IsSecure(ctx context.Context, rctx *fasthttpz.RequestCtx) (bool, error) {

	bytes, err := s.formatter.Format(ctx, map[string]any{})
	if err != nil {
		return false, fmt.Errorf("failed to execute custom security condition template: %w", err)
	}

	result, err := strconv.ParseBool(strings.Trim(string(bytes), "\n"))
	if err != nil {
		return false, fmt.Errorf("failed to parse custom security condition result as boolean: %w", err)
	}

	log.Debug().Str("condition", s.Condition.First()).Bool("result", result).Msgf("custom security condition evaluated")
	return result, nil
}

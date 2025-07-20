package config

import (
	"fmt"

	"github.com/42atomys/webhooked/format"
	securityNoop "github.com/42atomys/webhooked/security/noop"
	"github.com/rs/zerolog/log"
)

var (
	validators = []func(*Webhook) error{
		ensureResponseCompleteness,
		ensureSecurityCompleteness,
		ensureStorageCompleteness,
	}
)

func validateAndSetDefaults(wh *Webhook) error {
	for _, validator := range validators {
		if err := validator(wh); err != nil {
			return fmt.Errorf("error validating webhook %s: %w", wh.Name, err)
		}
	}

	return nil
}

func ensureResponseCompleteness(wh *Webhook) error {
	if wh.Response.ContentType == "" {
		wh.Response.ContentType = "application/json"
	}

	if wh.Response.StatusCode == 0 {
		wh.Response.StatusCode = 200
	} else if wh.Response.StatusCode < 100 || wh.Response.StatusCode > 599 {
		return ErrInvalidStatusCode
	}

	// Ensure response formatting is initialized correctly when not provided
	if wh.Response.Formatting == nil {
		formatting, err := format.New(format.Specs{TemplateString: string(defaultResponseTemplate)})
		if err != nil {
			return fmt.Errorf("error initializing default response formatting: %w", err)
		}
		wh.Response.Formatting = formatting
	}

	if !wh.Response.Formatting.HasTemplate() {
		wh.Response.Formatting.WithTemplate(defaultResponseTemplate)
	}

	return nil
}

func ensureSecurityCompleteness(wh *Webhook) error {
	if wh.Security.Type == "" {
		wh.Security.Type = "noop"
		wh.Security.Specs = &securityNoop.NoopSecuritySpec{}
		log.Warn().Msg("No security type specified, defaulting to noop")
	}

	if err := wh.Security.Specs.EnsureConfigurationCompleteness(); err != nil {
		return fmt.Errorf("error validating security %s: %w", wh.Security.Type, err)
	}

	if err := wh.Security.Specs.Initialize(); err != nil {
		return fmt.Errorf("error initializing security %s: %w", wh.Security.Type, err)
	}

	return nil
}

func ensureStorageCompleteness(wh *Webhook) error {
	for _, storage := range wh.Storage {
		if err := storage.Specs.EnsureConfigurationCompleteness(); err != nil {
			return fmt.Errorf("error validating storage %s: %w", storage.Type, err)
		}

		if err := storage.Specs.Initialize(); err != nil {
			return fmt.Errorf("error initializing storage %s: %w", storage.Type, err)
		}

		if !storage.Formatting.HasTemplateCompiled() {
			storage.Formatting.WithTemplate(defaultPayloadTemplate)
		}
	}

	return nil
}

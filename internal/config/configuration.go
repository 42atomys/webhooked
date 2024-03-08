package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"

	"atomys.codes/webhooked/pkg/factory"
	"atomys.codes/webhooked/pkg/storage"
)

var (
	currentConfig = &Configuration{}
	// ErrSpecNotFound is returned when the spec is not found
	ErrSpecNotFound = errors.New("spec not found")
	// defaultPayloadTemplate is the default template for the payload
	// when no template is defined
	defaultPayloadTemplate = `{{ .Payload }}`
	// defaultResponseTemplate is the default template for the response
	// when no template is defined
	defaultResponseTemplate = ``
)

// Load loads the configuration from the configuration file
// if an error is occurred, it will be returned
func Load(cfgFile string) error {
	var k = koanf.New(".")

	// Load YAML config.
	if err := k.Load(file.Provider(cfgFile), yaml.Parser()); err != nil {
		log.Error().Msgf("error loading config: %v", err)
	}

	// Load from environment variables
	err := k.Load(env.ProviderWithValue("WH_", ".", func(s, v string) (string, interface{}) {
		key := strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, "WH_")), "_", ".", -1)

		return key, v
	}), nil)
	if err != nil {
		log.Error().Msgf("error loading config: %v", err)
	}

	if os.Getenv("WH_DEBUG") == "true" {
		k.Print()
	}

	err = k.UnmarshalWithConf("", &currentConfig, koanf.UnmarshalConf{
		DecoderConfig: &mapstructure.DecoderConfig{
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToTimeDurationHookFunc(),
				factory.DecodeHook,
			),
			Result:           &currentConfig,
			WeaklyTypedInput: true,
		},
	})
	if err != nil {
		log.Fatal().Msgf("error loading config: %v", err)
		return err
	}

	for _, spec := range currentConfig.Specs {
		if err := loadSecurityFactory(spec); err != nil {
			return err
		}

		if spec.Formatting, err = loadTemplate(spec.Formatting, nil, defaultPayloadTemplate); err != nil {
			return fmt.Errorf("configured storage for %s received an error: %s", spec.Name, err.Error())
		}

		if err = loadStorage(spec); err != nil {
			return fmt.Errorf("configured storage for %s received an error: %s", spec.Name, err.Error())
		}

		if spec.Response.Formatting, err = loadTemplate(spec.Response.Formatting, nil, defaultResponseTemplate); err != nil {
			return fmt.Errorf("configured response for %s received an error: %s", spec.Name, err.Error())
		}
	}

	log.Info().Msgf("Load %d configurations", len(currentConfig.Specs))
	return Validate(currentConfig)
}

// loadSecurityFactory loads the security factory for the given spec
// if an error is occurred, return an error
func loadSecurityFactory(spec *WebhookSpec) error {
	spec.SecurityPipeline = factory.NewPipeline()
	for _, security := range spec.Security {
		for securityName, securityConfig := range security {
			f, ok := factory.GetFactoryByName(securityName)
			if !ok {
				return fmt.Errorf("security factory \"%s\" in %s specification is not a valid factory", securityName, spec.Name)
			}

			for _, input := range securityConfig.Inputs {
				f.WithInput(input.Name, input)
			}

			spec.SecurityPipeline.AddFactory(f.WithID(securityConfig.ID).WithConfig(securityConfig.Specs))
		}
	}
	log.Debug().Msgf("%d security factories loaded for spec %s", spec.SecurityPipeline.FactoryCount(), spec.Name)
	return nil
}

// Validate the configuration file and her content
func Validate(config *Configuration) error {
	var uniquenessName = make(map[string]bool)
	var uniquenessUrl = make(map[string]bool)

	for _, spec := range config.Specs {
		log.Debug().Str("name", spec.Name).Msgf("Load spec: %+v", spec)

		// Validate the uniqueness of all name
		if _, ok := uniquenessName[spec.Name]; ok {
			return fmt.Errorf("specification name %s must be unique", spec.Name)
		}
		uniquenessName[spec.Name] = true

		// Validate the uniqueness of all entrypoints
		if _, ok := uniquenessUrl[spec.EntrypointURL]; ok {
			return fmt.Errorf("specification entrypoint url %s must be unique", spec.EntrypointURL)
		}
		uniquenessUrl[spec.EntrypointURL] = true
	}

	return nil
}

// loadStorage registers the storage and validate it
// if the storage is not found or an error is occurred during the
// initialization or connection, the error is returned during the
// validation
func loadStorage(spec *WebhookSpec) (err error) {
	for _, s := range spec.Storage {
		s.Client, err = storage.Load(s.Type, s.Specs)
		if err != nil {
			return fmt.Errorf("storage %s cannot be loaded properly: %s", s.Type, err.Error())
		}

		if s.Formatting, err = loadTemplate(s.Formatting, spec.Formatting, defaultPayloadTemplate); err != nil {
			return fmt.Errorf("storage %s cannot be loaded properly: %s", s.Type, err.Error())
		}
	}

	log.Debug().Msgf("%d storages loaded for spec %s", len(spec.Storage), spec.Name)
	return
}

// loadTemplate loads the template for the given `spec`. When no spec is defined
// we try to load the template from the parentSpec and fallback to the default
// template if parentSpec is not given.
func loadTemplate(spec, parentSpec *FormattingSpec, defaultTemplate string) (*FormattingSpec, error) {
	if spec == nil {
		spec = &FormattingSpec{}
	}

	if spec.TemplateString != "" {
		spec.Template = spec.TemplateString
		return spec, nil
	}

	if spec.TemplatePath != "" {
		file, err := os.OpenFile(spec.TemplatePath, os.O_RDONLY, 0666)
		if err != nil {
			return spec, err
		}
		defer file.Close()

		var buffer bytes.Buffer
		_, err = io.Copy(&buffer, file)
		if err != nil {
			return spec, err
		}

		spec.Template = buffer.String()
		return spec, nil
	}

	if parentSpec != nil {
		if parentSpec.Template == "" {
			var err error
			parentSpec, err = loadTemplate(parentSpec, nil, defaultTemplate)
			if err != nil {
				return spec, err
			}
		}
		spec.Template = parentSpec.Template
	} else {
		spec.Template = defaultTemplate
	}

	return spec, nil
}

// Current returns the aftual configuration
func Current() *Configuration {
	return currentConfig
}

// GetSpec returns the spec for the given name, if no entry
// is found, ErrSpecNotFound is returned
func (c *Configuration) GetSpec(name string) (*WebhookSpec, error) {
	for _, spec := range c.Specs {
		if spec.Name == name {
			return spec, nil
		}
	}

	log.Error().Err(ErrSpecNotFound).Msgf("Spec %s not found", name)
	return nil, ErrSpecNotFound

}

// GetSpecByEndpoint returns the spec for the given endpoint, if no entry
// is found, ErrSpecNotFound is returned
func (c *Configuration) GetSpecByEndpoint(endpoint string) (*WebhookSpec, error) {
	for _, spec := range c.Specs {
		if spec.EntrypointURL == endpoint {
			return spec, nil
		}
	}

	return nil, ErrSpecNotFound
}

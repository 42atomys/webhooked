package config

import (
	"errors"
	"os"
	"strings"
	"sync"

	"github.com/42atomys/webhooked/format"
	"github.com/42atomys/webhooked/internal/valuable"
	"github.com/42atomys/webhooked/security"
	"github.com/42atomys/webhooked/storage"
	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog/log"
)

type Config struct {
	APIVersion APIVersion `json:"apiVersion"`
	Kind       Kind       `json:"kind"`
	Metadata   Metadata   `json:"metadata"`
	Specs      []*Spec    `json:"specs"`
}

type APIVersion string
type Kind string

const (
	APIVersionV1Alpha2 APIVersion = "v1alpha2"
	KindConfiguration  Kind       = "Configuration"
)

type Metadata struct {
	Name string `json:"name"`
}

type Spec struct {
	MetricsEnabled bool        `json:"metricsEnabled"`
	Throttling     *Throttling `json:"throttling"`
	Webhooks       []*Webhook  `json:"specs"`
}

type Throttling struct {
	Enabled bool `json:"enabled"`
	// MaxRequests is the maximum number of requests that can be processed
	// in a given time window.
	MaxRequests int `json:"maxRequests"`
	// Window is the time window in seconds.
	Window int `json:"window"`
	// Burst is the number of requests that can be processed in a single
	// burst.
	Burst int `json:"burst"`
	// BurstWindow is the time window in seconds for the burst.
	BurstWindow int `json:"burstWindow"`
	// QueueCapacity is the maximum number of requests that can be queued.
	QueueCapacity int `json:"queueCapacity"`
	// QueueTimeout is the maximum time a request can be queued.
	QueueTimeout int `json:"queueTimeout"`
	// QueueTimeoutCode is the status code to return when the queue times out.
	QueueTimeoutCode int `json:"queueTimeoutCode"`
}

type Webhook struct {
	Name          string             `json:"name"`
	EntrypointURL string             `json:"entrypointUrl"`
	Security      security.Security  `json:"security"`
	Storage       []*storage.Storage `json:"storage"`
	Response      Response           `json:"response"`
}

type TypedSpec[T any, S any] struct {
	Type  T `json:"type"`
	Specs S `json:"specs"`
}

type Response struct {
	Formatting  *format.Formatting `json:"formatting"`
	StatusCode  int                `json:"statusCode"`
	ContentType string             `json:"contentType"`
}

var (
	currentConfig = &Config{}
	// ErrSpecNotFound is returned when the spec is not found
	ErrSpecNotFound = errors.New("spec not found")
	// ErrInvalidStatusCode is returned when the status code is invalid
	ErrInvalidStatusCode = errors.New("invalid status code")
	// defaultPayloadTemplate is the default template for the payload
	// when no template is defined
	defaultPayloadTemplate = []byte(`{{ .Payload }}`)
	// defaultResponseTemplate is the default template for the response
	// when no template is defined
	defaultResponseTemplate = []byte(``)
	// webhooksPrefix is the prefix for the webhooks path in the URL
	// e.g. /webhooks/v1alpha2/github
	webhooksPrefix = []byte("/webhooks")
)

var (
	mutex = &sync.RWMutex{}
)

func Load(path string) error {
	mutex.Lock()
	defer mutex.Unlock()

	var k = koanf.New(".")

	// File provider
	fileProvider := file.Provider(path)
	if err := fileProvider.Watch(func(event any, err error) {
		if err != nil {
			log.Error().Msgf("error watching config file: %v", err)
		}

		log.Info().Msgf("config file changed, reloading config...")
		_ = fileProvider.Unwatch()
		if err := Load(path); err != nil {
			log.Error().Msgf("error reloading config: %v", err)
		}
	}); err != nil {
		log.Error().Msgf("error watching config file: %v", err)
		return err
	}

	// Load YAML config.
	if err := k.Load(fileProvider, yaml.Parser()); err != nil {
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
		return err
	}

	if os.Getenv("WH_DEBUG") == "true" {
		k.Print()
	}

	err = k.UnmarshalWithConf("", &currentConfig, koanf.UnmarshalConf{
		DecoderConfig: &mapstructure.DecoderConfig{
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				security.DecodeHook,
				format.DecodeHook,
				storage.DecodeHook,
				mapstructure.StringToTimeDurationHookFunc(),
				valuable.MapToValuableHookFunc(),
			),
			Result:           &currentConfig,
			WeaklyTypedInput: true,
		},
	})
	if err != nil {
		log.Fatal().Msgf("error loading config: %v", err)
		return err
	}

	webhooksCount := 0
	for _, spec := range currentConfig.Specs {
		for _, wh := range spec.Webhooks {
			if err := validateAndSetDefaults(wh); err != nil {
				return err
			}

			webhooksCount++
		}
	}

	log.Info().Msgf("Load %d configurations with %d webhooks from %s", len(currentConfig.Specs), webhooksCount, path)
	return nil
}

func Current() *Config {
	return currentConfig
}

func FetchWebhookByPath(path []byte) (*Webhook, error) {
	webhooksPrefixLen := len(webhooksPrefix) + len(currentConfig.APIVersion) + 1 // 1 for the slash
	if len(path) < webhooksPrefixLen {
		return nil, ErrSpecNotFound
	}

	path = path[webhooksPrefixLen:]
	for _, spec := range currentConfig.Specs {
		for _, wh := range spec.Webhooks {
			if wh.EntrypointURL == string(path) {
				return wh, nil
			}
		}
	}

	return nil, ErrSpecNotFound
}

func WebhooksEndpointPrefix() []byte {
	return webhooksPrefix
}

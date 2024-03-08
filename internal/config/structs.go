package config

import (
	"atomys.codes/webhooked/pkg/factory"
	"atomys.codes/webhooked/pkg/storage"
)

// Configuration is the struct contains all the configuration
// defined in the webhooks yaml file
type Configuration struct {
	// APIVerion is the version of the API that will be used
	APIVersion string `mapstructure:"apiVersion" json:"apiVersion"`
	// Observability is the configuration for observability
	Observability Observability `mapstructure:"observability" json:"observability"`
	// Specs is the configuration for the webhooks specs
	Specs []*WebhookSpec `mapstructure:"specs" json:"specs"`
}

// Observability is the struct contains the configuration for observability
// defined in the webhooks yaml file.
type Observability struct {
	// MetricsEnabled is the flag to enable or disable the prometheus metrics
	// endpoint and expose the metrics
	MetricsEnabled bool `mapstructure:"metricsEnabled" json:"metricsEnabled"`
}

// WebhookSpec is the struct contains the configuration for a webhook spec
// defined in the webhooks yaml file.
type WebhookSpec struct {
	// Name is the name of the webhook spec. It must be unique in the configuration
	// file. It is used to identify the webhook spec in the configuration file
	// and is defined by the user
	Name string `mapstructure:"name" json:"name"`
	// EntrypointURL is the URL of the entrypoint of the webhook spec. It must
	// be unique in the configuration file. It is defined by the user
	// It is used to identify the webhook spec when receiving a request
	EntrypointURL string `mapstructure:"entrypointUrl" json:"entrypointUrl"`
	// Security is the configuration for the security of the webhook spec
	// It is defined by the user and can be empty. See HasSecurity() method
	// to know if the webhook spec has security
	Security []map[string]Security `mapstructure:"security" json:"-"`
	// Format is used to define the payload format sent by the webhook spec
	// to all storages. Each storage can have its own format. When this
	// configuration is empty, the default formatting setting is used (body as JSON)
	// It is defined by the user and can be empty. See HasGlobalFormatting() method
	// to know if the webhook spec has format
	Formatting *FormattingSpec `mapstructure:"formatting" json:"-"`
	// SecurityPipeline is the security pipeline of the webhook spec
	// It is defined by the configuration loader. This field is not defined
	// by the user and cannot be overridden
	SecurityPipeline *factory.Pipeline `mapstructure:"-" json:"-"`
	// Storage is the configuration for the storage of the webhook spec
	// It is defined by the user and can be empty.
	Storage []*StorageSpec `mapstructure:"storage" json:"-"`
	// Response is the configuration for the response of the webhook sent
	// to the caller. It is defined by the user and can be empty.
	Response ResponseSpec `mapstructure:"response" json:"-"`
}

type ResponseSpec struct {
	// Formatting is used to define the response body sent by webhooked
	// to the webhook caller. When this configuration is empty, no response
	// body is sent. It is defined by the user and can be empty.
	Formatting *FormattingSpec `mapstructure:"formatting" json:"-"`
	// HTTPCode is the HTTP code of the response. It is defined by the user
	// and can be empty. (default: 200)
	HttpCode int `mapstructure:"httpCode" json:"httpCode"`
	// ContentType is the content type of the response. It is defined by the user
	// and can be empty. (default: plain/text)
	ContentType string `mapstructure:"contentType" json:"contentType"`
}

// Security is the struct contains the configuration for a security
// defined in the webhooks yaml file.
type Security struct {
	// ID is the ID of the security. It must be unique in the configuration
	// file. It is defined by the user and is used to identify the security
	// factory as .Outputs
	ID string `mapstructure:"id"`
	// Inputs is the configuration for the inputs of the security. It is
	// defined by the user and following the specification of the security
	// factory
	Inputs []*factory.InputConfig `mapstructure:"inputs"`
	// Specs is the configuration for the specs of the security. It is
	// defined by the user and following the specification of the security
	// factory
	Specs map[string]interface{} `mapstructure:",remain"`
}

// StorageSpec is the struct contains the configuration for a storage
// defined in the webhooks yaml file.
type StorageSpec struct {
	// Type is the type of the storage. It must be a valid storage type
	// defined in the storage package.
	Type string `mapstructure:"type" json:"type"`
	// Specs is the configuration for the storage. It is defined by the user
	// following the storage type specification
	// NOTE: this field is hidden for json to prevent mistake of the user
	//       when he use the custom formatting option and leak credentials
	Specs map[string]interface{} `mapstructure:"specs" json:"-"`
	// Format is used to define the payload format sent by the webhook spec
	// to this storage. If not defined, the format of the webhook spec is
	// used.
	// It is defined by the user and can be empty. See HasFormatting() method
	// to know if the webhook spec has format
	Formatting *FormattingSpec `mapstructure:"formatting" json:"-"`
	// Client is the storage client. It is defined by the configuration loader
	// and cannot be overridden
	Client storage.Pusher `mapstructure:"-" json:"-"`
}

// FormattingSpec is the struct contains the configuration to formatting the
// payload of the webhook spec. The field TempalteString is prioritized
// over the field TemplatePath when both are defined.
type FormattingSpec struct {
	// TemplatePath is the path to the template used to formatting the payload
	TemplatePath string `mapstructure:"templatePath"`
	// TemplateString is a plaintext template used to formatting the payload
	TemplateString string `mapstructure:"templateString"`
	// ResolvedTemplate is the template after resolving the template variables
	// It is defined by the configuration loader and cannot be overridden
	Template string `mapstructure:"-"`
}

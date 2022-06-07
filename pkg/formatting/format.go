package formatting

import (
	"bytes"
	"fmt"
	"net/http"
	"text/template"

	"github.com/rs/zerolog/log"

	"atomys.codes/webhooked/internal/config"
)

type TemplateData struct {
	tmplString string
	data       map[string]interface{}
}

// NewTemplateData returns a new TemplateData instance. It takes the template
// string as a parameter. The template string is the string that will be used
// to render the template. The data is the map of data that will be used to
// render the template.
func NewTemplateData(tmplString string) *TemplateData {
	return &TemplateData{
		tmplString: tmplString,
		data:       make(map[string]interface{}),
	}
}

// WithData adds a key-value pair to the data map. The key is the name of the
// variable and the value is the value of the variable.
func (d *TemplateData) WithData(name string, data interface{}) *TemplateData {
	d.data[name] = data
	return d
}

// WithRequest adds a http.Request object to the data map. The key of request is
// "Request".
func (d *TemplateData) WithRequest(r *http.Request) *TemplateData {
	d.WithData("Request", r)
	return d
}

// WithPayload adds a payload to the data map. The key of payload is "Payload".
// The payload is basically the body of the request.
func (d *TemplateData) WithPayload(payload []byte) *TemplateData {
	d.WithData("Payload", string(payload))
	return d
}

// WithSpec adds a webhookspec to the data map. The key of spec is "Spec".
func (d *TemplateData) WithSpec(spec *config.WebhookSpec) *TemplateData {
	d.WithData("Spec", spec)
	return d
}

// WithStorage adds a storage spec to the data map.
// The key of storage is "Storage".
func (d *TemplateData) WithStorage(spec *config.StorageSpec) *TemplateData {
	d.WithData("Storage", spec)
	return d
}

// WithConfig adds the current config to the data map.
// The key of config is "Config".
func (d *TemplateData) WithConfig() *TemplateData {
	d.WithData("Config", config.Current())
	return d
}

// Render returns the rendered template string. It takes the template string
// from the TemplateData instance and the data stored in the TemplateData
// instance. It returns an error if the template string is invalid or when
// rendering the template fails.
func (d *TemplateData) Render() (string, error) {
	log.Debug().Msgf("rendering template: %s", d.tmplString)

	t := template.New("formattingTmpl").Funcs(funcMap())
	t, err := t.Parse(d.tmplString)
	if err != nil {
		return "", fmt.Errorf("error in your template: %s", err.Error())
	}

	buf := new(bytes.Buffer)
	if err := t.Execute(buf, d.data); err != nil {
		return "", fmt.Errorf("error while filling your template: %s", err.Error())
	}

	return buf.String(), nil
}

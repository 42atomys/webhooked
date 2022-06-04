package formating

import (
	"bytes"
	"net/http"
	"text/template"

	"github.com/rs/zerolog/log"

	"atomys.codes/webhooked/internal/config"
)

type TemplateData struct {
	tmplString string
	data       map[string]interface{}
}

func NewTemplateData(tmplString string) *TemplateData {
	return &TemplateData{
		tmplString: tmplString,
		data:       make(map[string]interface{}),
	}
}

func (d *TemplateData) WithData(name string, data interface{}) *TemplateData {
	d.data[name] = data
	return d
}

func (d *TemplateData) WithRequest(r *http.Request) *TemplateData {
	d.WithData("Request", r)
	return d
}

func (d *TemplateData) WithPayload(payload []byte) *TemplateData {
	d.WithData("Payload", string(payload))
	return d
}

func (d *TemplateData) WithSpec(spec *config.WebhookSpec) *TemplateData {
	d.WithData("Spec", spec)
	return d
}

func (d *TemplateData) WithStorage(spec *config.StorageSpec) *TemplateData {
	d.WithData("Storage", spec)
	return d
}

func (d *TemplateData) WithConfig() *TemplateData {
	d.WithData("Config", config.Current())
	return d
}

func (d *TemplateData) Render() (string, error) {
	log.Debug().Msgf("rendering template: %s", d.tmplString)

	t := template.New("formatingTmpl").Funcs(funcMap())
	t, err := t.Parse(d.tmplString)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	if err := t.Execute(buf, d.data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

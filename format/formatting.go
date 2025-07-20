package format

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"sync"
	"text/template"

	"github.com/42atomys/webhooked/internal/contextutil"
	"github.com/go-sprout/sprout"
	"github.com/go-sprout/sprout/group/all"
	"github.com/valyala/fasthttp"
)

type Specs struct {
	TemplateString string `json:"templateString"`
	TemplatePath   string `json:"templatePath"`
}

type Formatting struct {
	specs      Specs
	template   *template.Template
	handler    sprout.Handler
	bufferPool sync.Pool
}

type TemplateFormatter interface {
	HasTemplate() bool
	HasTemplateCompiled() bool
	WithTemplate(template []byte) *Formatting
	Format(ctx context.Context, data map[string]any) ([]byte, error)
}

var (
	// ErrNoTemplate is returned when no template is defined in the Formatter
	// instance. Provide a template using the WithTemplate method.
	ErrNoTemplate = errors.New("no template defined")
)

func (f *Formatting) compileTemplate(specs Specs) error {
	var buffer bytes.Buffer

	if specs.TemplateString != "" {
		f.specs.TemplateString = specs.TemplateString
		buffer.WriteString(specs.TemplateString)
	}

	if specs.TemplatePath != "" {
		f.specs.TemplatePath = specs.TemplatePath
		file, err := os.OpenFile(specs.TemplatePath, os.O_RDONLY, 0666)
		if err != nil {
			return err
		}
		defer file.Close()

		var buffer bytes.Buffer
		_, err = io.Copy(&buffer, file)
		if err != nil {
			return err
		}
	}

	t, err := template.New("template").Funcs(f.handler.Build()).Parse(buffer.String())
	if err != nil {
		return fmt.Errorf("error while parsing your template: %s", err.Error())
	}

	f.template = t
	return nil
}

func New(specs Specs) (*Formatting, error) {
	f := &Formatting{
		handler: sprout.New(sprout.WithGroups(all.RegistryGroup())),
		bufferPool: sync.Pool{
			New: func() any {
				return new(bytes.Buffer)
			},
		},
	}
	if err := f.compileTemplate(specs); err != nil {
		return nil, err
	}

	return f, nil
}

func (f *Formatting) HasTemplate() bool {
	if f == nil {
		return false
	}

	return f.specs.TemplateString != "" || f.specs.TemplatePath != ""
}

func (f *Formatting) HasTemplateCompiled() bool {
	if f == nil {
		return false
	}

	return f.template != nil
}

func (f *Formatting) WithTemplate(template []byte) *Formatting {
	if f == nil {
		return nil
	}

	f.specs.TemplateString = string(template)
	return f
}

func (f *Formatting) Format(ctx context.Context, data map[string]any) ([]byte, error) {
	if f.template == nil {
		return nil, ErrNoTemplate
	}

	buf := f.bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer f.bufferPool.Put(buf)

	// Insert context data into the template data
	maps.Copy(data, templateData(ctx))

	if err := f.template.Execute(buf, data); err != nil {
		return nil, fmt.Errorf("error while filling your template: %s", err.Error())
	}

	return buf.Bytes(), nil
}

func templateData(ctx context.Context) map[string]any {
	rctx, ok := contextutil.RequestCtxFromContext[*fasthttp.RequestCtx](ctx)
	if !ok {
		return map[string]any{}
	}

	return map[string]any{
		"ConnID":      rctx.ConnID(),
		"ConnTime":    rctx.ConnTime(),
		"Host":        string(rctx.Host()),
		"IsTLS":       rctx.IsTLS(),
		"Method":      string(rctx.Method()),
		"QueryArgs":   rctx.QueryArgs(),
		"RemoteAddr":  rctx.RemoteAddr(),
		"RemoteIP":    rctx.RemoteIP(),
		"RequestTime": rctx.Time(),
		"URI":         rctx.URI(),
		"UserAgent":   string(rctx.UserAgent()),
		"Request":     &rctx.Request,
		"Payload":     string(rctx.Request.Body()),
	}
}

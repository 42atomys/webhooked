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

type TemplateContexter interface {
	TemplateContext() map[string]any
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
			return fmt.Errorf("error reading template file: %w", err)
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
		return nil, fmt.Errorf("error compiling template: %w", err)
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
	maps.Copy(data, compileContexts(ctx, data))

	if err := f.template.Execute(buf, data); err != nil {
		return nil, fmt.Errorf("error while filling your template: %s", err.Error())
	}

	return buf.Bytes(), nil
}

func compileContexts(ctx context.Context, extras ...map[string]any) map[string]any {
	specTemplateCtx, ok := contextutil.WebhookSpecFromContext[TemplateContexter](ctx)
	if !ok {
		specTemplateCtx = nil
	}

	storageTemplateCtx, ok := contextutil.StoreFromContext[TemplateContexter](ctx)
	if !ok {
		storageTemplateCtx = nil
	}

	requestTemplateCtx, ok := contextutil.RequestCtxFromContext[TemplateContexter](ctx)
	if !ok {
		requestTemplateCtx = nil
	}

	merged := MergeTemplateContexts(specTemplateCtx, storageTemplateCtx, requestTemplateCtx)

	for _, extra := range extras {
		for k, v := range extra {
			merged[k] = v
		}
	}
	return merged
}

func MergeTemplateContexts(ctxs ...TemplateContexter) map[string]any {
	merged := make(map[string]any)
	for _, ctx := range ctxs {
		if ctx == nil {
			continue
		}

		for k, v := range ctx.TemplateContext() {
			merged[k] = v
		}
	}
	return merged
}

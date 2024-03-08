package formatting

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
	"text/template"
)

type Formatter struct {
	tmplString string

	mu   sync.RWMutex // protect following field amd template parsing
	data map[string]interface{}
}

var (
	formatterCtxKey = struct{}{}
	// ErrNotFoundInContext is returned when the formatting data is not found in
	// the context. Use `FromContext` and `ToContext` to set and get the data in
	// the context.
	ErrNotFoundInContext = fmt.Errorf("unable to get the formatting data from the context")
	// ErrNoTemplate is returned when no template is defined in the Formatter
	// instance. Provide a template using the WithTemplate method.
	ErrNoTemplate = fmt.Errorf("no template defined")
)

// NewWithTemplate returns a new Formatter instance. It takes the template
// string as a parameter. The template string is the string that will be used
// to render the template. The data is the map of data that will be used to
// render the template.
// ! DEPRECATED: use New() and WithTemplate() instead
func NewWithTemplate(tmplString string) *Formatter {
	return &Formatter{
		tmplString: tmplString,
		data:       make(map[string]interface{}),
		mu:         sync.RWMutex{},
	}
}

// New returns a new Formatter instance. It takes no parameters. The template
// string must be set using the WithTemplate method. The data is the map of data
// that will be used to render the template.
func New() *Formatter {
	return &Formatter{
		data: make(map[string]interface{}),
		mu:   sync.RWMutex{},
	}
}

// WithTemplate sets the template string. The template string is the string that
// will be used to render the template.
func (d *Formatter) WithTemplate(tmplString string) *Formatter {
	d.tmplString = tmplString
	return d
}

// WithData adds a key-value pair to the data map. The key is the name of the
// variable and the value is the value of the variable.
func (d *Formatter) WithData(name string, data interface{}) *Formatter {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.data[name] = data
	return d
}

// WithRequest adds a http.Request object to the data map. The key of request is
// "Request".
func (d *Formatter) WithRequest(r *http.Request) *Formatter {
	d.WithData("Request", r)
	return d
}

// WithPayload adds a payload to the data map. The key of payload is "Payload".
// The payload is basically the body of the request.
func (d *Formatter) WithPayload(payload []byte) *Formatter {
	d.WithData("Payload", string(payload))
	return d
}

// Render returns the rendered template string. It takes the template string
// from the Formatter instance and the data stored in the Formatter
// instance. It returns an error if the template string is invalid or when
// rendering the template fails.
func (d *Formatter) Render() (string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.tmplString == "" {
		return "", ErrNoTemplate
	}

	t := template.New("formattingTmpl").Funcs(funcMap())
	t, err := t.Parse(d.tmplString)
	if err != nil {
		return "", fmt.Errorf("error in your template: %s", err.Error())
	}

	buf := new(bytes.Buffer)
	if err := t.Execute(buf, d.data); err != nil {
		return "", fmt.Errorf("error while filling your template: %s", err.Error())
	}

	if buf.String() == "<no value>" {
		return "", fmt.Errorf("template cannot be rendered, check your template")
	}

	return buf.String(), nil
}

// FromContext returns the Formatter instance stored in the context. It returns
// an error if the Formatter instance is not found in the context.
func FromContext(ctx context.Context) (*Formatter, error) {
	d, ok := ctx.Value(formatterCtxKey).(*Formatter)
	if !ok {
		return nil, ErrNotFoundInContext
	}
	return d, nil
}

// ToContext adds the Formatter instance to the context. It returns the context
// with the Formatter instance.
func ToContext(ctx context.Context, d *Formatter) context.Context {
	return context.WithValue(ctx, formatterCtxKey, d)
}

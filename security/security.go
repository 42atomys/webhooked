package security

import "github.com/valyala/fasthttp"

type Security struct {
	Type  string `json:"type"`
	Specs Specs  `json:"specs"`
}

type Specs interface {
	EnsureConfigurationCompleteness() error
	Initialize() error
	IsSecure(ctx *fasthttp.RequestCtx) (bool, error)
}

func (s *Security) IsSecure(ctx *fasthttp.RequestCtx) (bool, error) {
	return s.Specs.IsSecure(ctx)
}

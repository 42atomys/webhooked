package github

import (
	"bytes"
	"errors"

	"github.com/42atomys/webhooked/internal/valuable"
	"github.com/valyala/fasthttp"
)

type GitHubSecuritySpec struct {
	Secret *valuable.Valuable `json:"secret"`
}

const headerName = "X-Hub-Signature-256"

func (s *GitHubSecuritySpec) EnsureConfigurationCompleteness() error {
	return nil
}

func (s *GitHubSecuritySpec) Initialize() error {
	return nil
}

func (s *GitHubSecuritySpec) IsSecure(ctx *fasthttp.RequestCtx) (bool, error) {
	if s.Secret == nil || s.Secret.First() == "" {
		return false, errors.New("secret is required")
	}

	headerValue := ctx.Request.Header.Peek(headerName)
	if len(headerValue) == 0 {
		return false, nil
	}

	return bytes.Equal([]byte(s.Secret.First()), headerValue), nil
}

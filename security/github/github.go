package github

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/42atomys/webhooked/internal/fasthttpz"
	"github.com/42atomys/webhooked/internal/valuable"
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

func (s *GitHubSecuritySpec) IsSecure(ctx context.Context, rctx *fasthttpz.RequestCtx) (bool, error) {
	if s.Secret == nil || s.Secret.First() == "" {
		return false, errors.New("secret is required")
	}

	headerValue := rctx.Request.Header.Peek(headerName)
	if len(headerValue) == 0 {
		return false, nil
	}

	h := hmac.New(sha256.New, []byte(s.Secret.String()))
	h.Write(rctx.PostBody())
	expectedValue := "sha256=" + hex.EncodeToString(h.Sum(nil))

	return bytes.Equal([]byte(expectedValue), headerValue), nil
}

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWebhookSpec_HasSecurity(t *testing.T) {
	assert.False(t, WebhookSpec{Security: nil}.HasSecurity())
	// TODO: add tests for security
}

func TestWebhookSpec_HasGlobalFormatting(t *testing.T) {
	assert.False(t, WebhookSpec{Formatting: nil}.HasGlobalFormatting())
	assert.False(t, WebhookSpec{Formatting: &FormattingSpec{}}.HasGlobalFormatting())
	assert.False(t, WebhookSpec{Formatting: &FormattingSpec{TemplatePath: ""}}.HasGlobalFormatting())
	assert.False(t, WebhookSpec{Formatting: &FormattingSpec{TemplateString: ""}}.HasGlobalFormatting())
	assert.False(t, WebhookSpec{Formatting: &FormattingSpec{TemplatePath: "", TemplateString: ""}}.HasGlobalFormatting())
	assert.True(t, WebhookSpec{Formatting: &FormattingSpec{TemplatePath: "/_tmp/invalid_path", TemplateString: ""}}.HasGlobalFormatting())
	assert.True(t, WebhookSpec{Formatting: &FormattingSpec{TemplatePath: "/_tmp/invalid_path", TemplateString: "{{}}"}}.HasGlobalFormatting())
}

func TestWebhookSpec_HasFormatting(t *testing.T) {
	assert.False(t, StorageSpec{Formatting: nil}.HasFormatting())
	assert.False(t, StorageSpec{Formatting: &FormattingSpec{}}.HasFormatting())
	assert.False(t, StorageSpec{Formatting: &FormattingSpec{TemplatePath: ""}}.HasFormatting())
	assert.False(t, StorageSpec{Formatting: &FormattingSpec{TemplateString: ""}}.HasFormatting())
	assert.False(t, StorageSpec{Formatting: &FormattingSpec{TemplatePath: "", TemplateString: ""}}.HasFormatting())
	assert.True(t, StorageSpec{Formatting: &FormattingSpec{TemplatePath: "/_tmp/invalid_path", TemplateString: ""}}.HasFormatting())
	assert.True(t, StorageSpec{Formatting: &FormattingSpec{TemplatePath: "/_tmp/invalid_path", TemplateString: "{{}}"}}.HasFormatting())
}

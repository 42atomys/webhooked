package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWebhookSpec_HasSecurity(t *testing.T) {
	assert.False(t, WebhookSpec{Security: nil}.HasSecurity())
	// assert.False(t, WebhookSpec{Security: []map[string]map[string]interface{}{}}.HasSecurity())
	// assert.True(t, WebhookSpec{SecurityFactories: make([]*factory.Factory, 1)}.HasSecurity())
	// assert.True(t, WebhookSpec{SecurityFactories: make([]*factory.Factory, 2)}.HasSecurity())
}

func TestWebhookSpec_HasGlobalFormating(t *testing.T) {
	assert.False(t, WebhookSpec{Formating: nil}.HasGlobalFormating())
	assert.False(t, WebhookSpec{Formating: &FormatingSpec{}}.HasGlobalFormating())
	assert.False(t, WebhookSpec{Formating: &FormatingSpec{TemplatePath: ""}}.HasGlobalFormating())
	assert.False(t, WebhookSpec{Formating: &FormatingSpec{TemplateString: ""}}.HasGlobalFormating())
	assert.False(t, WebhookSpec{Formating: &FormatingSpec{TemplatePath: "", TemplateString: ""}}.HasGlobalFormating())
	assert.True(t, WebhookSpec{Formating: &FormatingSpec{TemplatePath: "/_tmp/invalid_path", TemplateString: ""}}.HasGlobalFormating())
	assert.True(t, WebhookSpec{Formating: &FormatingSpec{TemplatePath: "/_tmp/invalid_path", TemplateString: "{{}}"}}.HasGlobalFormating())
}

func TestWebhookSpec_HasFormating(t *testing.T) {
	assert.False(t, StorageSpec{Formating: nil}.HasFormating())
	assert.False(t, StorageSpec{Formating: &FormatingSpec{}}.HasFormating())
	assert.False(t, StorageSpec{Formating: &FormatingSpec{TemplatePath: ""}}.HasFormating())
	assert.False(t, StorageSpec{Formating: &FormatingSpec{TemplateString: ""}}.HasFormating())
	assert.False(t, StorageSpec{Formating: &FormatingSpec{TemplatePath: "", TemplateString: ""}}.HasFormating())
	assert.True(t, StorageSpec{Formating: &FormatingSpec{TemplatePath: "/_tmp/invalid_path", TemplateString: ""}}.HasFormating())
	assert.True(t, StorageSpec{Formating: &FormatingSpec{TemplatePath: "/_tmp/invalid_path", TemplateString: "{{}}"}}.HasFormating())
}

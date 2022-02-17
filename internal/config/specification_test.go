package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWebhookSpec_HasSecurity(t *testing.T) {
	assert.False(t, WebhookSpec{Security: nil}.HasSecurity())
	assert.False(t, WebhookSpec{Security: map[string]SecuritySpec{}}.HasSecurity())
	assert.True(t, WebhookSpec{Security: map[string]SecuritySpec{"test": {}}}.HasSecurity())
	assert.True(t, WebhookSpec{Security: map[string]SecuritySpec{"foo": {}, "bar": {}}}.HasSecurity())
}

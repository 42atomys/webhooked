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

package config

// HasSecurity returns true if the spec has a security layer
func (s WebhookSpec) HasSecurity() bool {
	return s.Security != nil && len(s.Security) > 0
}

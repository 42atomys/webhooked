package config

// HasSecurity returns true if the spec has a security factories
func (s WebhookSpec) HasSecurity() bool {
	return s.SecurityPipeline != nil && s.SecurityPipeline.HasFactories()
}

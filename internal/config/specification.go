package config

// HasSecurity returns true if the spec has a security factories
func (s WebhookSpec) HasSecurity() bool {
	return s.SecurityPipeline != nil && s.SecurityPipeline.HasFactories()
}

// HasGlobalFormatting returns true if the spec has a global formatting
func (s WebhookSpec) HasGlobalFormatting() bool {
	return s.Formatting != nil && (s.Formatting.TemplatePath != "" || s.Formatting.TemplateString != "")
}

// HasFormatting returns true if the storage spec has a formatting
func (s StorageSpec) HasFormatting() bool {
	return s.Formatting != nil && (s.Formatting.TemplatePath != "" || s.Formatting.TemplateString != "")
}

package config

// HasSecurity returns true if the spec has a security factories
func (s WebhookSpec) HasSecurity() bool {
	return s.SecurityPipeline != nil && s.SecurityPipeline.HasFactories()
}

// HasGlobalFormating returns true if the spec has a global formating
func (s WebhookSpec) HasGlobalFormating() bool {
	return s.Formating.TemplatePath != "" || s.Formating.TemplateString != ""
}

// HasFormating returns true if the storage spec has a formating
func (s StorageSpec) HasFormating() bool {
	return s.Formating != nil && (s.Formating.TemplatePath != "" || s.Formating.TemplateString != "")
}

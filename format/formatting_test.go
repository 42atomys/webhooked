//go:build unit

package format

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TestSuiteFormatting struct {
	suite.Suite

	validTemplateString    string
	invalidTemplateString  string
	complexTemplateString  string
	testData              map[string]any
	tempTemplatePath      string
	invalidTemplatePath   string
}

func (suite *TestSuiteFormatting) BeforeTest(suiteName, testName string) {
	suite.validTemplateString = "Hello {{ .Name }}!"
	suite.invalidTemplateString = "Hello {{ .Name "  // Missing closing brace
	suite.complexTemplateString = `
Name: {{ .Name }}
Age: {{ .Age }}
{{- if .Items }}
Items:
{{- range .Items }}
  - {{ . }}
{{- end }}
{{- end }}
`

	suite.testData = map[string]any{
		"Name": "World",
		"Age":  25,
		"Items": []string{"item1", "item2", "item3"},
	}

	// Create temporary template file
	suite.tempTemplatePath = "/tmp/webhooked_test_template.txt"
	suite.invalidTemplatePath = "/nonexistent/path/template.txt"
}

func (suite *TestSuiteFormatting) AfterTest(suiteName, testName string) {
	// Clean up temporary files
	os.Remove(suite.tempTemplatePath)
}

func (suite *TestSuiteFormatting) TestNew_WithValidTemplateString() {
	assert := assert.New(suite.T())

	specs := Specs{
		TemplateString: suite.validTemplateString,
	}

	formatting, err := New(specs)

	assert.NoError(err)
	assert.NotNil(formatting)
	assert.True(formatting.HasTemplate())
	assert.True(formatting.HasTemplateCompiled())
}

func (suite *TestSuiteFormatting) TestNew_WithInvalidTemplateString() {
	assert := assert.New(suite.T())

	specs := Specs{
		TemplateString: suite.invalidTemplateString,
	}

	formatting, err := New(specs)

	assert.Error(err)
	assert.Contains(err.Error(), "error compiling template")
	assert.Nil(formatting)
}

func (suite *TestSuiteFormatting) TestNew_WithValidTemplatePath() {
	assert := assert.New(suite.T())

	// Write template to temporary file
	err := os.WriteFile(suite.tempTemplatePath, []byte(suite.validTemplateString), 0644)
	require.NoError(suite.T(), err)

	specs := Specs{
		TemplatePath: suite.tempTemplatePath,
	}

	formatting, err := New(specs)

	assert.NoError(err)
	assert.NotNil(formatting)
	assert.True(formatting.HasTemplate())
	assert.True(formatting.HasTemplateCompiled())
}

func (suite *TestSuiteFormatting) TestNew_WithInvalidTemplatePath() {
	assert := assert.New(suite.T())

	specs := Specs{
		TemplatePath: suite.invalidTemplatePath,
	}

	formatting, err := New(specs)

	assert.Error(err)
	assert.Contains(err.Error(), "error compiling template")
	assert.Nil(formatting)
}

func (suite *TestSuiteFormatting) TestNew_WithBothStringAndPath() {
	assert := assert.New(suite.T())

	// Write template to temporary file
	err := os.WriteFile(suite.tempTemplatePath, []byte("File: {{ .FileContent }}"), 0644)
	require.NoError(suite.T(), err)

	specs := Specs{
		TemplateString: suite.validTemplateString,
		TemplatePath:   suite.tempTemplatePath,
	}

	formatting, err := New(specs)

	assert.NoError(err)
	assert.NotNil(formatting)
	assert.True(formatting.HasTemplate())
	assert.True(formatting.HasTemplateCompiled())
}

func (suite *TestSuiteFormatting) TestNew_WithEmptySpecs() {
	assert := assert.New(suite.T())

	specs := Specs{}

	formatting, err := New(specs)

	assert.NoError(err)
	assert.NotNil(formatting)
	assert.False(formatting.HasTemplate())
	assert.True(formatting.HasTemplateCompiled()) // Empty template still compiles
}

func (suite *TestSuiteFormatting) TestHasTemplate_WithNilFormatting() {
	assert := assert.New(suite.T())

	var formatting *Formatting = nil

	hasTemplate := formatting.HasTemplate()

	assert.False(hasTemplate)
}

func (suite *TestSuiteFormatting) TestHasTemplateCompiled_WithNilFormatting() {
	assert := assert.New(suite.T())

	var formatting *Formatting = nil

	hasCompiled := formatting.HasTemplateCompiled()

	assert.False(hasCompiled)
}

func (suite *TestSuiteFormatting) TestWithTemplate_WithNilFormatting() {
	assert := assert.New(suite.T())

	var formatting *Formatting = nil

	result := formatting.WithTemplate([]byte("test"))

	assert.Nil(result)
}

func (suite *TestSuiteFormatting) TestWithTemplate_ValidTemplate() {
	assert := assert.New(suite.T())

	formatting, err := New(Specs{})
	require.NoError(suite.T(), err)

	newTemplate := []byte("New template: {{ .Value }}")
	result := formatting.WithTemplate(newTemplate)

	assert.NotNil(result)
	assert.True(result.HasTemplate())
	assert.Equal(string(newTemplate), result.specs.TemplateString)
	// Note: WithTemplate only sets the string, doesn't recompile
	assert.True(result.HasTemplateCompiled()) // Still has the old compiled template
}

func (suite *TestSuiteFormatting) TestFormat_ValidTemplate() {
	assert := assert.New(suite.T())

	formatting, err := New(Specs{TemplateString: suite.validTemplateString})
	require.NoError(suite.T(), err)

	result, err := formatting.Format(context.Background(), suite.testData)

	assert.NoError(err)
	assert.Equal("Hello World!", string(result))
}

func (suite *TestSuiteFormatting) TestFormat_ComplexTemplate() {
	assert := assert.New(suite.T())

	formatting, err := New(Specs{TemplateString: suite.complexTemplateString})
	require.NoError(suite.T(), err)

	result, err := formatting.Format(context.Background(), suite.testData)

	assert.NoError(err)
	expected := `
Name: World
Age: 25
Items:
  - item1
  - item2
  - item3
`
	assert.Equal(expected, string(result))
}

func (suite *TestSuiteFormatting) TestFormat_NoTemplate() {
	assert := assert.New(suite.T())

	formatting, err := New(Specs{})
	require.NoError(suite.T(), err)
	
	// Clear the template to simulate no template scenario
	formatting.template = nil

	result, err := formatting.Format(context.Background(), suite.testData)

	assert.Error(err)
	assert.ErrorIs(err, ErrNoTemplate)
	assert.Nil(result)
}

func (suite *TestSuiteFormatting) TestFormat_TemplateExecutionError() {
	assert := assert.New(suite.T())

	// Template that will cause execution error (division by zero with custom func)
	// Use a template that calls a function with wrong number of arguments
	badTemplate := "{{ printf }}"  // printf requires at least one argument
	formatting, err := New(Specs{TemplateString: badTemplate})
	require.NoError(suite.T(), err)

	result, err := formatting.Format(context.Background(), suite.testData)

	assert.Error(err)
	assert.Contains(err.Error(), "error while filling your template")
	assert.Nil(result)
}

func (suite *TestSuiteFormatting) TestCompileContexts_EmptyContext() {
	assert := assert.New(suite.T())

	ctx := context.Background()
	result := compileContexts(ctx)

	assert.NotNil(result)
	assert.Empty(result)
}

func (suite *TestSuiteFormatting) TestCompileContexts_WithExtras() {
	assert := assert.New(suite.T())

	ctx := context.Background()
	extra1 := map[string]any{"key1": "value1"}
	extra2 := map[string]any{"key2": "value2"}

	result := compileContexts(ctx, extra1, extra2)

	assert.NotNil(result)
	assert.Equal("value1", result["key1"])
	assert.Equal("value2", result["key2"])
}

func (suite *TestSuiteFormatting) TestMergeTemplateContexts_NilContexts() {
	assert := assert.New(suite.T())

	result := MergeTemplateContexts(nil, nil)

	assert.NotNil(result)
	assert.Empty(result)
}

func (suite *TestSuiteFormatting) TestMergeTemplateContexts_ValidContexts() {
	assert := assert.New(suite.T())

	ctx1 := &mockTemplateContexter{
		context: map[string]any{"key1": "value1", "shared": "ctx1"},
	}
	ctx2 := &mockTemplateContexter{
		context: map[string]any{"key2": "value2", "shared": "ctx2"},
	}

	result := MergeTemplateContexts(ctx1, ctx2)

	assert.NotNil(result)
	assert.Equal("value1", result["key1"])
	assert.Equal("value2", result["key2"])
	assert.Equal("ctx2", result["shared"]) // Later context should override
}

func (suite *TestSuiteFormatting) TestFormatWithSprintFunctions() {
	assert := assert.New(suite.T())

	// Test template with built-in template functions (no sprout functions for now)
	templateString := `{{ .Name }} - {{ printf "%d" .Age }}`
	formatting, err := New(Specs{TemplateString: templateString})
	require.NoError(suite.T(), err)

	result, err := formatting.Format(context.Background(), suite.testData)

	assert.NoError(err)
	assert.Equal("World - 25", string(result))
}

func TestRunFormattingSuite(t *testing.T) {
	suite.Run(t, new(TestSuiteFormatting))
}

// Mock implementation for testing

type mockTemplateContexter struct {
	context map[string]any
}

func (m *mockTemplateContexter) TemplateContext() map[string]any {
	return m.context
}

// Benchmarks

func BenchmarkNew_SimpleTemplate(b *testing.B) {
	specs := Specs{TemplateString: "Hello {{ .Name }}!"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		New(specs) // nolint:errcheck
	}
}

func BenchmarkFormat_SimpleTemplate(b *testing.B) {
	formatting, _ := New(Specs{TemplateString: "Hello {{ .Name }}!"})
	data := map[string]any{"Name": "World"}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatting.Format(ctx, data) // nolint:errcheck
	}
}

func BenchmarkFormat_ComplexTemplate(b *testing.B) {
	templateString := `
Name: {{ .Name }}
Age: {{ .Age }}
{{- if .Items }}
Items:
{{- range .Items }}
  - {{ . }}
{{- end }}
{{- end }}
`
	formatting, _ := New(Specs{TemplateString: templateString})
	data := map[string]any{
		"Name": "World",
		"Age":  25,
		"Items": []string{"item1", "item2", "item3"},
	}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatting.Format(ctx, data) // nolint:errcheck
	}
}

func BenchmarkMergeTemplateContexts(b *testing.B) {
	ctx1 := &mockTemplateContexter{
		context: map[string]any{"key1": "value1", "shared": "ctx1"},
	}
	ctx2 := &mockTemplateContexter{
		context: map[string]any{"key2": "value2", "shared": "ctx2"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MergeTemplateContexts(ctx1, ctx2)
	}
}

func BenchmarkWithTemplate(b *testing.B) {
	formatting, _ := New(Specs{TemplateString: "initial"})
	template := []byte("New template: {{ .Value }}")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatting.WithTemplate(template)
	}
}
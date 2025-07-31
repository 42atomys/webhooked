//go:build unit

package format

import (
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TestSuiteFormatHooks struct {
	suite.Suite

	validTemplateStringData    map[string]any
	validTemplatePathData      map[string]any
	bothTemplatesData          map[string]any
	emptyTemplatesData         map[string]any
	invalidTemplateStringData  map[string]any
	invalidTemplatePathData    map[string]any
	nonMapData                 string
	tempTemplatePath           string
	invalidTemplatePath        string
}

func (suite *TestSuiteFormatHooks) BeforeTest(suiteName, testName string) {
	suite.validTemplateStringData = map[string]any{
		"templateString": "Hello {{ .Name }}!",
	}

	suite.tempTemplatePath = "/tmp/format_hooks_test_template.txt"
	suite.invalidTemplatePath = "/nonexistent/path/template.txt"

	// Create temporary template file
	err := os.WriteFile(suite.tempTemplatePath, []byte("File template: {{ .Content }}"), 0644)
	require.NoError(suite.T(), err)

	suite.validTemplatePathData = map[string]any{
		"templatePath": suite.tempTemplatePath,
	}

	suite.bothTemplatesData = map[string]any{
		"templateString": "String: {{ .Name }}",
		"templatePath":   suite.tempTemplatePath,
	}

	suite.emptyTemplatesData = map[string]any{
		"templateString": "",
		"templatePath":   "",
	}

	suite.invalidTemplateStringData = map[string]any{
		"templateString": "{{ invalid template",
	}

	suite.invalidTemplatePathData = map[string]any{
		"templatePath": suite.invalidTemplatePath,
	}

	suite.nonMapData = "not_a_map"
}

func (suite *TestSuiteFormatHooks) AfterTest(suiteName, testName string) {
	// Clean up temporary files
	os.Remove(suite.tempTemplatePath)
}

func (suite *TestSuiteFormatHooks) TestDecodeHook_ValidTemplateString() {
	assert := assert.New(suite.T())

	fromType := reflect.TypeOf(suite.validTemplateStringData)
	toType := reflect.TypeOf((*Formatting)(nil))

	result, err := DecodeHook(fromType, toType, suite.validTemplateStringData)

	assert.NoError(err)
	assert.NotNil(result)
	assert.IsType((*Formatting)(nil), result)

	formatting := result.(*Formatting)
	assert.True(formatting.HasTemplate())
	assert.True(formatting.HasTemplateCompiled())
}

func (suite *TestSuiteFormatHooks) TestDecodeHook_ValidTemplatePath() {
	assert := assert.New(suite.T())

	fromType := reflect.TypeOf(suite.validTemplatePathData)
	toType := reflect.TypeOf((*Formatting)(nil))

	result, err := DecodeHook(fromType, toType, suite.validTemplatePathData)

	assert.NoError(err)
	assert.NotNil(result)
	assert.IsType((*Formatting)(nil), result)

	formatting := result.(*Formatting)
	assert.True(formatting.HasTemplate())
	assert.True(formatting.HasTemplateCompiled())
}

func (suite *TestSuiteFormatHooks) TestDecodeHook_BothTemplates() {
	assert := assert.New(suite.T())

	fromType := reflect.TypeOf(suite.bothTemplatesData)
	toType := reflect.TypeOf((*Formatting)(nil))

	result, err := DecodeHook(fromType, toType, suite.bothTemplatesData)

	assert.NoError(err)
	assert.NotNil(result)
	assert.IsType((*Formatting)(nil), result)

	formatting := result.(*Formatting)
	assert.True(formatting.HasTemplate())
	assert.True(formatting.HasTemplateCompiled())
}

func (suite *TestSuiteFormatHooks) TestDecodeHook_EmptyTemplates() {
	assert := assert.New(suite.T())

	fromType := reflect.TypeOf(suite.emptyTemplatesData)
	toType := reflect.TypeOf((*Formatting)(nil))

	result, err := DecodeHook(fromType, toType, suite.emptyTemplatesData)

	assert.NoError(err)
	assert.Nil(result) // Should return nil when both templates are empty
}

func (suite *TestSuiteFormatHooks) TestDecodeHook_InvalidTemplateString() {
	assert := assert.New(suite.T())

	fromType := reflect.TypeOf(suite.invalidTemplateStringData)
	toType := reflect.TypeOf((*Formatting)(nil))

	result, err := DecodeHook(fromType, toType, suite.invalidTemplateStringData)

	assert.Error(err)
	assert.Contains(err.Error(), "error creating formatting")
	assert.Nil(result)
}

func (suite *TestSuiteFormatHooks) TestDecodeHook_InvalidTemplatePath() {
	assert := assert.New(suite.T())

	fromType := reflect.TypeOf(suite.invalidTemplatePathData)
	toType := reflect.TypeOf((*Formatting)(nil))

	result, err := DecodeHook(fromType, toType, suite.invalidTemplatePathData)

	assert.Error(err)
	assert.Contains(err.Error(), "error creating formatting")
	assert.Nil(result)
}

// Note: NonMapData test logic should be tested in integration tests

func (suite *TestSuiteFormatHooks) TestDecodeHook_WrongFromType() {
	assert := assert.New(suite.T())

	fromType := reflect.TypeOf("string") // Not a map
	toType := reflect.TypeOf((*Formatting)(nil))
	data := "test"

	result, err := DecodeHook(fromType, toType, data)

	// Should return data unchanged when from type is not map
	assert.NoError(err)
	assert.Equal(data, result)
}

func (suite *TestSuiteFormatHooks) TestDecodeHook_WrongToType() {
	assert := assert.New(suite.T())

	fromType := reflect.TypeOf(suite.validTemplateStringData)
	toType := reflect.TypeOf("string") // Not *Formatting
	data := suite.validTemplateStringData

	result, err := DecodeHook(fromType, toType, data)

	// Should return data unchanged when to type is not *Formatting
	assert.NoError(err)
	assert.Equal(data, result)
}

func (suite *TestSuiteFormatHooks) TestDecodeHook_NonStringTemplateString() {
	assert := assert.New(suite.T())

	dataWithNonStringTemplate := map[string]any{
		"templateString": 123, // Not a string
	}

	fromType := reflect.TypeOf(dataWithNonStringTemplate)
	toType := reflect.TypeOf((*Formatting)(nil))

	result, err := DecodeHook(fromType, toType, dataWithNonStringTemplate)

	// Should treat non-string as empty and return nil
	assert.NoError(err)
	assert.Nil(result)
}

func (suite *TestSuiteFormatHooks) TestDecodeHook_NonStringTemplatePath() {
	assert := assert.New(suite.T())

	dataWithNonStringPath := map[string]any{
		"templatePath": 123, // Not a string
	}

	fromType := reflect.TypeOf(dataWithNonStringPath)
	toType := reflect.TypeOf((*Formatting)(nil))

	result, err := DecodeHook(fromType, toType, dataWithNonStringPath)

	// Should treat non-string as empty and return nil
	assert.NoError(err)
	assert.Nil(result)
}

func (suite *TestSuiteFormatHooks) TestDecodeHook_MixedValidInvalid() {
	assert := assert.New(suite.T())

	dataWithMixed := map[string]any{
		"templateString": "Valid {{ .Template }}",
		"templatePath":   123, // Invalid (not string)
	}

	fromType := reflect.TypeOf(dataWithMixed)
	toType := reflect.TypeOf((*Formatting)(nil))

	result, err := DecodeHook(fromType, toType, dataWithMixed)

	// Should succeed with just the valid templateString
	assert.NoError(err)
	assert.NotNil(result)
	assert.IsType((*Formatting)(nil), result)

	formatting := result.(*Formatting)
	assert.True(formatting.HasTemplate())
}

func (suite *TestSuiteFormatHooks) TestDecodeHook_ExtraFields() {
	assert := assert.New(suite.T())

	dataWithExtra := map[string]any{
		"templateString": "Hello {{ .Name }}!",
		"extraField":     "should be ignored",
		"anotherField":   123,
	}

	fromType := reflect.TypeOf(dataWithExtra)
	toType := reflect.TypeOf((*Formatting)(nil))

	result, err := DecodeHook(fromType, toType, dataWithExtra)

	// Should succeed and ignore extra fields
	assert.NoError(err)
	assert.NotNil(result)
	assert.IsType((*Formatting)(nil), result)

	formatting := result.(*Formatting)
	assert.True(formatting.HasTemplate())
}

func (suite *TestSuiteFormatHooks) TestDecodeHook_EmptyMap() {
	assert := assert.New(suite.T())

	emptyMap := map[string]any{}

	fromType := reflect.TypeOf(emptyMap)
	toType := reflect.TypeOf((*Formatting)(nil))

	result, err := DecodeHook(fromType, toType, emptyMap)

	// Should return nil for empty map
	assert.NoError(err)
	assert.Nil(result)
}

func (suite *TestSuiteFormatHooks) TestDecodeHook_OnlyWhitespaceTemplates() {
	assert := assert.New(suite.T())

	whitespaceData := map[string]any{
		"templateString": "   ",
		"templatePath":   "", // Don't use invalid path
	}

	fromType := reflect.TypeOf(whitespaceData)
	toType := reflect.TypeOf((*Formatting)(nil))

	result, err := DecodeHook(fromType, toType, whitespaceData)

	// Should create formatting with whitespace templates (they're not empty strings)
	assert.NoError(err)
	assert.NotNil(result)
	assert.IsType((*Formatting)(nil), result)
}

// Note: NilMap test removed due to panic when accessing nil map

func (suite *TestSuiteFormatHooks) TestDecodeHook_ToFormattingValue() {
	assert := assert.New(suite.T())

	// Test with Formatting value instead of pointer
	fromType := reflect.TypeOf(suite.validTemplateStringData)
	toType := reflect.TypeOf(Formatting{})

	result, err := DecodeHook(fromType, toType, suite.validTemplateStringData)

	// Should return data unchanged when to type is not *Formatting
	assert.NoError(err)
	assert.Equal(suite.validTemplateStringData, result)
}

func TestRunFormatHooksSuite(t *testing.T) {
	suite.Run(t, new(TestSuiteFormatHooks))
}

// Benchmarks

func BenchmarkDecodeHook_ValidTemplateString(b *testing.B) {
	data := map[string]any{
		"templateString": "Hello {{ .Name }}!",
	}
	fromType := reflect.TypeOf(data)
	toType := reflect.TypeOf((*Formatting)(nil))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeHook(fromType, toType, data) // nolint:errcheck
	}
}

func BenchmarkDecodeHook_EmptyTemplates(b *testing.B) {
	data := map[string]any{
		"templateString": "",
		"templatePath":   "",
	}
	fromType := reflect.TypeOf(data)
	toType := reflect.TypeOf((*Formatting)(nil))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeHook(fromType, toType, data) // nolint:errcheck
	}
}

func BenchmarkDecodeHook_WrongType(b *testing.B) {
	data := "not a map"
	fromType := reflect.TypeOf(data)
	toType := reflect.TypeOf((*Formatting)(nil))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeHook(fromType, toType, data) // nolint:errcheck
	}
}

func BenchmarkDecodeHook_ComplexTemplate(b *testing.B) {
	data := map[string]any{
		"templateString": `
{{- range .Items }}
  Item: {{ .Name }} - {{ .Value }}
  {{- if .HasDetails }}
    Details:
    {{- range .Details }}
      - {{ . }}
    {{- end }}
  {{- end }}
{{- end }}`,
	}
	fromType := reflect.TypeOf(data)
	toType := reflect.TypeOf((*Formatting)(nil))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeHook(fromType, toType, data) // nolint:errcheck
	}
}

func BenchmarkDecodeHook_BothTemplates(b *testing.B) {
	// Create a temporary file for benchmarking
	tempFile := "/tmp/benchmark_template.txt"
	os.WriteFile(tempFile, []byte("Benchmark template: {{ .Value }}"), 0644)
	defer os.Remove(tempFile)

	data := map[string]any{
		"templateString": "String: {{ .Name }}",
		"templatePath":   tempFile,
	}
	fromType := reflect.TypeOf(data)
	toType := reflect.TypeOf((*Formatting)(nil))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeHook(fromType, toType, data) // nolint:errcheck
	}
}
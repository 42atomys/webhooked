---
description: >-
  Templates are powerful. Use them wisely to transform your webhook data exactly
  as needed.
---

# Formatting

## Introduction

Webhooked uses Go templates enhanced with 100+ functions from [go-sprout](https://github.com/go-sprout/sprout) to transform webhook data. This powerful templating system allows you to reshape, validate, and enrich data before storage or in responses.

{% hint style="success" %}
You can validate your template by using the command `webhooked --validate --config webhooked.yaml`
{% endhint %}

## Template Functions

Webhooked use internally [go-template](https://pkg.go.dev/text/template) for functionality and  [sprout](https://github.com/go-sprout/sprout) to provide 100+ functions.

Full documentation of sprout are available here : [Sprout](https://app.gitbook.com/o/IzEqF9iOTOuGvrXiqwxa/s/nfdPOfJ8fd1zSQvPrkEq/ "mention")&#x20;

## Template String

For simple template or a desire to have all configuration inside the yaml file, you can put your tempalte directly inline

```yaml
formatting:
  templateString: |
    {{- with $data := fromJSON .Payload -}}
    {{- $timestamp := $data.created_at | toDate "2006-01-02T15:04:05Z07:00" | unixEpoch -}}
    {
      "id": "{{ $data.id }}",
      "created": "{{ $timestamp }}",
      "reference": "{{ $data.id }}-{{ $timestamp }}"
    }
    {{- end -}}
```

## Template Files

For complex templates, usage of external files are recommended

```yaml
formatting:
  templatePath: /config/templates/webhook-transform.tmpl
```

Template file (`webhook-transform.tmpl`):

```go
{{- /* Complex webhook transformation template */ -}}
{{- $data := fromJSON .Payload -}}
{{- $config := dict 
  "environment" (env "ENVIRONMENT")
  "version" (env "VERSION")
  "region" (env "REGION")
-}}
{
  "metadata": {{ $config | toJSON }},
  "payload": {{ $data | toJSON }},
  "processed": "{{ now }}"
}
```

## Best Practices

1. **Always parse JSON once** and reuse the result
2. **Use `with` blocks** for nil-safety
3. **Provide defaults** for optional fields
4. **Use external files** for complex templates
5. **Cache computed values** in variables
6. **Document complex logic** with comments
7. **Test templates** with various payloads

## Performance Optimization

#### Parse Once, use multiple times

:white\_check\_mark: <mark style="color:$success;">**GOOD**</mark>: Parse one

```yaml
formatting:
  templateString: |
    {{ with $data := fromJSON .Payload }}
      ID: {{ $data.id }}
      Type: {{ $data.type }}
      User: {{ $data.user_id }}
    {{ end }}
```

:x: <mark style="color:red;">**BAD**</mark>: Parse multiple times

```yaml
formatting:
  templateString: |
    ID: {{ (fromJSON .Payload).id }}
    Type: {{ (fromJSON .Payload).type }}
    User: {{ (fromJSON .Payload).user_id }}
```

#### Avoid Unnecessary Operations

:white\_check\_mark: <mark style="color:$success;">**GOOD**</mark>: Cache computed values

```yaml
formatting:
  templateString: |
    {{ $timestamp := now }}
    {{ $id := uuidv4 }}
    {{ with $data := fromJSON .Payload }}
      {
        "id": "{{ $id }}",
        "created": "{{ $timestamp }}",
        "modified": "{{ $timestamp }}",
        "reference": "{{ $id }}-{{ $timestamp | unixEpoch }}"
      }
    {{ end }}
```

:x: <mark style="color:red;">**BAD**</mark>: Recall functions multiples times can cause bug and latency

```yaml
formatting:
  templateString: |
    {{ with $data := fromJSON .Payload }}
      {
        "id": "{{ uuidv4 }}",
        "created": "{{ now }}",
        "modified": "{{ now }}",
        "reference": "{{ uuidv4 }}-{{ now | unixEpoch }}"
      }
    {{ end }}
```

#### Avoid String building

:white\_check\_mark: <mark style="color:$success;">**GOOD**</mark>: Use template engine

```yaml
formatting:
  templateString: |
    {{ with $data := fromJSON .Payload }}
      {{ $data.type }}:{{ $data.id }}:{{ $data.timestamp }}
    {{ end }}
```

:x: <mark style="color:red;">**BAD**</mark>: use go functions to build end strings

```yaml
formatting:
  templateString: |
    {{ with $data := fromJSON .Payload }}
      {{ printf "%s:%s:%d" $data.type $data.id $data.timestamp }}
    {{ end }}
```

{% hint style="info" %}
This is provided as example, `printf` function aren't available on sprout project
{% endhint %}

## Common Issues and Solutions

### Issue: Template Syntax Errors

```yaml
# Error: unexpected "}" in operand
formatting:
  templateString: '{{ .Value }'  # Missing closing brace

# Fixed:
formatting:
  templateString: '{{ .Value }}'
```

### Issue: JSON Escaping

```yaml
# Problem: Invalid JSON output
formatting:
  templateString: '{"message": "{{ .Message }}"}'

# Solution: Use toJSON for proper escaping
formatting:
  templateString: '{"message": {{ .Message | toJSON }}}'
```

### Issue: Nil Pointer

```yaml
# Problem: Accessing non-existent field
formatting:
  templateString: '{{ $data.user.uame }}'

# Solution: Safe navigation using dig and default
formatting:
  templateString: |
    {{ $data | dig "user.uame" | default "Unknown" }}
```

---
description: >-
  Responses are what webhook providers see. Configure them carefully to ensure
  compatibility and proper acknowledgment.
---

# Response Layer

## Introduction

Configure how Webhooked responds to webhook requests, including status codes, headers, and response bodies.

## Response Context Variables

Variables available in security conditions:

| Variable             | Description                                            | Example                                   |
| -------------------- | ------------------------------------------------------ | ----------------------------------------- |
| `.SpecName`          | Name of the current spec as defined in config          | `"user-events"`                           |
| `.SpecEntrypointURL` | EntrypointURL of the current spec as defined in config | `"/user-events"`                          |
| `.ConnID`            | Unique connection ID                                   | `123549841`                               |
| `.ConnTime`          | Connection established time                            | time.Time object `2025-08-20T21:10:00Z`   |
| `.Host`              | Host header of request                                 | `"example.com"`                           |
| `.IsTLS`             | Whether request is HTTPS                               | `true`                                    |
| `.Method`            | HTTP method used                                       | `"POST"`                                  |
| `.Payload`           | Raw request body                                       | `{"data": "value"}`                       |
| `.QueryArgs`         | Query parameters object                                | `fasthttp.Args{"id":"123","token":"abc"}` |
| `.RemoteAddr`        | Remote Addr                                            | `"192.168.1.10:54321"`                    |
| `.RemoteIP`          | Remote network address                                 | `"192.168.1.1"`                           |
| `.RequestTime`       | Time when request was received                         | time.Time object `2025-08-20T21:10:00Z`   |
| `.Request`           | Full `fasthttp.Request` object                         | `&fasthttp.Request{...}`                  |
| `.URI`               | Request URI                                            | `"/webhooks/..."`                         |
| `.UserAgent`         | Client User-Agent header                               | `"Mozilla/5.0 (X11; Linux x86_64)"`       |

## Best Practices

1. **Always set appropriate status codes** (200 for success, 202 for async)
2. **Include request/correlation IDs** in responses
3. **Keep responses minimal** for performance
4. **Use consistent response formats** across webhooks
5. **Validate responses** match webhook provider expectations
6. **Set proper Content-Type** headers
7. **Include timestamps** for debugging
8. **Avoid sensitive data** in responses
9. **Test with actual webhook providers**

## Response Structure

```yaml
response:
  statusCode: 200                    # HTTP status code
  contentType: application/json      # Content-Type header
  formatting:                        # Response body template
    templateString: '{"status": "ok"}'
```

## Webhook-Specific Codes

```yaml
# GitHub webhooks expect 200-299
response:
  statusCode: 200

# Stripe webhooks expect exactly 200
response:
  statusCode: 200

# Slack webhooks expect 200 with specific body
response:
  statusCode: 200
  formatting:
    templateString: "OK"
```

## Examples

#### JSON Responses

```yaml
response:
  contentType: application/json
  formatting:
    templateString: |
      {
        "status": "received",
        "timestamp": "{{ now }}",
        "id": "{{ uuidv4 }}"
      }
```

#### Plain Text Responses

```yaml
response:
  contentType: text/plain
  formatting:
    templateString: "Webhook received and processed"
```

#### XML Responses

```yaml
response:
  contentType: application/xml
  formatting:
    templateString: |
      <?xml version="1.0" encoding="UTF-8"?>
      <response>
        <status>success</status>
        <id>{{ uuidv4 }}</id>
        <timestamp>{{ now }}</timestamp>
      </response>
```


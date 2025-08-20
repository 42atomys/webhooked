---
description: >-
  Security is not optional. Always authenticate webhooks in production
  environments.
---

# Security Layer

## Introduction

Security is paramount in webhook processing. Webhooked provides multiple authentication mechanisms to ensure only authorized sources can trigger your webhooks.

## Security Context Variables

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

## Security Providers

### GitHub

The GitHub provider validates webhook signatures using HMAC-SHA256.

```yaml
security:
  type: github
  specs:
    secret: # Valuable - see doc "Sourcing (Valuable)" for more info
      valueFrom:
        envRef: GITHUB_WEBHOOK_SECRET
```

#### How It Works

1. GitHub signs the payload with your secret using HMAC-SHA256
2. Sends signature in `X-Hub-Signature-256` header
3. Webhooked validates the signature matches
4. Rejects with 401 if validation fails

#### GitHub Webhook Setup

In your GitHub repository settings:

1. Go to Settings → Webhooks
2. Add webhook URL: `https://your-domain/webhooks/v1alpha2/your-path`
3. Set Content type: `application/json`
4. Set Secret: Your webhook secret
5. Select events to trigger webhook

#### Validation Example

```yaml
webhooks:
  - name: github-push
    entrypointUrl: /github/push
    security:
      type: github
      specs:
        secret:
          valueFrom:
            envRef: GITHUB_SECRET
    response:
      statusCode: 200
      formatting:
        templateString: |
          {
            "status": "validated",
            "event": "{{ .Request.Header.Peek "X-GitHub-Event" | toString }}",
            "delivery": "{{ .Request.Header.Peek "X-GitHub-Delivery" | toString }}"
          }
```

### Custom

The custom provider allows you to define authentication logic using Go templates.

#### Basic Token Authentication

```yaml
security:
  type: custom
  specs:
    condition: |
      {{ eq (.Request.Header.Peek "X-API-Key" | toString) (env "API_KEY") }}
```

#### Multiple Conditions

```yaml
security:
  type: custom
  specs:
    condition: |
      {{ and 
        (eq (.Request.Header.Peek "X-API-Key" | toString) (env "API_KEY"))
        (eq (.Request.Header.Peek "X-Tenant-ID" | toString) "tenant-123")
        (contains (.Request.Header.Peek "User-Agent" | toString) "MyApp")
      }}
```

### NoOp

The NoOp provider disables authentication entirely. **Use only in development!**

```yaml
security:
  type: noop
```

⚠️ **Warning**: Never use `noop` in production environments!

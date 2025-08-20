---
description: What is the Webhooked project ?
icon: address-card
---

# About

## What is Webhooked?

**Webhooked** is a high-performance, production-ready webhook gateway built in Go that acts as a universal receiver for incoming webhooks. It provides a robust pipeline for capturing, validating, transforming, storing, and responding to webhook events from any source.

Think of Webhooked as a smart proxy that sits between webhook providers (GitHub, Stripe, custom services) and your infrastructure (databases, message queues, applications), handling all the complexity of webhook processing while you focus on your business logic.

{% hint style="info" %}
_Webhooked is actively developed and we welcome contributions. Join us in building the future of webhook processing!_
{% endhint %}

## Why Webhooked?

### The Problem

Modern applications integrate with dozens of external services, each sending webhooks with different:

* Authentication methods
* Payload formats
* Retry mechanisms
* Rate limits
* Error handling requirements

Managing these individually leads to:

* **Scattered webhook logic** across multiple services
* **Inconsistent security** implementations
* **No unified monitoring** or observability
* **Complex retry logic** and error handling
* **Performance bottlenecks** in webhook processing

### The Solution

Webhooked provides a **centralized, declarative approach** to webhook management:

```yaml
# One configuration to rule them all
webhooks:
  - name: github-events
    entrypointUrl: /webhooks/github
    security:
      type: github
      specs:
        secret: ${{ env.GITHUB_SECRET }}
    storage:
      - type: redis
      - type: postgres
    response:
      statusCode: 200
```

## Key Features

### 🚀 Extreme Performance

* **50,000+ requests/second** on [commodity hardware](#user-content-fn-1)[^1]
* Built with `fasthttp` (10x faster than `net/http`)
* Kernel-level load balancing with `reuseport`
* Zero-allocation patterns for minimal GC pressure
* Concurrent storage operations

### 🔐 Enterprise-Grade Security

* **Multiple built-ins** authentication providers
* **Rate limiting** with token bucket algorithm
* **Request validation** and sanitization
* **Audit logging** for compliance

### 🔄 Powerful Data Transformation

* **Go template engine** with 100+ functions via [go-sprout](https://github.com/go-sprout/sprout)
* **JSON manipulation** and parsing
* **Custom formatting** per storage backend
* **Request/response transformation**

### 💾 Multi-Backend Storage

* **Parallel writes** to multiple backends
* **Support multiples backends**:
  * Message queuing
  * Persistent storage
  * More coming soon
* **Per-backend formatting** and transformation
* **Automatic retry** and error handling

### 📊 Observability

* **Prometheus metrics** out of the box
* **Structured JSON logging**
* **Request tracing** with correlation IDs
* **Health check endpoints**
* **Performance metrics** per webhook

### 🎛️ Zero-Code Configuration

* **Kubernetes-style YAML** configuration
* **Environment variable** substitution
* **File-based secrets** management
* **Hot reload** configuration changes
* **Validation** before applying changes

## Use Cases

### GitHub/GitLab CI/CD Integration

Receive repository events and trigger CI/CD pipelines:

```yaml
webhooks:
  - name: ci-trigger
    entrypointUrl: /github/push
    security:
      type: github
    storage:
      - type: rabbitmq  # Queue for CI workers
```

### Payment Processing

Handle Stripe/PayPal webhooks for payment events:

```yaml
webhooks:
  - name: payments
    entrypointUrl: /stripe/events
    security:
      type: stripe
    storage:
      - type: postgres  # Audit trail
      - type: redis     # Real-time processing
```

### Multi-Tenant SaaS

Route webhooks to tenant-specific backends:

```yaml
webhooks:
  - name: tenant-webhooks
    entrypointUrl: /tenants/{id}/webhooks
    storage:
      - type: postgres
        specs:
          query: INSERT INTO tenant_{{ .PathParams.id }}_events
```

### Event Aggregation

Collect events from multiple sources into a data lake:

```yaml
webhooks:
  - name: event-collector
    entrypointUrl: /events/{source}
    storage:
      - type: redis      # Hot data
      - type: postgres   # Warm data
      - type: s3         # Cold data (coming soon)
```

### Webhook Debugging

Development and testing of webhook integrations:

```yaml
webhooks:
  - name: debug
    entrypointUrl: /debug
    security:
      type: noop  # No auth for testing
    response:
      formatting:
        templateString: |
          {
            "received": {{ .Payload }},
            "headers": {{ .Request.Header | toJSON }}
          }
```

## Comparison with Alternatives

| Feature             | Webhooked  | Webhook.site | ngrok   |
| ------------------- | ---------- | ------------ | ------- |
| Performance         | 50K+ req/s | Limited      | Limited |
| Multi-storage       | ✅          | ❌            | ❌       |
| Security Providers  | Multiple   | Basic        | Tunnel  |
| Data Transformation | Advanced   | ❌            | ❌       |
| Self-hosted         | ✅          | ❌            | Partial |
| Configuration       | YAML       | UI           | CLI     |
| Monitoring          | Built-in   | Basic        | Basic   |
| Rate Limiting       | ✅          | Limited      | ❌       |
| Open Source         | ✅          | ❌            | ❌       |

{% hint style="info" %}
_Help is wanted to fullfil this section with more data._
{% endhint %}

## When to Use Webhooked

**Webhooked is perfect when you need:**

* High-performance webhook processing
* Multiple webhook sources with different auth methods
* Data transformation before storage
* Multiple storage backends
* Production-grade reliability
* Centralized webhook management
* Compliance and audit requirements

**Consider alternatives when:**

* You only need temporary webhook testing (use webhook.site)
* You need webhook tunneling to localhost (use ngrok)

## Core Philosophy

Webhooked follows these design principles:

1. **Performance First**: Every design decision prioritizes throughput and latency
2. **Configuration over Code**: Declarative YAML configuration for all features
3. **Security by Default**: Secure defaults with explicit opt-out
4. **Cloud Native**: Kubernetes-ready, 12-factor app principles
5. **Observable**: Metrics and logging built-in, not bolted-on
6. **Extensible**: Plugin architecture for custom providers

## Project Status

* **Current Version**: v1alpha2
* **Production Ready**: Yes, used in production by multiple companies
* **API Stability**: Alpha (breaking changes possible)
* **License**: Dual licensed (AGPL-3.0 / Commercial)
* **Maintained by**: 42Atomys and community contributors



[^1]: _tested on Apple Macbook Pro Intel 2019_
